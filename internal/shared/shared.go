// Package shared provides common utilities used across the quality-gate codebase.
package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// SafeReadFile reads a file with path traversal protection and 1MB size limit.
func SafeReadFile(file string) ([]byte, error) {
	// Resolve to absolute path (EvalSymlinks resolves macOS /var → /private/var)
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	// Resolve symlinks so paths are comparable (critical on macOS)
	realAbs, _ := filepath.EvalSymlinks(abs)
	if realAbs != "" {
		abs = realAbs
	}
	clean := filepath.Clean(abs)

	// Get repo root and verify the file is within it
	repoRoot, rootErr := getRepoRoot()
	if rootErr == nil {
		// Also resolve symlinks in repo root
		realRoot, _ := filepath.EvalSymlinks(repoRoot)
		if realRoot != "" {
			repoRoot = realRoot
		}
		repoRoot = filepath.Clean(repoRoot)
		rel, err := filepath.Rel(repoRoot, clean)
		if err != nil || strings.HasPrefix(rel, "..") {
			return nil, fmt.Errorf("path traversal blocked: %s (outside repo)", file)
		}
	}

	// Skip files larger than 1MB
	if info, err := os.Stat(clean); err != nil || info.Size() > 1<<20 {
		return nil, fmt.Errorf("file too large or unreadable: %s", file)
	}
	return os.ReadFile(clean)
}

func getRepoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
