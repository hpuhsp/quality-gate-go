package checker

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

func FormatCheck(proj detect.Result, autoFix bool) (ok bool, issues []string) {
	staged := GetStagedFiles()
	if len(staged) == 0 {
		return
	}

	switch proj.Language {
	case "kotlin":
		if shared.HasBin("ktlint") {
			ktFiles := filterByExt(staged, ".kt")
			if autoFix {
				if formatFiles("ktlint", append([]string{"-F"}, ktFiles...)...) {
					for _, f := range ktFiles {
						exec.Command("git", "add", "--", f).Run()
					}
				}
			} else {
				if !formatFiles("ktlint", ktFiles...) {
					issues = append(issues, "Kotlin formatting issues found (run with auto_fix:true or fix manually)")
				}
			}
		} else if len(filterByExt(staged, ".kt")) > 0 {
			issues = append(issues, "ktlint not installed — Kotlin formatting skipped")
		}
	case "java":
		if shared.HasBin("google-java-format") {
			javaFiles := filterByExt(staged, ".java")
			if autoFix {
				if formatFiles("google-java-format", append([]string{"--replace"}, javaFiles...)...) {
					for _, f := range javaFiles {
						exec.Command("git", "add", "--", f).Run()
					}
				}
			} else {
				if !formatFiles("google-java-format", append([]string{"--dry-run", "--set-exit-if-changed"}, javaFiles...)...) {
					issues = append(issues, "Java formatting issues found (run with auto_fix:true or fix manually)")
				}
			}
		} else if len(filterByExt(staged, ".java")) > 0 {
			issues = append(issues, "google-java-format not installed — Java formatting skipped")
		}
	case "javascript":
		if shared.HasBin("prettier") || shared.HasBin("npx") {
			jsFiles := filterByExt(staged, ".js", ".ts", ".jsx", ".tsx", ".json", ".css", ".md", ".yml", ".yaml", ".vue")
			if len(jsFiles) == 0 {
				break
			}
			bin := "prettier"
			if !shared.HasBin("prettier") {
				bin = "npx"
			}
			if autoFix {
				exec.Command(bin, append([]string{"prettier", "--write"}, jsFiles...)...).Run()
				for _, f := range jsFiles {
					exec.Command("git", "add", "--", f).Run()
				}
			} else {
				if err := exec.Command(bin, append([]string{"prettier", "--check"}, jsFiles...)...).Run(); err != nil {
					issues = append(issues, "JS/TS formatting issues found (run with auto_fix:true or fix manually)")
				}
			}
		} else if len(filterByExt(staged, ".js", ".ts")) > 0 {
			issues = append(issues, "prettier not installed — JS/TS formatting skipped")
		}
	case "go":
		if shared.HasBin("gofmt") {
			goFiles := filterByExt(staged, ".go")
			for _, f := range goFiles {
				if autoFix {
					exec.Command("gofmt", "-w", f).Run()
					exec.Command("git", "add", "--", f).Run()
				} else {
					out, _ := exec.Command("gofmt", "-d", f).Output()
					if len(out) > 0 {
						issues = append(issues, fmt.Sprintf("Go formatting issues in %s (run with auto_fix:true or fix manually)", f))
					}
				}
			}
		} else if len(filterByExt(staged, ".go")) > 0 {
			issues = append(issues, "gofmt not installed — Go formatting skipped")
		}
	}
	ok = len(issues) == 0
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
