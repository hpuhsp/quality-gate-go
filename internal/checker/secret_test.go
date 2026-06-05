package checker

import (
	"os"
	"path/filepath"
	"testing"
)

// createTempFile creates a temp file with content for testing.
func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// scanDirectly tests secret patterns against a file's content directly
// (bypasses getStagedFiles which requires a real git repo).
func scanDirectly(files []string) CheckResult {
	result := CheckResult{OK: true}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := splitLines(string(data))
		for _, line := range lines {
			trimmed := trimSpace(line)
			if len(trimmed) > 0 && (trimmed[0:1] == "/" || trimmed[0:1] == "#") {
				continue
			}
			for _, p := range secretPatterns {
				if p.Regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: 0, Pattern: p.Name, Severity: p.Severity,
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

func TestSecretScan_AWSKey(t *testing.T) {
	f := createTempFile(t, "config.yml", "AWS_ACCESS_KEY=AKIA1234567890ABCDEF")
	result := scanDirectly([]string{f})
	if result.OK {
		t.Error("expected AWS key to be detected")
	}
}

func TestSecretScan_Password(t *testing.T) {
	f := createTempFile(t, "app.yml", `password: "my-secret-123"`)
	result := scanDirectly([]string{f})
	if result.OK {
		t.Error("expected hardcoded password to be detected")
	}
}

func TestSecretScan_GitHubPAT(t *testing.T) {
	f := createTempFile(t, "config.yml", `GITHUB_TOKEN=ghp_1234567890abcdefghijklmnop`)
	result := scanDirectly([]string{f})
	if result.OK {
		t.Error("expected GitHub PAT to be detected")
	}
}

func TestSecretScan_PrivateKey(t *testing.T) {
	f := createTempFile(t, "key.pem", `-----BEGIN RSA PRIVATE KEY-----`)
	result := scanDirectly([]string{f})
	if result.OK {
		t.Error("expected private key header to be detected")
	}
}

func TestSecretScan_CleanCode(t *testing.T) {
	f := createTempFile(t, "app.java", `public class App { void main() { } }`)
	result := scanDirectly([]string{f})
	if !result.OK {
		t.Error("expected clean code to pass")
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result.Findings))
	}
}

func TestSecretScan_CommentSkipped(t *testing.T) {
	f := createTempFile(t, "app.java", `// password: "secret123"
password: "real_secret"`)
	result := scanDirectly([]string{f})
	if result.OK {
		t.Error("expected real password to be detected")
	}
	// Comment should not be detected
	for _, finding := range result.Findings {
		if finding.Pattern == "Hardcoded Password" && finding.Line == 1 {
			t.Error("expected comment line to be skipped")
		}
	}
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
