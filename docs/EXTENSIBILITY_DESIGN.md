# Go-Todo 可扩展性设计方案

## 项目现状

当前 go-todo 是一个 **CLI 应用**，采用分层架构：
- **命令层**: Cobra CLI 框架
- **业务逻辑层**: 核心任务管理逻辑（app/）
- **数据层**: 文件存储（JSON）
- **AI 集成**: DeepSeek API

---

## 扩展方案对比

| 方案 | 优点 | 缺点 | 适用场景 | 实现难度 |
|------|------|------|----------|----------|
| **1. RESTful API** | 跨语言、标准化、易测试 | 性能略低、需要服务端 | Web/移动应用集成 | ⭐⭐⭐ |
| **2. CLI 子进程调用** | 零改动、简单快速 | 性能差、难以调试 | 脚本、工作流集成 | ⭐ |
| **3. 共享库/SDK** | 性能最优、类型安全 | 仅限 Go 应用 | Go 微服务集成 | ⭐⭐ |
| **4. gRPC 服务** | 高性能、强类型、流式 | 学习曲线、复杂度高 | 微服务架构 | ⭐⭐⭐⭐ |
| **5. WebSocket** | 实时双向、推送支持 | 连接管理复杂 | 实时应用（如桌面通知） | ⭐⭐⭐ |

---

## 方案一：RESTful API（推荐）

### 架构设计

```
┌────────────────────────────────────────────────────────────┐
│                    客户端应用                               │
│  (Web/移动/桌面应用)                                        │
└────────────┬───────────────────────────────────────────────┘
             │ HTTP/JSON
             ↓
┌────────────────────────────────────────────────────────────┐
│                  HTTP Server (Gin/Echo)                     │
├────────────────────────────────────────────────────────────┤
│  Middleware: 认证、日志、限流、CORS                         │
├────────────────────────────────────────────────────────────┤
│  Routes:                                                    │
│    GET    /api/v1/todos          - 列出任务                │
│    GET    /api/v1/todos/:id      - 获取任务                │
│    POST   /api/v1/todos          - 创建任务                │
│    PUT    /api/v1/todos/:id      - 更新任务                │
│    DELETE /api/v1/todos/:id      - 删除任务                │
│    POST   /api/v1/todos/:id/complete - 完成任务            │
│    POST   /api/v1/todos/parse    - AI 解析自然语言         │
│    GET    /api/v1/backup         - 已完成任务              │
│    POST   /api/v1/backup/:id/restore - 恢复任务            │
├────────────────────────────────────────────────────────────┤
│              Handler Layer (api/)                           │
│  - 请求验证、响应格式化                                     │
├────────────────────────────────────────────────────────────┤
│            Service Layer (app/)                             │
│  - 复用现有业务逻辑                                         │
└─────────────┬──────────────────────────────────────────────┘
              ↓
    ┌─────────────────────┐
    │   Data Store (JSON) │
    └─────────────────────┘
```

### 目录结构

```
go-todo/
├── cmd/
│   ├── cli/           # 现有 CLI（重命名）
│   │   └── main.go
│   └── server/        # 新增 HTTP 服务器
│       └── main.go
├── api/               # 新增 API 层
│   ├── handlers/      # HTTP 处理器
│   │   ├── todo.go
│   │   ├── backup.go
│   │   └── ai.go
│   ├── middleware/    # 中间件
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── cors.go
│   ├── routes/        # 路由定义
│   │   └── router.go
│   └── dto/           # 数据传输对象
│       ├── request.go
│       └── response.go
├── app/               # 现有业务逻辑（保持不变）
├── internal/          # 现有内部包
└── docs/
    └── api/           # API 文档（OpenAPI/Swagger）
        └── openapi.yaml
```

### 实现示例

#### 1. API Handler (`api/handlers/todo.go`)

```go
package handlers

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"go-todo/app"
	"go-todo/api/dto"
)

type TodoHandler struct {
	service *app.TodoService
}

func NewTodoHandler(service *app.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

// ListTodos godoc
// @Summary 列出所有任务
// @Tags todos
// @Produce json
// @Param status query string false "状态过滤: pending, completed, in_progress"
// @Param urgent query string false "优先级过滤: low, medium, high, urgent"
// @Success 200 {object} dto.TodoListResponse
// @Router /api/v1/todos [get]
func (h *TodoHandler) ListTodos(c *gin.Context) {
	status := c.Query("status")
	urgent := c.Query("urgent")

	todos, err := h.service.List(status, urgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.TodoListResponse{
		Data: todos,
		Meta: dto.Meta{
			Total: len(todos),
		},
	})
}

// GetTodo godoc
// @Summary 获取单个任务
// @Tags todos
// @Produce json
// @Param id path int true "任务ID"
// @Success 200 {object} dto.TodoResponse
// @Router /api/v1/todos/{id} [get]
func (h *TodoHandler) GetTodo(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	todo, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, dto.TodoResponse{Data: todo})
}

// CreateTodo godoc
// @Summary 创建任务
// @Tags todos
// @Accept json
// @Produce json
// @Param request body dto.CreateTodoRequest true "任务信息"
// @Success 201 {object} dto.TodoResponse
// @Router /api/v1/todos [post]
func (h *TodoHandler) CreateTodo(c *gin.Context) {
	var req dto.CreateTodoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	todo, err := h.service.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.TodoResponse{Data: todo})
}

// ParseNaturalLanguage godoc
// @Summary AI 解析自然语言创建任务
// @Tags todos
// @Accept json
// @Produce json
// @Param request body dto.ParseRequest true "自然语言输入"
// @Success 200 {object} dto.ParseResponse
// @Router /api/v1/todos/parse [post]
func (h *TodoHandler) ParseNaturalLanguage(c *gin.Context) {
	var req dto.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	result, err := h.service.ParseAndExecute(req.Input, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ParseResponse{
		Intent: result.Intent,
		Tasks:  result.Tasks,
		Message: result.Message,
	})
}
```

#### 2. DTO 定义 (`api/dto/request.go`)

```go
package dto

import "time"

type CreateTodoRequest struct {
	TaskName      string        `json:"taskName" binding:"required,max=200"`
	TaskDesc      string        `json:"taskDesc" binding:"max=5000"`
	DueDate       *time.Time    `json:"dueDate"`
	Urgent        string        `json:"urgent" binding:"omitempty,oneof=low medium high urgent"`
	EventDuration *int          `json:"eventDuration"` // 分钟数

	// 循环任务
	IsRecurring       bool     `json:"isRecurring"`
	RecurringType     string   `json:"recurringType" binding:"omitempty,oneof=daily weekly monthly yearly"`
	RecurringInterval int      `json:"recurringInterval" binding:"omitempty,min=1,max=365"`
	RecurringWeekdays []int    `json:"recurringWeekdays" binding:"omitempty,dive,min=0,max=6"`
	RecurringMaxCount int      `json:"recurringMaxCount" binding:"omitempty,min=0,max=10000"`
}

type UpdateTodoRequest struct {
	TaskName      *string    `json:"taskName" binding:"omitempty,max=200"`
	TaskDesc      *string    `json:"taskDesc" binding:"omitempty,max=5000"`
	DueDate       *time.Time `json:"dueDate"`
	Urgent        *string    `json:"urgent" binding:"omitempty,oneof=low medium high urgent"`
	Status        *string    `json:"status" binding:"omitempty,oneof=pending completed in_progress"`
	EventDuration *int       `json:"eventDuration"`
}

type ParseRequest struct {
	Input    string `json:"input" binding:"required"`
	Language string `json:"language" binding:"omitempty,oneof=en zh"`
}
```

#### 3. 服务层重构 (`app/service.go`)

```go
package app

import (
	"context"
	"time"
	"go-todo/internal/ai"
)

// TodoService 业务逻辑服务
type TodoService struct {
	store     TodoStore
	aiClient  ai.Client
	config    *Config
}

func NewTodoService(store TodoStore, aiClient ai.Client, config *Config) *TodoService {
	return &TodoService{
		store:    store,
		aiClient: aiClient,
		config:   config,
	}
}

// List 列出任务（支持过滤）
func (s *TodoService) List(status, urgent string) ([]TodoItem, error) {
	todos := s.store.Load(false)

	// 应用过滤器
	var filtered []TodoItem
	for _, todo := range todos {
		if status != "" && todo.Status != status {
			continue
		}
		if urgent != "" && todo.Urgent != urgent {
			continue
		}
		filtered = append(filtered, todo)
	}

	return filtered, nil
}

// GetByID 根据 ID 获取任务
func (s *TodoService) GetByID(id int) (*TodoItem, error) {
	todos := s.store.Load(false)
	for _, todo := range todos {
		if todo.TaskID == id {
			return &todo, nil
		}
	}
	return nil, ErrTaskNotFound
}

// Create 创建任务
func (s *TodoService) Create(req interface{}) (*TodoItem, error) {
	// 实现创建逻辑（复用现有代码）
	// ...
	return &TodoItem{}, nil
}

// ParseAndExecute AI 解析并执行
func (s *TodoService) ParseAndExecute(input, lang string) (*ParseResult, error) {
	ctx := context.Background()

	// 构建 AI 提示词
	prompt := s.buildPrompt(input, lang)

	// 调用 AI
	response, err := s.aiClient.Chat(ctx, []ai.Message{
		{Role: "system", Content: prompt},
		{Role: "user", Content: input},
	})
	if err != nil {
		return nil, err
	}

	// 解析 AI 返回的 JSON
	var result ParseResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, err
	}

	// 执行意图操作
	if err := s.executeIntent(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type ParseResult struct {
	Intent  string     `json:"intent"`
	Tasks   []TodoItem `json:"tasks"`
	Message string     `json:"message"`
}
```

#### 4. 路由配置 (`api/routes/router.go`)

```go
package routes

import (
	"github.com/gin-gonic/gin"
	"go-todo/api/handlers"
	"go-todo/api/middleware"
)

func SetupRouter(todoHandler *handlers.TodoHandler, backupHandler *handlers.BackupHandler) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 认证中间件（可选）
		// v1.Use(middleware.Auth())

		// 任务相关
		todos := v1.Group("/todos")
		{
			todos.GET("", todoHandler.ListTodos)
			todos.GET("/:id", todoHandler.GetTodo)
			todos.POST("", todoHandler.CreateTodo)
			todos.PUT("/:id", todoHandler.UpdateTodo)
			todos.DELETE("/:id", todoHandler.DeleteTodo)
			todos.POST("/:id/complete", todoHandler.CompleteTodo)
			todos.POST("/parse", todoHandler.ParseNaturalLanguage)
		}

		// 备份相关
		backup := v1.Group("/backup")
		{
			backup.GET("", backupHandler.ListBackup)
			backup.GET("/:id", backupHandler.GetBackup)
			backup.POST("/:id/restore", backupHandler.RestoreTask)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
```

#### 5. 服务器入口 (`cmd/server/main.go`)

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"go-todo/api/handlers"
	"go-todo/api/routes"
	"go-todo/app"
	"go-todo/internal/ai"
)

func main() {
	// 加载配置
	config := app.LoadConfig()

	// 初始化存储
	store := app.NewFileTodoStore(config.TodoPath, config.BackupPath)

	// 初始化 AI 客户端
	aiClient := ai.NewDeepSeekClient(
		config.LLMBaseURL,
		config.APIKey,
		config.Model,
	)

	// 初始化服务
	todoService := app.NewTodoService(store, aiClient, config)

	// 初始化处理器
	todoHandler := handlers.NewTodoHandler(todoService)
	backupHandler := handlers.NewBackupHandler(todoService)

	// 设置路由
	router := routes.SetupRouter(todoHandler, backupHandler)

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
```

### API 文档（OpenAPI 3.0）

```yaml
openapi: 3.0.0
info:
  title: Go-Todo API
  version: 1.0.0
  description: AI-powered task management API
servers:
  - url: http://localhost:8080/api/v1
    description: Development server

paths:
  /todos:
    get:
      summary: 列出所有任务
      parameters:
        - in: query
          name: status
          schema:
            type: string
            enum: [pending, completed, in_progress]
        - in: query
          name: urgent
          schema:
            type: string
            enum: [low, medium, high, urgent]
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: '#/components/schemas/TodoItem'
                  meta:
                    type: object
                    properties:
                      total:
                        type: integer

    post:
      summary: 创建任务
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTodoRequest'
      responses:
        '201':
          description: 创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TodoResponse'

  /todos/{id}:
    get:
      summary: 获取单个任务
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: integer
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TodoResponse'
        '404':
          description: 任务不存在

  /todos/parse:
    post:
      summary: AI 解析自然语言
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                input:
                  type: string
                  example: "明天下午3点开会"
                language:
                  type: string
                  enum: [en, zh]
                  default: zh
      responses:
        '200':
          description: 解析成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  intent:
                    type: string
                  tasks:
                    type: array
                    items:
                      $ref: '#/components/schemas/TodoItem'
                  message:
                    type: string

components:
  schemas:
    TodoItem:
      type: object
      properties:
        taskId:
          type: integer
        taskName:
          type: string
        taskDesc:
          type: string
        status:
          type: string
          enum: [pending, completed, in_progress]
        urgent:
          type: string
          enum: [low, medium, high, urgent]
        dueDate:
          type: string
          format: date
        createTime:
          type: string
          format: date-time
        isRecurring:
          type: boolean
```

### 客户端示例

#### JavaScript/TypeScript

```typescript
// client.ts
class TodoClient {
  private baseURL: string;

  constructor(baseURL: string = 'http://localhost:8080/api/v1') {
    this.baseURL = baseURL;
  }

  async listTodos(filters?: { status?: string; urgent?: string }) {
    const params = new URLSearchParams(filters);
    const response = await fetch(`${this.baseURL}/todos?${params}`);
    return response.json();
  }

  async createTodo(data: CreateTodoRequest) {
    const response = await fetch(`${this.baseURL}/todos`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return response.json();
  }

  async parseNaturalLanguage(input: string, language: string = 'zh') {
    const response = await fetch(`${this.baseURL}/todos/parse`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ input, language }),
    });
    return response.json();
  }
}

// 使用示例
const client = new TodoClient();

// 自然语言创建任务
const result = await client.parseNaturalLanguage('明天下午3点开会讨论项目进度');
console.log(result);
// { intent: 'create', tasks: [...], message: '已创建任务' }
```

#### Python

```python
# client.py
import requests
from typing import Optional, List, Dict

class TodoClient:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url

    def list_todos(self, status: Optional[str] = None, urgent: Optional[str] = None) -> List[Dict]:
        params = {}
        if status:
            params['status'] = status
        if urgent:
            params['urgent'] = urgent

        response = requests.get(f"{self.base_url}/todos", params=params)
        response.raise_for_status()
        return response.json()['data']

    def create_todo(self, task_data: Dict) -> Dict:
        response = requests.post(f"{self.base_url}/todos", json=task_data)
        response.raise_for_status()
        return response.json()['data']

    def parse_natural_language(self, input_text: str, language: str = "zh") -> Dict:
        response = requests.post(
            f"{self.base_url}/todos/parse",
            json={"input": input_text, "language": language}
        )
        response.raise_for_status()
        return response.json()

# 使用示例
client = TodoClient()

# 自然语言创建
result = client.parse_natural_language("每周一早上9点团队站会")
print(result)
```

### 部署方式

#### Docker 容器化

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /server .

EXPOSE 8080
CMD ["./server"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  todo-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - API_KEY=${API_KEY}
      - TODO_LANG=zh
      - PORT=8080
    volumes:
      - todo-data:/root/.todo
    restart: unless-stopped

volumes:
  todo-data:
```

### 优缺点总结

✅ **优点**：
- 跨语言支持（任何语言都可调用）
- 标准化 HTTP 协议
- 易于测试和文档化（Swagger/OpenAPI）
- 支持水平扩展
- 前后端分离

❌ **缺点**：
- 需要运行服务器进程
- HTTP 开销（相比内存调用）
- 需要处理并发和认证

---

## 方案二：CLI 子进程调用

### 实现方式

保持现有 CLI 架构，其他应用通过 `exec` 调用。

#### Go 调用示例

```go
package main

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type TodoWrapper struct {
	cliPath string
}

func NewTodoWrapper(cliPath string) *TodoWrapper {
	return &TodoWrapper{cliPath: cliPath}
}

func (w *TodoWrapper) List() ([]TodoItem, error) {
	cmd := exec.Command(w.cliPath, "list", "--json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var todos []TodoItem
	if err := json.Unmarshal(output, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func (w *TodoWrapper) Parse(input string) error {
	cmd := exec.Command(w.cliPath, input)
	return cmd.Run()
}

// 使用示例
func main() {
	wrapper := NewTodoWrapper("/usr/local/bin/go-todo")

	// 列出任务
	todos, _ := wrapper.List()

	// 自然语言创建
	wrapper.Parse("明天下午3点开会")
}
```

#### Python 调用示例

```python
import subprocess
import json

class TodoCLI:
    def __init__(self, cli_path="/usr/local/bin/go-todo"):
        self.cli_path = cli_path

    def list(self):
        result = subprocess.run(
            [self.cli_path, "list", "--json"],
            capture_output=True,
            text=True
        )
        return json.loads(result.stdout)

    def parse(self, input_text):
        subprocess.run([self.cli_path, input_text], check=True)

# 使用
cli = TodoCLI()
todos = cli.list()
cli.parse("每周一早上9点开会")
```

### 优化建议

为了更好地支持子进程调用，需要改进 CLI：

#### 1. 添加 `--json` 标志（输出标准化）

```go
// cmd/root.go
var jsonOutput bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
}

// 在各个命令中使用
func listCmd(cmd *cobra.Command, args []string) error {
	todos := // ... 获取任务

	if jsonOutput {
		data, _ := json.Marshal(todos)
		fmt.Println(string(data))
	} else {
		// 原有的格式化输出
		app.List(&todos)
	}
	return nil
}
```

#### 2. 添加 `--silent` 标志（抑制交互）

```go
var silentMode bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&silentMode, "silent", false, "Suppress interactive prompts")
}
```

#### 3. 统一错误码

```go
// 定义退出码
const (
	ExitSuccess     = 0
	ExitUsageError  = 1
	ExitDataError   = 2
	ExitNotFound    = 3
	ExitAPIError    = 4
)

// 在命令中使用
if task == nil {
	os.Exit(ExitNotFound)
}
```

### 优缺点

✅ **优点**：
- 无需修改架构
- 实现最简单
- 保持 CLI 独立性

❌ **缺点**：
- 性能差（每次调用都启动新进程）
- 难以调试
- 缺少类型安全
- 进程间通信受限

---

## 方案三：共享库/SDK

### 架构设计

将核心逻辑提取为 Go 包，供其他 Go 应用直接引用。

```
go-todo/
├── pkg/                # 公开 API（可被外部引用）
│   ├── todosdk/
│   │   ├── client.go
│   │   ├── types.go
│   │   └── options.go
├── app/                # 内部实现（保持不变）
└── cmd/                # CLI 工具（基于 SDK）
```

### 实现示例

#### 公开 SDK (`pkg/todosdk/client.go`)

```go
package todosdk

import (
	"context"
	"go-todo/app"
	"go-todo/internal/ai"
)

// Client 是 go-todo SDK 的主客户端
type Client struct {
	service *app.TodoService
	config  *Config
}

// Config SDK 配置
type Config struct {
	StoragePath string
	BackupPath  string
	APIKey      string
	Model       string
	BaseURL     string
	Language    string
}

// NewClient 创建新的客户端实例
func NewClient(config *Config) (*Client, error) {
	// 初始化存储
	store := app.NewFileTodoStore(config.StoragePath, config.BackupPath)

	// 初始化 AI 客户端
	aiClient := ai.NewDeepSeekClient(config.BaseURL, config.APIKey, config.Model)

	// 创建服务
	appConfig := &app.Config{
		TodoPath:   config.StoragePath,
		BackupPath: config.BackupPath,
		Language:   config.Language,
	}
	service := app.NewTodoService(store, aiClient, appConfig)

	return &Client{
		service: service,
		config:  config,
	}, nil
}

// List 列出任务
func (c *Client) List(ctx context.Context, opts *ListOptions) ([]TodoItem, error) {
	status := ""
	urgent := ""
	if opts != nil {
		status = opts.Status
		urgent = opts.Urgent
	}
	return c.service.List(status, urgent)
}

// Get 获取单个任务
func (c *Client) Get(ctx context.Context, id int) (*TodoItem, error) {
	return c.service.GetByID(id)
}

// Create 创建任务
func (c *Client) Create(ctx context.Context, req *CreateRequest) (*TodoItem, error) {
	return c.service.Create(req)
}

// Parse 解析自然语言并执行
func (c *Client) Parse(ctx context.Context, input string) (*ParseResult, error) {
	return c.service.ParseAndExecute(input, c.config.Language)
}

// Complete 完成任务
func (c *Client) Complete(ctx context.Context, id int) error {
	return c.service.Complete(id)
}

// Delete 删除任务
func (c *Client) Delete(ctx context.Context, id int) error {
	return c.service.Delete(id)
}

// Update 更新任务
func (c *Client) Update(ctx context.Context, id int, req *UpdateRequest) (*TodoItem, error) {
	return c.service.Update(id, req)
}
```

#### 类型定义 (`pkg/todosdk/types.go`)

```go
package todosdk

import "time"

// TodoItem 任务项
type TodoItem struct {
	TaskID            int           `json:"taskId"`
	TaskName          string        `json:"taskName"`
	TaskDesc          string        `json:"taskDesc"`
	Status            string        `json:"status"`
	Urgent            string        `json:"urgent"`
	DueDate           string        `json:"dueDate"`
	CreateTime        time.Time     `json:"createTime"`
	EndTime           time.Time     `json:"endTime"`
	IsRecurring       bool          `json:"isRecurring"`
	RecurringType     string        `json:"recurringType,omitempty"`
	RecurringInterval int           `json:"recurringInterval,omitempty"`
}

// ListOptions 列表查询选项
type ListOptions struct {
	Status string // pending, completed, in_progress
	Urgent string // low, medium, high, urgent
	Limit  int
	Offset int
}

// CreateRequest 创建任务请求
type CreateRequest struct {
	TaskName          string
	TaskDesc          string
	DueDate           *time.Time
	Urgent            string
	IsRecurring       bool
	RecurringType     string
	RecurringInterval int
	RecurringWeekdays []int
	RecurringMaxCount int
}

// UpdateRequest 更新任务请求
type UpdateRequest struct {
	TaskName *string
	TaskDesc *string
	DueDate  *time.Time
	Urgent   *string
	Status   *string
}

// ParseResult AI 解析结果
type ParseResult struct {
	Intent  string     `json:"intent"`
	Tasks   []TodoItem `json:"tasks"`
	Message string     `json:"message"`
}
```

### 使用示例

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SongRunqi/go-todo/pkg/todosdk"
)

func main() {
	// 初始化客户端
	client, err := todosdk.NewClient(&todosdk.Config{
		StoragePath: "/home/user/.todo/todo.json",
		BackupPath:  "/home/user/.todo/todo_back.json",
		APIKey:      "your-api-key",
		Model:       "deepseek-chat",
		BaseURL:     "https://api.deepseek.com",
		Language:    "zh",
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// 自然语言创建任务
	result, err := client.Parse(ctx, "明天下午3点开会讨论项目进度")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Intent: %s, Message: %s\n", result.Intent, result.Message)

	// 列出所有任务
	todos, err := client.List(ctx, &todosdk.ListOptions{
		Status: "pending",
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, todo := range todos {
		fmt.Printf("[%d] %s - %s\n", todo.TaskID, todo.TaskName, todo.Status)
	}

	// 完成任务
	if err := client.Complete(ctx, 1); err != nil {
		log.Fatal(err)
	}
}
```

### Go Module 发布

```bash
# 1. 初始化模块（如果还没有）
go mod init github.com/SongRunqi/go-todo

# 2. 创建版本标签
git tag v1.0.0
git push origin v1.0.0

# 3. 其他项目引用
# go.mod
module myapp

require github.com/SongRunqi/go-todo v1.0.0
```

### 优缺点

✅ **优点**：
- 性能最优（内存调用）
- 类型安全（编译时检查）
- IDE 自动补全
- 零序列化开销
- 易于调试

❌ **缺点**：
- 仅限 Go 应用
- 紧耦合（需要重新编译）
- 版本管理复杂

---

## 方案四：gRPC 服务

### 架构设计

使用 Protocol Buffers 定义接口，提供高性能 RPC 服务。

```
┌──────────────────────────────────────┐
│       客户端 (任意语言)               │
│  gRPC Client (Go/Python/Java/...)    │
└────────────┬─────────────────────────┘
             │ gRPC (HTTP/2 + Protobuf)
             ↓
┌──────────────────────────────────────┐
│       gRPC Server (Go)               │
├──────────────────────────────────────┤
│  TodoService Implementation          │
│  - 拦截器: 日志、认证、限流          │
└────────────┬─────────────────────────┘
             ↓
      ┌──────────────┐
      │ App Service  │
      └──────────────┘
```

### Proto 定义 (`api/proto/todo.proto`)

```protobuf
syntax = "proto3";

package todo.v1;
option go_package = "go-todo/api/proto/todov1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// TodoService 任务管理服务
service TodoService {
  // 列出任务
  rpc ListTodos(ListTodosRequest) returns (ListTodosResponse);

  // 获取单个任务
  rpc GetTodo(GetTodoRequest) returns (TodoItem);

  // 创建任务
  rpc CreateTodo(CreateTodoRequest) returns (TodoItem);

  // 更新任务
  rpc UpdateTodo(UpdateTodoRequest) returns (TodoItem);

  // 删除任务
  rpc DeleteTodo(DeleteTodoRequest) returns (google.protobuf.Empty);

  // 完成任务
  rpc CompleteTodo(CompleteTodoRequest) returns (TodoItem);

  // AI 解析自然语言（流式）
  rpc ParseNaturalLanguage(ParseRequest) returns (stream ParseResponse);

  // 订阅任务变化（双向流）
  rpc WatchTodos(stream WatchRequest) returns (stream WatchResponse);
}

message TodoItem {
  int32 task_id = 1;
  string task_name = 2;
  string task_desc = 3;
  string status = 4; // pending, completed, in_progress
  string urgent = 5; // low, medium, high, urgent
  google.protobuf.Timestamp create_time = 6;
  google.protobuf.Timestamp due_date = 7;

  // 循环任务
  bool is_recurring = 10;
  string recurring_type = 11;
  int32 recurring_interval = 12;
  repeated int32 recurring_weekdays = 13;
  int32 recurring_max_count = 14;
}

message ListTodosRequest {
  string status = 1;
  string urgent = 2;
  int32 page_size = 3;
  string page_token = 4;
}

message ListTodosResponse {
  repeated TodoItem todos = 1;
  string next_page_token = 2;
  int32 total_count = 3;
}

message GetTodoRequest {
  int32 task_id = 1;
}

message CreateTodoRequest {
  string task_name = 1;
  string task_desc = 2;
  google.protobuf.Timestamp due_date = 3;
  string urgent = 4;

  // 循环任务
  bool is_recurring = 5;
  string recurring_type = 6;
  int32 recurring_interval = 7;
  repeated int32 recurring_weekdays = 8;
  int32 recurring_max_count = 9;
}

message UpdateTodoRequest {
  int32 task_id = 1;
  optional string task_name = 2;
  optional string task_desc = 3;
  optional google.protobuf.Timestamp due_date = 4;
  optional string urgent = 5;
  optional string status = 6;
}

message DeleteTodoRequest {
  int32 task_id = 1;
}

message CompleteTodoRequest {
  int32 task_id = 1;
}

message ParseRequest {
  string input = 1;
  string language = 2; // en, zh
}

message ParseResponse {
  string intent = 1;
  repeated TodoItem tasks = 2;
  string message = 3;
  bool is_final = 4; // 流式返回时标记最后一条
}

message WatchRequest {
  enum EventType {
    SUBSCRIBE = 0;
    UNSUBSCRIBE = 1;
  }
  EventType event_type = 1;
  repeated string filters = 2; // status, urgent 过滤器
}

message WatchResponse {
  enum ChangeType {
    CREATED = 0;
    UPDATED = 1;
    DELETED = 2;
    COMPLETED = 3;
  }
  ChangeType change_type = 1;
  TodoItem todo = 2;
  google.protobuf.Timestamp timestamp = 3;
}
```

### 生成代码

```bash
# 安装工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成 Go 代码
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/todo.proto
```

### 服务实现 (`api/grpc/server.go`)

```go
package grpc

import (
	"context"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	todov1 "go-todo/api/proto/todov1"
	"go-todo/app"
)

type TodoServer struct {
	todov1.UnimplementedTodoServiceServer
	service *app.TodoService
}

func NewTodoServer(service *app.TodoService) *TodoServer {
	return &TodoServer{service: service}
}

func (s *TodoServer) ListTodos(ctx context.Context, req *todov1.ListTodosRequest) (*todov1.ListTodosResponse, error) {
	todos, err := s.service.List(req.Status, req.Urgent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list todos: %v", err)
	}

	pbTodos := make([]*todov1.TodoItem, len(todos))
	for i, todo := range todos {
		pbTodos[i] = toPbTodoItem(&todo)
	}

	return &todov1.ListTodosResponse{
		Todos:      pbTodos,
		TotalCount: int32(len(todos)),
	}, nil
}

func (s *TodoServer) GetTodo(ctx context.Context, req *todov1.GetTodoRequest) (*todov1.TodoItem, error) {
	todo, err := s.service.GetByID(int(req.TaskId))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "task not found: %v", err)
	}
	return toPbTodoItem(todo), nil
}

func (s *TodoServer) CreateTodo(ctx context.Context, req *todov1.CreateTodoRequest) (*todov1.TodoItem, error) {
	createReq := fromPbCreateRequest(req)
	todo, err := s.service.Create(createReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create todo: %v", err)
	}
	return toPbTodoItem(todo), nil
}

func (s *TodoServer) ParseNaturalLanguage(req *todov1.ParseRequest, stream todov1.TodoService_ParseNaturalLanguageServer) error {
	// 流式返回解析过程
	result, err := s.service.ParseAndExecute(req.Input, req.Language)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to parse: %v", err)
	}

	// 发送结果
	pbTasks := make([]*todov1.TodoItem, len(result.Tasks))
	for i, task := range result.Tasks {
		pbTasks[i] = toPbTodoItem(&task)
	}

	return stream.Send(&todov1.ParseResponse{
		Intent:  result.Intent,
		Tasks:   pbTasks,
		Message: result.Message,
		IsFinal: true,
	})
}

func (s *TodoServer) WatchTodos(stream todov1.TodoService_WatchTodosServer) error {
	// 实现双向流：客户端订阅任务变化
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// 处理订阅请求
		// ... 实现事件推送逻辑
	}
}

// 辅助函数：转换数据模型
func toPbTodoItem(todo *app.TodoItem) *todov1.TodoItem {
	return &todov1.TodoItem{
		TaskId:            int32(todo.TaskID),
		TaskName:          todo.TaskName,
		TaskDesc:          todo.TaskDesc,
		Status:            todo.Status,
		Urgent:            todo.Urgent,
		CreateTime:        timestamppb.New(todo.CreateTime),
		DueDate:           timestamppb.New(todo.EndTime),
		IsRecurring:       todo.IsRecurring,
		RecurringType:     todo.RecurringType,
		RecurringInterval: int32(todo.RecurringInterval),
		RecurringWeekdays: int32SliceToInt32(todo.RecurringWeekdays),
		RecurringMaxCount: int32(todo.RecurringMaxCount),
	}
}
```

### 服务器启动 (`cmd/grpc-server/main.go`)

```go
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	todov1 "go-todo/api/proto/todov1"
	grpcserver "go-todo/api/grpc"
	"go-todo/app"
	"go-todo/internal/ai"
)

func main() {
	// 初始化服务
	config := app.LoadConfig()
	store := app.NewFileTodoStore(config.TodoPath, config.BackupPath)
	aiClient := ai.NewDeepSeekClient(config.LLMBaseURL, config.APIKey, config.Model)
	service := app.NewTodoService(store, aiClient, config)

	// 创建 gRPC 服务器
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			authInterceptor,
		),
	)

	// 注册服务
	todoServer := grpcserver.NewTodoServer(service)
	todov1.RegisterTodoServiceServer(grpcServer, todoServer)

	// 启用反射（用于 grpcurl 等工具）
	reflection.Register(grpcServer)

	// 启动服务器
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
```

### 客户端示例

#### Go Client

```go
package main

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	todov1 "go-todo/api/proto/todov1"
)

func main() {
	// 连接服务器
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := todov1.NewTodoServiceClient(conn)
	ctx := context.Background()

	// 创建任务
	todo, err := client.CreateTodo(ctx, &todov1.CreateTodoRequest{
		TaskName: "完成项目报告",
		Urgent:   "high",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created: %v", todo)

	// 列出任务
	resp, err := client.ListTodos(ctx, &todov1.ListTodosRequest{})
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range resp.Todos {
		log.Printf("[%d] %s - %s", t.TaskId, t.TaskName, t.Status)
	}

	// 流式解析
	stream, err := client.ParseNaturalLanguage(ctx, &todov1.ParseRequest{
		Input:    "明天下午3点开会",
		Language: "zh",
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			break
		}
		log.Printf("Intent: %s, Message: %s", resp.Intent, resp.Message)
	}
}
```

#### Python Client

```python
import grpc
from api.proto import todo_pb2, todo_pb2_grpc

def main():
    # 连接服务器
    channel = grpc.insecure_channel('localhost:50051')
    stub = todo_pb2_grpc.TodoServiceStub(channel)

    # 创建任务
    todo = stub.CreateTodo(todo_pb2.CreateTodoRequest(
        task_name="完成项目报告",
        urgent="high"
    ))
    print(f"Created: {todo}")

    # 列出任务
    response = stub.ListTodos(todo_pb2.ListTodosRequest())
    for todo in response.todos:
        print(f"[{todo.task_id}] {todo.task_name} - {todo.status}")

    # 流式解析
    for resp in stub.ParseNaturalLanguage(todo_pb2.ParseRequest(
        input="明天下午3点开会",
        language="zh"
    )):
        print(f"Intent: {resp.intent}, Message: {resp.message}")

if __name__ == '__main__':
    main()
```

### 优缺点

✅ **优点**：
- 高性能（HTTP/2 + Protobuf）
- 强类型（编译时检查）
- 跨语言（支持 10+ 语言）
- 支持流式和双向通信
- 自动代码生成
- 易于负载均衡

❌ **缺点**：
- 学习曲线陡峭
- 复杂度高（需要管理 .proto 文件）
- 调试较困难
- 不适合浏览器直接调用（需要 gRPC-Web）

---

## 方案五：WebSocket（实时通信）

适用于需要**实时推送任务变化**的场景（如桌面通知、多端同步）。

### 实现示例

```go
// api/websocket/hub.go
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *TaskEvent
	register   chan *Client
	unregister chan *Client
}

type TaskEvent struct {
	Type string      `json:"type"` // created, updated, completed, deleted
	Task *TodoItem   `json:"task"`
	Time time.Time   `json:"timestamp"`
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			delete(h.clients, client)
			close(client.send)
		case event := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- event:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// 客户端订阅
// ws://localhost:8080/ws
```

---

## 推荐方案组合

根据不同场景选择组合：

### 场景 1：Web/移动应用集成
**推荐**：RESTful API + WebSocket
- REST API 用于 CRUD 操作
- WebSocket 用于实时通知

### 场景 2：Go 微服务集成
**推荐**：共享库 SDK
- 直接引用 `pkg/todosdk`
- 性能最优，类型安全

### 场景 3：脚本/自动化
**推荐**：CLI 子进程 + `--json` 标志
- 简单快速
- 无需额外服务

### 场景 4：高性能企业级
**推荐**：gRPC + REST API（gRPC-Gateway）
- gRPC 用于服务间调用
- REST API 用于外部集成

---

## 实施路线图

### 阶段 1：基础改进（1-2 天）
- [ ] 添加 `--json` 输出标志
- [ ] 添加 `--silent` 模式
- [ ] 统一错误码
- [ ] 提取服务层（`app/service.go`）

### 阶段 2：RESTful API（1 周）
- [ ] 创建 API 层（handlers, routes, dto）
- [ ] 实现核心端点（CRUD + Parse）
- [ ] 添加中间件（日志、CORS）
- [ ] OpenAPI 文档
- [ ] 集成测试

### 阶段 3：SDK 封装（3-5 天）
- [ ] 提取公开 API（`pkg/todosdk`）
- [ ] 编写示例和文档
- [ ] 发布 Go Module

### 阶段 4：高级特性（可选）
- [ ] WebSocket 实时推送
- [ ] gRPC 接口
- [ ] 认证和授权
- [ ] 多租户支持

---

## 总结

| 方案 | 实施优先级 | 投入产出比 | 推荐指数 |
|------|----------|-----------|---------|
| **RESTful API** | 🔥 高 | ⭐⭐⭐⭐⭐ | ✅ 强烈推荐 |
| **CLI 优化** | 🔥 高 | ⭐⭐⭐⭐ | ✅ 立即实施 |
| **共享库 SDK** | 中 | ⭐⭐⭐⭐ | ✅ 推荐 |
| **gRPC** | 低 | ⭐⭐⭐ | ⚠️ 按需实施 |
| **WebSocket** | 低 | ⭐⭐⭐ | ⚠️ 按需实施 |

**建议优先实施**：
1. **立即**：CLI 优化（`--json`, `--silent`）
2. **短期**：RESTful API（覆盖 80% 集成需求）
3. **中期**：共享库 SDK（Go 生态集成）
4. **长期**：gRPC/WebSocket（按实际需求）
