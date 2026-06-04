package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hpuhsp/quality-gate/internal/detect"
)

func RunEnable() {
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		fmt.Println("❌ Not in a git repository.")
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	hooksDir := filepath.Join(home, ".quality-gate", "hooks")
	configDir := filepath.Join(home, ".quality-gate")

	os.MkdirAll(hooksDir, 0755)
	os.MkdirAll(configDir, 0755)

	// Detect project type
	proj := detect.Detect(".")

	// Discover tool paths for GUI compatibility
	exe, _ := os.Executable()
	extraPath := discoverToolPaths(proj)

	hook := func(name string) string {
		return fmt.Sprintf(`#!/bin/sh
# quality-gate %s hook — installed by 'quality-gate enable'
# Augment PATH for GUI clients (SourceTree, IDE plugins)
export PATH="%s:$PATH"
# Try PATH lookup first, then direct binary path
QG=$(command -v quality-gate 2>/dev/null || true)
if [ -n "$QG" ] && [ -x "$QG" ]; then exec "$QG" %s; fi
if [ -x "%s" ]; then exec "%s" %s; fi
echo "quality-gate: not found. Run: npm install -g github:hpuhsp/quality-gate"
exit 1
`, name, extraPath, name, exe, exe, name)
	}

	os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(hook("pre-commit")), 0755)
	os.WriteFile(filepath.Join(hooksDir, "pre-push"), []byte(hook("pre-push")), 0755)

	exec.Command("git", "config", "core.hooksPath", hooksDir).Run()

	configPath := filepath.Join(configDir, "config.yml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.WriteFile(configPath, []byte(`# quality-gate configuration
minCoverage: 60
autoFormat: true
blockOnSecrets: true
`), 0644)
	}

	fmt.Printf("✅ quality-gate enabled\n")
	fmt.Printf("   Hooks: %s\n", hooksDir)
	fmt.Printf("   Config: %s\n", configPath)
	fmt.Printf("   Binary: %s (%s/%s)\n", exe, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("   Project: %s\n", proj.Language)
	fmt.Printf("   Gates: secret-scan · syntax-check · sql-injection · auto-format\n")
	fmt.Printf("   GUI-ready: SourceTree, IDE Git plugins ✅\n")
}

// discoverToolPaths finds paths to external tools needed by quality-gate.
// At enable time, we know exactly where these tools live, so we can bake
// their directories into the hook script's PATH for GUI compatibility.
func discoverToolPaths(proj detect.Result) string {
	dirs := make(map[string]bool)

	// Always: ~/.local/bin (quality-gate itself and user-installed tools)
	home, _ := os.UserHomeDir()
	addIfExists(dirs, filepath.Join(home, ".local", "bin"))
	addIfExists(dirs, "/usr/local/bin")

	// Java (needed by ktlint / google-java-format)
	if javaHome := findJavaHome(); javaHome != "" {
		addIfExists(dirs, filepath.Join(javaHome, "bin"))
	}

	// Node.js (for JS/TS syntax check)
	if nodePath, err := exec.LookPath("node"); err == nil {
		addIfExists(dirs, filepath.Dir(nodePath))
	}
	// npm (for prettier via npx)
	if npmPath, err := exec.LookPath("npm"); err == nil {
		addIfExists(dirs, filepath.Dir(npmPath))
	}

	// Go (for gofmt)
	if goPath, err := exec.LookPath("go"); err == nil {
		addIfExists(dirs, filepath.Dir(goPath))
	}

	// Language-specific
	switch proj.Language {
	case "kotlin":
		if ktlintPath, err := exec.LookPath("ktlint"); err == nil {
			addIfExists(dirs, filepath.Dir(ktlintPath))
		}
	case "java":
		if gjfPath, err := exec.LookPath("google-java-format"); err == nil {
			addIfExists(dirs, filepath.Dir(gjfPath))
		}
	case "javascript":
		if prettierPath, err := exec.LookPath("prettier"); err == nil {
			addIfExists(dirs, filepath.Dir(prettierPath))
		}
	}

	// Assemble into PATH string
	paths := make([]string, 0, len(dirs))
	for d := range dirs {
		paths = append(paths, d)
	}
	return strings.Join(paths, ":")
}

func addIfExists(dirs map[string]bool, path string) {
	if _, err := os.Stat(path); err == nil {
		dirs[path] = true
	}
}

func findJavaHome() string {
	// Check JAVA_HOME env
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		if _, err := os.Stat(filepath.Join(jh, "bin", "java")); err == nil {
			return jh
		}
	}
	// Check common locations
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, ".local", "java"),
		"/Library/Java/JavaVirtualMachines",
		"/usr/local/opt/openjdk/libexec",
	}
	for _, c := range candidates {
		matches, _ := filepath.Glob(filepath.Join(c, "jdk*", "Contents", "Home"))
		for _, m := range matches {
			if _, err := os.Stat(filepath.Join(m, "bin", "java")); err == nil {
				return m
			}
		}
	}
	return ""
}

func RunDisable() {
	exec.Command("git", "config", "--unset", "core.hooksPath").Run()
	fmt.Println("✅ quality-gate disabled for this repo")
}
