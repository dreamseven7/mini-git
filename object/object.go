// object/object.go — Git 对象模型定义
//
// v2 没改，对象模型一开始就设计得够用了。把 Blob/Tree/Commit 三种对象定义清楚就行。
//
// Git 本质上是内容寻址的文件系统，三种对象的关系：
//   Commit → Tree → Blob (文件)
//             ├─ Blob (文件)
//             └─ Tree (子目录)
//                  └─ Blob (文件)
package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// Object 接口，所有对象都能序列化成字节
type Object interface {
	Serialize() []byte
}

// Hash 算 SHA1 哈希，返回 40 位十六进制字符串
// 注意是对"带头部的完整数据"算，不是只算文件内容
func Hash(obj Object) string {
	data := obj.Serialize()
	hash := sha1.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// Blob 存文件内容，不存文件名，文件名是 Tree 的事
type Blob struct {
	Size int
	Data []byte
}

func (b *Blob) Serialize() []byte {
	header := fmt.Sprintf("blob %d\x00", b.Size)
	return append([]byte(header), b.Data...)
}

// Tree 里的一个条目
type Entry struct {
	Mode string
	Name string
	Hash string
}

// Tree 记录某个时刻目录里有哪些文件，相当于目录快照
type Tree struct {
	Entries []Entry
}

// Serialize 先把所有条目拼成 content，再加头部
// 每个条目的哈希存的是 20 字节二进制，不是 40 位十六进制
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

// Commit 一次提交。Parent 为空就是首次提交，多个 parent 就是合并（目前不支持）
type Commit struct {
	TreeHash string
	Parent   string
	Author   string
	Message  string
}

func (c *Commit) Serialize() []byte {
	content := fmt.Sprintf("tree %s\n", c.TreeHash)
	if c.Parent != "" {
		content += fmt.Sprintf("parent %s\n", c.Parent)
	}
	content += fmt.Sprintf("author %s\n", c.Author)
	content += "\n"
	content += c.Message
	content += "\n"

	header := fmt.Sprintf("commit %d\x00", len(content))
	return append([]byte(header), []byte(content)...)
}
