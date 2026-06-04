package cmd

import (
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

func RunStatus() {
	proj := detect.Detect(".")

	// Git hook status
	hooksPath := "not set"
	if out, err := exec.Command("git", "config", "core.hooksPath").Output(); err == nil {
		hooksPath = string(out)
		if len(hooksPath) > 0 {
			hooksPath = hooksPath[:len(hooksPath)-1] // trim newline
		}
	}

	home, _ := os.UserHomeDir()
	qgHooks := filepath.Join(home, ".quality-gate", "hooks")
	enabled := hooksPath == qgHooks

	preCommitOk := shared.FileExists(filepath.Join(qgHooks, "pre-commit"))
	prePushOk := shared.FileExists(filepath.Join(qgHooks, "pre-push"))

	fmt.Println("quality-gate status")
	fmt.Println("──────────────────")
	fmt.Printf("  Enabled:     %s\n", boolIcon(enabled))
	fmt.Printf("  Hooks path:  %s\n", hooksPath)
	fmt.Printf("  pre-commit:  %s\n", fileIcon(preCommitOk))
	fmt.Printf("  pre-push:    %s\n", fileIcon(prePushOk))
	fmt.Println()
	fmt.Printf("  Project:     %s\n", proj.Language)
	fmt.Printf("  Build tool:  %s\n", proj.BuildTool)
	fmt.Printf("  Test FW:     %s\n", proj.TestFramework)
	fmt.Printf("  Has tests:   %s\n", boolIcon(proj.HasTests))
	fmt.Println()
	fmt.Printf("  Config:      %s\n", filepath.Join(home, ".quality-gate", "config.yml"))
}

func boolIcon(v bool) string {
	if v {
		return "✅ yes"
	}
	return "❌ no"
}

func fileIcon(v bool) string {
	if v {
		return "✅ installed"
	}
	return "❌ missing"
}

