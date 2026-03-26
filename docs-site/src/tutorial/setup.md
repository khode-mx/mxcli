# Setting Up

Before you can start exploring and modifying Mendix projects from the command line, you need three things:

1. **mxcli installed** on your machine
2. **A Mendix project** (an `.mpr` file) to work with
3. **A feel for the REPL**, the interactive shell where you'll spend most of your time

This chapter walks you through all three.

## Installation methods

There are three ways to get mxcli running:

- **Binary download** -- grab a pre-built binary from GitHub Releases. Quickest path if you just want to try it out.
- **Build from source** -- clone the repo and run `make build`. Useful if you want the latest unreleased changes or plan to contribute.
- **Dev Container** (recommended) -- run `mxcli init` on your Mendix project, open it in VS Code, and reopen in the container. This gives you mxcli, a JDK, Docker-in-Docker, and Claude Code all pre-configured in a sandboxed environment. This is the recommended approach, especially when pairing with AI coding assistants.

The next few pages cover each method, then walk you through opening a project and using the REPL.

> **Alpha software warning.** mxcli can corrupt your Mendix project files. Always work on a copy of your `.mpr` or use version control (Git) before making changes.
