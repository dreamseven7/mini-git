# xit — 从零手写的简化版 Git

一个用 Go 语言实现的迷你版本控制系统，旨在通过代码理解 Git 的核心原理。

## 已实现的功能

| 命令 | 说明 |
|------|------|
| `xit init` | 初始化一个 `.xit` 仓库 |
| `xit hash-object <文件>` | 把文件存为 Blob 对象，打印 SHA1 哈希 |
| `xit cat-file <哈希>` | 查看某个对象的内容 |
| `xit ls-tree <哈希>` | 查看 Tree 对象里的文件列表 |
| `xit add <文件>` | 暂存文件（添加到暂存区） |
| `xit commit -m "信息"` | 创建提交（自动构建 Tree→Commit→更新分支） |
| `xit log` | 查看当前分支的提交历史 |

## 项目结构

```
xit/
├── cmd/main.go         # 命令行入口
├── object/object.go    # Git 对象模型（Blob / Tree / Commit）
├── store/store.go      # 对象存储层（读写磁盘）
├── store/index.go      # 暂存区管理（Index）
├── go.mod              # Go 模块定义
├── test.txt            # 测试用文件
└── README.md
```

## 核心概念

### Git 对象模型

xit 实现了 Git 的三种核心对象：

- **Blob（文件）** — 存储文件的原始内容
- **Tree（目录）** — 记录目录中有哪些文件及其对应的 Blob 哈希
- **Commit（提交）** — 记录一次快照，包含 Tree 哈希、父提交、作者、提交信息

### 工作流程

```
编辑文件 → xit add → 暂存区(.xit/index) → xit commit → 生成 Commit 对象
                                                              ↓
                                                       更新分支引用
```

## 快速体验

```bash
cd xit
go build -o xit.exe ./cmd

# 初始化仓库
.\xit.exe init

# 暂存并提交
.\xit.exe add test.txt
.\xit.exe commit -m "初始提交"

# 查看历史
.\xit.exe log

# 底层命令
.\xit.exe cat-file <哈希>
.\xit.exe ls-tree <哈希>
```

## 后续计划

- [ ] `xit branch` — 分支管理
- [ ] `xit checkout` — 切换分支 / 恢复文件
- [ ] `xit status` — 查看工作区状态
- [ ] `xit diff` — 查看文件差异
