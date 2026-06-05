package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Secret.Enabled {
		t.Error("expected Secret.Enabled to be true by default")
	}
	if !cfg.Syntax.Enabled {
		t.Error("expected Syntax.Enabled to be true by default")
	}
	if !cfg.Security.Enabled {
		t.Error("expected Security.Enabled to be true by default")
	}
	if !cfg.Format.Enabled {
		t.Error("expected Format.Enabled to be true by default")
	}
	if cfg.Lint.Enabled {
		t.Error("expected Lint.Enabled to be false by default")
	}
	if cfg.Arch.Enabled {
		t.Error("expected Arch.Enabled to be false by default")
	}
	if cfg.Perf.MaxDuration != "3s" {
		t.Errorf("expected MaxDuration '3s', got '%s'", cfg.Perf.MaxDuration)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	tmp := t.TempDir()
	cfg := Load(tmp)
	if !cfg.Secret.Enabled {
		t.Error("expected defaults when no config file exists")
	}
}

func TestLoad_WithConfig(t *testing.T) {
	tmp := t.TempDir()
	yaml := `version: 1
secret:
  enabled: false
syntax:
  enabled: true
security:
  enabled: false
format:
  enabled: true
  auto_fix: true
`
	os.WriteFile(filepath.Join(tmp, "quality-gate.yaml"), []byte(yaml), 0644)

	cfg := Load(tmp)
	if cfg.Secret.Enabled {
		t.Error("expected Secret.Enabled to be false")
	}
	if !cfg.Syntax.Enabled {
		t.Error("expected Syntax.Enabled to be true")
	}
	if cfg.Security.Enabled {
		t.Error("expected Security.Enabled to be false")
	}
	if !cfg.Format.AutoFix {
		t.Error("expected Format.AutoFix to be true")
	}
}

func TestLoad_WithDotPrefix(t *testing.T) {
	tmp := t.TempDir()
	yaml := `version: 1
secret:
  enabled: false
`
	os.WriteFile(filepath.Join(tmp, ".quality-gate.yaml"), []byte(yaml), 0644)

	cfg := Load(tmp)
	if cfg.Secret.Enabled {
		t.Error("expected .quality-gate.yaml to be loaded")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"yes", false}, // not in ParseBool
		{"1", true},
		{"false", false},
		{"0", false},
		{"no", false},
	}
	for _, tt := range tests {
		result := parseBool(tt.input)
		if result != tt.expected {
			t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestLoad_ArchForbidden(t *testing.T) {
	tmp := t.TempDir()
	yaml := `version: 1
architecture:
  enabled: true
  forbidden: [controller->repository, ui->database]
`
	os.WriteFile(filepath.Join(tmp, "quality-gate.yaml"), []byte(yaml), 0644)

	cfg := Load(tmp)
	if !cfg.Arch.Enabled {
		t.Error("expected Arch.Enabled to be true")
	}
	if len(cfg.Arch.Forbidden) != 2 {
		t.Errorf("expected 2 forbidden rules, got %d", len(cfg.Arch.Forbidden))
	}
	if cfg.Arch.Forbidden[0] != "controller->repository" {
		t.Errorf("expected 'controller->repository', got '%s'", cfg.Arch.Forbidden[0])
	}
}
