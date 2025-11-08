# 任务提醒功能使用说明

## 功能概述

新增的提醒功能允许你为任务设置多个提醒时间，在任务开始前的指定时间通过系统通知提醒你。

## 核心功能

### 1. 设置提醒时间

为任务配置一个或多个提醒时间：

```bash
# 为任务 1 设置提前 1 小时提醒
./todo reminder set 1 1h

# 为任务 1 设置多个提醒：提前 1 天、1 小时和 15 分钟
./todo reminder set 1 1d 1h 15m

# 支持的时间格式
./todo reminder set 1 30m      # 30 分钟
./todo reminder set 1 2h       # 2 小时
./todo reminder set 1 1d       # 1 天
./todo reminder set 1 2d12h    # 2 天 12 小时
```

### 2. 启用/禁用提醒

```bash
# 启用任务的提醒（必须先设置提醒时间）
./todo reminder enable 1

# 禁用任务的提醒（保留配置，可以之后重新启用）
./todo reminder disable 1
```

### 3. 启动提醒守护进程

提醒功能需要一个后台服务持续运行：

```bash
# 启动守护进程（默认每 1 分钟检查一次）
./todo daemon

# 自定义检查间隔
./todo daemon --interval 30s   # 每 30 秒检查一次
./todo daemon --interval 5m    # 每 5 分钟检查一次

# 停止守护进程
# 按 Ctrl+C
```

## 使用示例

### 场景 1：每周课程提醒

假设你有一个每周三、周五 14:00-16:00 的课程：

```bash
# 1. 创建重复任务（使用 AI 或手动创建）
./todo "每周三和周五下午2点到4点上课"

# 2. 为课程设置提醒（假设任务 ID 是 5）
./todo reminder set 5 1d 1h 15m

# 这会在以下时间发送提醒：
# - 提前 1 天（周二/周四 14:00）
# - 提前 1 小时（当天 13:00）
# - 提前 15 分钟（当天 13:45）

# 3. 启动守护进程
./todo daemon
```

### 场景 2：单次事件提醒

```bash
# 1. 创建单次任务
./todo "明天下午3点开会"

# 2. 设置提醒（假设任务 ID 是 10）
./todo reminder set 10 1h 30m

# 3. 查看任务详情（可以看到提醒配置）
./todo get 10

# 4. 启动守护进程
./todo daemon
```

## 数据持久化

所有提醒配置都会保存在任务数据文件中（通常是 `~/.todo/todos.json`），包括：
- 是否启用提醒
- 提醒时间列表
- 已发送的提醒记录（避免重复发送）

## 系统要求

### Linux
需要安装 `notify-send`（通常已预装）：
```bash
sudo apt-get install libnotify-bin  # Debian/Ubuntu
sudo yum install libnotify           # CentOS/RHEL
```

### macOS
使用系统自带的 `osascript`，无需额外安装。

### Windows
使用 PowerShell 的内置通知功能，Windows 10+ 支持。

## 工作原理

1. **定时检查**：守护进程按指定间隔检查所有任务
2. **计算提醒时间**：对于每个启用提醒的任务，计算下一个事件时间和对应的提醒时间
3. **发送通知**：当当前时间达到提醒时间时，发送系统通知
4. **防止重复**：已发送的提醒会被记录，不会重复发送

## 重复任务支持

对于重复任务，提醒功能会：
- 为每个 occurrence（发生实例）独立跟踪提醒状态
- 自动为下一个周期的实例发送提醒
- 跳过已完成或已错过的实例

## 注意事项

1. **守护进程必须运行**：提醒功能需要 `todo daemon` 持续运行
2. **时间准确性**：提醒的准确性取决于检查间隔（默认 1 分钟）
3. **系统通知权限**：首次使用可能需要授予应用通知权限
4. **后台运行**：建议使用进程管理工具（如 systemd、supervisor）让守护进程在后台运行

## 后台运行守护进程

### 使用 nohup（简单方式）
```bash
nohup ./todo daemon > /tmp/todo-daemon.log 2>&1 &
```

### 使用 systemd（推荐，Linux）
创建文件 `~/.config/systemd/user/todo-daemon.service`：
```ini
[Unit]
Description=Todo Reminder Daemon
After=network.target

[Service]
Type=simple
ExecStart=/path/to/todo daemon
Restart=always

[Install]
WantedBy=default.target
```

启动服务：
```bash
systemctl --user enable todo-daemon
systemctl --user start todo-daemon
```

## 故障排查

### 通知未发送？
1. 检查守护进程是否运行
2. 检查提醒是否已启用：`./todo get <id>`
3. 检查系统通知权限
4. 查看日志（如果使用后台运行）

### 收到重复通知？
这不应该发生。如果出现，请检查是否运行了多个守护进程实例。

## 未来计划

- [ ] 支持声音提醒
- [ ] 支持自定义通知消息模板
- [ ] 支持邮件/短信提醒
- [ ] Web 界面查看和管理提醒
- [ ] 提醒历史记录查看
