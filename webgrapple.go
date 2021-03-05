package webgrapple

import (
	"github.com/foomo/webgrapple/clientnpm"
	"github.com/foomo/webgrapple/server"
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "webgrapple",
		Short: "reverse proxy a remote server and hijack routes",
		Long:  "a proxy and a client to take over routes of a remote server with local web servers",
	}
)

func init() {
	Command.AddCommand(server.Command)
	Command.AddCommand(clientnpm.Command)
}
