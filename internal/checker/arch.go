package checker

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hpuhsp/quality-gate-go/internal/config"
	"github.com/hpuhsp/quality-gate-go/internal/shared"
)

// Pre-compiled import extraction regexes (package level, compiled once).
var (
	reJavaImport = regexp.MustCompile(`(?m)^import\s+([\w.]+)`)
	reGoImport   = regexp.MustCompile(`"([\w./-]+)"`)
	reJSImport   = regexp.MustCompile(`(?m)(?:import|from)\s+['"]([^'"]+)['"]`)
)

// ArchCheck validates architectural layer constraints.
// Uses regex to detect class/layer from file paths and imports.
//
// Example rules:
//
//	controller->repository (forbidden)
//	ui->database (forbidden)
//	domain->infrastructure (forbidden)
func ArchCheck(staged []string, cfg config.ArchConfig) CheckResult {
	result := CheckResult{OK: true}
	if !cfg.Enabled || len(cfg.Forbidden) == 0 {
		return result
	}

	// Parse forbidden rules: "controller->repository"
	type rule struct {
		from string
		to   string
	}
	var rules []rule
	for _, f := range cfg.Forbidden {
		parts := strings.Split(f, "->")
		if len(parts) == 2 {
			rules = append(rules, rule{
				from: strings.TrimSpace(strings.ToLower(parts[0])),
				to:   strings.TrimSpace(strings.ToLower(parts[1])),
			})
		}
	}
	if len(rules) == 0 {
		return result
	}

	// Layer detection patterns (file path based)
	layerPatterns := map[string]*regexp.Regexp{
		"controller":     regexp.MustCompile(`(?i)(controller|handler|endpoint|resource)`),
		"service":        regexp.MustCompile(`(?i)(service|usecase|interactor)`),
		"repository":     regexp.MustCompile(`(?i)(repository|repo|dao|mapper)`),
		"domain":         regexp.MustCompile(`(?i)(domain|model|entity|aggregate)`),
		"infrastructure": regexp.MustCompile(`(?i)(infra|infrastructure|persistence|external)`),
		"ui":             regexp.MustCompile(`(?i)(view|component|page|widget|screen)`),
		"database":       regexp.MustCompile(`(?i)(database|db|migration|schema|sql)`),
	}

	// Classify each staged file into a layer
	type fileLayer struct {
		file  string
		layer string
	}
	var fileLayers []fileLayer
	for _, file := range staged {
		layer := classifyLayer(file, layerPatterns)
		if layer != "" {
			fileLayers = append(fileLayers, fileLayer{file: file, layer: layer})
		}
	}

	// For each file, check imports against forbidden rules
	for _, fl := range fileLayers {
		data, err := shared.SafeReadFile(fl.file)
		if err != nil {
			continue
		}
		content := string(data)
		imports := extractImports(content)

		for _, imp := range imports {
			impLayer := classifyImport(imp, layerPatterns)
			if impLayer == "" {
				continue
			}
			for _, r := range rules {
				if fl.layer == r.from && impLayer == r.to {
					result.Findings = append(result.Findings, Finding{
						File:     fl.file,
						Line:     0,
						Pattern:  fmt.Sprintf("Arch: %s → %s (forbidden)", fl.layer, impLayer),
						Severity: "warning",
					})
				}
			}
		}
	}

	if len(result.Findings) > 0 {
		result.OK = false
	}
	return result
}

func classifyLayer(path string, patterns map[string]*regexp.Regexp) string {
	lower := strings.ToLower(path)
	for layer, re := range patterns {
		if re.MatchString(lower) {
			return layer
		}
	}
	return ""
}

func classifyImport(imp string, patterns map[string]*regexp.Regexp) string {
	// Map import segments to layers
	segments := strings.Split(strings.ToLower(imp), ".")
	full := strings.Join(segments, " ")
	for layer, re := range patterns {
		if re.MatchString(full) {
			return layer
		}
	}
	return ""
}

func extractImports(content string) []string {
	var imports []string
	// Java/Kotlin: import com.example.controller.UserController
	for _, m := range reJavaImport.FindAllStringSubmatch(content, -1) {
		imports = append(imports, m[1])
	}
	// Go: import "github.com/user/repo/controller"
	for _, m := range reGoImport.FindAllStringSubmatch(content, -1) {
		if strings.Contains(m[1], ".") || strings.Contains(m[1], "/") {
			imports = append(imports, m[1])
		}
	}
	// JS/TS: import X from './controller/UserController'
	for _, m := range reJSImport.FindAllStringSubmatch(content, -1) {
		imports = append(imports, m[1])
	}
	return imports
}
