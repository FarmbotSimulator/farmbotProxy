package cmd

import (
	"github.com/FarmbotSimulator/FarmbotSessionManager/config"
	"github.com/FarmbotSimulator/FarmbotSessionManager/src/systemd"
	"github.com/spf13/cobra"
)

var force bool
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create service and config files",
	Long: `Create service and config files. Requires sudo
Will overwrite service file if exists, but will not overwrite config files if exist unless the -f flag is supplied.

Location for config files is in /etc/farmbotsimulator/farmbotsimulator.yaml

`,
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		production, _ := cmd.Flags().GetBool("prod")
		config.Config(force)
		systemd.Install(production)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolP("force", "f", false, "Overwrite config file if exists")
}
