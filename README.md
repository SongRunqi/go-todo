# Todo-Go

[![CI](https://github.com/SongRunqi/go-todo/actions/workflows/ci.yml/badge.svg)](https://github.com/SongRunqi/go-todo/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SongRunqi/go-todo)](https://goreportcard.com/report/github.com/SongRunqi/go-todo)
[![codecov](https://codecov.io/gh/SongRunqi/go-todo/branch/main/graph/badge.svg)](https://codecov.io/gh/SongRunqi/go-todo)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/SongRunqi/go-todo)](go.mod)

[English](README.md) | [中文](README_zh.md)

A powerful AI-powered todo management CLI application with Alfred integration, built in Go.

## Features

### Core Features
- **AI-Powered Task Management**: Uses LLM (DeepSeek by default) to intelligently parse natural language input
- **Alfred Integration**: Seamless integration with Alfred workflow for macOS users
- **Smart Task Parsing**: Automatically extracts task details, deadlines, and urgency from natural language
- **Comprehensive Task Operations**:
  - Create, list, get, update, complete, delete tasks
  - View and manage completed tasks (backup)
  - Restore completed tasks back to active list
- **Detailed Descriptions**: AI generates comprehensive task descriptions with context and expected outcomes
- **Priority Management**: Automatic urgency calculation based on deadlines
- **Time-Based Sorting**: Tasks sorted by due date with countdown timers
- **Multiple Output Formats**: JSON (Alfred-compatible) and Markdown formats

### Developer Features
- **🌍 Internationalization (i18n)**: Full support for English and Chinese
  - Auto-detects system language from environment
  - Switch language via `TODO_LANG` environment variable
  - All user-facing text fully translated
- **🎨 Colored Output**: Beautiful terminal output with color-coded messages
  - ✓ Green for success messages
  - ✗ Red for errors with actionable suggestions
  - ⚠ Yellow for warnings
  - ℹ Cyan for informational messages
- **⚡ Performance Optimized**: Comprehensive benchmarks and optimizations
- **🔍 Input Validation**: Robust validation layer with clear error messages
- **🧪 Well Tested**: 73%+ test coverage with unit and integration tests
- **📊 Structured Logging**: Zerolog-based logging with configurable levels
- **🔌 Pluggable AI Client**: Abstract AI interface supporting multiple LLM providers
- **💾 Memory Storage**: In-memory storage option for testing
- **🚀 CI/CD Pipeline**: Automated testing, linting, and multi-platform builds
- **🛠 Shell Completion**: Auto-completion support for Bash, Zsh, Fish, and PowerShell

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage](#usage)
- [Building for Different Platforms](#building-for-different-platforms)
- [Development](#development)
- [Testing](#testing)

## Installation

### Prerequisites

- **Go 1.21 or higher** - [Download Go](https://golang.org/dl/)
- **DeepSeek API key** (or compatible LLM API) - [Get API Key](https://platform.deepseek.com/)

Check your Go version:
```bash
go version  # Should be 1.21 or higher
```

### Quick Install

```bash
# Clone the repository
git clone https://github.com/SongRunqi/go-todo.git
cd go-todo

# Download dependencies
go mod download

# Build the application
go build -o todo main.go

# Verify the build
./todo --help
```

### Install Globally

To use `todo` from anywhere:

```bash
# Linux/macOS - copy to /usr/local/bin
sudo cp todo /usr/local/bin/todo

# Or copy to ~/bin (add ~/bin to PATH if needed)
mkdir -p ~/bin
cp todo ~/bin/todo
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Now use 'todo' from anywhere
todo list
```

## Quick Start

```bash
# 1. Set your API key
export API_KEY="your-deepseek-api-key-here"

# 2. Create a task using natural language
./todo "Buy groceries tomorrow evening"

# 3. List all tasks
./todo list

# 4. Complete a task
./todo complete 1

# 5. View completed tasks
./todo back
```

## Configuration

### Environment Variables

The application uses the following environment variables:

#### Required
- `API_KEY`: Your DeepSeek API key for LLM functionality (or `DEEPSEEK_API_KEY`)

#### Optional
- `TODO_LANG`: Set language for interface (defaults to auto-detect from system)
  - Supported values: `en` (English), `zh` (Chinese)
  - Auto-detects from `LANGUAGE`, `LC_ALL`, `LC_MESSAGES`, or `LANG` if not set
- `LLM_BASE_URL`: Custom LLM API endpoint (defaults to `https://api.deepseek.com/chat/completions`)
  - Use this to switch to other LLM providers (OpenAI, Claude, etc.)
- `LLM_MODEL`: Model to use (defaults to the provider's default)
- `LOG_LEVEL`: Logging level - `debug`, `info`, `warn`, `error` (default: `info`)
- `NO_COLOR`: Set to any value to disable colored output

### Configuration Examples

```bash
# Basic configuration (add to ~/.bashrc or ~/.zshrc)
export API_KEY="your-api-key-here"

# Full configuration
export API_KEY="your-api-key-here"
export LLM_BASE_URL="https://api.deepseek.com/chat/completions"
export LLM_MODEL="deepseek-chat"
export LOG_LEVEL="info"
export TODO_LANG="en"  # or "zh" for Chinese

# Use OpenAI instead
export API_KEY="your-openai-api-key"
export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
export LLM_MODEL="gpt-4"
```

## Internationalization (i18n)

Todo-Go supports multiple languages for all user-facing text.

### Supported Languages

- **English (en)**: Default language
- **中文 (zh)**: Simplified Chinese

### Setting Language

You can set the language using the `lang` command, which persists your preference, or use environment variables for temporary changes.

#### Using the lang command (Recommended)

```bash
# List available languages (Alfred-compatible JSON format)
./todo lang list

# Set language to Chinese (persists to config file)
./todo lang set zh

# Set language to English
./todo lang set en

# Check current language
./todo lang current
```

The language preference is saved to `~/.todo/config.json` and will persist across all future commands.

#### Using environment variables (Temporary)

```bash
# Use English (default)
./todo list

# Use Chinese
TODO_LANG=zh ./todo list

# Or set permanently in your shell configuration
export TODO_LANG=zh  # Add to ~/.bashrc or ~/.zshrc
./todo list
```

**Note**: Environment variable `TODO_LANG` takes priority over the config file setting.

### Auto-Detection

If `TODO_LANG` is not set, the application will auto-detect your system language from the following environment variables (in order):
1. `LANGUAGE`
2. `LC_ALL`
3. `LC_MESSAGES`
4. `LANG`

### Examples

**English:**
```bash
$ ./todo --help
A simple command-line TODO application that supports natural language input and AI-powered task management.
```

**Chinese:**
```bash
$ TODO_LANG=zh ./todo --help
一个简单的命令行待办事项应用，支持自然语言输入和 AI 驱动的任务管理。
```

All command help, error messages, validation messages, and output text will be displayed in your selected language.

## Usage

### Command Structure

Todo-Go now uses a modern CLI framework (Cobra) with clear command structure:

```bash
todo [command] [arguments] [flags]
```

**Available Commands:**
- `list` / `ls` - List all active todos
- `get <id>` - Get detailed information about a task
- `complete <id>` - Mark a task as completed
- `delete <id>` - Delete a task permanently
- `update <content>` - Update a task with Markdown or JSON
- `back` - List completed tasks
- `back get <id>` - View a completed task
- `back restore <id>` - Restore a completed task
- `help` - Get help about any command
- `completion` - Generate shell completion scripts

**Global Flags:**
- `--config` - Specify config file location
- `--verbose` / `-v` - Enable verbose output
- `--help` / `-h` - Show help for any command

**Environment Variables:**
- `LOG_LEVEL` - Set logging level (debug, info, warn, error) - defaults to "info"
- `NO_COLOR` - Disable colored output when set

**Natural Language (AI):** If you don't use a specific command, your input is treated as natural language and processed by AI.

### Shell Completion

Generate shell completion scripts for faster command entry:

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

All commands use the `./todo` executable (or `todo` if installed globally).

### Create Tasks

Create single or multiple tasks using natural language:

```bash
# Single task
./todo "Buy groceries by tomorrow evening"

# Multiple tasks (separated by semicolons)
./todo "Write report by Friday; Call client tomorrow; Review code by end of week"
```

The AI will automatically:
- Extract task name
- Generate detailed description with context
- Set due date and urgency level
- Calculate time remaining

### List Tasks

Display all active tasks:

```bash
./todo list
# or shorthand
./todo ls
```

Output format is Alfred-compatible JSON including:
- **Task ID**: `[1]` prefix for easy reference
- **Task Name**: With emoji indicator 🎯
- **Urgency Status**: Time remaining or overdue indicator
- **Description**: Detailed task context with status emoji (⌛️ pending, ✅ completed)

### View Completed Tasks (Backup)

List all completed/archived tasks:

```bash
./todo back
```

### Get Task Details

Retrieve detailed information about a specific task:

```bash
# Get active task
./todo get <task-id>

# Get completed task from backup
./todo "back get <task-id>"
```

Example output (Markdown format):
```markdown
# Task Name

- **Task ID:** 1
- **Status:** pending
- **User:** username
- **Due Date:** 2025-11-05
- **Urgency:** high
- **Created:** 2025-11-02 10:30:00
- **End Time:** 2025-11-05 18:00:00

## Description

Task description here...
```

### Complete Tasks

Mark a task as completed (moves to backup):

```bash
./todo "complete 1"
```

Completed tasks are archived in the backup file and removed from the active list.

### Restore Completed Tasks

Restore a completed task from backup back to active list:

```bash
./todo "back restore <task-id>"
```

The task status will change from "completed" to "pending".

### Update Tasks

Update an existing task using Markdown or JSON format:

```bash
# Using Markdown (recommended)
./todo update "# Updated Task Name

- **Task ID:** 1
- **Status:** pending
- **User:** username
- **Due Date:** 2025-11-10
- **Urgency:** high

## Description

Updated task description..."

# Using JSON
./todo update '{"taskId":1,"taskName":"Updated Task","taskDesc":"New description",...}'
```

### Delete Tasks

Remove a task permanently:

```bash
./todo "delete 1"
```

### Language Management

Manage language settings for the application:

```bash
# List available languages (Alfred-compatible JSON)
./todo lang list

# Set preferred language
./todo lang set en   # English
./todo lang set zh   # Chinese

# Show current language
./todo lang current
```

The language preference is saved to `~/.todo/config.json` and persists across all commands. See the [Internationalization](#internationalization-i18n) section for more details.

## Building for Different Platforms

### Build for Current Platform

```bash
# Standard build
go build -o todo main.go

# Optimized build (smaller binary)
go build -ldflags="-s -w" -o todo main.go
```

### Cross-Platform Builds

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o todo-linux-amd64 main.go

# Linux (arm64) - for Raspberry Pi, ARM servers
GOOS=linux GOARCH=arm64 go build -o todo-linux-arm64 main.go

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o todo-darwin-amd64 main.go

# macOS (Apple Silicon - M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o todo-darwin-arm64 main.go

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o todo-windows-amd64.exe main.go
```

### Build Script for All Platforms

Create a `build-all.sh` script:

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

    echo "Building $output_name..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o $output_name main.go
done

echo "All builds completed!"
```

Run it:
```bash
chmod +x build-all.sh
./build-all.sh
```

## AI-Powered Features

### Intelligent Task Description Generation

The LLM generates comprehensive descriptions that include:
1. **What needs to be done**: Specific action items
2. **Why it matters**: Purpose and expected outcomes
3. **Relevant details**: Dependencies, constraints, or context from your input

### Smart Intent Recognition

The AI automatically detects your intent:
- `create`: Adding new tasks
- `list`: Viewing all tasks
- `complete`: Marking tasks as done
- `delete`: Removing tasks

### Urgency Calculation

Tasks are automatically assigned urgency levels based on deadline:
- `urgent`: Very short timeframe
- `high`: Approaching deadline
- `medium`: Normal timeframe (default)
- `low`: Distant deadline

## Alfred Integration

### Alfred Item Format

Each task appears in Alfred with:
- **Title**: `[TaskID] 🎯 Task Name [Urgency Status]`
- **Subtitle**: `[Status] Task Description`
- **Arg**: Task ID for downstream actions

Example:
```
[1] 🎯 Buy groceries 还有2h 30m 截止
⌛️ Purchase fresh vegetables and fruits for the week...
```

### Status Indicators

- ⌛️ Pending task
- ✅ Completed task
- 还有 X时 X分 截止: Time remaining
- 已截止: Overdue

## File Storage

Tasks are stored in JSON files:
- **Active tasks**: `todos.json` (or custom location)
- **Completed tasks**: Backup file for archival

## Recent Updates

### Version 1.3.0 (Latest)

1. **Internationalization (i18n) Support**:
   - Full support for English and Chinese languages
   - Auto-detects system language or use `TODO_LANG` environment variable
   - All user-facing text fully translated (commands, messages, errors, etc.)

2. **CI/CD Pipeline**:
   - GitHub Actions automated testing
   - Multi-platform builds (Linux, macOS, Windows)
   - Linting and formatting checks
   - Coverage reporting integration

3. **UX Improvements**:
   - Colored terminal output with status indicators
   - Progress spinners for AI operations
   - Error messages with actionable suggestions
   - Shell completion support (Bash, Zsh, Fish, PowerShell)

4. **Performance Optimizations**:
   - Comprehensive benchmark suite
   - Optimized build flags
   - Performance baseline metrics

## Development

### Project Structure

```
go-todo/
├── main.go                      # Application entry point
├── cmd/                         # Command-line interface (Cobra)
│   ├── root.go                 # Root command and completion
│   ├── list.go                 # List command
│   ├── get.go                  # Get command
│   ├── complete.go             # Complete command
│   ├── delete.go               # Delete command
│   ├── update.go               # Update command
│   └── back.go                 # Backup commands
├── app/                         # Business logic
│   ├── command.go              # Core task operations
│   ├── commands.go             # Command implementations
│   ├── api.go                  # AI client wrapper
│   ├── storage.go              # File storage
│   ├── utils.go                # Utility functions
│   ├── types.go                # Data models
│   └── router.go               # Command router
├── parser/                      # Task parsing
│   ├── parser.go               # Markdown/JSON parser
│   └── parser_test.go          # Parser tests (94.6% coverage)
├── internal/                    # Internal packages
│   ├── i18n/                   # Internationalization
│   │   ├── i18n.go            # i18n package with embedded translations
│   │   └── translations/      # Translation files
│   │       ├── en.json        # English translations
│   │       └── zh.json        # Chinese translations
│   ├── logger/                 # Structured logging (zerolog)
│   ├── validator/              # Input validation
│   ├── ai/                     # AI client abstraction
│   │   ├── client.go          # Interface definition
│   │   ├── deepseek.go        # DeepSeek implementation
│   │   └── mock.go            # Mock for testing
│   ├── storage/                # Storage implementations
│   │   └── memory.go          # In-memory storage
│   └── output/                 # Terminal output
│       ├── color.go           # Colored output
│       └── spinner.go         # Progress indicators
├── .github/workflows/           # CI/CD pipeline
│   └── ci.yml                  # GitHub Actions
├── .golangci.yml               # Linter configuration
├── go.mod                      # Go module dependencies
├── go.sum                      # Dependency checksums
├── ROADMAP.md                  # Development roadmap
└── README.md                   # This file
```

### Technology Stack

- **CLI Framework**: [Cobra](https://github.com/spf13/cobra) - Modern command-line interface
- **Logging**: [Zerolog](https://github.com/rs/zerolog) - High-performance structured logging
- **Colors**: [Fatih Color](https://github.com/fatih/color) - Terminal color output
- **Spinner**: [Briandowns Spinner](https://github.com/briandowns/spinner) - Progress indicators
- **AI Provider**: DeepSeek API (configurable for other LLM providers)
- **Testing**: Go standard testing library with table-driven tests
- **CI/CD**: GitHub Actions with multi-platform builds

### Key Features by Package

**cmd/** - Command-line interface
- Cobra-based command structure
- Shell completion generation
- Natural language fallback

**app/** - Core business logic
- Task CRUD operations
- AI intent detection
- File-based persistence

**parser/** - Task parsing
- Markdown parser for task updates
- JSON parser for structured input
- Auto-format detection

**internal/logger** - Structured logging
- Configurable log levels
- Colored console output
- Error tracking

**internal/validator** - Input validation
- Task ID validation
- Field length checks
- Status and urgency validation

**internal/ai** - AI client abstraction
- Interface-based design
- DeepSeek implementation
- Mock client for testing

**internal/output** - Terminal output
- Colored success/error messages
- Progress spinners
- Actionable error suggestions

## Testing

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...
```

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Run Specific Tests

```bash
# Test specific package
go test ./app/
go test ./parser/

# Run specific test function
go test -run TestCreateTask ./app/

# Run tests matching pattern
go test -run "TestCreate.*" ./app/
```

### Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run benchmarks with memory stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkCreateTask ./app/

# Save benchmark results
go test -bench=. -benchmem ./... > benchmark.txt

# Compare benchmarks (requires benchstat)
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

### Current Test Coverage

- **app/**: 73.4% coverage
- **parser/**: 94.6% coverage
- **internal/validator/**: 90.2% coverage
- **internal/storage/**: 90.7% coverage
- **Overall**: 73%+ coverage
- **Total Tests**: 99+ test cases

### CI/CD Testing

Tests run automatically on:
- Every push to main/master
- Every pull request
- Multiple Go versions (1.21, 1.22, 1.23)
- Multiple platforms (Linux, macOS, Windows)

## Troubleshooting

### API Key Issues

If you get authentication errors:
```bash
# Verify your API key is set
echo $DEEPSEEK_API_KEY

# Re-export if needed
export DEEPSEEK_API_KEY="your-key-here"
```

### Custom LLM Provider

To use a different LLM provider:
```bash
# OpenAI example
export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
export DEEPSEEK_API_KEY="your-openai-api-key"

# Anthropic Claude example (via proxy)
export LLM_BASE_URL="https://your-claude-proxy.com/v1/chat/completions"
```

### Build Errors

Ensure all dependencies are installed:
```bash
go mod download
go mod tidy
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

[Add your license information here]

## Contact

[Add your contact information here]
