package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/sbaeurle/comb/metrics/config"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "evaluation",
	Short: "Lightweight metric system to collect benchmark data.",
}

var (
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

	rootCmd.AddCommand(serveCmd)
	// Add configuration options
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVar(&development, "development", false, "development mode")
	rootCmd.PersistentFlags().Int("buffer-size", 10, "channel size")
	rootCmd.PersistentFlags().Bool("plot", false, "enable result plotting")
	rootCmd.PersistentFlags().Int("port", 8000, "http port")
	viper.BindPFlag("BufferSize", rootCmd.PersistentFlags().Lookup("buffer-size"))
	viper.BindPFlag("Port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("GeneratePlots", rootCmd.PersistentFlags().Lookup("plot"))
}

func initConfig() {
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	// Check for configuration errors.
	// Exit application if errors are present.
	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	err = viper.Unmarshal(&cfg)
	cobra.CheckErr(err)

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
