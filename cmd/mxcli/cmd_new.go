// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/mendixlabs/mxcli/cmd/mxcli/docker"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <app-name>",
	Short: "Create a new Mendix project",
	Long: `Create a new Mendix project with all tooling configured.

This command performs the following steps:
  1. Downloads MxBuild for the specified Mendix version
  2. Creates a blank Mendix project using mx create-project
  3. Initializes AI tooling and devcontainer configuration (mxcli init)
  4. Downloads the correct mxcli binary for the devcontainer (linux)

Examples:
  mxcli new MyApp
  mxcli new MyApp --version 11.8.0
  mxcli new MyApp --version 10.24.0 --output-dir ./projects/my-app
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		mendixVersion, _ := cmd.Flags().GetString("version")
		outputDir, _ := cmd.Flags().GetString("output-dir")
		skipInit, _ := cmd.Flags().GetBool("skip-init")

		if mendixVersion == "" {
			fmt.Fprintln(os.Stderr, "Error: --version is required (e.g., --version 11.8.0)")
			os.Exit(1)
		}

		// Resolve output directory
		if outputDir == "" {
			outputDir = appName
		}
		absDir, err := filepath.Abs(outputDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		// Check if directory already exists and has content
		if entries, err := os.ReadDir(absDir); err == nil && len(entries) > 0 {
			fmt.Fprintf(os.Stderr, "Error: directory %s already exists and is not empty\n", absDir)
			os.Exit(1)
		}

		// Step 1: Download MxBuild
		fmt.Printf("Step 1/4: Downloading MxBuild %s...\n", mendixVersion)
		mxbuildPath, err := docker.DownloadMxBuild(mendixVersion, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading MxBuild: %v\n", err)
			os.Exit(1)
		}

		// Resolve mx binary from mxbuild path
		mxPath, err := docker.ResolveMx(mxbuildPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: could not find mx binary: %v\n", err)
			os.Exit(1)
		}

		// Step 2: Create project
		fmt.Printf("\nStep 2/4: Creating Mendix project '%s'...\n", appName)
		if err := os.MkdirAll(absDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			os.Exit(1)
		}

		mxCmd := exec.Command(mxPath, "create-project", "--app-name", appName)
		mxCmd.Dir = absDir
		mxCmd.Stdout = os.Stdout
		mxCmd.Stderr = os.Stderr
		if err := mxCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			os.Exit(1)
		}

		// Verify .mpr was created
		mprPath := filepath.Join(absDir, "App.mpr")
		if _, err := os.Stat(mprPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: mx create-project did not produce App.mpr in %s\n", absDir)
			os.Exit(1)
		}
		fmt.Printf("  Created %s\n", mprPath)

		// Step 3: Initialize tooling
		if !skipInit {
			fmt.Printf("\nStep 3/4: Initializing AI tooling...\n")
			initCmd.Run(initCmd, []string{absDir})
		} else {
			fmt.Printf("\nStep 3/4: Skipped (--skip-init)\n")
		}

		// Step 4: Ensure correct mxcli binary for devcontainer
		fmt.Printf("\nStep 4/4: Setting up mxcli binary...\n")
		mxcliBinPath := filepath.Join(absDir, "mxcli")
		if runtime.GOOS != "linux" {
			// Running on Windows/macOS — download the Linux binary for devcontainer
			tag := mxcliReleaseTag()
			fmt.Printf("  Downloading Linux mxcli (%s) for devcontainer...\n", tag)
			if err := downloadMxcliBinary("mendixlabs/mxcli", tag, "linux", "amd64", mxcliBinPath, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not download Linux binary: %v\n", err)
				fmt.Fprintln(os.Stderr, "  Run 'mxcli setup mxcli --output ./mxcli' inside the project to fix this.")
			}
		} else {
			// Running on Linux — copy ourselves
			self, err := os.Executable()
			if err == nil {
				selfBytes, err := os.ReadFile(self)
				if err == nil {
					if err := os.WriteFile(mxcliBinPath, selfBytes, 0755); err != nil {
						fmt.Fprintf(os.Stderr, "  Warning: could not copy mxcli binary: %v\n", err)
					} else {
						fmt.Printf("  Copied mxcli to %s\n", mxcliBinPath)
					}
				}
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: could not copy mxcli binary: %v\n", err)
			}
		}

		fmt.Printf("\n✓ Project '%s' created at %s\n", appName, absDir)
		fmt.Println("\nNext steps:")
		fmt.Println("  1. Open the project folder in VS Code")
		fmt.Println("  2. Reopen in Dev Container when prompted")
		fmt.Printf("  3. Run './mxcli -p App.mpr' to start working\n")
	},
}

func init() {
	newCmd.Flags().String("version", "", "Mendix version (e.g., 11.8.0) — required")
	newCmd.Flags().String("output-dir", "", "Output directory (default: ./<app-name>)")
	newCmd.Flags().Bool("skip-init", false, "Skip AI tooling initialization (mxcli init)")

	rootCmd.AddCommand(newCmd)
}
