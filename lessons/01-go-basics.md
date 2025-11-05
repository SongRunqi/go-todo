# Go 语言基础

## 目录
1. [什么是 Go](#什么是-go)
2. [Go 的特点](#go-的特点)
3. [基本语法](#基本语法)
4. [数据类型](#数据类型)
5. [变量和常量](#变量和常量)
6. [函数](#函数)
7. [控制结构](#控制结构)
8. [指针](#指针)
9. [结构体](#结构体)
10. [接口](#接口)
11. [错误处理](#错误处理)
12. [并发编程](#并发编程)

---

## 什么是 Go

Go（又称 Golang）是 Google 在 2009 年发布的一种开源编程语言。它由 Robert Griesemer、Rob Pike 和 Ken Thompson 设计。

### 为什么创造 Go？

- **简化软件开发**：当时的语言要么太复杂（C++），要么性能不够（Python、Ruby）
- **处理大规模系统**：Google 需要一种能够处理大规模分布式系统的语言
- **快速编译**：传统的 C++ 编译时间太长
- **内置并发支持**：多核处理器越来越普遍，需要更好的并发支持

### Go 适合做什么？

- **云服务和微服务**：Docker、Kubernetes 都是用 Go 写的
- **命令行工具**：编译成单个可执行文件，部署简单
- **Web 服务器和 API**：高性能、内置 HTTP 库
- **网络编程**：优秀的并发模型
- **DevOps 工具**：快速、可靠、易部署

---

## Go 的特点

### 1. 简洁的语法
```go
// 这是一个完整的 Go 程序
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

### 2. 快速编译
- Go 的编译速度非常快
- 可以像脚本语言一样快速迭代
- 编译成单个可执行文件，无需依赖

### 3. 静态类型
```go
var name string = "张三"  // 必须声明类型
age := 25                 // 类型推断
// age = "二十五"          // 错误！类型不匹配
```

### 4. 垃圾回收（GC）
- 自动管理内存，不需要手动 free
- 不用担心内存泄漏（大部分情况）

### 5. 内置并发支持
```go
// goroutine - 轻量级线程
go doSomething()  // 在新的 goroutine 中执行
```

### 6. 丰富的标准库
- HTTP 服务器/客户端
- JSON 处理
- 文件操作
- 加密
- 数据库驱动接口

### 7. 强大的工具链
- `go build` - 编译
- `go run` - 运行
- `go test` - 测试
- `go fmt` - 格式化代码
- `go mod` - 依赖管理

---

## 基本语法

### 包（Package）

每个 Go 文件都属于一个包：

```go
package main  // main 包是程序的入口

// 导入其他包
import "fmt"
import "time"

// 或者批量导入
import (
    "fmt"
    "time"
)
```

**重要概念：**
- `package main` 是特殊的包，表示这是一个可执行程序
- 其他包名通常和文件所在目录名相同
- 大写字母开头的函数、变量是**公开的**（可以被其他包使用）
- 小写字母开头的是**私有的**（只能在本包内使用）

### 注释

```go
// 这是单行注释

/*
这是多行注释
可以跨越多行
*/
```

### 分号

Go 语言**不需要**在每行末尾加分号（编译器会自动添加）：

```go
// 正确
fmt.Println("Hello")
fmt.Println("World")

// 不需要这样
fmt.Println("Hello");
fmt.Println("World");
```

---

## 数据类型

### 基本类型

#### 1. 布尔类型
```go
var isActive bool = true
var isComplete bool = false
```

#### 2. 整数类型
```go
// 有符号整数
var i8 int8 = 127          // -128 到 127
var i16 int16 = 32767      // -32768 到 32767
var i32 int32 = 2147483647
var i64 int64 = 9223372036854775807

// 无符号整数
var ui8 uint8 = 255        // 0 到 255
var ui16 uint16 = 65535
var ui32 uint32 = 4294967295
var ui64 uint64 = 18446744073709551615

// 常用的
var age int = 25           // 根据系统是 int32 或 int64
var count uint = 100       // 根据系统是 uint32 或 uint64
```

#### 3. 浮点数
```go
var price float32 = 19.99
var pi float64 = 3.14159265359  // 更高精度
```

#### 4. 字符串
```go
var name string = "张三"
var message string = `这是一个
多行字符串
使用反引号`

// 字符串是不可变的
s := "hello"
// s[0] = 'H'  // 错误！不能修改字符串
```

#### 5. 字符（rune）
```go
var letter rune = 'A'      // rune 是 int32 的别名
var chinese rune = '中'    // 可以表示 Unicode 字符
```

### 复合类型

#### 1. 数组（固定长度）
```go
// 声明一个包含 5 个整数的数组
var arr [5]int
arr[0] = 1
arr[1] = 2

// 声明并初始化
numbers := [5]int{1, 2, 3, 4, 5}

// 自动计算长度
colors := [...]string{"red", "green", "blue"}
```

#### 2. 切片（Slice - 动态数组）
```go
// 切片是最常用的
var s []int              // 声明切片
s = []int{1, 2, 3}      // 初始化

// 使用 make 创建
s2 := make([]int, 5)     // 长度为 5 的切片
s3 := make([]int, 0, 10) // 长度 0，容量 10

// 追加元素
s = append(s, 4)         // [1, 2, 3, 4]
s = append(s, 5, 6, 7)   // [1, 2, 3, 4, 5, 6, 7]

// 切片操作
s[1:3]    // [2, 3]  - 从索引 1 到 3（不包括 3）
s[:3]     // [1, 2, 3] - 从开始到 3
s[2:]     // [3, 4, 5, 6, 7] - 从 2 到结尾
```

#### 3. 映射（Map - 字典）
```go
// 声明并初始化
ages := map[string]int{
    "张三": 25,
    "李四": 30,
}

// 使用 make 创建
scores := make(map[string]int)
scores["数学"] = 95
scores["英语"] = 88

// 获取值
age := ages["张三"]        // 25

// 检查键是否存在
age, ok := ages["王五"]
if ok {
    fmt.Println("找到了:", age)
} else {
    fmt.Println("不存在")
}

// 删除键
delete(ages, "张三")

// 遍历
for name, age := range ages {
    fmt.Println(name, age)
}
```

---

## 变量和常量

### 变量声明

#### 方式 1：使用 var 关键字
```go
var name string = "张三"
var age int = 25
var price float64 = 19.99
```

#### 方式 2：类型推断
```go
var name = "张三"      // 推断为 string
var age = 25          // 推断为 int
var price = 19.99     // 推断为 float64
```

#### 方式 3：简短声明（最常用）
```go
name := "张三"         // 只能在函数内使用
age := 25
price := 19.99
```

#### 方式 4：批量声明
```go
var (
    name  string = "张三"
    age   int    = 25
    price float64 = 19.99
)
```

### 零值（Zero Value）

Go 中的变量如果没有初始化，会自动赋予零值：

```go
var i int        // 0
var f float64    // 0.0
var b bool       // false
var s string     // "" (空字符串)
var p *int       // nil
var slice []int  // nil
var m map[string]int  // nil
```

### 常量

常量使用 `const` 关键字，**不能**修改：

```go
const Pi = 3.14159
const MaxUsers = 100

// 批量声明
const (
    StatusOK = 200
    StatusNotFound = 404
    StatusError = 500
)

// iota - 自动递增
const (
    Monday = iota     // 0
    Tuesday           // 1
    Wednesday         // 2
    Thursday          // 3
    Friday            // 4
    Saturday          // 5
    Sunday            // 6
)
```

---

## 函数

### 基本函数

```go
// 基本函数定义
func greet(name string) {
    fmt.Println("你好,", name)
}

// 带返回值
func add(a int, b int) int {
    return a + b
}

// 相同类型的参数可以简写
func multiply(a, b int) int {
    return a * b
}

// 多个返回值（Go 的特色）
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")
    }
    return a / b, nil
}

// 使用方式
result, err := divide(10, 2)
if err != nil {
    fmt.Println("错误:", err)
} else {
    fmt.Println("结果:", result)
}
```

### 命名返回值

```go
func calculate(a, b int) (sum int, product int) {
    sum = a + b
    product = a * b
    return  // 自动返回 sum 和 product
}
```

### 可变参数

```go
func sum(numbers ...int) int {
    total := 0
    for _, num := range numbers {
        total += num
    }
    return total
}

// 使用
sum(1, 2, 3)           // 6
sum(1, 2, 3, 4, 5)     // 15
```

### 匿名函数和闭包

```go
// 匿名函数
add := func(a, b int) int {
    return a + b
}
result := add(3, 5)

// 闭包
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

c := counter()
fmt.Println(c())  // 1
fmt.Println(c())  // 2
fmt.Println(c())  // 3
```

### 方法（Method）

方法是带有接收者（receiver）的函数：

```go
type Person struct {
    Name string
    Age  int
}

// 值接收者
func (p Person) Greet() {
    fmt.Printf("你好，我是 %s\n", p.Name)
}

// 指针接收者（可以修改字段）
func (p *Person) HaveBirthday() {
    p.Age++
}

// 使用
p := Person{Name: "张三", Age: 25}
p.Greet()           // 你好，我是 张三
p.HaveBirthday()
fmt.Println(p.Age)  // 26
```

---

## 控制结构

### if 语句

```go
// 基本 if
if age >= 18 {
    fmt.Println("成年人")
}

// if-else
if age >= 18 {
    fmt.Println("成年人")
} else {
    fmt.Println("未成年人")
}

// if-else if-else
if score >= 90 {
    fmt.Println("优秀")
} else if score >= 60 {
    fmt.Println("及格")
} else {
    fmt.Println("不及格")
}

// if 带初始化语句（常用）
if err := doSomething(); err != nil {
    fmt.Println("错误:", err)
}
// err 只在 if 块内有效
```

### for 循环

Go 只有 `for` 循环，没有 `while`：

```go
// 传统 for 循环
for i := 0; i < 10; i++ {
    fmt.Println(i)
}

// 类似 while
i := 0
for i < 10 {
    fmt.Println(i)
    i++
}

// 无限循环
for {
    // 需要使用 break 退出
    if condition {
        break
    }
}

// 遍历数组/切片
numbers := []int{1, 2, 3, 4, 5}
for index, value := range numbers {
    fmt.Printf("索引: %d, 值: %d\n", index, value)
}

// 只要值，不要索引
for _, value := range numbers {
    fmt.Println(value)
}

// 只要索引
for index := range numbers {
    fmt.Println(index)
}

// 遍历 map
ages := map[string]int{"张三": 25, "李四": 30}
for name, age := range ages {
    fmt.Printf("%s: %d岁\n", name, age)
}

// 遍历字符串（按 rune 遍历，不是字节）
for index, char := range "Hello世界" {
    fmt.Printf("%d: %c\n", index, char)
}
```

### switch 语句

```go
// 基本 switch
switch day {
case "Monday":
    fmt.Println("星期一")
case "Tuesday":
    fmt.Println("星期二")
case "Wednesday":
    fmt.Println("星期三")
default:
    fmt.Println("其他")
}

// 多个条件
switch day {
case "Saturday", "Sunday":
    fmt.Println("周末")
default:
    fmt.Println("工作日")
}

// 不带表达式的 switch（相当于 if-else）
switch {
case score >= 90:
    fmt.Println("优秀")
case score >= 60:
    fmt.Println("及格")
default:
    fmt.Println("不及格")
}

// 带初始化语句
switch result := calculate(); result {
case 0:
    fmt.Println("零")
case 1:
    fmt.Println("一")
default:
    fmt.Println("其他")
}
```

### defer 语句

`defer` 会在函数返回前执行（常用于清理资源）：

```go
func readFile() {
    file, err := os.Open("test.txt")
    if err != nil {
        return
    }
    defer file.Close()  // 函数返回前会自动关闭文件

    // 读取文件...
}

// 多个 defer 按逆序执行（栈）
func example() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
    fmt.Println("主函数")
}
// 输出: 主函数 3 2 1
```

---

## 指针

指针存储变量的内存地址。

### 基本概念

```go
// 声明一个整数
x := 10

// 获取地址（使用 &）
p := &x
fmt.Println(p)   // 输出地址，例如: 0xc000012090

// 解引用（使用 *）
fmt.Println(*p)  // 10

// 通过指针修改值
*p = 20
fmt.Println(x)   // 20（原变量被修改了）
```

### 为什么需要指针？

1. **高效传递大型数据**：传递指针比复制整个数据结构快
2. **修改原始数据**：函数可以修改传入的变量

```go
// 不使用指针 - 不会修改原变量
func increment(n int) {
    n++
}

// 使用指针 - 会修改原变量
func incrementPtr(n *int) {
    *n++
}

x := 10
increment(x)
fmt.Println(x)    // 10（没有变化）

incrementPtr(&x)
fmt.Println(x)    // 11（被修改了）
```

### 指针与结构体

```go
type Person struct {
    Name string
    Age  int
}

// 不使用指针 - 修改副本
func birthday(p Person) {
    p.Age++
}

// 使用指针 - 修改原始数据
func birthdayPtr(p *Person) {
    p.Age++  // Go 自动解引用，不需要写 (*p).Age++
}

person := Person{Name: "张三", Age: 25}
birthday(person)
fmt.Println(person.Age)  // 25（没变）

birthdayPtr(&person)
fmt.Println(person.Age)  // 26（变了）
```

---

## 结构体

结构体是一组字段的集合，类似其他语言的"类"。

### 定义和使用

```go
// 定义结构体
type Person struct {
    Name    string
    Age     int
    Email   string
    Address string
}

// 创建实例 - 方式 1
var p1 Person
p1.Name = "张三"
p1.Age = 25

// 创建实例 - 方式 2
p2 := Person{
    Name:  "李四",
    Age:   30,
    Email: "lisi@example.com",
}

// 创建实例 - 方式 3（按顺序，不推荐）
p3 := Person{"王五", 28, "wangwu@example.com", "北京"}

// 使用指针创建
p4 := &Person{Name: "赵六", Age: 35}
```

### 嵌入（Embedding - 类似继承）

```go
type Address struct {
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address // 嵌入 Address，可以直接访问其字段
}

p := Person{
    Name: "张三",
    Age:  25,
    Address: Address{
        City:    "北京",
        Country: "中国",
    },
}

// 可以直接访问嵌入类型的字段
fmt.Println(p.City)     // 北京
fmt.Println(p.Country)  // 中国
```

### 结构体标签（Tags）

标签用于给字段添加元数据，常用于 JSON 序列化：

```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Password string `json:"-"`          // 不序列化
    Email    string `json:"email,omitempty"` // 为空时不序列化
}

user := User{ID: 1, Name: "张三", Email: ""}
data, _ := json.Marshal(user)
fmt.Println(string(data))  // {"id":1,"name":"张三"}
```

---

## 接口

接口定义了一组方法，任何实现了这些方法的类型都自动实现了该接口。

### 基本接口

```go
// 定义接口
type Speaker interface {
    Speak() string
}

// 定义结构体
type Dog struct {
    Name string
}

type Cat struct {
    Name string
}

// Dog 实现 Speaker 接口
func (d Dog) Speak() string {
    return "汪汪!"
}

// Cat 实现 Speaker 接口
func (c Cat) Speak() string {
    return "喵喵!"
}

// 使用接口
func makeSound(s Speaker) {
    fmt.Println(s.Speak())
}

dog := Dog{Name: "旺财"}
cat := Cat{Name: "咪咪"}

makeSound(dog)  // 汪汪!
makeSound(cat)  // 喵喵!
```

### 空接口（interface{}）

空接口可以接受任何类型：

```go
func printAnything(v interface{}) {
    fmt.Println(v)
}

printAnything(42)
printAnything("hello")
printAnything(true)
printAnything([]int{1, 2, 3})
```

### 类型断言

```go
var i interface{} = "hello"

// 类型断言
s := i.(string)
fmt.Println(s)  // hello

// 安全的类型断言
s, ok := i.(string)
if ok {
    fmt.Println("是字符串:", s)
}

n, ok := i.(int)
if !ok {
    fmt.Println("不是整数")
}
```

### 类型 switch

```go
func describe(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Printf("整数: %d\n", v)
    case string:
        fmt.Printf("字符串: %s\n", v)
    case bool:
        fmt.Printf("布尔值: %t\n", v)
    default:
        fmt.Printf("未知类型: %T\n", v)
    }
}
```

### 常用的标准接口

```go
// io.Reader - 读取数据
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer - 写入数据
type Writer interface {
    Write(p []byte) (n int, err error)
}

// fmt.Stringer - 字符串表示
type Stringer interface {
    String() string
}

// 实现 Stringer
type Person struct {
    Name string
    Age  int
}

func (p Person) String() string {
    return fmt.Sprintf("%s (%d岁)", p.Name, p.Age)
}

p := Person{"张三", 25}
fmt.Println(p)  // 张三 (25岁)
```

---

## 错误处理

Go 使用显式的错误处理，而不是异常（try-catch）。

### 基本错误处理

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")
    }
    return a / b, nil
}

// 使用
result, err := divide(10, 0)
if err != nil {
    fmt.Println("错误:", err)
    return
}
fmt.Println("结果:", result)
```

### 创建错误

```go
// 方式 1：使用 errors.New
import "errors"
err := errors.New("出错了")

// 方式 2：使用 fmt.Errorf（支持格式化）
err := fmt.Errorf("无法打开文件 %s", filename)

// 方式 3：自定义错误类型
type MyError struct {
    Code    int
    Message string
}

func (e *MyError) Error() string {
    return fmt.Sprintf("错误 %d: %s", e.Code, e.Message)
}
```

### 错误包装（Go 1.13+）

```go
// 包装错误
if err != nil {
    return fmt.Errorf("读取配置失败: %w", err)
}

// 检查错误
if errors.Is(err, os.ErrNotExist) {
    fmt.Println("文件不存在")
}

// 提取错误
var pathError *os.PathError
if errors.As(err, &pathError) {
    fmt.Println("路径错误:", pathError.Path)
}
```

### panic 和 recover

`panic` 用于严重错误（类似异常），`recover` 用于恢复：

```go
func riskyOperation() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("恢复自 panic:", r)
        }
    }()

    // 触发 panic
    panic("出大事了!")

    fmt.Println("这行不会执行")
}

riskyOperation()
fmt.Println("程序继续运行")
```

**注意**：在正常的错误处理中，应该使用 `error`，而不是 `panic`。`panic` 只用于真正无法恢复的情况。

---

## 并发编程

Go 的并发模型基于 CSP（Communicating Sequential Processes）。

### Goroutine

Goroutine 是轻量级线程：

```go
// 普通函数调用（同步）
doSomething()

// 在新的 goroutine 中执行（异步）
go doSomething()

// 匿名函数
go func() {
    fmt.Println("在 goroutine 中运行")
}()

// 示例：并发执行多个任务
for i := 0; i < 5; i++ {
    go func(n int) {
        fmt.Println("任务", n)
    }(i)
}

// 等待 goroutine 完成（否则主程序可能先退出）
time.Sleep(time.Second)
```

### Channel

Channel 用于 goroutine 之间的通信：

```go
// 创建 channel
ch := make(chan int)

// 发送数据到 channel
go func() {
    ch <- 42  // 发送 42
}()

// 从 channel 接收数据
value := <-ch
fmt.Println(value)  // 42

// 带缓冲的 channel
ch2 := make(chan int, 3)  // 容量为 3
ch2 <- 1
ch2 <- 2
ch2 <- 3
// ch2 <- 4  // 会阻塞，因为已满

// 关闭 channel
close(ch)

// 检查 channel 是否关闭
value, ok := <-ch
if !ok {
    fmt.Println("channel 已关闭")
}

// 遍历 channel
for value := range ch {
    fmt.Println(value)
}
```

### select 语句

`select` 用于等待多个 channel：

```go
ch1 := make(chan string)
ch2 := make(chan string)

go func() {
    time.Sleep(1 * time.Second)
    ch1 <- "来自 ch1"
}()

go func() {
    time.Sleep(2 * time.Second)
    ch2 <- "来自 ch2"
}()

// 等待第一个准备好的 channel
select {
case msg1 := <-ch1:
    fmt.Println(msg1)
case msg2 := <-ch2:
    fmt.Println(msg2)
case <-time.After(3 * time.Second):
    fmt.Println("超时")
}
```

### WaitGroup

用于等待多个 goroutine 完成：

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)  // 增加计数
    go func(n int) {
        defer wg.Done()  // 完成时减少计数
        fmt.Println("任务", n)
        time.Sleep(time.Second)
    }(i)
}

wg.Wait()  // 等待所有 goroutine 完成
fmt.Println("所有任务完成")
```

### Mutex（互斥锁）

用于保护共享数据：

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

// 使用
counter := &SafeCounter{}
var wg sync.WaitGroup

for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter.Increment()
    }()
}

wg.Wait()
fmt.Println(counter.Value())  // 1000
```

---

## 总结

这一课介绍了 Go 语言的基础知识：

1. **Go 的特点**：简洁、快速、并发支持强大
2. **基本语法**：包、导入、注释
3. **数据类型**：基本类型、数组、切片、映射
4. **变量和常量**：声明方式、零值
5. **函数**：基本函数、多返回值、方法
6. **控制结构**：if、for、switch、defer
7. **指针**：地址、解引用、为什么使用指针
8. **结构体**：定义、嵌入、标签
9. **接口**：定义、实现、类型断言
10. **错误处理**：error、panic、recover
11. **并发编程**：goroutine、channel、select、WaitGroup、Mutex

## 下一步

在下一课中，我们将学习：
- Go 项目的标准结构
- 如何组织代码
- 包的可见性和命名规则
- 模块和依赖管理

继续阅读 `02-go-project-structure.md`
