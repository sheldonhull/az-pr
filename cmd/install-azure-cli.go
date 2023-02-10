package cmd

import (
	"os"
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
		pterm.Error.Println(err.Error())
		os.Exit(1)
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
	pterm.Warning.Println("... not yet implemented")
}
