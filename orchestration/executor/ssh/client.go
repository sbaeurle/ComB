package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/sbaeurle/comb/orchestration/config"
	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	client *ssh.Client
	log    config.Logger
}

func newSSHClient(log config.Logger, keyPath string, username string, address string) (*sshClient, error) {
	// Verify address has specified port, otherwise add standard ssh port.
	if _, _, err := net.SplitHostPort(address); err != nil {
		address = net.JoinHostPort(address, "22")
	}

	keyFile, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(keyFile)
	if errors.Is(err, &ssh.PassphraseMissingError{}) {
		fmt.Println("Please enter private Key Passphrase:")
		reader := bufio.NewReader(os.Stdin)
		passphrase, _ := reader.ReadString('\n')
		key, err = ssh.ParsePrivateKeyWithPassphrase(keyFile, []byte(passphrase))
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User: username,
		// TODO: Replace with known_host parser
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		Timeout: 30 * time.Second,
	}

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return &sshClient{log: log, client: client}, nil
}

func (s *sshClient) executeCommand(command string) ([]byte, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	return session.CombinedOutput(command)
}

func (s *sshClient) issueCommand(done chan struct{}, command string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Setup Reader for Console Output
	r, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stopRead := make(chan struct{})
	go logOutput(s.log, stopRead, r, s.client.RemoteAddr().String())

	err = session.Run(command)
	done <- struct{}{}
	close(stopRead)
	return err
}

func logOutput(log config.Logger, done chan struct{}, reader io.Reader, node string) {
	b := bufio.NewReader(reader)
	for {
		select {
		case <-done:
			return
		default:
			tmp, err := b.ReadBytes('\n')
			if err == io.EOF {
				return
			}
			log.Debugf("%s: %s", node, tmp)
		}
	}
}

func (s *sshClient) copyFile(file io.Reader, location string) error {
	client, err := scp.NewClientBySSH(s.client)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.CopyFile(file, location, "0664")
	return err
}

func (s *sshClient) close() {
	s.client.Close()
}
