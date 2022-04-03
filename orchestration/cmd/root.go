package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sbaeurle/comb/orchestration/config"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "orchestration",
	Short: "Lightweight suite to schedule benchmark on configured devices.",
}

var (
	backend     string
	cfgFile     string
	development bool
	log         config.Logger
	cfg         config.Config
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(matchingCmd)

	// Add configuration options
	rootCmd.PersistentFlags().StringVar(&backend, "backend", "ssh", "backend used to schedule the workload. Available: (ssh, k8s, edge-io)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVar(&development, "development", false, "development mode")
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	// Check for configuration errors.
	// Exit application if errors are present.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	viper.Unmarshal(&cfg)

	var logger *zap.Logger
	switch development {
	case true:
		logger, err = zap.NewDevelopment()
	case false:
		logger, err = zap.NewProduction()
	}
	cobra.CheckErr(err)
	defer logger.Sync()
	log = logger.Sugar()
}
