package main

import (
	"fmt"
	"os"

	"github.com/hpuhsp/quality-gate/cmd"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "setup":
		cmd.RunSetup()
	case "doctor":
		cmd.RunDoctor()
	case "update":
		cmd.RunUpdate()
	case "enable":
		cmd.RunEnable()
	case "disable":
		cmd.RunDisable()
	case "status":
		cmd.RunStatus()
	case "pre-commit":
		cmd.RunPreCommit()
	case "pre-push":
		cmd.RunPrePush()
	case "gen-tests":
		cmd.RunGenTests(os.Args...)
	case "tool":
		if len(os.Args) > 2 && os.Args[2] == "gen-tests" {
			cmd.RunGenTests(os.Args...)
		} else {
			fmt.Println("quality-gate tool — Optional tools")
			fmt.Println("  tool gen-tests [file]   AI-generate test code (needs ANTHROPIC_API_KEY)")
		}
	case "version", "--version", "-v":
		fmt.Println(cmd.Version)
	case "--help", "-h":
		printHelp()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Printf(`quality-gate v%s — Shift-Left Quality Gates

  setup       First-run wizard
  doctor      Check & auto-install dependencies
  update      Update to latest version
  version     Show version information
  enable      Activate 4-gate pre-commit hook
  disable     Deactivate hooks
  status      Show status and project detection
  tool        Optional tools (test generation, etc.)
`, cmd.Version)
}
