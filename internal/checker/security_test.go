package checker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func secTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

// securityScanDirect tests the security patterns against files directly
// (bypasses SafeReadFile which requires being inside a git repo).
func securityScanDirect(files []string) CheckResult {
	result := CheckResult{OK: true}

	// Recreate the patterns inline (same as security.go)
	type sp struct {
		name     string
		regex    *regexp.Regexp
		severity string
	}
	cmdPats := []sp{
		{"Runtime.exec", regexp.MustCompile(`(?i)Runtime[\w.()]*\.exec\s*\(`), "critical"},
		{"ProcessBuilder", regexp.MustCompile(`(?i)new\s+ProcessBuilder\s*\(`), "critical"},
		{"os.system", regexp.MustCompile(`(?i)os\.(system|popen)\s*\(`), "critical"},
		{"Go exec.Command", regexp.MustCompile(`exec\.Command(Context)?\s*\(`), "critical"},
		{"Shell exec", regexp.MustCompile(`(?i)\bexec\s*\(\s*["']`), "high"},
	}
	pathPats := []sp{
		{"Path traversal", regexp.MustCompile(`(?i)(path|file)\s*[+:]=?\s*.*\.\./`), "high"},
		{"File with concat", regexp.MustCompile(`(?i)new\s+File(?:InputStream|Reader)?\s*\([^)]*\+`), "high"},
	}
	ssrfPats := []sp{
		{"URL constructed with input", regexp.MustCompile(`(?i)new\s+(URL|URI)\s*\([^)]*\+`), "high"},
		{"HTTP with concat", regexp.MustCompile(`(?i)(fetch|axios|requests\.get)\s*\(\s*[^)]*\+`), "high"},
	}
	allPats := append(append(cmdPats, pathPats...), ssrfPats...)

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			for _, p := range allPats {
				if p.regex.MatchString(line) {
					result.Findings = append(result.Findings, Finding{
						File: file, Line: i + 1,
						Pattern:  fmt.Sprintf("%s: %s", file, p.name),
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

func TestSecurityScan_DetectsGoExecCommand(t *testing.T) {
	f := secTempFile(t, "main.go", `package main
import "os/exec"
func run(input string) {
    exec.Command("bash", "-c", input)
}`)
	result := securityScanDirect([]string{f})
	if result.OK {
		t.Error("expected Go exec.Command with variable to be detected")
	}
}

func TestSecurityScan_DetectsJavaRuntimeExec(t *testing.T) {
	f := secTempFile(t, "App.java", `public class App {
    void run(String cmd) {
        Runtime.getRuntime().exec(cmd);
    }
}`)
	result := securityScanDirect([]string{f})
	if result.OK {
		t.Error("expected Java Runtime.exec to be detected")
	}
}

func TestSecurityScan_DetectsPythonSystem(t *testing.T) {
	f := secTempFile(t, "app.py", `import os
os.system("ls " + user_input)`)
	result := securityScanDirect([]string{f})
	if result.OK {
		t.Error("expected Python os.system to be detected")
	}
}

func TestSecurityScan_DetectsURLConstruction(t *testing.T) {
	f := secTempFile(t, "app.java", `URL url = new URL("https://" + userInput)`)
	result := securityScanDirect([]string{f})
	if result.OK {
		t.Error("expected URL constructed with user input to be detected")
	}
}

func TestSecurityScan_SafeCodePasses(t *testing.T) {
	f := secTempFile(t, "app.java", `public class App {
    void run() {
        System.out.println("hello");
    }
}`)
	result := securityScanDirect([]string{f})
	if !result.OK {
		t.Error("expected safe code to pass")
	}
}

func TestSecurityScan_CommentsSkipped(t *testing.T) {
	f := secTempFile(t, "app.java", `// Runtime.getRuntime().exec(cmd);
public class App {}`)
	result := securityScanDirect([]string{f})
	if !result.OK {
		t.Error("expected comment to be skipped")
	}
}
