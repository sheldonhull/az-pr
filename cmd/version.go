package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// version is the semver version, and is updated by changie
var ( //nolint:gochecknoglobals // this is fine, as version info
	// Version is the descriptive version, normally the tag from which the app was built.
	// Since git tags can be changed, use Commit instead as the most accurate version.
	version = "dev"
	// Commit is the git commit hash that the build was generated from.
	commit = "none"
	// Date is the date the binary was produced.
	date = "unknown"
	// buildName is the build name for easier confirmation on local builds that a build has changed.
	buildName = "unknown"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

// versionCmd represents the custom version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "List the current version info of the CLI",
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("az-pr")
		pterm.Info.Printfln("version: %s", version)
		pterm.Info.Printfln("commit: %s", commit)
		pterm.Info.Printfln("date: %s", date)
		pterm.Info.Printfln("buildName: %s", buildName)
	},
}
