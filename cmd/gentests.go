package cmd

import (
	"fmt"
	"os"
)

// RunGenTests generates unit tests using AI (requires ANTHROPIC_API_KEY).
func RunGenTests(args ...string) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY not set.")
		fmt.Println("   export ANTHROPIC_API_KEY=sk-ant-...")
		os.Exit(1)
	}
	// Find target file: skip command name and 'gen-tests' itself
	target := ""
	for i, a := range args {
		if a == "gen-tests" && i+1 < len(args) {
			target = args[i+1]
			break
		}
	}
	// Also check the very last arg if not found (for 'tool gen-tests file' pattern)
	if target == "" && len(args) > 0 {
		last := args[len(args)-1]
		if last != "gen-tests" && last != "tool" && last != "quality-gate" {
			target = last
		}
	}
	if target != "" {
		fmt.Printf("Generating tests for %s...\n", target)
	} else {
		fmt.Println("Generating tests for changed code...")
	}
	fmt.Println("⚠️  AI test generation requires gstack CLI or claude CLI.")
	fmt.Println("   Install: npm install -g gstack")
}
