package server

import (
	"github.com/foomo/webgrapple/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagAddresses      = []string{"https://localhost"}
	flagBackendURL     = ""
	flagPort           = ":80"
	flagCert           = ""
	flagKey            = ""
	flagServiceAddress = DefaultServiceAddress

	Command = &cobra.Command{
		Use:   "reverse-proxy",
		Short: "reverse proxy",
		Long:  `reverse proxy ....`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := utils.GetLogger()
			logger.Info("running a local reverse proxy server", zap.Strings("addresses", flagAddresses))
			errSign := run(logger, flagServiceAddress, flagBackendURL, flagAddresses, flagCert, flagKey)
			if errSign != nil {
				logger.Error("could not run server", zap.Error(errSign))
			}
		},
	}
)

func init() {
	Command.Flags().StringArrayVarP(&flagAddresses, "addresses", "a", flagAddresses, "what adresses to listen to / self sign a cert for")
	Command.Flags().StringVarP(&flagCert, "cert", "c", flagCert, "cert file relative path")
	Command.Flags().StringVarP(&flagKey, "key", "k", flagKey, "key file relative path")
	Command.Flags().StringVarP(&flagBackendURL, "backend", "b", flagBackendURL, "backend url")
	Command.Flags().StringVarP(&flagServiceAddress, "service-addr", "", flagServiceAddress, "service address url")
}
