package clientnpm

import (
	"os"

	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagDebugServerPort = 0
	flagStartVSCode     = false
	flagReverseProxyURL = server.DefaultServiceURL
	flagPort            = 0
	// Command use this for NPM support, when composing your own webgrapple
	Command = &cobra.Command{
		Use:   "client-npm",
		Short: "client to hook up a npm / Node.js server",
		Long: `allows you to webgrapple a Node.js server and has debugging support for vscode

- client-npm assumes that your Node.js server will respond on http://127.0.0.1:<flagPort>

		`,
		Run: func(cmd *cobra.Command, args []string) {
			logger := utils.GetLogger()
			wd, errWd := os.Getwd()
			if errWd != nil {
				logger.Error("could not determine working directory", zap.Error(errWd))
				return
			}
			npmCommand := "yarn"
			if len(args) > 0 {
				npmCommand = args[0]
			}
			npmArgs := []string{}
			if len(args) > 1 {
				npmArgs = args[1:]
			}
			errRun := Run(
				logger,
				flagReverseProxyURL,
				flagPort, flagDebugServerPort, flagStartVSCode,
				wd, npmCommand, npmArgs...,
			)
			if errRun != nil {
				logger.Error("run failed", zap.String("error", errRun.Error()))
			}
			logger.Info("shutting down")
		},
	}
)

func init() {
	Command.Flags().StringVar(&flagReverseProxyURL, "reverse-proxy-url", flagReverseProxyURL, "reverse proxy url")
	Command.Flags().IntVar(&flagDebugServerPort, "debug-port", flagDebugServerPort, "start debug session on the given port NODE_DEBUG_PORT will be set")
	Command.Flags().BoolVar(&flagStartVSCode, "debug-vscode", flagStartVSCode, "start a debug session in vscode, if no debug-port is defined it will be automatically assigned in NODE_DEBUG_PORT")
	Command.Flags().IntVar(&flagPort, "port", flagPort, "which port to use, if 0 client-npm will look for a free port and set env PORT")
}
