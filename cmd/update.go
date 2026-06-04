package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
	resp, err := http.Get("https://api.github.com/repos/hpuhsp/quality-gate-go/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	if release.TagName == "" {
		return "", fmt.Errorf("no tag_name in release")
	}
	// Strip "v" prefix: v2.0.0 → 2.0.0
	if len(release.TagName) > 1 && release.TagName[0] == 'v' {
		return release.TagName[1:], nil
	}
	return release.TagName, nil
}

func downloadAndReplace(version string) error {
	bin, _ := os.Executable()
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	// macOS should use amd64 binary (x86_64 emulation works on arm64)
	if goos == "darwin" {
		goarch = "amd64"
	}

	url := fmt.Sprintf("%s%s-%s-%s", updateURL, version, goos, goarch)

	// Download to temp, then replace
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

