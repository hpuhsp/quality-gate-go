package cmd

import (
	"fmt"
	"os"

	"github.com/hpuhsp/quality-gate/internal/checker"
	"github.com/hpuhsp/quality-gate/internal/detect"
)

func RunPreCommit() {
	proj := detect.Detect(".")
	hasErrors := false
	gatesTotal := 0

	// Gate 1: Secret scan
	gatesTotal++
	secretResult := checker.SecretScan()
	if !secretResult.OK {
		fmt.Print("\n❌ SECRET SCAN FAILED\n")
		checker.PrintFindings(secretResult.Findings, "secret")
		fmt.Print("  Fix: remove secrets before committing.\n\n")
		hasErrors = true
	} else if len(secretResult.Findings) > 0 {
		fmt.Println("⚠️  Secret scan: non-blocking patterns found")
		checker.PrintFindings(secretResult.Findings, "secret")
	}

	// Gate 2: Syntax check
	gatesTotal++
	syntaxResult := checker.SyntaxCheck()
	if !syntaxResult.OK {
		fmt.Print("\n❌ SYNTAX CHECK FAILED\n")
		for _, f := range syntaxResult.Findings {
			fmt.Printf("  %s: %s\n", f.File, f.Pattern)
		}
		fmt.Print("  Fix: correct the syntax errors before committing.\n\n")
		hasErrors = true
	}

	// Gate 3: SQL injection check
	gatesTotal++
	sqlResult := checker.SQLCheck()
	if !sqlResult.OK {
		fmt.Print("\n❌ SQL INJECTION RISK\n")
		checker.PrintSQLLines(sqlResult.Findings)
		fmt.Print("  Fix: use PreparedStatement / parameterized queries.\n\n")
		hasErrors = true
	} else if len(sqlResult.Findings) > 0 {
		fmt.Println("⚠️  SQL check: review these patterns")
		checker.PrintSQLLines(sqlResult.Findings)
	}

	// Gate 4: Auto-format
	if proj.Language != "unknown" {
		_, issues := checker.FormatCheck(proj)
		if len(issues) > 0 {
			fmt.Println("⚠️  Formatter not available:")
			checker.PrintFormatIssues(issues)
		}
	}

	if hasErrors {
		fmt.Printf("⛔ Commit blocked (%d gates total). Fix issues above.\n\n", gatesTotal)
		os.Exit(1)
	}

	tags := ""
	if proj.Language != "unknown" {
		tags = " (" + proj.Language
		if len(secretResult.Findings) > 0 {
			tags += ", secrets-warn"
		}
		tags += ")"
	}
	fmt.Printf("✅ pre-commit passed%s\n", tags)
}
