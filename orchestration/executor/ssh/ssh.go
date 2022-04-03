package ssh

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/sbaeurle/comb/orchestration/config"
)

// Executor to schedule benchmark runs using SSH and Docker.
type SSHExecutor struct {
	log       config.Logger
	cfg       *config.Config
	templates map[string]*template.Template
}

func NewSSHExecutor(log config.Logger, cfg *config.Config) (*SSHExecutor, error) {
	templates := make(map[string]*template.Template)
	for tag, command := range cfg.SSH.Commands {
		tmp, err := template.New(tag).Parse(command)
		if err != nil {
			return nil, err
		}
		templates[tag] = tmp
	}

	return &SSHExecutor{
		log:       log,
		cfg:       cfg,
		templates: templates,
	}, nil
}

func (s *SSHExecutor) VerifyEnvironment() []error {
	var nodes []string
	for _, group := range s.cfg.NodeGroups {
		nodes = append(nodes, group.Nodes...)
	}

	errCh := make(chan error, len(nodes))
	var wg sync.WaitGroup

	// Iterate over available nodes and verify access
	for _, node := range nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			// Verify ssh connection: host reachable + valid authentication.
			conn, err := newSSHClient(s.log, s.cfg.SSH.KeyFile, s.cfg.SSH.User, node)
			if err != nil {
				errCh <- fmt.Errorf("connection error %s: %w", node, err)
				return
			}
			defer conn.close()

			// Verifies docker is running + permissions are set correctly.
			_, err = conn.executeCommand("docker run --rm hello-world")
			if err != nil {
				errCh <- fmt.Errorf("docker error %s: %w", node, err)
			}

		}(node)
	}

	wg.Wait()
	close(errCh)

	// Forward all errors as array
	var errors []error
	for err := range errCh {
		s.log.Error(err)
	}

	return errors
}

func (s *SSHExecutor) RunMatching(matching map[string]string) error {
	nodeCache := make(map[string]*config.NodeGroup)
	for i, node := range s.cfg.NodeGroups {
		nodeCache[node.Name] = &s.cfg.NodeGroups[i]
	}

	workloadCache := make(map[string]*config.WorkloadConfig)
	for i, workload := range s.cfg.Workload {
		workloadCache[workload.Name] = &s.cfg.Workload[i]
	}

	possible, runs := generateRuns(matching, nodeCache, workloadCache)
	s.log.Debugf("Generated Runs: %s", possible)
	for i := 0; i < runs; i++ {
		tmp := make(map[string]string)
		var tags []string
		var nodes []string
		for v, workload := range s.cfg.Workload {
			p := possible[workload.Name]
			tags = append(tags, p[i%len(p)])

			n := nodeCache[matching[workload.Name]].Nodes
			nodes = append(nodes, n[v%len(n)])
			tmp[workload.Name] = fmt.Sprintf("%s-%s-%s", matching[workload.Name], n[v%len(n)], p[i%len(p)])
		}
		s.log.Infof("Start Benchmark Run: %v", tmp)
		err := s.singleRun(s.cfg.Workload, tmp, nodes, tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SSHExecutor) singleRun(workloads []config.WorkloadConfig, matching map[string]string, nodes []string, tags []string) error {
	var networkAddresses = make(map[string]string)
	done := make(chan struct{})

	if len(workloads) != len(nodes) && len(workloads) != len(tags) {
		return errors.New("slices need to be of equal length")
	}
	tmp, err := json.Marshal(matching)
	if err != nil {
		return err
	}

	resp, err := http.Post(s.cfg.Evaluation+"/start-run", "application/json", bytes.NewBuffer(tmp))
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	networkAddresses["Evaluation"] = s.cfg.Evaluation
	for i, workload := range workloads {
		node := nodes[i]
		tag := tags[i]
		networkAddresses[workload.Name] = node

		// Build shell tmp to be executed on node
		command, err := createCommand(s.templates, networkAddresses, workload, tag)
		if err != nil {
			return err
		}
		s.log.Debugf("Parsed Command %s: %s", workload.Name, command)

		// Create ssh connection to node for workload
		conn, err := newSSHClient(s.log, s.cfg.SSH.KeyFile, s.cfg.SSH.User, node)
		if err != nil {
			return fmt.Errorf("connection error %s: %w", node, err)
		}
		defer func(name string) {
			conn.executeCommand(fmt.Sprintf("docker rm -f %s", name))
			conn.close()
		}(workload.Name)

		// Prepare environment for benchmark
		err = prepareEnvironment(conn, workload.Image, tag, workload.LocalData)
		if err != nil {
			s.log.Errorf("%s:%s: %s", workload.Image, tag, err)
			return err
		}
		s.log.Infof("Schedule %s on %s", workload.Name, node)

		go func(name string) {
			err := conn.issueCommand(done, command)
			if err != nil {
				s.log.Errorf("%s, %s", name, err)
			}
		}(workload.Name)

		time.Sleep(10 * time.Second) // Sleep a second to help every container to startup correctly
	}

	<-done
	resp, err = http.Post(s.cfg.Evaluation+"/end-run", "application/json", nil)
	if err != nil || resp.StatusCode != http.StatusOK {
		return err
	}

	results := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return err
	}
	s.log.Infof("%v", results)
	return nil
}

func createCommand(cache map[string]*template.Template, mapping map[string]string, workload config.WorkloadConfig, tag string) (string, error) {
	var tmp strings.Builder
	wl := struct {
		config.WorkloadConfig
		Tag string
	}{workload, tag}
	err := cache[tag].Execute(&tmp, wl)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New("tmp").Parse(tmp.String())
	if err != nil {
		return "", err
	}
	var command strings.Builder
	err = tmpl.Execute(&command, mapping)
	if err != nil {
		return "", err
	}
	return command.String(), nil
}

func prepareEnvironment(connection *sshClient, image string, tag string, files []string) error {
	// Preload Image & ensure most recent image version TODO: Log console output
	_, err := connection.executeCommand(fmt.Sprintf("docker pull %s:%s", image, tag))
	if err != nil {
		return err
	}

	// Copy needed files to remote node
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		filename := strings.Split(file, "/")

		err = connection.copyFile(f, "/tmp/"+filename[len(filename)-1])
		if err != nil {
			return err
		}
	}
	return nil
}

func generateRuns(matching map[string]string, nodeCache map[string]*config.NodeGroup, workloadCache map[string]*config.WorkloadConfig) (map[string][]string, int) {
	possible := make(map[string][]string, len(matching))
	var runs int

	for wl, n := range matching {
		var possibleTags []string
		node := nodeCache[n]
		workload := workloadCache[wl]
		for _, tag := range workload.Tags {
			for _, cap := range node.Capabilities {
				if tag == cap {
					possibleTags = append(possibleTags, tag)
				}
			}
		}
		if len(possibleTags) > runs {
			runs = len(possibleTags)
		}
		possible[wl] = possibleTags
	}

	return possible, runs
}
