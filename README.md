# quality-gate

**Shift-Left Local Quality Gate** — pre-commit hooks that catch problems before they leave your machine. Single binary, zero dependencies, <10ms startup.

## Install

```bash
# macOS (amd64)
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# macOS (arm64 / Apple Silicon)
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-arm64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# Build from source (requires Go 1.22+)
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

## Languages

Syntax check supports: Java · Kotlin · JavaScript · TypeScript · C# · C++ · Go · Swift · Objective-C · Vue

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

## How it works

`quality-gate enable` sets `git config core.hooksPath` to `~/.quality-gate/hooks/`. The hook scripts call `quality-gate pre-commit`, which runs all 4 gates on staged files only. No files are added to your project.

```bash
quality-gate disable      # Removes core.hooksPath — hooks stop running
```

## Philosophy

> Install once, use everywhere. Catch problems at the keyboard, not in CI.

- **No project files**: uses `git config core.hooksPath`, not `.pre-commit-config.yaml`
- **No runtime deps**: single Go binary, no Node.js/Python required
- **Deterministic**: all gates are regex/compiler-based, no LLM API calls in hooks
- **Fast**: all 4 gates complete in under 3 seconds for typical commits

## Team config

Share rules across your team via a git repo:

```bash
export QG_REMOTE_REPO=https://gitlab.com/your-team/quality-gate-config.git
quality-gate enable
```

The remote `quality-gate-config.yml` is fetched and merged with local settings.

---

# quality-gate（中文）

**左移本地质量门禁** — 在代码离开你的机器之前就发现问题。单文件二进制，零依赖，<10ms 启动。

## 安装

```bash
# macOS Intel
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# macOS Apple Silicon
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-darwin-arm64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# 源码编译（需 Go 1.22+）
go install github.com/hpuhsp/quality-gate-go@latest
```

## 快速开始

```bash
cd your-project
quality-gate enable        # 一次性：激活钩子
quality-gate status        # 查看状态和项目检测
```

之后每次 `git commit` 自动执行 4 道闸门。

## 4 道闸门

| 闸门 | 检查内容 | 速度 |
|------|---------|:----:|
| 🔑 密钥扫描 | 密码、API Key、Token、私钥（17 条规则） | <1s |
| 📝 语法检查 | 括号不匹配、空 catch、预处理失衡（10 种语言） | <2s |
| 🛡️ SQL 注入 | 字符串拼接 SQL、裸 Statement、模板注入 | <1s |
| ✨ 自动格式化 | ktlint / prettier / google-java-format（自动修复） | 随工具 |

## 支持语言

语法检查支持：Java · Kotlin · JavaScript · TypeScript · C# · C++ · Go · Swift · Objective-C · Vue

## 命令

```
quality-gate setup        首次配置向导
quality-gate update       自更新到最新版
quality-gate enable       激活 4 道闸门
quality-gate disable      关闭
quality-gate status       查看状态和项目检测
quality-gate tool         可选工具
  tool gen-tests          AI 生成单元测试（需 ANTHROPIC_API_KEY）
```

## 原理

`quality-gate enable` 执行 `git config core.hooksPath ~/.quality-gate/hooks/`，钩子脚本在 commit 时调用 `quality-gate pre-commit`。项目目录**零文件增加**。

## 设计哲学

> 安装一次，到处使用。在敲键盘时就发现问题，不等 CI。

- **零文件侵入**：用 `git config core.hooksPath`，不污染项目目录
- **零运行时依赖**：单 Go 二进制，不需要 Node.js/Python
- **确定性**：全部基于正则/编译器检查，钩子中无 LLM API 调用
- **极快**：4 道闸门通常在 3 秒内完成

## 团队配置

通过 Git 仓库共享规则：

```bash
export QG_REMOTE_REPO=https://gitlab.com/your-team/quality-gate-config.git
quality-gate enable
```

远程 `quality-gate-config.yml` 与本地配置合并。
