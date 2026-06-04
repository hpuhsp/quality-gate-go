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
# amd64
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-linux-amd64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate

# arm64
curl -fsSL https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-linux-arm64 -o /usr/local/bin/quality-gate
chmod +x /usr/local/bin/quality-gate
```

### Windows

```powershell
# PowerShell（管理员运行）
Invoke-WebRequest https://github.com/hpuhsp/quality-gate-go/releases/latest/download/quality-gate-windows-amd64.exe -OutFile "$env:LOCALAPPDATA\quality-gate\quality-gate.exe"
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:LOCALAPPDATA\quality-gate", "User")
```

或从 [releases 页面](https://github.com/hpuhsp/quality-gate-go/releases/latest) 下载。

### 源码编译

```bash
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

语法检查：Java · Kotlin · JavaScript · TypeScript · C# · C++ · Go · Swift · Objective-C · Vue

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

`quality-gate enable` 执行 `git config core.hooksPath ~/.quality-gate/hooks/`，钩子脚本在 commit 时调用 `quality-gate pre-commit`，**只扫描暂存文件**。项目目录**零文件增加**。

```bash
quality-gate disable      # 取消 core.hooksPath，钩子停止运行
```

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
\n## 🚀 新特性 (Develop)\n- **项目级配置:** 支持 `.quality-gate.yaml`。\n- **规则同步:** 支持远程拉取最新安全规则。\n- **性能优化:** 仅扫描增量文件。\n- **自动修复:** 可在 setup 阶段开启自动格式化修复。
