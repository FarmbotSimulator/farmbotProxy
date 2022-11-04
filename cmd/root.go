package cmd

import (
	"fmt"
	"os"

	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
)

// var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "farmbotproxy",
	Short: "Farmbot Proxy.",
	Long: `Farmbot Proxy.
Manages sessions to farmbot backend server
	`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {  },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ascii := figlet4go.NewAsciiRender()

	// Adding the colors to RenderOptions
	options := figlet4go.NewRenderOptions()
	color1, _ := figlet4go.NewTrueColorFromHexString("885DBA")
	options.FontColor = []figlet4go.Color{
		// Colors can be given by default ansi color codes...
		figlet4go.ColorGreen,
		figlet4go.ColorYellow,
		figlet4go.ColorCyan,
		color1,
		// ...or by an TrueColor object with rgb values
		// figlet4go.TrueColor{13, 93, 186}, // too many values in struct literal
	}

	renderStr, _ := ascii.RenderOpts("Farmbot Proxy", options)

	// options.FontName = "larry3d"
	fmt.Print(renderStr)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolP("prod", "p", false, "Run in production")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// config.Config(false)
	// don't recreate service file

}
