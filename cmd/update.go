package cmd

import (
	"encoding/json"
	"fmt"
		"net/http"
	
	
		)

const updateURL = "https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-"

func RunUpdate() {
	current := Version
	fmt.Printf("quality-gate v%s\n", current)

	// Check latest version from GitHub releases
	latest, err := fetchLatestVersion()
	if err != nil {
		fmt.Println("⚠️  Could not fetch latest version (network/rate-limit). Falling back to direct update...")
		fmt.Println("Updating via go install...")
		if err := downloadAndReplace(""); err != nil {
			fmt.Printf("❌ Update failed: %v\n", err)
			return
		}
		fmt.Println("✅ Updated successfully")
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
		fmt.Println("   Download manually: https://github.com/hpuhsp/quality-gate-go/releases/latest")
		return
	}

	fmt.Printf("✅ Updated to v%s\n", latest)
}

func fetchLatestVersion() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/hpuhsp/quality-gate-go/tags")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil || len(tags) == 0 {
		return "", err
	}
	tagName := tags[0].Name
	if tagName == "" {
		return "", fmt.Errorf("no tag_name in release")
	}
	// Strip "v" prefix: v2.0.0 → 2.0.0
	if len(tagName) > 1 && tagName[0] == 'v' {
		return tagName[1:], nil
	}
	return tagName, nil
}

func downloadAndReplace(version string) error {
	return fmt.Errorf("Please run: git clone https://github.com/hpuhsp/quality-gate-go-go.git && cd quality-gate-go && go install .")
}