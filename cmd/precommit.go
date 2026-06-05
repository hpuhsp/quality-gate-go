package cmd

import (
	"fmt"
	"os"

	"github.com/hpuhsp/quality-gate-go/internal/checker"
	"github.com/hpuhsp/quality-gate-go/internal/config"
	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

func RunPreCommit() {
	proj := detect.Detect(".")
	cfg := config.Load(".")

	hasErrors := false
	gatesTotal := 0
	gatesPassed := 0

	// Get staged files once (shared across all checkers)
	staged := getStagedFiles()

	// Gate 1: Secret scan (regex + entropy)
	gatesTotal++
	if cfg.Secret.Enabled {
		secretResult := checker.SecretScan()
		// Also run entropy scan for high-entropy strings
		entropyFindings := runEntropyScan(staged)
		secretResult.Findings = append(secretResult.Findings, entropyFindings...)

		if !secretResult.OK {
			fmt.Print("\n❌ SECRET SCAN FAILED\n")
			checker.PrintFindings(secretResult.Findings)
			fmt.Print("  Fix: remove secrets before committing.\n\n")
			hasErrors = true
		} else if len(secretResult.Findings) > 0 {
			fmt.Println("⚠️  Secret scan: non-blocking patterns found")
			checker.PrintFindings(secretResult.Findings)
		} else {
			gatesPassed++
		}
	} else {
		gatesPassed++
	}

	// Gate 2: Syntax check
	gatesTotal++
	if cfg.Syntax.Enabled {
		syntaxResult := checker.SyntaxCheck()
		if !syntaxResult.OK {
			fmt.Print("\n❌ SYNTAX CHECK FAILED\n")
			for _, f := range syntaxResult.Findings {
				fmt.Printf("  %s: %s\n", f.File, f.Pattern)
			}
			fmt.Print("  Fix: correct the syntax errors before committing.\n\n")
			hasErrors = true
		} else {
			gatesPassed++
		}

		if len(syntaxResult.Skipped) > 0 {
			fmt.Printf("⚠️  Syntax check skipped for %d files (unsupported extension):\n", len(syntaxResult.Skipped))
			for _, f := range syntaxResult.Skipped {
				fmt.Printf("  - %s\n", f)
			}
		}
	} else {
		gatesPassed++
	}

	// Gate 3: SQL injection + security rules (command injection, path traversal, SSRF)
	gatesTotal++
	if cfg.Security.Enabled {
		sqlResult := checker.SQLCheck()
		secResult := checker.SecurityScan(staged)

		sqlResult.Findings = append(sqlResult.Findings, secResult.Findings...)
		if !sqlResult.OK || !secResult.OK {
			sqlResult.OK = false
		}

		if !sqlResult.OK {
			fmt.Print("\n❌ SECURITY CHECK FAILED\n")
			checker.PrintSQLLines(sqlResult.Findings)
			fmt.Print("  Fix: use parameterized queries, avoid Runtime.exec with user input.\n\n")
			hasErrors = true
		} else if len(sqlResult.Findings) > 0 {
			fmt.Println("⚠️  Security: review these patterns")
			checker.PrintSQLLines(sqlResult.Findings)
		} else {
			gatesPassed++
		}
	} else {
		gatesPassed++
	}

	// Gate 4: Auto-format (always counted, warning-only if formatter missing)
	gatesTotal++
	if cfg.Format.Enabled && proj.Language != "unknown" {
		_, issues := checker.FormatCheck(proj)
		if len(issues) > 0 {
			fmt.Println("⚠️  Formatter not available:")
			checker.PrintFormatIssues(issues)
		} else {
			gatesPassed++
		}
	} else {
		gatesPassed++
	}

	// Gate 5: Lint check (optional, enabled via config)
	if cfg.Lint.Enabled && proj.Language != "unknown" {
		gatesTotal++
		lintResult := checker.LintCheck(proj, staged)
		if !lintResult.OK {
			fmt.Print("\n❌ LINT CHECK FAILED\n")
			for _, f := range lintResult.Findings {
				fmt.Printf("  %s\n", f.Pattern)
			}
			fmt.Print("  Fix: resolve lint issues before committing.\n\n")
			hasErrors = true
		} else {
			gatesPassed++
		}
	}

	// Gate 6: Architecture rule check (optional, enabled via config)
	if cfg.Arch.Enabled && len(cfg.Arch.Forbidden) > 0 {
		gatesTotal++
		archResult := checker.ArchCheck(staged, cfg.Arch)
		if !archResult.OK {
			fmt.Print("\n⚠️  ARCHITECTURE VIOLATION\n")
			for _, f := range archResult.Findings {
				fmt.Printf("  %s: %s\n", f.File, f.Pattern)
			}
			fmt.Print("  Review: architectural layer violations detected.\n\n")
			hasErrors = true
		} else {
			gatesPassed++
		}
	}

	if hasErrors {
		fmt.Printf("⛔ Commit blocked (%d gates total). Fix issues above.\n\n", gatesTotal)
		os.Exit(1)
	}

	tags := ""
	if proj.Language != "unknown" {
		tags = " (" + proj.Language
		tags += fmt.Sprintf(", %d/%d gates", gatesPassed, gatesTotal)
		tags += ")"
	}
	fmt.Printf("✅ pre-commit passed%s\n", tags)
}

// getStagedFiles returns the list of staged files (called once, reused).
func getStagedFiles() []string {
	out, err := runCmd("git", "diff", "--cached", "--name-only", "--diff-filter=ACM")
	if err != nil {
		return nil
	}
	var files []string
	for _, f := range splitLines(out) {
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// runEntropyScan runs entropy-based secret detection on staged files.
func runEntropyScan(staged []string) []checker.Finding {
	var findings []checker.Finding
	for _, file := range staged {
		data, err := readFileSafe(file)
		if err != nil {
			continue
		}
		findings = append(findings, checker.EntropyScan(string(data), file)...)
	}
	return findings
}
