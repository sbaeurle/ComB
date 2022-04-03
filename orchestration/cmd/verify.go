package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/sbaeurle/comb/orchestration/executor"
	"github.com/sbaeurle/comb/orchestration/executor/ssh"
)

var initCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify configuration and access to all devices.",
	RunE:  verify,
}

func verify(cmd *cobra.Command, args []string) error {
	switch backend {
	case "ssh":
		var exec executor.Executor
		exec, err := ssh.NewSSHExecutor(log, &cfg)
		cobra.CheckErr(err)
		errs := exec.VerifyEnvironment()
		if len(errs) != 0 {
			cobra.CheckErr(errs)
		}
	default:
		return errors.New("incorrect benchmarking backend")
	}

	return nil
}
