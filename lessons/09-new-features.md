# go-todo 新功能详解

## 目录
1. [新功能概览](#新功能概览)
2. [国际化支持（i18n）](#国际化支持i18n)
3. [重复任务（Recurring Tasks）](#重复任务recurring-tasks)
4. [新命令介绍](#新命令介绍)
5. [安装脚本和 Makefile](#安装脚本和-makefile)
6. [实践练习](#实践练习)

---

## 新功能概览

自基础版本以来，go-todo 项目增加了许多强大的新功能：

### 主要新功能

| 功能 | 描述 | 命令 |
|------|------|------|
| 🌍 **国际化** | 支持多语言（中文、英文） | `lang` |
| 🔄 **重复任务** | 自动重复的任务（每日、每周、每月等） | 自然语言创建 |
| ⚙️ **初始化** | 快速初始化配置 | `init` |
| 📦 **压缩** | 将多个任务压缩为总结 | `compact` |
| 📋 **复制** | 复制现有任务 | `copy` |
| 🛠️ **构建工具** | Makefile 和安装脚本 | `make install` |

---

## 国际化支持（i18n）

### 什么是国际化？

国际化（Internationalization，简称 i18n）让应用支持多种语言。

### 项目中的实现

#### 1. 目录结构

```
internal/i18n/
├── i18n.go                      # i18n 核心逻辑
└── translations/
    ├── en.json                  # 英文翻译
    └── zh.json                  # 中文翻译
```

#### 2. 翻译文件格式

**translations/zh.json：**
```json
{
  "cmd.root.short": "AI 驱动的待办事项管理 CLI",
  "cmd.root.long": "Todo-Go 是一个 AI 驱动的命令行待办事项管理应用...",
  "cmd.list.short": "列出所有待办事项",
  "cmd.complete.short": "标记任务为已完成",
  "error.task_not_found": "未找到 ID 为 %d 的任务"
}
```

**translations/en.json：**
```json
{
  "cmd.root.short": "AI-powered todo management CLI",
  "cmd.root.long": "Todo-Go is an AI-powered command-line todo management application...",
  "cmd.list.short": "List all todos",
  "cmd.complete.short": "Mark task as completed",
  "error.task_not_found": "Task with ID %d not found"
}
```

#### 3. 使用方法

**在代码中使用翻译：**

```go
// internal/i18n/i18n.go
package i18n

import (
    "encoding/json"
    "fmt"
    "os"
)

var translations map[string]string
var currentLanguage string

// T 返回翻译后的文本
func T(key string, args ...interface{}) string {
    if text, ok := translations[key]; ok {
        if len(args) > 0 {
            return fmt.Sprintf(text, args...)
        }
        return text
    }
    return key  // 如果找不到翻译，返回 key
}

// SetLanguage 设置语言
func SetLanguage(lang string) error {
    currentLanguage = lang
    return loadTranslations(lang)
}
```

**在命令中使用：**

```go
// cmd/list.go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: i18n.T("cmd.list.short"),
    Long:  i18n.T("cmd.list.long"),
    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}
```

**在错误消息中使用：**

```go
// app/command.go
func GetTask(todos *[]TodoItem, id int) (*TodoItem, error) {
    for _, task := range *todos {
        if task.TaskID == id {
            return &task, nil
        }
    }
    return nil, fmt.Errorf(i18n.T("error.task_not_found", id))
}
```

#### 4. 语言设置命令

**查看当前语言：**
```bash
$ todo lang current
Current language: zh (中文)
```

**查看支持的语言：**
```bash
$ todo lang list
Available languages:
  en - English
  zh - 中文 (Chinese)
```

**设置语言：**
```bash
$ todo lang set zh
Language set to: zh (中文)

$ todo lang set en
Language set to: en (English)
```

#### 5. 配置文件中的语言设置

```bash
# 设置环境变量
export TODO_LANGUAGE=zh

# 或在配置文件中设置（如果实现了配置文件）
# ~/.todo/config.yaml
language: zh
```

---

## 重复任务（Recurring Tasks）

### 什么是重复任务？

重复任务会按照设定的规则自动重复，例如：
- **每日**：每天早上运动
- **每周**：每周一开会
- **每月**：每月 1 号交房租
- **工作日**：每个工作日写日报

### 数据结构

```go
// app/types.go
type TodoItem struct {
    TaskID     int       `json:"task_id"`
    TaskName   string    `json:"task_name"`
    // ... 其他字段 ...

    // 重复任务相关字段
    IsRecurring      bool      `json:"is_recurring"`       // 是否是重复任务
    RecurrenceRule   string    `json:"recurrence_rule"`    // 重复规则
    RecurrenceCount  int       `json:"recurrence_count"`   // 已重复次数
    MaxRecurrences   int       `json:"max_recurrences"`    // 最大重复次数（0=无限）
    ParentTaskID     int       `json:"parent_task_id"`     // 父任务ID
    NextOccurrence   time.Time `json:"next_occurrence"`    // 下次发生时间
}
```

### 重复规则格式

重复规则使用简单的字符串格式：

```
daily           # 每天
weekly          # 每周（相同星期几）
monthly         # 每月（相同日期）
yearly          # 每年（相同日期）
weekdays        # 工作日（周一到周五）
every 2 days    # 每 2 天
every 3 weeks   # 每 3 周
monday          # 每周一
tuesday         # 每周二
```

### 创建重复任务

#### 使用自然语言

```bash
# 每天早上 8 点运动
$ todo "每天早上8点运动"

# 每周一开会
$ todo "每周一上午10点团队会议"

# 每月 1 号交房租
$ todo "每月1号交房租"

# 工作日写日报
$ todo "每个工作日下午5点写日报"

# 限制次数：只重复 5 次
$ todo "未来 5 天每天复习英语"
```

#### AI 如何理解

AI 会分析你的输入，提取重复信息：

**输入：** "每周一上午10点团队会议"

**AI 响应：**
```json
{
  "intent": "create",
  "tasks": [{
    "taskName": "团队会议",
    "taskDesc": "每周一上午10点团队会议",
    "dueDate": "下周一 10:00",
    "is_recurring": true,
    "recurrence_rule": "monday",
    "urgent": "medium"
  }]
}
```

### 重复任务的工作原理

#### 1. 完成重复任务时

```go
// app/command.go
func CompleteRecurringTask(task *TodoItem, todos *[]TodoItem) error {
    // 1. 标记当前任务为完成
    task.Status = "completed"
    task.EndTime = time.Now()

    // 2. 计算下次发生时间
    nextTime := calculateNextOccurrence(task.NextOccurrence, task.RecurrenceRule)

    // 3. 创建新的任务实例
    if task.MaxRecurrences == 0 || task.RecurrenceCount < task.MaxRecurrences {
        newTask := TodoItem{
            TaskID:          generateNewID(todos),
            TaskName:        task.TaskName,
            TaskDesc:        task.TaskDesc,
            CreateTime:      time.Now(),
            DueDate:         formatTime(nextTime),
            IsRecurring:     true,
            RecurrenceRule:  task.RecurrenceRule,
            RecurrenceCount: task.RecurrenceCount + 1,
            MaxRecurrences:  task.MaxRecurrences,
            ParentTaskID:    task.ParentTaskID,
            NextOccurrence:  nextTime,
            Status:          "pending",
        }
        *todos = append(*todos, newTask)
    }

    return nil
}
```

#### 2. 计算下次发生时间

```go
func calculateNextOccurrence(current time.Time, rule string) time.Time {
    switch rule {
    case "daily":
        return current.AddDate(0, 0, 1)  // 加 1 天

    case "weekly":
        return current.AddDate(0, 0, 7)  // 加 7 天

    case "monthly":
        return current.AddDate(0, 1, 0)  // 加 1 月

    case "yearly":
        return current.AddDate(1, 0, 0)  // 加 1 年

    case "weekdays":
        next := current.AddDate(0, 0, 1)
        for next.Weekday() == time.Saturday || next.Weekday() == time.Sunday {
            next = next.AddDate(0, 0, 1)
        }
        return next

    case "monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday":
        // 找到下一个指定的星期几
        targetWeekday := parseWeekday(rule)
        next := current.AddDate(0, 0, 1)
        for next.Weekday() != targetWeekday {
            next = next.AddDate(0, 0, 1)
        }
        return next

    default:
        // 解析 "every N days/weeks/months" 格式
        return parseCustomRule(current, rule)
    }
}
```

### 查看重复任务

```bash
# 列出所有任务（包括重复任务）
$ todo list

# 查看特定任务的详情
$ todo get 1

# 输出示例：
Task ID: 1
Name: 团队会议
Description: 每周一上午10点团队会议
Status: pending
Due Date: 2025-11-11 10:00
Urgent: medium
Recurring: Yes
  Rule: monday (每周一)
  Count: 3/10 (第 3 次，共 10 次)
  Next: 2025-11-18 10:00
```

### 重复任务的限制

```bash
# 创建有限次数的重复任务
$ todo "未来 5 天每天早上 8 点跑步"

# AI 会设置：
# - is_recurring: true
# - recurrence_rule: "daily"
# - max_recurrences: 5
# - recurrence_count: 0

# 完成 5 次后，不再创建新任务
```

---

## 新命令介绍

### 1. init 命令 - 初始化配置

**用途：** 快速设置 go-todo 环境

```bash
# 初始化配置
$ todo init

# 输出：
Initializing go-todo...
✓ Created directory: ~/.todo
✓ Created config file: ~/.todo/config.yaml
✓ Created todo file: ~/.todo/todo.json
✓ Created backup file: ~/.todo/todo_back.json

Configuration:
  Language: zh
  Todo Path: ~/.todo/todo.json
  Backup Path: ~/.todo/todo_back.json

Setup complete! Try: todo "买牛奶"
```

**实现：**

```go
// cmd/init.go
var initCmd = &cobra.Command{
    Use:   "init",
    Short: i18n.T("cmd.init.short"),
    Long:  i18n.T("cmd.init.long"),
    Run: func(cmd *cobra.Command, args []string) {
        // 1. 创建目录
        todoDir := filepath.Join(os.Getenv("HOME"), ".todo")
        if err := os.MkdirAll(todoDir, 0755); err != nil {
            fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
            os.Exit(1)
        }

        // 2. 创建空的 todo 文件
        todoPath := filepath.Join(todoDir, "todo.json")
        if _, err := os.Stat(todoPath); os.IsNotExist(err) {
            os.WriteFile(todoPath, []byte("[]"), 0644)
        }

        // 3. 创建空的备份文件
        backupPath := filepath.Join(todoDir, "todo_back.json")
        if _, err := os.Stat(backupPath); os.IsNotExist(err) {
            os.WriteFile(backupPath, []byte("[]"), 0644)
        }

        // 4. 创建配置文件（可选）
        configPath := filepath.Join(todoDir, "config.yaml")
        if _, err := os.Stat(configPath); os.IsNotExist(err) {
            defaultConfig := `language: zh
log_level: info
`
            os.WriteFile(configPath, []byte(defaultConfig), 0644)
        }

        fmt.Println(i18n.T("cmd.init.success"))
    },
}
```

### 2. lang 命令 - 语言管理

**子命令：**
- `lang list` - 列出支持的语言
- `lang current` - 显示当前语言
- `lang set <lang>` - 设置语言

```bash
# 查看支持的语言
$ todo lang list
Available languages:
  en - English
  zh - 中文 (Chinese)

# 查看当前语言
$ todo lang current
Current language: zh (中文)

# 设置为英文
$ todo lang set en
Language set to: en (English)

# 设置为中文
$ todo lang set zh
Language set to: zh (中文)
```

### 3. compact 命令 - 任务压缩

**用途：** 将多个相关任务压缩成一个总结任务

```bash
# 压缩所有已完成的任务
$ todo compact

# AI 会：
# 1. 读取所有已完成的任务
# 2. 生成一个总结
# 3. 创建一个新的总结任务
# 4. 可选：删除原任务
```

**示例：**

**原任务：**
```
1. ✓ 买牛奶
2. ✓ 买面包
3. ✓ 买鸡蛋
```

**压缩后：**
```
新任务：购物清单（已完成）
描述：购买了牛奶、面包和鸡蛋
```

### 4. copy 命令 - 复制任务

**用途：** 复制现有任务，创建新任务

```bash
# 复制任务 1
$ todo copy 1

# 输出：
Task copied successfully!
New task ID: 5
```

**实现：**

```go
// cmd/copy.go
var copyCmd = &cobra.Command{
    Use:   "copy <id>",
    Short: i18n.T("cmd.copy.short"),
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        id, _ := strconv.Atoi(args[0])

        // 查找原任务
        var original *TodoItem
        for _, task := range *todos {
            if task.TaskID == id {
                original = &task
                break
            }
        }

        if original == nil {
            fmt.Fprintf(os.Stderr, i18n.T("error.task_not_found", id))
            os.Exit(1)
        }

        // 创建副本
        newTask := *original
        newTask.TaskID = generateNewID(todos)
        newTask.CreateTime = time.Now()
        newTask.Status = "pending"

        // 添加到列表
        *todos = append(*todos, newTask)
        store.Save(todos, false)

        fmt.Printf(i18n.T("cmd.copy.success", newTask.TaskID))
    },
}
```

---

## 安装脚本和 Makefile

### install.sh - 安装脚本

**用途：** 一键安装 go-todo

```bash
# 下载并安装
curl -fsSL https://raw.githubusercontent.com/SongRunqi/go-todo/main/install.sh | bash

# 或本地安装
./install.sh
```

**install.sh 做什么？**

```bash
#!/bin/bash

# 1. 检测操作系统和架构
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# 2. 下载对应的二进制文件
URL="https://github.com/SongRunqi/go-todo/releases/latest/download/todo-${OS}-${ARCH}"
curl -L $URL -o todo

# 3. 添加执行权限
chmod +x todo

# 4. 移动到系统路径
sudo mv todo /usr/local/bin/

# 5. 初始化配置
todo init

echo "Installation complete! Try: todo --help"
```

### Makefile - 构建工具

**常用命令：**

```bash
# 构建
make build

# 运行测试
make test

# 安装到系统
make install

# 清理
make clean

# 交叉编译（所有平台）
make build-all

# 查看帮助
make help
```

**Makefile 内容：**

```makefile
# 变量定义
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags="-X main.version=$(VERSION) -s -w"
BUILD_DIR := bin

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build:
	@echo "Building todo..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/todo

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# 测试覆盖率
.PHONY: coverage
coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 安装
.PHONY: install
install: build
	@echo "Installing todo..."
	@cp $(BUILD_DIR)/todo /usr/local/bin/
	@echo "Installation complete!"

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

# 交叉编译
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/todo-linux-amd64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/todo-darwin-amd64
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/todo-darwin-arm64
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/todo-windows-amd64.exe
	@echo "Build complete!"

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  test       - Run tests"
	@echo "  coverage   - Generate coverage report"
	@echo "  install    - Install to /usr/local/bin"
	@echo "  clean      - Remove build artifacts"
	@echo "  build-all  - Build for all platforms"
	@echo "  help       - Show this help message"
```

---

## 实践练习

### 练习 1：使用国际化

1. **查看当前语言**
   ```bash
   todo lang current
   ```

2. **切换到英文**
   ```bash
   todo lang set en
   todo list
   ```

3. **切换回中文**
   ```bash
   todo lang set zh
   todo list
   ```

4. **添加新的翻译**

   编辑 `internal/i18n/translations/zh.json`：
   ```json
   {
     "custom.greeting": "你好，%s！",
     "custom.farewell": "再见！"
   }
   ```

   在代码中使用：
   ```go
   fmt.Println(i18n.T("custom.greeting", "张三"))
   ```

### 练习 2：创建重复任务

1. **每天运动**
   ```bash
   todo "每天早上 8 点运动 30 分钟"
   ```

2. **每周会议**
   ```bash
   todo "每周一上午 10 点团队会议"
   ```

3. **查看重复任务详情**
   ```bash
   todo get 1
   ```

4. **完成重复任务**
   ```bash
   todo complete 1
   todo list  # 查看是否生成了新的任务
   ```

### 练习 3：使用新命令

1. **初始化新环境**
   ```bash
   # 备份当前配置
   mv ~/.todo ~/.todo.backup

   # 初始化
   todo init

   # 恢复
   rm -rf ~/.todo
   mv ~/.todo.backup ~/.todo
   ```

2. **复制任务**
   ```bash
   todo copy 1
   todo list
   ```

3. **使用 Makefile**
   ```bash
   make build
   make test
   make coverage
   ```

### 练习 4：为国际化添加新语言

假设我们要添加日语支持：

1. **创建翻译文件**

   `internal/i18n/translations/ja.json`：
   ```json
   {
     "cmd.root.short": "AIを活用したTodo管理CLI",
     "cmd.list.short": "すべてのタスクを表示",
     "cmd.complete.short": "タスクを完了としてマーク"
   }
   ```

2. **更新 i18n.go**

   ```go
   // internal/i18n/i18n.go
   var supportedLanguages = map[string]string{
       "en": "English",
       "zh": "中文",
       "ja": "日本語",  // 新增
   }
   ```

3. **更新 lang list 命令**

   ```go
   // cmd/lang.go
   case "list":
       fmt.Println("Available languages:")
       fmt.Println("  en - English")
       fmt.Println("  zh - 中文 (Chinese)")
       fmt.Println("  ja - 日本語 (Japanese)")  // 新增
   ```

4. **测试**
   ```bash
   todo lang set ja
   todo list
   ```

---

## 总结

### 新功能带来的好处

1. **国际化**
   - 支持多语言用户
   - 易于添加新语言
   - 本地化用户体验

2. **重复任务**
   - 自动化日常任务
   - 减少手动创建
   - 智能任务管理

3. **新命令**
   - `init` - 快速开始
   - `lang` - 语言管理
   - `compact` - 任务整理
   - `copy` - 快速复制

4. **开发工具**
   - Makefile - 简化构建
   - install.sh - 一键安装
   - 更好的开发体验

### 学习要点

1. **国际化实现**
   - JSON 翻译文件
   - T() 函数的使用
   - 语言切换机制

2. **重复任务设计**
   - 规则解析
   - 时间计算
   - 任务生成

3. **命令实现**
   - Cobra 子命令
   - 参数处理
   - 错误处理

4. **构建和部署**
   - Makefile 使用
   - 交叉编译
   - 安装脚本

## 下一步

现在你已经学习了 go-todo 的所有核心功能和最新特性！

**继续探索：**
1. 阅读源代码，理解实现细节
2. 尝试添加新的翻译语言
3. 实现自定义的重复规则
4. 为项目贡献代码

**推荐实践：**
- 使用 go-todo 管理你的日常任务
- 根据需求添加新功能
- 分享你的使用经验
- 为项目提交 Pull Request

祝你学习愉快！🎉
