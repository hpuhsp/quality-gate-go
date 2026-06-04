package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const updateURL = "https://github.com/hpuhsp/quality-gate/releases/latest/download/quality-gate-"

func RunUpdate() {
	current := Version
	fmt.Printf("quality-gate v%s\n", current)

	// Check latest version from GitHub releases
	latest, err := fetchLatestVersion()
	if err != nil {
		fmt.Println("⚠️  Could not check latest version.")
		return
	}

	fmt.Printf("  Installed: v%s\n", current)
	fmt.Printf("  Latest:    v%s\n", latest)

	if current == latest {
		fmt.Println("✅ Already up to date.")
		return
	}

	fmt.Printf("\nUpdating v%s → v%s...\n", current, latest)
	if err := downloadAndReplace(latest); err != nil {
		fmt.Printf("❌ Update failed: %v\n", err)
		fmt.Println("   Download manually: https://github.com/hpuhsp/quality-gate/releases/latest")
		return
	}

	fmt.Printf("✅ Updated to v%s\n", latest)
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/hpuhsp/quality-gate/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	s := string(body)
	// Extract tag_name
	idx := strings.Index(s, `"tag_name":"`)
	if idx < 0 {
		return "", fmt.Errorf("no tag_name")
	}
	rest := s[idx+len(`"tag_name":"`):]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return "", fmt.Errorf("malformed")
	}
	return strings.TrimPrefix(rest[:end], "v"), nil
}

func downloadAndReplace(version string) error {
	bin, _ := os.Executable()
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	url := fmt.Sprintf("%s%s-%s-%s", updateURL, version, goos, goarch)
	if goos == "darwin" {
		goarch = "amd64"
	}
	_ = url

	// Simple approach: re-download binary
	tmp := filepath.Join(os.TempDir(), "quality-gate-update")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, data, 0755); err != nil {
		return err
	}

	// Replace current binary
	// On macOS/Linux: move tmp → bin
	return os.Rename(tmp, bin)
}

// RunPrePush placeholder — CI handles real tests.
func RunPrePush() {
	fmt.Println("⚠️  Unit tests are handled by CI, not local pre-push.")
	fmt.Println("   Pre-push hook is active but skips tests by default.")
	fmt.Println("   Run 'quality-gate setup' to enable local test execution.")
}

// RunGenTests placeholder for AI test generation.
func RunGenTests(args ...string) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY not set.")
		fmt.Println("   export ANTHROPIC_API_KEY=sk-ant-...")
		os.Exit(1)
	}
	target := ""
	if len(args) > 2 {
		target = args[len(args)-1]
	}
	if target != "" {
		fmt.Printf("Generating tests for %s...\n", target)
	} else {
		fmt.Println("Generating tests for changed code...")
	}
	fmt.Println("⚠️  AI test generation requires gstack CLI or claude CLI.")
	fmt.Println("   Install: npm install -g gstack")
}
