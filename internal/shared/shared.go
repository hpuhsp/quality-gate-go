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

// RunInstall executes a shell command and returns combined output on failure.
// WARNING: This function passes the command string to "sh -c". Only call it with
// hardcoded, trusted command strings. Never pass user input to this function.
func RunInstall(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.TrimSpace(string(out)))
	}
	return nil
}

// repoRoot resolves the git repo root for the current working directory.
// It is NOT cached — each call queries git. This avoids stale results when
// the working directory changes (e.g., submodules, chained hooks).
func repoRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	val := strings.TrimSpace(string(out))
	// Resolve symlinks (critical on macOS /var → /private/var)
	if real, err := filepath.EvalSymlinks(val); err == nil && real != "" {
		val = real
	}
	return filepath.Clean(val), nil
}

// SafeReadFile reads a file with path traversal protection and 1MB size limit.
func SafeReadFile(file string) ([]byte, error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	realAbs, err := filepath.EvalSymlinks(abs)
	if err == nil && realAbs != "" {
		abs = realAbs
	}
	clean := filepath.Clean(abs)

	// Get repo root (not cached — re-resolved each call for correctness)
	gitRoot, rootErr := repoRoot()
	if rootErr != nil {
		return nil, fmt.Errorf("cannot verify path safety: not in a git repo (%v)", rootErr)
	}

	rel, err := filepath.Rel(gitRoot, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("path traversal blocked: %s (outside repo)", file)
	}

	// Read and validate size atomically via the file content
	data, err := os.ReadFile(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file, err)
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("file too large: %s (%d bytes)", file, len(data))
	}
	return data, nil
}
