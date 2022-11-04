package cmd

import (
	"github.com/FarmbotSimulator/farmbotProxy/src/systemd"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start farmbotproxy service",
	Long:  `Start farmbotproxy service`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		systemd.Start(production)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
