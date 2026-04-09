// SPDX-License-Identifier: Apache-2.0

// Package docker implements Docker build and deployment support for Mendix projects.
package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// mxbuildBinaryName returns the platform-specific mxbuild binary name.
func mxbuildBinaryName() string {
	if runtime.GOOS == "windows" {
		return "mxbuild.exe"
	}
	return "mxbuild"
}

// mxbuildBinaryNames returns all candidate binary names for mxbuild.
// On Windows, the Linux binary ("mxbuild") may also be cached when downloaded
// for Docker, so both names must be checked.
func mxbuildBinaryNames() []string {
	if runtime.GOOS == "windows" {
		return []string{"mxbuild.exe", "mxbuild"}
	}
	return []string{"mxbuild"}
}

// findMxBuildInDir looks for the mxbuild binary inside a directory.
// Checks: dir/mxbuild, dir/modeler/mxbuild (Mendix installation layout).
func findMxBuildInDir(dir string) string {
	for _, bin := range mxbuildBinaryNames() {
		candidates := []string{
			filepath.Join(dir, bin),
			filepath.Join(dir, "modeler", bin),
		}
		for _, c := range candidates {
			if info, err := os.Stat(c); err == nil && !info.IsDir() {
				return c
			}
		}
	}
	return ""
}

// resolveMxBuild finds the MxBuild executable.
// Priority: explicit path > PATH lookup > OS-specific known locations.
// The explicit path can be the binary itself or a directory containing it
// (e.g., a Mendix installation root with modeler/mxbuild inside).
func resolveMxBuild(explicitPath string) (string, error) {
	if explicitPath != "" {
		info, err := os.Stat(explicitPath)
		if err != nil {
			return "", fmt.Errorf("mxbuild not found at %s: %w", explicitPath, err)
		}
		// If it's a directory, look for the binary inside it
		if info.IsDir() {
			if found := findMxBuildInDir(explicitPath); found != "" {
				return found, nil
			}
			return "", fmt.Errorf("mxbuild binary not found inside directory %s (looked for mxbuild and modeler/mxbuild)", explicitPath)
		}
		return explicitPath, nil
	}

	// Try PATH
	if p, err := exec.LookPath("mxbuild"); err == nil {
		return p, nil
	}

	// Try cached downloads (~/.mxcli/mxbuild/*/modeler/mxbuild)
	if p := AnyCachedMxBuildPath(); p != "" {
		return p, nil
	}

	// Try OS-specific known locations
	for _, pattern := range mxbuildSearchPaths() {
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			// Return the last match (likely newest version)
			return matches[len(matches)-1], nil
		}
	}

	return "", fmt.Errorf("mxbuild not found; install Mendix Studio Pro or specify --mxbuild-path")
}

// mxbuildSearchPaths returns OS-specific glob patterns for MxBuild.
func mxbuildSearchPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\Mendix\*\modeler\mxbuild.exe`,
			`C:\Program Files (x86)\Mendix\*\modeler\mxbuild.exe`,
		}
	case "darwin":
		return []string{
			"/Applications/Mendix/*/modeler/mxbuild",
		}
	default: // linux
		home, _ := os.UserHomeDir()
		paths := []string{
			"/opt/mendix/*/modeler/mxbuild",
		}
		if home != "" {
			paths = append(paths, filepath.Join(home, ".mendix/*/modeler/mxbuild"))
		}
		return paths
	}
}

// resolveJDK21 finds a JDK 21 installation.
// Priority: JAVA_HOME (verify version) > macOS java_home > java in PATH (verify version) > OS-specific known locations.
func resolveJDK21() (string, error) {
	// Try JAVA_HOME
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		if isJDK21(javaHome) {
			return javaHome, nil
		}
		// JAVA_HOME set but wrong version — continue searching
	}

	// On macOS, use /usr/libexec/java_home to find the real JDK path.
	// The /usr/bin/java shim is not a symlink, so EvalSymlinks won't resolve
	// it to the actual JDK, and MxBuild needs the real installation path.
	if runtime.GOOS == "darwin" {
		if javaHome, err := resolveMacOSJavaHome(); err == nil && isJDK21(javaHome) {
			return javaHome, nil
		}
	}

	// Try java in PATH
	if javaPath, err := exec.LookPath("java"); err == nil {
		// Resolve symlinks to find the actual installation
		resolved, err := filepath.EvalSymlinks(javaPath)
		if err == nil {
			javaPath = resolved
		}
		// java binary is typically at <jdk>/bin/java
		javaHome := filepath.Dir(filepath.Dir(javaPath))
		if isJDK21(javaHome) {
			return javaHome, nil
		}
	}

	// Try OS-specific known locations
	for _, pattern := range jdkSearchPaths() {
		matches, _ := filepath.Glob(pattern)
		for _, m := range matches {
			if isJDK21(m) {
				return m, nil
			}
		}
	}

	return "", fmt.Errorf("JDK 21 not found; set JAVA_HOME or install JDK 21")
}

// resolveMacOSJavaHome uses /usr/libexec/java_home to find a JDK 21 on macOS.
func resolveMacOSJavaHome() (string, error) {
	out, err := exec.Command("/usr/libexec/java_home", "-v", "21").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// jdkSearchPaths returns OS-specific glob patterns for JDK installations.
func jdkSearchPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{
			`C:\Program Files\Eclipse Adoptium\jdk-21*`,
			`C:\Program Files\Java\jdk-21*`,
			`C:\Program Files\Microsoft\jdk-21*`,
		}
	case "darwin":
		return []string{
			"/Library/Java/JavaVirtualMachines/temurin-21*/Contents/Home",
			"/Library/Java/JavaVirtualMachines/jdk-21*/Contents/Home",
		}
	default: // linux
		return []string{
			"/usr/lib/jvm/java-21-*",
			"/usr/lib/jvm/temurin-21-*",
		}
	}
}

// isJDK21 checks if the given JAVA_HOME points to a JDK 21 installation.
func isJDK21(javaHome string) bool {
	javaBin := filepath.Join(javaHome, "bin", "java")
	if runtime.GOOS == "windows" {
		javaBin += ".exe"
	}
	if _, err := os.Stat(javaBin); err != nil {
		return false
	}

	out, err := exec.Command(javaBin, "-version").CombinedOutput()
	if err != nil {
		return false
	}

	return jdk21VersionRegex.Match(out)
}

var jdk21VersionRegex = regexp.MustCompile(`version "21[\.\s"]`)

// javaVersionString runs java -version and returns the output for diagnostics.
func javaVersionString(javaHome string) string {
	javaBin := filepath.Join(javaHome, "bin", "java")
	out, err := exec.Command(javaBin, "-version").CombinedOutput()
	if err != nil {
		return "(unknown)"
	}
	lines := strings.SplitN(string(out), "\n", 2)
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}
	return "(unknown)"
}
