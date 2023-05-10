package cmd

import (
	"os"

	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/req"
	"github.com/spf13/cobra"
)

// fireworks turns on fireworks mode for confetti command
var fireworks bool

// confettiCmd represents the confetti command
var confettiCmd = &cobra.Command{
	Use:   "confetti",
	Short: "Life needs more 🎉",
	Long:  `being able to access confetti on demand is almost as good as tacos on demand`,
	Run: func(cmd *cobra.Command, args []string) {
		pterm.Debug.Printfln("resolving confetti")
		binary, err := req.ResolveBinaryByInstall("confetty", "github.com/maaslalani/confetty@latest")
		if err != nil {
			pterm.Error.Printfln("issue setting up confetty: %v", err)
		}
		pterm.Debug.Printfln("running confetti")
		additionalConfettiArgs := []string{}

		if fireworks {
			pterm.Info.Printfln("you wanted fireworks... i'll give you fireworks. 🎆")
			additionalConfettiArgs = append(additionalConfettiArgs, "fireworks")
		}
		if err := sh.RunV(binary, additionalConfettiArgs...); err != nil {
			pterm.Warning.Printfln("i'm sorry that confetti didn't seem to work. complain to sheldon 😀")
			os.Exit(1)
		}
	},
}
