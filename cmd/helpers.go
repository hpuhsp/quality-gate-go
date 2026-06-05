package cmd

import (
	"os"
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

// readFileSafe reads a file with size limit (1MB) using shared.SafeReadFile.
func readFileSafe(file string) ([]byte, error) {
	return shared.SafeReadFile(file)
}

// hasBin checks if a command exists in PATH.
func hasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
