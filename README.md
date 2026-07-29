# xit — 从零手写的简化版 Git

一个用 Go 语言实现的迷你版本控制系统，旨在通过代码理解 Git 的核心原理。

## 已实现的功能

| 命令 | 说明 |
|------|------|
| `xit init` | 初始化一个 `.xit` 仓库 |
| `xit hash-object <文件>` | 把文件存为 Blob 对象，打印 SHA1 哈希 |
| `xit cat-file <哈希>` | 查看某个对象的内容 |
| `xit ls-tree <哈希>` | 查看 Tree 对象里的文件列表 |

## 项目结构

```
xit/
├── cmd/main.go         # 命令行入口
├── object/object.go    # Git 对象模型（Blob / Tree / Commit）
├── store/store.go      # 对象存储层（读写磁盘）
├── go.mod              # Go 模块定义
├── test.txt            # 测试用文件
└── README.md
```

## 快速体验

```bash
cd xit
go build -o xit.exe ./cmd
.\xit.exe init
.\xit.exe hash-object test.txt
.\xit.exe cat-file 271a4d0eb32d2fa090c7867f10ca4aefa7eb5450
```

## 后续计划

- [ ] `xit add` — 暂存文件
- [ ] `xit commit` — 创建提交
- [ ] `xit log` — 查看提交历史
- [ ] `xit branch` — 分支管理
