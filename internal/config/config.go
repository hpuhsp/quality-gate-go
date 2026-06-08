// Package config provides configuration loading for quality-gate.
// Reads .quality-gate.yaml from the repo root, with sensible defaults.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config represents the quality-gate configuration.
type Config struct {
	Version  int          `yaml:"version"`
	Secret   GateConfig   `yaml:"secret"`
	Syntax   GateConfig   `yaml:"syntax"`
	Format   FormatConfig `yaml:"format"`
	Lint     GateConfig   `yaml:"lint"`
	Security GateConfig   `yaml:"security"`
	Arch     ArchConfig   `yaml:"architecture"`
}

type GateConfig struct {
	Enabled bool `yaml:"enabled"`
}

type FormatConfig struct {
	Enabled bool `yaml:"enabled"`
	AutoFix bool `yaml:"auto_fix"`
}

type ArchConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Forbidden []string `yaml:"forbidden"`
}

// DefaultConfig returns sensible defaults (all gates enabled).
func DefaultConfig() Config {
	return Config{
		Version:  1,
		Secret:   GateConfig{Enabled: true},
		Syntax:   GateConfig{Enabled: true},
		Format:   FormatConfig{Enabled: true, AutoFix: false},
		Lint:     GateConfig{Enabled: false}, // disabled until linters installed
		Security: GateConfig{Enabled: true},
		Arch:     ArchConfig{Enabled: false},
	}
}

// Load reads .quality-gate.yaml from root, falling back to defaults.
func Load(root string) Config {
	cfg := DefaultConfig()

	configPath := filepath.Join(root, ".quality-gate.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg
	}

	f, err := os.Open(configPath)
	if err != nil {
		return cfg
	}
	defer f.Close()

	// Minimal YAML parser — handles flat keys, top-level sections, and inline comments
	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Strip inline comments (e.g., "enabled: true  # comment")
		if idx := strings.Index(line, " #"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header (e.g., "secret:")
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") {
			section = strings.TrimSuffix(line, ":")
			continue
		}

		// Key-value (e.g., "enabled: true")
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch section {
		case "secret":
			if key == "enabled" {
				cfg.Secret.Enabled = parseBool(val)
			}
		case "syntax":
			if key == "enabled" {
				cfg.Syntax.Enabled = parseBool(val)
			}
		case "format":
			if key == "enabled" {
				cfg.Format.Enabled = parseBool(val)
			}
			if key == "auto_fix" {
				cfg.Format.AutoFix = parseBool(val)
			}
		case "lint":
			if key == "enabled" {
				cfg.Lint.Enabled = parseBool(val)
			}
		case "security":
			if key == "enabled" {
				cfg.Security.Enabled = parseBool(val)
			}
		case "architecture":
			if key == "enabled" {
				cfg.Arch.Enabled = parseBool(val)
			}
			if key == "forbidden" {
				// Parse inline array: [controller->repository, ui->database]
				// NOTE: only inline [a, b] format is supported, not YAML dash-lists
				inner := strings.Trim(val, "[]")
				if inner != "" {
					cfg.Arch.Forbidden = strings.Split(inner, ",")
					for i, s := range cfg.Arch.Forbidden {
						cfg.Arch.Forbidden[i] = strings.TrimSpace(s)
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "quality-gate: config scan error: %v\n", err)
	}

	return cfg
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	b, _ := strconv.ParseBool(s)
	return b
}
