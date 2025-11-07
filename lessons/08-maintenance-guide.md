# go-todo 维护和扩展指南

## 目录
1. [项目维护](#项目维护)
2. [添加新功能](#添加新功能)
3. [常见任务](#常见任务)
4. [问题排查](#问题排查)
5. [性能优化](#性能优化)
6. [部署和分发](#部署和分发)
7. [贡献指南](#贡献指南)

---

## 项目维护

### 定期维护任务

#### 1. 更新依赖

```bash
# 查看可更新的依赖
go list -u -m all

# 更新所有依赖到最新版本
go get -u ./...
go mod tidy

# 更新特定依赖
go get -u github.com/spf13/cobra
go mod tidy

# 运行测试确保更新没有问题
go test ./...
```

#### 2. 检查安全漏洞

```bash
# 安装 govulncheck
go install golang.org/x/vuln/cmd/govulncheck@latest

# 检查漏洞
govulncheck ./...
```

#### 3. 代码质量检查

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行 linter
golangci-lint run

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...
```

#### 4. 检查测试覆盖率

```bash
# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 添加新功能

### 示例 1：添加"优先级"命令

假设我们要添加一个命令来设置任务优先级：

#### 步骤 1：在 cmd 目录添加命令文件

```go
// cmd/priority.go
package cmd

import (
    "fmt"
    "os"
    "strconv"

    "github.com/spf13/cobra"
    "github.com/SongRunqi/go-todo/app"
)

var priorityCmd = &cobra.Command{
    Use:   "priority <id> <priority>",
    Short: "Set task priority",
    Long:  `Set the priority of a task. Priority: low, medium, high, urgent`,
    Args:  cobra.ExactArgs(2),
    Run: func(cmd *cobra.Command, args []string) {
        // 解析 ID
        id, err := strconv.Atoi(args[0])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Invalid ID: %v\n", err)
            os.Exit(1)
        }

        // 获取优先级
        priority := args[1]

        // 创建上下文
        ctx := &app.Context{
            Store: store,
            Todos: todos,
            Args:  []string{"todo", "priority", args[0], priority},
        }

        // 执行命令
        priorityCommand := &app.PriorityCommand{
            ID:       id,
            Priority: priority,
        }

        if err := priorityCommand.Execute(ctx); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

func init() {
    rootCmd.AddCommand(priorityCmd)
}
```

#### 步骤 2：在 app 目录添加业务逻辑

```go
// app/commands.go
type PriorityCommand struct {
    ID       int
    Priority string
}

func (c *PriorityCommand) Execute(ctx *Context) error {
    return SetPriority(ctx.Todos, c.ID, c.Priority, ctx.Store)
}
```

```go
// app/command.go
func SetPriority(todos *[]TodoItem, id int, priority string, store TodoStore) error {
    // 1. 验证优先级
    validPriorities := []string{"low", "medium", "high", "urgent"}
    valid := false
    for _, p := range validPriorities {
        if priority == p {
            valid = true
            break
        }
    }
    if !valid {
        return fmt.Errorf("invalid priority: %s (must be: low, medium, high, urgent)", priority)
    }

    // 2. 查找任务
    found := false
    for i := range *todos {
        if (*todos)[i].TaskID == id {
            (*todos)[i].Urgent = priority
            found = true
            break
        }
    }

    if !found {
        return fmt.Errorf("task with ID %d not found", id)
    }

    // 3. 保存
    if err := store.Save(todos, false); err != nil {
        return fmt.Errorf("failed to save: %w", err)
    }

    output.PrintSuccess(fmt.Sprintf("Set task %d priority to %s", id, priority))
    return nil
}
```

#### 步骤 3：添加测试

```go
// app/command_test.go
func TestSetPriority(t *testing.T) {
    tests := []struct {
        name      string
        todos     []TodoItem
        id        int
        priority  string
        wantErr   bool
        checkFunc func(t *testing.T, todos []TodoItem)
    }{
        {
            name: "valid priority",
            todos: []TodoItem{
                {TaskID: 1, TaskName: "Task 1", Urgent: "medium"},
            },
            id:       1,
            priority: "high",
            wantErr:  false,
            checkFunc: func(t *testing.T, todos []TodoItem) {
                if todos[0].Urgent != "high" {
                    t.Errorf("expected priority high, got %s", todos[0].Urgent)
                }
            },
        },
        {
            name: "invalid priority",
            todos: []TodoItem{
                {TaskID: 1, TaskName: "Task 1"},
            },
            id:       1,
            priority: "invalid",
            wantErr:  true,
        },
        {
            name:     "task not found",
            todos:    []TodoItem{},
            id:       999,
            priority: "high",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            store := &FileTodoStore{
                Path: filepath.Join(tmpDir, "todo.json"),
            }

            err := SetPriority(&tt.todos, tt.id, tt.priority, store)

            if (err != nil) != tt.wantErr {
                t.Errorf("SetPriority() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            if tt.checkFunc != nil {
                tt.checkFunc(t, tt.todos)
            }
        })
    }
}
```

#### 步骤 4：测试和文档

```bash
# 运行测试
go test ./app -run TestSetPriority

# 测试命令
go run main.go priority 1 high

# 更新 README
# 添加新命令的说明
```

### 示例 2：添加数据库支持

#### 步骤 1：定义接口（已有）

```go
// app/types.go
type TodoStore interface {
    Load(backup bool) ([]TodoItem, error)
    Save(todos *[]TodoItem, backup bool) error
}
```

#### 步骤 2：实现数据库存储

```go
// app/storage_db.go
package app

import (
    "database/sql"
    "encoding/json"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

type DBTodoStore struct {
    db *sql.DB
}

func NewDBTodoStore(dbPath string) (*DBTodoStore, error) {
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }

    // 创建表
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY,
            data TEXT NOT NULL,
            is_backup INTEGER DEFAULT 0
        )
    `)
    if err != nil {
        return nil, err
    }

    return &DBTodoStore{db: db}, nil
}

func (s *DBTodoStore) Load(backup bool) ([]TodoItem, error) {
    isBackup := 0
    if backup {
        isBackup = 1
    }

    rows, err := s.db.Query("SELECT data FROM todos WHERE is_backup = ?", isBackup)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var todos []TodoItem
    for rows.Next() {
        var data string
        if err := rows.Scan(&data); err != nil {
            return nil, err
        }

        var todo TodoItem
        if err := json.Unmarshal([]byte(data), &todo); err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }

    return todos, nil
}

func (s *DBTodoStore) Save(todos *[]TodoItem, backup bool) error {
    isBackup := 0
    if backup {
        isBackup = 1
    }

    // 开始事务
    tx, err := s.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 清空旧数据
    _, err = tx.Exec("DELETE FROM todos WHERE is_backup = ?", isBackup)
    if err != nil {
        return err
    }

    // 插入新数据
    for _, todo := range *todos {
        data, err := json.Marshal(todo)
        if err != nil {
            return err
        }

        _, err = tx.Exec(
            "INSERT INTO todos (id, data, is_backup) VALUES (?, ?, ?)",
            todo.TaskID, string(data), isBackup,
        )
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

#### 步骤 3：修改配置

```go
// app/config.go
type Config struct {
    // ... 现有字段 ...
    StorageType string  // "file" 或 "db"
    DBPath      string  // 数据库路径
}

func LoadConfig() Config {
    return Config{
        // ... 现有配置 ...
        StorageType: getEnv("STORAGE_TYPE", "file"),
        DBPath:      getEnv("DB_PATH", defaultDBPath()),
    }
}

func defaultDBPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".todo", "todo.db")
}
```

#### 步骤 4：修改初始化逻辑

```go
// cmd/root.go
PersistentPreRun: func(cmd *cobra.Command, args []string) {
    // ... 其他初始化 ...

    // 初始化存储
    if config.StorageType == "db" {
        dbStore, err := app.NewDBTodoStore(config.DBPath)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Failed to init database: %v\n", err)
            os.Exit(1)
        }
        store = dbStore
    } else {
        store = &app.FileTodoStore{
            Path:       config.TodoPath,
            BackupPath: config.BackupPath,
        }
    }

    // ... 其他初始化 ...
},
```

---

## 常见任务

### 修改 AI Prompt

```go
// app/command.go
const cmd = `
<System>
你是一个待办事项助手。

// 修改这里的内容来改变 AI 行为
`
```

**修改后测试：**
```bash
go run main.go "测试新的 prompt"
```

### 修改数据结构

如果要给 TodoItem 添加新字段：

```go
// app/types.go
type TodoItem struct {
    // ... 现有字段 ...
    Tags     []string  `json:"tags"`      // 新增：标签
    Priority int       `json:"priority"`  // 新增：数字优先级
}
```

**注意事项：**
1. 添加 JSON 标签
2. 更新相关函数
3. 更新测试
4. 考虑向后兼容性

### 切换 AI 提供商

```bash
# 使用 OpenAI
export API_KEY="sk-your-openai-key"
export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
export LLM_MODEL="gpt-4"

# 使用自定义 API
export API_KEY="your-key"
export LLM_BASE_URL="https://your-api.com/v1/chat/completions"
export LLM_MODEL="your-model"
```

### 添加新的验证规则

```go
// internal/validator/validator.go
func ValidateTag(tag string) error {
    if len(tag) > 20 {
        return errors.New("tag too long (max 20 characters)")
    }
    // 只允许字母、数字和下划线
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, tag)
    if !matched {
        return errors.New("tag contains invalid characters")
    }
    return nil
}
```

---

## 问题排查

### 常见问题

#### 1. 找不到配置文件

**症状：**
```
Error loading todos: failed to read file: open ~/.todo/todo.json: no such file or directory
```

**解决方案：**
```bash
# 创建目录
mkdir -p ~/.todo

# 创建空文件
echo "[]" > ~/.todo/todo.json
echo "[]" > ~/.todo/todo_back.json
```

#### 2. AI API 调用失败

**症状：**
```
Error: AI request failed: Post "https://api.deepseek.com": dial tcp: lookup api.deepseek.com: no such host
```

**排查步骤：**
```bash
# 1. 检查环境变量
echo $API_KEY
echo $LLM_BASE_URL

# 2. 测试网络连接
curl https://api.deepseek.com

# 3. 检查 API Key
curl -H "Authorization: Bearer $API_KEY" https://api.deepseek.com/v1/models

# 4. 启用详细日志
export LOG_LEVEL=debug
go run main.go "测试"
```

#### 3. JSON 解析失败

**症状：**
```
Error: failed to parse intent response: invalid character 'x' looking for beginning of value
```

**排查步骤：**
```bash
# 1. 启用 debug 日志查看原始响应
export LOG_LEVEL=debug

# 2. 检查 AI 响应格式
# 查看 logger 输出的原始响应

# 3. 可能的原因：
# - AI 返回了 markdown 代码块
# - AI 返回了非 JSON 内容
# - API 返回了错误信息
```

#### 4. 文件权限问题

**症状：**
```
Error: failed to write file: permission denied
```

**解决方案：**
```bash
# 检查文件权限
ls -la ~/.todo/

# 修复权限
chmod 644 ~/.todo/todo.json
chmod 755 ~/.todo/
```

### 调试技巧

#### 1. 启用详细日志

```bash
export LOG_LEVEL=debug
go run main.go list
```

#### 2. 使用 Delve 调试

```bash
# 调试程序
dlv debug

# 设置断点
(dlv) break app.CreateTask
(dlv) continue

# 查看变量
(dlv) print task
(dlv) print *todos
```

#### 3. 添加临时日志

```go
// 在关键位置添加日志
logger.Debug().Interface("task", task).Msg("Creating task")
logger.Debug().Int("todoCount", len(*todos)).Msg("Todo count")
```

#### 4. 查看生成的文件

```bash
# 查看 todo.json
cat ~/.todo/todo.json | jq .

# 监控文件变化
watch -n 1 cat ~/.todo/todo.json
```

---

## 性能优化

### 1. 减少文件 I/O

**当前实现：**
```go
// 每次操作都保存
CreateTask(todos, task)
store.Save(todos, false)
```

**优化：批量操作**
```go
// 批量创建后一次性保存
for _, task := range tasks {
    CreateTask(todos, task)
}
store.Save(todos, false)  // 只保存一次
```

### 2. 使用缓存

```go
// app/cache.go
type TodoCache struct {
    todos       []TodoItem
    lastUpdated time.Time
    ttl         time.Duration
}

func (c *TodoCache) Get() ([]TodoItem, bool) {
    if time.Since(c.lastUpdated) > c.ttl {
        return nil, false
    }
    return c.todos, true
}
```

### 3. 优化 AI 调用

```go
// 添加超时
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

client.Chat(ctx, messages)
```

### 4. 并发处理

```go
// 并发加载多个文件
var wg sync.WaitGroup
var todos, backupTodos []TodoItem
var todosErr, backupErr error

wg.Add(2)

go func() {
    defer wg.Done()
    todos, todosErr = store.Load(false)
}()

go func() {
    defer wg.Done()
    backupTodos, backupErr = store.Load(true)
}()

wg.Wait()
```

### 基准测试

```go
// app/command_bench_test.go
func BenchmarkCreateTask(b *testing.B) {
    todos := []TodoItem{}
    task := TodoItem{
        TaskName: "Benchmark task",
        TaskDesc: "Testing performance",
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        todos = []TodoItem{}  // 重置
        CreateTask(&todos, &task)
    }
}

func BenchmarkListTodos(b *testing.B) {
    // 准备 1000 个任务
    todos := make([]TodoItem, 1000)
    for i := 0; i < 1000; i++ {
        todos[i] = TodoItem{
            TaskID:   i + 1,
            TaskName: fmt.Sprintf("Task %d", i+1),
        }
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        List(&todos)
    }
}
```

运行基准测试：
```bash
go test -bench=. -benchmem ./app
```

---

## 部署和分发

### 构建可执行文件

```bash
# 构建当前平台
go build -o todo

# Linux
GOOS=linux GOARCH=amd64 go build -o todo-linux-amd64

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o todo-darwin-amd64

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o todo-darwin-arm64

# Windows
GOOS=windows GOARCH=amd64 go build -o todo-windows-amd64.exe
```

### 优化构建

```bash
# 减小文件大小
go build -ldflags="-s -w" -o todo

# 添加版本信息
VERSION=1.0.0
go build -ldflags="-X main.version=$VERSION" -o todo
```

```go
// main.go
package main

import (
    "fmt"
    "github.com/SongRunqi/go-todo/cmd"
)

var version = "dev"  // 会被构建时替换

func main() {
    if len(os.Args) > 1 && os.Args[1] == "version" {
        fmt.Printf("todo version %s\n", version)
        return
    }
    cmd.Execute()
}
```

### 使用 Makefile

```makefile
# Makefile
.PHONY: build test install clean

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags="-X main.version=$(VERSION) -s -w"

build:
	go build $(LDFLAGS) -o bin/todo

test:
	go test -v -cover ./...

install: build
	cp bin/todo /usr/local/bin/

clean:
	rm -rf bin/

# 交叉编译
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/todo-linux-amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/todo-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/todo-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/todo-windows-amd64.exe
```

使用：
```bash
make build
make test
make install
make build-all
```

### 创建发布包

```bash
# 创建 release 脚本
#!/bin/bash
# release.sh

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: ./release.sh <version>"
    exit 1
fi

# 构建
make build-all

# 创建压缩包
cd bin
tar -czf todo-linux-amd64-${VERSION}.tar.gz todo-linux-amd64
tar -czf todo-darwin-amd64-${VERSION}.tar.gz todo-darwin-amd64
tar -czf todo-darwin-arm64-${VERSION}.tar.gz todo-darwin-arm64
zip todo-windows-amd64-${VERSION}.zip todo-windows-amd64.exe

# 生成校验和
sha256sum *.tar.gz *.zip > checksums.txt

echo "Release ${VERSION} created successfully!"
```

### Docker 化

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o todo

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/todo .

ENTRYPOINT ["./todo"]
```

```bash
# 构建镜像
docker build -t todo:latest .

# 运行
docker run -it --rm \
    -e API_KEY=$API_KEY \
    -v ~/.todo:/root/.todo \
    todo:latest list
```

---

## 贡献指南

### 开发流程

#### 1. Fork 和 Clone

```bash
# Fork 仓库（在 GitHub 上点击 Fork）

# Clone 你的 fork
git clone https://github.com/YOUR_USERNAME/go-todo.git
cd go-todo

# 添加上游仓库
git remote add upstream https://github.com/SongRunqi/go-todo.git
```

#### 2. 创建功能分支

```bash
# 更新 main 分支
git checkout main
git pull upstream main

# 创建新分支
git checkout -b feature/my-new-feature
```

#### 3. 开发和测试

```bash
# 开发你的功能

# 运行测试
go test ./...

# 运行 linter
golangci-lint run

# 格式化代码
go fmt ./...
```

#### 4. 提交更改

```bash
# 添加更改
git add .

# 提交（使用清晰的提交信息）
git commit -m "feat: add priority command"

# 推送到你的 fork
git push origin feature/my-new-feature
```

#### 5. 创建 Pull Request

1. 在 GitHub 上打开你的 fork
2. 点击 "New Pull Request"
3. 填写 PR 描述
4. 等待审查

### 提交信息规范

使用约定式提交（Conventional Commits）：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型：**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档变更
- `style`: 代码格式（不影响代码运行）
- `refactor`: 重构
- `test`: 添加或修改测试
- `chore`: 构建过程或辅助工具的变动

**示例：**
```
feat(cmd): add priority command

Add a new command to set task priority.

Closes #123
```

### 代码审查清单

提交 PR 前检查：

- [ ] 代码遵循项目风格
- [ ] 添加了必要的测试
- [ ] 所有测试通过
- [ ] 文档已更新
- [ ] 没有引入新的警告
- [ ] 提交信息清晰

---

## 总结

### 维护要点

1. **定期更新依赖**：保持安全和性能
2. **保持测试覆盖率**：至少 70%
3. **代码质量检查**：使用 linter 和 vet
4. **安全检查**：使用 govulncheck

### 扩展指南

1. **添加新命令**：cmd + app + 测试
2. **修改数据结构**：考虑兼容性
3. **切换存储**：实现 TodoStore 接口
4. **优化性能**：使用基准测试

### 部署建议

1. **构建优化**：使用 ldflags 减小大小
2. **交叉编译**：支持多平台
3. **容器化**：使用 Docker
4. **自动化**：使用 Makefile 和 CI/CD

## 下一步

恭喜你完成了所有课程！现在你应该：

1. **理解 Go 语言基础**
2. **掌握项目结构**
3. **了解依赖管理**
4. **熟悉 Cobra 框架**
5. **理解项目架构**
6. **能够阅读和修改代码**
7. **会编写测试**
8. **能够维护和扩展项目**

### 实践建议

1. **尝试添加新功能**
2. **修改 AI Prompt**
3. **切换 AI 提供商**
4. **优化性能**
5. **贡献代码**

### 学习资源

- [Go 官方文档](https://golang.org/doc/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Cobra 文档](https://cobra.dev/)
- [Go 语言圣经（中文版）](https://gopl-zh.github.io/)

祝你在 Go 语言的学习之旅中取得成功！
