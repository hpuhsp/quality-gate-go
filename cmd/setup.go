package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hpuhsp/quality-gate-go/internal/shared"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

type toolDep struct {
	bin     string
	install string
	purpose string
}

func langDeps(lang string) []toolDep {
	switch lang {
	case "kotlin":
		return []toolDep{{"ktlint", "brew install ktlint", "Kotlin code formatting"}}
	case "java":
		return []toolDep{{"google-java-format", "brew install google-java-format", "Java code formatting"}}
	case "javascript":
		return []toolDep{{"prettier", "npm install -g prettier", "JS/TS/CSS formatting"}}
	case "go":
		return []toolDep{{"gofmt", "bundled with Go", "Go code formatting"}}
	}
	return nil
}

func RunSetup() {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".quality-gate")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.yml")

	// Detect project
	proj := detect.Detect(".")

	fmt.Print(`
╔══════════════════════════════════════════╗
║   quality-gate — First Run Setup         ║
╚══════════════════════════════════════════╝

Press Enter to accept defaults.
`)

	// ── Step 1: Project detection ─────────────────────────────────────
	fmt.Printf("─── Project Detection ───\n")
	fmt.Printf("  Language:     %s\n", proj.Language)
	fmt.Printf("  Build tool:   %s\n", proj.BuildTool)
	fmt.Printf("  Test FW:      %s\n", proj.TestFramework)
	fmt.Println()

	// ── Step 2: Language-specific dependencies ────────────────────────
	deps := langDeps(proj.Language)
	missing := make([]toolDep, 0)
	for _, d := range deps {
		if !shared.HasBin(d.bin) {
			missing = append(missing, d)
		}
	}

	if len(missing) > 0 {
		fmt.Println("─── Required Tools ───")
		for _, d := range deps {
			if shared.HasBin(d.bin) {
				fmt.Printf("  ✅ %s (%s)\n", d.bin, d.purpose)
			} else {
				fmt.Printf("  ❌ %s — %s\n", d.bin, d.install)
			}
		}
		fmt.Println()

		if len(missing) > 0 {
			answer := prompt("Install missing tools now?", "y")
			if answer == "y" || answer == "Y" || answer == "yes" || answer == "" {
				for _, d := range missing {
					fmt.Printf("  Installing %s...\n", d.bin)
					err := runInstallCmd(d.install)
					if err != nil {
						fmt.Printf("    ❌ Failed: %v\n", err)
						fmt.Printf("    Manual: %s\n", d.install)
					} else {
						fmt.Printf("    ✅ %s installed\n", d.bin)
					}
				}
				fmt.Println()
			}
		}
	}

	// ── Step 3: AI Test Generation ────────────────────────────────────
	fmt.Println("─── AI Test Generation ───")
	fmt.Println("quality-gate can auto-generate unit tests using AI.")
	apiKey := prompt("Anthropic API Key (sk-ant-..., leave empty to skip)", "")

	// ── Step 4: Test execution ─────────────────────────────────────────
	fmt.Println("\n─── Test Execution ───")
	fmt.Println("  ✅ yes — Run tests locally on git push (slower)")
	fmt.Println("  ❌ no  — Let CI handle tests (recommended for teams)")
	pushTests := prompt("Run unit tests on git push?", "no")

	// ── Step 5: Formatting ─────────────────────────────────────────────
	fmt.Println("\n─── Code Formatting ───")
	autoFmt := prompt("Auto-format code on commit?", "yes")

	// ── Step 6: Coverage ───────────────────────────────────────────────
	minCov := prompt("Minimum test coverage %", "60")

	// ── Write config ───────────────────────────────────────────────────
	config := fmt.Sprintf(`# quality-gate configuration
# Generated: %s
minCoverage: %s
autoFormat: %s
runTestsOnPush: %s
`, timestamp(), minCov, yesNo(autoFmt), yesNo(pushTests))

	if apiKey != "" {
		config += "\n# Store API keys in environment, not config files\n"
		config += "# Run: export ANTHROPIC_API_KEY=sk-ant-...\n"
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		fmt.Printf("❌ Failed to write config: %v\n", err)
		os.Exit(1)
	}

	// ── API key note ──────────────────────────────────────────────────
	if apiKey != "" {
		fmt.Printf("\n⚠️  API key NOT stored. Persist it:\n")
		fmt.Print("   echo 'export ANTHROPIC_API_KEY=sk-ant-...' >> ~/.zshrc\n")
	}

	fmt.Printf("\n✅ Setup complete! Config: %s\n\n", configPath)
	fmt.Println("Next:")
	fmt.Println("  quality-gate enable    Activate hooks for this repo")
	fmt.Println("  quality-gate status    Verify everything is working")
	fmt.Println("  quality-gate doctor    Check dependencies anytime")
}

func prompt(question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("%s: ", question)
	}
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
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

func runInstallCmd(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return nil
}

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
// Feature: Async Nudge & Interactive Setup implemented
