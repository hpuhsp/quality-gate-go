package checker

import (
	"regexp"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// Pre-compiled patterns (package level, compiled once)
type secPattern struct {
	name     string
	regex    *regexp.Regexp
	severity string
}

var cmdPatterns = []secPattern{
	// Java
	{"Runtime.exec", regexp.MustCompile(`(?i)Runtime[\w.()]*\.exec\s*\(`), "critical"},
	{"ProcessBuilder", regexp.MustCompile(`(?i)new\s+ProcessBuilder\s*\(`), "critical"},
	// Python
	{"os.system", regexp.MustCompile(`(?i)os\.(system|popen)\s*\(`), "critical"},
	// Go
	{"Go exec.Command", regexp.MustCompile(`exec\.Command(Context)?\s*\(`), "critical"},
	// Generic
	{"Shell exec", regexp.MustCompile(`(?i)\bexec\s*\(\s*["']`), "high"},
}

var pathPatterns = []secPattern{
	{"Path traversal (../)", regexp.MustCompile(`(?i)(path|file)\s*[+:]=?\s*.*\.\./`), "high"},
	{"File with concat", regexp.MustCompile(`(?i)new\s+File(?:InputStream|Reader)?\s*\([^)]*\+`), "high"},
	{"open() with concat", regexp.MustCompile(`(?i)open\s*\(\s*[^)]*\+`), "medium"},
}

var ssrfPatterns = []secPattern{
	{"URL constructed with input", regexp.MustCompile(`(?i)new\s+(URL|URI)\s*\([^)]*\+`), "high"},
	{"HTTP with concat", regexp.MustCompile(`(?i)(fetch|axios|requests\.get)\s*\(\s*[^)]*\+`), "high"},
	{"HttpClient URL", regexp.MustCompile(`(?i)(HttpGet|HttpPost|HttpURLConnection)\s*\(`), "medium"},
}

// SecurityScan detects Command Injection, Path Traversal, and SSRF patterns.
func SecurityScan(staged []string) CheckResult {
	result := CheckResult{OK: true}
	allPatterns := make([]secPattern, 0, len(cmdPatterns)+len(pathPatterns)+len(ssrfPatterns))
	allPatterns = append(append(append(allPatterns, cmdPatterns...), pathPatterns...), ssrfPatterns...)

	for _, file := range staged {
		data, err := shared.SafeReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		blockDepth := 0
		for i, line := range lines {
			// Track /* ... */ block comments (depth counter handles nesting)
			blockDepth += strings.Count(line, "/*") - strings.Count(line, "*/")
			if blockDepth < 0 {
				blockDepth = 0
			}
			if blockDepth > 0 {
				continue
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			for _, p := range allPatterns {
				if p.regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1,
						Pattern:  p.name,
						Severity: p.severity,
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
