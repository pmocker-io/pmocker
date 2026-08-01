# PMocker M6 CLI 闭环阶段设计文档

## 1. 背景与目标

### 1.1 当前状态

M1-M5 已完成：

* M1 骨架：go.work + gva subtree + CLI 框架

* M2 内核：EAV 引擎 + 三维权限 + 工作流引擎

* M3 后端：9 个 gva 插件模块（server 端）

* M4 镜像：.pmi 镜像格式 + OCI 解析 + diff/upgrade 命令

* M5 前端：9 个模块 23 个页面

当前 CLI 命令状态：

* 已实现（M4）：`version` / `images` / `inspect` / `rmi` / `diff` / `upgrade`

* 占位 stub：`run` / `ps` / `stop` / `start` / `rm` / `commit` / `export`

### 1.2 目标

实现需求文档 M5 · CLI 闭环：

1. 7 个占位命令全部实现
2. 一条命令（`pmocker run`）启动完整 PM 系统
3. `commit` / `export` / `diff` / `upgrade` 全闭环
4. 文档与安装脚本

### 1.3 验收标准

* `pmocker run --image pmbok6-hybrid:latest -p 8080` 能启动 gva server，浏览器可访问

* `pmocker ps` 列出运行中实例

* `pmocker stop` / `pmocker start` 能正确管理实例生命周期

* `pmocker rm -v` 删除实例和数据卷

* `pmocker commit` 从实例导出新镜像到缓存

* `pmocker export` 导出 .pmi 文件到指定路径

***

## 2. 架构设计

### 2.1 进程模型

pmocker 作为进程管理器，fork/exec 预构建的 gva server 二进制。

* pmocker CLI 不依赖 gva 运行时，仅通过系统调用管理进程

* 每个 PMSystem 实例 = 一个 gva server 进程 + 独立数据卷 + 独立端口

* pmocker 记录 PID，通过 PID 管理进程生命周期（SIGTERM 停止）

### 2.2 二进制构建

首次 `pmocker run` 时检测 `~/.pmocker/bin/gva-server` 是否存在：

* 不存在 → 自动 `go build` gva server 二进制 + `npm run build` 前端

* 已存在 → 跳过构建，直接使用

构建产物：

* `~/.pmocker/bin/gva-server`：gva server 二进制

* gva server 内置静态文件服务（web/dist），用户只需访问端口

### 2.3 实例注册表

SQLite 数据库 `~/.pmocker/instances/instances.db`：

```sql
CREATE TABLE instances (
  id           TEXT PRIMARY KEY,          -- UUID
  name         TEXT UNIQUE NOT NULL,      -- 用户可读名称
  image_digest TEXT NOT NULL,             -- 引用 images 缓存的 digest
  image_ref    TEXT,                      -- 原始镜像名:tag
  port         INTEGER NOT NULL,          -- 暴露端口
  volume_id    TEXT NOT NULL,             -- 数据卷 UUID
  pid          INTEGER DEFAULT 0,        -- 进程 PID，0 = 已停止
  status       TEXT DEFAULT 'stopped',    -- running / stopped
  created_at   DATETIME NOT NULL,
  started_at   DATETIME,
  stopped_at   DATETIME
);
```

与 images 缓存的兼容性：

* `image_digest` 直接引用 `~/.pmocker/images/index.json` 中的 digest

* `pmocker commit` 产出新镜像写入 images 缓存

* `pmocker inspect <name|id>` 可同时查镜像和实例信息

### 2.4 数据卷结构

```
~/.pmocker/volumes/<volume_id>/
├── system.db      # gva 系统库（用户/角色/菜单/权限）
├── project.db     # 项目库（pm_entities + pm_attrs EAV + 业务索引）
└── uploads/        # 交付物文件存储
```

首次 `pmocker run` 时：

1. 创建数据卷目录
2. 从镜像 schema 层提取实体类型/字段定义，灌入 project.db
3. 从镜像 seed 层提取种子数据，灌入 system.db（默认菜单/角色/用户）

***

## 3. 组件设计

### 3.1 `cli/internal/instance/store.go` — 实例注册表

职责：SQLite 实例记录的 CRUD

```go
type Instance struct {
    ID           string
    Name         string
    ImageDigest  string
    ImageRef     string
    Port         int
    VolumeID     string
    PID          int
    Status       string  // "running" | "stopped"
    CreatedAt    time.Time
    StartedAt    *time.Time
    StoppedAt    *time.Time
}

type Store interface {
    Create(inst *Instance) error
    GetByID(id string) (*Instance, error)
    GetByName(name string) (*Instance, error)
    List(includeStopped bool) ([]*Instance, error)
    Update(inst *Instance) error
    Delete(id string) error
}
```

### 3.2 `cli/internal/instance/manager.go` — 进程管理

职责：gva server 进程的 fork/exec/stop/start

```go
type Manager struct {
    store      Store
    binPath    string  // ~/.pmocker/bin/gva-server
    volumeBase string  // ~/.pmocker/volumes
}

func (m *Manager) Run(imageRef string, opts RunOptions) (*Instance, error)
// 1. 解析镜像 ref → digest（通过 image store）
// 2. 创建数据卷
// 3. 提取镜像层灌入数据卷
// 4. fork gva-server（传入端口、数据卷路径）
// 5. 写入实例记录

func (m *Manager) Stop(idOrName string) error
// 发 SIGTERM 给 gva-server 进程，等待退出，更新 status

func (m *Manager) Start(idOrName string) error
// 重新 fork gva-server，使用已有数据卷

func (m *Manager) Remove(idOrName string, removeVolume bool) error
// 停止进程（如运行中）→ 删除实例记录 → 可选删除数据卷
```

### 3.3 `cli/internal/instance/volume.go` — 数据卷管理

职责：数据卷的创建/删除/路径管理

```go
type VolumeManager struct {
    base string  // ~/.pmocker/volumes
}

func (v *VolumeManager) Create() (volumeID string, err error)
// 创建 ~/.pmocker/volumes/<uuid>/{system.db,project.db,uploads/}

func (v *VolumeManager) Path(volumeID string) string
// 返回数据卷根路径

func (v *VolumeManager) Remove(volumeID string) error
// 删除整个数据卷目录
```

### 3.4 `cli/internal/builder/builder.go` — 二进制构建

职责：首次运行时自动构建 gva server 二进制

```go
type Builder struct {
    binPath    string  // ~/.pmocker/bin/gva-server
    gvaServer  string  // gva/server 源码路径
    gvaWeb     string  // gva/web 源码路径
}

func (b *Builder) Ensure() error
// 检查 binPath 是否存在，不存在则构建

func (b *Builder) buildServer() error
// cd gva/server && go build -o ~/.pmocker/bin/gva-server .

func (b *Builder) buildWeb() error
// cd gva/web && npm install && npm run build
```

***

## 4. 命令实现

### 4.1 `pmocker run`

```
pmocker run --image <.pmi或镜像名:tag> [flags]
```

Flags（已在 M1 定义）：

* `-i, --image`：指定 .pmi 镜像文件或镜像名

* `-p, --port`：暴露端口（默认 8080）

* `-n, --name`：实例名称

* `-d, --db`：数据库驱动（sqlite|mysql|postgres）

* `--db-dsn`：数据库 DSN

* `-v, --volume`：数据卷路径

* `--admin-password`：管理员密码

流程：

1. `builder.Ensure()` — 确保二进制存在
2. 解析镜像 ref → digest（本地缓存 or .pmi 文件导入）
3. 创建数据卷
4. 提取镜像 schema 层 → 灌入 project.db
5. 提取镜像 seed 层 → 灌入 system.db
6. fork gva-server（传入：端口、数据卷路径、DB 路径、admin-password）
7. 写入实例记录
8. 输出实例信息 + 访问 URL

### 4.2 `pmocker ps`

```
pmocker ps [-a]
```

查询 SQLite 实例表，表格输出：

* ID（前 12 位）

* NAME

* IMAGE

* PORT

* STATUS（running/stopped）

* CREATED

`-a` 显示已停止实例。

### 4.3 `pmocker stop`

```
pmocker stop <name|id>
```

1. 查找实例
2. 发 SIGTERM 给 PID
3. 等待进程退出（超时 30s 后 SIGKILL）
4. 更新 status=stopped, pid=0

### 4.4 `pmocker start`

```
pmocker start <name|id>
```

1. 查找实例（status=stopped）
2. fork gva-server（使用已有数据卷和端口）
3. 更新 status=running, pid=新PID

### 4.5 `pmocker rm`

```
pmocker rm <name|id> [-v]
```

1. 如运行中，先 stop
2. 删除实例记录
3. `-v`：同时删除数据卷

### 4.6 `pmocker commit`

```
pmocker commit <name|id> -t <新镜像名:tag> [-m "说明"]
```

1. 查找实例
2. 从实例数据卷读取当前 schema + 数据
3. 打包为 .pmi 镜像（4 层：plugins/schema/theme/assets）
4. 写入 images 缓存（index.json + blob）
5. 输出新镜像 digest

### 4.7 `pmocker export`

```
pmocker export <name|id> -o <file.pmi>
```

1. 查找实例
2. 从实例数据卷读取当前状态
3. 打包为 .pmi 文件
4. 直接落盘到指定路径（不写入缓存）

***

## 5. gva server 启动参数

pmocker fork gva-server 时传入环境变量：

```
PMOCKER_INSTANCE_ID=<uuid>
PMOCKER_VOLUME_PATH=~/.pmocker/volumes/<volume_id>
PMOCKER_PORT=<port>
PMOCKER_DB_DRIVER=sqlite
PMOCKER_SYSTEM_DB=~/.pmocker/volumes/<volume_id>/system.db
PMOCKER_PROJECT_DB=~/.pmocker/volumes/<volume_id>/project.db
PMOCKER_ADMIN_PASSWORD=<password>
GIN_MODE=release
```

gva server 需要修改启动逻辑（通过 `init()` 或环境变量读取 PMocker 配置），这部分在 M6 中实现。

***

## 6. 文件结构

```
cli/
├── cmd/
│   ├── root.go          # 已有
│   ├── version.go       # 已有
│   ├── run.go           # 修改：实现完整逻辑
│   ├── ps.go            # 新增：ps + stop + start + rm
│   ├── stop.go          # 新增（或合并到 ps.go）
│   ├── start.go         # 新增
│   ├── rm.go            # 新增
│   ├── commit.go        # 新增：commit + export
│   ├── export.go        # 新增
│   ├── images.go        # 已有
│   ├── inspect.go       # 修改：支持实例 inspect
│   ├── rmi.go           # 已有
│   ├── diff.go          # 已有
│   ├── upgrade.go       # 已有
│   └── commands.go      # 删除（stub 命令迁移到具体文件）
├── internal/
│   ├── instance/
│   │   ├── store.go     # SQLite 实例注册表
│   │   ├── manager.go   # 进程管理
│   │   └── volume.go    # 数据卷管理
│   └── builder/
│       └── builder.go   # gva server 二进制构建
├── go.mod
├── go.sum
└── main.go
```

***

## 7. 跨平台考虑

### 7.1 进程管理

* Windows：`os/exec` + `taskkill /PID /T`（SIGTERM 等价）

* Unix：`syscall.SIGTERM` + 超时后 `SIGKILL`

### 7.2 路径处理

* `os.UserHomeDir()` 获取用户目录

* `filepath.Join()` 拼接路径，兼容跨平台

***

## 8. 测试策略

* `store.go`：SQLite CRUD 单元测试（内存 DB）

* `volume.go`：数据卷创建/删除测试（临时目录）

* `manager.go`：进程 fork/stop 集成测试（使用简单 echo server 替代 gva）

* `builder.go`：构建逻辑测试（mock go build）

* CLI 命令：端到端测试（构建 → run → ps → stop → rm）

***

## 9. 范围边界

### M6 交付

* 7 个 CLI 命令实现

* SQLite 实例注册表

* 数据卷管理

* gva server 二进制自动构建

* gva server 环境变量启动适配

* 文档与安装脚本

### M6 不交付

* `pmocker pull` / `push` / `build`（v1.1）

* 多实例负载均衡

* WebSocket 实时通知（v1.2）

* 前端 dev 模式（仅生产模式）

