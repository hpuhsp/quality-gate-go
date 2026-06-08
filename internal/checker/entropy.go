package checker

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// Entropy-based secret detection.
// High-entropy strings (entropy > 4.5 bits/char, length > 20) are likely secrets.
// This catches secrets that don't match any known pattern (e.g., base64-encoded keys).

const (
	entropyThreshold = 4.5 // bits per character
	minEntropyLen    = 20  // minimum string length to consider
)

// entropyStrings extracts high-entropy strings from source code.
// Returns findings for strings that look like secrets but match no known pattern.
func EntropyScan(content string, file string) []Finding {
	var findings []Finding

	// Match quoted strings (single, double, backtick)
	re := regexp.MustCompile(`["\x60]([A-Za-z0-9+/=_\-]{` + fmt.Sprintf("%d", minEntropyLen) + `,})["\x60]`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, m := range matches {
		s := m[1]
		// Skip common non-secret patterns
		if isCommonNonSecret(s) {
			continue
		}
		// Calculate Shannon entropy
		e := shannonEntropy(s)
		if e >= entropyThreshold {
			lineNum := findLineNumber(content, m[0])
			findings = append(findings, Finding{
				File:     file,
				Line:     lineNum,
				Pattern:  fmt.Sprintf("High-entropy string (entropy=%.1f, len=%d)", e, len(s)),
				Severity: "high",
			})
		}
	}

	return findings
}

// shannonEntropy calculates Shannon entropy in bits per character.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, c := range s {
		freq[c]++
	}
	length := float64(len([]rune(s)))
	entropy := 0.0
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// isCommonNonSecret filters out strings that are high-entropy but not secrets.

var (
	reUUID       = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	reGitSHA     = regexp.MustCompile(`(?i)^[0-9a-f]{40}$`)
	reHexShort   = regexp.MustCompile(`^[0-9a-fA-F]{16,64}$`)
	reDomain     = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.[a-zA-Z]{2,}(\.[a-zA-Z]{2,})?$`)
	reHexOrMixed = regexp.MustCompile(`[a-fA-F]{4,}`)
)

func isCommonNonSecret(s string) bool {
	// UUIDs: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if reUUID.MatchString(s) {
		return true
	}
	// Git SHAs
	if reGitSHA.MatchString(s) {
		return true
	}
	// Base64-encoded images/data (very long, starts with known prefixes)
	if len(s) > 100 && strings.HasPrefix(s, "iVBOR") {
		return true
	}
	// Hex strings (color codes, hashes)
	if reHexOrMixed.MatchString(s) && len(s) < 64 {
		return true
	}
	// URL-safe strings with dots (domain names encoded)
	// Only skip if it looks like a real domain (word.tld pattern, not arbitrary dots)
	if reDomain.MatchString(s) {
		return true
	}
	return false
}

func findLineNumber(content, substr string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, substr) {
			return i + 1
		}
	}
	return 0
}
