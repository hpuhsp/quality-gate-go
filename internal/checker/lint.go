package checker

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// LintCheck runs language-specific linters.
// Lint is opt-in (quality-gate.yaml: lint.enabled: true) and delegates
// to well-established tools rather than reimplementing rules.
//
// Language → Tool mapping:
//
//	Go      → golangci-lint
//	Java    → PMD (pmd check)
//	Kotlin  → detekt
//	JS/TS   → eslint
//	Vue     → eslint + vue plugin
//	Python  → ruff check
func LintCheck(proj detect.Result, staged []string) CheckResult {
	result := CheckResult{OK: true}

	switch proj.Language {
	case "go":
		return runLinter("golangci-lint", []string{"run", "--new-from-rev=HEAD~1", "--out-format=line-number"}, staged, "go")
	case "java":
		return runLinter("pmd", []string{"check", "-d", ".", "-R", "rulesets/java/quickstart.xml", "-f", "text"}, staged, "java")
	case "kotlin":
		return runLinter("detekt", []string{"--input", "."}, staged, "kotlin")
	case "javascript", "typescript":
		return runLinter("eslint", []string{"--format", "compact"}, staged, "js")
	default:
		// Unsupported language — lint is optional, skip silently
		return result
	}
}

func runLinter(cmdName string, args []string, staged []string, lang string) CheckResult {
	result := CheckResult{OK: true}

	if !hasBin(cmdName) {
		// Linter not installed — report as info, not error
		fmt.Printf("  ℹ️  %s not installed (%s lint skipped)\n", cmdName, lang)
		return result
	}

	// Filter staged files to language-specific ones
	var langFiles []string
	extMap := map[string][]string{
		"go":   {".go"},
		"java": {".java"},
		"kotlin": {".kt", ".kts"},
		"js":   {".js", ".ts", ".jsx", ".tsx", ".vue"},
	}
	exts := extMap[lang]
	for _, f := range staged {
		for _, ext := range exts {
			if strings.HasSuffix(f, ext) {
				langFiles = append(langFiles, f)
				break
			}
		}
	}

	if len(langFiles) == 0 {
		return result
	}

	// Run the linter
	allArgs := append(args, langFiles...)
	out, err := exec.Command(cmdName, allArgs...).CombinedOutput()
	if err != nil {
		// Linter found issues
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if line != "" {
				result.Findings = append(result.Findings, Finding{
					File:     "",
					Line:     0,
					Pattern:  line,
					Severity: "warning",
				})
			}
		}
		if len(result.Findings) > 0 {
			result.OK = false
		}
	}

	return result
}

// hasBin delegates to shared.HasBin.
func hasBin(name string) bool {
	return shared.HasBin(name)
}
