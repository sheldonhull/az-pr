//nolint:gochecknoglobals // cobra uses globals for commands
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alessio/shellescape"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"github.com/magefile/mage/sh"

	"github.com/go-git/go-git/v5"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "🚀 Create a new PR",
	Long: `Use this command to create a new PR.
	It will ask you a few questions and help you create a PR with an interactive prompt.

	If you don't like confetti, you can disable it by setting the environment variable CONFETTI=1.
	I'll probably add a flag to disable in the future, but till then yeah 🎊🎉`,
	Run: func(cmd *cobra.Command, args []string) {
		createPR()
	},
}

// evaluateScopeMonths is how far back in git history to evaluate scopes to provide with autocompletion.
const evaluateScopeMonths int = 6

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

func conventionalCommitTypeOptions() []huh.Option[string] {
	return huh.NewOptions(
		"ci",
		"build",
		"feat",
		"fix",
		"refactor",
		"style",
		"chore",
		"test",
		"docs",
		"perf",
		"revert",
		"security",
	)
}

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

var (
	commit, scope      string
	title, description string
	confirm            bool
)

// customKeyMap uses the default keymap, but overrides certain keys so it doesn't have to all be redefined.
func customKeyMap() *huh.KeyMap {
	df := huh.NewDefaultKeyMap()
	df.Quit = key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl", "exit"))
	df.Text.NewLine = key.NewBinding(key.WithKeys("enter", "ctrl+j"), key.WithHelp("enter / ctrl+j", "new line"))
	df.Text.Next = key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next"))
	df.Input.AcceptSuggestion = key.NewBinding(key.WithKeys("right"), key.WithHelp("→", "next"))
	return df
}

func gatherInput() (title, description, workitems string, draft bool) {
	if Debug {
		pterm.EnableDebugMessages()
	}
	var err error
	var commitType, scope string
	// var confirm bool
	// var okWithEmptyDescription bool

	// while this can return a collection, maxResults means return the most popular single scope in last evaluateScopeMonths period.
	// placeholderForScope := "lowercase-hypen-separated"

	//	TODO: make this prettier in future :-p, cause I want to do something with the list of scopes

	_, _ = pterm.DefaultInteractiveConfirm.WithDefaultText("pausing for warning output: press any key to continue").Show()
	nf := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("type").
				Value(&commitType).
				Description("press / to quickly filter by typing").
				Options(conventionalCommitTypeOptions()...),
			// WithHeight(5),
		),
	).WithKeyMap(customKeyMap()).
		WithTheme(huh.ThemeDracula())

	if err = nf.Run(); err != nil {
		pterm.Warning.Printfln("issue gather input, either by user cancellation, or other issue. I goofed. not your fault. ## ShouldHaveDoneTDD: %v", err)
		os.Exit(0)
	}
	var suggestedScope []ScopeCount

	suggestedScope, err = GetScopesInLastMonths(evaluateScopeMonths, 6, commitType)
	if err != nil {
		pterm.Warning.Printfln("suggested scope logic errored, continuing: %v", err)
		// placeholderForScope = ""
		_, _ = pterm.DefaultInteractiveConfirm.WithDefaultText("pausing for warning output: press any key to continue").Show()
	}
	var scopesToSuggest []string
	for _, s := range suggestedScope {
		scopesToSuggest = append(scopesToSuggest, s.Scope)
	}
	// var suggestionsForScope []string{}
	scopePlaceholder := DynamicScopeSuggestion(commitType)
	// if len(scopeSuggestions) == 0 {
	// 	pterm.Debug.Println("scope suggestions are just a single result, so using that value instead of scope array") // TODO: cleanup in future, it's redundant
	// 	suggestionsForScope = []string{scopePlaceholder}
	// }

	nf = huh.NewForm(
		huh.NewGroup(
			// TODO: have the suggestions populate the string array list
			huh.NewInput().Title("scope").Inline(true).Value(&scope).Suggestions(scopesToSuggest).Placeholder(scopePlaceholder).CharLimit(20),
			huh.NewInput().Title("title").Inline(true).Value(&title).Placeholder("use lower case, present tense").CharLimit(72),
			huh.NewText().
				Title("PR description").
				Placeholder("- Tell me more... ").
				Value(&description),
			// NOTE: it is confusing as it doesn't get inserted into current group, therefore disabling and letting user choose for now
			// related context: https://github.com/charmbracelet/huh/issues/108
			// Validate(func(t string) error {
			// 	if t == "" {
			// 		if err := huh.NewConfirm().
			// 			Title("Is it ok to proceed with an empty description?").
			// 			Value(&okWithEmptyDescription).Run(); err != nil {
			// 			return fmt.Errorf("NewConfirm for NewText Validate failure: %v", err)
			// 		}
			// 		if okWithEmptyDescription {
			// 			return nil
			// 		}
			// 		return fmt.Errorf("no input was provided, so try again")
			// 	}
			// 	return nil
			// }).Value(&description),
			huh.NewConfirm().Title("draft pr?").Inline(true).Value(&draft),
			huh.NewInput().Title("workitems").Inline(true).Value(&workitems).Placeholder("(optional) space separated"),
			// huh.NewConfirm().Title("submit?").Inline(true).Value(&confirm),
		),
	).
		WithKeyMap(customKeyMap()).
		WithTheme(huh.ThemeDracula())

	pterm.DefaultSection.Println("PR Creation")

	if err = nf.Run(); err != nil {
		pterm.Warning.Printfln("issue gather input, either by user cancellation, or other issue. I goofed. not your fault. ## ShouldHaveDoneTDD: %v", err)
		os.Exit(0)
	}

	if scope != "" {
		scope = "(" + scope + "):"
	} else {
		scope = ":"
	}
	title = fmt.Sprintf("%s%s %s%s", commitType, scope, emojify(commitType), title)
	pterm.Info.Println(title)

	pterm.Info.Printfln("\n%s", description)

	// if !confirm {
	// 	pterm.Warning.Printfln("you selected to not submit, so exiting without further action")
	// 	os.Exit(0)
	// }
	return title, description, workitems, draft
}

// trunk-ignore(golangci-lint/funlen)
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

func detectSSH() bool {
	if Debug {
		pterm.EnableDebugMessages()
	}
	// Open the repository
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		pterm.Error.Printfln("error opening repository: %v", err)
		os.Exit(1)
	}

	// Get the URL of the remote
	remote, err := repo.Remote("origin")
	if err != nil {
		pterm.Error.Printfln("error getting remote: %v", err)
		os.Exit(1)
	}
	url := remote.Config().URLs[0]

	// Check if the URL is an ssh URL
	_, err = transport.NewEndpoint(url)
	if err != nil {
		pterm.Debug.Println("The remote was cloned via https")
		return false
	} else {
		pterm.Debug.Println("The remote was cloned via ssh")
		return true
	}
}

func createPR() { //nolint:funlen,cyclop // this is a cli tool, not a library, ok with longer workflow command
	if Debug {
		pterm.EnableDebugMessages()
	}
	var branchName string
	var err error
	isSSH := detectSSH()

	// NOTE: can't use autodetect with ssh so have to be specific:
	// Per error: DevOps SSH URLs are not supported for repo auto-detection yet. https://github.com/Microsoft/azure-devops-cli-extension/issues/142
	// Attempt to calculate for user based on main/master pattern.
	if isSSH {
		branchName, err = getUpstreamBranch()
		if err != nil {
			pterm.Warning.Printfln("isSSH check for getUpstreamBranch: %v", err)
		}
	}

	title, description, workitems, draft := gatherInput()
	args := []string{
		"repos", "pr", "create",
		"--output", "json", // i had issues when passing at end, so make this the first arg
		"--title", title,
		"--draft", fmt.Sprintf("%t", draft),
		"--auto-complete", "true",
		"--delete-source-branch", "true",
		"--squash",
		"--transition-work-items", "true",
		"--open",
		"--merge-commit-message", title,
	}

	if isSSH && branchName != "" {
		pterm.Debug.Printfln("isSSH && branchName is: %s", branchName)
		args = append(args, "--target-branch")
		args = append(args, branchName)
	} else {
		pterm.Debug.Println("target-branch is not set so default branch set in Azure Repos will be used")
		args = append(args, "--detect")
	}

	args = append(args, "--description")
	args = append(args, description)

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

	pterm.Debug.Printfln("az %s", shellescape.QuoteCommand(args))
	out, err := cmd.CombinedOutput()
	if err != nil {
		pterm.Error.Printf("failure running azure-cli via az-cli:\n%v\n\n", err)
		pterm.Error.Printfln("out: %s", out)
		pterm.Error.Printfln("err: %v", err)

		pterm.Info.Println("try again by rerunning this generated command")
		pterm.Info.Println("az " + shellescape.QuoteCommand(args))
		os.Exit(1)
	}
	prResponse := PullRequestResponse{}
	if err := json.Unmarshal(out, &prResponse); err != nil {
		pterm.Error.Printf("unmarshal failure: %v\n", err)
		pterm.Debug.Printf("out:\n%s\n", string(out))
		os.Exit(1)
	}

	// to give better control when running in container, i want to output the url to the console to control click.
	url := fmt.Sprintf(
		"%s/pullrequest/%d",
		prResponse.Repository.WebURL,
		prResponse.PullRequestID,
	)
	pterm.Success.Printf("Pull Request Url: %s\n", url)

	// Try to match against a pr item number, and if so then append.
	// If not, bypass the entire process of trying to link to work-items.
	reg := regexp.MustCompile(`\d{5,7}`)
	if reg.MatchString(workitems) {
		associateWorkItemIDargs := []string{
			"repos", "pr", "work-item", "add",
			"--id", fmt.Sprintf("%d", prResponse.PullRequestID),
			"--work-items",
		}

		pterm.Success.Printf("Work Item IDs: %s\n", workitems)
		// Argument escaping seems to have issues with spaces in string
		// This as a workaround to turn the space delimited string into each being an individual argument to pass to command processor.
		itemIDs := strings.Split(workitems, " ")
		associateWorkItemIDargs = append(associateWorkItemIDargs, itemIDs...)
		pterm.Info.Println("az " + shellescape.QuoteCommand(associateWorkItemIDargs))

		if err := sh.Run("az", associateWorkItemIDargs...); err != nil {
			pterm.Error.Printf("failure associating work-items via az-cli:\n%v\n\n", err)

			os.Exit(0)
		}
	} else {
		pterm.Info.Println("no work items to associate")
	}

	_ = pterm.DefaultBigText.WithLetters(pterm.NewLettersFromString("CELEBRATE")).Render()
	if os.Getenv("NO_CONFETTI") == "1" {
		pterm.Debug.Printfln("no fun, no confetty, exiting")
		os.Exit(0)
	}

	time.Sleep(1 * time.Second)
}

type ScopeCount struct {
	Scope string
	Count int
}

type ByCount []ScopeCount

func (a ByCount) Len() int           { return len(a) }
func (a ByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCount) Less(i, j int) bool { return a[i].Count > a[j].Count }

func GetScopesInLastMonths(months, maxResults int, commitType string) ([]ScopeCount, error) {
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		pterm.Error.Printfln("error opening repository: %v", err)

		return nil, err
	}

	ref, err := r.Head()
	if err != nil {
		return nil, err
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	// Bulk review
	// var commitTypes []string
	// for _, i := range conventionalCommitTypeOptions() {
	// 	commitTypes = append(commitTypes, i.Value)
	// }
	// listOfCommitTypes := strings.Join(commitTypes, "|")
	scopeCounts := make(map[string]int)
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Author.When.After(time.Now().AddDate(0, -months, 0)) {
			re := regexp.MustCompile(fmt.Sprintf(`(?:%s)\((.*?)\):`, commitType))
			submatchall := re.FindAllStringSubmatch(c.Message, -1)
			for _, element := range submatchall {
				scopeCounts[element[1]]++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var scopes ByCount
	for k, v := range scopeCounts {
		scopes = append(scopes, ScopeCount{k, v})
	}

	sort.Sort(scopes)
	if len(scopes) > maxResults {
		scopes = scopes[:maxResults]
	}
	pterm.Debug.Printfln("total scope count: %d", len(scopes))
	pterm.Debug.Printfln("scope counts: %v", scopes)
	return scopes, nil
}

// DynamicScopeSuggestion calculates the best possible scope for the current commit type based on history.
func DynamicScopeSuggestion(commitType string) string {
	if commitType == "" {
		return ""
	}
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		pterm.Error.Printfln("error opening repository: %v", err)
		return ""
	}

	ref, err := r.Head()
	if err != nil {
		return ""
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return ""
	}
	scopeCounts := make(map[string]int)
	err = cIter.ForEach(func(c *object.Commit) error {
		if c.Author.When.After(time.Now().AddDate(0, -evaluateScopeMonths, 0)) {
			re := regexp.MustCompile(fmt.Sprintf(`(?:%s)\((.*?)\):`, commitType))
			submatchall := re.FindAllStringSubmatch(c.Message, -1)
			for _, element := range submatchall {
				scopeCounts[element[1]]++
			}
		}
		return nil
	})
	if err != nil {
		return ""
	}

	var scopes ByCount
	for k, v := range scopeCounts {
		scopes = append(scopes, ScopeCount{k, v})
	}

	sort.Sort(scopes)
	if len(scopes) > 1 {
		scopes = scopes[:1]
	}
	pterm.Debug.Printfln("total scope count: %d", len(scopes))
	pterm.Debug.Printfln("scope counts: %v", scopes)

	if len(scopes) > 0 {
		scopeRecommendation := scopes[0].Scope
		pterm.Debug.Printfln("best recommendation is: %s", scopeRecommendation)
		return scopeRecommendation
	}
	return ""
}
