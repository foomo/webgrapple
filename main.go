package main

import (
	"github.com/foomo/webgrapple/clientnpm"
	"github.com/foomo/webgrapple/server"
	"github.com/foomo/webgrapple/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	rootCmd = &cobra.Command{
		Use:   "webgrapple",
		Short: "reverse proxy a remote server and hijack routes",
		Long:  "a proxy and a client to take over routes of a remote server with local web servers",
	}
)

func init() {
	rootCmd.AddCommand(server.Command)
	rootCmd.AddCommand(clientnpm.Command)
}

func main() {
	errExecute := rootCmd.Execute()
	if errExecute != nil {
		utils.GetLogger().Error("execution error", zap.Error(errExecute))
	}
}
