// store/index.go — 暂存区管理
//
// v2 新增。之前没有暂存区，add 和 commit 是这次才加上的。
// 暂存区就是你 commit 前放文件的地方，想好要提交哪些文件就先 add 进来。
package store

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IndexEntry 暂存区里的一条记录
type IndexEntry struct {
	Mode string // 文件权限，目前一律 100644
	Hash string // 对应的 blob 哈希
	Name string // 文件路径
}

// ReadIndex 读取 .xit/index，返回所有暂存的文件
// 文件不存在就返回 nil，表示暂存区是空的
func ReadIndex() ([]IndexEntry, error) {
	f, err := os.Open(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []IndexEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// SplitN 限制 3 次，防止文件名里有空格被切碎
		parts := strings.SplitN(line, " ", 3)
		if len(parts) == 3 {
			entries = append(entries, IndexEntry{
				Mode: parts[0],
				Hash: parts[1],
				Name: parts[2],
			})
		}
	}
	return entries, scanner.Err()
}

// WriteIndex 把暂存区写回磁盘，传入 nil 就是清空
func WriteIndex(entries []IndexEntry) error {
	dir := filepath.Dir(indexPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var sb strings.Builder
	for _, e := range entries {
		sb.WriteString(e.Mode)
		sb.WriteString(" ")
		sb.WriteString(e.Hash)
		sb.WriteString(" ")
		sb.WriteString(e.Name)
		sb.WriteString("\n")
	}
	return os.WriteFile(indexPath, []byte(sb.String()), 0644)
}

// AddToIndex 往暂存区加一个文件，如果已经存在同名的就替换
// 比如先 add 了 test.txt，改了之后又 add 一次，后一个覆盖前一个
func AddToIndex(mode, hash, name string) error {
	entries, err := ReadIndex()
	if err != nil {
		return err
	}

	found := false
	for i, e := range entries {
		if e.Name == name {
			entries[i] = IndexEntry{Mode: mode, Hash: hash, Name: name}
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, IndexEntry{Mode: mode, Hash: hash, Name: name})
	}

	return WriteIndex(entries)
}
