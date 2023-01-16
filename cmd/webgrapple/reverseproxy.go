package webgrapple

import (
	"github.com/foomo/webgrapple/pkg/server"
	"github.com/foomo/webgrapple/pkg/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const DefaultServiceAddress = "127.0.0.1:8888"

var (
	flagAddresses      = []string{"https://localhost"}
	flagBackendURL     = ""
	flagCert           = ""
	flagKey            = ""
	flagServiceAddress = DefaultServiceAddress

	serverCmd = &cobra.Command{
		Use:   "reverse-proxy",
		Short: "reverse proxy",
		Long:  `reverse proxy ....`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := utils.GetLogger()
			logger.Info("running a local reverse proxy server", zap.Strings("addresses", flagAddresses))
			errRun := server.Run(
				cmd.Context(),
				logger.Sugar(),
				flagServiceAddress,
				flagBackendURL,
				flagAddresses,
				flagCert,
				flagKey,
			)
			if errRun != nil {
				logger.Error("could not run server", zap.Error(errRun))
			}
		},
	}
)

func init() {
	serverCmd.Flags().StringArrayVarP(&flagAddresses, "addresses", "a", flagAddresses, "what adresses to listen to / self sign a cert for")
	serverCmd.Flags().StringVar(&flagCert, "cert", flagCert, "cert file relative path")
	serverCmd.Flags().StringVar(&flagKey, "key", flagKey, "key file relative path")
	serverCmd.Flags().StringVar(&flagBackendURL, "backend", flagBackendURL, "backend url")
	serverCmd.Flags().StringVar(&flagServiceAddress, "service-addr", flagServiceAddress, "service address url")
}
