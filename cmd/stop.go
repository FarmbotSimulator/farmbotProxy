package cmd

import (
	"github.com/FarmbotSimulator/FarmbotSessionManager/src/systemd"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop farmbotsimulator service",
	Long:  `Stop farmbotsimulator service`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		systemd.Stop(production)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
