package cmd

import (
	"github.com/foomo/webgrapple/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagAddresses   = []string{"https://localhost"}
	flagBackendURL  = ""
	reverseProxyCmd = &cobra.Command{
		Use:   "reverse-proxy",
		Short: "reverse proxy",
		Long:  `reverse proxy ....`,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("running a local reverse proxy server", zap.Strings("addresses", flagAddresses))
			errSign := server.Run(logger, flagServiceAddress, flagBackendURL, flagAddresses, flagCert, flagKey)
			if errSign != nil {
				logger.Error("could not run server", zap.Error(errSign))
			}

		},
	}
)

func init() {
	reverseProxyCmd.Flags().StringArrayVarP(&flagAddresses, "addresses", "a", flagAddresses, "what adresses to listen to / self sign a cert for")
	reverseProxyCmd.Flags().StringVarP(&flagCert, "cert", "c", flagCert, "cert file relative path")
	reverseProxyCmd.Flags().StringVarP(&flagKey, "key", "k", flagKey, "key file relative path")
	reverseProxyCmd.Flags().StringVarP(&flagBackendURL, "backend", "b", flagBackendURL, "backend url")
	reverseProxyCmd.Flags().StringVarP(&flagServiceAddress, "service-addr", "", flagServiceAddress, "service address url")
}
