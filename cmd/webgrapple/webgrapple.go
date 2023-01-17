package webgrapple

import (
	"github.com/spf13/cobra"
)

var (
	Command = &cobra.Command{
		Use:   "webgrapple",
		Short: "reverse proxy a remote server and hijack routes",
		Long:  "a proxy and a client to take over routes of a remote server with local web servers",
	}
	flagPort = 0
)

func init() {
	Command.AddCommand(serverCmd)
	Command.AddCommand(clientNPMCmd)
}
