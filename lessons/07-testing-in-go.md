# Go 测试和调试

## 目录
1. [Go 测试基础](#go-测试基础)
2. [编写单元测试](#编写单元测试)
3. [表驱动测试](#表驱动测试)
4. [Mock 和桩](#mock-和桩)
5. [测试覆盖率](#测试覆盖率)
6. [基准测试](#基准测试)
7. [调试技巧](#调试技巧)
8. [go-todo 中的测试](#go-todo-中的测试)

---

## Go 测试基础

### 测试文件命名

测试文件必须以 `_test.go` 结尾：

```
mypackage/
├── user.go          # 源代码
└── user_test.go     # 测试代码
```

### 测试函数命名

测试函数必须：
1. 以 `Test` 开头
2. 接收 `*testing.T` 参数

```go
func TestAdd(t *testing.T) {
    // 测试代码
}
```

### 运行测试

```bash
# 运行当前包的所有测试
go test

# 运行所有包的测试
go test ./...

# 显示详细输出
go test -v

# 运行特定测试
go test -run TestAdd

# 运行匹配模式的测试
go test -run TestAdd.*
```

---

## 编写单元测试

### 简单测试

```go
// math.go
package math

func Add(a, b int) int {
    return a + b
}
```

```go
// math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    // 1. 准备（Arrange）
    a := 2
    b := 3
    expected := 5

    // 2. 执行（Act）
    result := Add(a, b)

    // 3. 断言（Assert）
    if result != expected {
        t.Errorf("Add(%d, %d) = %d; want %d", a, b, result, expected)
    }
}
```

### testing.T 的常用方法

```go
func TestExample(t *testing.T) {
    // 记录错误但继续执行
    t.Error("出错了")
    t.Errorf("出错了: %v", err)

    // 记录错误并立即停止
    t.Fatal("严重错误")
    t.Fatalf("严重错误: %v", err)

    // 记录日志
    t.Log("调试信息")
    t.Logf("调试信息: %v", value)

    // 标记测试失败
    t.Fail()

    // 标记测试失败并停止
    t.FailNow()

    // 跳过测试
    if runtime.GOOS == "windows" {
        t.Skip("在 Windows 上跳过")
    }
}
```

### 子测试

```go
func TestMath(t *testing.T) {
    // 子测试 1
    t.Run("Add", func(t *testing.T) {
        result := Add(2, 3)
        if result != 5 {
            t.Errorf("expected 5, got %d", result)
        }
    })

    // 子测试 2
    t.Run("Subtract", func(t *testing.T) {
        result := Subtract(5, 3)
        if result != 2 {
            t.Errorf("expected 2, got %d", result)
        }
    })
}
```

**运行特定子测试：**
```bash
go test -run TestMath/Add
```

---

## 表驱动测试

表驱动测试是 Go 中最常用的测试模式。

### 基本模式

```go
func TestAdd(t *testing.T) {
    // 定义测试用例表
    tests := []struct {
        name     string  // 测试名称
        a        int     // 输入 a
        b        int     // 输入 b
        expected int     // 期望输出
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
        {"mixed", -2, 3, 1},
        {"zeros", 0, 0, 0},
        {"with zero", 5, 0, 5},
    }

    // 遍历测试用例
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

**好处：**
1. **清晰的测试用例**：所有用例一目了然
2. **易于添加**：只需在表中添加一行
3. **减少重复代码**：测试逻辑只写一次

### 测试错误情况

```go
func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a, b      float64
        expected  float64
        expectErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"divide by zero", 10, 0, 0, true},
        {"negative", -10, 2, -5, false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Divide(tt.a, tt.b)

            // 检查错误
            if tt.expectErr {
                if err == nil {
                    t.Error("expected error but got nil")
                }
                return
            }

            // 检查结果
            if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
            if result != tt.expected {
                t.Errorf("got %f, want %f", result, tt.expected)
            }
        })
    }
}
```

---

## Mock 和桩

### 为什么需要 Mock？

测试时我们不想：
- 真的调用外部 API
- 真的访问数据库
- 真的发送邮件

### 使用接口实现 Mock

```go
// 定义接口
type UserStore interface {
    GetUser(id int) (*User, error)
    SaveUser(user *User) error
}

// 真实实现
type DBUserStore struct {
    db *sql.DB
}

func (s *DBUserStore) GetUser(id int) (*User, error) {
    // 真的查询数据库
}

// Mock 实现
type MockUserStore struct {
    users map[int]*User
}

func (s *MockUserStore) GetUser(id int) (*User, error) {
    user, ok := s.users[id]
    if !ok {
        return nil, errors.New("user not found")
    }
    return user, nil
}

func (s *MockUserStore) SaveUser(user *User) error {
    s.users[user.ID] = user
    return nil
}
```

### 测试时使用 Mock

```go
func TestGetUserInfo(t *testing.T) {
    // 创建 Mock
    mockStore := &MockUserStore{
        users: map[int]*User{
            1: {ID: 1, Name: "张三"},
        },
    }

    // 使用 Mock 进行测试
    service := NewUserService(mockStore)
    user, err := service.GetUserInfo(1)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Name != "张三" {
        t.Errorf("expected 张三, got %s", user.Name)
    }
}
```

### go-todo 中的 Mock 示例

```go
// internal/ai/mock.go
type MockClient struct {
    response string
    err      error
}

func (m *MockClient) Chat(ctx context.Context, messages []Message) (string, error) {
    if m.err != nil {
        return "", m.err
    }
    return m.response, nil
}

// 测试中使用
func TestAICommand(t *testing.T) {
    // 创建 Mock 响应
    mockResponse := `{
        "intent": "create",
        "tasks": [{
            "taskName": "买牛奶",
            "taskDesc": "去超市买两盒牛奶",
            "urgent": "medium"
        }]
    }`

    mockClient := &MockClient{response: mockResponse}

    // 测试...
}
```

---

## 测试覆盖率

### 查看覆盖率

```bash
# 运行测试并生成覆盖率报告
go test -cover

# 输出：
# ok      mypackage    0.123s    coverage: 73.4% of statements
```

### 生成详细报告

```bash
# 生成覆盖率文件
go test -coverprofile=coverage.out

# 查看覆盖率详情
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

### 覆盖率报告示例

```
github.com/SongRunqi/go-todo/app/command.go:63:     DoI             85.7%
github.com/SongRunqi/go-todo/app/command.go:113:    CreateTask      92.3%
github.com/SongRunqi/go-todo/app/command.go:142:    List            100.0%
github.com/SongRunqi/go-todo/app/command.go:162:    Complete        78.6%
total:                                               (statements)    73.4%
```

### 设置覆盖率目标

```bash
# 要求至少 80% 覆盖率
go test -cover -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
# 如果低于 80%，CI 失败
```

---

## 基准测试

基准测试用于性能测试。

### 编写基准测试

```go
// 基准测试函数以 Benchmark 开头
func BenchmarkAdd(b *testing.B) {
    // b.N 会自动调整，直到结果稳定
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

func BenchmarkString Concatenation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        result := ""
        for j := 0; j < 100; j++ {
            result += "a"
        }
    }
}

func BenchmarkStringBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        for j := 0; j < 100; j++ {
            sb.WriteString("a")
        }
        _ = sb.String()
    }
}
```

### 运行基准测试

```bash
# 运行基准测试
go test -bench=.

# 输出：
# BenchmarkAdd-8                  1000000000    0.25 ns/op
# BenchmarkStringConcatenation-8  100000        15234 ns/op
# BenchmarkStringBuilder-8        5000000       256 ns/op
```

**结果解读：**
- `BenchmarkAdd-8`：测试名称，`-8` 表示 GOMAXPROCS=8
- `1000000000`：执行次数（b.N）
- `0.25 ns/op`：每次操作耗时

### 基准测试选项

```bash
# 指定运行时间
go test -bench=. -benchtime=10s

# 显示内存分配
go test -bench=. -benchmem

# 输出：
# BenchmarkAdd-8    1000000000    0.25 ns/op    0 B/op    0 allocs/op
```

### 比较基准测试

```bash
# 第一次测试
go test -bench=. > old.txt

# 修改代码后
go test -bench=. > new.txt

# 使用 benchcmp 比较（需要安装）
go install golang.org/x/tools/cmd/benchcmp@latest
benchcmp old.txt new.txt
```

---

## 调试技巧

### 1. 使用 fmt.Println

最简单的调试方法：

```go
func problematicFunction(x int) int {
    fmt.Println("x =", x)  // 调试输出
    result := x * 2
    fmt.Println("result =", result)
    return result
}
```

### 2. 使用 log 包

比 fmt 更好，带时间戳：

```go
import "log"

func main() {
    log.Println("程序开始")
    log.Printf("处理用户 %d", userID)

    if err != nil {
        log.Fatal("严重错误:", err)  // 打印并退出
    }
}
```

### 3. 使用 Delve 调试器

Delve 是 Go 的官方调试器。

#### 安装 Delve

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

#### 使用 Delve

```bash
# 调试当前包
dlv debug

# 调试测试
dlv test

# 调试指定文件
dlv debug main.go
```

#### Delve 命令

```
(dlv) break main.main    # 在 main 函数设置断点
(dlv) break file.go:10   # 在第 10 行设置断点
(dlv) continue           # 继续执行
(dlv) next               # 下一行（不进入函数）
(dlv) step               # 下一行（进入函数）
(dlv) print x            # 打印变量 x
(dlv) locals             # 打印所有局部变量
(dlv) stack              # 打印调用栈
(dlv) quit               # 退出
```

### 4. 使用 VS Code 调试

**.vscode/launch.json：**

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}"
        }
    ]
}
```

### 5. 使用 pprof 性能分析

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    go func() {
        http.ListenAndServe("localhost:6060", nil)
    }()

    // 你的程序逻辑...
}
```

访问性能分析：
```bash
# CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine 分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

---

## go-todo 中的测试

### 测试文件结构

```
app/
├── command.go
├── command_test.go    # command.go 的测试
├── storage.go
├── storage_test.go    # storage.go 的测试
├── utils.go
└── utils_test.go      # utils.go 的测试
```

### 示例：测试 CreateTask

```go
// app/command_test.go
func TestCreateTask(t *testing.T) {
    tests := []struct {
        name      string
        task      TodoItem
        wantErr   bool
        checkFunc func(t *testing.T, todos []TodoItem)
    }{
        {
            name: "valid task",
            task: TodoItem{
                TaskName: "测试任务",
                TaskDesc: "这是一个测试任务",
                Urgent:   "medium",
            },
            wantErr: false,
            checkFunc: func(t *testing.T, todos []TodoItem) {
                if len(todos) != 1 {
                    t.Errorf("expected 1 todo, got %d", len(todos))
                }
                if todos[0].TaskID != 1 {
                    t.Errorf("expected ID 1, got %d", todos[0].TaskID)
                }
            },
        },
        {
            name: "empty task name",
            task: TodoItem{
                TaskName: "",
                TaskDesc: "描述",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 准备：空的 todos 列表
            todos := []TodoItem{}

            // 执行
            err := CreateTask(&todos, &tt.task)

            // 断言：检查错误
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
                return
            }

            // 断言：自定义检查
            if tt.checkFunc != nil {
                tt.checkFunc(t, todos)
            }
        })
    }
}
```

### 示例：测试存储

```go
// app/storage_test.go
func TestFileTodoStore_SaveAndLoad(t *testing.T) {
    // 创建临时目录
    tmpDir := t.TempDir()
    todoPath := filepath.Join(tmpDir, "todo.json")

    // 创建存储
    store := &FileTodoStore{
        Path:       todoPath,
        BackupPath: filepath.Join(tmpDir, "todo_back.json"),
    }

    // 准备测试数据
    todos := []TodoItem{
        {
            TaskID:   1,
            TaskName: "任务1",
            Status:   "pending",
        },
    }

    // 测试保存
    err := store.Save(&todos, false)
    if err != nil {
        t.Fatalf("Save() error = %v", err)
    }

    // 测试加载
    loaded, err := store.Load(false)
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    // 验证
    if len(loaded) != 1 {
        t.Errorf("expected 1 todo, got %d", len(loaded))
    }
    if loaded[0].TaskID != 1 {
        t.Errorf("expected ID 1, got %d", loaded[0].TaskID)
    }
}
```

### 示例：测试 AI 命令（使用 Mock）

```go
// app/commands_test.go
func TestAICommand_Execute(t *testing.T) {
    // Mock AI 响应
    mockResponse := `{
        "intent": "create",
        "tasks": [{
            "taskId": -1,
            "taskName": "买牛奶",
            "taskDesc": "去超市买两盒牛奶",
            "urgent": "medium",
            "status": "pending"
        }]
    }`

    // 创建上下文
    todos := []TodoItem{}
    tmpDir := t.TempDir()
    store := &FileTodoStore{
        Path:       filepath.Join(tmpDir, "todo.json"),
        BackupPath: filepath.Join(tmpDir, "todo_back.json"),
    }

    ctx := &Context{
        Store: store,
        Todos: &todos,
        Args:  []string{"todo", "买牛奶"},
        Config: &Config{
            APIKey:  "mock-key",
            BaseURL: "http://mock.api",
            Model:   "mock-model",
        },
    }

    // 注入 Mock 客户端
    // (这需要修改代码以支持依赖注入)

    // 执行命令
    cmd := &AICommand{}
    err := cmd.Execute(ctx)

    if err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    // 验证任务已创建
    if len(*ctx.Todos) != 1 {
        t.Errorf("expected 1 todo, got %d", len(*ctx.Todos))
    }
}
```

### 运行 go-todo 的测试

```bash
# 运行所有测试
go test ./...

# 显示详细输出
go test -v ./...

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行特定包的测试
go test ./app

# 运行特定测试
go test -run TestCreateTask ./app
```

---

## 测试最佳实践

### 1. 测试命名清晰

```go
// ✅ 好的命名
func TestCreateTask_WithValidInput_Success(t *testing.T) {}
func TestCreateTask_WithEmptyName_ReturnsError(t *testing.T) {}

// ❌ 不好的命名
func TestCreateTask1(t *testing.T) {}
func TestCreateTask2(t *testing.T) {}
```

### 2. 使用表驱动测试

```go
// ✅ 使用表驱动测试
tests := []struct {
    name string
    // ...
}{ /* ... */ }

// ❌ 重复的测试函数
func TestAdd1(t *testing.T) { /* ... */ }
func TestAdd2(t *testing.T) { /* ... */ }
```

### 3. 测试边界条件

```go
tests := []struct {
    // ...
}{
    {"empty input", /* ... */},
    {"single item", /* ... */},
    {"maximum size", /* ... */},
    {"negative number", /* ... */},
    {"zero value", /* ... */},
}
```

### 4. 使用 t.Helper()

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // 标记为辅助函数，错误会指向调用处
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    result := doSomething()
    assertEqual(t, result, 42)  // 错误会指向这一行
}
```

### 5. 使用 t.Cleanup()

```go
func TestWithCleanup(t *testing.T) {
    // 设置
    file, _ := os.Create("test.txt")

    // 注册清理函数
    t.Cleanup(func() {
        file.Close()
        os.Remove("test.txt")
    })

    // 测试逻辑...
}
```

### 6. 避免测试之间的依赖

```go
// ❌ 不好：测试之间有依赖
var globalData []int

func TestA(t *testing.T) {
    globalData = append(globalData, 1)
}

func TestB(t *testing.T) {
    // 依赖 TestA 的结果
    if len(globalData) == 0 { /* ... */ }
}

// ✅ 好：每个测试独立
func TestA(t *testing.T) {
    data := []int{}
    data = append(data, 1)
}

func TestB(t *testing.T) {
    data := []int{1}
    // 使用本地数据
}
```

---

## 总结

### 测试工具

| 工具 | 用途 |
|------|------|
| `go test` | 运行测试 |
| `go test -cover` | 测试覆盖率 |
| `go test -bench` | 基准测试 |
| `dlv` | 调试器 |
| `pprof` | 性能分析 |

### 测试类型

1. **单元测试**：测试单个函数
2. **集成测试**：测试多个组件协作
3. **基准测试**：性能测试
4. **示例测试**：可执行的文档

### 关键概念

1. **表驱动测试**：Go 的最佳实践
2. **Mock**：使用接口实现测试替身
3. **覆盖率**：衡量测试质量
4. **子测试**：组织相关测试

## 下一步

在最后一课中，我们将学习：
- 如何维护这个项目
- 如何添加新功能
- 常见问题排查
- 部署和分发

继续阅读 `08-maintenance-guide.md`
