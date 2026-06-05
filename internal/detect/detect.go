// Package detect — Project type auto-detection (mirrors CI detect.sh logic)
package detect

import (
	"github.com/hpuhsp/quality-gate-go/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

// Result holds detected project metadata.
type Result struct {
	Language      string
	BuildTool     string
	TestFramework string
	HasTests      bool
	HasDockerfile bool
}

// Detect scans the project root and returns language/build/test info.
func Detect(root string) Result {
	r := Result{Language: "unknown", BuildTool: "unknown", TestFramework: "unknown"}

	switch {
	case shared.FileExists(filepath.Join(root, "build.gradle.kts")) || hasExt(root, ".kt"):
		r.Language = "kotlin"
	case shared.FileExists(filepath.Join(root, "build.gradle")) || hasExt(root, ".java"):
		r.Language = "java"
	case shared.FileExists(filepath.Join(root, "package.json")):
		r.Language = "javascript"
	case shared.FileExists(filepath.Join(root, "go.mod")):
		r.Language = "go"
	}

	// Build tool
	switch r.Language {
	case "kotlin", "java":
		if shared.FileExists(filepath.Join(root, "gradlew")) {
			r.BuildTool = "gradle-wrapper"
		} else {
			r.BuildTool = "gradle"
		}
	case "javascript":
		switch {
		case shared.FileExists(filepath.Join(root, "pnpm-lock.yaml")):
			r.BuildTool = "pnpm"
		case shared.FileExists(filepath.Join(root, "yarn.lock")):
			r.BuildTool = "yarn"
		default:
			r.BuildTool = "npm"
		}
	case "go":
		r.BuildTool = "go"
	}

	// Test framework
	switch r.Language {
	case "kotlin", "java":
		r.TestFramework = "JUnit5"
		r.HasTests = dirExists(filepath.Join(root, "src", "test"))
	case "javascript":
		r.TestFramework = detectJSTestFramework(root)
		r.HasTests = dirExists(filepath.Join(root, "test")) ||
			dirExists(filepath.Join(root, "__tests__")) ||
			dirExists(filepath.Join(root, "src", "__tests__"))
	case "go":
		r.TestFramework = "testing"
		r.HasTests = hasGoTestFiles(root)
	}

	// Dockerfile
	r.HasDockerfile = shared.FileExists(filepath.Join(root, "Dockerfile")) || shared.FileExists(filepath.Join(root, filepath.Join("docker", "Dockerfile")))

	return r
}

func detectJSTestFramework(root string) string {
	data, err := shared.SafeReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return "unknown"
	}
	s := string(data)
	switch {
	case strings.Contains(s, `"vitest"`):
		return "Vitest"
	case strings.Contains(s, `"jest"`):
		return "Jest"
	case strings.Contains(s, `"mocha"`):
		return "Mocha"
	}
	return "unknown"
}

func hasGoTestFiles(root string) bool {
	found := false
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if strings.HasSuffix(p, "_test.go") {
			found = true
		}
		return nil
	})
	return found
}



func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func hasExt(root, ext string) bool {
	found := false
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == ".gradle") {
			return filepath.SkipDir
		}
		if strings.HasSuffix(p, ext) {
			found = true
		}
		return nil
	})
	return found
}
