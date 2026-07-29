// object/object.go — Git 对象模型
//
// v2→v3：加了 Time 字段
// v4：Parent 改成 Parents，支持合并提交（多个 parent）
package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Object interface {
	Serialize() []byte
}

func Hash(obj Object) string {
	data := obj.Serialize()
	hash := sha1.Sum(data)
	return fmt.Sprintf("%x", hash)
}

type Blob struct {
	Size int
	Data []byte
}

func (b *Blob) Serialize() []byte {
	header := fmt.Sprintf("blob %d\x00", b.Size)
	return append([]byte(header), b.Data...)
}

type Entry struct {
	Mode string
	Name string
	Hash string
}

type Tree struct {
	Entries []Entry
}

func (t *Tree) Serialize() []byte {
	var content []byte
	for _, entry := range t.Entries {
		line := fmt.Sprintf("%s %s\x00", entry.Mode, entry.Name)
		content = append(content, []byte(line)...)
		hashBytes, _ := hex.DecodeString(entry.Hash)
		content = append(content, hashBytes...)
	}
	header := fmt.Sprintf("tree %d\x00", len(content))
	return append([]byte(header), content...)
}

// Commit 一次提交。Parents 里可以有一个（普通提交）或两个（合并提交）父提交
type Commit struct {
	TreeHash string
	Parents  []string // v4 改了，原来叫 Parent（单数）
	Author   string
	Message  string
	Time     time.Time
}

func (c *Commit) Serialize() []byte {
	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	for _, p := range c.Parents {
		content += fmt.Sprintf("parent %s\n", p)
	}
	content += fmt.Sprintf("author %s\n", c.Author)
	content += fmt.Sprintf("timestamp %d\n", c.Time.Unix())
	content += "\n"
	content += c.Message
	content += "\n"

	header := fmt.Sprintf("commit %d\x00", len(content))
	return append([]byte(header), []byte(content)...)
}

// HasParent 检查某个哈希是否在 parent 列表里
// 合并时判断"是不是已经合并过了"会用到
func (c *Commit) HasParent(hash string) bool {
	for _, p := range c.Parents {
		if p == hash {
			return true
		}
	}
	return false
}

// FirstParent 返回第一个 parent（普通提交只有一个，合并提交中第一个是当前分支原来的 HEAD）
func (c *Commit) FirstParent() string {
	if len(c.Parents) > 0 {
		return c.Parents[0]
	}
	return ""
}

// MergeMessage 生成合并提交的默认信息
func MergeMessage(branch string) string {
	return fmt.Sprintf("Merge branch '%s'", branch)
}

// CommitParentsLine 把 parent 列表变成"parent xxx"字符串，方便打印
func CommitParentsLine(hashes []string) string {
	if len(hashes) == 0 {
		return ""
	}
	return strings.Join(hashes, ", ")
}
