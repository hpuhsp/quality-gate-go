package cmd

import (
	"os/exec"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// runCmd runs a shell command and returns stdout.
func runCmd(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// splitLines splits a string by newlines.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

// hasBin delegates to shared.HasBin (avoids duplication).
func hasBin(name string) bool {
	return shared.HasBin(name)
}
