package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var clientNPMCmd = &cobra.Command{
	Use:   "client-npm",
	Short: "client to hook up a npm server",
	Long:  `long client ....`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("running a local client thing for npm")
	},
}
