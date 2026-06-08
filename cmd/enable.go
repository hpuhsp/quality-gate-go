package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

func RunEnable() {
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		fmt.Println("❌ Not in a git repository.")
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".quality-gate")
	hooksDir := filepath.Join(configDir, "hooks")

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		fmt.Printf("❌ Cannot create hooks directory: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Printf("❌ Cannot create config directory: %v\n", err)
		os.Exit(1)
	}

	proj := detect.Detect(".")
	exe, _ := os.Executable()
	extraPath := discoverToolPaths(proj)

	// Generate hook scripts (shell scripts work on all platforms —
	// Git for Windows includes bash/sh that executes hooks).
	hook := func(name string) string {
		return fmt.Sprintf(`#!/bin/sh
# quality-gate %s hook — installed by 'quality-gate enable'
export PATH="%s:$PATH"
if QG=$(command -v quality-gate 2>/dev/null) && [ -n "$QG" ] && [ -x "$QG" ]; then
  exec "$QG" %s
fi
if [ -x "%s" ]; then exec "%s" %s; fi
echo "quality-gate: not found. Download from https://github.com/hpuhsp/quality-gate-go/releases"
exit 1
`, name, extraPath, name, exe, exe, name)
	}

	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(hook("pre-commit")), 0755); err != nil {
		fmt.Printf("❌ Failed to write pre-commit hook: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-push"), []byte(hook("pre-push")), 0755); err != nil {
		fmt.Printf("❌ Failed to write pre-push hook: %v\n", err)
		os.Exit(1)
	}

	// Windows: also write .bat wrappers for native cmd/PowerShell git clients
	if runtime.GOOS == "windows" {
		bat := func(name string) string {
			return fmt.Sprintf(`@echo off
REM quality-gate %s hook (Windows)
set "PATH=%s;%%PATH%%"
"%s" %s
`, name, strings.ReplaceAll(extraPath, ":", ";"), exe, name)
		}
		if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit.bat"), []byte(bat("pre-commit")), 0644); err != nil {
			fmt.Printf("❌ Failed to write pre-commit.bat: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(hooksDir, "pre-push.bat"), []byte(bat("pre-push")), 0644); err != nil {
			fmt.Printf("❌ Failed to write pre-push.bat: %v\n", err)
			os.Exit(1)
		}
	}

	if err := exec.Command("git", "config", "core.hooksPath", hooksDir).Run(); err != nil {
		fmt.Printf("⚠️  Failed to set git config: %v\n", err)
	}

	fmt.Printf("✅ quality-gate enabled\n")
	fmt.Printf("   Hooks:  %s\n", hooksDir)
	fmt.Printf("   Binary: %s (%s/%s)\n", exe, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("   Project: %s\n", proj.Language)
	fmt.Printf("   Gates: secret-scan · syntax-check · security-scan · auto-format · lint · architecture\n")
	if runtime.GOOS == "windows" {
		fmt.Printf("   Windows: .sh + .bat hooks generated\n")
	} else {
		fmt.Printf("   GUI-ready: SourceTree, IDE Git plugins ✅\n")
	}
}

func discoverToolPaths(proj detect.Result) string {
	dirs := make(map[string]bool)
	home, _ := os.UserHomeDir()

	if runtime.GOOS == "windows" {
		// Windows tool paths
		addIfExists(dirs, filepath.Join(home, ".local", "bin"))
		addIfExists(dirs, filepath.Join(home, "AppData", "Roaming", "npm"))
		addIfExists(dirs, filepath.Join(home, "go", "bin"))
		addIfExists(dirs, `C:\Program Files\nodejs`)
		addIfExists(dirs, `C:\Program Files\Go\bin`)
		// Java
		if jh := findJavaHome(); jh != "" {
			addIfExists(dirs, filepath.Join(jh, "bin"))
		}
		// ktlint / prettier via npm
		if ktlintPath, err := exec.LookPath("ktlint"); err == nil {
			addIfExists(dirs, filepath.Dir(ktlintPath))
		}
		if prettierPath, err := exec.LookPath("prettier"); err == nil {
			addIfExists(dirs, filepath.Dir(prettierPath))
		}
	} else {
		// Unix tool paths
		addIfExists(dirs, filepath.Join(home, ".local", "bin"))
		addIfExists(dirs, "/usr/local/bin")
		if javaHome := findJavaHome(); javaHome != "" {
			addIfExists(dirs, filepath.Join(javaHome, "bin"))
		}
		if nodePath, err := exec.LookPath("node"); err == nil {
			addIfExists(dirs, filepath.Dir(nodePath))
		}
		if npmPath, err := exec.LookPath("npm"); err == nil {
			addIfExists(dirs, filepath.Dir(npmPath))
		}
		if goPath, err := exec.LookPath("go"); err == nil {
			addIfExists(dirs, filepath.Dir(goPath))
		}
	}

	// Language-specific tools (cross-platform)
	switch proj.Language {
	case "kotlin":
		if p, _ := exec.LookPath("ktlint"); p != "" {
			addIfExists(dirs, filepath.Dir(p))
		}
	case "java":
		if p, _ := exec.LookPath("google-java-format"); p != "" {
			addIfExists(dirs, filepath.Dir(p))
		}
	case "javascript":
		if p, _ := exec.LookPath("prettier"); p != "" {
			addIfExists(dirs, filepath.Dir(p))
		}
	}

	paths := make([]string, 0, len(dirs))
	for d := range dirs {
		paths = append(paths, d)
	}
	// Use platform-appropriate separator (bash on all platforms for hooks)
	return strings.Join(paths, ":")
}

func addIfExists(dirs map[string]bool, path string) {
	if _, err := os.Stat(path); err == nil {
		dirs[path] = true
	}
}

func findJavaHome() string {
	if jh := os.Getenv("JAVA_HOME"); jh != "" {
		if _, err := os.Stat(filepath.Join(jh, "bin", "java")); err == nil {
			return jh
		}
		if _, err := os.Stat(filepath.Join(jh, "bin", "java.exe")); err == nil {
			return jh
		}
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		candidates := []string{
			filepath.Join(home, ".local", "java"),
			`C:\Program Files\Java`,
			`C:\Program Files\Eclipse Adoptium`,
			`C:\Program Files\Amazon Corretto`,
		}
		for _, c := range candidates {
			matches, _ := filepath.Glob(filepath.Join(c, "jdk*"))
			for _, m := range matches {
				if _, err := os.Stat(filepath.Join(m, "bin", "java.exe")); err == nil {
					return m
				}
			}
		}
	} else {
		candidates := []string{
			filepath.Join(home, ".local", "java"),
			"/Library/Java/JavaVirtualMachines",
			"/usr/local/opt/openjdk/libexec",
			"/usr/lib/jvm",
		}
		for _, c := range candidates {
			matches, _ := filepath.Glob(filepath.Join(c, "jdk*", "Contents", "Home"))
			for _, m := range matches {
				if _, err := os.Stat(filepath.Join(m, "bin", "java")); err == nil {
					return m
				}
			}
			// Linux-style: /usr/lib/jvm/jdk-17
			matches2, _ := filepath.Glob(filepath.Join(c, "jdk*"))
			for _, m := range matches2 {
				if _, err := os.Stat(filepath.Join(m, "bin", "java")); err == nil {
					return m
				}
			}
		}
	}
	return ""
}

func RunDisable() {
	if err := exec.Command("git", "config", "--unset", "core.hooksPath").Run(); err != nil {
		fmt.Printf("⚠️  quality-gate may not be enabled for this repo (%v)\n", err)
		return
	}
	fmt.Println("✅ quality-gate disabled for this repo")
}
