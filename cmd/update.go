package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// RunUpdate checks for the latest release and shows update instructions.
// Self-updating a running binary is unreliable across platforms;
// instead we point the user to the correct install command.
func RunUpdate() {
	current := Version
	fmt.Printf("quality-gate v%s (%s/%s)\n", current, runtime.GOOS, runtime.GOARCH)

	latest, err := fetchLatestTag()
	if err != nil {
		fmt.Println("⚠️  Could not check latest version (network or rate-limit).")
		fmt.Println("   Visit: https://github.com/hpuhsp/quality-gate-go/releases")
		return
	}

	fmt.Printf("  Installed: v%s\n", current)
	fmt.Printf("  Latest:    v%s\n", latest)

	if current == latest {
		fmt.Println("✅ Already up to date.")
		return
	}

	fmt.Printf("\n📦 v%s is available. To update:\n", latest)
	switch runtime.GOOS {
	case "darwin", "linux":
		fmt.Printf("   curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/download/v%s/quality-gate-%s-%s -o /usr/local/bin/quality-gate && chmod +x /usr/local/bin/quality-gate\n", latest, runtime.GOOS, runtime.GOARCH)
	case "windows":
		fmt.Printf("   Invoke-WebRequest https://github.com/hpuhsp/quality-gate-go/releases/download/v%s/quality-gate-windows-amd64.exe -OutFile $env:LOCALAPPDATA\\quality-gate\\quality-gate.exe\n", latest)
	}
	fmt.Println()
	fmt.Println("   Or build from source: go install github.com/hpuhsp/quality-gate-go@latest")
}

func fetchLatestTag() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/hpuhsp/quality-gate-go/releases/latest")
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
	if len(release.TagName) > 1 && release.TagName[0] == 'v' {
		return release.TagName[1:], nil
	}
	return release.TagName, nil
}
