package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/sbaeurle/comb/metrics/config"
	"github.com/sbaeurle/comb/metrics/modules"
)

type ControlService struct {
	log     config.Logger
	cfg     config.Config
	modz    map[string]modules.Module
	mapping map[string]string
	root    string
	path    string
	run     int
}

func NewControlService(log config.Logger, cfg config.Config, modz map[string]modules.Module) (*ControlService, error) {
	return &ControlService{log: log, cfg: cfg, modz: modz}, nil
}

func (cs *ControlService) StartRun(w http.ResponseWriter, r *http.Request) {
	var tmp map[string]string
	err := json.NewDecoder(r.Body).Decode(&tmp)
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cs.mapping = tmp
	cs.run++

	cs.path, err = filepath.Abs(fmt.Sprintf("%s/run%03d", cs.root, cs.run))
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = os.Mkdir(cs.path, 0755)
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, m := range cs.modz {
		err = m.StartMeasurement(cs.path)
		if err != nil {
			cs.log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (cs *ControlService) EndRun(w http.ResponseWriter, r *http.Request) {
	results := make(map[string]map[string]float64)
	for k, m := range cs.modz {
		tmp, err := m.CollectMetrics()
		if err != nil {
			cs.log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		results[k] = tmp
	}

	output := struct {
		Matching map[string]string             `json:"matching"`
		Results  map[string]map[string]float64 `json:"results"`
	}{
		Matching: cs.mapping,
		Results:  results,
	}

	tmp, err := json.Marshal(output)
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	path, err := filepath.Abs(fmt.Sprintf("%s/%s", cs.path, "results.json"))
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = f.Write(tmp)
	if err != nil {
		cs.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(tmp)
}

func (cs *ControlService) EndBenchmark(w http.ResponseWriter, r *http.Request) {
	if cs.cfg.GeneratePlots {
		cmd := exec.Command("python3", cs.cfg.PlottingScript, "--runs", fmt.Sprintf("%d", cs.run), "--path", cs.root)
		output, err := cmd.CombinedOutput()
		if err != nil {
			cs.log.Errorf("%v: %v", err, output)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func (cs *ControlService) StartBenchmark(w http.ResponseWriter, r *http.Request) {
	cs.root = fmt.Sprintf("%s/%s", cs.cfg.RootFolder, time.Now().Format("20060102_150405"))
	err := os.MkdirAll(cs.root, 0755)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cs.run = 0

	w.WriteHeader(http.StatusOK)
}
