package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunSetup() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".quality-gate")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yml")

	fmt.Print(`
╔══════════════════════════════════════════╗
║   quality-gate — First Run Setup         ║
╚══════════════════════════════════════════╝

This wizard configures your quality-gate settings.
Press Enter to accept defaults.
`)

	// API Key
	fmt.Println("─── AI Test Generation ───")
	apiKey := prompt("Anthropic API Key (sk-ant-...)", "")

	// Tests on push
	fmt.Println("\n─── Test Execution ───")
	pushTests := prompt("Run unit tests on git push? (yes/no)", "no")

	// Auto format
	fmt.Println("\n─── Code Formatting ───")
	autoFmt := prompt("Auto-format code on commit? (yes/no)", "yes")

	// Coverage
	minCov := prompt("Minimum test coverage %", "60")

	config := fmt.Sprintf(`# quality-gate configuration
# Generated setup
minCoverage: %s
autoFormat: %s
runTestsOnPush: %s
`, minCov, yesNo(autoFmt), yesNo(pushTests))

	if apiKey != "" {
		config += "\n# Store API keys in environment, not config files\n"
		config += "# export ANTHROPIC_API_KEY=sk-ant-...\n"
	}

	os.WriteFile(configPath, []byte(config), 0644)

	if apiKey != "" {
		fmt.Printf("\n⚠️  API key NOT stored. Add to your shell profile:\n")
		fmt.Printf("   echo 'export ANTHROPIC_API_KEY=%s' >> ~/.zshrc\n", apiKey)
	}

	fmt.Printf("\n✅ Setup complete! Config: %s\n\n", configPath)
	fmt.Println("Next: quality-gate enable")
}

func prompt(question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("%s: ", question)
	}
	var answer string
	fmt.Scanln(&answer)
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultVal
	}
	return answer
}

func yesNo(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "yes" || s == "y" || s == "true" {
		return "true"
	}
	return "false"
}
