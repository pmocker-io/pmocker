# M14 镜像生命周期闭环（commit/export/upgrade 三件套）

## 基本信息

- 提出日期：2026-08-16
- 当前状态：`active`
- 需求类型：平台核心能力补全（P0-2，项目审阅报告最高优先级）
- 优先级：高
- 需求文件：`aiDoc/memory/business/active/m14-image-lifecycle.md`

## 用户原始意图摘要

落实「项目管理系统的 Docker」核心承诺：`commit`/`export` 应把实例数据卷真实状态（sqlite + dist + uploads）打包进 `.pmi`；`upgrade` 应真正执行 diff→migration。此前三者均为「复制原始镜像/打印计划的壳」（项目审阅报告 P0-2）。

## 影响范围

- 后端：`pkg/pmocker/oci`（新增 data 层）、`pkg/pmocker/diff`（迁移 SQL 化，未做）
- CLI：`cli/cmd/commit.go`、`cli/cmd/export.go`（重写）、`cli/cmd/upgrade.go`（未做）、`cli/internal/instance/snapshot.go`、`commit.go`（新建）
- 文档：README commit/export 能力描述、M14 计划文档、本记忆
- 里程碑：M14（docs/superpowers/plans/2026-08-16-m14-image-lifecycle.md）

## 涉及对象

- 模块：pkg/pmocker/oci（LayerTypeData）、cli/internal/instance（快照 + commit 构建）
- 接口：`commit <name> -t <tag>`、`export <name> -o <file>`、`upgrade <name> --to <image>`（后两阶段）
- 数据：实例数据卷 → tar（system.db WAL checkpoint + dist + uploads）

## 已确认约束

- commit 前需**停止实例**或依赖 WAL checkpoint（快照实现用 `PRAGMA wal_checkpoint(TRUNCATE)`）
- commit 镜像 config.Version 加时间戳，保证 digest 唯一（避免覆盖源镜像）
- 原镜像的非 data 层（schema/plugin/theme/assets）原样保留，旧 data 层丢弃重建
- `upgrade` 迁移需备份 sqlite + 失败回滚（未实施，M14 后续 Task 5）
- commit：`type(scope): description`，中文描述

## 当前进展

- 2026-08-16：M14 立项计划文档 + 项目审阅报告归档（5a6c05a0）
- 2026-08-16：T1 OCI data 层（a6c53853）——`LayerTypeData` + 构建/回读测试
- 2026-08-16：T2 实例数据卷快照（dfa44587）——WAL checkpoint + tar 打包 system.db/dist/uploads，3 单测
- 2026-08-16：T3/T4 commit/export（c954b4fc + 96f38167）——原层 + data 层构建新镜像；**端到端验证通过**：commit pms-v12 产出 4 层镜像（schema/plugins/theme/data），data 层 317 文件含真实 system.db(2.4MB) + dist
- 2026-08-16：README 更新 commit/export 真实能力

## 后续待办

- [ ] **M14-T5 upgrade 真迁移**（最大未完成项）：diff → SQL 迁移（CREATE TABLE/ADD COLUMN/seed upsert）+ sqlite 备份回滚
- [ ] M14-T6 记忆/README 收尾（本文件已建）
- [ ] 镜像「可重建」验证：用 commit 产出的 .pmi 起新实例，确认数据可重现

## 更新规则

- 本文件只承载「镜像生命周期闭环」这一功能点
- upgrade 落地后在本文件标记完成并移入 `done/`
