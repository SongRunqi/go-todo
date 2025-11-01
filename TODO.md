# Todo-Go 重构任务清单

> 根据架构评审结果，按优先级和严重程度整理的改进任务

---

## 🔴 Critical - 必须立即修复（阻塞性问题）

### 1. 修复接口实现不一致 ⚠️ BLOCKING
**问题：** `TodoStore` 接口定义与 `FileTodoStore` 实现签名不匹配
- [ ] 修改 `FileTodoStore.Load()` 方法，返回 `([]TodoItem, error)`
- [ ] 修改所有调用 `Load()` 的地方，处理返回的 error
- [ ] 确保接口实现完全匹配接口定义

**文件：** `storage.go`, `types.go`
**预计时间：** 30 分钟
**影响：** 编译错误级别问题

---

### 2. 移除硬编码路径 🔧
**问题：** 代码无法在其他环境运行，无法测试
- [ ] 创建 `config.go` 文件
- [ ] 实现配置加载函数（支持环境变量 + 默认值）
- [ ] 从 `storage.go` 中移除硬编码的路径常量
- [ ] 更新 `main.go`，从配置读取路径

**相关环境变量：**
```bash
TODO_PATH=/path/to/todo.json
TODO_BACKUP_PATH=/path/to/todo_back.json
```

**文件：** 新建 `config.go`, 修改 `storage.go`, `main.go`
**预计时间：** 1 小时
**影响：** 可移植性、可测试性

---

### 3. 消除全局变量 🌍
**问题：** `fileTodoStore` 全局变量导致代码难以测试和并发不安全
- [ ] 移除 `main.go` 中的 `var fileTodoStore FileTodoStore`
- [ ] 重构所有使用 `fileTodoStore` 的函数，通过参数传递
- [ ] 在 `main()` 函数中创建 store 实例并传递

**影响函数：**
- `DoI()`
- `Complete()`
- `DeleteTask()`
- `RestoreTask()`

**文件：** `main.go`, `command.go`
**预计时间：** 1.5 小时
**影响：** 可测试性、并发安全

---

### 4. 统一错误处理策略 ⚡
**问题：** 错误处理不一致，有些函数吞掉错误，有些正确返回
- [ ] 为所有命令函数添加 `error` 返回值
  - `Complete()`
  - `CreateTask()`
  - `List()`
  - `GetTask()`
  - `UpdateTask()`
  - `DeleteTask()`
  - `RestoreTask()`
- [ ] 在 `main.go` 中统一处理错误
- [ ] 移除函数内部的 `fmt.Printf` 错误输出，改为返回 error
- [ ] 使用 `fmt.Errorf` 和 `%w` 包装错误上下文

**文件：** `command.go`, `main.go`
**预计时间：** 2 小时
**影响：** 可维护性、错误追踪

---

## 🟠 High Priority - 高优先级（严重影响代码质量）

### 5. 重构 main.go - 引入命令模式 🎯
**问题：** `main.go` 102 行代码混杂多种职责，if-else 嵌套过深
- [ ] 创建 `Command` 接口
  ```go
  type Command interface {
      Execute(ctx *Context) error
  }
  ```
- [ ] 实现具体命令结构体：
  - [ ] `ListCommand`
  - [ ] `BackCommand`
  - [ ] `BackGetCommand`
  - [ ] `BackRestoreCommand`
  - [ ] `CompleteCommand`
  - [ ] `DeleteCommand`
  - [ ] `GetCommand`
  - [ ] `UpdateCommand`
  - [ ] `AICommand` (处理自然语言输入)
- [ ] 创建 `Router` 结构负责命令分发
- [ ] 重构 `main.go`，精简到 20 行以内

**建议结构：**
```
internal/
  ├── command/
  │   ├── interface.go    (Command 接口)
  │   ├── router.go       (命令路由)
  │   ├── list.go
  │   ├── complete.go
  │   ├── restore.go
  │   └── ...
```

**文件：** `main.go`, 新建 `internal/command/` 目录
**预计时间：** 4 小时
**影响：** 可维护性、可扩展性

---

### 6. 拆分 UpdateTask 函数 ✂️
**问题：** 200+ 行的函数违反单一职责原则
- [ ] 创建 `parser` 包
- [ ] 实现 `ParseMarkdown(string) (TodoItem, error)` 函数
- [ ] 实现 `ParseJSON(string) (TodoItem, error)` 函数
- [ ] 实现 `ParseCompactFormat(string) (TodoItem, error)` 函数
- [ ] 重构 `UpdateTask`，使用解析器 + 更新逻辑

**文件：** 新建 `pkg/parser/`, 修改 `command.go`
**预计时间：** 2 小时
**影响：** 可读性、可测试性

---

### 7. 引入 CLI 框架 (Cobra) 🐍
**问题：** 手动字符串解析容易出错且难以扩展
- [ ] 添加依赖：`go get github.com/spf13/cobra`
- [ ] 创建根命令和子命令结构
- [ ] 迁移现有命令到 Cobra 命令
- [ ] 添加命令帮助文档和参数验证

**示例结构：**
```
todo list
todo get <id>
todo complete <id>
todo back
todo back get <id>
todo back restore <id>
```

**文件：** 新建 `cmd/` 目录, 修改 `main.go`
**预计时间：** 3 小时
**影响：** 用户体验、可维护性

---

### 8. 添加单元测试（测试覆盖率 0% → 60%+） 🧪
**问题：** 没有任何测试，重构风险极高
- [ ] 为 `FileTodoStore` 编写测试
  - [ ] `TestLoad`
  - [ ] `TestSave`
- [ ] 为命令函数编写测试
  - [ ] `TestComplete`
  - [ ] `TestRestoreTask`
  - [ ] `TestCreateTask`
- [ ] 为解析器编写测试（完成任务 #6 后）
- [ ] 为 AI 客户端编写 Mock 测试
- [ ] 配置 GitHub Actions 运行测试

**文件：** 新建 `*_test.go` 文件
**预计时间：** 6 小时
**影响：** 代码质量、重构安全性

---

## 🟡 Medium Priority - 中优先级（提升代码质量）

### 9. 改进日志系统 📝
**问题：** 日志和用户输出混用，调试信息污染输出
- [ ] 引入结构化日志库（推荐 `zap` 或 `zerolog`）
- [ ] 区分日志级别（DEBUG, INFO, ERROR）
- [ ] 将所有 `log.Println` 迁移到新日志系统
- [ ] 确保用户输出只使用 `fmt` 包
- [ ] 支持通过环境变量控制日志级别

**文件：** 新建 `pkg/logger/`, 修改所有文件
**预计时间：** 2 小时

---

### 10. 添加输入验证层 ✅
**问题：** 缺少对用户输入和函数参数的验证
- [ ] 创建验证器工具函数
- [ ] 在所有命令入口验证参数
  - [ ] ID 范围验证（必须 > 0）
  - [ ] 字符串非空验证
  - [ ] Slice 非 nil 验证
- [ ] 统一验证错误消息格式

**文件：** 新建 `pkg/validator/`, 修改 `command.go`
**预计时间：** 1.5 小时

---

### 11. 实现 TodoStore 的内存实现（用于测试） 💾
**问题：** 测试依赖文件系统，速度慢且不可靠
- [ ] 创建 `MemoryTodoStore` 实现 `TodoStore` 接口
- [ ] 使用 map 存储数据
- [ ] 在单元测试中使用内存实现替代文件实现

**文件：** 新建 `storage/memory.go`
**预计时间：** 1 小时

---

### 12. 分离 AI 客户端逻辑 🤖
**问题：** AI 调用逻辑耦合在 `api.go` 和 `main.go` 中
- [ ] 创建 `AIClient` 接口
- [ ] 实现 `OpenAIClient` 结构体
- [ ] 将 prompt 移到配置文件
- [ ] 支持不同 AI 提供商（通过接口）

**文件：** 新建 `internal/ai/`, 修改 `api.go`, `main.go`
**预计时间：** 2 小时

---

### 13. RestoreTask 的语义修正 🔄
**问题：** 当前 restore 会从 backup 删除任务，与用户预期不符
- [ ] 修改 `RestoreTask` 逻辑，restore 时保留 backup 中的原始记录
- [ ] 可选：添加 `--move` 标志支持移动语义（删除 backup）

**文件：** `command.go`
**预计时间：** 15 分钟

---

## 🟢 Low Priority - 低优先级（锦上添花）

### 14. 项目结构重组 📁
- [ ] 采用标准 Go 项目布局
  ```
  todo-go/
  ├── cmd/todo/           # 主程序入口
  ├── internal/           # 私有代码
  │   ├── app/
  │   ├── command/
  │   ├── domain/
  │   ├── storage/
  │   └── ai/
  ├── pkg/                # 公共库
  │   ├── parser/
  │   └── validator/
  ├── configs/            # 配置示例
  ├── go.mod
  └── README.md
  ```

**预计时间：** 2 小时

---

### 15. 添加 CI/CD 管道 🚀
- [ ] 创建 `.github/workflows/ci.yml`
- [ ] 配置自动化测试
- [ ] 配置代码质量检查（golangci-lint）
- [ ] 配置自动发布

**预计时间：** 1 小时

---

### 16. 改进用户体验 🎨
- [ ] 添加彩色输出支持（使用 `fatih/color`）
- [ ] 改进错误消息的友好度
- [ ] 添加进度指示器（长时间操作）
- [ ] 支持配置文件（YAML/TOML）

**预计时间：** 2 小时

---

### 17. 性能优化 ⚡
- [ ] 使用 `sync.Pool` 优化内存分配（如果需要）
- [ ] 添加缓存层（如果文件读取频繁）
- [ ] 使用 goroutine 并发处理多任务操作

**预计时间：** 2 小时

---

### 18. 文档完善 📚
- [ ] 编写详细的 README.md
  - [ ] 安装说明
  - [ ] 使用示例
  - [ ] 配置说明
  - [ ] 架构图
- [ ] 添加代码注释（GoDoc 风格）
- [ ] 创建 CONTRIBUTING.md
- [ ] 添加架构设计文档（ADR）

**预计时间：** 3 小时

---

## 📊 执行建议

### Phase 1: 基础修复（1周）
按顺序完成 Critical 任务 #1-4，确保代码基础稳固。

### Phase 2: 架构重构（2周）
完成 High Priority 任务 #5-8，大幅提升代码质量。

### Phase 3: 质量提升（1周）
完成 Medium Priority 任务 #9-13。

### Phase 4: 锦上添花（可选）
根据实际需求选择性完成 Low Priority 任务。

---

## 🎯 关键里程碑

- [ ] **Milestone 1**: 修复所有 Critical 问题，代码可以编译运行
- [ ] **Milestone 2**: 完成命令模式重构，通过 20+ 单元测试
- [ ] **Milestone 3**: 测试覆盖率达到 60%+
- [ ] **Milestone 4**: 重构完成，代码评分达到 8/10

---

## 📝 备注

- 每完成一个任务，建议创建一个 git commit
- 重构前务必确保有足够的测试覆盖
- 可以使用功能分支开发，避免直接在 main 分支修改
- 建议使用 TDD（测试驱动开发）方式进行重构

---

**最后更新：** 2025-11-02
**评审人：** Claude (Architecture Review)