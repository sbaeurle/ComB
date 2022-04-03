package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/sbaeurle/comb/orchestration/matching"
)

var matchingCmd = &cobra.Command{
	Use:   "matching",
	Short: "Generate Matchings",
	RunE:  listMatching,
}

func listMatching(cmd *cobra.Command, args []string) error {
	matchings := matching.GenerateSchedules(cfg)

	tmp, err := json.Marshal(matchings)
	if err != nil {
		log.Error(err)
		return err
	}

	f, err := os.OpenFile("matchings.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error(err)
		return err
	}
	defer f.Close()

	_, err = f.Write(tmp)
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}
