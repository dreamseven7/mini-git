package object

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// ========== 接口 ==========

type Object interface {
	Serialize() []byte
}

func Hash(obj Object) string {
	data := obj.Serialize()
	hash := sha1.Sum(data)
	return fmt.Sprintf("%x", hash)
}

// ========== Blob ==========

type Blob struct {
	Size int
	Data []byte
}

func (b *Blob) Serialize() []byte {
	header := fmt.Sprintf("blob %d\x00", b.Size)
	return append([]byte(header), b.Data...)
}

// ========== Tree ==========

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

// ========== Commit ==========

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
