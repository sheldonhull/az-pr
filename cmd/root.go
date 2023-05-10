//nolint:gochecknoglobals // cobra uses globals for commands
package cmd

import (
	"os"

	shell "github.com/brianstrauch/cobra-shell"
	"github.com/c-bata/go-prompt"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "az-pr",
	Short: "Create pull requests with the azure-cli.",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("az-pr")
		pterm.Info.Printfln("version: %s", version)
		pterm.Info.Println("press ctrl+d to exit")
		shell := shell.New(cmd, nil)

		err := shell.Execute()
		if err != nil {
			pterm.Error.Println(err)
		}
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cc.Init(&cc.Config{
		RootCmd:  rootCmd,
		Headings: cc.HiCyan + cc.Bold + cc.Underline,
		Commands: cc.HiYellow + cc.Bold,
		Example:  cc.Italic,
		ExecName: cc.Bold,
		Flags:    cc.Bold,
	})
	rootCmd.AddCommand(confettiCmd)
	confettiCmd.Flags().BoolVarP(&fireworks, "fireworks", "", false, "enable fireworks mode")

	// finally register interactive
	rootCmd.AddCommand(shell.New(
		rootCmd,
		nil,
		// testcmd,
		prompt.OptionTitle("tada"),
		prompt.OptionPrefix(">>> "),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionMaxSuggestion(10), //nolint:gomnd // ok to leave this here
		prompt.OptionInputTextColor(prompt.Yellow),
		prompt.OptionCompletionOnDown(),
		prompt.OptionCompletionWordSeparator(""),
	))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var Debug bool

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.az-pr.yaml)")
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "enable debug output")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	if Debug {
		pterm.EnableDebugMessages()
	}
	// rootCmd.Flags().BoolP("debug", "d", false, "Enable debug logging")
}
