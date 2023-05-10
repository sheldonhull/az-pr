//nolint:gochecknoglobals // cobra uses globals for commands
package cmd

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
	"github.com/bitfield/script"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// installAzureCliCmd represents the installAzureCli command
var installAzureCliCmd = &cobra.Command{
	Use:   "install-azure-cli",
	Short: "install-azcli",
	Long: `Attempts to install azure-cli based on the current runtime.

While this won't install the package managers, it will try to detect the following to help install:

- MacOS: Homebrew
- Linux: Snap, Homebrew, Curl Install
- Windows: Scoop & Chocolatey
`,
	Run: func(cmd *cobra.Command, args []string) {
		installAzureCLI()
	},
}

func init() {
	rootCmd.AddCommand(installAzureCliCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installAzureCliCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installAzureCliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func installAzureCLI() {
	// check for az command to be found and not error
	binary, err := exec.LookPath("az")
	if err == nil {
		pterm.Info.Printfln("az already installed, found: %q", binary)
		return
	}

	switch runtime.GOOS {
	case "darwin":
		installAzureCLIMacOS()
	case "linux":
		installAzureCLILinux()
	case "windows":
		installAzureCLIWindows()
	default:
		pterm.Error.Printfln("Unsupported OS: %s", runtime.GOOS)
	}
}

func installAzureCLIMacOS() {
	pterm.DefaultSection.Printfln("installing azure-cli via brew")
	_, _ = script.Exec("brew install azure-cli").Stdout()
}

func installAzureCLILinux() {
	pterm.Info.Println("not using apt-install as it can be very out of date")
	var response bool
	qs := &survey.Confirm{
		Message: "Would you like to install or upgrade azure-cli via curl (debian/ubuntu only)?",
		Default: false,
	}

	err := survey.AskOne(qs, &response)
	if err != nil {
		pterm.Warning.Printfln("installAzureCLILinux() you changed your mind: %v", err)
		os.Exit(0)
	}

	if !response {
		pterm.Info.Println("Skipping install of azure-cli")
		os.Exit(0)
		return
	}

	pterm.Debug.Println("exec.LookPath(\"az\") failed")
	pterm.Debug.Println("os.IsNotExist(err) was true, so attempting to start install")
	pterm.Warning.Println("Attempting to install missing azure-cli for you")
	_, _ = script.Get("https://aka.ms/InstallAzureCLIDeb").Exec("sudo bash").Stdout()

	// cmd := exec.Command("az", "config", "set", "extension.use_dynamic_install=yes_prompt")
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// if err := cmd.Run(); err != nil {
	// 	pterm.Error.Printfln("init() az config set extension: %v", err)
	// }
	pterm.Success.Println("init(): az config set extension.use_dynamic_install=yes_prompt")
	pterm.Success.Println("azure-cli installed successfully")
}

func installAzureCLIWindows() {
	installType, _ := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"scoop", "choco", "winget"}).
		Show()
	pterm.Info.Printfln("attempting install by %s", installType)

	if installType == "scoop" {
		// if scoop is found in path then use this to install azure-cli
		_, err := exec.LookPath("scoop")
		if err == nil {
			pterm.DefaultSection.Printfln("installing azure-cli via scoop")
			_, _ = script.Exec("scoop install azure-cli").Stdout()
			return
		} else {
			pterm.Warning.Printfln("scoop not found in path: %v", err)
		}
	}

	if installType == "choco" {

		// if choco is found then use this to install azure-cli
		_, err := exec.LookPath("choco")
		if err == nil {
			pterm.DefaultSection.Printfln("installing azure-cli via choco")
			_, _ = script.Exec("choco install azure-cli").Stdout()
			return
		} else {
			pterm.Warning.Printfln("choco not found in path: %v", err)
		}
	}

	// if winget is found then use this to install azure-cli
	if installType == "winget" {
		_, err := exec.LookPath("winget")
		if err == nil {
			pterm.DefaultSection.Printfln("installing azure-cli via winget")
			_, _ = script.Exec("winget install -e --id Microsoft.AzureCLI").Stdout()
			return
		} else {
			pterm.Warning.Printfln("winget not found in path: %v", err)
		}
	}

	// else warn can't to go install manually.
}
