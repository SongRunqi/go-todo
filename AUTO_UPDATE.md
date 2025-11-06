# Auto-Update Feature - Technical Documentation

[English](#english) | [中文](#中文)

---

## English

### Overview

Todo-Go implements a serverless auto-update mechanism using GitHub Releases. This approach eliminates the need for a dedicated update server while providing secure, reliable updates for all platforms.

### Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │ ─────>  │    GitHub    │ <─────  │   GitHub    │
│   (todo)    │  Check  │   Releases   │  Upload │   Actions   │
│             │  Update │     API      │  Binary │   (CI/CD)   │
└─────────────┘         └──────────────┘         └─────────────┘
       │                       │
       │  Download Binary      │
       └───────────────────────┘
```

### Components

#### 1. Version Management (`internal/version/version.go`)

**Purpose**: Store and manage version information

**Key Features**:
- Version number (e.g., "1.0.0")
- Git commit hash
- Build date
- Platform information (OS/Architecture)

**Implementation**:
```go
// Version information is injected at build time using -ldflags
var (
    Version   = "dev"           // Set via -X flag
    GitCommit = "unknown"       // Set via -X flag
    BuildDate = "unknown"       // Set via -X flag
)
```

**Build-time Injection**:
```bash
go build -ldflags="-X github.com/SongRunqi/go-todo/internal/version.Version=1.0.0"
```

#### 2. Update Manager (`internal/updater/updater.go`)

**Purpose**: Handle version checking and binary updates

**Key Features**:
- Check for updates via GitHub Releases API
- Download platform-specific binaries
- Verify integrity using SHA256 checksums
- Replace current binary safely with automatic rollback
- Cross-platform support (Linux, macOS, Windows)

**Update Flow**:

```
1. Check for Updates
   └─> GET https://api.github.com/repos/{owner}/{repo}/releases/latest
   └─> Parse JSON response
   └─> Compare version numbers

2. Download Binary
   └─> Select platform-specific asset (e.g., todo-linux-amd64-todo)
   └─> Download from GitHub CDN
   └─> Store in memory

3. Verify Integrity
   └─> Download .sha256 checksum file
   └─> Calculate SHA256 of downloaded binary
   └─> Compare checksums

4. Replace Binary
   └─> Backup current binary (*.backup)
   └─> Write new binary to temp file (*.new)
   └─> Rename temp file to replace current binary
   └─> Delete backup (or restore on failure)
```

**Safety Mechanisms**:
- **Backup**: Original binary is backed up before replacement
- **Atomic Replace**: Uses `os.Rename()` for atomic file operations
- **Auto-Rollback**: Restores backup if update fails
- **Checksum Verification**: Ensures binary integrity

#### 3. CLI Commands

**`todo version`** - Display version information
```bash
$ todo version
Todo-Go Version Information:
  Version:    1.0.0
  Git Commit: a1b2c3d
  Build Date: 2024-01-15T10:30:00Z
  Go Version: go1.23
  Platform:   linux/amd64
```

**`todo upgrade`** - Upgrade to latest version
```bash
$ todo upgrade
✓ New version available: v1.0.0 -> v1.1.0
  Do you want to upgrade? (y/N): y
⟳ Downloading and installing update...
✓ Successfully upgraded to v1.1.0
ℹ Please restart the application to use the new version
```

**`todo upgrade --check`** - Check for updates without installing
```bash
$ todo upgrade --check
ℹ New version available: v1.0.0 -> v1.1.0

  Release Notes: New Features
  - Added auto-update feature
  - Improved performance
```

#### 4. CI/CD Pipeline (`.github/workflows/release.yml`)

**Purpose**: Automate the build and release process

**Trigger**: When a version tag is pushed (e.g., `v1.0.0`)

**Build Matrix**:
- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

**Build Process**:
```yaml
1. Checkout code
2. Set up Go environment
3. Extract version from tag (v1.0.0 → 1.0.0)
4. Get Git commit hash and build date
5. Build binary with version information:
   go build -ldflags="-s -w
     -X version.Version=1.0.0
     -X version.GitCommit=a1b2c3d
     -X version.BuildDate=2024-01-15T10:30:00Z"
6. Generate SHA256 checksum
7. Upload to GitHub Release
```

**Release Assets**:
```
todo-linux-amd64-todo         (Binary for Linux x86_64)
todo-linux-amd64-todo.sha256  (Checksum)
todo-linux-arm64-todo         (Binary for Linux ARM64)
todo-linux-arm64-todo.sha256  (Checksum)
todo-darwin-amd64-todo        (Binary for macOS Intel)
todo-darwin-amd64-todo.sha256 (Checksum)
todo-darwin-arm64-todo        (Binary for macOS Apple Silicon)
todo-darwin-arm64-todo.sha256 (Checksum)
todo-windows-amd64-todo.exe   (Binary for Windows)
todo-windows-amd64-todo.exe.sha256 (Checksum)
```

### Update Process Flow

#### Phase 1: Version Check

```
Client                           GitHub API
  |                                  |
  |----GET /releases/latest-------->|
  |                                  |
  |<-------Release JSON-------------|
  |                                  |
  +--> Parse version                |
  +--> Compare with current         |
  +--> Display result               |
```

#### Phase 2: Download & Verify

```
Client                           GitHub CDN
  |                                  |
  |----Download binary------------->|
  |<-------Binary data--------------|
  |                                  |
  |----Download checksum----------->|
  |<-------SHA256 hash--------------|
  |                                  |
  +--> Calculate SHA256             |
  +--> Verify checksum              |
```

#### Phase 3: Safe Replacement

```
Filesystem Operations
  |
  +--> Read current executable path
  +--> Create backup: todo -> todo.backup
  +--> Write new binary: todo.new
  +--> Atomic rename: todo.new -> todo
  +--> Set executable permissions (0755)
  +--> Delete backup
```

### Security Considerations

#### 1. **HTTPS Only**
- All communication with GitHub uses HTTPS
- Prevents man-in-the-middle attacks

#### 2. **SHA256 Verification**
- Every binary includes a checksum file
- Client verifies integrity before installation
- Prevents corrupted or tampered downloads

#### 3. **Atomic Operations**
- Uses `os.Rename()` for atomic file replacement
- Either succeeds completely or fails safely
- No partial updates

#### 4. **Automatic Rollback**
- Backup created before update
- Restored automatically on failure
- User never left with broken binary

#### 5. **GitHub API Rate Limiting**
- Unauthenticated: 60 requests/hour
- Sufficient for normal update checks
- Can add GitHub token for higher limits

### Platform-Specific Considerations

#### Linux
- Binary location: Typically `~/.local/bin/todo` or `/usr/local/bin/todo`
- Permissions: Must maintain executable bit (0755)
- Symlinks: Automatically resolved using `filepath.EvalSymlinks()`

#### macOS
- Same considerations as Linux
- Apple Silicon (ARM64) requires separate binary
- Intel and ARM binaries are not compatible

#### Windows
- Binary extension: `.exe` required
- File locking: Windows locks running executables
- Update may require elevated permissions for system directories

### Advantages of This Approach

#### ✅ **Serverless**
- No dedicated update server required
- Leverages GitHub's infrastructure
- Zero hosting costs

#### ✅ **Secure**
- HTTPS communication
- SHA256 checksum verification
- Atomic file operations
- Automatic rollback

#### ✅ **Reliable**
- GitHub's 99.9% uptime SLA
- Global CDN distribution
- Fast downloads worldwide

#### ✅ **Multi-Platform**
- Supports Linux, macOS, Windows
- Architecture-aware (amd64, arm64)
- Single update mechanism for all platforms

#### ✅ **Integrated CI/CD**
- Automated builds on tag push
- Consistent versioning
- Release notes generation

#### ✅ **User-Friendly**
- Simple commands: `todo upgrade`
- Interactive confirmation
- Clear progress indicators
- Helpful error messages

### Limitations

#### ⚠️ **GitHub Dependency**
- Requires internet connection
- Subject to GitHub availability
- API rate limits (60/hour unauthenticated)

#### ⚠️ **Binary Size**
- Each release includes multiple binaries
- ~5-10 MB per platform
- Storage accumulates over time

#### ⚠️ **Permission Issues**
- May fail if binary is in protected directory
- User must have write permissions
- Consider suggesting user-local installation

### Troubleshooting

#### Issue: "Permission denied" during update

**Cause**: Binary is in a protected directory (e.g., `/usr/bin`)

**Solution**:
```bash
# Move to user directory
mv /usr/bin/todo ~/.local/bin/todo

# Or run update with sudo (not recommended)
sudo todo upgrade
```

#### Issue: "Checksum verification failed"

**Cause**: Download corrupted or network issue

**Solution**:
```bash
# Retry the update
todo upgrade

# Or download manually from GitHub Releases
```

#### Issue: "Already running the latest version" but outdated

**Cause**: Using development build without version info

**Solution**:
```bash
# Install official release
wget https://github.com/SongRunqi/go-todo/releases/latest/download/todo-linux-amd64-todo
chmod +x todo-linux-amd64-todo
mv todo-linux-amd64-todo ~/.local/bin/todo
```

### Future Enhancements

#### 🔮 **Automatic Background Checks**
- Check for updates on startup
- Configurable check interval
- Silent background updates

#### 🔮 **Delta Updates**
- Download only changed bytes
- Reduce bandwidth usage
- Faster updates for large binaries

#### 🔮 **Update Channels**
- Stable, Beta, Nightly channels
- Allow users to opt into pre-releases
- Separate release tracks

#### 🔮 **Rollback Command**
- `todo rollback` to previous version
- Maintain version history
- Quick recovery from bad updates

#### 🔮 **Update Notifications**
- Desktop notifications
- Email alerts
- RSS/Atom feed

### Release Workflow

#### For Maintainers

**1. Prepare Release**
```bash
# Update CHANGELOG.md
# Bump version in code if needed
# Commit changes
git add .
git commit -m "chore: prepare release v1.1.0"
```

**2. Create Tag**
```bash
# Create annotated tag
git tag -a v1.1.0 -m "Release v1.1.0"

# Push tag to GitHub
git push origin v1.1.0
```

**3. Automated Build**
- GitHub Actions detects tag
- Builds binaries for all platforms
- Generates checksums
- Creates GitHub Release
- Uploads all assets

**4. Verify Release**
```bash
# Check release page
open https://github.com/SongRunqi/go-todo/releases/latest

# Test update
todo upgrade --check
```

### Testing

#### Unit Tests
```bash
# Test updater logic (without actual updates)
go test ./internal/updater/...
```

#### Integration Tests
```bash
# Test with a real (but test) release
GITHUB_REPO=test-repo todo upgrade --check
```

#### Manual Testing
```bash
# Build old version
git checkout v1.0.0
make build
./todo version

# Build new version
git checkout main
make build VERSION=1.1.0

# Test update
./todo upgrade
./todo version
```

---

## 中文

### 概述

Todo-Go 实现了一个基于 GitHub Releases 的无服务器自动更新机制。这种方法无需专用的更新服务器，同时为所有平台提供安全可靠的更新。

### 架构

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   客户端     │ ─────>  │    GitHub    │ <─────  │   GitHub    │
│   (todo)    │  检查   │   Releases   │  上传   │   Actions   │
│             │  更新   │     API      │  二进制 │   (CI/CD)   │
└─────────────┘         └──────────────┘         └─────────────┘
       │                       │
       │  下载二进制文件        │
       └───────────────────────┘
```

### 组件

#### 1. 版本管理 (`internal/version/version.go`)

**目的**：存储和管理版本信息

**主要特性**：
- 版本号（例如 "1.0.0"）
- Git 提交哈希
- 构建日期
- 平台信息（操作系统/架构）

**实现**：
```go
// 版本信息在构建时通过 -ldflags 注入
var (
    Version   = "dev"           // 通过 -X 标志设置
    GitCommit = "unknown"       // 通过 -X 标志设置
    BuildDate = "unknown"       // 通过 -X 标志设置
)
```

**构建时注入**：
```bash
go build -ldflags="-X github.com/SongRunqi/go-todo/internal/version.Version=1.0.0"
```

#### 2. 更新管理器 (`internal/updater/updater.go`)

**目的**：处理版本检查和二进制文件更新

**主要特性**：
- 通过 GitHub Releases API 检查更新
- 下载特定平台的二进制文件
- 使用 SHA256 校验和验证完整性
- 安全替换当前二进制文件并自动回滚
- 跨平台支持（Linux、macOS、Windows）

**更新流程**：

```
1. 检查更新
   └─> GET https://api.github.com/repos/{owner}/{repo}/releases/latest
   └─> 解析 JSON 响应
   └─> 比较版本号

2. 下载二进制文件
   └─> 选择特定平台的资产（例如 todo-linux-amd64-todo）
   └─> 从 GitHub CDN 下载
   └─> 存储在内存中

3. 验证完整性
   └─> 下载 .sha256 校验和文件
   └─> 计算下载的二进制文件的 SHA256
   └─> 比较校验和

4. 替换二进制文件
   └─> 备份当前二进制文件 (*.backup)
   └─> 将新二进制文件写入临时文件 (*.new)
   └─> 重命名临时文件以替换当前二进制文件
   └─> 删除备份（或在失败时恢复）
```

**安全机制**：
- **备份**：替换前备份原始二进制文件
- **原子替换**：使用 `os.Rename()` 进行原子文件操作
- **自动回滚**：更新失败时恢复备份
- **校验和验证**：确保二进制文件完整性

#### 3. CLI 命令

**`todo version`** - 显示版本信息
```bash
$ todo version
Todo-Go 版本信息：
  版本：      1.0.0
  Git 提交：  a1b2c3d
  构建日期：  2024-01-15T10:30:00Z
  Go 版本：   go1.23
  平台：      linux/amd64
```

**`todo upgrade`** - 升级到最新版本
```bash
$ todo upgrade
✓ 发现新版本：v1.0.0 -> v1.1.0
  是否要升级？(y/N): y
⟳ 正在下载并安装更新...
✓ 成功升级到 v1.1.0
ℹ 请重启应用程序以使用新版本
```

**`todo upgrade --check`** - 仅检查更新，不安装
```bash
$ todo upgrade --check
ℹ 发现新版本：v1.0.0 -> v1.1.0

  版本说明：新功能
  - 添加了自动更新功能
  - 改进了性能
```

#### 4. CI/CD 流水线 (`.github/workflows/release.yml`)

**目的**：自动化构建和发布流程

**触发器**：推送版本标签时（例如 `v1.0.0`）

**构建矩阵**：
- **Linux**：amd64、arm64
- **macOS**：amd64（Intel）、arm64（Apple Silicon）
- **Windows**：amd64

**构建过程**：
```yaml
1. 检出代码
2. 设置 Go 环境
3. 从标签提取版本（v1.0.0 → 1.0.0）
4. 获取 Git 提交哈希和构建日期
5. 使用版本信息构建二进制文件：
   go build -ldflags="-s -w
     -X version.Version=1.0.0
     -X version.GitCommit=a1b2c3d
     -X version.BuildDate=2024-01-15T10:30:00Z"
6. 生成 SHA256 校验和
7. 上传到 GitHub Release
```

**发布资产**：
```
todo-linux-amd64-todo         (Linux x86_64 二进制文件)
todo-linux-amd64-todo.sha256  (校验和)
todo-linux-arm64-todo         (Linux ARM64 二进制文件)
todo-linux-arm64-todo.sha256  (校验和)
todo-darwin-amd64-todo        (macOS Intel 二进制文件)
todo-darwin-amd64-todo.sha256 (校验和)
todo-darwin-arm64-todo        (macOS Apple Silicon 二进制文件)
todo-darwin-arm64-todo.sha256 (校验和)
todo-windows-amd64-todo.exe   (Windows 二进制文件)
todo-windows-amd64-todo.exe.sha256 (校验和)
```

### 更新流程

#### 阶段 1：版本检查

```
客户端                        GitHub API
  |                                  |
  |----GET /releases/latest-------->|
  |                                  |
  |<-------Release JSON-------------|
  |                                  |
  +--> 解析版本                      |
  +--> 与当前版本比较                 |
  +--> 显示结果                      |
```

#### 阶段 2：下载和验证

```
客户端                        GitHub CDN
  |                                  |
  |----下载二进制文件--------------->|
  |<-------二进制数据----------------|
  |                                  |
  |----下载校验和------------------->|
  |<-------SHA256 哈希---------------|
  |                                  |
  +--> 计算 SHA256                  |
  +--> 验证校验和                    |
```

#### 阶段 3：安全替换

```
文件系统操作
  |
  +--> 读取当前可执行文件路径
  +--> 创建备份：todo -> todo.backup
  +--> 写入新二进制文件：todo.new
  +--> 原子重命名：todo.new -> todo
  +--> 设置可执行权限 (0755)
  +--> 删除备份
```

### 安全考虑

#### 1. **仅 HTTPS**
- 与 GitHub 的所有通信使用 HTTPS
- 防止中间人攻击

#### 2. **SHA256 验证**
- 每个二进制文件都包含校验和文件
- 客户端在安装前验证完整性
- 防止损坏或被篡改的下载

#### 3. **原子操作**
- 使用 `os.Rename()` 进行原子文件替换
- 要么完全成功，要么安全失败
- 不会出现部分更新

#### 4. **自动回滚**
- 更新前创建备份
- 失败时自动恢复
- 用户永远不会留下损坏的二进制文件

#### 5. **GitHub API 速率限制**
- 未认证：60 次请求/小时
- 足够正常的更新检查
- 可以添加 GitHub token 以获得更高的限制

### 平台特定考虑

#### Linux
- 二进制文件位置：通常是 `~/.local/bin/todo` 或 `/usr/local/bin/todo`
- 权限：必须保持可执行位 (0755)
- 符号链接：使用 `filepath.EvalSymlinks()` 自动解析

#### macOS
- 与 Linux 相同的考虑
- Apple Silicon (ARM64) 需要单独的二进制文件
- Intel 和 ARM 二进制文件不兼容

#### Windows
- 二进制文件扩展名：需要 `.exe`
- 文件锁定：Windows 锁定正在运行的可执行文件
- 更新可能需要系统目录的提升权限

### 此方法的优势

#### ✅ **无服务器**
- 无需专用更新服务器
- 利用 GitHub 的基础设施
- 零托管成本

#### ✅ **安全**
- HTTPS 通信
- SHA256 校验和验证
- 原子文件操作
- 自动回滚

#### ✅ **可靠**
- GitHub 的 99.9% 正常运行时间 SLA
- 全球 CDN 分发
- 全球快速下载

#### ✅ **多平台**
- 支持 Linux、macOS、Windows
- 架构感知（amd64、arm64）
- 所有平台的单一更新机制

#### ✅ **集成 CI/CD**
- 标签推送时自动构建
- 一致的版本控制
- 自动生成发布说明

#### ✅ **用户友好**
- 简单的命令：`todo upgrade`
- 交互式确认
- 清晰的进度指示器
- 有用的错误消息

### 限制

#### ⚠️ **依赖 GitHub**
- 需要互联网连接
- 受 GitHub 可用性影响
- API 速率限制（未认证 60/小时）

#### ⚠️ **二进制文件大小**
- 每个版本包含多个二进制文件
- 每个平台约 5-10 MB
- 存储随时间累积

#### ⚠️ **权限问题**
- 如果二进制文件在受保护的目录中可能失败
- 用户必须具有写权限
- 考虑建议用户本地安装

### 故障排除

#### 问题："更新期间权限被拒绝"

**原因**：二进制文件在受保护的目录中（例如 `/usr/bin`）

**解决方案**：
```bash
# 移动到用户目录
mv /usr/bin/todo ~/.local/bin/todo

# 或使用 sudo 运行更新（不推荐）
sudo todo upgrade
```

#### 问题："校验和验证失败"

**原因**：下载损坏或网络问题

**解决方案**：
```bash
# 重试更新
todo upgrade

# 或从 GitHub Releases 手动下载
```

#### 问题："已经是最新版本"但实际过时

**原因**：使用没有版本信息的开发构建

**解决方案**：
```bash
# 安装官方版本
wget https://github.com/SongRunqi/go-todo/releases/latest/download/todo-linux-amd64-todo
chmod +x todo-linux-amd64-todo
mv todo-linux-amd64-todo ~/.local/bin/todo
```

### 未来增强

#### 🔮 **自动后台检查**
- 启动时检查更新
- 可配置的检查间隔
- 静默后台更新

#### 🔮 **增量更新**
- 仅下载更改的字节
- 减少带宽使用
- 大型二进制文件的更快更新

#### 🔮 **更新通道**
- 稳定版、测试版、夜间版通道
- 允许用户选择预发布版本
- 独立的发布轨道

#### 🔮 **回滚命令**
- `todo rollback` 回到上一版本
- 维护版本历史
- 从错误更新中快速恢复

#### 🔮 **更新通知**
- 桌面通知
- 电子邮件警报
- RSS/Atom 源

### 发布工作流程

#### 对于维护者

**1. 准备发布**
```bash
# 更新 CHANGELOG.md
# 如需要，在代码中更新版本号
# 提交更改
git add .
git commit -m "chore: 准备发布 v1.1.0"
```

**2. 创建标签**
```bash
# 创建带注释的标签
git tag -a v1.1.0 -m "发布 v1.1.0"

# 推送标签到 GitHub
git push origin v1.1.0
```

**3. 自动构建**
- GitHub Actions 检测标签
- 为所有平台构建二进制文件
- 生成校验和
- 创建 GitHub Release
- 上传所有资产

**4. 验证发布**
```bash
# 检查发布页面
open https://github.com/SongRunqi/go-todo/releases/latest

# 测试更新
todo upgrade --check
```

### 测试

#### 单元测试
```bash
# 测试更新器逻辑（不进行实际更新）
go test ./internal/updater/...
```

#### 集成测试
```bash
# 使用真实（但测试）的发布进行测试
GITHUB_REPO=test-repo todo upgrade --check
```

#### 手动测试
```bash
# 构建旧版本
git checkout v1.0.0
make build
./todo version

# 构建新版本
git checkout main
make build VERSION=1.1.0

# 测试更新
./todo upgrade
./todo version
```
