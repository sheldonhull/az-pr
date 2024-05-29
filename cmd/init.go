package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "⚙️ Set up azure-cli for correct project and plugins",
	Long: `Azure CLI requires plugins such as azure-devops to be installed.

This configures the cli to install the plugin and not require prompting.
Additionally, it prompts for the preferred project.

This can let you change projects for PR creation.
I've not found the auto-detection to be reliable, so this works better to be intentional on where to push a PR.`,
	Run: func(cmd *cobra.Command, args []string) {
		setup()
		pterm.Success.Println("init() success")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func setup() {
	if Debug {
		pterm.EnableDebugMessages()
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cmd := exec.Command("az", "extension", "add", "--name", "azure-devops")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		pterm.Error.Printfln("init() az extension add --name azure-devops: %v", err)
	}
	pterm.Success.Println("init() az extension add --name azure-devops")

	answers := struct {
		Organization string `survey:"organization"`
		Project      string `survey:"Project"`
	}{}

	// the questions to ask
	qs := []*survey.Question{
		{
			Name: "organization",
			Prompt: &survey.Input{
				Message: "organization (in format: 'https://dev.azure.com/ORG')",
				Suggest: func(toComplete string) []string {
					return []string{"https://dev.azure.com/"}
				},
			},
		},
		{
			Name:   "project",
			Prompt: &survey.Input{Message: "Project"},
		},
	}

	// perform the questions
	err := survey.Ask(qs, &answers)
	if err != nil {
		pterm.Warning.Printfln("init() you changed your mind: %v", err)
		os.Exit(0)
	}

	cmd = exec.Command("az", "devops", "configure", "--defaults", fmt.Sprintf("organization=%s", answers.Organization), fmt.Sprintf("project=%s", answers.Project))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		pterm.Error.Printfln("init() az extension add --name azure-devops: %v", err)
	}
	pterm.Success.Printfln("init() %s", cmd.String())
}
