package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func RunEnable() {
	// Verify in git repo
	if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
		fmt.Println("❌ Not in a git repository.")
		os.Exit(1)
	}

	home, _ := os.UserHomeDir()
	hooksDir := filepath.Join(home, ".quality-gate", "hooks")
	configDir := filepath.Join(home, ".quality-gate")

	os.MkdirAll(hooksDir, 0755)
	os.MkdirAll(configDir, 0755)

	// Write hook scripts using the current binary
	exe, _ := os.Executable()
	hook := func(name string) string {
		return fmt.Sprintf(`#!/bin/sh
# quality-gate %s hook
QG=$(command -v quality-gate 2>/dev/null || true)
if [ -n "$QG" ] && [ -x "$QG" ]; then exec "$QG" %s; fi
if [ -x "%s" ]; then exec "%s" %s; fi
echo "quality-gate: not found. Run: npm install -g github:hpuhsp/quality-gate"
exit 1
`, name, name, exe, exe, name)
	}

	os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(hook("pre-commit")), 0755)
	os.WriteFile(filepath.Join(hooksDir, "pre-push"), []byte(hook("pre-push")), 0755)

	exec.Command("git", "config", "core.hooksPath", hooksDir).Run()

	// Write default config if not exists
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
	fmt.Printf("   Gates: secret-scan · syntax-check · sql-injection · auto-format\n")
}

// RunDisable deactivates hooks.
func RunDisable() {
	exec.Command("git", "config", "--unset", "core.hooksPath").Run()
	fmt.Println("✅ quality-gate disabled for this repo")
}
