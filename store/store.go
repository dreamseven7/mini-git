// store/store.go — 存储层，负责把对象写到磁盘或从磁盘读出来
//
// v2 改动：index 相关代码移到 index.go 了，这个文件只保留核心的读写逻辑
package store

import (
	"fmt"
	"os"
	"path/filepath"

	"xit/object"
)

const gitDir = ".xit"
const indexPath = ".xit/index"

// Init 创建 .xit/objects/ 和 .xit/refs/heads/ 两个目录
// HEAD 和 index 文件等用的时候再创建，空文件反而麻烦
func Init() error {
	dirs := []string{
		filepath.Join(gitDir, "objects"),
		filepath.Join(gitDir, "refs", "heads"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %w", dir, err)
		}
	}
	return nil
}

// objectPath 根据哈希算磁盘路径，比如 "271a..." → ".xit/objects/27/1a..."
// 拿前两位当目录名是为了避免一个目录里文件太多，Git 也是这么干的
func objectPath(hash string) string {
	return filepath.Join(gitDir, "objects", hash[:2], hash[2:])
}

// Write 把对象存到 .xit/objects 里，返回它的 SHA1 哈希
// 流程：算哈希 → 序列化 → 算路径 → 创建目录 → 写文件
func Write(obj object.Object) (string, error) {
	hash := object.Hash(obj)
	data := obj.Serialize()
	path := objectPath(hash)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败 %s: %w", dir, err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("写入对象失败 %s: %w", path, err)
	}
	return hash, nil
}

// Read 根据哈希读取对象内容，返回的是带 Git 头部的原始字节
// 调用方得自己跳过头部拿真正的数据
func Read(hash string) ([]byte, error) {
	path := objectPath(hash)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取对象失败 %s: %w", path, err)
	}
	return data, nil
}
