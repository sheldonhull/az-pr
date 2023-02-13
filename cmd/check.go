//nolint:gochecknoglobals // cobra uses globals for commands
package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "🧪 Check your system for the correct requirements",
	Long: `Checks your system for tooling that is required and environment variables.

For example:

- Required environment variable for authenticating with Azure CLI
- Detection of the Azure CLI.
`,
	Run: func(cmd *cobra.Command, args []string) {
		check()
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func check() {
	failureCount := 0
	report := pterm.TableData{
		[]string{"Status", "Check", "Value", "Notes"},
	}
	checkReport, err := checkAzEnv()
	if err != nil {
		failureCount++
	}
	report = append(report, checkReport)
	checkReport, err = checkAzureCLI()
	if err != nil {
		failureCount++
	}
	report = append(report, checkReport)

	primary := pterm.NewStyle(pterm.FgLightCyan, pterm.BgGray, pterm.Bold)
	pterm.DefaultHeader.Printfln("az-pr requirements check")
	if err := pterm.DefaultTable.WithHasHeader().
		WithBoxed(true).
		WithHeaderStyle(primary).
		WithData(report).Render(); err != nil {
		pterm.Error.WithShowLineNumber(true).WithLineNumberOffset(1).Printfln(
			"pterm.DefaultTable.WithHasHeader of variable information failed. Continuing...%v",
			err,
		)
	}
	if failureCount > 0 {
		pterm.Error.Printfln("failureCount: %d", failureCount)
		os.Exit(1)
	} else {
		pterm.Success.Printfln("no failures: %d", failureCount)
	}
}

// checkAzureCLI checks if Azure CLI is installed.
func checkAzureCLI() ([]string, error) {
	_, err := exec.LookPath("az")
	if err != nil {
		// pterm.Error.Println("🧪 Azure CLI is not installed.\n" +
		// 	"This is required to be able to create pull requests.\n" +
		// 	"👉 Recommend installing Azure CLI: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli")
		return []string{"❌", "azure-cli", "required", "not installed"}, errors.New("Azure CLI not installed")
	}
	return []string{"✅", "azure-cli", "installed", ""}, nil
}

// checkAzEnv checks if AZURE_DEVOPS_EXT_PAT is set as env variable which is required for the Azure CLI to function.
func checkAzEnv() ([]string, error) {
	_, ok := os.LookupEnv("AZURE_DEVOPS_EXT_PAT")
	if !ok || os.Getenv("AZURE_DEVOPS_EXT_PAT") == "" {
		// pterm.Error.Println("🧪 AZURE_DEVOPS_EXT_PAT is not set as env variable.\n" +
		// 	"This is required to be able to create pull requests.\n" +
		// 	"👉 Recommend export AZURE_DEVOPS_EXT_PAT=\"\" in your $HOME/.envrc, .zshenv, .bashprofile, etc.")
		return []string{"❌", "AZURE_DEVOPS_EXT_PAT", "not set", "env var is required for auth"}, errors.New("AZURE_DEVOPS_EXT_PAT not set")
	}
	return []string{"✅", "AZURE_DEVOPS_EXT_PAT", "detected", ""}, nil
}
