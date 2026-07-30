// cmd/main.go — xit 命令行入口
//
// v4 加了 branch 和 merge：
//   xit branch               列出分支
//   xit branch <名字>         创建分支
//   xit branch -d <名字>      删除分支
//   xit merge <分支>          合并分支
//
// 顺带把 Commit.Parent 改成了 Parents（支持多个），
。
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

	if command != "init" && command != "help" && command != "--help" {
		if !isRepoInit() {
			fmt.Println("错误：当前目录不是 xit 仓库（或找不到 .xit 目录）")
			fmt.Println("请先执行 xit init")
			return
		}
	}

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
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit add <文件路径...>")
			return
		}
		cmdAdd(os.Args[2:])
	case "commit":
		msgs := []string{}
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "-m" && i+1 < len(os.Args) {
				msgs = append(msgs, os.Args[i+1])
				i++
			}
		}
		if len(msgs) == 0 {
			fmt.Println("用法: xit commit -m \"信息\" [-m \"更多信息\"]")
			return
		}
		cmdCommit(strings.Join(msgs, "\n\n"))
	case "log":
		cmdLog()
	case "status":
		cmdStatus()
	case "checkout":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit checkout <哈希值> -- <文件>")
			fmt.Println("       xit checkout <分支名>")
			return
		}
		cmdCheckout(os.Args[2:])
	case "diff":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit diff <哈希1> <哈希2>")
			return
		}
		cmdDiff(os.Args[2], os.Args[3])
	// v4 新增
	case "branch":
		cmdBranch(os.Args[2:])
	case "merge":
		if len(os.Args) < 3 {
			fmt.Println("用法: xit merge <分支名>")
			return
		}
		cmdMerge(os.Args[2])
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
	fmt.Println("  xit add <文件...>      -- 暂存文件（支持 *.go 通配符）")
	fmt.Println("  xit commit -m \"信息\"   -- 创建提交（多个 -m 自动拼接）")
	fmt.Println("  xit log                -- 查看提交历史")
	fmt.Println("  xit status             -- 查看工作区状态")
	fmt.Println("  xit checkout <哈希> -- <文件>  -- 恢复文件")
	fmt.Println("  xit checkout <分支>    -- 切换分支")
	fmt.Println("  xit diff <哈希1> <哈希2> -- 比较差异")
	// v4
	fmt.Println("  xit branch             -- 列出分支")
	fmt.Println("  xit branch <名字>      -- 创建分支")
	fmt.Println("  xit branch -d <名字>   -- 删除分支")
	fmt.Println("  xit merge <分支>       -- 合并分支到当前分支")
}

// ===== 辅助函数 =====

func isRepoInit() bool {
	_, err := os.Stat(".xit/HEAD")
	return err == nil
}

func getAuthor() string {
	if a := os.Getenv("XIT_AUTHOR"); a != "" {
		return a
	}
	return "xit 用户 <user@xit>"
}

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

func getBranchRef(branch string) string {
	refPath := filepath.Join(".xit", "refs", "heads", branch)
	data, err := os.ReadFile(refPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func setBranchRef(branch, hash string) error {
	refPath := filepath.Join(".xit", "refs", "heads", branch)
	dir := filepath.Dir(refPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(refPath, []byte(hash+"\n"), 0644)
}

func getCommitTreeHash(hash string) string {
	data, err := store.Read(hash)
	if err != nil {
		return ""
	}
	var content []byte
	for i, b := range data {
		if b == 0 {
			content = data[i+1:]
			break
		}
	}
	if content == nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "tree ") {
			return strings.TrimPrefix(line, "tree ")
		}
	}
	return ""
}

func getFilesFromTree(treeHash string) map[string]string {
	result := map[string]string{}
	if treeHash == "" {
		return result
	}
	collectTreeFiles(treeHash, "", result)
	return result
}

func collectTreeFiles(treeHash, prefix string, result map[string]string) {
	data, err := store.Read(treeHash)
	if err != nil {
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
		if len(parts) < 2 {
			break
		}
		mode := parts[0]
		name := parts[1]

		hashBytes := make([]byte, 20)
		_, err = reader.Read(hashBytes)
		if err != nil {
			break
		}
		hash := fmt.Sprintf("%x", hashBytes)
		fullPath := filepath.Join(prefix, name)

		if strings.HasPrefix(mode, "04") {
			collectTreeFiles(hash, fullPath, result)
		} else {
			result[fullPath] = hash
		}
	}
}

func getFileMode(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "100644"
	}
	mode := info.Mode()
	if mode&0100 != 0 {
		return "100755"
	}
	return "100644"
}

// commitInfo 解析 commit 对象的结果
type commitInfo struct {
	treeHash   string
	parents    []string // v4 改成数组
	author     string
	timestamp  int64
	message    string
}

func parseCommit(hash string) *commitInfo {
	data, err := store.Read(hash)
	if err != nil {
		return nil
	}
	var content []byte
	for i, b := range data {
		if b == 0 {
			content = data[i+1:]
			break
		}
	}
	if content == nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	info := &commitInfo{}
	msgStart := 0

	for i, line := range lines {
		if strings.HasPrefix(line, "tree ") {
			info.treeHash = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent ") {
			info.parents = append(info.parents, strings.TrimPrefix(line, "parent "))
		} else if strings.HasPrefix(line, "author ") {
			info.author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "timestamp ") {
			ts := strings.TrimPrefix(line, "timestamp ")
			info.timestamp, _ = strconv.ParseInt(ts, 10, 64)
		} else if line == "" {
			msgStart = i + 1
			break
		}
	}

	msgLines := lines[msgStart:]
	info.message = strings.TrimSpace(strings.Join(msgLines, "\n"))
	return info
}

// isAncestor 检查某个 commit 是不是另一个的祖先
// 用来判断能不能快进合并。从 target 开始沿着 parent 向上找，
// 如果能找到 base，说明 base 是 target 的祖先。
func isAncestor(ancestorHash, targetHash string) bool {
	if ancestorHash == targetHash {
		return true
	}
	info := parseCommit(targetHash)
	if info == nil {
		return false
	}
	for _, p := range info.parents {
		if isAncestor(ancestorHash, p) {
			return true
		}
	}
	return false
}

// ===== 原始命令 =====

func cmdInit() {
	err := store.Init()
	if err != nil {
		fmt.Printf("初始化失败: %v\n", err)
		return
	}

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
	blob := &object.Blob{Size: len(data), Data: data}
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

func cmdAdd(args []string) {
	files := []string{}
	for _, arg := range args {
		if strings.ContainsAny(arg, "*?[") {
			matches, err := filepath.Glob(arg)
			if err != nil {
				fmt.Printf("通配符解析失败 %s: %v\n", arg, err)
				continue
			}
			files = append(files, matches...)
		} else {
			files = append(files, arg)
		}
	}
	if len(files) == 0 {
		fmt.Println("没有匹配的文件")
		return
	}
	for _, filePath := range files {
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("读取文件失败 %s: %v\n", filePath, err)
			continue
		}
		blob := &object.Blob{Size: len(data), Data: data}
		hash, err := store.Write(blob)
		if err != nil {
			fmt.Printf("存储对象失败 %s: %v\n", filePath, err)
			continue
		}
		mode := getFileMode(filePath)
		err = store.AddToIndex(mode, hash, filePath)
		if err != nil {
			fmt.Printf("添加到暂存区失败 %s: %v\n", filePath, err)
			continue
		}
		fmt.Printf("已暂存 %s（%s）\n", filePath, hash)
	}
}

// cmdCommit 提交
// v4 改用 Parents 数组，但普通提交只有一个 parent，逻辑不变
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

	parents := []string{}
	if parentHash != "" {
		parents = append(parents, parentHash)
	}

	commit := &object.Commit{
		TreeHash: treeHash,
		Parents:  parents,
		Author:   getAuthor(),
		Message:  message,
		Time:     time.Now(),
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

	if err := store.WriteIndex(nil); err != nil {
		fmt.Printf("清空暂存区失败: %v\n", err)
		return
	}

	fmt.Printf("[%s %s] %s\n", branch, commitHash[:8], message)
	fmt.Printf("  %d 个文件已提交\n", len(entries))
}

// cmdLog 查看提交历史
// v4 适配了多 parent：合并提交会显示所有 parent，遍历时只走第一个
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
		info := parseCommit(hash)
		if info == nil {
			break
		}

		fmt.Printf("提交: %s\n", hash)
		if len(info.parents) > 0 {
			if len(info.parents) == 1 {
				fmt.Printf("父提交: %s\n", info.parents[0])
			} else {
				// 合并提交有多个 parent
				fmt.Printf("合并: %s\n", strings.Join(info.parents, ", "))
			}
		}
		fmt.Printf("作者: %s\n", info.author)
		t := time.Unix(info.timestamp, 0)
		fmt.Printf("日期: %s\n", t.Format("2006-01-02 15:04:05"))
		fmt.Printf("\n    %s\n\n", info.message)

		// 沿第一个 parent 走（和 git log 默认行为一致）
		if len(info.parents) > 0 {
			hash = info.parents[0]
		} else {
			hash = ""
		}

		if hash != "" {
			fmt.Println(strings.Repeat("-", 50))
		}
	}
}

// cmdStatus 查看状态
func cmdStatus() {
	branch := getHeadBranch()
	commitHash := getBranchRef(branch)

	entries, _ := store.ReadIndex()
	if len(entries) > 0 {
		fmt.Println("要提交的变更：")
		for _, e := range entries {
			fmt.Printf("  新文件: %s\n", e.Name)
		}
		fmt.Println()
	} else {
		fmt.Println("无要提交的变更（使用 xit add 暂存文件）")
		fmt.Println()
	}

	if commitHash != "" {
		treeHash := getCommitTreeHash(commitHash)
		headFiles := getFilesFromTree(treeHash)

		modified := 0
		for path, hash := range headFiles {
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			blob := &object.Blob{Size: len(data), Data: data}
			currentHash := object.Hash(blob)
			if currentHash != hash {
				if modified == 0 {
					fmt.Println("未暂存的变更：")
				}
				fmt.Printf("  修改: %s\n", path)
				modified++
			}
		}
		if modified > 0 {
			fmt.Println()
		}
	}

	untracked := []string{}
	headFiles := map[string]bool{}
	if commitHash != "" {
		treeHash := getCommitTreeHash(commitHash)
		for k := range getFilesFromTree(treeHash) {
			headFiles[k] = true
		}
	}
	stagedFiles := map[string]bool{}
	for _, e := range entries {
		stagedFiles[e.Name] = true
	}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(path, ".xit") || strings.HasPrefix(path, ".git") {
				return filepath.SkipDir
			}
			return nil
		}
		if headFiles[path] || stagedFiles[path] {
			return nil
		}
		if strings.HasSuffix(path, ".exe") || strings.HasPrefix(path, ".") {
			return nil
		}
		untracked = append(untracked, path)
		return nil
	})

	if len(untracked) > 0 {
		fmt.Println("未跟踪的文件：")
		for _, f := range untracked {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}
}

func cmdCheckout(args []string) {
	if len(args) >= 3 && args[1] == "--" {
		hash := args[0]
		filePath := args[2]

		treeHash := getCommitTreeHash(hash)
		if treeHash == "" {
			fmt.Printf("找不到 commit: %s\n", hash)
			return
		}

		files := getFilesFromTree(treeHash)
		blobHash, ok := files[filePath]
		if !ok {
			fmt.Printf("在 commit %s 中找不到文件 %s\n", hash, filePath)
			return
		}

		data, err := store.Read(blobHash)
		if err != nil {
			fmt.Printf("读取对象失败: %v\n", err)
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
			content = data
		}

		os.WriteFile(filePath, content, 0644)
		fmt.Printf("已恢复 %s 到 commit %s 时的版本\n", filePath, hash[:8])
		return
	}

	branchName := args[0]
	commitHash := getBranchRef(branchName)
	if commitHash == "" {
		fmt.Printf("分支不存在: %s\n", branchName)
		return
	}

	os.WriteFile(".xit/HEAD", []byte(fmt.Sprintf("ref: refs/heads/%s\n", branchName)), 0644)
	fmt.Printf("已切换到分支 '%s'\n", branchName)
}

func cmdDiff(hash1, hash2 string) {
	data1, err := store.Read(hash1)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash1, err)
		return
	}
	data2, err := store.Read(hash2)
	if err != nil {
		fmt.Printf("读取对象失败 %s: %v\n", hash2, err)
		return
	}

	content1 := skipHeader(data1)
	content2 := skipHeader(data2)

	lines1 := strings.Split(string(content1), "\n")
	lines2 := strings.Split(string(content2), "\n")

	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}

	fmt.Printf("--- %s\n", hash1[:8])
	fmt.Printf("+++ %s\n", hash2[:8])

	for i := 0; i < maxLen; i++ {
		line1 := ""
		line2 := ""
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}
		if line1 != line2 {
			if i < len(lines1) {
				fmt.Printf("-%s\n", line1)
			}
			if i < len(lines2) {
				fmt.Printf("+%s\n", line2)
			}
		}
	}
}

func skipHeader(data []byte) []byte {
	for i, b := range data {
		if b == 0 {
			return data[i+1:]
		}
	}
	return data
}

// ===== v4 新增：branch / merge =====

// cmdBranch 分支管理
//
// 三种用法：
//   xit branch          → 列出所有分支，当前分支标 *
//   xit branch <名字>   → 创建新分支（指向当前 HEAD）
//   xit branch -d <名字> → 删除分支
func cmdBranch(args []string) {
	// 没参数 = 列出分支
	if len(args) == 0 {
		current := getHeadBranch()
		entries, err := os.ReadDir(filepath.Join(".xit", "refs", "heads"))
		if err != nil {
			fmt.Printf("读取分支列表失败: %v\n", err)
			return
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if name == current {
				fmt.Printf("* %s\n", name)
			} else {
				fmt.Printf("  %s\n", name)
			}
		}
		return
	}

	// -d = 删除分支
	if args[0] == "-d" {
		if len(args) < 2 {
			fmt.Println("用法: xit branch -d <分支名>")
			return
		}
		name := args[1]
		current := getHeadBranch()
		if name == current {
			fmt.Printf("错误：不能删除当前所在的分支 '%s'\n", name)
			return
		}
		refPath := filepath.Join(".xit", "refs", "heads", name)
		if err := os.Remove(refPath); err != nil {
			fmt.Printf("删除分支失败: %v\n", err)
			return
		}
		fmt.Printf("已删除分支 '%s'\n", name)
		return
	}

	// 创建分支
	name := args[0]
	hash := getBranchRef(name)
	if hash != "" {
		fmt.Printf("错误：分支 '%s' 已存在\n", name)
		return
	}

	currentHash := getBranchRef(getHeadBranch())
	if currentHash == "" {
		fmt.Println("错误：当前分支没有提交记录，无法创建分支")
		return
	}

	if err := setBranchRef(name, currentHash); err != nil {
		fmt.Printf("创建分支失败: %v\n", err)
		return
	}
	fmt.Printf("已创建分支 '%s'（指向 %s）\n", name, currentHash[:8])
}

// cmdMerge 合并分支到当前分支
//
// 两种情况：
//   1. 快进合并（fast-forward）：当前分支是目标分支的祖先，直接把指针往前移
//   2. 合并提交：两条分叉的历史，创建一个合并 commit（两个 parent）
//
// 合并 commit 的 tree 用的是当前分支的 tree（简单处理，不搞三路合并）
func cmdMerge(branch string) {
	currentBranch := getHeadBranch()
	currentHash := getBranchRef(currentBranch)
	branchHash := getBranchRef(branch)

	if currentHash == "" {
		fmt.Println("错误：当前分支没有提交记录")
		return
	}
	if branchHash == "" {
		fmt.Printf("错误：分支 '%s' 不存在\n", branch)
		return
	}
	if currentHash == branchHash {
		fmt.Println("已经在同一个提交上了，无需合并")
		return
	}

	// 检查是不是快进合并（当前分支是目标分支的祖先）
	if isAncestor(currentHash, branchHash) {
		// 快进：直接把当前分支指针移到目标分支那里
		setBranchRef(currentBranch, branchHash)
		fmt.Printf("快进合并: 将 %s 向前推进到 %s\n", currentBranch, branchHash[:8])
		fmt.Printf("[%s %s] merge branch '%s'\n", currentBranch, branchHash[:8], branch)
		return
	}

	// 非快进：创建一个合并提交
	// tree 沿用当前分支的 tree（简单粗暴）
	info := parseCommit(currentHash)
	if info == nil {
		fmt.Println("错误：无法读取当前提交")
		return
	}

	commit := &object.Commit{
		TreeHash: info.treeHash,
		Parents:  []string{currentHash, branchHash},
		Author:   getAuthor(),
		Message:  object.MergeMessage(branch),
		Time:     time.Now(),
	}
	commitHash, err := store.Write(commit)
	if err != nil {
		fmt.Printf("写入合并提交失败: %v\n", err)
		return
	}

	if err := setBranchRef(currentBranch, commitHash); err != nil {
		fmt.Printf("更新分支引用失败: %v\n", err)
		return
	}

	fmt.Printf("合并分支 '%s' 到 %s\n", branch, currentBranch)
	fmt.Printf("合并提交: %s\n", commitHash[:8])
}
