package checker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// Pre-compiled regex patterns (package level, not recompiled per call)
var (
	tagRe = regexp.MustCompile(`</?([a-zA-Z][\w-]*)`)
	// Match empty catch blocks including those with only comments inside
	emptyCatchRe = regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*(?://[^\n]*)?\s*\}`)
	ifdefRe      = regexp.MustCompile(`(?m)^#ifdef|#ifndef|#if\b`)
	endifRe      = regexp.MustCompile(`(?m)^#endif`)
)

var extToLang = map[string]string{
	".java": "java", ".kt": "kotlin", ".kts": "kotlin",
	".js": "js", ".jsx": "js", ".ts": "ts", ".tsx": "ts",
	".vue": "vue",
	".cs":  "csharp",
	".cpp": "cpp", ".cc": "cpp", ".cxx": "cpp", ".h": "cpp", ".hpp": "cpp",
	".go":    "go",
	".swift": "swift",
	".m":     "objc", ".mm": "objc",
}

func getStagedSrc() []string {
	files := GetStagedFiles()
	var src []string
	// Ignore obvious binary/image extensions
	ignored := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".zip": true, ".jar": true, ".exe": true, ".dll": true, ".so": true,
		".class": true, ".bin": true, ".tar": true, ".gz": true,
	}
	for _, f := range files {
		ext := filepath.Ext(f)
		if !ignored[ext] {
			src = append(src, f)
		}
	}
	return src
}

func SyntaxCheck() CheckResult {
	result := CheckResult{OK: true}
	files := getStagedSrc()
	if len(files) == 0 {
		return result
	}

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file))
		lang := extToLang[ext]

		var issues []string
		if lang == "" {
			issues = checkBracketsAndQuotes(file)
			if len(issues) == 0 {
				result.Skipped = append(result.Skipped, file)
			}
		} else {
			issues = checkFile(file, lang)
		}

		for _, issue := range issues {
			result.Findings = append(result.Findings, Finding{
				File: file, Line: 0, Pattern: issue, Severity: "error",
			})
			result.OK = false
		}
	}

	return result
}

func checkFile(file, lang string) []string {
	var issues []string
	// Skip files larger than 1MB
	if info, err := os.Stat(file); err != nil || info.Size() > 1<<20 {
		return nil
	}
	data, err := shared.SafeReadFile(file)
	if err != nil {
		return nil
	}
	content := string(data)

	// Universal bracket check (skip for Go — gofmt handles it better,
	// and regex-heavy Go source can false-positive on bracket matching).
	if lang != "go" {
		issues = append(issues, checkBrackets(content, file)...)
	}

	switch lang {
	case "js", "ts":
		issues = append(issues, checkNodeSyntax(file)...)
	case "vue":
		issues = append(issues, checkVue(content, file)...)
	case "java", "kotlin", "csharp":
		issues = append(issues, checkJVM(content, file)...)
	case "cpp":
		issues = append(issues, checkCpp(content, file)...)
	case "go":
		issues = append(issues, checkGo(file)...)
	case "objc":
		issues = append(issues, checkObjC(content, file)...)
	}
	return issues
}

func checkBrackets(content, file string) []string {
	var issues []string
	stack := make([]rune, 0)
	pairs := map[rune]rune{'{': '}', '[': ']', '(': ')'}
	inLineComment := false
	inBlockComment := false
	inString := false
	prev := rune(0)

	for _, ch := range content {
		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			prev = ch
			continue
		}
		if inBlockComment {
			if prev == '*' && ch == '/' {
				inBlockComment = false
			}
			prev = ch
			continue
		}
		if inString {
			if ch == '"' && prev != '\\' {
				inString = false
			}
			prev = ch
			continue
		}
		if prev == '/' && ch == '/' {
			inLineComment = true
			prev = ch
			continue
		}
		if prev == '/' && ch == '*' {
			inBlockComment = true
			prev = ch
			continue
		}
		if ch == '"' {
			inString = true
			prev = ch
			continue
		}
		prev = ch

		if ch == '{' || ch == '[' || ch == '(' {
			stack = append(stack, ch)
		} else if ch == '}' || ch == ']' || ch == ')' {
			if len(stack) == 0 {
				issues = append(issues, fmt.Sprintf("%s: unexpected '%c'", file, ch))
				break
			}
			last := stack[len(stack)-1]
			if pairs[last] != ch {
				issues = append(issues, fmt.Sprintf("%s: mismatched bracket — expected '%c', got '%c'", file, pairs[last], ch))
				break
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		parts := make([]string, len(stack))
		for i, r := range stack {
			parts[i] = fmt.Sprintf("'%c'", r)
		}
		issues = append(issues, fmt.Sprintf("%s: unmatched brackets: %s (%d unclosed)", file, strings.Join(parts, ", "), len(stack)))
	}
	return issues
}

func checkNodeSyntax(file string) []string {
	var issues []string
	cmd := exec.Command("node", "--check", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		for _, l := range strings.Split(string(out), "\n") {
			if strings.Contains(l, "SyntaxError") {
				issues = append(issues, fmt.Sprintf("%s: %s", file, strings.TrimSpace(l)))
				break
			}
		}
	}
	return issues
}

func checkVue(content, file string) []string {
	var issues []string
	// Check template tags
	selfClose := map[string]bool{"br": true, "hr": true, "img": true, "input": true, "meta": true, "link": true}
	tagStack := make([]string, 0)
	for _, m := range tagRe.FindAllStringSubmatch(content, -1) {
		tag := m[0]
		name := strings.ToLower(m[1])
		if strings.HasPrefix(tag, "</") {
			if len(tagStack) > 0 && tagStack[len(tagStack)-1] == name {
				tagStack = tagStack[:len(tagStack)-1]
			}
		} else if !selfClose[name] {
			tagStack = append(tagStack, name)
		}
	}
	if len(tagStack) > 0 {
		issues = append(issues, fmt.Sprintf("%s: Vue template — possibly unclosed tags: %s", file, strings.Join(tagStack, ", ")))
	}
	return issues
}

func checkJVM(content, file string) []string {
	var issues []string
	// Empty catch
	if matched := emptyCatchRe.MatchString(content); matched {
		issues = append(issues, fmt.Sprintf("%s: empty catch block(s)", file))
	}
	// Unclosed string — skip escaped quotes (\") when counting
	unescaped := strings.ReplaceAll(content, "\\\"", "")
	if strings.Count(unescaped, "\"")%2 != 0 {
		issues = append(issues, fmt.Sprintf("%s: unclosed string literal", file))
	}
	return issues
}

func checkCpp(content, file string) []string {
	var issues []string
	ifdefs := len(ifdefRe.FindAllString(content, -1))
	endifs := len(endifRe.FindAllString(content, -1))
	if ifdefs != endifs {
		issues = append(issues, fmt.Sprintf("%s: preprocessor imbalance (#if*= %d, #endif= %d)", file, ifdefs, endifs))
	}
	return issues
}

func checkGo(file string) []string {
	var issues []string
	// Try gofmt
	cmd := exec.Command("gofmt", "-e", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		for _, l := range strings.Split(string(out), "\n") {
			if strings.Contains(l, file) {
				issues = append(issues, fmt.Sprintf("%s: gofmt — %s", file, strings.TrimSpace(l)))
				break
			}
		}
	}
	return issues
}

func checkObjC(content, file string) []string {
	var issues []string
	implementations := strings.Count(content, "@implementation") + strings.Count(content, "@interface")
	ends := strings.Count(content, "@end")
	if implementations != ends {
		issues = append(issues, fmt.Sprintf("%s: @implementation/@interface (%d) vs @end (%d) mismatch", file, implementations, ends))
	}
	return issues
}

func checkBracketsAndQuotes(file string) []string {
	content, err := shared.SafeReadFile(file)
	if err != nil {
		return nil
	}
	str := string(content)
	var issues []string
	if strings.Count(str, "{") != strings.Count(str, "}") {
		issues = append(issues, "Unmatched braces {}")
	}
	if strings.Count(str, "[") != strings.Count(str, "]") {
		issues = append(issues, "Unmatched brackets []")
	}
	// simple single/double quotes check (skip escaped quotes)
	unescapedQ := strings.ReplaceAll(str, "\\\"", "")
	if strings.Count(unescapedQ, "\"")%2 != 0 {
		issues = append(issues, "Unmatched double quotes \"")
	}
	return issues
}
