# M8 - PMocker MCP 集成设计

* **日期**: 2026-08-01

* **阶段**: M8

* **状态**: 已确认待实现

* **依赖**: M3（9 大模块插件）、M6（CLI + PMI 镜像）、M7（实例化运行）

## 背景与目标

PMocker v1.0 MVP 已交付 9 大项目管理模块、CLI 工具、PMI 镜像格式和实例化运行能力。但 AI 编辑器（Claude Code、Cursor、Trae 等）无法直接操作 PMocker 的业务 API。

GVA v3.0 已内置 MCP 独立服务（端口 8889，基于 mark3labs/mcp-go），提供 17 个内置工具（API/菜单/字典/权限/组织等）和动态工具注册机制。本阶段通过复用 GVA MCP 基础设施，让 AI 编辑器能调用 PMocker 的项目管理能力。

**目标**：AI 编辑器连接 MCP 后，能用自然语言完成创建需求、查询进度、关键路径分析、变更影响分析等项目管理操作。

## 关键决策

| 决策项       | 选择               | 理由                                      |
| --------- | ---------------- | --------------------------------------- |
| 部署形态      | 开发环境 + 实例模式两者都要  | 开发时加速调试，实例模式供实战使用                       |
| 工具暴露方式    | 混合方式（内置 + 动态）    | 核心业务用内置（灵活嵌入 CPM/EVM 算法），CRUD 用动态（快速覆盖） |
| 实例 MCP 启动 | pmocker run 自动启动 | 开箱即用，用户无感知                              |
| 内置工具范围    | 8 个项目管理工具        | 覆盖 PMBOK 核心场景，平衡工作量                     |

## 1. 架构概览

复用 GVA MCP 独立进程，不重复造轮子。PMocker 的 MCP 能力作为 GVA MCP 的扩展。

```
开发模式（gva/server 目录）
  gva-server:8888  ←─── MCP:8889 ←── AI 编辑器
  (主服务+PMocker插件)   (转发请求,带x-token)

实例模式（pmocker run）
  gva-server:8080  ←─── MCP:8899 ←── AI 编辑器
  (pmocker自动拉起两进程, MCP随实例生命周期)
```

**三层工具暴露**：

1. GVA 原生 17 工具（已有）：API/菜单/字典/权限/组织等
2. PMocker 内置 8 工具（新增）：项目管理核心业务（CPM/EVM/影响分析等）
3. PMocker 动态工具（新增）：9 模块 CRUD API 通过 seed.yaml 自动注册

**鉴权链路不变**：AI 编辑器 → MCP(不鉴权) → gva-server(带 x-token，JWT+Casbin 校验)。PMocker 在 AutoInitIfEmpty 中自动签发长期 API Token（有效期 -1=100 年），写入实例元数据。

## 2. 内置工具定义（8 个）

工具文件位置：`gva/server/mcp/pmocker_*.go`，每个实现 `McpTool` 接口，在 `init()` 中 `RegisterTool`。

| 工具名                          | 描述     | 关键参数                                                            | 调用后端 API                           | 返回                       |
| ---------------------------- | ------ | --------------------------------------------------------------- | ---------------------------------- | ------------------------ |
| `pmocker_create_requirement` | 创建需求   | title, description, priority, module(可选，指定归属模块如 scope/schedule) | POST /pmocker/requirement/create   | 需求ID + 字段                |
| `pmocker_query_progress`     | 查询项目进度 | project\_id, scope(all\|active\|delayed)                        | GET /pmocker/schedule/progress     | 完成率/里程碑/延迟任务             |
| `pmocker_critical_path`      | 关键路径分析 | project\_id, view(list\|gantt)                                  | GET /pmocker/schedule/criticalPath | 关键路径链/总工期/浮动             |
| `pmocker_impact_analysis`    | 变更影响分析 | change\_id 或 from\_entity+to\_entity                            | GET /pmocker/change/impact         | 受影响需求/任务/成本              |
| `pmocker_risk_matrix`        | 风险矩阵评估 | project\_id, threshold(可选)                                      | GET /pmocker/risk/matrix           | 高/中/低分布 + Top 风险         |
| `pmocker_evm_calc`           | 挣值计算   | project\_id, as\_of\_date(可选)                                   | GET /pmocker/cost/evm              | PV/EV/AC/SPI/CPI/EAC/VAC |
| `pmocker_query_issues`       | 问题查询   | project\_id, status, severity, assignee                         | GET /pmocker/issue/list            | 问题列表 + 统计                |
| `pmocker_manage_status`      | 状态流转   | entity\_type, entity\_id, action, comment                       | POST /pmocker/workflow/transition  | 新状态 + 工作流历史              |

**HTTP 调用机制**：复用 GVA MCP 的 `http_client.go`，工具内部通过 MCP 上游客户端调用 gva-server（带 x-token 头），不直接访问 DB。

## 3. 动态工具与 seed.yaml 自动注册

**覆盖范围**：9 大模块的 CRUD API，排除已是内置工具的接口。约 30+ 个动态工具。

**注册机制**：在 PMocker 插件的 `seed.yaml` 中预定义动态工具绑定，AutoInitIfEmpty 初始化时自动插入到 GVA 的 `sys_mcp_tools` 表。MCP 进程启动时通过 `registerDynamicTools` 读取并注册。

**seed.yaml 工具绑定格式**：

```yaml
mcp_dynamic_tools:
  - name: "pmocker_requirement_list"
    description: "查询需求列表，支持分页和筛选"
    api_path: "/api/pmocker/requirement/list"
    method: "GET"
    parameters:
      - name: "page"
        type: "integer"
        description: "页码"
        required: false
      - name: "pageSize"
        type: "integer"
        description: "每页数量"
        required: false
      - name: "keyword"
        type: "string"
        description: "需求标题关键字"
        required: false

  - name: "pmocker_requirement_create"
    description: "创建需求"
    api_path: "/api/pmocker/requirement/create"
    method: "POST"
    parameters:
      - name: "title"
        type: "string"
        description: "需求标题"
        required: true
      - name: "priority"
        type: "string"
        description: "优先级 (high/medium/low)"
        required: false
```

**9 模块动态工具分布**：

| 模块          | 动态工具数 | 工具                                             |
| ----------- | ----- | ---------------------------------------------- |
| requirement | 4     | list, create, update, delete                   |
| scope       | 4     | list, create, update, delete                   |
| schedule    | 5     | list, create, update, delete, milestones       |
| cost        | 5     | list, create, update, delete, budgets          |
| risk        | 4     | list, create, update, delete（matrix 是内置）       |
| issue       | 4     | list, create, update, delete（list 也是内置，动态版更简单） |
| eps         | 4     | list, create, update, delete                   |
| deliverable | 5     | list, create, update, delete, versions         |
| change      | 4     | list, create, update, delete（impact 是内置）       |

**去重策略**：`issue/list` 和 `risk/matrix` 同时有内置和动态版本。内置版参数更丰富（聚合统计），动态版是纯 API 透传。两者工具名不同，不冲突。

## 4. 实例模式 MCP 自动启动

**双进程架构**：`pmocker run` 拉起 gva-server + MCP 两个进程，共享数据卷。

**端口分配**：

* gva-server 端口 = 用户指定（如 8080）

* MCP 端口 = 8899

* 端口冲突处理：若 MCP 端口被占用，逐步 +1 探测（8900、8901...），最多探测 10 个端口，仍冲突则报错并提示用户用 `-p` 指定其他 gva-server 端口

**MCP 配置生成**：为每个实例生成独立的 `mcp_config.yaml`，写入数据卷：

```yaml
mcp:
  name: PMOCKER_MCP_<实例ID前8位>
  version: v1.0.0
  path: /mcp
  addr: 8899              
  base_url: http://127.0.0.1:8899/mcp
  upstream_base_url: http://127.0.0.1:8080  # 指向实例gva-server
  auth_header: x-token
  request_timeout: 15
```

**鉴权 Token 自动签发**：

* AutoInitIfEmpty 完成后，调用 API Token 接口签发长期 Token（有效期 -1=100 年）

* Token 写入实例元数据（`instances` 表新增 `mcp_token` 字段）

* `pmocker ps` 显示 MCP 端口

* `pmocker inspect <id>` 输出完整 MCP 连接配置（含 token，供 AI 编辑器直接复制）

**进程生命周期**：

| pmocker 命令 | gva-server | MCP                |
| ---------- | ---------- | ------------------ |
| `run`      | 启动         | 启动（gva-server 就绪后） |
| `stop`     | 停止         | 停止（先于 gva-server）  |
| `start`    | 启动         | 启动                 |
| `rm`       | 停止+删除      | 停止+删除              |

**实现位置**：

* `cli/internal/instance/manager.go`：`Start` 方法启动 gva-server 后，再启动 MCP 子进程

* `cli/internal/instance/manager.go`：`Stop` 方法先停 MCP，再停 gva-server

* `gva/server/initialize/auto_init.go`：AutoInitIfEmpty 末尾签发 Token

* `cli/internal/instance/store.go`：instances 表新增 mcp\_port、mcp\_token 字段

**pmocker inspect 输出示例**：

```
实例 ID: 04072407-37d0
gva-server: http://localhost:8080
MCP 服务:   http://localhost:8899/mcp
MCP Token:  eyJhbGciOiJIUzI1NiIs...

AI 编辑器配置（Claude Code）:
  claude mcp add --transport http pmocker http://localhost:8899/mcp --header "x-token: eyJhbGci..."
```

## 5. 开发模式接入 + 测试策略

**开发模式接入（零改动）**：

* gva-server:8888（已有）+ MCP:8889（已有 `cmd/mcp/config.yaml`）

* PMocker 8 个内置工具放在 `gva/server/mcp/pmocker_*.go`，编译进 gva-server 二进制

* 开发时启动：`cd gva/server && go run ./cmd/mcp -config ./cmd/mcp/config.yaml`

* AI 编辑器连 `http://127.0.0.1:8889/mcp`，token 用 GVA 后台「API Token」页面签发

**动态工具开发模式验证**：

* seed.yaml 中预定义的动态工具，在 AutoInitIfEmpty 时插入 `sys_mcp_tools` 表

* 开发模式重置数据库后，重新初始化即生效

* MCP 进程启动时 `registerDynamicTools` 自动加载

**测试策略**：

| 层级      | 测试内容                             | 方式                                                  |
| ------- | -------------------------------- | --------------------------------------------------- |
| 单元测试    | 8 个内置工具的参数解析、错误处理                | `gva/server/mcp/pmocker_*_test.go`，mock HTTP client |
| 集成测试    | MCP 工具 → gva-server API → DB 全链路 | 启动 gva-server + MCP，用 mcp-go client 调用工具            |
| seed 测试 | 动态工具 seed.yaml 解析 + 数据库插入        | 验证 `sys_mcp_tools` 表记录数                             |
| 实例 E2E  | `pmocker run` → MCP 就绪 → 工具可调用   | 浏览器 + curl 验证 `/health` 和工具列表                       |

## 实现范围（文件清单）

**新增文件**：

* `gva/server/mcp/pmocker_create_requirement.go` + 测试

* `gva/server/mcp/pmocker_query_progress.go` + 测试

* `gva/server/mcp/pmocker_critical_path.go` + 测试

* `gva/server/mcp/pmocker_impact_analysis.go` + 测试

* `gva/server/mcp/pmocker_risk_matrix.go` + 测试

* `gva/server/mcp/pmocker_evm_calc.go` + 测试

* `gva/server/mcp/pmocker_query_issues.go` + 测试

* `gva/server/mcp/pmocker_manage_status.go` + 测试

**修改文件**：

* `gva/server/plugin/pmocker/*/seed.yaml`：9 个模块各新增 mcp\_dynamic\_tools 段

* `gva/server/initialize/auto_init.go`：AutoInitIfEmpty 末尾签发 API Token

* `cli/internal/instance/manager.go`：Start/Stop 管理 MCP 子进程

* `cli/internal/instance/store.go`：instances 表新增 mcp\_port、mcp\_token 字段

* `cli/cmd/inspect.go`：输出 MCP 连接信息

## 验收标准

1. **开发模式**：`go run ./cmd/mcp` 启动后，AI 编辑器能列出 PMocker 的 8 个内置工具 + 30+ 动态工具
2. **实例模式**：`pmocker run` 后，`pmocker inspect` 输出 MCP 连接信息，AI 编辑器可连接并调用工具
3. **工具调用**：`pmocker_query_progress` 能返回真实项目进度数据
4. **无 404**：所有内置工具调用的后端 API 都已注册（需先修复 scope/schedule/eps 的路由 404）

