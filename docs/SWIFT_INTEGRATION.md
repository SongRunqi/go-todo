# Swift App 集成方案

本文档介绍如何在 Swift iOS/macOS 应用中集成 go-todo 核心功能。

---

## 方案对比

| 方案 | 优点 | 缺点 | 推荐场景 |
|------|------|------|----------|
| **gomobile (推荐)** | 本地运行、性能最优、离线可用 | 不支持多端同步 | 个人Todo应用 |
| **REST API** | 多端同步、云端存储 | 需要服务器、网络依赖 | 协作型应用 |
| **混合方案** | 兼具两者优势 | 复杂度较高 | 企业级应用 |

---

## 方案一：gomobile 本地集成（推荐）

### 架构图

```
┌─────────────────────────────────────────────┐
│         Swift App (iOS/macOS)               │
│  ┌─────────────────────────────────────┐   │
│  │      SwiftUI Views                   │   │
│  │  - TodoListView                      │   │
│  │  - TodoDetailView                    │   │
│  │  - CreateTodoView                    │   │
│  └────────────┬─────────────────────────┘   │
│               │                              │
│  ┌────────────▼─────────────────────────┐   │
│  │    TodoViewModel (Swift)             │   │
│  │  - ObservableObject                  │   │
│  │  - @Published var todos: [Todo]      │   │
│  └────────────┬─────────────────────────┘   │
│               │ Swift API                    │
│  ┌────────────▼─────────────────────────┐   │
│  │    TodoSDK.framework                 │   │
│  │  (gomobile 生成)                     │   │
│  │  - TodoServiceCreate()               │   │
│  │  - TodoServiceList()                 │   │
│  │  - TodoServiceParse()                │   │
│  └────────────┬─────────────────────────┘   │
└───────────────┼──────────────────────────────┘
                │ Go Code (编译为 Framework)
   ┌────────────▼─────────────────────────┐
   │      Go Business Logic               │
   │  ┌──────────────────────────────┐    │
   │  │  pkg/mobile/                 │    │
   │  │  - service.go (gomobile接口) │    │
   │  └────────┬─────────────────────┘    │
   │           │                           │
   │  ┌────────▼─────────────────────┐    │
   │  │  app/ (现有业务逻辑)          │    │
   │  │  - TodoService               │    │
   │  │  - FileTodoStore             │    │
   │  │  - AI Client                 │    │
   │  └──────────────────────────────┘    │
   └────────────┬──────────────────────────┘
                │
   ┌────────────▼──────────────────────────┐
   │  本地存储 (iOS Documents)             │
   │  ~/Library/Application Support/todo/  │
   │  - todo.json                          │
   │  - todo_back.json                     │
   └───────────────────────────────────────┘
```

---

## 实施步骤

### 步骤 1：安装 gomobile

```bash
# 安装 gomobile
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest

# 初始化 gomobile
gomobile init
```

### 步骤 2：创建 Mobile SDK

创建专门的 mobile 包，封装核心功能：

#### 目录结构

```
go-todo/
├── pkg/
│   └── mobile/              # 新增：gomobile 接口层
│       ├── service.go       # 主要 API
│       ├── types.go         # 数据类型（简化）
│       └── callback.go      # 回调接口
├── app/                     # 现有业务逻辑
├── cmd/
│   ├── cli/                 # CLI 工具
│   └── mobile-build/        # 构建脚本
│       └── build.sh
└── examples/
    └── swift/               # Swift 示例代码
        ├── TodoApp/
        └── TodoAppMac/
```

#### `pkg/mobile/service.go`

```go
// Package mobile provides gomobile-compatible API for go-todo
// This package is designed to be called from Swift/Objective-C
package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"go-todo/app"
	"go-todo/internal/ai"
)

// TodoService 是给 Swift 调用的主要服务
// gomobile 不支持复杂类型，所以使用简化的接口
type TodoService struct {
	service  *app.TodoService
	store    *app.FileTodoStore
	aiClient ai.Client
}

// NewTodoService 创建服务实例
// storagePath: iOS Documents 目录路径
// apiKey: DeepSeek API Key
// language: "zh" 或 "en"
func NewTodoService(storagePath, apiKey, language string) (*TodoService, error) {
	if storagePath == "" {
		return nil, fmt.Errorf("storage path is required")
	}

	// 设置存储路径
	todoPath := filepath.Join(storagePath, "todo.json")
	backupPath := filepath.Join(storagePath, "todo_back.json")

	// 初始化存储
	store := app.NewFileTodoStore(todoPath, backupPath)

	// 初始化 AI 客户端
	var aiClient ai.Client
	if apiKey != "" {
		aiClient = ai.NewDeepSeekClient(
			"https://api.deepseek.com",
			apiKey,
			"deepseek-chat",
		)
	} else {
		aiClient = ai.NewMockClient() // 离线模式
	}

	// 创建业务服务
	config := &app.Config{
		TodoPath:   todoPath,
		BackupPath: backupPath,
		Language:   language,
	}
	service := app.NewTodoService(store, aiClient, config)

	return &TodoService{
		service:  service,
		store:    store,
		aiClient: aiClient,
	}, nil
}

// ListTodos 列出所有任务
// 返回 JSON 字符串（gomobile 限制）
func (s *TodoService) ListTodos(status, urgent string) (string, error) {
	todos, err := s.service.List(status, urgent)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(todos)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetTodo 获取单个任务
func (s *TodoService) GetTodo(id int64) (string, error) {
	todo, err := s.service.GetByID(int(id))
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(todo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateTodo 创建任务
// taskJSON: JSON 格式的任务数据
func (s *TodoService) CreateTodo(taskJSON string) (string, error) {
	var req struct {
		TaskName          string  `json:"taskName"`
		TaskDesc          string  `json:"taskDesc"`
		DueDate           *string `json:"dueDate"` // ISO8601 格式
		Urgent            string  `json:"urgent"`
		IsRecurring       bool    `json:"isRecurring"`
		RecurringType     string  `json:"recurringType"`
		RecurringInterval int     `json:"recurringInterval"`
	}

	if err := json.Unmarshal([]byte(taskJSON), &req); err != nil {
		return "", err
	}

	// 转换为内部请求类型
	createReq := &app.CreateRequest{
		TaskName:          req.TaskName,
		TaskDesc:          req.TaskDesc,
		Urgent:            req.Urgent,
		IsRecurring:       req.IsRecurring,
		RecurringType:     req.RecurringType,
		RecurringInterval: req.RecurringInterval,
	}

	// 解析日期
	if req.DueDate != nil && *req.DueDate != "" {
		t, err := time.Parse(time.RFC3339, *req.DueDate)
		if err == nil {
			createReq.DueDate = &t
		}
	}

	todo, err := s.service.Create(createReq)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(todo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UpdateTodo 更新任务
func (s *TodoService) UpdateTodo(id int64, updateJSON string) (string, error) {
	var req app.UpdateRequest
	if err := json.Unmarshal([]byte(updateJSON), &req); err != nil {
		return "", err
	}

	todo, err := s.service.Update(int(id), &req)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(todo)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeleteTodo 删除任务
func (s *TodoService) DeleteTodo(id int64) error {
	return s.service.Delete(int(id))
}

// CompleteTodo 完成任务
func (s *TodoService) CompleteTodo(id int64) error {
	return s.service.Complete(int(id))
}

// ParseNaturalLanguage AI 解析自然语言
// 返回 JSON: {intent: string, tasks: [], message: string}
func (s *TodoService) ParseNaturalLanguage(input, language string) (string, error) {
	result, err := s.service.ParseAndExecute(input, language)
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetBackupTodos 获取已完成任务
func (s *TodoService) GetBackupTodos() (string, error) {
	todos := s.store.Load(true) // true = 加载备份

	data, err := json.Marshal(todos)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RestoreTodo 恢复已完成任务
func (s *TodoService) RestoreTodo(id int64) error {
	todos := s.store.Load(false)
	backupTodos := s.store.Load(true)

	return app.RestoreTask(&todos, &backupTodos, int(id), s.store)
}

// ExportToJSON 导出所有数据（用于备份）
func (s *TodoService) ExportToJSON() (string, error) {
	todos := s.store.Load(false)
	backups := s.store.Load(true)

	export := map[string]interface{}{
		"todos":   todos,
		"backups": backups,
		"exportedAt": time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(export)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ImportFromJSON 导入数据（用于恢复）
func (s *TodoService) ImportFromJSON(jsonData string) error {
	var data struct {
		Todos   []app.TodoItem `json:"todos"`
		Backups []app.TodoItem `json:"backups"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}

	// 保存数据
	if err := s.store.Save(data.Todos, false); err != nil {
		return err
	}
	if err := s.store.Save(data.Backups, true); err != nil {
		return err
	}

	return nil
}
```

#### `pkg/mobile/types.go`

```go
package mobile

// TodoFilter 过滤条件
type TodoFilter struct {
	Status string // pending, completed, in_progress
	Urgent string // low, medium, high, urgent
}

// 由于 gomobile 限制，主要使用 JSON 字符串传递复杂数据
// Swift 端会解码为结构化对象
```

### 步骤 3：构建 iOS/macOS Framework

#### 构建脚本 `cmd/mobile-build/build.sh`

```bash
#!/bin/bash
set -e

echo "Building go-todo framework for iOS and macOS..."

# 清理旧文件
rm -rf build/
mkdir -p build/

# 构建 iOS framework (支持 iOS 12.0+)
echo "Building for iOS (arm64, simulator)..."
gomobile bind \
  -target=ios \
  -iosversion=12.0 \
  -o build/TodoSDK.xcframework \
  -ldflags="-s -w" \
  github.com/SongRunqi/go-todo/pkg/mobile

# 构建 macOS framework (Apple Silicon + Intel)
echo "Building for macOS (universal)..."
gomobile bind \
  -target=macos \
  -o build/TodoSDKMac.framework \
  -ldflags="-s -w" \
  github.com/SongRunqi/go-todo/pkg/mobile

echo "✅ Build complete!"
echo "iOS Framework: build/TodoSDK.xcframework"
echo "macOS Framework: build/TodoSDKMac.framework"
echo ""
echo "Next steps:"
echo "1. Drag TodoSDK.xcframework into your Xcode project"
echo "2. Add it to 'Frameworks, Libraries, and Embedded Content'"
echo "3. Import in Swift: import TodoSDK"
```

运行构建：

```bash
cd go-todo
chmod +x cmd/mobile-build/build.sh
./cmd/mobile-build/build.sh
```

### 步骤 4：Swift 集成

#### 4.1 Xcode 项目配置

1. 创建新的 iOS App 项目
2. 将 `TodoSDK.xcframework` 拖入项目
3. 在 **General** → **Frameworks, Libraries, and Embedded Content** 中添加

#### 4.2 Swift 数据模型

```swift
// Models/Todo.swift
import Foundation

struct Todo: Identifiable, Codable {
    let taskId: Int
    var taskName: String
    var taskDesc: String
    var status: TodoStatus
    var urgent: TodoUrgency
    let createTime: Date
    var dueDate: Date?

    // 循环任务
    var isRecurring: Bool
    var recurringType: RecurringType?
    var recurringInterval: Int?

    var id: Int { taskId }
}

enum TodoStatus: String, Codable {
    case pending
    case inProgress = "in_progress"
    case completed
}

enum TodoUrgency: String, Codable {
    case low, medium, high, urgent

    var color: Color {
        switch self {
        case .low: return .gray
        case .medium: return .blue
        case .high: return .orange
        case .urgent: return .red
        }
    }
}

enum RecurringType: String, Codable {
    case daily, weekly, monthly, yearly
}

// 创建请求
struct CreateTodoRequest: Codable {
    let taskName: String
    let taskDesc: String
    let dueDate: String? // ISO8601
    let urgent: String
    let isRecurring: Bool
    let recurringType: String?
    let recurringInterval: Int?
}

// AI 解析结果
struct ParseResult: Codable {
    let intent: String
    let tasks: [Todo]
    let message: String
}
```

#### 4.3 Todo Service (Swift Wrapper)

```swift
// Services/TodoService.swift
import Foundation
import TodoSDK // gomobile 生成的 framework

class TodoService: ObservableObject {
    private var goService: MobileTodoService?

    @Published var todos: [Todo] = []
    @Published var backupTodos: [Todo] = []
    @Published var isLoading = false
    @Published var errorMessage: String?

    init() {
        setupService()
        loadTodos()
    }

    private func setupService() {
        // 获取 Documents 目录
        let documentsPath = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask)[0].path

        // 从 UserDefaults 读取 API Key
        let apiKey = UserDefaults.standard.string(forKey: "deepseek_api_key") ?? ""
        let language = Locale.current.languageCode == "zh" ? "zh" : "en"

        do {
            // 调用 Go 代码创建服务
            self.goService = try MobileNewTodoService(documentsPath, apiKey, language)
        } catch {
            print("Failed to initialize TodoService: \(error)")
            self.errorMessage = "初始化失败: \(error.localizedDescription)"
        }
    }

    // MARK: - CRUD Operations

    func loadTodos() {
        guard let service = goService else { return }

        isLoading = true

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                // 调用 Go 方法获取 JSON
                let jsonString = try service.listTodos("", urgent: "")
                let data = jsonString.data(using: .utf8)!

                // 解码为 Swift 对象
                let decoder = JSONDecoder()
                decoder.dateDecodingStrategy = .iso8601
                let todos = try decoder.decode([Todo].self, from: data)

                DispatchQueue.main.async {
                    self?.todos = todos
                    self?.isLoading = false
                }
            } catch {
                DispatchQueue.main.async {
                    self?.errorMessage = "加载失败: \(error.localizedDescription)"
                    self?.isLoading = false
                }
            }
        }
    }

    func createTodo(name: String, description: String, dueDate: Date?, urgent: TodoUrgency) {
        guard let service = goService else { return }

        let request = CreateTodoRequest(
            taskName: name,
            taskDesc: description,
            dueDate: dueDate?.ISO8601Format(),
            urgent: urgent.rawValue,
            isRecurring: false,
            recurringType: nil,
            recurringInterval: nil
        )

        isLoading = true

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                let encoder = JSONEncoder()
                let jsonData = try encoder.encode(request)
                let jsonString = String(data: jsonData, encoding: .utf8)!

                // 调用 Go 方法
                _ = try service.createTodo(jsonString)

                // 重新加载列表
                DispatchQueue.main.async {
                    self?.loadTodos()
                }
            } catch {
                DispatchQueue.main.async {
                    self?.errorMessage = "创建失败: \(error.localizedDescription)"
                    self?.isLoading = false
                }
            }
        }
    }

    func completeTodo(_ todo: Todo) {
        guard let service = goService else { return }

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                try service.completeTodo(Int64(todo.id))

                DispatchQueue.main.async {
                    self?.loadTodos()
                }
            } catch {
                DispatchQueue.main.async {
                    self?.errorMessage = "完成失败: \(error.localizedDescription)"
                }
            }
        }
    }

    func deleteTodo(_ todo: Todo) {
        guard let service = goService else { return }

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                try service.deleteTodo(Int64(todo.id))

                DispatchQueue.main.async {
                    self?.loadTodos()
                }
            } catch {
                DispatchQueue.main.async {
                    self?.errorMessage = "删除失败: \(error.localizedDescription)"
                }
            }
        }
    }

    // MARK: - AI Natural Language

    func parseNaturalLanguage(_ input: String, completion: @escaping (Result<ParseResult, Error>) -> Void) {
        guard let service = goService else {
            completion(.failure(NSError(domain: "TodoService", code: -1)))
            return
        }

        let language = Locale.current.languageCode == "zh" ? "zh" : "en"

        DispatchQueue.global(qos: .userInitiated).async {
            do {
                let jsonString = try service.parseNaturalLanguage(input, language: language)
                let data = jsonString.data(using: .utf8)!

                let decoder = JSONDecoder()
                decoder.dateDecodingStrategy = .iso8601
                let result = try decoder.decode(ParseResult.self, from: data)

                DispatchQueue.main.async {
                    completion(.success(result))
                    // 重新加载以显示新任务
                    self.loadTodos()
                }
            } catch {
                DispatchQueue.main.async {
                    completion(.failure(error))
                }
            }
        }
    }

    // MARK: - Backup

    func loadBackupTodos() {
        guard let service = goService else { return }

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                let jsonString = try service.getBackupTodos()
                let data = jsonString.data(using: .utf8)!

                let decoder = JSONDecoder()
                decoder.dateDecodingStrategy = .iso8601
                let todos = try decoder.decode([Todo].self, from: data)

                DispatchQueue.main.async {
                    self?.backupTodos = todos
                }
            } catch {
                print("Failed to load backup: \(error)")
            }
        }
    }

    func restoreTodo(_ todo: Todo) {
        guard let service = goService else { return }

        DispatchQueue.global(qos: .userInitiated).async { [weak self] in
            do {
                try service.restoreTodo(Int64(todo.id))

                DispatchQueue.main.async {
                    self?.loadTodos()
                    self?.loadBackupTodos()
                }
            } catch {
                DispatchQueue.main.async {
                    self?.errorMessage = "恢复失败: \(error.localizedDescription)"
                }
            }
        }
    }
}

// Date 扩展
extension Date {
    func ISO8601Format() -> String {
        let formatter = ISO8601DateFormatter()
        return formatter.string(from: self)
    }
}
```

#### 4.4 SwiftUI Views

```swift
// Views/TodoListView.swift
import SwiftUI

struct TodoListView: View {
    @StateObject private var todoService = TodoService()
    @State private var showingCreateSheet = false
    @State private var showingNaturalLanguageSheet = false
    @State private var searchText = ""

    var filteredTodos: [Todo] {
        if searchText.isEmpty {
            return todoService.todos
        }
        return todoService.todos.filter {
            $0.taskName.localizedCaseInsensitiveContains(searchText)
        }
    }

    var body: some View {
        NavigationView {
            List {
                ForEach(filteredTodos) { todo in
                    TodoRow(todo: todo, todoService: todoService)
                }
                .onDelete(perform: deleteTodos)
            }
            .navigationTitle("任务")
            .searchable(text: $searchText, prompt: "搜索任务")
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    Button(action: { showingCreateSheet = true }) {
                        Image(systemName: "plus")
                    }
                }
                ToolbarItem(placement: .navigationBarLeading) {
                    Button("AI") {
                        showingNaturalLanguageSheet = true
                    }
                }
            }
            .sheet(isPresented: $showingCreateSheet) {
                CreateTodoView(todoService: todoService)
            }
            .sheet(isPresented: $showingNaturalLanguageSheet) {
                NaturalLanguageView(todoService: todoService)
            }
            .refreshable {
                todoService.loadTodos()
            }
            .overlay {
                if todoService.isLoading {
                    ProgressView()
                }
            }
        }
    }

    private func deleteTodos(at offsets: IndexSet) {
        for index in offsets {
            todoService.deleteTodo(filteredTodos[index])
        }
    }
}

// Views/TodoRow.swift
struct TodoRow: View {
    let todo: Todo
    @ObservedObject var todoService: TodoService

    var body: some View {
        HStack(alignment: .top, spacing: 12) {
            // 完成按钮
            Button(action: {
                todoService.completeTodo(todo)
            }) {
                Image(systemName: todo.status == .completed ? "checkmark.circle.fill" : "circle")
                    .foregroundColor(todo.status == .completed ? .green : .gray)
                    .font(.title2)
            }
            .buttonStyle(PlainButtonStyle())

            VStack(alignment: .leading, spacing: 4) {
                Text(todo.taskName)
                    .font(.headline)
                    .strikethrough(todo.status == .completed)

                if !todo.taskDesc.isEmpty {
                    Text(todo.taskDesc)
                        .font(.subheadline)
                        .foregroundColor(.secondary)
                        .lineLimit(2)
                }

                HStack {
                    // 优先级标签
                    Label(todo.urgent.rawValue.capitalized, systemImage: "exclamationmark.circle")
                        .font(.caption)
                        .foregroundColor(todo.urgent.color)

                    // 截止日期
                    if let dueDate = todo.dueDate {
                        Label(formatDate(dueDate), systemImage: "calendar")
                            .font(.caption)
                            .foregroundColor(.secondary)
                    }

                    // 循环任务标记
                    if todo.isRecurring {
                        Image(systemName: "repeat")
                            .font(.caption)
                            .foregroundColor(.blue)
                    }
                }
            }

            Spacer()
        }
        .padding(.vertical, 4)
    }

    private func formatDate(_ date: Date) -> String {
        let formatter = DateFormatter()
        formatter.dateStyle = .short
        formatter.timeStyle = .short
        return formatter.string(from: date)
    }
}

// Views/CreateTodoView.swift
struct CreateTodoView: View {
    @Environment(\.dismiss) var dismiss
    @ObservedObject var todoService: TodoService

    @State private var taskName = ""
    @State private var taskDesc = ""
    @State private var urgent: TodoUrgency = .medium
    @State private var dueDate = Date()
    @State private var hasDueDate = false

    var body: some View {
        NavigationView {
            Form {
                Section("基本信息") {
                    TextField("任务名称", text: $taskName)
                    TextField("描述（可选）", text: $taskDesc, axis: .vertical)
                        .lineLimit(3...6)
                }

                Section("详情") {
                    Picker("优先级", selection: $urgent) {
                        Text("低").tag(TodoUrgency.low)
                        Text("中").tag(TodoUrgency.medium)
                        Text("高").tag(TodoUrgency.high)
                        Text("紧急").tag(TodoUrgency.urgent)
                    }

                    Toggle("设置截止日期", isOn: $hasDueDate)

                    if hasDueDate {
                        DatePicker("截止日期", selection: $dueDate)
                    }
                }
            }
            .navigationTitle("创建任务")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("取消") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button("创建") {
                        createTodo()
                    }
                    .disabled(taskName.isEmpty)
                }
            }
        }
    }

    private func createTodo() {
        todoService.createTodo(
            name: taskName,
            description: taskDesc,
            dueDate: hasDueDate ? dueDate : nil,
            urgent: urgent
        )
        dismiss()
    }
}

// Views/NaturalLanguageView.swift
struct NaturalLanguageView: View {
    @Environment(\.dismiss) var dismiss
    @ObservedObject var todoService: TodoService

    @State private var inputText = ""
    @State private var isProcessing = false
    @State private var result: ParseResult?

    var body: some View {
        NavigationView {
            VStack(spacing: 20) {
                // 输入区域
                VStack(alignment: .leading, spacing: 8) {
                    Text("用自然语言描述任务")
                        .font(.headline)

                    TextField("例如：明天下午3点开会讨论项目进度", text: $inputText, axis: .vertical)
                        .textFieldStyle(.roundedBorder)
                        .lineLimit(3...6)

                    Text("支持中英文，AI 会自动解析时间、优先级等信息")
                        .font(.caption)
                        .foregroundColor(.secondary)
                }
                .padding()

                // 处理按钮
                Button(action: processInput) {
                    HStack {
                        if isProcessing {
                            ProgressView()
                                .progressViewStyle(CircularProgressViewStyle())
                        }
                        Text(isProcessing ? "处理中..." : "创建任务")
                    }
                    .frame(maxWidth: .infinity)
                    .padding()
                    .background(Color.blue)
                    .foregroundColor(.white)
                    .cornerRadius(10)
                }
                .disabled(inputText.isEmpty || isProcessing)
                .padding(.horizontal)

                // 结果显示
                if let result = result {
                    VStack(alignment: .leading, spacing: 12) {
                        Text("✅ \(result.message)")
                            .font(.subheadline)
                            .foregroundColor(.green)

                        if !result.tasks.isEmpty {
                            Text("已创建 \(result.tasks.count) 个任务")
                                .font(.caption)
                                .foregroundColor(.secondary)
                        }
                    }
                    .padding()
                    .background(Color.green.opacity(0.1))
                    .cornerRadius(8)
                    .padding(.horizontal)
                }

                Spacer()

                // 示例
                VStack(alignment: .leading, spacing: 8) {
                    Text("示例：")
                        .font(.caption)
                        .foregroundColor(.secondary)

                    ForEach(examples, id: \.self) { example in
                        Button(example) {
                            inputText = example
                        }
                        .font(.caption)
                        .foregroundColor(.blue)
                    }
                }
                .padding()
            }
            .navigationTitle("AI 创建")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("关闭") { dismiss() }
                }
            }
        }
    }

    private let examples = [
        "明天下午3点开会",
        "每周一早上9点团队站会",
        "下周五提交项目报告，高优先级"
    ]

    private func processInput() {
        isProcessing = true

        todoService.parseNaturalLanguage(inputText) { result in
            isProcessing = false

            switch result {
            case .success(let parseResult):
                self.result = parseResult

                // 2秒后自动关闭
                DispatchQueue.main.asyncAfter(deadline: .now() + 2) {
                    dismiss()
                }
            case .failure(let error):
                // 显示错误
                print("Parse error: \(error)")
            }
        }
    }
}
```

#### 4.5 App 入口

```swift
// TodoApp.swift
import SwiftUI

@main
struct TodoApp: App {
    var body: some Scene {
        WindowGroup {
            TodoListView()
        }
    }
}
```

### 步骤 5：配置 API Key

```swift
// Views/SettingsView.swift
struct SettingsView: View {
    @AppStorage("deepseek_api_key") private var apiKey = ""
    @State private var showingAlert = false

    var body: some View {
        Form {
            Section("AI 配置") {
                SecureField("DeepSeek API Key", text: $apiKey)

                Link("获取 API Key", destination: URL(string: "https://platform.deepseek.com")!)
                    .font(.caption)
                    .foregroundColor(.blue)
            }

            Section {
                Button("保存") {
                    showingAlert = true
                }
            }
        }
        .navigationTitle("设置")
        .alert("已保存", isPresented: $showingAlert) {
            Button("确定", role: .cancel) { }
        }
    }
}
```

---

## 构建和运行

### 1. 构建 Go Framework

```bash
cd go-todo
./cmd/mobile-build/build.sh
```

### 2. 添加到 Xcode

1. 将 `build/TodoSDK.xcframework` 拖入 Xcode 项目
2. 在项目设置中添加到 **Embed Frameworks**

### 3. 运行应用

```bash
# iOS 模拟器
Cmd + R

# 真机调试
需要 Apple Developer 账号
```

---

## 性能优化

### 1. 减少 Framework 体积

```bash
# 使用 -ldflags 压缩
gomobile bind -ldflags="-s -w" ...

# 预期大小：
# - iOS framework: ~8-12 MB (arm64 + simulator)
# - macOS framework: ~15-20 MB (universal)
```

### 2. 异步操作

所有 Go 调用都应该在后台线程：

```swift
DispatchQueue.global(qos: .userInitiated).async {
    // 调用 Go 代码
    let result = try service.listTodos(...)

    DispatchQueue.main.async {
        // 更新 UI
    }
}
```

### 3. 缓存策略

```swift
class TodoService {
    private var cachedTodos: [Todo] = []
    private var lastRefresh: Date?

    func loadTodos(forceRefresh: Bool = false) {
        let shouldRefresh = forceRefresh ||
            lastRefresh == nil ||
            Date().timeIntervalSince(lastRefresh!) > 60

        if !shouldRefresh {
            self.todos = cachedTodos
            return
        }

        // 从 Go 加载...
    }
}
```

---

## 离线模式

如果不配置 API Key，应用仍可工作（不支持 AI 解析）：

```go
// pkg/mobile/service.go
if apiKey == "" {
    aiClient = ai.NewMockClient() // 使用 Mock 客户端
}
```

Swift 端处理：

```swift
func parseNaturalLanguage(_ input: String) {
    let apiKey = UserDefaults.standard.string(forKey: "deepseek_api_key")

    if apiKey?.isEmpty ?? true {
        // 显示提示：需要配置 API Key
        showAPIKeyAlert()
        return
    }

    // 正常调用
}
```

---

## 优缺点总结

### ✅ 优点

1. **性能优异**：直接内存调用，无网络开销
2. **完全离线**：数据本地存储，无需服务器
3. **一次编写，多端复用**：Go 代码可用于 iOS、macOS、Android
4. **类型安全**：Swift 调用 Go 有编译时检查
5. **应用体积小**：Framework 仅 8-12 MB

### ⚠️ 注意事项

1. **gomobile 限制**：
   - 不支持 Go 泛型
   - 不支持复杂类型（需使用 JSON 传递）
   - 接口必须返回 error

2. **调试困难**：Go 代码崩溃时难以追踪

3. **不支持多端同步**：数据仅存储在本地

---

## 下一步扩展

如果需要多端同步，可以添加：

1. **iCloud 同步**（Swift 端实现）
2. **REST API 可选同步**（混合方案）
3. **CloudKit 集成**

需要我帮你实现哪个部分？
