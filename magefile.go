//go:build mage
// +build mage

// ⚡ Core Mage Tasks
package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/ci"
	"github.com/sheldonhull/magetools/fancy"

	// mage:import
	_ "github.com/sheldonhull/magetools/docgen"
	// mage:import
	"github.com/sheldonhull/magetools/gotools"
)

// artifactDirectory is a directory containing artifacts for the project and shouldn't be committed to source.
const artifactDirectory = ".artifacts"

// cacheDirectory is a directory containing other artifacts for development that should persist between builds, such as temporary configs or testing charts.
const cacheDirectory = ".cache"

const permissionUserReadWriteExecute = 0o0777

// 📁 createDirectories creates the local working directories for build artifacts and tooling.
func createDirectories() error {
	for _, dir := range []string{artifactDirectory, cacheDirectory} {
		if err := os.MkdirAll(dir, permissionUserReadWriteExecute); err != nil {
			pterm.Error.Printf("failed to create dir: [%s] with error: %v\n", dir, err)

			return err
		}
		pterm.Success.Printf("✅ [%s] dir created\n", dir)
	}

	return nil
}

// ⚡ Init runs multiple tasks to initialize all the requirements for running a project for a new contributor.
func Init() {
	fancy.IntroScreen(ci.IsCI())
	pterm.Success.Println("running Init()...")

	mg.SerialDeps(
		Clean,
		createDirectories,
	)

	pterm.DefaultSection.Println("CI Tooling")
	if ci.IsCI() {
		pterm.Success.Println("done with CI specific tooling. since detected in CI context, ending init early as core requirements met")
		return
	}

	mg.SerialDeps(
		(gotools.Go{}.Tidy),
		(gotools.Go{}.Init),
	)

	pterm.DefaultSection.Println("Setup Project Specific Tools")
	// Aqua install is run in devcontainer/codespace automatically.
	// Might require setup outside of this project and then this will work.
	if err := sh.RunV("aqua", "install"); err != nil {
		pterm.Warning.Printfln("aqua install not successful.\n" +
			"This is optional, but will ensure every tool for the project is installed and matching version." +
			"To install see developer docs or go to https://aquaproj.github.io/docs/reference/install")
	}
}

// 🗑️ Clean up after yourself.
func Clean() {
	pterm.Success.Println("Cleaning...")
	for _, dir := range []string{artifactDirectory, cacheDirectory} {
		err := os.RemoveAll(dir)
		if err != nil {
			pterm.Error.Printf("failed to removeall: [%s] with error: %v\n", dir, err)
		}
		pterm.Success.Printf("🧹 [%s] dir removed\n", dir)
	}
	mg.Deps(createDirectories)
}

// Release using github cli (for now)
func Release() error {
	version, changelogFile, err := getVersion()
	if err != nil {
		pterm.Error.Printfln("failed to get version: %v", err)
		return err
	}
	return sh.Run("gh", "release", "create", version, "--title", version, "--notes-file", changelogFile, "--target", "main")
}

// getVersion returns the version and path for the changefile to use for the semver and release notes.
func getVersion() (releaseVersion, cleanPath string, err error) {
	releaseVersion, err = sh.Output("changie", "latest")
	if err != nil {
		pterm.Error.Printfln("changie pulling latest release note version failure: %v", err)
		return "", "", err
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanPath = filepath.Join(".changes", cleanVersion+".md")
	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanPath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}
	return cleanVersion, cleanPath, nil
}

// ⚙ Run builds the binary into the local artifact direction and launches for testing.
func Run() error {
	start := time.Now()
	defer func() {
		duration := durafmt.Parse(time.Since(start))
		pterm.Success.Printfln("✅ Run() took %s", duration)
	}()
	targetBuildFile := filepath.Join(artifactDirectory, "az-pr")
	if err := sh.RunV("go", "build", "-o", targetBuildFile, "main.go"); err != nil {
		return err
	}
	return sh.RunV(targetBuildFile, "shell", "--debug")
}

// 📦 Install the tool from remote.
// This can help catch odd issues with embed for example.
func Install() error {
	return sh.RunV("go", "install", "github.com/sheldonhull/az-pr@latest")
}
