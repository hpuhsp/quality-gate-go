package checker

import (
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

func FormatCheck(proj detect.Result) (ok bool, issues []string) {
	ok = true
	staged := getStagedFiles()
	if len(staged) == 0 {
		return
	}

	switch proj.Language {
	case "kotlin":
		if shared.HasBin("ktlint") {
			ktFiles := filterByExt(staged, ".kt")
			if formatFiles("ktlint", ktFiles...) {
				for _, f := range ktFiles {
					exec.Command("git", "add", f).Run()
				}
			}
		} else if len(filterByExt(staged, ".kt")) > 0 {
			issues = append(issues, "ktlint not installed — Kotlin formatting skipped")
		}
	case "java":
		if shared.HasBin("google-java-format") {
			javaFiles := filterByExt(staged, ".java")
			if formatFiles("google-java-format", append([]string{"--replace"}, javaFiles...)...) {
				for _, f := range javaFiles {
					exec.Command("git", "add", f).Run()
				}
			}
		} else if len(filterByExt(staged, ".java")) > 0 {
			issues = append(issues, "google-java-format not installed — Java formatting skipped")
		}
	case "javascript":
		if shared.HasBin("prettier") || shared.HasBin("npx") {
			jsFiles := filterByExt(staged, ".js", ".ts", ".jsx", ".tsx", ".json", ".css", ".md", ".yml", ".yaml")
			args := []string{"prettier", "--write"}
			if !shared.HasBin("prettier") {
				args = []string{"npx", "prettier", "--write"}
			}
			if len(jsFiles) > 0 {
				exec.Command(args[0], append(args[1:], jsFiles...)...).Run()
				for _, f := range jsFiles {
					exec.Command("git", "add", f).Run()
				}
			}
		} else if len(filterByExt(staged, ".js", ".ts")) > 0 {
			issues = append(issues, "prettier not installed — JS/TS formatting skipped")
		}
	}
	return
}

func filterByExt(files []string, exts ...string) []string {
	var out []string
	for _, f := range files {
		ext := filepath.Ext(f)
		for _, e := range exts {
			if ext == e {
				out = append(out, f)
				break
			}
		}
	}
	return out
}

func formatFiles(cmd string, files ...string) bool {
	if len(files) == 0 {
		return false
	}
	c := exec.Command(cmd, files...)
	err := c.Run()
	return err == nil
}


func PrintFormatIssues(issues []string) {
	for _, i := range issues {
		fmt.Printf("  ⚠️  %s\n", i)
	}
}
