# AZ-PR

![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/0-percent-optimized.svg)
![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/contains-tasty-spaghetti-code.svg)
![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/not-a-bug-a-feature.svg)
![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/uses-badges.svg)
![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/works-on-my-machine.svg)
![forthebadge](https://raw.githubusercontent.com/BraveUX/for-the-badge/master/src/images/badges/you-didnt-ask-for-this.svg)

A go cli app which wraps up the azure-cli for simplifying setup and creation of azure-devops repos pull requests without having to resort to the web ui.

## Disclaimer

It's a wrapper and done with limited time so lots to cleanup, but since it's replacing a bash script it's still better than nothing! 😀

Feel free to create a PR or an issue for any issues experienced and if I can improve it I will.
Hope it helps you out in your workflow a bit and helps you avoid the Azure DevOps web ui a litle bit longer.

## Install

Use Go 1.18+.
This should grab the latest tagged version.
Upgrade the same way.

```shell
go install github.com/sheldonhull/az-pr@latest
```

## Check Requirements

You need to have an access token configured, so this will check and make sure the requirements to run are setup correctly.

```shell
az-pr check
```

## Install from Source With Go

If you don't have Go tools typically setup, then here's something to add your profile to make sure the binary is found.
This would be your `$HOME/.zshenv` or `$HOME/.bashrc`.

```shell
export PATH="$(go env GOPATH)/bin:${PATH}"
```

Find this file by typing `code $PROFILE` in a PowerShell prompt to open/create it on demand.
This profile is different depending on where you load it from, so VSCode has a unique profile as well as a normal terminal.
[Profile Files](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_profiles?view=powershell-7.3)

```powershell
$GOPATH=&go env GOPATH
$ENV:PATH = '{0}{1}{2}' -f (Join-Path $GOPATH 'bin'), [IO.Path]::PathSeparator, $ENV:PATH
```

## Why

I could do this in a shell script (which I've done), but wanted to make it easy for a Go dev to run with a nice prompting experience that was compatible with Linux/MacOS/Windows.
I use conventional commits in all repos, and this should make it much easier to create pull requests without dealing with the web-ui.

In addition, it will set auto-complete and link work-items.

## Why A Wrapper

The cli has a lot of logic already built in.
I couldn't find a documented Azure DevOps Go SDK for repos.

Might redo with API calls, but the cli is an easy quick win. 😀
