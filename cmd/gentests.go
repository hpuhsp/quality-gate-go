package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"path/filepath"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/detect"
)

// RunGenTests generates unit tests using Anthropic Claude API.
// Requires ANTHROPIC_API_KEY environment variable.
func RunGenTests(args ...string) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ ANTHROPIC_API_KEY not set.")
		fmt.Println("   export ANTHROPIC_API_KEY=sk-ant-...")
		os.Exit(1)
	}

	target := ""
	for i, a := range args {
		if a == "gen-tests" && i+1 < len(args) {
			target = args[i+1]
			break
		}
	}
	if target == "" && len(args) > 0 {
		last := args[len(args)-1]
		if last != "gen-tests" && last != "tool" && last != "quality-gate" {
			target = last
		}
	}

	proj := detect.Detect(".")
	if proj.Language == "unknown" {
		fmt.Println("❌ Unknown project type. Cannot generate tests.")
		os.Exit(1)
	}

	// Collect sources
	var sources []string
	if target != "" {
		data, err := shared.SafeReadFile(target)
		if err != nil {
			fmt.Printf("❌ Cannot read %s: %v\n", target, err)
			os.Exit(1)
		}
		sources = []string{fmt.Sprintf("### %s\n```%s\n%s\n```", target, proj.Language, string(data))}
	} else {
		// Get git diff of changed files
		out, _ := exec.Command("git", "diff", "--name-only", "--diff-filter=ACM").Output()
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f == "" || strings.Contains(f, "test") || strings.Contains(f, "Test") {
				continue
			}
			data, err := shared.SafeReadFile(f)
			if err != nil {
				continue
			}
			sources = append(sources, fmt.Sprintf("### %s\n```%s\n%s\n```", f, proj.Language, string(data)))
		}
	}
	if len(sources) == 0 {
		fmt.Println("ℹ️  No source files to generate tests for.")
		os.Exit(0)
	}

	// Determine test directory and framework
	testDir := "src/test/" + proj.Language
	if proj.Language == "javascript" {
		testDir = "test"
	} else if proj.Language == "go" {
		testDir = ""
	}
	ext := fileExt(proj.Language)

	prompt := fmt.Sprintf(`Generate unit tests for this %s code. Use %s framework.

## Requirements
- Cover all public functions/methods
- Include edge cases (null, empty, boundary)
- Tests must compile and run without modifications
- Return ONLY the test file code, no explanations

## Test file location: %s/

## Source code:
%s

Output format: a single code block with filename:
`+"```"+`%s filename:%s
// test code here
`+"```"+``, proj.Language, proj.TestFramework, testDir, strings.Join(sources, "\n\n"), proj.Language, "TestClassName."+ext)

	fmt.Printf("Generating tests for %d source file(s) (%s/%s)...\n", len(sources), proj.Language, proj.TestFramework)

	response, err := callAnthropic(apiKey, prompt)
	if err != nil {
		fmt.Printf("❌ API error: %v\n", err)
		os.Exit(1)
	}

	// Parse and write test files
	filesWritten := parseAndWriteTests(response, testDir, ext)
	if filesWritten == 0 {
		fmt.Println("⚠️  Could not extract test code from AI response.")
		fmt.Println("--- Response preview ---")
		if len(response) > 500 {
			response = response[:500] + "..."
		}
		fmt.Println(response)
		return
	}

	fmt.Printf("\n✅ %d test file(s) written:\n", filesWritten)
	fmt.Println("\nReview the generated tests, then:")
	fmt.Println("  git add " + testDir)
	fmt.Println("  git commit -m 'add AI-generated tests'")
	fmt.Println("  git push")
}

func fileExt(lang string) string {
	switch lang {
	case "kotlin":
		return "kt"
	case "java":
		return "java"
	case "javascript":
		return "js"
	case "go":
		return "go"
	default:
		return "txt"
	}
}

func callAnthropic(apiKey, prompt string) (string, error) {
	body := map[string]interface{}{
		"model":      "claude-sonnet-4-6",
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10MB cap
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(raw))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return result.Content[0].Text, nil
}

func parseAndWriteTests(response, testDir, ext string) int {
	written := 0
	os.MkdirAll(testDir, 0755)

	// Extract code blocks with filenames
	blocks := strings.Split(response, "```")
	for i := 1; i < len(blocks); i += 2 {
		block := blocks[i]
		// Extract filename if present
		lines := strings.SplitN(block, "\n", 2)
		header := lines[0]
		code := ""
		if len(lines) > 1 {
			code = lines[1]
		}

		fileName := ""
		if strings.Contains(header, "filename:") {
			parts := strings.SplitN(header, "filename:", 2)
			if len(parts) > 1 {
				fileName = strings.TrimSpace(parts[1])
			}
		}
		if fileName == "" {
			fileName = fmt.Sprintf("GeneratedTest%d.%s", written+1, ext)
		}
		if !strings.Contains(fileName, ".") {
			fileName = strings.TrimSuffix(fileName, "."+ext) + "_test." + ext
		}

		filePath := filepath.Join(testDir, fileName)
		// P0-1: Validate output path stays within testDir
		absTestDir, _ := filepath.Abs(testDir)
		absFilePath, _ := filepath.Abs(filePath)
		if !strings.HasPrefix(absFilePath, absTestDir+string(os.PathSeparator)) {
			fmt.Printf("  ⚠️  skipped: path traversal in filename %q\n", fileName)
			continue
		}
		if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
			fmt.Printf("  ⚠️  Failed to write %s: %v\n", filePath, err)
			continue
		}
		fmt.Printf("  ✅ %s\n", filePath)
		written++
	}
	return written
}
