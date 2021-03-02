package clientnpm

import (
	"os"

	"github.com/foomo/webgrapple/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagServiceMimetypes = []string{}
	flagServicePath      = ""
	flagNPMDebug         = false
	flagStartVSCode      = true
	Command              = &cobra.Command{
		Use:   "client-npm",
		Short: "client to hook up a npm server",
		Long:  `long client ....`,
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
				logger, flagNPMDebug, flagStartVSCode,
				flagServicePath, flagServiceMimetypes,
				wd, npmCommand, npmArgs...,
			)
			if errRun != nil {
				logger.Error("run failed", zap.Error(errRun))
			}
		},
	}
)

func init() {
	Command.Flags().BoolVar(&flagNPMDebug, "debug", flagNPMDebug, "start with a debug port")
	Command.Flags().BoolVar(&flagStartVSCode, "debug-vscode", flagNPMDebug, "start a debug session in vscode")
	Command.Flags().StringVar(&flagServicePath, "service-path", flagServicePath, "register service for a path")
	Command.Flags().StringArrayVar(&flagServiceMimetypes, "service-mimes", flagServiceMimetypes, "register service for mime types")
}
