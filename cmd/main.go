// =============================================
// cmd/main.go — xit 命令行入口
//
// xit 是一个简化版的 Git 版本控制工具
// 它的核心逻辑是：
//   1. 初始化一个 .xit 仓库（类似 git init）
//   2. 把文件存成 Git 对象（Blob / Tree / Commit）
//   3. 用 SHA1 哈希给每个对象算一个唯一地址
//   4. 按哈希值的前两位分目录存储（.xit/objects/xx/xxxx...）
//
// 用法：
//   xit init              — 初始化仓库
//   xit hash-object <文件> — 把文件存成 blob，打印哈希
//   xit cat-file <哈希>    — 查看某个对象的内容
//   xit ls-tree <哈希>     — 查看 tree 对象里的文件列表
// =============================================

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"xit/object"
	"xit/store"
)

// =============================================
// main 函数：程序入口
//
// 流程：
//   1. 读取命令行参数（os.Args）
//      os.Args[0] = 程序名（"./xit"）
//      os.Args[1] = 子命令（"init" / "hash-object" / "cat-file" / "ls-tree"）
//      os.Args[2] = 子命令的参数（文件名 / 哈希值等）
//   2. 根据不同的子命令，调用不同的函数去处理
//   3. 如果命令不认识，打印帮助信息
// =============================================
func main() {
	// 如果用户没有输入任何子命令（只有程序名本身）
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// os.Args[1] 就是第一个参数，也就是子命令的名字
	command := os.Args[1]

	// 根据不同的命令，执行不同的逻辑
	switch command {
	case "init":
		// xit init
		cmdInit()

	case "hash-object":
		// xit hash-object <文件名>
		// 需要第二个参数：文件名
		if len(os.Args) < 3 {
			fmt.Println("用法: xit hash-object <文件路径>")
			return
		}
		cmdHashObject(os.Args[2])

	case "cat-file":
		// xit cat-file <哈希值>
		if len(os.Args) < 3 {
			fmt.Println("用法: xit cat-file <哈希值>")
			return
		}
		cmdCatFile(os.Args[2])

	case "ls-tree":
		// xit ls-tree <哈希值>
		if len(os.Args) < 3 {
			fmt.Println("用法: xit ls-tree <哈希值>")
			return
		}
		cmdLsTree(os.Args[2])

	default:
		// 不认识这个命令，打印帮助信息
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

// =============================================
// printUsage：打印帮助信息
//
// 当用户输入了错误的命令时，告诉他应该怎么写
// =============================================
func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  xit init              -- 初始化 .xit 仓库")
	fmt.Println("  xit hash-object <文件> -- 把文件存为 blob 对象")
	fmt.Println("  xit cat-file <哈希>    -- 查看对象内容")
	fmt.Println("  xit ls-tree <哈希>     -- 查看 tree 对象的内容")
}

// =============================================
// cmdInit：执行 xit init
//
// 这个命令做了三件事：
//   1. 在当前目录创建 .xit 文件夹
//   2. 在 .xit 里创建 objects 文件夹（存所有 Git 对象）
//   3. 在 .xit 里创建 refs/heads 文件夹（存分支引用）
//
// Git 仓库的结构：
//   .xit/              <- 仓库根目录（隐藏文件夹）
//   +-- objects/       <- 存放所有对象文件（按哈希分目录）
//   |   +-- ab/        <- 哈希前两位做目录名
//   |   |   +-- cd...  <- 哈希后 38 位做文件名
//   |   +-- ...
//   +-- refs/          <- 存放引用（分支、标签）
//       +-- heads/     <- 存放分支（如 main）
//           +-- main   <- main 分支指向的 commit 哈希
// =============================================
func cmdInit() {
	// 调用 store.Init() 来创建目录
	// store.Init 里面会创建 .xit/objects 和 .xit/refs/heads
	err := store.Init()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}
	fmt.Println("已初始化空的 xit 仓库在 .xit/")
}

// =============================================
// cmdHashObject：执行 xit hash-object <文件路径>
//
// 这个命令做了三件事：
//   1. 读取文件内容到内存
//   2. 把文件内容封装成 Blob 对象（加上 "blob 长度\x00" 头部）
//   3. 计算 SHA1 哈希，存入 .xit/objects，并打印哈希值
//
// 为什么叫 hash-object：
//   在 Git 里，任何东西（文件、目录、提交）都被称为"对象"
//   每个对象都有一个唯一的 SHA1 哈希作为编号
//   这个命令就是给一个文件计算哈希并存储
// =============================================
func cmdHashObject(filePath string) {
	// 1. 读取文件内容
	// os.ReadFile 是 Go 标准库提供的函数
	// 它会读取整个文件到内存中，返回 []byte
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败 %s: %v\n", filePath, err)
		return
	}

	// 2. 创建 Blob 对象
	// Blob 是 Git 里最基本的对象类型，代表一个文件
	// Size 是文件大小（字节数）
	// Data 是文件的原始内容
	blob := &object.Blob{
		Size: len(data),
		Data: data,
	}

	// 3. 把 Blob 存入 .xit/objects
	// store.Write 会做三件事：
	//   a) 调用 blob.Serialize() 生成带头部的完整数据
	//   b) 用 object.Hash(blob) 计算 SHA1 哈希
	//   c) 按哈希前两位创建子目录（.xit/objects/ab/）
	//   d) 把数据写入文件（.xit/objects/ab/cd...）
	hash, err := store.Write(blob)
	if err != nil {
		fmt.Printf("写入对象失败: %v\n", err)
		return
	}

	// 4. 打印哈希值
	// 这个哈希值就是该文件的唯一标识
	// 以后可以用这个哈希来查看文件内容（cat-file）
	fmt.Println(hash)
}

// =============================================
// cmdCatFile：执行 xit cat-file <哈希值>
//
// 这个命令做了两件事：
//   1. 根据哈希值从 .xit/objects 中读取对象数据
//   2. 打印对象的原始内容（去掉头部的纯数据部分）
//
// Git 对象的存储格式：
//   blob 5\x00hello         <- 头部（类型 + 大小 + 空字符）
//   ^^^^^^                  <- 数据从这里开始才是真正的文件内容
//   store.Read 返回的是完整数据（包含头部）
//   我们需要跳过头部来获取真实内容
// =============================================
func cmdCatFile(hash string) {
	// 1. 从 .xit/objects 中读取对象
	data, err := store.Read(hash)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash, err)
		return
	}

	// 2. 跳过头部，只打印真实内容
	// 数据格式： "blob 5\x00hello"
	//           头部结束     ^ 这里是 \x00，ASCII 0
	// 找到第一个 \x00 的位置，它后面的就是真实数据
	for i, b := range data {
		if b == 0 { // 找到了分隔符 \x00
			// i+1 是真实数据的起始位置
			// 只打印真实数据，不打印头部
			fmt.Print(string(data[i+1:]))
			return
		}
	}

	// 如果没找到 \x00，说明数据格式有问题
	fmt.Print(string(data))
}

// =============================================
// cmdLsTree：执行 xit ls-tree <哈希值>
//
// Tree 对象的存储格式：
//   100644 hello\x00<20字节哈希>100755 main.go\x00<20字节哈希>
//
// 每个文件条目包含：
//   - 权限（Mode）：100644（普通文件）或 100755（可执行文件）
//   - 文件名（Name）：文件的名字
//   - 哈希（Hash）：文件内容对应的 Blob 哈希（十六进制字符串）
// =============================================
func cmdLsTree(hash string) {
	// 1. 读取 tree 对象的数据
	data, err := store.Read(hash)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash, err)
		return
	}

	// 2. 跳过头部（找到 \x00 分隔符）
	// tree 对象的格式也是 "tree 长度\x00内容"
	var content []byte
	for i, b := range data {
		if b == 0 {
			content = data[i+1:] // 跳过 \x00 本身
			break
		}
	}

	if content == nil {
		fmt.Println("无效的 tree 对象")
		return
	}

	// 3. 解析 tree 内容
	// tree 的内容是一串连续的条目：
	//   "100644 main.go\x00<20字节哈希>100755 script.sh\x00<20字节哈希>"
	reader := bufio.NewReader(strings.NewReader(string(content)))

	for {
		// 3.1 读取模式（mode）和文件名（name）
		// 从当前位置读到 \x00 为止，得到 "100644 hello" 这样的字符串
		nameWithMode, err := reader.ReadString(0)
		if err != nil {
			break
		}

		// 去掉末尾的 \x00
		nameWithMode = strings.TrimRight(nameWithMode, "\x00")

		// 按空格分割，得到权限和文件名
		parts := strings.SplitN(nameWithMode, " ", 2)

		// 3.2 读取 20 字节的二进制哈希
		hashBytes := make([]byte, 20)
		_, err = reader.Read(hashBytes)
		if err != nil {
			break
		}

		mode := ""
		name := ""
		if len(parts) >= 2 {
			mode = parts[0]
			name = parts[1]
		} else {
			mode = parts[0]
			name = ""
		}

		// 3.3 打印条目
		fmt.Printf("%s %s\t%s\n", mode, fmt.Sprintf("%x", hashBytes), name)
	}
}
