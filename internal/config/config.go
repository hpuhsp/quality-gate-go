package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var ConfigHome = filepath.Join(os.Getenv("HOME"), ".quality-gate")

type Config struct {
	MinCoverage   string `yaml:"minCoverage"`
	AutoFormat    string `yaml:"autoFormat"`
	RunTestsOnPush string `yaml:"runTestsOnPush"`
	Remote        string `yaml:"remote"`
	RemoteRef     string `yaml:"remoteRef"`
	CustomHooks   map[string]string `yaml:"customHooks"`
}

func Load() Config {
	c := Config{MinCoverage: "60", AutoFormat: "true", RunTestsOnPush: "no"}

	// Local config
	localPath := filepath.Join(ConfigHome, "config.yml")
	if data, err := os.ReadFile(localPath); err == nil {
		var local Config
		if yaml.Unmarshal(data, &local) == nil {
			if local.MinCoverage != "" {
				c.MinCoverage = local.MinCoverage
			}
			if local.AutoFormat != "" {
				c.AutoFormat = local.AutoFormat
			}
			if local.RunTestsOnPush != "" {
				c.RunTestsOnPush = local.RunTestsOnPush
			}
			if local.Remote != "" {
				c.Remote = local.Remote
			}
			if local.RemoteRef != "" {
				c.RemoteRef = local.RemoteRef
			}
		}
	}

	
	// Project config overlay
	if projData, err := os.ReadFile(".quality-gate.yaml"); err == nil {
		var projCfg Config
		if yaml.Unmarshal(projData, &projCfg) == nil {
			if len(projCfg.CustomHooks) > 0 {
				c.CustomHooks = projCfg.CustomHooks
			}
		}
	}

	// Remote config overlay
	remote := os.Getenv("QG_REMOTE_REPO")
	if remote == "" {
		remote = c.Remote
	}
	if remote != "" && isValidURL(remote) {
		fetchRemote(remote, c.RemoteRef, &c)
	}

	return c
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
		func() {
			cmd := exec.Command("git", "pull", "--ff-only")
			cmd.Dir = dir
			cmd.Run()
		}()
	} else {
		exec.Command("git", "clone", "--depth", "1", "--branch", ref, repo, dir).Run()
	}
	data, err := os.ReadFile(filepath.Join(dir, "quality-gate-config.yml"))
	if err != nil {
		return
	}
	var remoteCfg Config
	if yaml.Unmarshal(data, &remoteCfg) == nil {
		if remoteCfg.MinCoverage != "" {
			c.MinCoverage = remoteCfg.MinCoverage
		}
		if remoteCfg.AutoFormat != "" {
			c.AutoFormat = remoteCfg.AutoFormat
		}
	}
}
