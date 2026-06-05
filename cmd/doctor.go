package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

type dep struct {
	name     string
	bin      string
	required bool
	install  string
	purpose  string
}

var deps = []dep{
	{name: "git", bin: "git", required: true, purpose: "hook management + staged file scanning"},
	{name: "node", bin: "node", required: false, purpose: "JS/TS syntax check"},
	{name: "gofmt", bin: "gofmt", required: false, purpose: "Go syntax check (bundled with Go)"},
	{name: "ktlint", bin: "ktlint", required: false, install: "brew install ktlint", purpose: "Kotlin formatting"},
	{name: "google-java-format", bin: "google-java-format", required: false, install: "brew install google-java-format", purpose: "Java formatting"},
	{name: "prettier", bin: "prettier", required: false, install: "npm install -g prettier", purpose: "JS/TS/CSS/MD formatting"},
}

func RunDoctor() {
	proj := detect.Detect(".")

	fmt.Println("quality-gate doctor")
	fmt.Println("──────────────────")
	fmt.Println()

	missingRequired := false
	missing := make([]dep, 0)
	ok := 0

	for _, d := range deps {
		if shared.HasBin(d.bin) {
			ok++
			fmt.Printf("  ✅ %-22s %s\n", d.name, d.purpose)
		} else {
			if d.required {
				fmt.Printf("  ❌ %-22s %s — MUST INSTALL\n", d.name, d.purpose)
				missingRequired = true
			} else {
				fmt.Printf("  ⚠️  %-22s %s\n", d.name, d.purpose)
				missing = append(missing, d)
			}
		}
	}

	fmt.Printf("\n  %d/%d deps satisfied\n", ok, len(deps))
	fmt.Println()

	// Language-specific recommendations
	if proj.Language != "unknown" {
		fmt.Printf("  Detected: %s (%s)\n", proj.Language, proj.BuildTool)
		switch proj.Language {
		case "kotlin":
			if !shared.HasBin("ktlint") {
				fmt.Println("  💡 Install ktlint for auto-formatting: brew install ktlint")
			}
		case "java":
			if !shared.HasBin("google-java-format") {
				fmt.Println("  💡 Install google-java-format: brew install google-java-format")
			}
		case "javascript":
			if !shared.HasBin("prettier") && !shared.HasBin("npx") {
				fmt.Println("  💡 Install prettier: npm install -g prettier")
			}
		case "go":
			if !shared.HasBin("gofmt") {
				fmt.Println("  💡 gofmt should come with Go. Check your Go installation.")
			}
		}
	}

	if missingRequired {
		fmt.Println()
		fmt.Println("❌ Required dependencies missing. quality-gate cannot function without git.")
		return
	}

	if len(missing) == 0 {
		fmt.Println("✅ All dependencies satisfied.")
		return
	}

	fmt.Println()
	fmt.Println("Missing optional dependencies:")
	for _, d := range missing {
		if d.install != "" {
			fmt.Printf("  %s — %s\n", d.name, d.install)
		}
	}

	fmt.Println()
	fmt.Print("Auto-install missing dependencies? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer != "y" && answer != "Y" && answer != "yes" {
		fmt.Println("  Skipped. Run 'quality-gate doctor' anytime to check again.")
		return
	}

	fmt.Println()
	for _, d := range missing {
		if d.install == "" {
			continue
		}
		fmt.Printf("Installing %s...\n", d.name)
		err := runInstall(d.install)
		if err != nil {
			fmt.Printf("  ❌ Failed: %v\n", err)
		} else {
			fmt.Printf("  ✅ %s installed\n", d.name)
		}
	}

	fmt.Println("\nRun 'quality-gate doctor' again to verify.")
}


func runInstall(cmd string) error {
	c := exec.Command("sh", "-c", cmd)
	out, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(out))
	}
	return nil
}
