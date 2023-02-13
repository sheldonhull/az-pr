//nolint:gochecknoglobals // cobra uses globals for commands
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/go-git/go-git/v5"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "🚀 Create a new PR",
	Long:  `Use this command to create a new PR. It will ask you a few questions and help you create a PR with an interactive prompt.`,
	Run: func(cmd *cobra.Command, args []string) {
		createPR()
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
		Validate: survey.Required,
	},
	{
		Name:      "scope",
		Prompt:    &survey.Input{Message: "Scope (optional, enter to skip)"},
		Transform: survey.ToLower,
	},
	{
		Name:      "title",
		Prompt:    &survey.Input{Message: "Title"},
		Transform: survey.ToLower,
		Validate:  survey.Required,
	},
	// ,
	// {
	// 	Name: "description",
	// 	Prompt: &survey.Multiline{
	// 		Message: "Description (imperative & active voice)",
	// 		Help:    "Write imperative and active voice \nwith multiple lines for each bullet point.",
	// 	},
	// },
}

func gatherInput() (title, description string) {
	if Debug {
		pterm.EnableDebugMessages()
	}
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
		pterm.Warning.Printfln("gatherInput() you must have forgotten something: %v", err)
		os.Exit(0)
	}
	p := tea.NewProgram(initialModel())

	mod, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}
	myModelInstance, ok := mod.(model)
	if !ok {
		pterm.Warning.Printfln("gatherInput() issue in gather input from text input. I goofed. not your fault. ## ShouldHaveDoneTDD: %v", err)
	}
	description = myModelInstance.textarea.Value()
	if answers.Scope != "" {
		answers.Scope = "(" + answers.Scope + "):"
	} else {
		answers.Scope = ":"
	}
	title = fmt.Sprintf("%s%s %s%s", answers.Type, answers.Scope, emojify(answers.Type), answers.Title)
	pterm.Info.Println(title)

	// pterm.Info.Println("\n" + answers.Description)
	pterm.Info.Printfln("\n%s", description)
	return title, description
}

func getUpstreamBranch() (branchName string, err error) {
	if Debug {
		pterm.EnableDebugMessages()
	}
	r, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("unable to open git repo: %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return "", fmt.Errorf("unable to get head from git: %w", err)
	}

	pterm.Debug.Printfln("[Type: %+v]\n[Hash: %+v]\n[Name: %+v]\n[Target: %+v]\n[String: %+v]\n[Name.Short: %+v]",
		head.Type(),
		head.Hash(),
		head.Name(),
		head.Target(),
		head.String(),
		head.Name().Short())

	bl, err := r.Branches()
	if err != nil {
		return "", fmt.Errorf("unable to get branches from git: %w", err)
	}

	var detectMaster bool
	var detectMain bool

	err = bl.ForEach(func(b *plumbing.Reference) error {
		pterm.Debug.Printfln("\tBranch.ForEach: [Type: %+v][Hash: %+v][Name: %+v][Target: %+v][Name.Short: %+v]",

			b.Type(),
			b.Hash(),
			b.Name(),
			b.Target(),
			b.Name().Short())

		switch b.Name().Short() {
		case "master":
			detectMaster = true
		case "main":
			detectMain = true
		default:
		}
		return nil
	})

	if detectMaster && detectMain {
		pterm.Warning.Println("things seem to be confusing here. you have a main and a master branch")
		if err := survey.AskOne(&survey.Input{
			Message: "Which branch do you want to use as target for push?",
			Default: "main",
			Suggest: func(toComplete string) []string {
				return []string{"main", "master"}
			},
		}, &branchName); err != nil {
			return "", fmt.Errorf("failed to input branch name: %w", err)
		}
	} else {
		if detectMaster {
			branchName = "master"
		} else if detectMain {
			branchName = "main"
		} else {
			branchName = "" // just for clarity.. not really needed :-)
			pterm.Warning.Println("unable to detect main or master branch")
		}
		pterm.Info.Printfln("autodetected target branch of: %s", branchName)
	}

	return branchName, err
}

func createPR() {
	if Debug {
		pterm.EnableDebugMessages()
	}
	branchName, err := getUpstreamBranch()
	if err != nil {
		pterm.Error.Printfln("createPR: %v", err)
		os.Exit(1)
	}

	title, description := gatherInput()
	args := []string{
		"repos", "pr", "create",
		"--output", "json", // i had issues when passing at end, so make this the first arg
		"--title", title,
		"--auto-complete", "true",
		"--delete-source-branch", "true",
		"--squash",
		"--transition-work-items", "true",
		"--open",
		"--target-branch", branchName, // can't use autodetect with ssh so have to be specific: Per error: DevOps SSH URLs are not supported for repo auto-detection yet. https://github.com/Microsoft/azure-devops-cli-extension/issues/142
		"--description", description,
	}

	// Repository contains the response URL for the PR
	type Repository struct {
		WebURL string `json:"webUrl"` //nolint:tagliatelle // this is output from azure-cli I don't control
	}
	// PullRequestResponse contains the response from the PR creation, captured the azure-cli.
	type PullRequestResponse struct {
		PullRequestID int        `json:"pullRequestId"` //nolint:tagliatelle // this is output from azure-cli I don't control
		Repository    Repository `json:"repository"`    //nolint:tagliatelle // this is output from azure-cli I don't control
	}

	cmd := exec.Command("az", args...)

	pterm.Debug.Printfln("az %s", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		pterm.Error.Printf("failure running azure-cli via az-cli:\n%v\n\n", err)
		pterm.Error.Printfln("out: %s", out)
		pterm.Error.Printfln("err: %v", err)
	}
	prResponse := PullRequestResponse{}
	if err := json.Unmarshal(out, &prResponse); err != nil {
		pterm.Error.Printf("unmarshal failure: %v\n", err)
		pterm.Debug.Printf("out:\n%s\n", string(out))
	}

	// to give better control when running in container, i want to output the url to the console to control click.
	url := fmt.Sprintf(
		"%s/pullrequest/%d",
		prResponse.Repository.WebURL,
		prResponse.PullRequestID,
	)
	pterm.Success.Printf("Pull Request Url: %s\n", url)
}

type errMsg error

type model struct {
	textarea textarea.Model
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "- Implement tacos in app..."
	ti.Focus()

	return model{
		textarea: ti,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			pterm.Warning.Println("ctrl+c pressed. exiting... make up your mind please")
			os.Exit(0)
		case tea.KeyCtrlD:
			return m, tea.Quit
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf(
		"Summary of changes and why?\n\n%s\n\n%s",
		m.textarea.View(),
		"(ctrl+d to save, ctrl+c to quit)",
	) + "\n\n"
}
