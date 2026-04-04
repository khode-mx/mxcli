# Setting Up

Before you can start exploring and modifying Mendix projects from the command line, you need three things:

1. **mxcli installed** on your machine
2. **A Mendix project** (an `.mpr` file) to work with
3. **A feel for the REPL**, the interactive shell where you'll spend most of your time

This chapter walks you through all three.

## Installation methods

There are five ways to get started with mxcli:

- **[Playground](installation.md#playground-zero-install)** -- open the [mxcli Playground](https://github.com/mendixlabs/mxcli-playground) in a GitHub Codespace. Zero install, runs in your browser with a sample Mendix project, tutorials, and example scripts. Best way to try mxcli for the first time.
- **`mxcli new`** (recommended for new projects) -- run `mxcli new MyApp --version 11.8.0` to create a new Mendix project from scratch with all tooling and Dev Container configured. One command does everything: downloads MxBuild, creates the project, sets up AI tools, and installs the correct mxcli binary.
- **Binary download** -- grab a pre-built binary from GitHub Releases. Quickest path if you want to use mxcli on your own project.
- **Build from source** -- clone the repo and run `make build`. Useful if you want the latest unreleased changes or plan to contribute.
- **Dev Container** (recommended for existing projects) -- run `mxcli init` on your Mendix project, open it in VS Code, and reopen in the container. This gives you mxcli, a JDK, Docker-in-Docker, and Claude Code all pre-configured in a sandboxed environment. This is the recommended approach, especially when pairing with AI coding assistants.

The next few pages cover each method, then walk you through opening a project and using the REPL.

> **Alpha software warning.** mxcli can corrupt your Mendix project files. Always work on a copy of your `.mpr` or use version control (Git) before making changes.
