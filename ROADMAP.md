# Todo-Go 后续改进任务清单

> 已完成的 Critical 和部分 High Priority 任务，以下是剩余的改进计划

---

## 📋 任务总览

- **High Priority 剩余**: 1 项（Cobra CLI）
- **Medium Priority**: 5 项（日志、验证等）
- **低优先级**: 4 项（项目重组、CI/CD、UX、性能）

---

## 🟠 High Priority - 剩余任务

### ✅ 已完成
- [x] 修复接口实现不一致 ✅
- [x] 移除硬编码路径 ✅
- [x] 消除全局变量 ✅
- [x] 统一错误处理策略 ✅
- [x] 拆分 UpdateTask 函数 ✅
- [x] 添加单元测试（69.7% 覆盖率）✅

### 🔶 待完成

#### 7. 引入 CLI 框架 (Cobra) 🐍
**问题：** 手动字符串解析容易出错且难以扩展
**优先级：** High
**预计时间：** 3-4 小时

**任务清单：**
- [ ] 添加 Cobra 依赖：`go get github.com/spf13/cobra@latest`
- [ ] 创建 `cmd/` 目录结构
- [ ] 实现根命令 (root command)
- [ ] 迁移现有命令到 Cobra 子命令
  - [ ] `list` / `ls` - 列出所有任务
  - [ ] `get <id>` - 获取任务详情
  - [ ] `complete <id>` - 完成任务
  - [ ] `delete <id>` - 删除任务
  - [ ] `update <content>` - 更新任务
  - [ ] `back` - 列出备份任务
  - [ ] `back get <id>` - 获取备份任务
  - [ ] `back restore <id>` - 恢复任务
- [ ] 添加命令帮助文档和使用示例
- [ ] 添加全局 flags（如 `--config`, `--verbose`）
- [ ] 实现参数验证
- [ ] 保留自然语言输入作为默认行为
- [ ] 更新 README 文档
- [ ] 更新测试以适配新结构

**预期命令结构：**
```bash
todo list                    # 列出任务
todo get 1                   # 获取任务
todo complete 1              # 完成任务
todo back                    # 查看备份
todo back restore 1          # 恢复任务
todo "明天写报告"            # 自然语言（AI）
todo --help                  # 帮助信息
```

**文件结构：**
```
cmd/
├── root.go              # 根命令
├── list.go              # list 子命令
├── get.go               # get 子命令
├── complete.go          # complete 子命令
├── delete.go            # delete 子命令
├── update.go            # update 子命令
└── back.go              # back 及其子命令
```

---

## 🟡 Medium Priority - 中优先级任务

### 9. 改进日志系统 📝
**问题：** 日志和用户输出混用，调试信息污染输出
**优先级：** Medium
**预计时间：** 2-3 小时

**任务清单：**
- [ ] 选择日志库（推荐 `zerolog` 或 `zap`）
  - `zerolog`: 更轻量，JSON 格式
  - `zap`: 更强大，性能更好
- [ ] 创建 `internal/logger/` 包
- [ ] 实现日志初始化函数
- [ ] 定义日志级别
  - DEBUG: 详细调试信息
  - INFO: 一般信息
  - WARN: 警告信息
  - ERROR: 错误信息
- [ ] 迁移所有 `log.Println` 到新日志系统
- [ ] 确保用户输出只使用 `fmt` 包
- [ ] 添加环境变量控制日志级别（`LOG_LEVEL`）
- [ ] 添加日志输出目标配置（stdout/file）
- [ ] 更新测试（可能需要 Mock logger）

**示例代码：**
```go
// internal/logger/logger.go
package logger

import "github.com/rs/zerolog"

var Log zerolog.Logger

func Init(level string) {
    // 初始化逻辑
}

// 使用
logger.Log.Debug().Msg("Parsing markdown")
logger.Log.Info().Str("taskId", "1").Msg("Task created")
logger.Log.Error().Err(err).Msg("Failed to save")
```

---

### 10. 添加输入验证层 ✅
**问题：** 缺少对用户输入和函数参数的验证
**优先级：** Medium
**预计时间：** 1.5 小时

**任务清单：**
- [ ] 创建 `internal/validator/` 包
- [ ] 实现验证器工具函数
  - [ ] `ValidateTaskID(id int) error` - ID 必须 > 0
  - [ ] `ValidateTaskName(name string) error` - 非空验证
  - [ ] `ValidateStatus(status string) error` - 枚举验证
  - [ ] `ValidateUrgency(urgent string) error` - 枚举验证
- [ ] 在所有命令入口添加验证
- [ ] 统一验证错误消息格式
- [ ] 添加验证器测试

**示例代码：**
```go
// internal/validator/validator.go
package validator

import "fmt"

func ValidateTaskID(id int) error {
    if id <= 0 {
        return fmt.Errorf("task ID must be greater than 0, got: %d", id)
    }
    return nil
}

var validStatuses = map[string]bool{
    "pending":   true,
    "completed": true,
}

func ValidateStatus(status string) error {
    if !validStatuses[status] {
        return fmt.Errorf("invalid status: %s", status)
    }
    return nil
}
```

---

### 11. 实现 TodoStore 的内存实现（用于测试） 💾
**问题：** 测试依赖文件系统，速度慢且不可靠
**优先级：** Medium
**预计时间：** 1 小时

**任务清单：**
- [ ] 创建 `internal/storage/memory.go`
- [ ] 实现 `MemoryTodoStore` 结构体
- [ ] 实现 `TodoStore` 接口
  - [ ] `Load(backup bool) ([]TodoItem, error)`
  - [ ] `Save(todos []TodoItem, backup bool) error`
- [ ] 使用 `map` 或 `sync.Map` 存储数据
- [ ] 添加测试
- [ ] 更新现有测试以使用内存实现（可选）

**示例代码：**
```go
// internal/storage/memory.go
package storage

import "sync"

type MemoryTodoStore struct {
    data       map[string][]TodoItem
    mu         sync.RWMutex
}

func NewMemoryStore() *MemoryTodoStore {
    return &MemoryTodoStore{
        data: make(map[string][]TodoItem),
    }
}

func (m *MemoryTodoStore) Load(backup bool) ([]TodoItem, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    key := "active"
    if backup {
        key = "backup"
    }

    todos, ok := m.data[key]
    if !ok {
        return []TodoItem{}, nil
    }
    return todos, nil
}
```

---

### 12. 分离 AI 客户端逻辑 🤖
**问题：** AI 调用逻辑耦合在 `api.go` 和 `commands.go` 中
**优先级：** Medium
**预计时间：** 2 小时

**任务清单：**
- [ ] 创建 `internal/ai/` 包
- [ ] 定义 `AIClient` 接口
  ```go
  type AIClient interface {
      Chat(ctx context.Context, messages []Message) (string, error)
  }
  ```
- [ ] 实现 `DeepSeekClient` 结构体
- [ ] 将 prompt 移到配置文件或常量
- [ ] 重构 `AICommand` 使用新接口
- [ ] 创建 `MockAIClient` 用于测试
- [ ] 添加对其他 AI 提供商的支持结构
- [ ] 更新测试

**目录结构：**
```
internal/ai/
├── client.go          # AIClient 接口定义
├── deepseek.go        # DeepSeek 实现
├── mock.go            # Mock 实现（测试用）
├── prompts.go         # Prompt 常量
└── client_test.go     # 测试
```

---

### 13. RestoreTask 的语义修正 🔄
**问题：** 当前 restore 会从 backup 删除任务，与用户预期不符
**优先级：** Medium
**预计时间：** 15 分钟

**任务清单：**
- [ ] 修改 `RestoreTask` 逻辑
  - 选项 A: Restore 时保留 backup 中的记录（复制语义）
  - 选项 B: 添加 `--move` 标志支持移动语义
- [ ] 更新命令帮助文档说明行为
- [ ] 更新测试验证新行为
- [ ] 更新 README 文档

---

## 🟢 Low Priority - 低优先级任务

### 14. 项目结构重组 📁
**优先级：** Low
**预计时间：** 3-4 小时

**任务清单：**
- [ ] 采用标准 Go 项目布局
- [ ] 创建新目录结构
  ```
  todo-go/
  ├── cmd/
  │   └── todo/           # 主程序入口
  │       └── main.go
  ├── internal/           # 私有代码
  │   ├── app/            # 应用逻辑
  │   ├── command/        # 命令实现
  │   ├── domain/         # 领域模型
  │   ├── storage/        # 存储实现
  │   ├── ai/             # AI 客户端
  │   ├── logger/         # 日志系统
  │   └── validator/      # 验证器
  ├── pkg/                # 公共库（可被外部使用）
  │   └── parser/         # 解析器（从 internal 移出）
  ├── configs/            # 配置文件示例
  │   └── config.example.yaml
  ├── scripts/            # 构建脚本
  ├── .github/
  │   └── workflows/      # CI/CD
  ├── go.mod
  ├── go.sum
  ├── README.md
  ├── LICENSE
  └── .gitignore
  ```
- [ ] 迁移文件到新结构
- [ ] 更新 import 路径
- [ ] 更新测试
- [ ] 确保所有测试通过
- [ ] 更新文档

---

### 15. 添加 CI/CD 管道 🚀
**优先级：** Low
**预计时间：** 2-3 小时

**任务清单：**
- [ ] 创建 `.github/workflows/ci.yml`
- [ ] 配置自动化测试
  - [ ] 多 Go 版本测试矩阵 (1.21, 1.22, 1.23)
  - [ ] 运行 `go test ./...`
  - [ ] 生成覆盖率报告
  - [ ] 上传到 Codecov/Coveralls（可选）
- [ ] 配置代码质量检查
  - [ ] `golangci-lint` 运行
  - [ ] `go vet`
  - [ ] `go fmt` 检查
  - [ ] `go mod verify`
- [ ] 配置构建
  - [ ] 多平台构建（Linux, macOS, Windows）
  - [ ] 生成二进制文件
- [ ] 配置自动发布（可选）
  - [ ] Git tags 触发
  - [ ] GitHub Releases
  - [ ] 二进制文件上传
- [ ] 添加 README badges
  - [ ] Build status
  - [ ] Coverage
  - [ ] Go Report Card
  - [ ] License

**示例 CI 配置：**
```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}

    - name: Build
      run: go build -v ./...

    - name: Test
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Coverage
      run: go tool cover -html=coverage.out -o coverage.html

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: golangci/golangci-lint-action@v3
```

---

### 16. 改进用户体验 🎨
**优先级：** Low
**预计时间：** 2-3 小时

**任务清单：**
- [ ] 添加彩色输出支持
  - [ ] 使用 `github.com/fatih/color`
  - [ ] 成功消息：绿色
  - [ ] 错误消息：红色
  - [ ] 警告消息：黄色
  - [ ] 任务标题：粗体/蓝色
- [ ] 改进错误消息的友好度
  - [ ] 提供可操作的建议
  - [ ] 包含示例用法
- [ ] 添加进度指示器（长时间操作）
  - [ ] 使用 `github.com/schollz/progressbar`
  - [ ] AI 调用时显示 spinner
- [ ] 支持配置文件（YAML/TOML）
  - [ ] `.todo.yaml` 或 `~/.config/todo/config.yaml`
  - [ ] 加载优先级：CLI flags > 环境变量 > 配置文件 > 默认值
- [ ] 添加交互式模式（可选）
  - [ ] `todo interactive`
  - [ ] 菜单选择
- [ ] 添加 shell 补全脚本
  - [ ] Bash
  - [ ] Zsh
  - [ ] Fish

**示例彩色输出：**
```go
import "github.com/fatih/color"

// 成功消息
color.Green("✓ Task %d created successfully\n", taskID)

// 错误消息
color.Red("✗ Error: Task %d not found\n", taskID)

// 任务标题
color.New(color.FgCyan, color.Bold).Println(task.TaskName)
```

---

### 17. 性能优化 ⚡
**优先级：** Low
**预计时间：** 2-3 小时

**任务清单：**
- [ ] 性能分析
  - [ ] 使用 `go test -bench` 创建基准测试
  - [ ] 使用 `pprof` 分析性能瓶颈
- [ ] 优化建议（根据实际情况）
  - [ ] 使用 `sync.Pool` 优化频繁分配（如果需要）
  - [ ] 添加缓存层（如果文件读取频繁）
  - [ ] 使用 goroutine 并发处理多任务操作
  - [ ] 优化 JSON 序列化（考虑使用 `easyjson`）
- [ ] 添加性能基准测试
  - [ ] `BenchmarkList`
  - [ ] `BenchmarkCreate`
  - [ ] `BenchmarkParse`
- [ ] 文档化性能特性

**示例基准测试：**
```go
func BenchmarkList(b *testing.B) {
    // 准备数据
    todos := make([]TodoItem, 100)
    for i := 0; i < 100; i++ {
        todos[i] = TodoItem{
            TaskID: i + 1,
            TaskName: fmt.Sprintf("Task %d", i),
            EndTime: time.Now().Add(time.Hour),
        }
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        List(&todos)
    }
}
```

---

## 📊 任务执行建议

### Phase 1: 高优先级剩余任务（1 周）
1. ✅ 引入 Cobra CLI 框架（3-4 小时）

### Phase 2: 中优先级任务（1-2 周）
1. 改进日志系统（2-3 小时）
2. 添加输入验证层（1.5 小时）
3. 分离 AI 客户端逻辑（2 小时）
4. 实现内存存储（1 小时）
5. 修正 RestoreTask 语义（15 分钟）

### Phase 3: 低优先级任务（可选，1-2 周）
1. 项目结构重组（3-4 小时）
2. CI/CD 配置（2-3 小时）
3. UX 改进（2-3 小时）
4. 性能优化（2-3 小时）

---

## 🎯 关键里程碑

- [ ] **Milestone 1**: Cobra CLI 集成完成
- [ ] **Milestone 2**: 日志和验证系统完善
- [ ] **Milestone 3**: AI 客户端模块化
- [ ] **Milestone 4**: 标准项目结构
- [ ] **Milestone 5**: CI/CD 和自动化完善
- [ ] **Milestone 6**: 代码质量达到 9/10 分

---

## 📈 质量目标

- **测试覆盖率**: 保持 > 70%
- **代码质量**: Go Report Card A+
- **文档完整性**: 100%
- **性能**: 所有操作 < 100ms (除 AI 调用)
- **可维护性**: 平均函数长度 < 50 行
- **可扩展性**: 易于添加新命令和功能

---

## 🔗 相关资源

- **Cobra 文档**: https://cobra.dev/
- **Zerolog 文档**: https://github.com/rs/zerolog
- **Go 项目布局**: https://github.com/golang-standards/project-layout
- **golangci-lint**: https://golangci-lint.run/
- **Codecov**: https://about.codecov.io/

---

**创建日期**: 2025-11-05
**最后更新**: 2025-11-05
**进度**: High Priority 6/7 完成 (85.7%)
