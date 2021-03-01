package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reverseProxyCmd = &cobra.Command{
	Use:   "reverse-proxy",
	Short: "reverse proxy",
	Long:  `reverse proxy ....`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("running a local reverse proxy server on port", flagPort)
	},
}

func init() {
	reverseProxyCmd.Flags().StringVarP(&flagPort, "port", "p", ":8080", "define a local port")
}
