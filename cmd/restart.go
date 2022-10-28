package cmd

import (
	"github.com/FarmbotSimulator/FarmbotSessionManager/src/systemd"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart farmbotsimulator service",
	Long:  `Restart farmbotsimulator service`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		systemd.Restart(production)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
