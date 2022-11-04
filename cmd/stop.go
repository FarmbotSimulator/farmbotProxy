package cmd

import (
	"github.com/FarmbotSimulator/farmbotProxy/src/systemd"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop farmbotproxy service",
	Long:  `Stop farmbotproxy service`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		systemd.Stop(production)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
