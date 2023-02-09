/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "🚀 Create a new PR",
	Long:  `Use this command to create a new PR. It will ask you a few questions and help you create a PR with an interactive prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		gatherInput()
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//nolint:gochecknoglobals // gochecknoglobals: these are globals used for conventional commit and more visible as globals for updates
var (
	// _conventionalCommitTypes is a collection of conventional commit types for PR creation.
	_conventionalCommitTypes = []string{
		"feat",
		"fix",
		"chore",
		"refactor",
		"test",
		"docs",
		"style",
		"perf",
		"ci",
		"build",
		"revert",
	}
)

// emojify returns a nice emoji for the given commit type.
// Emoji's make it easier to smile. :).
// Trailing space is to ensure it doesn't run into the scope message.
func emojify(commitTypeString string) string {
	switch commitTypeString {
	case "feat":
		return "✨" + " "
	case "fix":
		return "🐛" + " "
	case "chore":
		return "🔨" + " "
	case "refactor":
		return "🛠️" + " "
	case "test":
		return "🧪" + " "
	case "docs":
		return "📘" + " "
	case "style":
		return "🎨" + " "
	case "perf":
		return "⚡" + " "
	case "ci":
		return "🚀" + " "
	case "build":
		return "👷" + " "
	case "revert":
		return "💩" + " "
	default:
		return "" + " "
	}
}

// the questions to ask
var qs = []*survey.Question{
	{
		Name: "type",
		Prompt: &survey.Select{
			Message: "Choose a conventional commit type",
			Options: _conventionalCommitTypes,
		},
	},
	{
		Name:      "scope",
		Prompt:    &survey.Input{Message: "Scope (optional)"},
		Transform: survey.ToLower,
	},
	{
		Name:      "title",
		Prompt:    &survey.Input{Message: "Title"},
		Transform: survey.ToLower,
	},
	{
		Name:   "description",
		Prompt: &survey.Multiline{Message: "Description (imperative & active voice)", Help: "Write imperative and active voice \nwith multiple lines for each bullet point."},
	},
}

func gatherInput() {
	var err error
	// the answers will be written to this struct
	answers := struct {
		Type        string `survey:"type"`
		Scope       string `survey:"color"`
		Title       string `survey:"title"`
		Description string `survey:"description"`
	}{}

	// perform the questions
	err = survey.Ask(qs, &answers)
	if err != nil {
		pterm.Error.Printfln("gatherInput: %v", err)
		os.Exit(1)
	}
	if answers.Scope != "" {
		answers.Scope = "(" + answers.Scope + "):"
	} else {
		answers.Scope = ":"
	}

	pterm.Info.Printfln("%s%s %s%s", answers.Type, answers.Scope, emojify(answers.Type), answers.Title)

	pterm.Info.Println("\n" + answers.Description)
	fmt.Println("") // will need to the processing here.
}
