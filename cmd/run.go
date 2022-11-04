package cmd

import (
	"fmt"
	"os"

	"github.com/FarmbotSimulator/farmbotProxy/src/server"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run farmbotproxy",
	Long:  `Run farmbotproxy.`,
	Run: func(cmd *cobra.Command, args []string) {
		production, _ := cmd.Flags().GetBool("prod")
		if production {
			os.Setenv("env", "prod")
			fmt.Println("Starting Server in Production Enviroment")
		} else {
			os.Setenv("env", "dev")
			fmt.Println("Starting Server in Development Enviroment")
		}
		server.Start(production)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
