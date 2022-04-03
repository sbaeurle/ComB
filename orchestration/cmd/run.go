package cmd

import (
	"errors"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/sbaeurle/comb/orchestration/executor"
	"github.com/sbaeurle/comb/orchestration/executor/ssh"
	"github.com/sbaeurle/comb/orchestration/matching"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run benchmark and display results.",
	RunE:  benchmark,
}

func benchmark(cmd *cobra.Command, args []string) error {
	switch backend {
	case "ssh":
		var exec executor.Executor
		exec, err := ssh.NewSSHExecutor(log, &cfg)
		if err != nil {
			return err
		}

		matchings := matching.GenerateSchedules(cfg)
		log.Infof("Generated Matchings: %v", matchings)

		_, err = http.Post(cfg.Evaluation+"/start-benchmark", "application/json", nil)
		if err != nil {
			return err
		}
		for _, matching := range matchings {
			err := exec.RunMatching(matching)
			cobra.CheckErr(err)
		}
		_, err = http.Post(cfg.Evaluation+"/end-benchmark", "application/json", nil)
		if err != nil {
			return err
		}
	default:
		return errors.New("incorrect benchmarking backend")
	}
	return nil
}
