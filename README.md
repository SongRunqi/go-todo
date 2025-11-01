# Todo-Go

A powerful AI-powered todo management CLI application with Alfred integration, built in Go.

## Features

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

## Installation

### Prerequisites

- Go 1.x or higher
- DeepSeek API key (or compatible LLM API)

### Build from Source

```bash
git clone <repository-url>
cd todo-go
go build -o todo .
```

This will create an executable named `todo` in the current directory.

## Configuration

### Environment Variables

The application uses the following environment variables:

#### Required
- `DEEPSEEK_API_KEY`: Your DeepSeek API key for LLM functionality

#### Optional
- `LLM_BASE_URL`: Custom LLM API endpoint (defaults to `https://api.deepseek.com/chat/completions`)
  - Use this to switch to other LLM providers (OpenAI, Claude, etc.)
  - Example: `export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"`

### Example Configuration

```bash
# Add to your ~/.bashrc or ~/.zshrc
export DEEPSEEK_API_KEY="your-api-key-here"
export LLM_BASE_URL="https://api.deepseek.com/chat/completions"  # Optional
```

## Usage

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

### Version 1.2.0 (Latest)

1. **Backup Management Commands**:
   - `back get <id>`: View completed task details from backup
   - `back restore <id>`: Restore completed tasks back to active list
2. **Markdown Output for Tasks**: `get` command now outputs tasks in clean Markdown format
3. **Task ID in Alfred Items**: Alfred titles now begin with `[TaskID]` for easy reference
4. **Enhanced AI Descriptions**: Improved LLM prompt to generate detailed, meaningful task descriptions with full context
5. **Configurable LLM Endpoint**: Added `LLM_BASE_URL` environment variable support for flexible API provider configuration

## Development

### Project Structure

```
todo-go/
├── main.go        # Application entry point
├── types.go       # Data structures and models
├── command.go     # Business logic and LLM integration
├── api.go         # HTTP API client for LLM
├── utils.go       # Alfred conversion utilities
├── storage.go     # File-based persistence
├── go.mod         # Go module dependencies
└── README.md      # This file
```

### Key Functions

- `TransToAlfredItem()`: Converts tasks to Alfred JSON format
- `DoI()`: Intent detection and command routing
- `Chat()`: LLM API communication
- `CreateTask()`: Task creation with ID generation
- `List()`: Sorted task display
- `Complete()`: Task completion and archival
- `RestoreTask()`: Restore completed tasks back to active list
- `GetTask()`: Retrieve and format task details in Markdown
- `UpdateTask()`: Update tasks using Markdown or JSON format

### Building the Project

```bash
# Build with custom output name
go build -o todo .

# Build for different platforms
GOOS=darwin GOARCH=amd64 go build -o todo-macos .
GOOS=linux GOARCH=amd64 go build -o todo-linux .
GOOS=windows GOARCH=amd64 go build -o todo.exe .
```

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
