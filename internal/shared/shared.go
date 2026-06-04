// Package shared provides common utilities used across the quality-gate codebase.
package shared

import (
	"os"
	"os/exec"
)

// HasBin checks whether a command exists in PATH.
func HasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// FileExists checks whether a file or directory exists at the given path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
