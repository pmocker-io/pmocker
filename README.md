# PMocker

> Docker for Project Management Systems

PMocker 把不同方法论（PMBOK 第六版 / PRINCE2 / ISO 21502 / 敏捷）的项目管理系统封装为可分享、可版本化、可差异升级的 `.pmi` 镜像，让个人/团队「一条命令启动一套完整的项目管理系统」。

## 当前状态

v1.0 MVP 已达成，M1-M12 里程碑全部完成（M1-M8 已推送 `origin/main`，M9-M12 已实现）：

| 里程碑 | 交付内容 | 状态 |
|------|------|------|
| M1 · 骨架 | 仓库 + go.work + gva subtree + 空 CLI | 已完成 |
| M2 · 内核 | EAV 引擎 + 三维权限 + 声明式工作流引擎 | 已完成 |
| M3 · 模块 | 9 个 gva 插件（server + web + pmocker 元数据） | 已完成 |
| M4 · 镜像 | `.pmi` 镜像打包 + OCI 解析 + diff/upgrade | 已完成 |
| M5 · 前端 | 23 个页面组件 + 9 个 API 封装 + 特色视图 | 已完成 |
| M6 · CLI | 12 个 CLI 命令 + 默认镜像 `pmbok6-hybrid.pmi` | 已完成 |
| M7 · 集成修复 | 实例模式自动初始化 + 静态资源服务 + 商业授权裁剪 | 已完成 |
| M8 · MCP | MCP 集成（8 内置工具 + 78 动态工具 + 双进程架构） | 已完成 |
| M9 · 字段补全 | 10 模块 schema/seed 字段补全 + team 团队管理模块 | 已完成 |
| M10 · 表单适配 | 前端 DynamicForm 动态表单 + Schema API | 已完成 |
| M11 · 可视化 | 甘特图 / 检入检出 / 变更 diff / 跨模块联动 | 已完成 |
| M12 · 业务闭环 | 基线 / 偏差 / 事件引擎 / 完成度 / 仪表盘 / PMO / 结项 / 任务中心 / 项目工作台 | 已完成 |

> 完整设计见 [需求文档.MD](需求文档.MD)，实现计划见 [docs/superpowers/plans/](docs/superpowers/plans/)。

## 快速开始

### 1. 构建 CLI

```bash
make build-cli
# 产物：cli/pmocker.exe
```

### 2. 启动一个 PMSystem（实例模式）

```bash
# 用默认 PMBOK 第六版混合型镜像启动一个实例，暴露 8080 端口
./cli/pmocker.exe run -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080

# 首次运行会自动构建 gva-server / gva-mcp 二进制和前端 dist
# 启动后访问 http://localhost:8080，默认账号 admin / 123456（验证码已自动禁用）
```

实例启动后：
- 9 大 PM 模块菜单自动授予管理员（无需手动在后台勾选）
- 数据库自动初始化（`PMOCKER_AUTO_INIT=1`）
- MCP 子进程在 8899 端口启动，AI 编辑器可接入

### 3. 常用 CLI 命令

```bash
pmocker ps                       # 查看运行中实例
pmocker stop <name|id>          # 停止实例（数据卷保留）
pmocker start <name|id>         # 重启已停止实例
pmocker logs <name|id> [-f] [--mcp|--access] [-n N]  # 查看实例日志（-f 跟踪，--mcp 看 MCP，--access 看访问日志，-n 最后 N 行）
pmocker inspect <name|id>       # 查看实例元信息 + MCP 连接配置
pmocker images                  # 列出本地镜像缓存
pmocker commit <name> -t <tag>  # 把实例当前状态导出为 .pmi 镜像
pmocker export <name> -o <file> # 导出为 .pmi 文件
pmocker diff <old.pmi> <new.pmi>   # 对比两镜像 schema 差异
pmocker upgrade <name> --to <new.pmi>  # 升级实例到新镜像版本
pmocker rm <name> [-v]          # 删除实例（-v 同时删数据卷）
```

> v1 不含 `pull` / `push` / `build`（PMocker Hub 后期）。

### 4. 环境变量

| 变量 | 说明 | 默认值 |
|------|------|------|
| `PMOCKER_HOME` | PMocker 数据根目录（images/instances/volumes） | `~/.pmocker` |
| `PMOCKER_AUTO_INIT` | 实例模式自动初始化数据库 | 未设置（`pmocker run` 自动置 1） |
| `PMOCKER_ADMIN_PASSWORD` | 管理员初始密码 | `123456` |
| `PMOCKER_MCP_TOKEN` | MCP 子进程调用后端的 API Token | 自动签发（100 年有效期） |

> Windows 下建议设置 `PMOCKER_HOME` 到项目目录避免 `~/.pmocker` 写入权限问题。

## 三层抽象

| 概念 | 说明 | Docker 类比 |
|------|------|-------------|
| **PMI 镜像** | 项目管理系统模板定义（模块清单、schema、菜单、权限、seed、主题、前端资源），内容寻址（SHA256） | Docker 镜像 |
| **PMSystem** | 运行中的 PM 系统实例，独立 gin-vue-admin 进程 + 独立数据卷 | Docker 容器 |
| **项目** | PMSystem 内管理的项目数据（WBS、进度、风险…），存储于数据卷 | 容器内应用数据 |

## 10 大模块

v1 默认镜像 `pmbok6-hybrid.pmi` 内置 10 个 gva 插件模块，对标 PMBOK 第六版 + 行业最佳实践：

| 模块 | 说明 | 特色视图/算法 |
|------|------|------|
| 需求管理 requirement | 需求分类、优先级、追踪矩阵 | 追踪矩阵表格 |
| 范围管理 scope | WBS 分解、范围基线 | el-tree 树形分解 |
| 交付物管理 deliverable | 版本控制、审批流、Check-in/Out | 版本追踪 |
| 进度管理 schedule | 任务依赖、关键路径 CPM、里程碑 | echarts 关键路径图 |
| 成本管理 cost | 预算分解、挣值 EVM | EVM 指标、S 曲线 |
| 风险管理 risk | 5×5 概率影响矩阵、应对策略 | 风险矩阵 |
| 问题管理 issue | 问题类型、看板视图 | kanban 看板 |
| 变更管理 change | CCB 流程、影响分析 | CCB 评审、变更日志 |
| 组织级项目 EPS | EPS 树、项目组合管理 | 树形组织 |
| 团队管理 team | 成员/角色/培训/绩效（PMBOK 资源管理） | 利用率计算 |

聚合视图层（M12）：
- **项目仪表盘 / PMO 看板 / 结项归档**：健康度 RAG、成本偏差、资源负荷
- **个人任务中心 / 项目工作台**：4 类任务聚合 + "我关注"子视图（P0/P1 可见性）
- **基线快照 / 偏差分析**：3 类基线 + 字段级 diff + SPI/CPI
- **事件引擎**：工作流节点 hook 触发基线生成、完成度刷新、变更应用

## MCP 集成

PMocker 通过 MCP（Model Context Protocol）让 AI 编辑器调用项目管理能力：

- **8 个内置工具**：`pmocker_create_requirement`、`pmocker_evm_calc`、`pmocker_critical_path`、`pmocker_risk_matrix` 等，位于 [gva/server/mcp/pmocker_*.go](gva/server/mcp/)
- **78 个动态工具**：10 模块全部 API 自动绑定为 MCP 工具，由 `AutoInitIfEmpty` 注册到 `sys_mcp_tools` 表
- **双进程架构**：实例模式下 `pmocker run` 同时拉起 gva-server（8080）和 MCP 服务（8899），共享数据卷，自动签发长期 API Token

在 AI 编辑器中接入（`pmocker inspect` 输出配置）：

```bash
# Claude Code
claude mcp add --transport http pmocker http://localhost:8899/mcp --header "x-token: <token>"
```

```json
// Cursor .cursor/mcp.json
{
  "mcpServers": {
    "pmocker": { "url": "http://localhost:8899/mcp", "headers": { "x-token": "<token>" } }
  }
}
```

## 项目结构

```
pmocker/
├── go.work                  # Go Workspace：cli / gva / pkg 三模块
├── cli/                     # PMocker CLI（Go + Cobra，12 命令）
│   ├── cmd/                 # run/ps/start/stop/rm/commit/export/images/inspect/rmi/diff/upgrade
│   └── internal/            # builder / instance / image
├── gva/                     # gin-vue-admin v3.0.0（Git Subtree）
│   ├── server/
│   │   ├── plugin/pmocker_*/   # ★ 10 个 PM 模块 gva 插件（后端）
│   │   ├── mcp/                 # MCP 工具（内置 + 动态注册）
│   │   ├── initialize/auto_init.go  # 实例模式自动初始化
│   │   └── service/pmocker/     # EAV 引擎 + RBAC + 工作流
│   └── web/
│       └── src/view/pmocker/    # ★ 23 个页面组件
├── pkg/                     # 共享库（被 cli 和 gva 引用）
│   └── pmocker/
│       ├── eav/             # EAV 类型 + schema 加载
│       ├── oci/             # .pmi 镜像格式（manifest/config/layer）
│       ├── workflow/        # 声明式工作流引擎
│       ├── diff/            # 镜像 diff + migration 生成
│       └── plugin/          # PMockerPlugin 接口 + 注册器
├── images/pmbok6-hybrid/    # 默认镜像源（9 模块 + 3 层）
└── docs/superpowers/plans/  # 里程碑实现计划（M1-M12）
```

### 插件元数据扩展

每个 PM 模块作为 gva 标准插件实现，额外含 `pmocker/` 元数据子目录：

```
gva/server/plugin/pmocker_<mod>/
├── plugin.go                # gva Plugin 接口 + InitPMocker() 钩子
├── api/ router/ service/    # gva 标准目录
└── pmocker/
    ├── schema.yaml          # EAV 字段定义
    ├── seed.yaml            # 字典/种子数据
    ├── menu.yaml            # 菜单树
    ├── workflows/*.yaml     # 声明式工作流模板
    └── manifest.yaml        # 插件清单
```

## 技术栈

- **CLI**: Go 1.24 + Cobra
- **后端**: Go + Gin（gin-vue-admin v3.0.0）
- **前端**: Vue 3 + Element Plus + Pinia + UnoCSS
- **MCP**: mark3labs/mcp-go
- **数据库**: SQLite（默认）/ MySQL / PostgreSQL
- **镜像格式**: 简化 OCI（`.pmi`，内容寻址）
- **上游追踪**: Git Subtree（`--prefix=gva --squash`）

## 开发

```bash
# 开发模式：后端 :8888 + 前端 :8080 + MCP :8889
make run-gva          # gva 后端
make run-gva-web       # gva 前端（Vite dev）
cd gva/server && go run ./cmd/mcp -config ./cmd/mcp/config.yaml   # MCP 独立进程

# 运行测试
make test
```

PMocker 自定义代码统一放在 `gva/server/plugin/pmocker_*`、`gva/server/service/pmocker`、`gva/server/model/pmocker`，避免与 gva 上游冲突。

## 后续演进

| 版本 | 主题 | 说明 |
|------|------|------|
| v1.0 | MVP（当前） | 单机 + 9 模块 + 12 CLI + MCP，PMBOK 第六版混合镜像 |
| v1.1 | PMocker Hub | `pull` / `push` / `login` / `search`，Git 仓库驱动 |
| v1.2 | 团队协作 | MySQL/PostgreSQL、多用户、并发锁、WebSocket |
| v1.3 | 方法论镜像 | PRINCE2 / ISO 21502 / 纯敏捷预置镜像 |
| v2.0 | 企业 PMO | 私有 Hub、跨实例组合视图、SSO、审计合规 |

## License

MIT
