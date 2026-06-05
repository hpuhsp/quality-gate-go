package checker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// SecurityScan detects Command Injection, Path Traversal, and SSRF patterns.
// These supplement the SQL injection checker with broader security coverage.
func SecurityScan(staged []string) CheckResult {
	result := CheckResult{OK: true}

	// Command Injection patterns
	cmdPatterns := []struct {
		name string
		regex *regexp.Regexp
		severity string
	}{
		{"Runtime.exec() with string concat", regexp.MustCompile(`(?i)Runtime\.(getRuntime\(\)\.)?exec\s*\(\s*["'][^"']*\+`), "critical"},
		{"ProcessBuilder with user input", regexp.MustCompile(`(?i)new\s+ProcessBuilder\s*\([^)]*\+`), "critical"},
		{"os.system() with string concat", regexp.MustCompile(`(?i)os\.(system|popen)\s*\(\s*["'][^"']*\+`), "critical"},
		{"exec() with string concat", regexp.MustCompile(`(?i)\bexec\s*\(\s*["'][^"']*["']\s*\+`), "high"},
		{"Go exec.Command with variable", regexp.MustCompile(`exec\.Command\s*\([^)]*\+`), "critical"},
		{"Go exec.CommandContext with variable", regexp.MustCompile(`exec\.CommandContext\s*\([^)]*\+`), "critical"},
	}

	// Path Traversal patterns
	pathPatterns := []struct {
		name string
		regex *regexp.Regexp
		severity string
	}{
		{"Path with ../ in variable", regexp.MustCompile(`(?i)path\s*[+:]=?\s*.*\.\./`), "high"},
		{"File read with user-controlled path", regexp.MustCompile(`(?i)new\s+File\s*\([^)]*\+`), "high"},
		{"FileInputStream with concat", regexp.MustCompile(`(?i)new\s+FileInputStream\s*\([^)]*\+`), "high"},
		{"open() with user input", regexp.MustCompile(`(?i)open\s*\(\s*[^)]*\+`), "medium"},
	}

	// SSRF patterns
	ssrfPatterns := []struct {
		name string
		regex *regexp.Regexp
		severity string
	}{
		{"URL from user input", regexp.MustCompile(`(?i)new\s+(URL|URI)\s*\(\s*[^)]*\+`), "high"},
		{"HTTP request with concat URL", regexp.MustCompile(`(?i)(HttpGet|HttpPost|HttpURLConnection|requests\.get|fetch)\s*\(\s*["'][^"']*\+`), "high"},
		{"OkHttp/Retrofit with variable URL", regexp.MustCompile(`(?i)(HttpUrl\.parse|\.url)\s*\(\s*[^)]*\+`), "medium"},
	}

	for _, file := range staged {
		data, err := shared.SafeReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}

			// Command Injection
			for _, p := range cmdPatterns {
				if p.regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1, Pattern: fmt.Sprintf("CMD: %s", p.name), Severity: p.severity,
					})
				}
			}

			// Path Traversal
			for _, p := range pathPatterns {
				if p.regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1, Pattern: fmt.Sprintf("PATH: %s", p.name), Severity: p.severity,
					})
				}
			}

			// SSRF
			for _, p := range ssrfPatterns {
				if p.regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1, Pattern: fmt.Sprintf("SSRF: %s", p.name), Severity: p.severity,
					})
				}
			}
		}
	}

	for _, f := range result.Findings {
		if f.Severity == "critical" || f.Severity == "high" {
			result.OK = false
			break
		}
	}
	return result
}
