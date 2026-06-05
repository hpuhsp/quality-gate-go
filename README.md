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
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-linux-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate
```

### Windows

```powershell
Invoke-WebRequest https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\quality-gate\quality-gate.exe"
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\quality-gate", "User")
```

## Quick Start

```bash
cd your-project
quality-gate setup          # Interactive wizard: detect project, install tools
quality-gate doctor         # Verify all dependencies
quality-gate enable         # Activate pre-commit hook
quality-gate status         # See what's active and detected
```

## 6 Quality Gates

| Gate | What it catches | Speed |
|------|----------------|:----:|
| 🔑 **Secret Scan** | Passwords, API keys, tokens + **Shannon entropy detection** (32+ patterns) | <1s |
| 📝 **Syntax Check** | Bracket mismatch, empty catch, preprocessor imbalance (10 languages) | <2s |
| 🛡️ **Security Scan** | SQL injection + Command Injection + Path Traversal + SSRF | <1s |
| ✨ **Auto-Format** | ktlint / prettier / google-java-format (auto-fixed on commit) | varies |
| 🔍 **Lint** (opt-in) | golangci-lint / PMD / Detekt / ESLint | <5s |
| 🏗️ **Architecture** (opt-in) | Layer dependency violations (e.g., controller→repository forbidden) | <1s |

## Supported Languages

Java · Kotlin · JavaScript · TypeScript · Go · C# · C++ · Vue · Swift · Objective-C · Python

## Configuration

Create `quality-gate.yaml` in your project root to customize gates:

```yaml
version: 1

secret:
  enabled: true          # Secret scan + entropy detection

syntax:
  enabled: true          # Syntax check (10 languages)

security:
  enabled: true          # SQL injection + CMD injection + Path traversal + SSRF

format:
  enabled: true          # Auto-format on commit
  auto_fix: false        # Set true to auto-fix, false to check-only

lint:
  enabled: false          # Language-specific linters (requires tools installed)

architecture:
  enabled: false
  forbidden:             # Architecture layer constraints
    - controller->repository
    - ui->database
    - domain->infrastructure

performance:
  max_duration: 3s       # Max commit hook duration
```

No config file needed — all gates run with sensible defaults.

## Commands

```
quality-gate setup        First-run wizard: detect project, install tools
quality-gate doctor       Check & auto-install dependencies
quality-gate enable       Activate pre-commit hook
quality-gate disable      Deactivate hooks
quality-gate status       Show status and project detection
quality-gate update       Check for latest version
quality-gate tool         Optional tools
  tool gen-tests          [experimental] AI-generate tests (needs ANTHROPIC_API_KEY)
```

## How It Works

`quality-gate enable` sets `git config core.hooksPath` to `~/.quality-gate/hooks/`. The hook scripts call `quality-gate pre-commit`, which runs all 6 gates on **staged files only**. No files are added to your project.

```bash
quality-gate disable      # Removes core.hooksPath — hooks stop running
```

## Philosophy

> Install once, use everywhere. Catch problems at the keyboard, not in CI.

- **Zero project files**: uses `git config core.hooksPath`, not `.pre-commit-config.yaml`
- **Zero runtime deps**: single Go binary, no Node.js/Python required
- **Configurable**: per-gate enable/disable via `quality-gate.yaml`
- **Fast**: all 6 gates complete in under 3 seconds for typical commits

## Team Config

```yaml
# quality-gate.yaml in your project root
# Commit this file — team shares the same rules
version: 1
security:
  enabled: true
lint:
  enabled: true
architecture:
  enabled: true
  forbidden:
    - controller->repository
    - ui->database
```

## Why quality-gate-go?

| | Husky | quality-gate-go |
|---|---|---|
| **Dependencies** | Node.js runtime required | Single Go binary |
| **Startup** | ~200ms | <10ms |
| **Multi-language** | JS/TS focused | 11 languages supported |
| **Secret detection** | Not built-in | 32 patterns + entropy |
| **Security rules** | Not built-in | SQL/CMD/Path/SSRF |
| **Architecture rules** | Not built-in | Layer constraints |
| **Config** | `.huskyrc` | `quality-gate.yaml` |

## Configuration Examples

- [Java / Spring Boot](examples/java/.quality-gate.yaml)
- [Android (Kotlin/Gradle)](examples/android/.quality-gate.yaml)
- [Vue / Frontend](examples/vue/.quality-gate.yaml)
- [Uni-app](examples/uni-app/package.json)
