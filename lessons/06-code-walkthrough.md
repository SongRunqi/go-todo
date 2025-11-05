# go-todo 代码详细解析

## 目录
1. [程序入口](#程序入口)
2. [初始化流程](#初始化流程)
3. [核心 CRUD 操作](#核心-crud-操作)
4. [AI 集成详解](#ai-集成详解)
5. [存储系统](#存储系统)
6. [命令路由](#命令路由)
7. [实用技巧](#实用技巧)

---

## 程序入口

### main.go

```go
package main

import "github.com/SongRunqi/go-todo/cmd"

func main() {
    cmd.Execute()
}
```

**解析：**
- **极简设计**：main 函数只有一行
- **职责单一**：只负责启动应用
- **所有逻辑在 cmd 包中**

**为什么这样设计？**
1. **清晰的入口**：任何人都能一眼看懂程序从哪里开始
2. **易于测试**：可以在测试中直接调用 `cmd.Execute()`
3. **符合 Go 惯例**：标准的 CLI 应用结构

---

## 初始化流程

### cmd/root.go - 根命令定义

```go
var rootCmd = &cobra.Command{
    Use:   "todo [natural language input]",
    Short: "AI-powered todo management CLI",
    Long:  `Todo-Go is an AI-powered command-line todo management application.`,

    // 在任何命令执行前运行（包括子命令）
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 1. 初始化日志
        logLevel := os.Getenv("LOG_LEVEL")
        if logLevel == "" {
            logLevel = "info"
        }
        logger.Init(logLevel)

        // 2. 加载配置
        config = app.LoadConfig()

        // 3. 初始化存储
        store = &app.FileTodoStore{
            Path:       config.TodoPath,
            BackupPath: config.BackupPath,
        }

        // 4. 加载待办事项
        var err error
        loadedTodos, err := store.Load(false)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
            os.Exit(1)
        }
        todos = &loadedTodos
        currentTime = time.Now()
    },

    // 根命令的主逻辑
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) > 0 {
            // 有参数：处理为自然语言
            handleNaturalLanguage(args)
        } else {
            // 无参数：显示帮助
            cmd.Help()
        }
    },
}
```

**PersistentPreRun 详解：**

这个函数在**所有命令**执行前运行，确保应用状态正确初始化。

```
用户执行: todo list
         ↓
    PersistentPreRun  ← 先执行这个
         ↓
    listCmd.Run       ← 再执行这个
```

**为什么使用 PersistentPreRun？**
1. **避免重复代码**：不需要在每个命令中都初始化
2. **确保一致性**：所有命令都有相同的初始状态
3. **集中管理**：初始化逻辑在一个地方

### Execute 函数

```go
func Execute() {
    // 禁用 Cobra 的默认错误处理
    rootCmd.SilenceErrors = true
    rootCmd.SilenceUsage = true

    if err := rootCmd.Execute(); err != nil {
        errStr := err.Error()

        // 特殊处理：未知命令 → 自然语言
        if len(os.Args) > 1 && strings.Contains(errStr, "unknown command") {
            // 手动初始化（因为 PersistentPreRun 没有运行）
            logLevel := os.Getenv("LOG_LEVEL")
            if logLevel == "" {
                logLevel = "info"
            }
            logger.Init(logLevel)

            config = app.LoadConfig()
            store = &app.FileTodoStore{
                Path:       config.TodoPath,
                BackupPath: config.BackupPath,
            }

            loadedTodos, loadErr := store.Load(false)
            if loadErr != nil {
                fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", loadErr)
                os.Exit(1)
            }
            todos = &loadedTodos
            currentTime = time.Now()

            // 作为自然语言处理
            handleNaturalLanguage(os.Args[1:])
            return
        }

        // 其他错误：正常打印
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**这段代码的巧妙之处：**

1. **未知命令的优雅降级**
   ```bash
   $ todo list        # 执行 list 命令
   $ todo buy milk    # buy 是未知命令，作为自然语言处理
   ```

2. **SilenceErrors 和 SilenceUsage**
   - 防止 Cobra 自动打印错误和用法
   - 我们自己控制错误处理

3. **手动初始化**
   - 因为未知命令不会触发 PersistentPreRun
   - 需要手动执行相同的初始化逻辑

---

## 核心 CRUD 操作

### app/command.go - CreateTask

```go
func CreateTask(todos *[]TodoItem, task *TodoItem) error {
    // 1. 验证输入
    if err := validator.ValidateTaskName(task.TaskName); err != nil {
        return err
    }

    // 2. 生成新 ID（找到最大 ID + 1）
    maxID := 0
    for _, t := range *todos {
        if t.TaskID > maxID {
            maxID = t.TaskID
        }
    }
    task.TaskID = maxID + 1

    // 3. 设置默认值
    if task.CreateTime.IsZero() {
        task.CreateTime = time.Now()
    }
    if task.Status == "" {
        task.Status = "pending"
    }
    if task.User == "" {
        task.User = "default"
    }

    // 4. 添加到列表
    *todos = append(*todos, *task)

    return nil
}
```

**关键点：**

1. **指针参数**：`todos *[]TodoItem`
   ```go
   // 使用指针，可以修改原始切片
   *todos = append(*todos, *task)
   ```

2. **ID 生成策略**
   - 遍历找最大 ID
   - 简单但有效（适合小规模数据）
   - 更复杂的系统可能使用 UUID

3. **零值检查**：`task.CreateTime.IsZero()`
   ```go
   // time.Time 的零值是 "0001-01-01 00:00:00"
   if task.CreateTime.IsZero() {
       // 设置默认值
   }
   ```

### app/command.go - Complete

```go
func Complete(todos *[]TodoItem, task *TodoItem, store *FileTodoStore) error {
    // 1. 查找任务
    found := false
    var targetTask *TodoItem
    for i := range *todos {
        if (*todos)[i].TaskID == task.TaskID {
            found = true
            targetTask = &(*todos)[i]
            break
        }
    }

    if !found {
        return fmt.Errorf("task with ID %d not found", task.TaskID)
    }

    // 2. 标记为完成
    targetTask.Status = "completed"
    targetTask.EndTime = time.Now()

    // 3. 从活动列表移除
    newTodos := make([]TodoItem, 0, len(*todos)-1)
    for i := range *todos {
        if (*todos)[i].TaskID != task.TaskID {
            newTodos = append(newTodos, (*todos)[i])
        }
    }
    *todos = newTodos

    // 4. 保存活动列表
    if err := store.Save(todos, false); err != nil {
        return fmt.Errorf("failed to save todos: %w", err)
    }

    // 5. 添加到备份
    backupTodos, err := store.Load(true)
    if err != nil {
        return fmt.Errorf("failed to load backup: %w", err)
    }

    backupTodos = append(backupTodos, *targetTask)

    // 6. 保存备份
    if err := store.Save(&backupTodos, true); err != nil {
        return fmt.Errorf("failed to save backup: %w", err)
    }

    output.PrintTaskCompleted(targetTask.TaskID, targetTask.TaskName)
    return nil
}
```

**重点分析：**

1. **查找任务**
   ```go
   for i := range *todos {
       if (*todos)[i].TaskID == task.TaskID {
           targetTask = &(*todos)[i]  // 获取指针
           break
       }
   }
   ```
   - 为什么用 `&(*todos)[i]`？
   - 因为 `range` 返回的是副本
   - 我们需要原始元素的指针

2. **移除元素**
   ```go
   newTodos := make([]TodoItem, 0, len(*todos)-1)
   for i := range *todos {
       if (*todos)[i].TaskID != task.TaskID {
           newTodos = append(newTodos, (*todos)[i])
       }
   }
   *todos = newTodos
   ```
   - Go 没有内置的"删除切片元素"方法
   - 需要创建新切片，复制非目标元素

3. **双文件操作**
   - 从 `todo.json` 移除
   - 添加到 `todo_back.json`
   - 确保数据不丢失

4. **错误包装**：`fmt.Errorf("failed to save: %w", err)`
   - `%w` 包装错误，保留错误链
   - 可以用 `errors.Is()` 和 `errors.As()` 检查

---

## AI 集成详解

### app/command.go - System Prompt

```go
const cmd = `
<System>
You are a todo helper agent. Your task is to analyze user input and determine their intent along with any tasks they want to create.

Key behaviors:
1. Identify the user's primary intent from the <ability> tag options
2. If the user wants to create tasks, treat ';' as a separator for multiple tasks
3. Return intent as a separate, independent attribute
4. Return tasks array only when user wants to create tasks (intent="create")

<ability>
<item>
    <name>create</name>
    <desc>user wants to create one or more tasks</desc>
</item>
<item>
    <name>delete</name>
    <desc>user wants to delete a task</desc>
</item>
<item>
    <name>list</name>
    <desc>user wants to see all the todolist</desc>
</item>
<item>
    <name>complete</name>
    <desc>user wants to complete a task</desc>
</item>
</ability>

Return format (remove markdown code fence):
{
    "intent": "create|delete|list|complete",
    "tasks": [
        {
            "taskId": -1,
            "user": "if not mentioned, You is default",
            "createTime": "use current time",
            "endTime": "place end time based on the current time",
            "taskName": "Extract a clear, concise title from the user's input. Use key words from their message without adding creative interpretations.",
            "taskDesc": "Summarize the user's input directly and factually. Use the exact words and intent from the user's message. Do not add creative interpretations or assumptions. Keep it concise (1-2 sentences) and preserve the original meaning.",
            "dueDate": "give a clear due date",
            "urgent": "low, medium, high, urgent, select one, default is medium, calculate this by time left"
        }
    ]
}

Note: Only include "tasks" array when intent is "create". For other intents, omit the tasks field or return empty array.
`
```

**Prompt 设计要点：**

1. **清晰的角色定义**
   ```
   You are a todo helper agent.
   ```

2. **结构化能力列表**
   ```xml
   <ability>
   <item><name>create</name></item>
   ...
   </ability>
   ```
   - 使用 XML 标签让 AI 更容易理解
   - 明确列出所有支持的操作

3. **明确的输出格式**
   ```json
   {
       "intent": "create|delete|list|complete",
       "tasks": [...]
   }
   ```
   - 要求返回 JSON
   - 指定了字段名和类型

4. **处理边界情况**
   - 分号分隔多个任务
   - 只在 create 时包含 tasks 数组
   - 提供默认值指导

### app/command.go - DoI 函数

```go
func DoI(todoStr string, todos *[]TodoItem, store *FileTodoStore) error {
    // 1. 解析 AI 响应
    var intentResponse IntentResponse
    removedata := removeJsonTag(todoStr)  // 移除 markdown 代码块标记
    err := json.Unmarshal([]byte(removedata), &intentResponse)
    if err != nil {
        logger.ErrorWithErr(err, "Failed to parse intent response")
        return fmt.Errorf("failed to parse intent response: %w", err)
    }

    logger.Infof("Intent: %s, Number of tasks: %d",
                 intentResponse.Intent, len(intentResponse.Tasks))

    // 2. 根据意图路由
    switch intentResponse.Intent {
    case "create":
        // 创建所有任务
        for i := range intentResponse.Tasks {
            task := &intentResponse.Tasks[i]
            if err := CreateTask(todos, task); err != nil {
                return fmt.Errorf("failed to create task: %w", err)
            }
            output.PrintTaskCreated(task.TaskID, task.TaskName)
        }
        // 批量保存
        err := store.Save(todos, false)
        if err != nil {
            return fmt.Errorf("failed to save todos batch: %w", err)
        }

    case "list":
        if err := List(todos); err != nil {
            return fmt.Errorf("failed to list todos: %w", err)
        }

    case "complete":
        if len(intentResponse.Tasks) > 0 {
            if err := Complete(todos, &intentResponse.Tasks[0], store); err != nil {
                return fmt.Errorf("failed to complete task: %w", err)
            }
        }

    case "delete":
        if len(intentResponse.Tasks) > 0 {
            if err := DeleteTask(todos, intentResponse.Tasks[0].TaskID, store); err != nil {
                return fmt.Errorf("failed to delete task: %w", err)
            }
        }

    default:
        return fmt.Errorf("unknown intent: %s", intentResponse.Intent)
    }

    return nil
}
```

**关键技术点：**

1. **removeJsonTag 函数**
   ```go
   func removeJsonTag(s string) string {
       // AI 可能返回: ```json\n{...}\n```
       // 需要移除 ``` 标记
       s = strings.TrimPrefix(s, "```json")
       s = strings.TrimPrefix(s, "```")
       s = strings.TrimSuffix(s, "```")
       return strings.TrimSpace(s)
   }
   ```

2. **意图路由**
   - 使用 switch 语句
   - 每个意图对应一个操作
   - 清晰且易于扩展

3. **批量创建优化**
   ```go
   // 先创建所有任务
   for i := range intentResponse.Tasks {
       CreateTask(todos, task)
   }
   // 最后一次性保存
   store.Save(todos, false)
   ```
   - 避免多次文件 I/O
   - 提高性能

### app/api.go - Chat 函数

```go
func Chat(config Config, input string, todoList []TodoItem) (string, error) {
    // 1. 构建上下文消息
    contextMessage := buildContextMessage(todoList)

    // 2. 构建消息列表
    messages := []ai.Message{
        {
            Role:    "system",
            Content: cmd + contextMessage,  // system prompt + 当前任务列表
        },
        {
            Role:    "user",
            Content: input,  // 用户输入
        },
    }

    // 3. 调用 AI API
    client := ai.NewDeepSeekClient(config.APIKey, config.BaseURL, config.Model)
    response, err := client.Chat(context.Background(), messages)
    if err != nil {
        return "", fmt.Errorf("AI request failed: %w", err)
    }

    return response, nil
}
```

**buildContextMessage 函数：**

```go
func buildContextMessage(todoList []TodoItem) string {
    if len(todoList) == 0 {
        return "\n\nCurrent todo list is empty."
    }

    var sb strings.Builder
    sb.WriteString("\n\nCurrent todo list:\n")

    for _, todo := range todoList {
        sb.WriteString(fmt.Sprintf(
            "- ID: %d, Name: %s, Status: %s, Due: %s, Urgent: %s\n",
            todo.TaskID,
            todo.TaskName,
            todo.Status,
            todo.DueDate,
            todo.Urgent,
        ))
    }

    return sb.String()
}
```

**为什么提供上下文？**
- AI 需要知道当前有哪些任务
- 避免创建重复任务
- 更智能的 ID 分配

---

## 存储系统

### app/storage.go - FileTodoStore

```go
type FileTodoStore struct {
    Path       string  // ~/.todo/todo.json
    BackupPath string  // ~/.todo/todo_back.json
}

// Load 加载待办事项
func (s *FileTodoStore) Load(backup bool) ([]TodoItem, error) {
    // 1. 选择文件路径
    path := s.Path
    if backup {
        path = s.BackupPath
    }

    // 2. 读取文件
    data, err := os.ReadFile(path)
    if err != nil {
        // 文件不存在 → 返回空列表
        if os.IsNotExist(err) {
            return []TodoItem{}, nil
        }
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // 3. 解析 JSON
    var todos []TodoItem
    if err := json.Unmarshal(data, &todos); err != nil {
        return nil, fmt.Errorf("failed to parse JSON: %w", err)
    }

    return todos, nil
}

// Save 保存待办事项
func (s *FileTodoStore) Save(todos *[]TodoItem, backup bool) error {
    // 1. 选择文件路径
    path := s.Path
    if backup {
        path = s.BackupPath
    }

    // 2. 确保目录存在
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }

    // 3. 序列化为 JSON（带缩进）
    data, err := json.MarshalIndent(todos, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %w", err)
    }

    // 4. 写入文件
    if err := os.WriteFile(path, data, 0644); err != nil {
        return fmt.Errorf("failed to write file: %w", err)
    }

    return nil
}
```

**技术细节：**

1. **文件不存在的处理**
   ```go
   if os.IsNotExist(err) {
       return []TodoItem{}, nil  // 返回空切片而不是 nil
   }
   ```
   - 第一次运行时文件不存在是正常的
   - 返回空切片，而不是错误

2. **目录创建**
   ```go
   os.MkdirAll(dir, 0755)
   ```
   - `MkdirAll` 类似 `mkdir -p`
   - 会创建所有必要的父目录
   - `0755` 是权限：`rwxr-xr-x`

3. **JSON 格式化**
   ```go
   json.MarshalIndent(todos, "", "  ")
   ```
   - 第二个参数：前缀（通常为空）
   - 第三个参数：缩进（2 个空格）
   - 生成人类可读的 JSON

4. **文件权限**
   ```go
   os.WriteFile(path, data, 0644)
   ```
   - `0644` = `rw-r--r--`
   - 所有者可读写，其他人只读

---

## 命令路由

### app/router.go

```go
func RouteCommand(ctx *Context) error {
    // 1. 获取命令名称
    if len(ctx.Args) < 2 {
        return fmt.Errorf("no command provided")
    }
    cmdName := ctx.Args[1]

    // 2. 路由到对应的命令
    var cmd Command
    switch cmdName {
    case "list", "ls", "l":
        cmd = &ListCommand{}
    case "get", "g":
        cmd = &GetCommand{}
    case "complete", "done", "c":
        cmd = &CompleteCommand{}
    case "delete", "del", "rm":
        cmd = &DeleteCommand{}
    case "update", "u":
        cmd = &UpdateCommand{}
    case "back", "backup", "b":
        cmd = &BackCommand{}
    default:
        // 未知命令 → AI 处理
        cmd = &AICommand{}
    }

    // 3. 执行命令
    return cmd.Execute(ctx)
}
```

**设计模式：命令模式**

```go
// 命令接口
type Command interface {
    Execute(ctx *Context) error
}

// 具体命令
type ListCommand struct{}

func (c *ListCommand) Execute(ctx *Context) error {
    return List(ctx.Todos)
}
```

**好处：**
1. **统一接口**：所有命令都实现 `Execute` 方法
2. **易于扩展**：添加新命令只需实现接口
3. **便于测试**：可以单独测试每个命令

---

## 实用技巧

### 1. 错误包装

```go
// ❌ 不好：丢失上下文
if err != nil {
    return err
}

// ✅ 好：添加上下文
if err != nil {
    return fmt.Errorf("failed to load todos: %w", err)
}
```

**使用 `%w` 的好处：**
```go
err := loadTodos()
// 错误链：failed to load todos: failed to read file: permission denied

// 可以检查原始错误
if errors.Is(err, os.ErrPermission) {
    // 处理权限错误
}
```

### 2. 零值检查

```go
// time.Time 的零值检查
if task.CreateTime.IsZero() {
    task.CreateTime = time.Now()
}

// 字符串零值检查
if task.Status == "" {
    task.Status = "pending"
}

// 切片零值检查
if len(todos) == 0 {
    return []TodoItem{}
}
```

### 3. 使用 strings.Builder

```go
// ❌ 低效：字符串拼接
var result string
for _, item := range items {
    result += item  // 每次都创建新字符串
}

// ✅ 高效：使用 Builder
var sb strings.Builder
for _, item := range items {
    sb.WriteString(item)  // 在同一个缓冲区中
}
result := sb.String()
```

### 4. 延迟关闭资源

```go
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()  // 函数结束时自动关闭

// 继续处理文件...
// 不用担心忘记关闭
```

### 5. 使用 context

```go
// 带超时的 HTTP 请求
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
resp, err := client.Do(req)
```

---

## 总结

### 代码组织

1. **分层清晰**：cmd → app → internal
2. **职责单一**：每个函数做好一件事
3. **接口驱动**：便于测试和扩展

### Go 编程技巧

1. **指针使用**：修改原始数据时使用指针
2. **错误处理**：使用 `%w` 包装错误
3. **零值检查**：善用 Go 的零值特性
4. **defer**：确保资源释放

### 设计模式

1. **命令模式**：统一的命令接口
2. **策略模式**：可替换的 AI 客户端
3. **依赖注入**：通过 Context 传递依赖

## 下一步

在下一课中，我们将学习：
- 如何编写测试
- Go 的测试工具
- 调试技巧
- 性能优化

继续阅读 `07-testing-in-go.md`
