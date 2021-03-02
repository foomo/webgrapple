package cmd

import (
	"github.com/foomo/webgrapple/server"
	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(reverseProxyCmd)
	rootCmd.AddCommand(clientNPMCmd)
}
