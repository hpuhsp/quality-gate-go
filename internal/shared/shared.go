// Package shared provides common utilities used across the quality-gate codebase.
package shared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
func RunInstall(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), strings.TrimSpace(string(out)))
	}
	return nil
}

// Cached repo root — resolved once per process, never again.
var (
	repoRootOnce sync.Once
	repoRootVal  string
	repoRootErr  error
)

func cachedRepoRoot() (string, error) {
	repoRootOnce.Do(func() {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			repoRootErr = err
			return
		}
		repoRootVal = strings.TrimSpace(string(out))
		// Resolve symlinks (critical on macOS /var → /private/var)
		real, _ := filepath.EvalSymlinks(repoRootVal)
		if real != "" {
			repoRootVal = real
		}
		repoRootVal = filepath.Clean(repoRootVal)
	})
	return repoRootVal, repoRootErr
}

// SafeReadFile reads a file with path traversal protection and 1MB size limit.
func SafeReadFile(file string) ([]byte, error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	realAbs, _ := filepath.EvalSymlinks(abs)
	if realAbs != "" {
		abs = realAbs
	}
	clean := filepath.Clean(abs)

	// Get repo root (cached)
	repoRoot, rootErr := cachedRepoRoot()
	if rootErr != nil {
		return nil, fmt.Errorf("cannot verify path safety: not in a git repo (%v)", rootErr)
	}

	rel, err := filepath.Rel(repoRoot, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("path traversal blocked: %s (outside repo)", file)
	}

	// Skip files larger than 1MB
	if info, err := os.Stat(clean); err != nil || info.Size() > 1<<20 {
		return nil, fmt.Errorf("file too large or unreadable: %s", file)
	}
	return os.ReadFile(clean)
}
