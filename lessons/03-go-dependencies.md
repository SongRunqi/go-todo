# Go 依赖管理和模块系统

## 目录
1. [什么是 Go 模块](#什么是-go-模块)
2. [go.mod 文件详解](#gomod-文件详解)
3. [go.sum 文件详解](#gosum-文件详解)
4. [常用命令](#常用命令)
5. [添加和管理依赖](#添加和管理依赖)
6. [版本管理](#版本管理)
7. [替换和排除依赖](#替换和排除依赖)
8. [工作区模式](#工作区模式)
9. [最佳实践](#最佳实践)

---

## 什么是 Go 模块

Go 模块（Go Modules）是 Go 1.11 引入的依赖管理系统，在 Go 1.16 之后成为默认方式。

### 模块 vs 包

- **包（Package）**：代码组织的基本单位，一个目录就是一个包
- **模块（Module）**：一组相关包的集合，由 `go.mod` 文件定义

```
模块 (github.com/SongRunqi/go-todo)
├── 包 (main)
├── 包 (app)
├── 包 (cmd)
└── 包 (internal/ai)
```

### 为什么需要模块系统？

在 Go 1.11 之前，Go 使用 GOPATH 来管理代码，存在很多问题：
- 所有项目必须放在 GOPATH 目录下
- 无法指定依赖的版本
- 无法隔离不同项目的依赖

**Go 模块解决了这些问题：**
- 可以在任何地方创建项目
- 精确控制依赖版本
- 自动下载和管理依赖
- 支持语义化版本控制

---

## go.mod 文件详解

`go.mod` 是模块定义文件，类似于 Node.js 的 `package.json` 或 Python 的 `requirements.txt`。

### 创建 go.mod

```bash
# 在项目根目录执行
go mod init github.com/yourusername/projectname

# 例如
go mod init github.com/SongRunqi/go-todo
```

这会创建一个 `go.mod` 文件：

```go
module github.com/SongRunqi/go-todo

go 1.24
```

### go.mod 文件结构

```go
// 模块路径（模块的唯一标识符）
module github.com/SongRunqi/go-todo

// Go 版本
go 1.24

// 直接依赖
require (
    github.com/spf13/cobra v1.10.1
    github.com/rs/zerolog v1.34.0
)

// 间接依赖（被直接依赖引入的）
require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/mattn/go-colorable v0.1.13 // indirect
)

// 替换依赖（用本地或其他版本替换）
replace github.com/old/package => github.com/new/package v1.2.3
replace github.com/local/package => ../local/package

// 排除某个版本（因为有 bug 或安全问题）
exclude github.com/broken/package v1.0.0
```

### 模块路径命名规则

```go
// ✅ 推荐：使用 GitHub 路径
module github.com/username/projectname

// ✅ 使用其他托管平台
module gitlab.com/username/projectname
module bitbucket.org/username/projectname

// ✅ 使用自己的域名
module example.com/myproject

// ✅ 本地学习项目可以简单命名
module myproject
module todo-app
```

### 依赖的标注

```go
require (
    // 直接依赖（你的代码直接 import 的）
    github.com/spf13/cobra v1.10.1

    // 间接依赖（标记为 indirect）
    github.com/spf13/pflag v1.0.5 // indirect
)
```

**indirect 表示：**
1. 你的代码没有直接导入，是通过其他包引入的
2. 或者该依赖的 go.mod 文件不完整

---

## go.sum 文件详解

`go.sum` 是校验和文件，类似于 npm 的 `package-lock.json`。

### 为什么需要 go.sum？

- **安全性**：确保下载的依赖没有被篡改
- **可重复构建**：保证每次构建使用完全相同的依赖

### go.sum 文件内容

```
github.com/spf13/cobra v1.10.1 h1:e5/vxKd/rZsfSJMUX1agtjeTDf+qv1/JdBF8gg5k9ZM=
github.com/spf13/cobra v1.10.1/go.mod h1:...
```

每行包含：
1. 模块路径和版本
2. 哈希值（`h1:xxx`）
3. Go 模块哈希值（`go.mod h1:xxx`）

### 应该提交 go.sum 吗？

**是的！** go.sum 应该提交到版本控制系统（Git）。

```bash
# .gitignore 中不应该包含 go.sum
# ❌ 不要这样
go.sum

# ✅ go.sum 应该被提交
git add go.mod go.sum
git commit -m "Update dependencies"
```

---

## 常用命令

### 1. 初始化模块

```bash
# 创建新模块
go mod init [模块路径]

# 示例
go mod init github.com/yourusername/myproject
```

### 2. 添加依赖

```bash
# 方式 1：直接在代码中 import，然后运行
go mod tidy

# 方式 2：使用 go get 添加
go get github.com/spf13/cobra

# 添加特定版本
go get github.com/spf13/cobra@v1.10.1

# 添加最新版本
go get github.com/spf13/cobra@latest

# 添加特定分支
go get github.com/spf13/cobra@master
```

### 3. 整理依赖

```bash
# 添加缺失的依赖，移除未使用的依赖
go mod tidy

# 这是最常用的命令！每次修改依赖后都应该运行
```

### 4. 下载依赖

```bash
# 下载 go.mod 中的所有依赖到本地缓存
go mod download

# 下载并验证
go mod verify
```

### 5. 查看依赖

```bash
# 列出所有依赖（包括间接依赖）
go list -m all

# 列出直接依赖
go list -m -f '{{if not .Indirect}}{{.Path}}{{end}}' all

# 查看依赖树
go mod graph

# 查看为什么需要某个依赖
go mod why github.com/spf13/cobra
```

### 6. 升级依赖

```bash
# 升级所有依赖到最新的小版本（patch 和 minor）
go get -u ./...

# 升级到最新的 patch 版本
go get -u=patch ./...

# 升级特定包
go get -u github.com/spf13/cobra

# 升级到特定版本
go get github.com/spf13/cobra@v1.10.1
```

### 7. 清理缓存

```bash
# 清理模块缓存
go clean -modcache

# 查看模块缓存位置
go env GOMODCACHE
```

### 8. Vendor 模式

```bash
# 将依赖复制到 vendor 目录
go mod vendor

# 使用 vendor 目录构建
go build -mod=vendor
```

---

## 添加和管理依赖

### 实际示例：为项目添加依赖

假设我们要添加一个 HTTP 路由库 `gin`：

#### 步骤 1：添加依赖

```bash
go get github.com/gin-gonic/gin
```

这会：
1. 下载 gin 及其依赖
2. 更新 go.mod 文件
3. 更新 go.sum 文件

#### 步骤 2：在代码中使用

```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    r.Run()
}
```

#### 步骤 3：整理依赖

```bash
go mod tidy
```

### 查看 go.mod 变化

```go
module myproject

go 1.24

require (
    github.com/gin-gonic/gin v1.9.1
)

require (
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
    // ... 更多间接依赖
)
```

### 移除未使用的依赖

如果你删除了代码中的 import，运行：

```bash
go mod tidy
```

Go 会自动移除不再需要的依赖。

---

## 版本管理

Go 模块使用**语义化版本控制**（Semantic Versioning）。

### 语义化版本

版本号格式：`v主版本号.次版本号.修订号`

```
v1.2.3
│ │ │
│ │ └─ 修订号（Patch）：bug 修复
│ └─── 次版本号（Minor）：新功能，向后兼容
└───── 主版本号（Major）：破坏性变更，不向后兼容
```

### 版本示例

```go
require (
    github.com/spf13/cobra v1.10.1      // 正式版本
    github.com/example/pkg v0.0.0-20230101120000-abcdef123456  // 伪版本（没有 tag）
    github.com/another/pkg v1.2.3-beta.1  // 预发布版本
    github.com/local/pkg v2.0.0+incompatible  // 不兼容的 v2+
)
```

### 主版本 2+ 的特殊规则

当模块升级到 v2 或更高版本时，模块路径需要包含主版本号：

```go
// v0 和 v1
module github.com/example/mymodule
require github.com/example/mymodule v1.2.3

// v2+（路径需要加上 /v2）
module github.com/example/mymodule/v2
require github.com/example/mymodule/v2 v2.0.0

// v3+
module github.com/example/mymodule/v3
require github.com/example/mymodule/v3 v3.0.0
```

**为什么这样做？**
- 允许同时使用同一模块的不同主版本
- 清晰地表明代码有破坏性变更

### 版本选择

```bash
# 最新版本
go get github.com/example/pkg@latest

# 特定版本
go get github.com/example/pkg@v1.2.3

# 特定分支
go get github.com/example/pkg@master
go get github.com/example/pkg@dev

# 特定提交
go get github.com/example/pkg@abc123

# 降级到特定版本
go get github.com/example/pkg@v1.0.0
```

### 最小版本选择（MVS）

Go 使用"最小版本选择"算法：

```
你的模块依赖：
  A v1.2.0 → B v1.1.0
  C v2.0.0 → B v1.3.0

Go 会选择 B v1.3.0（满足所有要求的最小版本）
```

**不是**选择最新版本，而是选择满足所有约束的最低版本。

---

## 替换和排除依赖

### replace - 替换依赖

#### 用途 1：使用本地版本开发

```go
module myproject

go 1.24

require github.com/example/library v1.0.0

// 用本地路径替换
replace github.com/example/library => ../library
```

```bash
myproject/
├── go.mod
└── ../library/      # 本地开发的库
    └── go.mod
```

#### 用途 2：使用 fork 的版本

```go
// 使用你 fork 的版本替换原版本
replace github.com/original/package => github.com/yourname/package v1.0.1
```

#### 用途 3：修复依赖问题

```go
// 某个依赖有 bug，使用修复后的版本
replace github.com/broken/package v1.0.0 => github.com/broken/package v1.0.1
```

### exclude - 排除特定版本

```go
module myproject

go 1.24

// 排除有严重 bug 的版本
exclude github.com/example/buggy v1.2.0
```

当 Go 尝试使用被排除的版本时，会自动选择下一个可用版本。

### 实际示例

```go
module github.com/SongRunqi/go-todo

go 1.24

require (
    github.com/spf13/cobra v1.10.1
)

// 开发时使用本地版本的 cobra
replace github.com/spf13/cobra => /Users/song/projects/cobra

// 排除有 bug 的版本
exclude github.com/spf13/cobra v1.8.0
```

**注意**：`replace` 和 `exclude` 只在主模块中生效，在依赖中无效。

---

## 工作区模式

Go 1.18 引入了工作区模式，用于同时开发多个相关模块。

### 创建工作区

```bash
# 创建工作区
go work init

# 添加模块到工作区
go work use ./myapp
go work use ./mylib

# 查看工作区配置
cat go.work
```

### go.work 文件

```go
go 1.24

use (
    ./myapp
    ./mylib
)

// 也可以在工作区级别 replace
replace github.com/example/pkg => ../external-pkg
```

### 工作区示例

```
workspace/
├── go.work          # 工作区配置
├── myapp/
│   ├── go.mod
│   └── main.go
└── mylib/
    ├── go.mod
    └── lib.go
```

**myapp/main.go：**
```go
package main

import (
    "fmt"
    "workspace/mylib"  // 使用工作区中的 mylib
)

func main() {
    mylib.Hello()
}
```

**好处：**
- 不需要在 go.mod 中使用 `replace`
- 可以同时修改多个模块
- 提交时不会包含临时的 `replace` 指令

**注意**：`go.work` 不应该提交到 Git！

```bash
# .gitignore
go.work
go.work.sum
```

---

## 最佳实践

### 1. 定期运行 go mod tidy

```bash
# 每次修改依赖后
go mod tidy

# 提交前确保依赖正确
git diff go.mod go.sum
git add go.mod go.sum
git commit -m "Update dependencies"
```

### 2. 固定依赖版本

```go
// ✅ 好的实践：使用具体版本
require github.com/example/pkg v1.2.3

// ❌ 不好的实践：不要使用 latest
// go get github.com/example/pkg@latest 会获取具体版本
```

### 3. 谨慎升级依赖

```bash
# 升级前先检查变更日志
# 在测试环境验证后再升级生产环境

# 只升级 patch 版本（较安全）
go get -u=patch ./...

# 升级所有依赖（需要充分测试）
go get -u ./...
go mod tidy
go test ./...
```

### 4. 使用 go mod vendor 管理私有部署

```bash
# 将依赖复制到 vendor 目录
go mod vendor

# 提交 vendor 目录（可选）
git add vendor/
git commit -m "Vendor dependencies"

# 构建时使用 vendor
go build -mod=vendor
```

**何时使用 vendor：**
- 网络受限的环境
- 需要确保依赖完全可控
- 某些 CI/CD 环境要求

### 5. 私有模块配置

如果你使用私有 Git 仓库：

```bash
# 配置 GOPRIVATE
go env -w GOPRIVATE=github.com/yourcompany/*

# 配置 Git 使用 SSH 而不是 HTTPS
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### 6. 使用 go.work 进行本地开发

```bash
# 不要提交 go.work 到 Git
echo "go.work" >> .gitignore
echo "go.work.sum" >> .gitignore

# 使用 go.work 代替 replace
go work init
go work use ./myapp ./mylib
```

### 7. 定期检查漏洞

```bash
# 使用 govulncheck 检查依赖中的安全漏洞
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

### 8. 依赖版本管理策略

```bash
# 开发分支：可以使用较新的依赖
git checkout develop
go get -u ./...

# 生产分支：保守升级，只修复 bug
git checkout main
go get -u=patch ./...

# 提交前运行测试
go mod tidy
go test ./...
go vet ./...
```

---

## go-todo 项目的依赖分析

让我们看看 go-todo 项目的依赖：

### go.mod 文件

```go
module github.com/SongRunqi/go-todo

go 1.24.5

require (
    github.com/briandowns/spinner v1.23.2
    github.com/fatih/color v1.18.0
    github.com/rs/zerolog v1.34.0
    github.com/spf13/cobra v1.10.1
)

require (
    github.com/inconshreveable/mousetrap v1.1.0 // indirect
    github.com/mattn/go-colorable v0.1.13 // indirect
    github.com/mattn/go-isatty v0.0.20 // indirect
    golang.org/x/sys v0.25.0 // indirect
    golang.org/x/term v0.24.0 // indirect
)
```

### 直接依赖分析

1. **github.com/spf13/cobra v1.10.1**
   - CLI 框架
   - 提供命令行参数解析、子命令支持
   - 最重要的依赖

2. **github.com/rs/zerolog v1.34.0**
   - 结构化日志库
   - 高性能、零分配
   - 用于应用程序日志记录

3. **github.com/fatih/color v1.18.0**
   - 终端彩色输出
   - 用于美化命令行界面

4. **github.com/briandowns/spinner v1.23.2**
   - 终端进度指示器
   - 显示加载动画

### 间接依赖分析

这些是直接依赖自动引入的：

1. **github.com/inconshreveable/mousetrap**
   - Cobra 在 Windows 上需要

2. **github.com/mattn/go-colorable**
   - color 包的依赖
   - Windows 终端颜色支持

3. **github.com/mattn/go-isatty**
   - 检测是否在终端中运行

4. **golang.org/x/sys**
   - 系统底层调用

5. **golang.org/x/term**
   - 终端控制

### 如何添加新依赖

假设要添加 JSON 验证库：

```bash
# 1. 添加依赖
go get github.com/go-playground/validator/v10

# 2. 在代码中使用
# app/validator.go

# 3. 整理依赖
go mod tidy

# 4. 测试
go test ./...

# 5. 提交
git add go.mod go.sum
git commit -m "Add validator dependency"
```

---

## 常见问题

### 1. go.mod 和 go.sum 冲突了怎么办？

```bash
# 重新生成 go.sum
go mod tidy
```

### 2. 依赖下载失败怎么办？

```bash
# 使用代理
go env -w GOPROXY=https://goproxy.cn,direct

# 或者
export GOPROXY=https://goproxy.io,direct

# 常用代理：
# - https://goproxy.cn（中国）
# - https://goproxy.io（全球）
# - https://proxy.golang.org（官方）
```

### 3. 如何查看某个包的所有版本？

```bash
go list -m -versions github.com/spf13/cobra
```

### 4. 如何回退到之前的依赖版本？

```bash
# 查看 git 历史中的 go.mod
git checkout HEAD~1 -- go.mod go.sum
go mod download

# 或者直接指定版本
go get github.com/spf13/cobra@v1.9.0
go mod tidy
```

### 5. 如何完全重新下载依赖？

```bash
# 清理缓存
go clean -modcache

# 重新下载
go mod download
```

---

## 总结

### 关键要点

1. **go.mod** - 定义模块和依赖
   - 模块路径
   - Go 版本
   - 依赖列表

2. **go.sum** - 依赖校验和
   - 确保安全性
   - 应该提交到 Git

3. **常用命令**
   - `go mod init` - 初始化
   - `go mod tidy` - 整理依赖
   - `go get` - 添加/升级依赖
   - `go mod download` - 下载依赖

4. **版本管理**
   - 使用语义化版本
   - v2+ 需要在路径中包含版本号

5. **最佳实践**
   - 定期运行 go mod tidy
   - 提交 go.mod 和 go.sum
   - 谨慎升级依赖
   - 使用 GOPROXY 加速下载

## 下一步

在下一课中，我们将学习：
- Cobra CLI 框架详解
- 如何创建命令行应用
- 命令、标志和参数
- go-todo 如何使用 Cobra

继续阅读 `04-cobra-cli-framework.md`
