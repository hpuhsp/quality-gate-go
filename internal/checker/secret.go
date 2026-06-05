package checker

import (
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type Finding struct {
	File     string
	Line     int
	Pattern  string
	Severity string
}

type CheckResult struct {
	OK       bool
	Findings []Finding
	Skipped  []string
}

var secretPatterns = []struct {
	Name     string
	Regex    *regexp.Regexp
	Severity string
}{
	{Name: "AWS Access Key", Regex: regexp.MustCompile(`(?:AKIA|ASIA)[A-Z0-9]{16}`), Severity: "critical"},
	{Name: "Alibaba Cloud AK", Regex: regexp.MustCompile(`LTAI[A-Za-z0-9]{16,}`), Severity: "critical"},
	{Name: "Tencent Cloud AK", Regex: regexp.MustCompile(`AKID[A-Za-z0-9]{13,}`), Severity: "critical"},
	{Name: "GitHub PAT", Regex: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{20,}`), Severity: "critical"},
	{Name: "GitLab PAT", Regex: regexp.MustCompile(`glpat-[A-Za-z0-9_-]{20,}`), Severity: "critical"},
	{Name: "OpenAI API Key", Regex: regexp.MustCompile(`sk-(?:proj-)?[A-Za-z0-9]{32,}`), Severity: "critical"},
	{Name: "Anthropic API Key", Regex: regexp.MustCompile(`sk-ant-(?:api|admin)[0-9]{2}-[A-Za-z0-9_-]{60,}`), Severity: "critical"},
	{Name: "Hardcoded Password", Regex: regexp.MustCompile(`(?i)(?:password|passwd|pwd)\s*[:=]\s*['\"][^'\"]{4,}['\"]`), Severity: "critical"},
	{Name: "Generic API Key", Regex: regexp.MustCompile(`(?i)(?:api_?key|api_?secret|secret_?key)\s*[:=]\s*['\"][A-Za-z0-9_-]{10,}['\"]`), Severity: "critical"},
	{Name: "Generic Token", Regex: regexp.MustCompile(`(?i)(?:auth_?token|access_?token)\s*[:=]\s*['\"][A-Za-z0-9_-]{10,}['\"]`), Severity: "critical"},
	{Name: "Private Key Header", Regex: regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`), Severity: "critical"},
	{Name: "AWS Secret Key", Regex: regexp.MustCompile(`(?i)aws_secret_access_key\s*[:=]\s*['\"][A-Za-z0-9/+=]{20,}`), Severity: "critical"},
	{Name: "JDBC Connection", Regex: regexp.MustCompile(`jdbc:[a-z]+://[^/]+/[^\s\"'?]+(?:\?[^\s\"']*)?.*(?:user|password)=`), Severity: "critical"},
	{Name: "JWT Token", Regex: regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`), Severity: "high"},
	{Name: "OAuth Bearer", Regex: regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-._~+/]{20,}`), Severity: "high"},
	{Name: "DB Connection String", Regex: regexp.MustCompile(`(?i)(?:DATABASE_URL|DB_URL|MONGO_URI|REDIS_URL)\s*[:=]\s*['\"][^'\"]{10,}`), Severity: "high"},
	{Name: "Slack Webhook", Regex: regexp.MustCompile(`https://hooks\.slack\.com/services/[A-Za-z0-9/]+`), Severity: "medium"},
	{Name: "Feishu Webhook", Regex: regexp.MustCompile(`https://open\.feishu\.cn/open-apis/bot/v2/hook/[A-Za-z0-9-]+`), Severity: "medium"},
}

var skipExt = regexp.MustCompile(`(?i)\.(png|jpe?g|gif|ico|svg|woff2?|ttf|eot|zip|tar|gz|jar|war|class|lock|map|min\.\w+|pb|bin|exe|dll|so|dylib|wasm|mp4|mp3|pdf)$`)

func SecretScan() CheckResult {
	result := CheckResult{OK: true}

	staged := getStagedFiles()
	if len(staged) == 0 {
		return result
	}

	for _, file := range staged {
		if skipExt.MatchString(file) || strings.Contains(file, "node_modules/") {
			continue
		}
		data, err := shared.SafeReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		inBlockComment := false
		for i, line := range lines {
			// Track multi-line /* ... */ block comments
			if inBlockComment {
				if strings.Contains(line, "*/") {
					inBlockComment = false
				}
				continue
			}
			if strings.Contains(line, "/*") && !strings.Contains(line, "*/") {
				inBlockComment = true
				continue
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			for _, p := range secretPatterns {
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

func getStagedFiles() []string {
	out, err := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACM").Output()
	if err != nil {
		return nil
	}
	var files []string
	for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files
}

// PrintFindings formats findings to stdout.
func PrintFindings(findings []Finding) {
	for _, f := range findings {
		fmt.Printf("  %s:%d — %s [%s]\n", f.File, f.Line, f.Pattern, f.Severity)
	}
}
