package cmd

import (
	"os"

	"github.com/foomo/webgrapple/clientnpm"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flagNPMDebug    = false
	flagStartVSCode = true
	clientNPMCmd    = &cobra.Command{
		Use:   "client-npm",
		Short: "client to hook up a npm server",
		Long:  `long client ....`,
		Run: func(cmd *cobra.Command, args []string) {
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
			errRun := clientnpm.Run(logger, flagNPMDebug, flagStartVSCode, wd, npmCommand, npmArgs...)
			if errRun != nil {
				logger.Error("run failed", zap.Error(errRun))
			}
		},
	}
)

func init() {
	clientNPMCmd.Flags().BoolVar(&flagNPMDebug, "debug", flagNPMDebug, "start with a debug port")
	clientNPMCmd.Flags().BoolVar(&flagStartVSCode, "debug-vscode", flagNPMDebug, "start a debug session in vscode")
}
