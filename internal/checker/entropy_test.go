package checker

import (
	"testing"
)

func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minEntropy float64
		maxEntropy float64
	}{
		{"single char", "aaaaaa", 0.0, 0.1},
		{"two chars alternating", "ababab", 0.9, 1.1},
		{"high entropy base64", "xK9mP2vL5nQ8wZ3jR7tY1bF6gH4dS0aE", 4.0, 6.0},
		{"empty string", "", 0.0, 0.1},
		{"numeric", "1234567890", 3.0, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := shannonEntropy(tt.input)
			if e < tt.minEntropy || e > tt.maxEntropy {
				t.Errorf("shannonEntropy(%q) = %.2f, want [%.1f, %.1f]", tt.input, e, tt.minEntropy, tt.maxEntropy)
			}
		})
	}
}

func TestIsCommonNonSecret(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"UUID", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", true},
		{"hex short", "abcdef1234567890abcdef1234567890", true},
		{"domain", "example.com", true},
		{"high entropy secret", "xK9mP2vL5nQ8wZ3jR7tY1bF6gH4dS0aE", false},
		{"plain text", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCommonNonSecret(tt.input)
			if result != tt.expected {
				t.Errorf("isCommonNonSecret(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEntropyScan_FindsHighEntropy(t *testing.T) {
	content := `const token = "xK9mP2vL5nQ8wZ3jR7tY1bF6gH4dS0aE"`
	findings := EntropyScan(content, "test.go")

	found := false
	for _, f := range findings {
		if f.Severity == "high" {
			found = true
		}
	}
	if !found {
		t.Error("expected high-entropy finding for token")
	}
}

func TestEntropyScan_IgnoresNormalCode(t *testing.T) {
	content := `const name = "hello world"`
	findings := EntropyScan(content, "test.go")
	if len(findings) > 0 {
		t.Errorf("expected 0 findings for normal code, got %d", len(findings))
	}
}
