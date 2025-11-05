# Cobra CLI 框架详解

## 目录
1. [什么是 Cobra](#什么是-cobra)
2. [Cobra 的核心概念](#cobra-的核心概念)
3. [创建第一个 Cobra 应用](#创建第一个-cobra-应用)
4. [命令（Commands）](#命令commands)
5. [标志（Flags）](#标志flags)
6. [参数（Arguments）](#参数arguments)
7. [子命令](#子命令)
8. [生命周期钩子](#生命周期钩子)
9. [自动补全](#自动补全)
10. [go-todo 中的 Cobra 使用](#go-todo-中的-cobra-使用)

---

## 什么是 Cobra

Cobra 是 Go 语言最流行的 CLI（命令行界面）框架。

### 谁在使用 Cobra？

许多著名的 Go 项目都使用 Cobra：
- **Kubernetes** (`kubectl`)
- **Docker** (`docker`)
- **GitHub CLI** (`gh`)
- **Hugo**（静态网站生成器）
- **Terraform**

### 为什么选择 Cobra？

1. **易于使用**：简单的 API，快速构建 CLI
2. **功能强大**：支持子命令、标志、参数
3. **自动生成**：帮助信息、shell 补全
4. **标准化**：遵循 POSIX 标准
5. **活跃维护**：由 spf13（Steve Francia）创建和维护

### 安装 Cobra

```bash
# 添加依赖
go get -u github.com/spf13/cobra@latest

# 安装 cobra-cli 工具（可选，用于快速生成代码）
go install github.com/spf13/cobra-cli@latest
```

---

## Cobra 的核心概念

### 三个核心组件

```
应用程序
└── Commands（命令）
    ├── Flags（标志）
    └── Args（参数）
```

#### 1. Commands（命令）

命令是应用的行为，例如：

```bash
git clone    # clone 是命令
docker run   # run 是命令
kubectl get  # get 是命令
```

#### 2. Flags（标志）

标志修改命令的行为，以 `-` 或 `--` 开头：

```bash
git clone --depth 1           # --depth 是标志
docker run -d nginx           # -d 是标志
kubectl get pods --all-namespaces  # --all-namespaces 是标志
```

#### 3. Args（参数）

参数是传递给命令的值：

```bash
git clone https://github.com/user/repo  # URL 是参数
docker run nginx                         # nginx 是参数
kubectl get pods my-pod                  # my-pod 是参数
```

### 命令结构示例

```bash
todo list --status pending
│    │    │      │
│    │    │      └─ 标志值
│    │    └──────── 标志
│    └───────────── 子命令
└────────────────── 根命令
```

---

## 创建第一个 Cobra 应用

### 最简单的 Cobra 应用

```go
package main

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"
)

func main() {
    // 创建根命令
    var rootCmd = &cobra.Command{
        Use:   "hello",
        Short: "Hello 是一个简单的 CLI",
        Long:  `这是一个使用 Cobra 构建的简单命令行应用。`,
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Hello, Cobra!")
        },
    }

    // 执行命令
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
```

运行：

```bash
$ go run main.go
Hello, Cobra!
```

### 标准项目结构

Cobra 应用通常使用以下结构：

```
myapp/
├── cmd/
│   ├── root.go      # 根命令
│   ├── list.go      # list 子命令
│   ├── get.go       # get 子命令
│   └── delete.go    # delete 子命令
├── main.go          # 程序入口
└── go.mod
```

**main.go：**
```go
package main

import "myapp/cmd"

func main() {
    cmd.Execute()
}
```

**cmd/root.go：**
```go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "我的应用程序",
    Long:  `这是我的应用程序的详细描述。`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    // 在这里初始化标志和配置
}
```

---

## 命令（Commands）

### Command 结构

```go
type Command struct {
    Use   string   // 命令的使用方式
    Short string   // 简短描述
    Long  string   // 详细描述
    Run   func(cmd *Command, args []string)  // 执行函数
    // ... 更多字段
}
```

### 创建命令

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出所有项目",
    Long:  `列出所有项目的详细信息。`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("列出所有项目...")
        // 实际逻辑
    },
}
```

### 命令字段说明

#### Use（必需）

定义命令的使用方式：

```go
Use: "clone <url>"    // clone 命令需要一个 url 参数
Use: "list [filter]"  // list 命令有一个可选的 filter 参数
```

#### Aliases（别名）

为命令创建别名：

```go
var listCmd = &cobra.Command{
    Use:     "list",
    Aliases: []string{"ls", "l"},  // 可以用 ls 或 l 代替 list
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("列表...")
    },
}
```

```bash
$ myapp list     # 正常使用
$ myapp ls       # 使用别名
$ myapp l        # 使用别名
```

#### Example

提供使用示例：

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出项目",
    Example: `  myapp list
  myapp list --status pending
  myapp list --limit 10`,
    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}
```

#### RunE（带错误处理的 Run）

```go
var getCmd = &cobra.Command{
    Use:   "get <id>",
    Short: "获取项目",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) < 1 {
            return fmt.Errorf("需要提供 ID")
        }
        // 返回错误会自动打印并退出
        return processItem(args[0])
    },
}
```

**Run vs RunE：**
- `Run`：不返回错误
- `RunE`：返回错误，Cobra 会自动处理

---

## 标志（Flags）

标志用于修改命令的行为。

### 本地标志（Local Flags）

只对当前命令有效：

```go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出项目",
    Run: func(cmd *cobra.Command, args []string) {
        limit, _ := cmd.Flags().GetInt("limit")
        fmt.Printf("限制: %d\n", limit)
    },
}

func init() {
    // 添加本地标志
    listCmd.Flags().IntP("limit", "l", 10, "限制结果数量")
}
```

使用：

```bash
$ myapp list --limit 20
$ myapp list -l 20
```

### 持久标志（Persistent Flags）

对当前命令及其所有子命令有效：

```go
var rootCmd = &cobra.Command{
    Use: "myapp",
}

func init() {
    // 添加持久标志
    rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出")
}
```

现在所有命令都可以使用 `--verbose`：

```bash
$ myapp --verbose
$ myapp list --verbose
$ myapp get 1 --verbose
```

### 标志类型

Cobra 支持多种标志类型：

```go
// 字符串
cmd.Flags().String("name", "", "名称")
cmd.Flags().StringP("name", "n", "", "名称")  // 带短标志

// 整数
cmd.Flags().Int("count", 0, "数量")
cmd.Flags().IntP("count", "c", 0, "数量")

// 布尔值
cmd.Flags().Bool("force", false, "强制执行")
cmd.Flags().BoolP("force", "f", false, "强制执行")

// 字符串切片
cmd.Flags().StringSlice("tags", []string{}, "标签列表")

// 其他类型
cmd.Flags().Float64("rate", 0.0, "比率")
cmd.Flags().Duration("timeout", 0, "超时时间")
```

### 绑定标志到变量

```go
var (
    name    string
    count   int
    verbose bool
)

func init() {
    // 直接绑定到变量
    rootCmd.Flags().StringVar(&name, "name", "", "名称")
    rootCmd.Flags().IntVar(&count, "count", 0, "数量")
    rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
}

var rootCmd = &cobra.Command{
    Use: "myapp",
    Run: func(cmd *cobra.Command, args []string) {
        // 直接使用变量
        fmt.Printf("名称: %s, 数量: %d, 详细: %v\n", name, count, verbose)
    },
}
```

### 必需标志

```go
func init() {
    cmd.Flags().String("name", "", "名称")
    cmd.MarkFlagRequired("name")  // 标记为必需
}
```

### 标志依赖

```go
func init() {
    cmd.Flags().String("username", "", "用户名")
    cmd.Flags().String("password", "", "密码")

    // password 依赖 username
    cmd.MarkFlagsRequiredTogether("username", "password")
}
```

---

## 参数（Arguments）

参数是传递给命令的位置参数。

### 参数验证

Cobra 提供了内置的参数验证器：

```go
var getCmd = &cobra.Command{
    Use:   "get <id>",
    Short: "获取项目",
    Args:  cobra.ExactArgs(1),  // 必须正好 1 个参数
    Run: func(cmd *cobra.Command, args []string) {
        id := args[0]
        fmt.Printf("获取项目: %s\n", id)
    },
}
```

### 内置验证器

```go
// 不接受任何参数
Args: cobra.NoArgs

// 必须有至少 N 个参数
Args: cobra.MinimumNArgs(1)

// 最多 N 个参数
Args: cobra.MaximumNArgs(2)

// 正好 N 个参数
Args: cobra.ExactArgs(1)

// 接受任意数量的参数（默认）
Args: cobra.ArbitraryArgs

// 只接受有效的参数（从 ValidArgs 中）
Args: cobra.OnlyValidArgs

// 组合多个验证器
Args: cobra.MatchAll(
    cobra.ExactArgs(2),
    cobra.OnlyValidArgs,
)
```

### 自定义验证器

```go
var getCmd = &cobra.Command{
    Use:  "get <id>",
    Args: func(cmd *cobra.Command, args []string) error {
        if len(args) != 1 {
            return fmt.Errorf("需要正好 1 个参数")
        }
        // 验证 ID 是数字
        if _, err := strconv.Atoi(args[0]); err != nil {
            return fmt.Errorf("ID 必须是数字")
        }
        return nil
    },
    Run: func(cmd *cobra.Command, args []string) {
        // ...
    },
}
```

---

## 子命令

### 添加子命令

```go
// cmd/root.go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "我的应用",
}

func init() {
    // 添加子命令
    rootCmd.AddCommand(listCmd)
    rootCmd.AddCommand(getCmd)
    rootCmd.AddCommand(deleteCmd)
}

// cmd/list.go
var listCmd = &cobra.Command{
    Use:   "list",
    Short: "列出项目",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("列出项目...")
    },
}
```

使用：

```bash
$ myapp list
$ myapp get 1
$ myapp delete 1
```

### 嵌套子命令

```go
var backupCmd = &cobra.Command{
    Use:   "backup",
    Short: "备份管理",
}

var backupListCmd = &cobra.Command{
    Use:   "list",
    Short: "列出备份",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("列出备份...")
    },
}

var backupRestoreCmd = &cobra.Command{
    Use:   "restore <id>",
    Short: "恢复备份",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("恢复备份: %s\n", args[0])
    },
}

func init() {
    // 添加子命令到 backup
    backupCmd.AddCommand(backupListCmd)
    backupCmd.AddCommand(backupRestoreCmd)

    // 添加 backup 到根命令
    rootCmd.AddCommand(backupCmd)
}
```

使用：

```bash
$ myapp backup list
$ myapp backup restore 1
```

---

## 生命周期钩子

Cobra 提供了多个钩子函数，在命令执行的不同阶段运行。

### 钩子执行顺序

```
PersistentPreRun  (父命令)
    ↓
PersistentPreRun  (当前命令)
    ↓
PreRun            (当前命令)
    ↓
Run               (当前命令)
    ↓
PostRun           (当前命令)
    ↓
PersistentPostRun (当前命令)
    ↓
PersistentPostRun (父命令)
```

### 使用示例

```go
var rootCmd = &cobra.Command{
    Use: "myapp",

    // 在任何命令之前运行（包括子命令）
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("1. Root PersistentPreRun")
        // 初始化配置、数据库连接等
    },

    // 在命令之前运行
    PreRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("2. Root PreRun")
    },

    // 命令的主要逻辑
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("3. Root Run")
    },

    // 在命令之后运行
    PostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("4. Root PostRun")
    },

    // 在任何命令之后运行（包括子命令）
    PersistentPostRun: func(cmd *cobra.Command, args []string) {
        fmt.Println("5. Root PersistentPostRun")
        // 清理资源
    },
}
```

### 带错误处理的钩子

```go
var rootCmd = &cobra.Command{
    Use: "myapp",

    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        // 初始化数据库
        if err := initDatabase(); err != nil {
            return fmt.Errorf("数据库初始化失败: %w", err)
        }
        return nil
    },

    RunE: func(cmd *cobra.Command, args []string) error {
        // 主逻辑
        return nil
    },
}
```

---

## 自动补全

Cobra 自动生成 Shell 补全脚本。

### 添加补全命令

```go
var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "生成 shell 补全脚本",
    ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        switch args[0] {
        case "bash":
            cmd.Root().GenBashCompletion(os.Stdout)
        case "zsh":
            cmd.Root().GenZshCompletion(os.Stdout)
        case "fish":
            cmd.Root().GenFishCompletion(os.Stdout, true)
        case "powershell":
            cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
        }
    },
}

func init() {
    rootCmd.AddCommand(completionCmd)
}
```

### 使用补全

```bash
# Bash
$ source <(myapp completion bash)

# Zsh
$ source <(myapp completion zsh)

# Fish
$ myapp completion fish | source
```

---

## go-todo 中的 Cobra 使用

### 项目结构

```
go-todo/
├── main.go
└── cmd/
    ├── root.go       # 根命令
    ├── list.go       # list 命令
    ├── get.go        # get 命令
    ├── complete.go   # complete 命令
    ├── delete.go     # delete 命令
    ├── update.go     # update 命令
    └── back.go       # back 命令（备份管理）
```

### main.go

```go
package main

import "github.com/SongRunqi/go-todo/cmd"

func main() {
    cmd.Execute()
}
```

非常简洁！所有逻辑都在 `cmd` 包中。

### cmd/root.go（简化版）

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/SongRunqi/go-todo/app"
)

var rootCmd = &cobra.Command{
    Use:   "todo [natural language input]",
    Short: "AI-powered todo management CLI",
    Long:  `Todo-Go 是一个 AI 驱动的命令行待办事项管理应用。`,

    // 在所有命令之前执行
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 初始化日志
        logger.Init(logLevel)

        // 加载配置
        config = app.LoadConfig()

        // 初始化存储
        store = &app.FileTodoStore{
            Path:       config.TodoPath,
            BackupPath: config.BackupPath,
        }

        // 加载待办事项
        todos, _ = store.Load(false)
    },

    // 根命令的逻辑
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) > 0 {
            // 处理自然语言输入
            handleNaturalLanguage(args)
        } else {
            cmd.Help()
        }
    },
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**关键点：**

1. **PersistentPreRun**：在所有命令之前执行
   - 初始化日志
   - 加载配置
   - 加载待办事项

2. **Run**：处理自然语言输入或显示帮助

### cmd/list.go

```go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/SongRunqi/go-todo/app"
)

var listCmd = &cobra.Command{
    Use:     "list",
    Aliases: []string{"ls", "l"},  // 别名
    Short:   "List all todos",
    Run: func(cmd *cobra.Command, args []string) {
        ctx := &app.Context{
            Store: store,
            Todos: todos,
            Args:  append([]string{"todo", "list"}, args...),
        }

        // 执行 list 命令
        listCommand := &app.ListCommand{}
        if err := listCommand.Execute(ctx); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

func init() {
    // 将 listCmd 添加到 rootCmd
    rootCmd.AddCommand(listCmd)
}
```

**关键点：**

1. **Aliases**：`ls` 和 `l` 都可以使用
2. **init()**：将命令添加到根命令
3. **分离业务逻辑**：实际逻辑在 `app.ListCommand` 中

### cmd/get.go

```go
var getCmd = &cobra.Command{
    Use:   "get <id>",
    Short: "Get todo by ID",
    Args:  cobra.ExactArgs(1),  // 必须有 1 个参数
    Run: func(cmd *cobra.Command, args []string) {
        id, err := strconv.Atoi(args[0])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Invalid ID: %v\n", err)
            os.Exit(1)
        }

        ctx := &app.Context{
            Store: store,
            Todos: todos,
            Args:  []string{"todo", "get", strconv.Itoa(id)},
        }

        getCommand := &app.GetCommand{}
        if err := getCommand.Execute(ctx); err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    },
}

func init() {
    rootCmd.AddCommand(getCmd)
}
```

**关键点：**

1. **Args: cobra.ExactArgs(1)**：验证参数数量
2. **参数解析**：将字符串 ID 转换为整数
3. **错误处理**：无效 ID 时退出

### cmd/back.go（嵌套命令）

```go
// 主 backup 命令
var backCmd = &cobra.Command{
    Use:     "back",
    Aliases: []string{"backup", "b"},
    Short:   "Backup management commands",
}

// 列出备份
var backListCmd = &cobra.Command{
    Use:     "list",
    Aliases: []string{"ls"},
    Short:   "List backup todos",
    Run: func(cmd *cobra.Command, args []string) {
        // 逻辑...
    },
}

// 恢复备份
var backRestoreCmd = &cobra.Command{
    Use:   "restore <id>",
    Short: "Restore todo from backup",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // 逻辑...
    },
}

func init() {
    // 添加子命令到 back
    backCmd.AddCommand(backListCmd)
    backCmd.AddCommand(backGetCmd)
    backCmd.AddCommand(backRestoreCmd)

    // 添加 back 到根命令
    rootCmd.AddCommand(backCmd)
}
```

**使用：**

```bash
$ todo back list          # 列出备份
$ todo back get 1         # 查看备份
$ todo back restore 1     # 恢复备份
```

### 未知命令的处理

go-todo 有一个巧妙的设计：当用户输入未知命令时，将其作为自然语言处理。

```go
func Execute() {
    rootCmd.SilenceErrors = true
    rootCmd.SilenceUsage = true

    if err := rootCmd.Execute(); err != nil {
        errStr := err.Error()
        if strings.Contains(errStr, "unknown command") {
            // 将未知命令作为自然语言处理
            handleNaturalLanguage(os.Args[1:])
            return
        }
        // 其他错误正常处理
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

**效果：**

```bash
$ todo list              # 执行 list 命令
$ todo buy milk          # 未知命令，作为自然语言处理
$ todo "完成报告"         # 自然语言
```

---

## 最佳实践

### 1. 命令设计原则

```go
// ✅ 好的命令设计
Use:   "get <id>"        // 清晰的参数
Use:   "list [filter]"   // 可选参数用方括号

// ❌ 不好的命令设计
Use:   "get"             // 缺少参数说明
```

### 2. 提供有用的帮助信息

```go
var listCmd = &cobra.Command{
    Use:   "list [flags]",
    Short: "列出所有待办事项",  // 简短描述
    Long: `列出所有待办事项的详细信息。

可以使用标志过滤和排序结果。`,  // 详细描述

    Example: `  todo list
  todo list --status pending
  todo list --limit 10`,  // 使用示例
}
```

### 3. 使用别名提高便利性

```go
Aliases: []string{"ls", "l"}  // list -> ls -> l
Aliases: []string{"rm"}       // delete -> rm
```

### 4. 分离业务逻辑

```go
// ❌ 不好：业务逻辑在 cmd 中
var listCmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        // 100 行业务逻辑...
    },
}

// ✅ 好：业务逻辑在单独的包中
var listCmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        app.ListTodos(store, config)
    },
}
```

### 5. 使用 PersistentPreRun 进行初始化

```go
var rootCmd = &cobra.Command{
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // 所有命令都需要的初始化
        initLogger()
        loadConfig()
        connectDatabase()
    },
}
```

### 6. 提供有意义的错误信息

```go
// ❌ 不好
return fmt.Errorf("错误")

// ✅ 好
return fmt.Errorf("无法找到 ID 为 %d 的待办事项", id)
```

### 7. 支持 Shell 补全

```go
// 添加 completion 命令
rootCmd.AddCommand(completionCmd)

// 提供有效参数
var getCmd = &cobra.Command{
    ValidArgs: []string{"1", "2", "3"},
}
```

---

## 总结

### Cobra 核心概念

1. **Commands**：应用的行为
2. **Flags**：修改命令的行为
3. **Args**：位置参数

### 关键功能

1. **子命令**：`rootCmd.AddCommand(subCmd)`
2. **标志**：本地标志和持久标志
3. **参数验证**：`Args: cobra.ExactArgs(1)`
4. **生命周期钩子**：`PersistentPreRun`, `Run`, `PostRun`
5. **自动补全**：`GenBashCompletion`, `GenZshCompletion`

### go-todo 中的应用

1. **结构化命令**：`list`, `get`, `complete`, `delete`, `update`, `back`
2. **嵌套命令**：`back list`, `back restore`
3. **自然语言回退**：未知命令作为 AI 输入处理
4. **统一初始化**：`PersistentPreRun` 加载配置和数据

## 下一步

在下一课中，我们将深入分析 go-todo 项目：
- 项目整体架构
- 各个模块的作用
- 代码组织方式
- 设计模式

继续阅读 `05-project-overview.md`
