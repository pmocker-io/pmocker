# PMocker

> Docker for Project Management Systems

PMocker 将项目管理系统封装为可分享的 `.pmi` 镜像，一条命令启动完整的项目管理系统。

## 快速开始

```bash
# 构建 CLI
make build-cli

# 查看版本
./cli/pmocker.exe version

# 查看帮助
./cli/pmocker.exe --help
```

## 三层抽象

| 概念 | 说明 | Docker 类比 |
|------|------|-------------|
| **PMI 镜像** | 项目管理系统模板 | Docker 镜像 |
| **PMSystem** | 运行中的 PM 系统实例 | Docker 容器 |
| **项目** | PMSystem 内管理的项目数据 | 容器内应用数据 |

## 项目结构

```
pmocker/
├── cli/       # PMocker CLI (Go + Cobra)
├── gva/       # gin-vue-admin v3 (Git Subtree)
├── pkg/       # 共享库 (EAV/OCI/Workflow/RBAC)
├── docs/      # 文档
└── go.work    # Go Workspace
```

详见 [需求文档.MD](需求文档.MD)。

## 技术栈

- CLI: Go 1.24 + Cobra
- 后端: gin-vue-admin v3
- 前端: Vue 3 + Element Plus + Pinia
- 数据库: SQLite (默认) / MySQL / PostgreSQL

## 开发

```bash
# 启动 gva 后端
make run-gva

# 启动 gva 前端
make run-gva-web
```

## License

MIT
