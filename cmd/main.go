// cmd/main.go — xit 命令行入口
//
// v2 新增：add / commit / log 三个命令，以及 getHeadBranch / getBranchRef / setBranchRef 辅助函数
// v2 改进：cmdInit 现在会自动创建 .xit/HEAD 文件（之前忘了这茬）
//
// 写这个项目的时候一直在想：Git 到底是怎么工作的？
// 边看资料边敲代码，很多东西试了好几种写法才确定下来。
// 注释里会记一些当时的想法和踩过的坑。
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xit/object"
	"xit/store"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "init":
		cmdInit()
	case "hash-object":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit hash-object <文件路径>")
			return
		}
		cmdHashObject(os.Args[2])
	case "cat-file":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit cat-file <哈希值>")
			return
		}
		cmdCatFile(os.Args[2])
	case "ls-tree":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit ls-tree <哈希值>")
			return
		}
		cmdLsTree(os.Args[2])
	// v2 新增的三个命令
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit add <文件路径>")
			return
		}
		cmdAdd(os.Args[2])
	case "commit":
		msg := ""
		for i := 2; i < len(os.Args)-1; i++ {
			if os.Args[i] == "-m" {
				msg = os.Args[i+1]
				break
			}
		}
		if msg == "" {
			fmt.Println("用法: xit commit -m \"提交信息\"")
			return
		}
		cmdCommit(msg)
	case "log":
		cmdLog()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  xit init              -- 初始化 .xit 仓库")
	fmt.Println("  xit hash-object <文件> -- 把文件存为 blob 对象")
	fmt.Println("  xit cat-file <哈希>    -- 查看对象内容")
	fmt.Println("  xit ls-tree <哈希>     -- 查看 tree 对象的内容")
	// v2 新增
	fmt.Println("  xit add <文件>         -- 暂存文件")
	fmt.Println("  xit commit -m \"信息\"   -- 创建提交")
	fmt.Println("  xit log                -- 查看提交历史")
}

// cmdInit 初始化仓库
// v2 改了：加了创建 HEAD 文件的逻辑，不然 commit 不知道当前在哪个分支
func cmdInit() {
	err := store.Init()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

	// 默认分支叫 main，不叫 master 了
	headDir := filepath.Dir(".xit/HEAD")
	os.MkdirAll(headDir, 0755)
	os.WriteFile(".xit/HEAD", []byte("ref: refs/heads/main\n"), 0644)

	fmt.Println("已初始化空的 xit 仓库在 .xit/")
}

func cmdHashObject(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败 %s: %v\n", filePath, err)
		return
	}

	blob := &object.Blob{
		Size: len(data),
		Data: data,
	}

	hash, err := store.Write(blob)
	if err != nil {
		fmt.Printf("写入对象失败: %v\n", err)
		return
	}

	fmt.Println(hash)
}

func cmdCatFile(hash string) {
	data, err := store.Read(hash)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash, err)
		return
	}

	// 找 \x00 分隔符，后面的才是真数据
	for i, b := range data {
		if b == 0 {
			fmt.Print(string(data[i+1:]))
			return
		}
	}

	fmt.Print(string(data))
}

func cmdLsTree(hash string) {
	data, err := store.Read(hash)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash, err)
		return
	}

	var content []byte
	for i, b := range data {
		if b == 0 {
			content = data[i+1:]
			break
		}
	}

	if content == nil {
		fmt.Println("无效的 tree 对象")
		return
	}

	reader := bufio.NewReader(strings.NewReader(string(content)))

	for {
		nameWithMode, err := reader.ReadString(0)
		if err != nil {
			break
		}

		nameWithMode = strings.TrimRight(nameWithMode, "\x00")
		parts := strings.SplitN(nameWithMode, " ", 2)

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

		fmt.Printf("%s %s\t%s\n", mode, fmt.Sprintf("%x", hashBytes), name)
	}
}

// ===== v2 新增：以下都是辅助函数和新命令 =====

// getHeadBranch 读 HEAD 文件得到当前分支名
// 比如 "ref: refs/heads/main" → 返回 "main"
// 文件不存在就默认 main（兼容旧版 xit init 创建的仓库）
func getHeadBranch() string {
	data, err := os.ReadFile(".xit/HEAD")
	if err != nil {
		return "main"
	}
	line := strings.TrimSpace(string(data))
	if strings.HasPrefix(line, "ref: ") {
		parts := strings.Split(line, "/")
		return parts[len(parts)-1]
	}
	return "main"
}

// getBranchRef 读分支文件，拿到最新的 commit 哈希
// 还没提交过就返回空字符串
func getBranchRef(branch string) string {
	refPath := filepath.Join(".xit", "refs", "heads", branch)
	data, err := os.ReadFile(refPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// setBranchRef 更新分支指向的 commit 哈希
// commit 成功后调用这个
func setBranchRef(branch, hash string) error {
	refPath := filepath.Join(".xit", "refs", "heads", branch)
	dir := filepath.Dir(refPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(refPath, []byte(hash+"\n"), 0644)
}

// cmdAdd 暂存文件：读文件 → 存成 blob → 写入 index
func cmdAdd(filePath string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败 %s: %v\n", filePath, err)
		return
	}

	blob := &object.Blob{
		Size: len(data),
		Data: data,
	}
	hash, err := store.Write(blob)
	if err != nil {
		fmt.Printf("存储对象失败: %v\n", err)
		return
	}

	// 100644 是普通文件，可执行文件应该是 100755，回头再处理
	err = store.AddToIndex("100644", hash, filePath)
	if err != nil {
		fmt.Printf("添加到暂存区失败: %v\n", err)
		return
	}

	fmt.Printf("已暂存 %s（%s）\n", filePath, hash)
}

// cmdCommit 提交：读 index → 建 tree → 建 commit → 更新分支 → 清空 index
//
// 当时写这个函数的时候有个纠结：commit 完了要不要清空 index？
// 想了想还是要清，因为 index 里的文件已经被"消费"了，
// 下次 commit 前得重新 add。Git 也是这么干的。
//
// 还有个问题：如果第 2 步（写 tree）成功了，第 4 步（写 commit）失败了，
// 就会留下一个没人引用的 tree 对象。真正的 Git 有 gc 来回收，
// xit 暂时不管了，反正不影响仓库完整性。
func cmdCommit(message string) {
	entries, err := store.ReadIndex()
	if err != nil {
		fmt.Printf("读取暂存区失败: %v\n", err)
		return
	}
	if len(entries) == 0 {
		fmt.Println("没有文件被暂存，请先使用 xit add")
		return
	}

	tree := &object.Tree{}
	for _, e := range entries {
		tree.Entries = append(tree.Entries, object.Entry{
			Mode: e.Mode,
			Name: e.Name,
			Hash: e.Hash,
		})
	}
	treeHash, err := store.Write(tree)
	if err != nil {
		fmt.Printf("写入 Tree 对象失败: %v\n", err)
		return
	}

	branch := getHeadBranch()
	parentHash := getBranchRef(branch)

	commit := &object.Commit{
		TreeHash: treeHash,
		Parent:   parentHash,
		Author:   "xit 用户 <user@xit>",
		Message:  message,
	}
	commitHash, err := store.Write(commit)
	if err != nil {
		fmt.Printf("写入 Commit 对象失败: %v\n", err)
		return
	}

	if err := setBranchRef(branch, commitHash); err != nil {
		fmt.Printf("更新分支引用失败: %v\n", err)
		return
	}

	// 提交完了，清空暂存区
	if err := store.WriteIndex(nil); err != nil {
		fmt.Printf("清空暂存区失败: %v\n", err)
		return
	}

	fmt.Printf("[%s %s] %s\n", branch, commitHash[:8], message)
	fmt.Printf("  %d 个文件已提交\n", len(entries))
}

// cmdLog 从最新 commit 开始沿着 parent 链往回打印
//
// 有个偷懒的地方：日期用了当前时间而不是 commit 里的时间。
// 因为 Commit 结构体当时设计的时候忘了加时间字段……回头补上。
// TODO: 在 Commit 里加 Time 字段
func cmdLog() {
	branch := getHeadBranch()
	hash := getBranchRef(branch)

	if hash == "" {
		fmt.Println("当前分支没有提交记录")
		return
	}

	fmt.Printf("分支: %s\n", branch)
	fmt.Println(strings.Repeat("-", 50))

	for hash != "" {
		data, err := store.Read(hash)
		if err != nil {
			fmt.Printf("读取 commit 失败: %v\n", err)
			break
		}

		var content []byte
		for i, b := range data {
			if b == 0 {
				content = data[i+1:]
				break
			}
		}
		if content == nil {
			break
		}

		lines := strings.Split(string(content), "\n")

		parentHash := ""
		author := ""
		msgStart := 0

		for i, line := range lines {
			if strings.HasPrefix(line, "parent ") {
				parentHash = strings.TrimPrefix(line, "parent ")
			} else if strings.HasPrefix(line, "author ") {
				author = strings.TrimPrefix(line, "author ")
			} else if line == "" {
				msgStart = i + 1
				break
			}
		}

		msgLines := lines[msgStart:]
		message := strings.TrimSpace(strings.Join(msgLines, "\n"))

		fmt.Printf("提交: %s\n", hash)
		if parentHash != "" {
			fmt.Printf("父提交: %s\n", parentHash)
		}
		fmt.Printf("作者: %s\n", author)
		fmt.Printf("日期: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("\n    %s\n\n", message)

		hash = parentHash

		if hash != "" {
			fmt.Println(strings.Repeat("-", 50))
		}
	}
}
