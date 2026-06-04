# quality-gate

**Shift-Left Local Quality Gate** — pre-commit hooks that catch problems before they leave your machine. Single binary, zero dependencies, <10ms startup.

[中文文档](README_CN.md)

## Install

### macOS

```bash
# Intel
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# Apple Silicon
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-arm64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate
```

### Linux

```bash
# amd64
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-linux-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# arm64
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-linux-arm64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate
```

### Windows

```powershell
# PowerShell (run as Administrator)
Invoke-WebRequest https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\quality-gate\quality-gate.exe"
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\quality-gate", "User")
```

Or download from [releases page](https://github.com/hpuhsp/quality-gate-go/releases/latest).

### Build from source

```bash
go install github.com/hpuhsp/quality-gate-go@latest
```

## Quick Start

```bash
cd your-project
quality-gate enable        # One-time setup: activates hooks
quality-gate status        # See what's active and detected
```

That's it. Every `git commit` now runs 4 gates.

## Gates

| Gate | What it catches | Speed |
|------|----------------|:----:|
| 🔑 Secret scan | Passwords, API keys, tokens, private keys (17 patterns) | <1s |
| 📝 Syntax check | Bracket mismatch, empty catch, preprocessor imbalance (10 languages) | <2s |
| 🛡️ SQL injection | String-concatenated SQL, raw Statement, template injection | <1s |
| ✨ Auto-format | ktlint / prettier / google-java-format (auto-fixed on commit) | varies |

## Supported Languages

Syntax check: Java · Kotlin · JavaScript · TypeScript · C# · C++ · Go · Swift · Objective-C · Vue

## Commands

```
quality-gate setup        First-run wizard
quality-gate update       Self-update to latest version
quality-gate enable       Activate 4-gate pre-commit hook
quality-gate disable      Deactivate hooks
quality-gate status       Show status and project detection
quality-gate tool         Optional tools
  tool gen-tests          AI-generate unit tests (needs ANTHROPIC_API_KEY)
```

## How It Works

`quality-gate enable` sets `git config core.hooksPath` to `~/.quality-gate/hooks/`. The hook scripts call `quality-gate pre-commit`, which runs all 4 gates on **staged files only**. No files are added to your project.

```bash
quality-gate disable      # Removes core.hooksPath — hooks stop running
```

## Philosophy

> Install once, use everywhere. Catch problems at the keyboard, not in CI.

- **Zero project files**: uses `git config core.hooksPath`, not `.pre-commit-config.yaml`
- **Zero runtime deps**: single Go binary, no Node.js/Python required
- **Deterministic**: all gates are regex/compiler-based, no LLM API calls in hooks
- **Fast**: all 4 gates complete in under 3 seconds for typical commits

## Team Config

Share rules across your team via a git repo:

```bash
export QG_REMOTE_REPO=https://gitlab.com/your-team/quality-gate-config.git
quality-gate enable
```

The remote `quality-gate-config.yml` is fetched and merged with local settings.

## 🚀 New Features in Develop
- **Project-level config:** `.quality-gate.yaml` support.
- **Auto-update rules:** Sync security rules from remote.
- **Staged files only:** Optimized performance.
- **Auto-fix:** Configurable via `quality-gate setup --auto-fix=true`.

## Why quality-gate-go? (vs. Husky)
While tools like **Husky** are excellent for pure frontend ecosystems, they require a Node.js environment to run. In a polyglot enterprise team (Java, Go, Python, Vue, Android), forcing every backend and mobile developer to install Node.js just for Git hooks creates immense friction.

**quality-gate-go** solves this:
- 🚀 **Zero Dependencies:** Distributed as a single compiled Go binary. No Node/NPM required.
- ⚡ **Lightning Fast:** <10ms startup time, offering a frictionless developer experience.
- 🌍 **Polyglot Ready:** Instantly deployable across backend, frontend, and mobile repositories without environmental headaches.
