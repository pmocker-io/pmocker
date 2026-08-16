# M14 镜像生命周期闭环（commit/export/upgrade 三件套）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 CLI 的 `commit`/`export`/`upgrade` 从「复制原始镜像的壳」落实为「镜像生命周期闭环」——commit/export 抓取实例数据卷真实状态（sqlite + dist + uploads + 业务数据）打包为新的 `.pmi`；upgrade 真正执行 diff→migration（SQLite 加列/建表/seed upsert）并支持备份回滚。对齐需求文档 1.4「OCI 镜像 + 差异升级」核心承诺。

**Architecture:** 复用 `pkg/pmocker/oci`（4 类层：plugins/schema/theme/assets）扩展为支持数据层；commit 把实例数据卷（system.db + dist + uploads）作为新层写入镜像；upgrade 解包新旧镜像 → `diff.DiffImages` → 生成 SQL 迁移 → 在实例 sqlite 上执行（备份 → 迁移 → 校验 → 回滚）。

**Tech Stack:** Go 1.25 / Cobra / modernc.org/sqlite（CLI 侧读实例 sqlite）/ gorm + sqlite（gva 侧执行迁移）/ OCI tar+manifest

## Global Constraints

- CLI module：`github.com/pmocker-io/pmocker/cli`；不修改 gva 原始文件（subtree pull 不冲突）
- 实例数据卷路径：`<home>/.pmocker/volumes/<volume_id>/`（或 `.pmocker-data/volumes/`）
- commit 需**停止实例**或**WAL checkpoint 后复制** sqlite（避免拷贝不一致）
- upgrade 迁移前**先备份 sqlite**（`.bak`），失败自动回滚
- 分层约束：CLI 不依赖 gva 源码包；复用 `pkg/pmocker/{oci,diff,image}`
- commit：`type(scope): description`，中文描述

---

## File Structure

### 新增
| 文件 | 职责 |
|------|------|
| `pkg/pmocker/oci/layer_data.go`（改造 builder） | 支持 data 层（实例数据） |
| `cli/internal/instance/snapshot.go`（新建） | 实例数据卷 → tar 快照（sqlite WAL checkpoint + dist + uploads） |
| `cli/internal/migrate/executor.go`（新建） | 执行 diff 生成的 SQL 迁移（备份/迁移/校验/回滚） |
| `cli/cmd/commit.go`（重写） | 抓真实态打包新镜像 |
| `cli/cmd/export.go`（重写） | 导出实例为独立 `.pmi` |
| `cli/cmd/upgrade.go`（重写） | diff → 迁移执行 |

### 修改
| 文件 | 修改 |
|------|------|
| `pkg/pmocker/diff/diff.go` | 迁移操作支持真实 SQL（CREATE TABLE / ADD COLUMN / seed upsert） |
| `cli/internal/builder/builder.go` | 同步新层到镜像缓存 |

---

## Task 1: OCI data 层支持

**Files:**
- Modify: `pkg/pmocker/oci/types.go` / `builder.go`
- Create: `pkg/pmocker/oci/layer_test.go`

**Interfaces:**
- Produces: `LayerTypeData LayerType = "data"`、`NewDataLayer(tarBytes)`、`BuildImage` 支持任意 LayerType

- [ ] **Step 1: types.go 增加 LayerTypeData**
- [ ] **Step 2: builder.go 的 descriptorFor 支持 data 层 mediaType**
- [ ] **Step 3: 测试构建含 data 层的镜像并回读**
- [ ] **Step 4: 提交 `feat(oci): data layer for instance snapshot`**

## Task 2: 实例数据卷快照

**Files:**
- Create: `cli/internal/instance/snapshot.go` + `snapshot_test.go`

**Interfaces:**
- Produces: `SnapshotVolume(vm *VolumeManager, volID string, outTar string) error`（WAL checkpoint + tar system.db/dist/uploads）

- [ ] **Step 1: sqlite WAL checkpoint（PRAGMA wal_checkpoint(TRUNCATE)）后复制 system.db**
- [ ] **Step 2: tar 打包（system.db + dist + uploads）**
- [ ] **Step 3: 测试（临时卷 → 快照 → 校验 tar 内容）**
- [ ] **Step 4: 提交 `feat(cli): instance volume snapshot`**

## Task 3: commit 抓真实态

**Files:**
- Rewrite: `cli/cmd/commit.go`

**Interfaces:**
- Consumes: `instance.SnapshotVolume`、`oci.BuildImage`、`image.Store`
- 行为：停止实例（或 checkpoint）→ 快照数据卷 → 原镜像层 + data 层 → 构建新 `.pmi` → 注册到镜像库

- [ ] **Step 1: 重写 commit（复用原镜像 layer，追加 data layer）**
- [ ] **Step 2: 支持 --message 写入 manifest annotations**
- [ ] **Step 3: 端到端验证（commit → 新镜像 inspect 含 data 层）**
- [ ] **Step 4: 提交 `feat(cli): commit captures instance state`**

## Task 4: export 导出实例镜像

**Files:**
- Rewrite: `cli/cmd/export.go`

**Interfaces:**
- Consumes: commit 产出的新镜像
- 行为：导出实例当前态 `.pmi` 到指定路径（与 commit 同源，独立文件）

- [ ] **Step 1: 重写 export（复用 commit 快照逻辑）**
- [ ] **Step 2: 提交 `feat(cli): export instance image`**

## Task 5: upgrade 真正执行迁移

**Files:**
- Rewrite: `cli/cmd/upgrade.go`
- Modify: `pkg/pmocker/diff/diff.go`
- Create: `cli/internal/migrate/executor.go` + `executor_test.go`

**Interfaces:**
- Consumes: `diff.DiffImages` → `GenerateMigration`（SQL 操作）→ `migrate.Executor`
- 行为：解包新旧镜像 → diff → 备份 sqlite → 逐条执行（CREATE TABLE / ADD COLUMN / seed upsert）→ 校验 → 失败回滚

- [ ] **Step 1: diff.GenerateMigration 输出真实 SQL 操作（当前仅描述）**
- [ ] **Step 2: executor.go（modernc sqlite 打开实例库执行迁移 + 备份/回滚）**
- [ ] **Step 3: 测试（旧 schema 库 → 迁移 → 新表/列存在；失败注入回滚）**
- [ ] **Step 4: 重写 upgrade 命令接入 executor**
- [ ] **Step 5: 端到端验证（pmbok6-hybrid 旧镜像 → 加字段新镜像 → 实例数据保留）**
- [ ] **Step 6: 提交 `feat(cli): upgrade executes real migration`**

## Task 6: 文档 + 记忆登记

**Files:**
- Modify: `README.md`（镜像生命周期章节）、`gva/aiDoc/memory/business/active/pmocker-config.md`（或新文件）

- [ ] **Step 1: README 更新 commit/export/upgrade 真实能力说明**
- [ ] **Step 2: 业务记忆登记 M14**
- [ ] **Step 3: 提交 `docs(pmocker): M14 image lifecycle docs`**

---

## 完成检查清单

- [ ] commit 产出的镜像包含实例真实数据（sqlite 数据可在新实例重现）
- [ ] export 产出独立可用的 `.pmi`
- [ ] upgrade 真执行迁移且不丢用户数据；失败可回滚
- [ ] 全部单测通过；端到端验证通过
- [ ] README/记忆层同步
