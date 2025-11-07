# Todo-Go

[![CI](https://github.com/SongRunqi/go-todo/actions/workflows/ci.yml/badge.svg)](https://github.com/SongRunqi/go-todo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SongRunqi/go-todo)](https://goreportcard.com/report/github.com/SongRunqi/go-todo)
[![codecov](https://codecov.io/gh/SongRunqi/go-todo/branch/main/graph/badge.svg)](https://codecov.io/gh/SongRunqi/go-todo)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SongRunqi/go-todo)](go.mod)

[English](README.md) | [中文](README_zh.md)

一个功能强大的 AI 驱动的待办事项管理命令行应用，支持 Alfred 集成，使用 Go 语言构建。

## 功能特性

### 核心功能
- **AI 驱动的任务管理**：使用 LLM（默认为 DeepSeek）智能解析自然语言输入
- **Alfred 集成**：与 Alfred 工作流无缝集成（macOS 用户）
- **智能任务解析**：自动从自然语言中提取任务详情、截止日期和紧急程度
- **全面的任务操作**：
  - 创建、列表、获取、更新、完成、删除任务
  - 查看和管理已完成任务（备份）
  - 恢复已完成的任务到活动列表
- **详细描述**：AI 生成包含上下文和预期结果的综合任务描述
- **优先级管理**：基于截止日期自动计算紧急程度
- **基于时间的排序**：任务按到期日期排序，带有倒计时定时器
- **多种输出格式**：JSON（兼容 Alfred）和 Markdown 格式

### 开发者功能
- **🌍 国际化 (i18n)**：完整支持中文和英文
  - 自动从系统环境检测语言
  - 通过 `TODO_LANG` 环境变量切换语言
  - 所有用户界面文本完全翻译
- **🎨 彩色输出**：漂亮的终端彩色输出
  - ✓ 绿色表示成功消息
  - ✗ 红色表示错误并提供可操作建议
  - ⚠ 黄色表示警告
  - ℹ 青色表示信息
- **⚡ 性能优化**：全面的基准测试和优化
- **🔍 输入验证**：强大的验证层，提供清晰的错误消息
- **🧪 充分测试**：73%+ 的测试覆盖率，包含单元和集成测试
- **📊 结构化日志**：基于 Zerolog 的日志系统，可配置日志级别
- **🔌 可插拔的 AI 客户端**：抽象 AI 接口，支持多个 LLM 提供商
- **💾 内存存储**：用于测试的内存存储选项
- **🚀 CI/CD 流水线**：自动化测试、代码检查和多平台构建
- **🛠 Shell 补全**：支持 Bash、Zsh、Fish 和 PowerShell 自动补全

## 目录

- [安装](#安装)
- [快速开始](#快速开始)
- [配置](#配置)
- [使用方法](#使用方法)
- [跨平台构建](#跨平台构建)
- [开发](#开发)
- [测试](#测试)

## 安装

### 前置要求

- **Go 1.21 或更高版本** - [下载 Go](https://golang.org/dl/)
- **DeepSeek API 密钥**（或兼容的 LLM API）- [获取 API 密钥](https://platform.deepseek.com/)

检查你的 Go 版本：
```bash
go version  # 应该是 1.21 或更高版本
```

### 推荐：使用安装脚本

最简单的安装方法：

```bash
# 克隆仓库
git clone https://github.com/SongRunqi/go-todo.git
cd go-todo

# 运行安装脚本
chmod +x install.sh
./install.sh
```

脚本会：
- ✓ 构建优化的二进制文件
- ✓ 安装到 `~/.local/bin/todo`
- ✓ 初始化待办目录和配置
- ✓ 引导您选择语言

### 替代方法：使用 Makefile

```bash
# 克隆仓库
git clone https://github.com/SongRunqi/go-todo.git
cd go-todo

# 安装并初始化（推荐）
make init

# 或仅安装（不初始化）
make install

# 或仅构建（二进制文件在当前目录）
make build
```

运行 `make help` 查看所有可用命令。

### 手动安装

```bash
# 克隆仓库
git clone https://github.com/SongRunqi/go-todo.git
cd go-todo

# 下载依赖
go mod download

# 构建应用
go build -ldflags="-s -w" -o todo main.go

# 安装到 ~/.local/bin
mkdir -p ~/.local/bin
cp todo ~/.local/bin/
chmod +x ~/.local/bin/todo

# 如果还未添加到 PATH（添加到您的 shell 配置）
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# 初始化待办环境
todo init
todo list
```

## 快速开始

```bash
# 1. 初始化待办环境（如果还未完成）
todo init

# 2. 设置你的 API 密钥
export API_KEY="your-deepseek-api-key-here"

# 3. 设置首选语言（可选，初始化时会询问）
todo lang set zh    # 或 'en' 表示英文

# 4. 使用自然语言创建任务
todo "明天晚上买菜"

# 5. 列出所有任务
todo list

# 6. 完成任务
todo complete 1

# 7. 查看已完成的任务
todo back
```

## 配置

### 环境变量

应用使用以下环境变量：

#### 必需
- `API_KEY`：你的 DeepSeek API 密钥用于 LLM 功能（或使用 `DEEPSEEK_API_KEY`）

#### 可选
- `TODO_LANG`：设置界面语言（默认自动从系统检测）
  - 支持的值：`en`（英文）、`zh`（中文）
  - 如果未设置，会从 `LANGUAGE`、`LC_ALL`、`LC_MESSAGES` 或 `LANG` 自动检测
- `LLM_BASE_URL`：自定义 LLM API 端点（默认为 `https://api.deepseek.com/chat/completions`）
  - 用于切换到其他 LLM 提供商（OpenAI、Claude 等）
- `LLM_MODEL`：要使用的模型（默认为提供商的默认模型）
- `LOG_LEVEL`：日志级别 - `debug`、`info`、`warn`、`error`（默认：`info`）
- `NO_COLOR`：设置任意值以禁用彩色输出

### 配置示例

```bash
# 基本配置（添加到 ~/.bashrc 或 ~/.zshrc）
export API_KEY="your-api-key-here"

# 完整配置
export API_KEY="your-api-key-here"
export LLM_BASE_URL="https://api.deepseek.com/chat/completions"
export LLM_MODEL="deepseek-chat"
export LOG_LEVEL="info"
export TODO_LANG="zh"  # 或 "en" 使用英文

# 改用 OpenAI
export API_KEY="your-openai-api-key"
export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
export LLM_MODEL="gpt-4"
```

## 国际化 (i18n)

Todo-Go 支持多种语言，所有用户界面文本均可翻译。

### 支持的语言

- **中文 (zh)**：简体中文
- **English (en)**：英文

### 设置语言

使用 `lang` 命令设置您的首选语言。设置将保存到 `~/.todo/config.json` 并在所有命令中保持有效。

```bash
# 列出可用语言（Alfred 兼容的 JSON 格式）
./todo lang list

# 设置语言为中文
./todo lang set zh

# 设置语言为英文
./todo lang set en

# 查看当前语言
./todo lang current
```

### 自动检测

如果配置文件中未设置语言，应用程序将从以下环境变量（按顺序）自动检测系统语言：
1. `LANGUAGE`
2. `LC_ALL`
3. `LC_MESSAGES`
4. `LANG`

### 示例

**中文：**
```bash
$ ./todo lang set zh
$ ./todo --help
一个简单的命令行待办事项应用，支持自然语言输入和 AI 驱动的任务管理。
```

**English:**
```bash
$ ./todo lang set en
$ ./todo --help
A simple command-line TODO application that supports natural language input and AI-powered task management.
```

所有命令帮助、错误消息、验证消息和输出文本都将以您选择的语言显示。

## 使用方法

### 命令结构

Todo-Go 现在使用现代 CLI 框架（Cobra），命令结构清晰：

```bash
todo [命令] [参数] [标志]
```

**可用命令：**
- `list` / `ls` - 列出所有活动任务
- `get <id>` - 获取任务详细信息
- `complete <id>` - 标记任务为已完成
- `delete <id>` - 永久删除任务
- `update <内容>` - 使用 Markdown 或 JSON 更新任务
- `back` - 列出已完成的任务
- `back get <id>` - 查看已完成的任务
- `back restore <id>` - 恢复已完成的任务
- `help` - 获取任何命令的帮助
- `completion` - 生成 shell 补全脚本

**全局标志：**
- `--config` - 指定配置文件位置
- `--verbose` / `-v` - 启用详细输出
- `--help` / `-h` - 显示任何命令的帮助

**环境变量：**
- `LOG_LEVEL` - 设置日志级别（debug、info、warn、error）- 默认为 "info"
- `NO_COLOR` - 设置后禁用彩色输出

**自然语言（AI）：** 如果不使用特定命令，你的输入将被视为自然语言并由 AI 处理。

### Shell 补全

生成 shell 补全脚本以加快命令输入：

```bash
# Bash
todo completion bash > /etc/bash_completion.d/todo

# Zsh
todo completion zsh > "${fpath[1]}/_todo"

# Fish
todo completion fish > ~/.config/fish/completions/todo.fish

# PowerShell
todo completion powershell > todo.ps1
```

所有命令都使用 `./todo` 可执行文件（如果全局安装则使用 `todo`）。

### 创建任务

使用自然语言创建单个或多个任务：

```bash
# 单个任务
./todo "明天晚上之前买菜"

# 多个任务（用分号分隔）
./todo "周五之前写报告; 明天给客户打电话; 本周末之前审查代码"
```

AI 将自动：
- 提取任务名称
- 生成带有上下文的详细描述
- 设置截止日期和紧急程度
- 计算剩余时间

### 列出任务

显示所有活动任务：

```bash
./todo list
# 或简写
./todo ls
```

输出格式为 Alfred 兼容的 JSON，包括：
- **任务 ID**：`[1]` 前缀便于引用
- **任务名称**：带有表情符号指示器 🎯
- **紧急状态**：剩余时间或逾期指示器
- **描述**：详细的任务上下文，带有状态表情符号（⌛️ 待处理，✅ 已完成）

### 查看已完成的任务（备份）

列出所有已完成/存档的任务：

```bash
./todo back
```

### 获取任务详情

检索特定任务的详细信息：

```bash
# 获取活动任务
./todo get <task-id>

# 从备份获取已完成的任务
./todo "back get <task-id>"
```

示例输出（Markdown 格式）：
```markdown
# 任务名称

- **任务 ID：** 1
- **状态：** 待处理
- **用户：** 用户名
- **截止日期：** 2025-11-05
- **紧急程度：** 高
- **创建时间：** 2025-11-02 10:30:00
- **结束时间：** 2025-11-05 18:00:00

## 描述

任务描述在这里...
```

### 完成任务

将任务标记为已完成（移动到备份）：

```bash
./todo complete 1
```

已完成的任务会存档到备份文件中，并从活动列表中删除。

### 恢复已完成的任务

从备份中恢复已完成的任务到活动列表：

```bash
./todo "back restore <task-id>"
```

任务状态将从"已完成"变为"待处理"。

### 更新任务

使用 Markdown 或 JSON 格式更新现有任务：

```bash
# 使用 Markdown（推荐）
./todo update "# 更新的任务名称

- **任务 ID：** 1
- **状态：** 待处理
- **用户：** 用户名
- **截止日期：** 2025-11-10
- **紧急程度：** 高

## 描述

更新的任务描述..."

# 使用 JSON
./todo update '{"taskId":1,"taskName":"更新的任务","taskDesc":"新描述",...}'
```

### 删除任务

永久删除任务：

```bash
./todo delete 1
```

### 语言管理

管理应用程序的语言设置：

```bash
# 列出可用语言（Alfred 兼容的 JSON 格式）
./todo lang list

# 设置首选语言
./todo lang set en   # 英文
./todo lang set zh   # 中文

# 显示当前语言
./todo lang current
```

语言偏好将保存到 `~/.todo/config.json` 并在所有命令中保持有效。更多详情请参阅[国际化](#国际化-i18n)部分。

## 跨平台构建

### 为当前平台构建

```bash
# 标准构建
go build -o todo main.go

# 优化构建（更小的二进制文件）
go build -ldflags="-s -w" -o todo main.go
```

### 跨平台构建

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o todo-linux-amd64 main.go

# Linux (arm64) - 适用于树莓派、ARM 服务器
GOOS=linux GOARCH=arm64 go build -o todo-linux-arm64 main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o todo-darwin-amd64 main.go

# macOS (Apple Silicon - M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o todo-darwin-arm64 main.go

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o todo-windows-amd64.exe main.go
```

### 为所有平台构建的脚本

创建一个 `build-all.sh` 脚本：

```bash
#!/bin/bash
platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name="todo-${GOOS}-${GOARCH}"

    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    echo "正在构建 $output_name..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $output_name main.go
done

echo "所有构建完成！"
```

运行它：
```bash
chmod +x build-all.sh
./build-all.sh
```

## AI 驱动的功能

### 智能任务描述生成

LLM 生成包含以下内容的综合描述：
1. **需要做什么**：具体的行动项
2. **为什么重要**：目的和预期结果
3. **相关详情**：从你的输入中获取的依赖项、约束或上下文

### 智能意图识别

AI 自动检测你的意图：
- `create`：添加新任务
- `list`：查看所有任务
- `complete`：标记任务为已完成
- `delete`：删除任务

### 紧急程度计算

任务根据截止日期自动分配紧急程度级别：
- `urgent`：非常短的时间框架
- `high`：即将到期
- `medium`：正常时间框架（默认）
- `low`：遥远的截止日期

## Alfred 集成

### Alfred 项目格式

每个任务在 Alfred 中显示为：
- **标题**：`[任务ID] 🎯 任务名称 [紧急状态]`
- **副标题**：`[状态] 任务描述`
- **Arg**：任务 ID 用于下游操作

示例：
```
[1] 🎯 买菜 还有2h 30m 截止
⌛️ 购买新鲜蔬菜和水果...
```

### 状态指示器

- ⌛️ 待处理任务
- ✅ 已完成任务
- 还有 X时 X分 截止：剩余时间
- 已截止：逾期

## 文件存储

任务存储在 JSON 文件中：
- **活动任务**：`todos.json`（或自定义位置）
- **已完成任务**：备份文件用于存档

## 开发

### 项目结构

```
go-todo/
├── main.go                      # 应用程序入口点
├── cmd/                         # 命令行界面（Cobra）
│   ├── root.go                 # 根命令和补全
│   ├── list.go                 # 列表命令
│   ├── get.go                  # 获取命令
│   ├── complete.go             # 完成命令
│   ├── delete.go               # 删除命令
│   ├── update.go               # 更新命令
│   └── back.go                 # 备份命令
├── app/                         # 业务逻辑
│   ├── command.go              # 核心任务操作
│   ├── commands.go             # 命令实现
│   ├── api.go                  # AI 客户端包装器
│   ├── storage.go              # 文件存储
│   ├── utils.go                # 实用函数
│   ├── types.go                # 数据模型
│   └── router.go               # 命令路由器
├── parser/                      # 任务解析
│   ├── parser.go               # Markdown/JSON 解析器
│   └── parser_test.go          # 解析器测试（94.6% 覆盖率）
├── internal/                    # 内部包
│   ├── i18n/                   # 国际化
│   │   ├── i18n.go            # i18n 包（嵌入式翻译）
│   │   └── translations/      # 翻译文件
│   │       ├── en.json        # 英文翻译
│   │       └── zh.json        # 中文翻译
│   ├── logger/                 # 结构化日志（zerolog）
│   ├── validator/              # 输入验证
│   ├── ai/                     # AI 客户端抽象
│   │   ├── client.go          # 接口定义
│   │   ├── deepseek.go        # DeepSeek 实现
│   │   └── mock.go            # 测试模拟
│   ├── storage/                # 存储实现
│   │   └── memory.go          # 内存存储
│   └── output/                 # 终端输出
│       ├── color.go           # 彩色输出
│       └── spinner.go         # 进度指示器
├── .github/workflows/           # CI/CD 流水线
│   └── ci.yml                  # GitHub Actions
├── .golangci.yml               # Linter 配置
├── go.mod                      # Go 模块依赖
├── go.sum                      # 依赖校验和
├── ROADMAP.md                  # 开发路线图
└── README.md                   # 英文文档
```

### 技术栈

- **CLI 框架**：[Cobra](https://github.com/spf13/cobra) - 现代命令行界面
- **日志**：[Zerolog](https://github.com/rs/zerolog) - 高性能结构化日志
- **颜色**：[Fatih Color](https://github.com/fatih/color) - 终端彩色输出
- **进度条**：[Briandowns Spinner](https://github.com/briandowns/spinner) - 进度指示器
- **AI 提供商**：DeepSeek API（可配置其他 LLM 提供商）
- **测试**：Go 标准测试库，表驱动测试
- **CI/CD**：GitHub Actions，多平台构建

### 按包分类的主要功能

**cmd/** - 命令行界面
- 基于 Cobra 的命令结构
- Shell 补全生成
- 自然语言回退

**app/** - 核心业务逻辑
- 任务 CRUD 操作
- AI 意图检测
- 基于文件的持久化

**parser/** - 任务解析
- 任务更新的 Markdown 解析器
- 结构化输入的 JSON 解析器
- 自动格式检测

**internal/logger** - 结构化日志
- 可配置的日志级别
- 彩色控制台输出
- 错误跟踪

**internal/validator** - 输入验证
- 任务 ID 验证
- 字段长度检查
- 状态和紧急程度验证

**internal/ai** - AI 客户端抽象
- 基于接口的设计
- DeepSeek 实现
- 用于测试的模拟客户端

**internal/output** - 终端输出
- 彩色成功/错误消息
- 进度旋转器
- 可操作的错误建议

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行带详细输出的测试
go test -v ./...

# 运行带覆盖率的测试
go test -cover ./...

# 运行带竞态检测的测试
go test -race ./...
```

### 生成覆盖率报告

```bash
# 生成覆盖率配置文件
go test -coverprofile=coverage.out ./...

# 在终端查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 覆盖率报告
go tool cover -html=coverage.out -o coverage.html

# 在浏览器中打开
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### 运行特定测试

```bash
# 测试特定包
go test ./app/
go test ./parser/

# 运行特定测试函数
go test -run TestCreateTask ./app/

# 运行匹配模式的测试
go test -run "TestCreate.*" ./app/
```

### 基准测试

```bash
# 运行所有基准测试
go test -bench=. ./...

# 运行带内存统计的基准测试
go test -bench=. -benchmem ./...

# 运行特定基准测试
go test -bench=BenchmarkCreateTask ./app/

# 保存基准测试结果
go test -bench=. -benchmem ./... > benchmark.txt

# 比较基准测试（需要 benchstat）
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

### 当前测试覆盖率

- **app/**：73.4% 覆盖率
- **parser/**：94.6% 覆盖率
- **internal/validator/**：90.2% 覆盖率
- **internal/storage/**：90.7% 覆盖率
- **总体**：73%+ 覆盖率
- **总测试数**：99+ 测试用例

### CI/CD 测试

测试自动运行于：
- 每次推送到 main/master
- 每个 pull request
- 多个 Go 版本（1.21、1.22、1.23）
- 多个平台（Linux、macOS、Windows）

## 故障排除

### API 密钥问题

如果遇到认证错误：
```bash
# 验证 API 密钥是否已设置
echo $API_KEY

# 如需要，重新导出
export API_KEY="your-key-here"
```

### 自定义 LLM 提供商

使用不同的 LLM 提供商：
```bash
# OpenAI 示例
export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
export API_KEY="your-openai-api-key"

# Anthropic Claude 示例（通过代理）
export LLM_BASE_URL="https://your-claude-proxy.com/v1/chat/completions"
```

### 构建错误

确保所有依赖已安装：
```bash
go mod download
go mod tidy
```

### 权限被拒绝

```bash
# 使其可执行
chmod +x todo

# 然后运行
./todo list
```

## 最近更新

### 版本 1.3.0（最新）

1. **国际化 (i18n) 支持**：
   - 完整支持中文和英文
   - 自动检测系统语言或使用 `TODO_LANG` 环境变量
   - 所有用户界面文本完全翻译（命令、消息、错误等）

2. **CI/CD 流水线**：
   - GitHub Actions 自动化测试
   - 多平台构建（Linux、macOS、Windows）
   - 代码检查和格式化检查
   - 覆盖率报告集成

3. **UX 改进**：
   - 彩色终端输出，带有状态指示器
   - AI 操作进度旋转器
   - 带有可操作建议的错误消息
   - Shell 补全支持（Bash、Zsh、Fish、PowerShell）

4. **性能优化**：
   - 全面的基准测试套件
   - 优化的构建标志
   - 性能基线指标

## 贡献

欢迎贡献！请随时提交问题或拉取请求。

## 许可证

MIT License

## 联系方式

[添加你的联系信息]
