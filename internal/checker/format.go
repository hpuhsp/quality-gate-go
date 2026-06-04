package checker

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/hpuhsp/quality-gate/internal/detect"
)

func FormatCheck(proj detect.Result) (ok bool, issues []string) {
	ok = true
	staged := getStagedFiles()
	if len(staged) == 0 {
		return
	}

	switch proj.Language {
	case "kotlin":
		if hasBin("ktlint") {
			ktFiles := filterByExt(staged, ".kt")
			if formatFiles("ktlint", ktFiles...) {
				for _, f := range ktFiles {
					exec.Command("git", "add", f).Run()
				}
			}
		}
	case "java":
		if hasBin("google-java-format") {
			javaFiles := filterByExt(staged, ".java")
			if formatFiles("google-java-format", append([]string{"--replace"}, javaFiles...)...) {
				for _, f := range javaFiles {
					exec.Command("git", "add", f).Run()
				}
			}
		}
	case "javascript":
		if hasBin("prettier") || hasBin("npx") {
			jsFiles := filterByExt(staged, ".js", ".ts", ".jsx", ".tsx", ".json", ".css", ".md", ".yml", ".yaml")
			args := []string{"prettier", "--write"}
			if !hasBin("prettier") {
				args = []string{"npx", "prettier", "--write"}
			}
			if len(jsFiles) > 0 {
				exec.Command(args[0], append(args[1:], jsFiles...)...).Run()
				for _, f := range jsFiles {
					exec.Command("git", "add", f).Run()
				}
			}
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

func hasBin(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// PrintFormatIssues reports format problems.
func PrintFormatIssues(issues []string) {
	for _, i := range issues {
		fmt.Printf("  %s\n", i)
	}
}
