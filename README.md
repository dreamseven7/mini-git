# xit — 从零手写的简化版 Git

一个用 Go 语言实现的迷你版本控制系统。

## 功能

| 命令 | 说明 |
|------|------|
| `xit init` | 初始化仓库 |
| `xit hash-object <文件>` | 文件存为 blob，返回哈希 |
| `xit cat-file <哈希>` | 查看对象内容 |
| `xit ls-tree <哈希>` | 查看 tree 对象 |
| `xit add <文件...>` | 暂存文件（支持 `*.go` 通配符） |
| `xit commit -m "信息"` | 创建提交（多个 `-m` 自动拼接） |
| `xit log` | 查看提交历史 |
| `xit status` | 查看工作区状态 |
| `xit checkout <哈希> -- <文件>` | 恢复文件 |
| `xit checkout <分支>` | 切换分支 |
| `xit diff <哈希1> <哈希2>` | 比较差异 |
| `xit branch` | 列出分支 |
| `xit branch <名字>` | 创建分支 |
| `xit branch -d <名字>` | 删除分支 |
| `xit merge <分支>` | 合并分支到当前分支 |

## 项目结构

```
xit/
├── cmd/main.go         # 命令行入口
├── object/object.go    # Git 对象模型
├── store/store.go      # 对象存储层
├── store/index.go      # 暂存区管理
├── go.mod
├── test-all.ps1        # 功能测试脚本
└── README.md
```

## 环境变量

- `XIT_AUTHOR` — 设置作者信息，默认 `xit 用户 <user@xit>`
