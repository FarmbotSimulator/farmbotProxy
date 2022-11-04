package cmd

import (
	"github.com/FarmbotSimulator/farmbotProxy/src/systemd"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart farmbotproxy service",
	Long:  `Restart farmbotproxy service`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		systemd.Restart(production)
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
