# go-todo 项目架构概览

## 目录
1. [项目简介](#项目简介)
2. [项目结构](#项目结构)
3. [核心模块](#核心模块)
4. [数据流](#数据流)
5. [设计模式](#设计模式)
6. [配置系统](#配置系统)
7. [存储系统](#存储系统)
8. [AI 集成](#ai-集成)

---

## 项目简介

go-todo 是一个 **AI 驱动的命令行待办事项管理应用**。

### 核心特性

1. **双模式操作**
   - **结构化命令**：`todo list`, `todo get 1`, `todo complete 1`
   - **自然语言**：`todo "明天买牛奶"`, `todo "周五前完成报告"`

2. **AI 智能解析**
   - 使用 DeepSeek API 理解自然语言
   - 自动提取任务名称、描述、截止日期、紧急程度

3. **完整的 CRUD 操作**
   - 创建、读取、更新、删除待办事项
   - 标记完成、恢复已完成项目

4. **备份系统**
   - 完成的任务自动归档
   - 可以查看和恢复历史任务

5. **Alfred 集成**
   - 为 Alfred workflow 优化的 JSON 输出
   - macOS 用户的最佳体验

---

## 项目结构

```
go-todo/
├── main.go                    # 程序入口
│
├── cmd/                       # Cobra 命令定义
│   ├── root.go               # 根命令，初始化逻辑
│   ├── list.go               # list 命令
│   ├── get.go                # get 命令
│   ├── complete.go           # complete 命令
│   ├── delete.go             # delete 命令
│   ├── update.go             # update 命令
│   └── back.go               # back 命令（备份管理）
│
├── app/                       # 核心业务逻辑
│   ├── types.go              # 数据结构定义
│   ├── command.go            # 核心 CRUD 操作
│   ├── commands.go           # Command 接口实现
│   ├── router.go             # 命令路由
│   ├── api.go                # AI 客户端封装
│   ├── config.go             # 配置加载
│   ├── storage.go            # 文件存储
│   ├── utils.go              # 工具函数
│   └── app_main.go           # 主应用逻辑
│
├── internal/                  # 内部包（不能被外部导入）
│   ├── ai/                   # AI 客户端
│   │   ├── client.go         # Client 接口定义
│   │   ├── deepseek.go       # DeepSeek 实现
│   │   └── mock.go           # Mock 实现（测试用）
│   │
│   ├── logger/               # 日志系统
│   │   └── logger.go         # Zerolog 封装
│   │
│   ├── validator/            # 输入验证
│   │   └── validator.go      # 验证函数
│   │
│   ├── storage/              # 内存存储（测试用）
│   │   └── memory.go
│   │
│   └── output/               # 输出格式化
│       ├── color.go          # 彩色输出
│       └── spinner.go        # 加载动画
│
├── parser/                    # 解析器
│   └── parser.go             # Markdown/JSON 解析
│
├── go.mod                     # Go 模块定义
├── go.sum                     # 依赖校验和
├── README.md                  # 项目文档
└── README_zh.md               # 中文文档
```

### 目录说明

#### `/cmd` - 命令层

- **职责**：处理命令行交互
- **不包含**：业务逻辑
- **只做**：参数解析、调用 app 层

#### `/app` - 应用层

- **职责**：核心业务逻辑
- **包含**：CRUD 操作、AI 调用、数据处理
- **可以**：被测试、被复用

#### `/internal` - 内部库

- **职责**：可复用的内部组件
- **特点**：不能被外部项目导入
- **包含**：日志、验证、AI 客户端等

#### `/parser` - 解析器

- **职责**：解析 Markdown 和 JSON 格式的任务
- **用于**：`update` 命令

---

## 核心模块

### 1. 数据模型（app/types.go）

#### TodoItem - 待办事项

```go
type TodoItem struct {
    TaskID     int       `json:"task_id"`      // 任务 ID
    CreateTime time.Time `json:"create_time"`  // 创建时间
    EndTime    time.Time `json:"end_time"`     // 完成时间
    User       string    `json:"user"`         // 用户
    TaskName   string    `json:"task_name"`    // 任务名称
    TaskDesc   string    `json:"task_desc"`    // 任务描述
    Status     string    `json:"status"`       // 状态（pending/completed）
    DueDate    string    `json:"due_date"`     // 截止日期
    Urgent     string    `json:"urgent"`       // 紧急程度
}
```

#### AlfredItem - Alfred 输出格式

```go
type AlfredItem struct {
    UID          string `json:"uid"`           // 唯一标识
    Title        string `json:"title"`         // 标题
    Subtitle     string `json:"subtitle"`      // 副标题
    Arg          string `json:"arg"`           // 参数
    Autocomplete string `json:"autocomplete"`  // 自动补全
    Icon         Icon   `json:"icon"`          // 图标
}
```

#### Context - 上下文

```go
type Context struct {
    Store       TodoStore    // 存储接口
    Todos       *[]TodoItem  // 待办事项列表
    Args        []string     // 命令参数
    CurrentTime time.Time    // 当前时间
    Config      *Config      // 配置
}
```

### 2. 命令系统（app/commands.go）

#### Command 接口

```go
type Command interface {
    Execute(ctx *Context) error
}
```

#### 命令实现

1. **ListCommand** - 列出所有任务
2. **GetCommand** - 获取单个任务
3. **CompleteCommand** - 标记完成
4. **DeleteCommand** - 删除任务
5. **UpdateCommand** - 更新任务
6. **BackCommand** - 备份管理
7. **AICommand** - 自然语言处理

### 3. 存储系统（app/storage.go）

#### TodoStore 接口

```go
type TodoStore interface {
    Load(backup bool) ([]TodoItem, error)
    Save(todos []TodoItem, backup bool) error
}
```

#### FileTodoStore 实现

```go
type FileTodoStore struct {
    Path       string  // 主文件路径（~/.todo/todo.json）
    BackupPath string  // 备份路径（~/.todo/todo_back.json）
}
```

**双文件系统：**
- `todo.json` - 活动任务
- `todo_back.json` - 已完成任务

### 4. AI 系统（internal/ai/）

#### Client 接口

```go
type Client interface {
    Chat(ctx context.Context, messages []Message) (string, error)
}
```

#### DeepSeekClient 实现

```go
type DeepSeekClient struct {
    APIKey  string
    BaseURL string
    Model   string
}
```

**支持的 API：**
- DeepSeek（默认）
- OpenAI（兼容）
- 任何 OpenAI 兼容的 API

### 5. 日志系统（internal/logger/）

```go
// 初始化日志
logger.Init("info")

// 使用日志
logger.Info().Msg("应用启动")
logger.Error().Err(err).Msg("发生错误")
logger.Debug().Str("key", "value").Msg("调试信息")
```

**日志级别：**
- `debug` - 调试信息
- `info` - 一般信息
- `warn` - 警告
- `error` - 错误

### 6. 输出系统（internal/output/）

```go
// 彩色输出
output.PrintSuccess("操作成功")
output.PrintError("操作失败")
output.PrintWarning("警告信息")
output.PrintInfo("提示信息")

// 加载动画
spinner := output.NewSpinner("正在处理...")
spinner.Start()
// ... 处理 ...
spinner.Stop()
```

---

## 数据流

### 场景 1：结构化命令

```
用户输入: todo list
    ↓
main.go
    ↓
cmd.Execute()
    ↓
cmd/root.go (PersistentPreRun)
    ├─ 初始化 logger
    ├─ 加载 config
    ├─ 初始化 store
    └─ 加载 todos
    ↓
cmd/list.go (Run)
    ↓
app.ListCommand.Execute()
    ↓
app/command.go (List 函数)
    ├─ 遍历 todos
    ├─ 格式化为 AlfredResponse
    └─ 输出 JSON
    ↓
输出到终端
```

### 场景 2：自然语言命令

```
用户输入: todo "明天买牛奶"
    ↓
main.go
    ↓
cmd.Execute()
    ↓
cmd/root.go (PersistentPreRun)
    ├─ 初始化
    └─ 加载数据
    ↓
cmd/root.go (Run)
    ├─ 检测到参数
    └─ 调用 handleNaturalLanguage()
    ↓
app.AICommand.Execute()
    ↓
app/api.go (Chat 函数)
    ├─ 构建 prompt
    ├─ 调用 AI API
    └─ 解析 JSON 响应
    ↓
app/command.go (DoI 函数)
    ├─ 识别意图（intent）
    └─ 路由到相应操作
    ↓
app/command.go (CreateTask)
    ├─ 创建 TodoItem
    ├─ 添加到 todos
    └─ 保存到文件
    ↓
输出成功信息
```

### 场景 3：完成任务

```
用户输入: todo complete 1
    ↓
cmd/complete.go
    ↓
app.CompleteCommand.Execute()
    ↓
app/command.go (DoComplete)
    ├─ 查找任务（ID = 1）
    ├─ 标记为 completed
    ├─ 设置 EndTime
    ├─ 从 todos 移除
    ├─ 保存到 todo.json
    ├─ 加载 backup todos
    ├─ 添加到 backup
    └─ 保存到 todo_back.json
    ↓
输出成功信息
```

### 场景 4：恢复任务

```
用户输入: todo back restore 1
    ↓
cmd/back.go (backRestoreCmd)
    ↓
app/command.go (RestoreTask)
    ├─ 从 backup 查找任务
    ├─ 修改状态为 pending
    ├─ 从 backup 移除
    ├─ 保存 backup
    ├─ 添加到活动 todos
    └─ 保存 todos
    ↓
输出成功信息
```

---

## 设计模式

### 1. 命令模式（Command Pattern）

```go
// 命令接口
type Command interface {
    Execute(ctx *Context) error
}

// 具体命令
type ListCommand struct{}
func (c *ListCommand) Execute(ctx *Context) error {
    // 实现...
}

// 使用
cmd := &ListCommand{}
cmd.Execute(ctx)
```

**好处：**
- 封装操作
- 易于添加新命令
- 便于测试

### 2. 策略模式（Strategy Pattern）

```go
// AI 客户端接口
type Client interface {
    Chat(ctx context.Context, messages []Message) (string, error)
}

// 不同的实现
type DeepSeekClient struct { /* ... */ }
type MockClient struct { /* ... */ }

// 可以轻松切换
var client Client
client = &DeepSeekClient{}  // 生产环境
client = &MockClient{}      // 测试环境
```

**好处：**
- 可替换实现
- 易于测试
- 支持多个 AI 提供商

### 3. 依赖注入（Dependency Injection）

```go
type Context struct {
    Store TodoStore  // 注入存储实现
    Config *Config   // 注入配置
}

func (c *ListCommand) Execute(ctx *Context) error {
    // 使用注入的依赖
    todos := ctx.Todos
    store := ctx.Store
}
```

**好处：**
- 松耦合
- 易于测试
- 灵活配置

### 4. 适配器模式（Adapter Pattern）

```go
// app/api.go - 封装 AI 客户端
func Chat(config Config, messages []Message) (string, error) {
    client := ai.NewDeepSeekClient(config.APIKey, config.BaseURL, config.Model)
    return client.Chat(context.Background(), messages)
}
```

**好处：**
- 隐藏底层实现细节
- 统一接口
- 易于替换

### 5. 仓储模式（Repository Pattern）

```go
type TodoStore interface {
    Load(backup bool) ([]TodoItem, error)
    Save(todos []TodoItem, backup bool) error
}

type FileTodoStore struct {
    Path       string
    BackupPath string
}
```

**好处：**
- 抽象数据访问
- 易于切换存储（文件→数据库）
- 便于测试（使用内存存储）

---

## 配置系统

### 配置结构（app/config.go）

```go
type Config struct {
    APIKey     string  // LLM API 密钥
    BaseURL    string  // API 基础 URL
    Model      string  // 模型名称
    TodoPath   string  // 活动任务文件路径
    BackupPath string  // 备份文件路径
}
```

### 配置来源

1. **环境变量**（优先）
2. **默认值**

```go
func LoadConfig() Config {
    return Config{
        APIKey:     getEnv("API_KEY", getEnv("DEEPSEEK_API_KEY", "")),
        BaseURL:    getEnv("LLM_BASE_URL", "https://api.deepseek.com"),
        Model:      getEnv("LLM_MODEL", "deepseek-chat"),
        TodoPath:   getEnv("TODO_PATH", defaultTodoPath()),
        BackupPath: getEnv("TODO_BACKUP_PATH", defaultBackupPath()),
    }
}
```

### 环境变量

```bash
# AI 配置
export API_KEY="your-api-key"
export LLM_BASE_URL="https://api.deepseek.com"
export LLM_MODEL="deepseek-chat"

# 存储配置
export TODO_PATH="$HOME/.todo/todo.json"
export TODO_BACKUP_PATH="$HOME/.todo/todo_back.json"

# 日志配置
export LOG_LEVEL="info"  # debug, info, warn, error

# 输出配置
export NO_COLOR=1  # 禁用彩色输出
```

---

## 存储系统

### 文件格式

#### todo.json（活动任务）

```json
[
  {
    "task_id": 1,
    "create_time": "2025-11-05T10:00:00Z",
    "end_time": "0001-01-01T00:00:00Z",
    "user": "song",
    "task_name": "买牛奶",
    "task_desc": "去超市买两盒牛奶",
    "status": "pending",
    "due_date": "2025-11-06",
    "urgent": "medium"
  }
]
```

#### todo_back.json（已完成任务）

```json
[
  {
    "task_id": 2,
    "create_time": "2025-11-04T10:00:00Z",
    "end_time": "2025-11-05T15:30:00Z",
    "user": "song",
    "task_name": "完成报告",
    "task_desc": "写月度工作报告",
    "status": "completed",
    "due_date": "2025-11-05",
    "urgent": "high"
  }
]
```

### 存储操作

```go
// 加载活动任务
todos, err := store.Load(false)

// 加载备份任务
backupTodos, err := store.Load(true)

// 保存活动任务
err := store.Save(todos, false)

// 保存备份任务
err := store.Save(backupTodos, true)
```

### 数据目录

```
~/.todo/
├── todo.json       # 活动任务
└── todo_back.json  # 已完成任务
```

---

## AI 集成

### AI 工作流

```
用户输入: "明天下午 3 点开会"
    ↓
构建 Prompt:
    系统提示词: "你是一个待办事项助手..."
    用户消息: "明天下午 3 点开会"
    ↓
调用 DeepSeek API
    ↓
AI 响应 JSON:
{
  "intent": "create",
  "name": "开会",
  "desc": "明天下午 3 点开会",
  "due": "明天",
  "urgent": "medium"
}
    ↓
解析 JSON
    ↓
创建 TodoItem
    ↓
保存到文件
```

### System Prompt

```go
const systemPrompt = `你是一个待办事项管理助手。
用户会用自然语言告诉你任务，你需要：
1. 识别意图（create/list/complete/delete/update）
2. 提取任务信息
3. 返回 JSON 格式

JSON 格式：
{
  "intent": "create",
  "name": "任务名称",
  "desc": "任务描述",
  "due": "截止日期",
  "urgent": "紧急程度(low/medium/high/urgent)"
}
`
```

### AI 客户端实现

```go
type DeepSeekClient struct {
    APIKey  string
    BaseURL string
    Model   string
}

func (c *DeepSeekClient) Chat(ctx context.Context, messages []Message) (string, error) {
    // 构建请求
    reqBody := ChatRequest{
        Model:    c.Model,
        Messages: messages,
    }

    // 调用 API
    resp, err := http.Post(c.BaseURL, body)

    // 解析响应
    var chatResp ChatResponse
    json.NewDecoder(resp.Body).Decode(&chatResp)

    return chatResp.Choices[0].Message.Content, nil
}
```

### 支持的 AI 提供商

1. **DeepSeek**（默认）
   ```bash
   export API_KEY="your-deepseek-key"
   export LLM_BASE_URL="https://api.deepseek.com/chat/completions"
   export LLM_MODEL="deepseek-chat"
   ```

2. **OpenAI**
   ```bash
   export API_KEY="your-openai-key"
   export LLM_BASE_URL="https://api.openai.com/v1/chat/completions"
   export LLM_MODEL="gpt-4"
   ```

3. **自定义**（任何 OpenAI 兼容的 API）
   ```bash
   export API_KEY="your-key"
   export LLM_BASE_URL="https://your-api.com/v1/chat/completions"
   export LLM_MODEL="your-model"
   ```

---

## 总结

### 架构特点

1. **分层清晰**
   - cmd 层：命令行交互
   - app 层：业务逻辑
   - internal 层：可复用组件

2. **模块化设计**
   - 每个模块职责单一
   - 接口驱动开发
   - 易于测试和维护

3. **可扩展性**
   - 易于添加新命令
   - 易于切换 AI 提供商
   - 易于切换存储方式

4. **用户友好**
   - 双模式操作
   - 彩色输出
   - 详细的错误信息

### 核心组件

| 组件 | 职责 | 位置 |
|------|------|------|
| **命令层** | 处理用户输入 | cmd/ |
| **应用层** | 业务逻辑 | app/ |
| **AI 客户端** | 自然语言处理 | internal/ai/ |
| **存储** | 数据持久化 | app/storage.go |
| **日志** | 日志记录 | internal/logger/ |
| **输出** | 格式化输出 | internal/output/ |
| **验证** | 输入验证 | internal/validator/ |

## 下一步

在下一课中，我们将：
- 逐行解析关键代码
- 理解每个函数的作用
- 学习 Go 编程技巧
- 了解如何修改和扩展

继续阅读 `06-code-walkthrough.md`
