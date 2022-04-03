package cmd

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/sbaeurle/comb/metrics/routes"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve Metric System.",
	RunE:  serve,
}

func serve(cmd *cobra.Command, args []string) error {
	r := mux.NewRouter()

	modz, err := routes.RegisterRoutes(r, log, cfg)
	if err != nil {
		return err
	}

	control, err := routes.NewControlService(log, cfg, modz)
	if err != nil {
		return nil
	}

	r.HandleFunc("/start-benchmark", control.StartBenchmark).Methods("POST")
	r.HandleFunc("/end-benchmark", control.EndBenchmark).Methods("POST")
	r.HandleFunc("/start-run", control.StartRun).Methods("POST")
	r.HandleFunc("/end-run", control.EndRun).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r))

	return nil
}
