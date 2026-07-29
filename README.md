# xit — 从零写的简化版 Git

![Go Version](https://img.shields.io/badge/Go-1.23-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-学习中-orange)

一个用 Go 语言实现的迷你版本控制系统。不依赖任何第三方库，纯标准库实现，旨在通过手写代码理解 Git 的核心原理。

## 功能特性

- **对象存储** — Blob（文件） / Tree（目录快照） / Commit（提交）三种 Git 核心对象
- **内容寻址** — SHA1 哈希作为唯一地址，相同内容永远相同哈希
- **暂存区机制** — 选择性提交，想好要提交哪些文件再 commit
- **提交历史** — parent 链串联所有提交，完整追溯每次变更
- **分支管理** — 创建/切换/删除/合并分支，支持快进合并与合并提交
- **文件恢复** — 从任意历史版本恢复文件
- **差异对比** — 逐行比较两个版本的差异
- **工作区状态** — 一目了然看清暂存/未暂存/未跟踪的文件
- **通配符支持** — `xit add *.go` 批量暂存
- **可执行文件检测** — 自动识别 100755 权限

## 命令速览

| 命令 | 说明 |
|------|------|
| `xit init` | 初始化仓库 |
| `xit add <文件...>` | 暂存文件（支持 `*.go` 通配符） |
| `xit commit -m "信息"` | 创建提交（多个 `-m` 自动拼接） |
| `xit log` | 查看提交历史 |
| `xit status` | 查看工作区状态 |
| `xit branch` | 列出/创建/删除分支 |
| `xit merge <分支>` | 合并分支到当前分支 |
| `xit checkout <分支>` | 切换分支 |
| `xit checkout <哈希> -- <文件>` | 从指定版本恢复文件 |
| `xit diff <哈希1> <哈希2>` | 比较两个版本差异 |
| `xit hash-object <文件>` | 底层：存为 blob |
| `xit cat-file <哈希>` | 底层：查看对象内容 |
| `xit ls-tree <哈希>` | 底层：查看 tree 对象 |

## 技术栈

| 层次 | 技术 |
|------|------|
| 语言 | Go 1.23+ |
| 哈希 | crypto/sha1 |
| 存储 | 文件系统（`.xit/objects/xx/xxxx...`） |
| 依赖 | 零外部依赖，纯标准库 |

## 项目结构

```
xit/
├── cmd/main.go         # 命令行入口
├── object/object.go    # Git 对象模型（Blob / Tree / Commit）
├── store/store.go      # 对象存储层（读写磁盘）
├── store/index.go      # 暂存区管理
├── go.mod              # 模块定义
├── test-all.ps1        # 功能测试脚本
└── README.md
```

## 快速开始

### 前置要求

- Go 1.23 或更高版本

### 构建

```bash
cd xit
go build -o xit.exe ./cmd
```

### 快速体验

```bash
# 初始化仓库
.\xit.exe init

# 创建文件并暂存
"hello xit" | Set-Content readme.txt -NoNewline
.\xit.exe add readme.txt

# 提交
.\xit.exe commit -m "第一次提交"

# 查看历史
.\xit.exe log

# 查看状态
.\xit.exe status
```

### 分支与合并

```bash
# 基于当前提交创建分支
.\xit.exe branch feature

# 切换到新分支
.\xit.exe checkout feature

# 在新分支上修改并提交
# ...

# 切回主分支并合并
.\xit.exe checkout main
.\xit.exe merge feature
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `XIT_AUTHOR` | 作者信息 | `xit 用户 <user@xit>` |

## 设计思路

Git 本质上是内容寻址的文件系统。三种对象的关系：

```
Commit → Tree → Blob（文件）
          ├─ Blob（文件）
          └─ Tree（子目录）
               └─ Blob（文件）
```

- **Blob** 存文件内容，不存文件名（文件名在 Tree 里）
- **Tree** 存目录快照，记录文件名→Blob 哈希的映射
- **Commit** 存一次提交，指向 Tree + 父提交 + 作者 + 时间

每个对象通过 SHA1 哈希唯一标识，哈希的前两位作为目录名分片存储（`.xit/objects/ab/cd...`），避免单目录文件过多。


