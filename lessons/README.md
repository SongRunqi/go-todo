# Go 语言学习指南 - 基于 go-todo 项目

欢迎来到 Go 语言学习指南！这是一套完整的教程，帮助你从零开始学习 Go 语言，并深入理解 go-todo 项目。

## 📚 课程概览

这套教程共 9 课，从 Go 语言基础到项目最新功能，循序渐进。每一课都包含详细的解释、代码示例和实践建议。

### 课程列表

| 课程 | 主题 | 难度 | 预计时间 |
|------|------|------|----------|
| [01. Go 语言基础](01-go-basics.md) | Go 的基本语法、数据类型、控制结构 | ⭐ 入门 | 2-3 小时 |
| [02. Go 项目结构](02-go-project-structure.md) | 包管理、项目组织、命名规则 | ⭐ 入门 | 1-2 小时 |
| [03. Go 依赖管理](03-go-dependencies.md) | go.mod、go.sum、依赖管理 | ⭐⭐ 初级 | 1-2 小时 |
| [04. Cobra CLI 框架](04-cobra-cli-framework.md) | CLI 开发、命令设计、标志和参数 | ⭐⭐ 初级 | 2-3 小时 |
| [05. go-todo 项目架构](05-project-overview.md) | 项目结构、模块设计、数据流 | ⭐⭐⭐ 中级 | 2-3 小时 |
| [06. 代码详细解析](06-code-walkthrough.md) | 逐行代码分析、设计模式、技巧 | ⭐⭐⭐ 中级 | 3-4 小时 |
| [07. 测试和调试](07-testing-in-go.md) | 单元测试、Mock、基准测试、调试 | ⭐⭐⭐ 中级 | 2-3 小时 |
| [08. 维护和扩展](08-maintenance-guide.md) | 添加功能、问题排查、部署 | ⭐⭐⭐⭐ 高级 | 2-3 小时 |
| [09. 项目新功能](09-new-features.md) | 国际化、重复任务、新命令 | ⭐⭐⭐ 中级 | 2-3 小时 |

**总计：约 17-26 小时**

---

## 🎯 学习目标

完成这套课程后，你将能够：

### 基础能力
- ✅ 理解 Go 语言的基本语法和特性
- ✅ 掌握 Go 项目的标准结构
- ✅ 使用 Go Modules 管理依赖
- ✅ 使用 Cobra 构建 CLI 应用

### 进阶能力
- ✅ 理解 go-todo 项目的架构设计
- ✅ 阅读和修改 Go 代码
- ✅ 编写单元测试和基准测试
- ✅ 使用调试工具排查问题

### 高级能力
- ✅ 为项目添加新功能
- ✅ 优化代码性能
- ✅ 部署和分发 Go 应用
- ✅ 为开源项目做贡献

---

## 🚀 开始学习

### 前置要求

1. **安装 Go**
   ```bash
   # 检查 Go 是否已安装
   go version

   # 如果未安装，请访问: https://golang.org/dl/
   ```

2. **安装 Git**
   ```bash
   git --version
   ```

3. **代码编辑器**
   - 推荐：VS Code + Go 扩展
   - 或者：GoLand、Vim、Sublime Text 等

4. **基础知识**
   - 熟悉命令行操作
   - 了解基本的编程概念（变量、函数、循环等）

### 学习路径

#### 📖 路径 1：完整学习（推荐）

适合 Go 初学者，按顺序学习所有课程：

```
01 → 02 → 03 → 04 → 05 → 06 → 07 → 08
```

**学习建议：**
- 每课学习后动手实践
- 修改代码，观察结果
- 完成课后练习
- 遇到问题先尝试自己解决

#### ⚡ 路径 2：快速入门

已有编程基础，想快速了解项目：

```
01 (快速浏览) → 04 → 05 → 06
```

**学习建议：**
- 重点关注 Go 的特殊之处
- 理解项目架构
- 尝试修改代码

#### 🎯 路径 3：按需学习

有特定目标，选择相关课程：

- **想学 Go 基础？** → 课程 01, 02, 03
- **想学 CLI 开发？** → 课程 04
- **想理解项目？** → 课程 05, 06
- **想写测试？** → 课程 07
- **想添加功能？** → 课程 08

---

## 📝 课程详情

### 第 1 课：Go 语言基础

**你将学到：**
- Go 的历史和特点
- 基本语法和数据类型
- 函数和方法
- 结构体和接口
- 错误处理
- 并发编程（goroutine 和 channel）

**实践项目：**
- 编写"Hello World"
- 创建简单的计算器
- 实现并发任务处理

---

### 第 2 课：Go 项目结构

**你将学到：**
- 包（Package）的概念
- 标准项目布局（cmd、internal、pkg）
- 命名规则和最佳实践
- 可见性规则（大小写）

**实践项目：**
- 创建一个多包项目
- 组织代码结构
- 实现包之间的调用

---

### 第 3 课：Go 依赖管理

**你将学到：**
- go.mod 和 go.sum 文件
- 如何添加和更新依赖
- 版本管理策略
- 私有模块配置
- vendor 模式

**实践项目：**
- 创建一个使用外部库的项目
- 升级依赖版本
- 解决依赖冲突

---

### 第 4 课：Cobra CLI 框架

**你将学到：**
- Cobra 框架基础
- 命令、标志和参数
- 子命令设计
- 生命周期钩子
- Shell 补全

**实践项目：**
- 创建一个简单的 CLI 工具
- 添加子命令和标志
- 实现命令补全

---

### 第 5 课：go-todo 项目架构

**你将学到：**
- go-todo 的目录结构
- 核心模块和职责
- 数据流和处理流程
- 设计模式应用
- AI 集成原理

**实践项目：**
- 画出项目架构图
- 追踪一个命令的执行流程
- 理解每个模块的作用

---

### 第 6 课：代码详细解析

**你将学到：**
- 程序入口和初始化
- CRUD 操作实现
- AI 集成详解
- 存储系统设计
- 实用编程技巧

**实践项目：**
- 阅读关键代码
- 添加日志输出观察执行
- 修改功能逻辑

---

### 第 7 课：测试和调试

**你将学到：**
- 单元测试编写
- 表驱动测试
- Mock 和桩
- 基准测试
- 调试工具使用

**实践项目：**
- 为现有代码编写测试
- 使用 Mock 测试 AI 功能
- 运行基准测试
- 使用 Delve 调试

---

### 第 8 课：维护和扩展

**你将学到：**
- 项目维护任务
- 添加新功能的步骤
- 常见问题排查
- 性能优化技巧
- 部署和分发

**实践项目：**
- 添加一个新命令
- 切换 AI 提供商
- 构建跨平台可执行文件
- 创建 Docker 镜像

---

## 💡 学习建议

### 1. 边学边做

不要只看不做！每学完一个概念，立即动手实践：

```bash
# 创建实验文件
cd ~/go-experiments
mkdir lesson01
cd lesson01
go mod init experiments/lesson01

# 编写代码
vim main.go

# 运行
go run main.go
```

### 2. 修改项目代码

在 go-todo 项目中实验：

```bash
# 创建实验分支
git checkout -b experiment/my-feature

# 修改代码
# ... 编辑 ...

# 测试
go run main.go list

# 如果搞砸了，重置
git checkout main
git branch -D experiment/my-feature
```

### 3. 写笔记

记录你的学习过程：

```markdown
# 我的 Go 学习笔记

## 2025-11-05：第 1 课

### 学到的概念
- goroutine 是轻量级线程
- channel 用于 goroutine 通信

### 遇到的问题
- 不理解为什么要用指针

### 解决方案
- 看了第 6 课的例子后明白了

### 代码片段
...
```

### 4. 提问和讨论

- GitHub Issues：在项目仓库提问
- Go 社区：[Go Forum](https://forum.golangbridge.org/)
- Stack Overflow：搜索和提问
- Reddit：[r/golang](https://www.reddit.com/r/golang/)

### 5. 阅读官方文档

- [Go 官方教程](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go 标准库文档](https://pkg.go.dev/std)

---

## 🎓 练习项目

完成课程后，尝试这些项目巩固知识：

### 初级项目

1. **Todo CLI（不用 AI）**
   - 实现基本的 CRUD 操作
   - 使用 JSON 文件存储
   - 练习 Cobra 和文件操作

2. **简单的 HTTP 服务器**
   - 使用 net/http 包
   - 实现几个 API 端点
   - 练习 JSON 处理

3. **文件处理工具**
   - 批量重命名文件
   - 搜索文件内容
   - 练习文件系统操作

### 中级项目

1. **为 go-todo 添加功能**
   - 添加标签系统
   - 添加搜索功能
   - 添加导出功能（Markdown、CSV）

2. **构建 RESTful API**
   - 使用 gin 或 echo 框架
   - 实现用户认证
   - 连接数据库

3. **爬虫程序**
   - 抓取网页内容
   - 并发下载
   - 数据存储

### 高级项目

1. **微服务**
   - 构建多个服务
   - 服务间通信（gRPC）
   - 服务发现和负载均衡

2. **完整的 Web 应用**
   - 前后端分离
   - WebSocket 实时通信
   - 部署到云平台

3. **为开源项目贡献**
   - 选择一个 Go 开源项目
   - 修复 bug 或添加功能
   - 提交 Pull Request

---

## 📚 推荐资源

### 书籍

1. **《Go 语言圣经》（The Go Programming Language）**
   - 作者：Alan Donovan, Brian Kernighan
   - 适合：系统学习 Go

2. **《Go 语言实战》（Go in Action）**
   - 作者：William Kennedy, Brian Ketelsen, Erik St. Martin
   - 适合：实践导向学习

3. **《Concurrency in Go》**
   - 作者：Katherine Cox-Buday
   - 适合：深入理解并发

### 在线资源

1. **官方资源**
   - [Go 官网](https://go.dev/)
   - [Go Tour](https://go.dev/tour/)
   - [Go Playground](https://play.golang.org/)

2. **教程网站**
   - [Go by Example](https://gobyexample.com/)
   - [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/)
   - [Gophercises](https://gophercises.com/)

3. **视频课程**
   - [freeCodeCamp Go 教程](https://www.youtube.com/watch?v=YS4e4q9oBaU)
   - [Tech With Tim](https://www.youtube.com/watch?v=8uiZC0l4Ajw)

4. **博客和文章**
   - [Go 官方博客](https://go.dev/blog/)
   - [Dave Cheney's Blog](https://dave.cheney.net/)
   - [Ardan Labs Blog](https://www.ardanlabs.com/blog/)

### 工具和插件

1. **VS Code 扩展**
   - Go (官方)
   - Go Test Explorer
   - Go Doc

2. **命令行工具**
   - `golangci-lint` - 代码检查
   - `gopls` - 语言服务器
   - `dlv` - 调试器

---

## 🤝 获取帮助

### 遇到问题？

1. **检查课程内容**
   - 重新阅读相关章节
   - 查看代码示例

2. **搜索答案**
   - Google
   - Stack Overflow
   - GitHub Issues

3. **提问**
   - 在项目仓库开 issue
   - 在 Go 论坛提问
   - 加入 Go 社区

### 提问技巧

好的问题示例：

```markdown
## 问题：无法理解 channel 的工作原理

### 我的理解
我知道 channel 用于 goroutine 通信，但不明白为什么...

### 我尝试的代码
```go
// 代码...
```

### 遇到的错误
```
fatal error: all goroutines are asleep - deadlock!
```

### 我的疑问
1. 为什么会死锁？
2. 应该如何正确使用？
```

---

## 🎉 完成课程后

恭喜完成所有课程！现在你可以：

### 1. 巩固知识
- 重复做练习项目
- 阅读 Go 标准库源码
- 研究优秀的开源项目

### 2. 深入学习
- 学习 Go 的内部实现
- 研究性能优化
- 学习分布式系统设计

### 3. 贡献社区
- 为开源项目做贡献
- 写技术博客
- 帮助其他学习者

### 4. 实际应用
- 在工作中使用 Go
- 开发个人项目
- 参加 Go 相关的活动

---

## 📞 联系和反馈

### 发现错误？

如果你在课程中发现错误或不清楚的地方：

1. 在 GitHub 仓库开 issue
2. 提交 Pull Request 修正
3. 发送反馈邮件

### 课程改进建议

欢迎提供改进建议：

- 哪些地方讲得太简略？
- 哪些地方需要更多例子？
- 希望增加哪些内容？

---

## 📄 版权信息

这些课程材料基于 go-todo 项目创建，旨在帮助学习者理解 Go 语言和项目实践。

- **项目地址**：https://github.com/SongRunqi/go-todo
- **许可证**：MIT License
- **作者**：Claude (AI Assistant)
- **创建日期**：2025-11-05

---

## 🌟 开始学习

准备好了吗？让我们开始吧！

👉 **[第 1 课：Go 语言基础](01-go-basics.md)**

---

*祝你学习愉快！如果觉得这些课程有帮助，请给项目一个 star ⭐*
