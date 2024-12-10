package main

import (
	"fmt"
	"hash"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/magefile/mage/sh"
	"github.com/pterm/pterm"
	"github.com/sheldonhull/magetools/pkg/magetoolsutils"
)

// FNV64a hashes using fnv64a algorithm
//
// Sourced from: https://github.com/shomali11/util/blob/master/xhashes/xhashes.go
func FNV64a(text string) uint64 {
	algorithm := fnv.New64a()
	return uint64Hasher(algorithm, text)
}

// uint64Hasher returns a uint64
//
// Sourced from: https://github.com/shomali11/util/blob/master/xhashes/xhashes.go
func uint64Hasher(algorithm hash.Hash64, text string) uint64 {
	_, _ = algorithm.Write([]byte(text)) // TODO: look at if error handling here matters later
	return algorithm.Sum64()
}

func randomBuildName() (petname string) {
	v := time.Now().Unix()
	if err := gofakeit.Seed(v); err != nil {
		pterm.Error.Printfln("gofakeit.Seed() failed: %v", err)
		return "unknown"
	}
	animal := gofakeit.Animal()
	adjective := gofakeit.AdjectiveDescriptive()
	petname = strings.ToLower(strings.Join([]string{adjective, animal}, "-"))
	pterm.Info.Printfln("buildName based on Version: %s\n", petname)
	petname = strings.Trim(petname, "'")
	return petname
}

func randomBuildBasedOnVersion() (petname string) {
	releaseVersion, err := sh.Output("changie", "latest")
	if err != nil {
		pterm.Warning.Printfln("changie pulling latest release note version failure: %v", err)
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	// make cleanVersion reflect an int64 value so it can be used for seeding
	cleanVersionSeed := FNV64a(cleanVersion)

	if err := gofakeit.Seed(cleanVersionSeed); err != nil {
		pterm.Error.Printfln("gofakeit.Seed() failed: %v", err)
		return "unknown"
	}
	animal := gofakeit.Animal()
	adjective := gofakeit.AdjectiveDescriptive()
	petname = strings.ToLower(strings.Join([]string{adjective, animal}, "-"))
	pterm.Info.Printfln("random Pet Calculated at Runtime: %s\n", petname)
	petname = strings.Trim(petname, "'")
	return petname
}

func checkEnvVar(envVar string, required bool) (string, error) {
	envVarValue := os.Getenv(envVar)
	if envVarValue == "" && required {
		pterm.Error.Printfln(
			"%s is required and unable to proceed without this being provided. terminating task.",
			envVar,
		)
		return "", fmt.Errorf("%s is required", envVar)
	}
	if envVarValue == "" {
		pterm.Debug.Printfln(
			"checkEnvVar() found no value for: %q, however this is marked as optional, so not exiting task",
			envVar,
		)
	}
	pterm.Debug.Printfln("checkEnvVar() found value: %q=%q", envVar, envVarValue)
	return envVarValue, nil
}

// 🔨 Build builds the project for the current platform.
func Build() error {
	magetoolsutils.CheckPtermDebug()

	releaserArgs := []string{
		"build",
		"--clean",
		"--auto-snapshot",
		"--single-target",
		"--skip=validate",
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	return sh.RunWithV(
		map[string]string{
			"BUILD_NAME": randomBuildName(),
		}, "goreleaser", releaserArgs...) // "--skip-announce",.
}

// 🔨 BuildAll builds all the binaries defined in the project, for all platforms. This includes Docker image generation but skips publish.
// If there is no additional platforms configured in the task, then basically this will just be the same as `mage build`.
func BuildAll() error {
	magetoolsutils.CheckPtermDebug()
	releaserArgs := []string{
		"release",
		"--auto-snapshot",
		"--clean",
		"--skip=publish,validate",
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)
	return sh.RunWithV(
		map[string]string{
			"BUILD_NAME": randomBuildName(),
		}, "goreleaser", releaserArgs...)
	// To pass in explicit version mapping, you can do this. I'm not using at this time.
	// Return sh.RunWithV(map[string]string{
	// 	"GORELEASER_CURRENT_TAG": "latest",
	// }, binary, releaserArgs...)
}

// 🔨 Release generates a release for the current platform.
func Release() error {
	magetoolsutils.CheckPtermDebug()

	// if _, err := checkEnvVar("DOCKER_ORG", true); err != nil {
	// 	return err
	// }

	releaseVersion, err := sh.Output("changie", "latest")
	if err != nil {
		pterm.Warning.Printfln("changie pulling latest release note version failure: %v", err)
	}
	cleanVersion := strings.TrimSpace(releaseVersion)
	cleanpath := filepath.Join(".changes", cleanVersion+".md")

	if os.Getenv("GITHUB_WORKSPACE") != "" {
		cleanpath = filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".changes", cleanVersion+".md")
	}

	releaserArgs := []string{
		"release",
		"--clean",
		"--skip=validate",
		fmt.Sprintf("--release-notes=%s", cleanpath),
	}
	pterm.Debug.Printfln("goreleaser: %+v", releaserArgs)

	return sh.RunWithV(map[string]string{
		"GORELEASER_CURRENT_TAG": cleanVersion,
		"BUILD_NAME":             randomBuildBasedOnVersion(),
	},
		"goreleaser",
		releaserArgs...,
	)
}
