# Go 项目结构和组织

## 目录
1. [Go 项目的基本结构](#go-项目的基本结构)
2. [包的组织](#包的组织)
3. [标准项目布局](#标准项目布局)
4. [命名规则](#命名规则)
5. [可见性规则](#可见性规则)
6. [项目示例](#项目示例)
7. [最佳实践](#最佳实践)

---

## Go 项目的基本结构

### 最简单的 Go 项目

```
myproject/
├── go.mod          # 模块定义文件（类似 package.json）
├── go.sum          # 依赖版本锁定文件（类似 package-lock.json）
└── main.go         # 主程序文件
```

`main.go` 示例：
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 稍微复杂一点的项目

```
myproject/
├── go.mod
├── go.sum
├── main.go         # 程序入口
├── config.go       # 配置相关
├── database.go     # 数据库相关
└── utils.go        # 工具函数
```

**注意**：当项目中有多个 `.go` 文件时，它们必须在同一个包中（如果在同一目录下）。

例如，所有文件都应该以 `package main` 开头：

```go
// main.go
package main

func main() {
    // 可以直接调用同包中的函数
    result := add(1, 2)
}
```

```go
// utils.go
package main

func add(a, b int) int {
    return a + b
}
```

---

## 包的组织

### 什么是包（Package）？

包是 Go 中组织代码的基本单位。每个目录是一个包，包中的所有 `.go` 文件必须声明相同的包名。

### 包名规则

1. **包名通常与目录名相同**
   ```
   myproject/
   └── math/
       ├── add.go      # package math
       └── subtract.go # package math
   ```

2. **包名使用小写**，不要用下划线或驼峰：
   - ✅ `package user`
   - ✅ `package http`
   - ❌ `package User`
   - ❌ `package user_service`
   - ❌ `package userService`

3. **包名要简洁有意义**：
   - ✅ `package json`
   - ✅ `package http`
   - ❌ `package utility`
   - ❌ `package common`

### 创建和使用包

**项目结构：**
```
myproject/
├── go.mod
├── main.go
└── math/
    ├── add.go
    └── multiply.go
```

**math/add.go：**
```go
package math

// Add 函数（大写开头，公开的）
func Add(a, b int) int {
    return a + b
}

// 内部使用的函数（小写开头，私有的）
func validate(n int) bool {
    return n > 0
}
```

**math/multiply.go：**
```go
package math

func Multiply(a, b int) int {
    return a * b
}
```

**main.go：**
```go
package main

import (
    "fmt"
    "myproject/math"  // 导入自己的包
)

func main() {
    result := math.Add(1, 2)
    fmt.Println(result)  // 3

    result = math.Multiply(3, 4)
    fmt.Println(result)  // 12

    // math.validate(5)  // 错误！validate 是私有的
}
```

### 导入包的方式

```go
// 单个导入
import "fmt"

// 多个导入
import (
    "fmt"
    "time"
    "myproject/math"
)

// 给包起别名
import (
    m "myproject/math"
)
// 使用: m.Add(1, 2)

// 导入但不使用（只执行 init 函数）
import _ "github.com/lib/pq"

// 将包的所有公开标识符导入当前包（不推荐）
import . "fmt"
// 使用: Println("hello") 而不是 fmt.Println("hello")
```

---

## 标准项目布局

Go 社区有一个广泛认可的项目结构标准：[golang-standards/project-layout](https://github.com/golang-standards/project-layout)

### 常见目录结构

```
myproject/
├── cmd/                    # 主要应用程序
│   ├── myapp/
│   │   └── main.go        # 应用程序入口
│   └── another-tool/
│       └── main.go
├── internal/               # 私有应用程序和库代码
│   ├── app/
│   ├── pkg/
│   └── config/
├── pkg/                    # 可以被外部使用的库代码
│   ├── utils/
│   └── errors/
├── api/                    # API 定义（OpenAPI、Protocol Buffers）
├── web/                    # Web 应用资源（静态文件、模板）
├── configs/                # 配置文件
├── scripts/                # 构建、安装、分析等脚本
├── build/                  # 打包和 CI 相关
├── deployments/            # 部署配置（Docker、Kubernetes）
├── test/                   # 额外的测试文件
├── docs/                   # 设计和用户文档
├── examples/               # 示例代码
├── third_party/            # 外部工具和代码
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 模块校验和
├── Makefile                # 构建脚本
└── README.md               # 项目说明
```

### 重要目录说明

#### `/cmd` - 应用程序入口

```
cmd/
├── myapp/
│   └── main.go        # myapp 的入口
└── worker/
    └── main.go        # worker 的入口
```

每个应用程序应该有自己的目录。`main.go` 应该很简短，主要逻辑放在 `/internal` 或 `/pkg` 中。

```go
// cmd/myapp/main.go
package main

import (
    "myproject/internal/app"
)

func main() {
    app.Run()  // 主逻辑在 internal/app 中
}
```

#### `/internal` - 私有代码

`/internal` 是特殊的目录，Go 编译器会阻止其他项目导入这个目录下的代码。

```
internal/
├── app/           # 应用程序逻辑
├── database/      # 数据库相关
├── service/       # 业务逻辑
└── config/        # 配置管理
```

**为什么使用 internal？**
- 防止外部项目依赖你的内部实现
- 可以自由重构，不用担心破坏外部依赖

#### `/pkg` - 公共库代码

可以被外部项目安全导入的代码。

```
pkg/
├── utils/         # 工具函数
├── errors/        # 错误定义
└── logger/        # 日志工具
```

**何时使用 pkg？**
- 如果你的代码可以被其他项目使用
- 如果你想明确哪些代码是"公开 API"

**注意**：很多项目不使用 `/pkg`，直接把库代码放在根目录下。两种方式都可以。

#### 其他常见目录

```
/api          # API 规范文件
    ├── openapi.yaml
    └── proto/

/web          # 前端资源
    ├── static/
    ├── templates/
    └── public/

/configs      # 配置文件示例
    ├── config.yaml.example
    └── development.yaml

/scripts      # 构建、安装脚本
    ├── build.sh
    └── install.sh

/test         # 额外的测试应用和测试数据
    ├── integration/
    └── testdata/

/docs         # 设计文档和用户文档
    ├── design.md
    └── api.md

/examples     # 使用示例
    └── simple-example/
        └── main.go
```

---

## 命名规则

### 文件命名

- 使用小写字母
- 使用下划线分隔单词（虽然包名不用下划线，但文件名可以）
- 测试文件以 `_test.go` 结尾

```
user.go           # 用户相关
user_service.go   # 用户服务
user_test.go      # 用户测试
```

### 包命名

```go
// ✅ 好的包名
package user
package http
package json
package time

// ❌ 不好的包名
package utilities
package common
package base
package my_package
```

### 变量命名

#### 局部变量：使用短名称

```go
// ✅ 好的局部变量名
for i := 0; i < 10; i++ {
    // i 的作用域很小，短名称就够了
}

// 在小作用域内
u := getCurrentUser()
s := u.Name

// ❌ 不必要的长名称
for index := 0; index < 10; index++ {
    // ...
}
```

#### 导出的变量：使用清晰的名称

```go
// ✅ 好的导出变量名
package config

var DefaultTimeout = 30 * time.Second
var MaxRetries = 3

// ❌ 不清晰的名称
var Timeout = 30  // 什么的超时？单位是什么？
var Max = 3       // 什么的最大值？
```

### 函数命名

#### 使用驼峰命名法（camelCase 或 PascalCase）

```go
// ✅ 私有函数（小写开头）
func getUserByID(id int) *User {
    // ...
}

// ✅ 公开函数（大写开头）
func GetUserByID(id int) *User {
    // ...
}

// ❌ 使用下划线
func get_user_by_id(id int) *User {
    // ...
}
```

#### Getter 和 Setter

Go 不使用 `Get` 前缀：

```go
type Person struct {
    name string
    age  int
}

// ✅ Getter 不用 Get 前缀
func (p *Person) Name() string {
    return p.name
}

// ✅ Setter 使用 Set 前缀
func (p *Person) SetName(name string) {
    p.name = name
}

// ❌ 不要这样
func (p *Person) GetName() string {
    return p.name
}
```

### 接口命名

单方法接口通常以 `-er` 结尾：

```go
// ✅ 好的接口名
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

type Closer interface {
    Close() error
}

// 组合接口
type ReadCloser interface {
    Reader
    Closer
}

// ✅ 多方法接口使用名词
type UserRepository interface {
    Create(user *User) error
    FindByID(id int) (*User, error)
    Update(user *User) error
    Delete(id int) error
}
```

### 常量命名

```go
// ✅ 使用驼峰命名
const MaxConnections = 100
const DefaultTimeout = 30 * time.Second

// ✅ 一组相关常量
const (
    StatusOK       = 200
    StatusNotFound = 404
    StatusError    = 500
)

// ❌ 不要用全大写+下划线（这是 C 风格）
const MAX_CONNECTIONS = 100
const DEFAULT_TIMEOUT = 30
```

---

## 可见性规则

Go 使用简单的规则控制可见性：**首字母大写**就是公开的。

### 导出（公开）

```go
package user

// 导出的结构体
type User struct {
    ID   int     // 导出的字段
    Name string  // 导出的字段
    age  int     // 未导出的字段（私有）
}

// 导出的函数
func NewUser(name string) *User {
    return &User{Name: name}
}

// 导出的方法
func (u *User) GetAge() int {
    return u.age
}

// 导出的常量
const MaxAge = 150

// 导出的变量
var DefaultUser = &User{Name: "Guest"}
```

### 未导出（私有）

```go
package user

// 未导出的结构体
type credentials struct {
    username string
    password string
}

// 未导出的函数
func validatePassword(password string) bool {
    return len(password) >= 8
}

// 未导出的常量
const minPasswordLength = 8

// 未导出的变量
var defaultPassword = "changeme"
```

### 在其他包中使用

```go
// main.go
package main

import (
    "fmt"
    "myproject/user"
)

func main() {
    // ✅ 可以访问导出的
    u := user.NewUser("张三")
    fmt.Println(u.Name)  // 可以
    fmt.Println(u.ID)    // 可以

    // ❌ 不能访问未导出的
    // fmt.Println(u.age)           // 编译错误
    // user.validatePassword("pwd") // 编译错误
}
```

---

## 项目示例

让我们创建一个简单的博客系统来演示项目结构：

```
blog/
├── cmd/
│   └── blog/
│       └── main.go              # 程序入口
├── internal/
│   ├── app/
│   │   └── app.go               # 应用主逻辑
│   ├── model/
│   │   ├── post.go              # 文章模型
│   │   └── user.go              # 用户模型
│   ├── repository/
│   │   ├── post_repository.go   # 文章数据访问
│   │   └── user_repository.go   # 用户数据访问
│   ├── service/
│   │   ├── post_service.go      # 文章业务逻辑
│   │   └── user_service.go      # 用户业务逻辑
│   └── handler/
│       ├── post_handler.go      # 文章 HTTP 处理
│       └── user_handler.go      # 用户 HTTP 处理
├── pkg/
│   ├── database/
│   │   └── db.go                # 数据库连接
│   └── logger/
│       └── logger.go            # 日志工具
├── configs/
│   └── config.yaml              # 配置文件
├── go.mod
└── go.sum
```

### cmd/blog/main.go

```go
package main

import (
    "log"
    "blog/internal/app"
)

func main() {
    if err := app.Run(); err != nil {
        log.Fatal(err)
    }
}
```

### internal/app/app.go

```go
package app

import (
    "blog/internal/handler"
    "blog/pkg/database"
    "blog/pkg/logger"
    "net/http"
)

func Run() error {
    // 初始化日志
    logger.Init()

    // 初始化数据库
    db, err := database.Connect()
    if err != nil {
        return err
    }
    defer db.Close()

    // 设置路由
    http.HandleFunc("/posts", handler.PostHandler)
    http.HandleFunc("/users", handler.UserHandler)

    // 启动服务器
    logger.Info("服务器启动在 :8080")
    return http.ListenAndServe(":8080", nil)
}
```

### internal/model/post.go

```go
package model

import "time"

type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    AuthorID  int       `json:"author_id"`
    CreatedAt time.Time `json:"created_at"`
}
```

### internal/repository/post_repository.go

```go
package repository

import (
    "blog/internal/model"
    "database/sql"
)

type PostRepository struct {
    db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
    return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *model.Post) error {
    // 数据库操作
    return nil
}

func (r *PostRepository) FindByID(id int) (*model.Post, error) {
    // 数据库操作
    return nil, nil
}
```

### internal/service/post_service.go

```go
package service

import (
    "blog/internal/model"
    "blog/internal/repository"
)

type PostService struct {
    repo *repository.PostRepository
}

func NewPostService(repo *repository.PostRepository) *PostService {
    return &PostService{repo: repo}
}

func (s *PostService) CreatePost(title, content string, authorID int) (*model.Post, error) {
    // 验证逻辑
    if title == "" {
        return nil, fmt.Errorf("标题不能为空")
    }

    // 创建文章
    post := &model.Post{
        Title:     title,
        Content:   content,
        AuthorID:  authorID,
        CreatedAt: time.Now(),
    }

    if err := s.repo.Create(post); err != nil {
        return nil, err
    }

    return post, nil
}
```

### internal/handler/post_handler.go

```go
package handler

import (
    "blog/internal/service"
    "encoding/json"
    "net/http"
)

var postService *service.PostService

func PostHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        handleGetPosts(w, r)
    case "POST":
        handleCreatePost(w, r)
    default:
        http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
    }
}

func handleGetPosts(w http.ResponseWriter, r *http.Request) {
    // 获取文章列表
    posts, err := postService.GetAllPosts()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func handleCreatePost(w http.ResponseWriter, r *http.Request) {
    // 创建文章
    // ...
}
```

### pkg/logger/logger.go

```go
package logger

import (
    "log"
    "os"
)

var (
    infoLogger  *log.Logger
    errorLogger *log.Logger
)

func Init() {
    infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
    errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func Info(message string) {
    infoLogger.Println(message)
}

func Error(message string) {
    errorLogger.Println(message)
}
```

---

## 最佳实践

### 1. 保持包的单一职责

每个包应该有一个明确的目的：

```
✅ 好的组织
/user         # 用户相关的所有内容
/order        # 订单相关的所有内容
/payment      # 支付相关的所有内容

❌ 不好的组织
/models       # 所有模型混在一起
/handlers     # 所有处理器混在一起
/utils        # 所有工具函数混在一起
```

### 2. 避免循环依赖

```
❌ 循环依赖
package user
import "myproject/order"  // user 依赖 order

package order
import "myproject/user"   // order 依赖 user
// 这会导致编译错误！
```

解决方法：
- 提取共同依赖到新包
- 使用接口解耦
- 重新设计包结构

### 3. 使用 internal 保护内部实现

```
myproject/
├── internal/           # 只能被本项目导入
│   └── service/
└── pkg/                # 可以被外部项目导入
    └── client/
```

### 4. 每个包都应该有清晰的文档

```go
// Package user provides user management functionality.
//
// This package handles user authentication, authorization,
// and profile management.
package user

// User represents a user in the system.
type User struct {
    // ID is the unique identifier for the user.
    ID int
    // Name is the user's display name.
    Name string
}
```

### 5. 测试文件和源文件放在同一目录

```
user/
├── user.go           # 源文件
├── user_test.go      # 测试文件
├── service.go
└── service_test.go
```

### 6. 使用 `testdata` 目录存放测试数据

```
user/
├── user.go
├── user_test.go
└── testdata/         # 测试数据（Go 工具会忽略这个目录）
    ├── input.json
    └── expected.json
```

### 7. init 函数要慎用

```go
package config

import "os"

var DBConnection string

// init 函数在包被导入时自动执行
func init() {
    DBConnection = os.Getenv("DB_CONNECTION")
}
```

**注意**：
- init 函数会自动执行，可能让代码难以理解
- 测试时难以控制 init 的执行
- 尽量使用显式的初始化函数

### 8. 一个目录一个包

```
❌ 错误：同一目录下有不同的包
myproject/
└── services/
    ├── user.go      # package user
    └── order.go     # package order
    // 编译错误！

✅ 正确：每个包一个目录
myproject/
└── services/
    ├── user/
    │   └── user.go  # package user
    └── order/
        └── order.go # package order
```

---

## 总结

### 关键要点

1. **包是 Go 组织代码的基本单位**
   - 一个目录 = 一个包
   - 包名通常与目录名相同

2. **标准项目布局**
   - `/cmd` - 应用程序入口
   - `/internal` - 私有代码
   - `/pkg` - 公共库代码

3. **命名规则**
   - 包名：小写，简洁
   - 文件名：小写，下划线分隔
   - 函数/类型：驼峰命名法

4. **可见性规则**
   - 大写开头 = 公开（导出）
   - 小写开头 = 私有（未导出）

5. **最佳实践**
   - 保持包的单一职责
   - 避免循环依赖
   - 使用 internal 保护内部实现
   - 测试文件和源文件放在一起

## 下一步

在下一课中，我们将学习：
- Go 模块系统（go.mod）
- 如何管理依赖
- 如何发布自己的包
- 依赖版本控制

继续阅读 `03-go-dependencies.md`
