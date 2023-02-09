# AZ-PR

A go cli app which wraps up the azure-cli for simplifying setup and creation of azure-devops repos pull requests without having to resort to the web ui.

## Run

Use Go 1.18+.
This should grab the latest tagged version.
Upgrade the same way.

```shell
go install github.com/sheldonhull/az-pr@latest
```

## Why

I could do this in a shell script (which I've done), but wanted to make it easy for a Go dev to run with a nice prompting experience that was compatible with Linux/MacOS/Windows.
I use conventional commits in all repos, and this should make it much easier to create pull requests without dealing with the web-ui.

In addition, it will set auto-complete and link work-items.

## Why A Wrapper

The cli has a lot of logic already built in.
I couldn't find a documented Azure DevOps Go SDK for repos.

Might redo with API calls, but the cli is an easy quick win. 😀
