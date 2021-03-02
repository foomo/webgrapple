package cmd

import (
	"fmt"

	"github.com/foomo/webgrapple/server"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// Used for flags.
	logger  *zap.Logger
	cfgFile string

	flagPort           = ":80"
	flagCert           = ""
	flagKey            = ""
	flagServiceAddress = server.DefaultServiceAddress

	rootCmd = &cobra.Command{
		Use:   "webgrapple",
		Short: "reverse proxy a remote server and hijack routes",
		Long:  "a proxy and a client to take over routes of a remote server with local web servers",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	//	rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")

	rootCmd.AddCommand(reverseProxyCmd)
	rootCmd.AddCommand(clientNPMCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
