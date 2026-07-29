package store

import (
	"fmt"
	"os"
	"path/filepath"

	"xit/object"
)

const gitDir = ".xit"

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

func objectPath(hash string) string {
	return filepath.Join(gitDir, "objects", hash[:2], hash[2:])
}

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

func Read(hash string) ([]byte, error) {
	path := objectPath(hash)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取对象失败 %s: %w", path, err)
	}
	return data, nil
}
