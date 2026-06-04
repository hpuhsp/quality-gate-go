package checker

import (
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"fmt"
	"regexp"
	"strings"
)

var sqliPatterns = []struct {
	Name     string
	Regex    *regexp.Regexp
	Severity string
}{
	{Name: "SQL string concat", Regex: regexp.MustCompile(`(?i)["']\s*\+.*SELECT\b|SELECT\b.*\+`), Severity: "critical"},
	{Name: "SQL string concat", Regex: regexp.MustCompile(`(?i)["']\s*\+.*INSERT\b|INSERT\b.*\+`), Severity: "critical"},
	{Name: "SQL string concat", Regex: regexp.MustCompile(`(?i)["']\s*\+.*UPDATE\b|UPDATE\b.*\+`), Severity: "critical"},
	{Name: "SQL string concat", Regex: regexp.MustCompile(`(?i)["']\s*\+.*DELETE\b|DELETE\b.*\+`), Severity: "critical"},
	{Name: "SQL string concat", Regex: regexp.MustCompile(`(?i)["']\s*\+.*WHERE\b|WHERE\b.*\+`), Severity: "critical"},
	{Name: "SQL with String.format", Regex: regexp.MustCompile(`(?i)String\.format\s*\(\s*["'][^"']*(?:SELECT|INSERT|UPDATE|DELETE)\b`), Severity: "critical"},
	{Name: "Raw SQL execution with concat", Regex: regexp.MustCompile(`(?i)\.execute(?:Query|Update)\s*\(\s*["'][^"']*\+`), Severity: "critical"},
	{Name: "SQL in template literal", Regex: regexp.MustCompile("(?i)`\\s*(?:SELECT|INSERT|UPDATE|DELETE)\\b[^`]*\\$\\{[^}]*\\}[^`]*`"), Severity: "critical"},
	{Name: "JdbcTemplate string SQL", Regex: regexp.MustCompile(`(?i)jdbcTemplate\.\w+\s*\(\s*["'][^"']*\+`), Severity: "high"},
	{Name: "NativeQuery string concat", Regex: regexp.MustCompile(`(?i)createNativeQuery\s*\(\s*["'][^"']*\+`), Severity: "high"},
	{Name: "Raw Statement usage", Regex: regexp.MustCompile(`(?i)\.createStatement\s*\(\s*\)|Statement\s+stmt\s*=`), Severity: "medium"},
	{Name: "Dynamic ORDER BY", Regex: regexp.MustCompile(`(?i)ORDER\s+BY\s*\+`), Severity: "medium"},
	{Name: "Dynamic GROUP BY", Regex: regexp.MustCompile(`(?i)GROUP\s+BY\s*\+`), Severity: "medium"},
}

var srcExt = regexp.MustCompile(`(?i)\.(java|kt|js|ts|py|php)$`)

func SQLCheck() CheckResult {
	result := CheckResult{OK: true}
	staged := getStagedFiles()
	if len(staged) == 0 {
		return result
	}

	for _, file := range staged {
		if !srcExt.MatchString(file) {
			continue
		}
		data, err := shared.SafeReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			for _, p := range sqliPatterns {
				if p.Regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1, Pattern: p.Name, Severity: p.Severity,
					})
				}
			}
		}
	}

	for _, f := range result.Findings {
		if f.Severity == "critical" {
			result.OK = false
			break
		}
	}
	return result
}

// PrintSQLLines formats SQL findings.
func PrintSQLLines(findings []Finding) {
	for _, f := range findings {
		fmt.Printf("  %s:%d — %s [%s]\n", f.File, f.Line, f.Pattern, f.Severity)
	}
}
