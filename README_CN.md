# quality-gate

**左移本地质量门禁** — 在代码离开你的机器之前就发现问题。单文件二进制，零依赖，<10ms 启动。

[English Docs](README.md)

## 安装

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

## 快速开始

```bash
cd your-project
quality-gate setup          # 交互式向导：检测项目、安装工具
quality-gate doctor         # 验证依赖
quality-gate enable         # 激活 pre-commit 钩子
quality-gate status         # 查看状态
```

## 6 道质量闸门

| 闸门 | 检查内容 | 速度 |
|------|---------|:----:|
| 🔑 **密钥扫描** | 密码、API Key、Token + **Shannon 熵值检测**（18 规则） | <1s |
| 📝 **语法检查** | 括号不匹配、空 catch、预处理失衡（10 种语言） | <2s |
| 🛡️ **安全扫描** | SQL 注入 + 命令注入 + 路径穿越 + SSRF | <1s |
| ✨ **自动格式化** | ktlint / prettier / google-java-format（提交时自动修正） | 随工具 |
| 🔍 **Lint**（可选） | golangci-lint / PMD / Detekt / ESLint | <5s |
| 🏗️ **架构规则**（可选） | 层依赖约束（如 controller→repository 禁止） | <1s |

## 支持语言

Java · Kotlin · JavaScript · TypeScript · Go · C# · C++ · Vue · Swift · Objective-C

## 配置文件

在项目根目录创建 `.quality-gate.yaml` 自定义闸门：

```yaml
version: 1

secret:
  enabled: true          # 密钥扫描 + 熵值检测

syntax:
  enabled: true          # 语法检查（10 种语言）

security:
  enabled: true          # SQL 注入 + 命令注入 + 路径穿越 + SSRF

format:
  enabled: true          # 提交时自动格式化
  auto_fix: false        # true=自动修正，false=仅检查

lint:
  enabled: false          # 语言专用 Lint（需安装对应工具）

architecture:
  enabled: false
  forbidden: [controller->repository, ui->database, domain->infrastructure]  # 架构层约束

# 所有闸门使用合理默认值，无需配置文件。
```

无需配置文件——所有闸门使用合理默认值。

## 命令

```
quality-gate setup        首次配置向导：检测项目、安装工具
quality-gate doctor       检查并自动安装依赖
quality-gate enable       激活 pre-commit 钩子
quality-gate disable      关闭钩子
quality-gate status       查看状态和项目检测
quality-gate update       检查最新版本
quality-gate tool         可选工具
  tool gen-tests          [实验性] AI 生成测试（需 ANTHROPIC_API_KEY）
```

## 原理

`quality-gate enable` 设置 `git config core.hooksPath` 指向 `~/.quality-gate/hooks/`。钩子脚本调用 `quality-gate pre-commit`，对**暂存文件**执行 6 道闸门。项目目录零文件增加。

```bash
quality-gate disable      # 取消 core.hooksPath，钩子停止运行
```

## 设计哲学

> 安装一次，到处使用。在键盘前发现问题，不等 CI。

- **零文件侵入**：用 `git config core.hooksPath`，不污染项目目录
- **零运行时依赖**：单 Go 二进制，不需要 Node.js/Python
- **可配置**：每道闸门可通过 `.quality-gate.yaml` 独立开关
- **极快**：6 道闸门通常在 3 秒内完成

## 团队配置

```yaml
# .quality-gate.yaml 放在项目根目录
# 提交到仓库——团队共享同一套规则
version: 1
security:
  enabled: true
lint:
  enabled: true
architecture:
  enabled: true
  forbidden: [controller->repository, ui->database]
```

## 为什么选择 quality-gate-go？

| | Husky | quality-gate-go |
|---|---|---|
| **依赖** | 需要 Node.js 运行时 | 单 Go 二进制 |
| **启动** | ~200ms | <10ms |
| **多语言** | 专注 JS/TS | 支持 10 种语言 |
| **密钥检测** | 无内置 | 18 规则 + 熵值检测 |
| **安全规则** | 无内置 | SQL/命令注入/路径穿越/SSRF |
| **架构规则** | 无内置 | 层依赖约束 |
| **配置** | `.huskyrc` | `.quality-gate.yaml` |

## 配置示例

- [Java / Spring Boot](examples/java/.quality-gate.yaml)
- [Android (Kotlin/Gradle)](examples/android/.quality-gate.yaml)
- [Vue / 前端](examples/vue/.quality-gate.yaml)
- [iOS (Swift/ObjC)](examples/ios/.quality-gate.yaml)
