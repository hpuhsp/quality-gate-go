package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ConfigHome is the quality-gate configuration directory.
var ConfigHome = filepath.Join(userHomeDir(), ".quality-gate")

type Config struct {
	MinCoverage    string `yaml:"minCoverage"`
	AutoFormat     string `yaml:"autoFormat"`
	RunTestsOnPush string `yaml:"runTestsOnPush"`
	Remote         string `yaml:"remote"`
	RemoteRef      string `yaml:"remoteRef"`
}

// Load reads configuration from local, project, and remote sources.
func Load() Config {
	c := Config{MinCoverage: "60", AutoFormat: "true", RunTestsOnPush: "no"}

	// Local config (~/.quality-gate/config.yml)
	localPath := filepath.Join(ConfigHome, "config.yml")
	if data, err := os.ReadFile(localPath); err == nil {
		var local Config
		if yaml.Unmarshal(data, &local) == nil {
			overlayConfig(&c, &local)
		}
	}

	// Remote config overlay (shared team config)
	remote := os.Getenv("QG_REMOTE_REPO")
	if remote == "" {
		remote = c.Remote
	}
	if remote != "" && isValidURL(remote) {
		fetchRemote(remote, c.RemoteRef, &c)
	}

	return c
}

func overlayConfig(dst, src *Config) {
	if src.MinCoverage != "" {
		dst.MinCoverage = src.MinCoverage
	}
	if src.AutoFormat != "" {
		dst.AutoFormat = src.AutoFormat
	}
	if src.RunTestsOnPush != "" {
		dst.RunTestsOnPush = src.RunTestsOnPush
	}
	if src.Remote != "" {
		dst.Remote = src.Remote
	}
	if src.RemoteRef != "" {
		dst.RemoteRef = src.RemoteRef
	}
}

func isValidURL(u string) bool {
	return strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") ||
		strings.HasPrefix(u, "git@") || strings.HasPrefix(u, "ssh://")
}

func fetchRemote(repo, ref string, c *Config) {
	if ref == "" {
		ref = "main"
	}
	dir := filepath.Join(ConfigHome, "remote-config")
	if _, err := os.Stat(dir); err == nil {
		cmd := exec.Command("git", "pull", "--ff-only")
		cmd.Dir = dir
		cmd.Run() // optional — failure is non-fatal
	} else {
		exec.Command("git", "clone", "--depth", "1", "--branch", ref, repo, dir).Run()
	}
	data, err := os.ReadFile(filepath.Join(dir, "quality-gate-config.yml"))
	if err != nil {
		return
	}
	var remoteCfg Config
	if yaml.Unmarshal(data, &remoteCfg) == nil {
		overlayConfig(c, &remoteCfg)
	}
}

func userHomeDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return os.Getenv("HOME")
}
