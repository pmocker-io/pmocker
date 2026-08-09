# M13 初始配置管理模块 设计文档

> 版本：v1.0
> 日期：2026-08-09
> 状态：设计已确认，待用户审阅
> 需求记忆：`aiDoc/memory/business/active/pmocker-config.md`

---

## 1. 背景与问题陈述

### 1.1 当前状态

PMocker M1-M12 完成，10 模块 + 聚合视图已落地。但项目管理的**种子/元数据配置**存在以下问题：

- **字段定义 / 实体类型 / 工作流 / 字典**目前由各插件 `pmocker/schema.yaml`/`seed.yaml`/`menu.yaml` 经 `loader` 在启动时灌入 `pm_field_defs`/`pm_entity_types`/`pm_workflow_defs` 等元表，**无 Web 管理入口**，修改需改 YAML + 重建镜像
- **状态与流转**硬编码在前端 `statusTransitions.js`，改状态机需改前端代码
- **业务种子数据**（3 项目任务/风险/问题等）在 `business_seed.yaml`，无页面编辑能力
- 组织架构/用户/角色/权限由 gva 内置 superAdmin 管理，但缺少统一配置入口

### 1.2 目标

新增「初始配置管理」模块（`pmocker_config` 插件），以 Web 页面方式管理各模块的字段、初始值、字典、状态流转、工作流、业务种子数据，配置项带**状态机管理**（草稿/评审/发布/归档），支持编辑、复制复用、删除；配置完善后**导出 YAML 固化到镜像源**，支撑 v1.1 升级。

---

## 2. 已确认决策（brainstorm 结论）

| # | 决策点 | 结论 |
|---|--------|------|
| 1 | 架构 | **方案 A**：EAV 元表直接 CRUD + 导出生成 YAML（复用现有表/loader/动态表单） |
| 2 | 生效机制 | **直写 DB + 导出固化**；仅 published 配置生效（状态机驱动） |
| 3 | 状态流转配置 | 新建 `pm_state_defs` 表，前端 `statusTransitions.js` 改读 API |
| 4 | 组织形态 | 独立 `pmocker_config` 插件（server + web + pmocker 元数据） |
| 5 | 页面组织 | 单菜单「初始配置」多子页（VerticalTabLayout） |
| 6 | 业务种子范围 | 完整 CRUD（新增/编辑/删除业务实体） |
| 7 | 组织/用户/权限 | 入口跳转 gva superAdmin，不重复建设 |
| 8 | 导出产物 | YAML 三件套（schema.yaml/seed.yaml/menu.yaml）到镜像源 |
| 9 | 状态管理 | 配置项统一状态机 `draft → reviewing → published → archived`，archived 可恢复 draft，draft 可删除；简化单人流转 |
| 10 | 复用语义 | 一键复制为 draft |

---

## 3. 方案架构

### 3.1 总体架构

```
┌────────────────────────────────────────────────────────┐
│                    前端配置层                            │
│  实体类型 │ 字段定义 │ 字典 │ 状态流转 │ 工作流 │ 业务种子 │
│            （VerticalTabLayout 单菜单多子页）            │
├────────────────────────────────────────────────────────┤
│                pmocker_config 插件（后端）              │
│  ConfigService（CRUD） │ StateMachineService │ Export  │
├────────────────────────────────────────────────────────┤
│                元数据层（已有，加 status）                │
│  pm_entity_types │ pm_field_defs │ pm_workflow_defs     │
│  pm_relation_types │ sys_dictionary（+明细）             │
│  新增：pm_state_defs（状态流转定义）                     │
├────────────────────────────────────────────────────────┤
│                业务数据层（EAV，已有）                   │
│  pm_entities │ pm_attrs（业务种子数据完整 CRUD）          │
└────────────────────────────────────────────────────────┘
```

### 3.2 数据模型

#### 3.2.1 现有表加 `status` 列（迁移）

| 表 | 新增列 | 说明 |
|----|--------|------|
| `pm_entity_types` | `status` VARCHAR(16) DEFAULT 'published' | 实体类型配置状态 |
| `pm_field_defs` | `status` VARCHAR(16) DEFAULT 'published' | 字段定义配置状态 |
| `pm_workflow_defs` | `status` VARCHAR(16) DEFAULT 'published' | 工作流定义配置状态 |
| `pm_relation_types` | `status` VARCHAR(16) DEFAULT 'published' | 关系类型配置状态 |

> **兼容**：默认值 `published`，现有 loader 灌入的数据自动为已发布，无缝兼容现网实例。

#### 3.2.2 新增 `pm_state_defs`（状态流转定义表）

```sql
CREATE TABLE pm_state_defs (
  id            BIGINT PRIMARY KEY AUTOINCREMENT,
  entity_type   VARCHAR(64) NOT NULL,   -- 实体类型（task/issue/change_request/...）
  status        VARCHAR(32) NOT NULL,   -- 状态值（open/in_progress/closed/...）
  label         VARCHAR(64),            -- 状态显示名（待处理/已关闭/...）
  tag_type      VARCHAR(16),            -- el-tag 类型（info/warning/success/danger/''）
  sort          INT DEFAULT 0,          -- 排序
  actions_json  TEXT,                   -- JSON: [{label, target, api, type}]
  config_status VARCHAR(16) DEFAULT 'published',  -- 配置自身状态（draft/reviewing/published/archived）
  UNIQUE(entity_type, status)
);
```

对应前端 `statusTransitions.js` 的结构，页面可配置每个状态的标签、样式、可执行流转动作。

#### 3.2.3 状态机（所有配置对象统一）

```
draft ──提交评审──> reviewing ──发布──> published ──归档──> archived
  ▲                    │                                  │
  └────编辑/保存────────┘                                  │
  └──────────────恢复为草稿────────────────────────────────┘
```

- `draft`：可编辑、可删除、可复制
- `reviewing`：待发布（MVP 单人流转，仅状态标记）
- `published`：**生效**（动态表单/列表/工作流读取）
- `archived`：归档只读，可恢复为 draft

### 3.3 插件结构（遵循 aiDoc 插件规范）

```
gva/server/plugin/pmocker_config/
├── plugin.go              # interfaces.Plugin + PMockerPlugin
├── model/                 # request/response struct
│   ├── entity_type.go
│   ├── field_def.go
│   ├── dictionary.go
│   ├── state_def.go
│   ├── workflow_def.go
│   └── seed_entity.go
├── api/                   # ConfigApi
│   ├── entity_type.go
│   ├── field_def.go
│   ├── dictionary.go
│   ├── state_def.go
│   ├── workflow_def.go
│   ├── seed_entity.go
│   └── export.go
├── service/               # 复用 pmocker 核心 service + 新写
│   └── config_service.go
├── router/                # InitConfig（中间件链四件套）
├── initialize/            # router.go + menu.go + api.go
└── pmocker/
    ├── menu.yaml          # 菜单注册
    ├── api.yaml           # API 注册（Casbin）
    └── manifest.yaml      # 插件清单
```

### 3.4 后端核心 Service

| Service | 职责 |
|---------|------|
| `ConfigService` | 6 类配置对象的 CRUD + `CopyAsDraft(id)` 复制为草稿 |
| `StateMachineService` | 统一状态流转：`SubmitReview/Approve(发布)/Archive/Restore/Delete`，校验状态合法性 |
| `ExportService` | 读 DB published 配置 → 序列化 → 生成 `schema.yaml`/`seed.yaml`/`menu.yaml` → 写镜像源 |

### 3.5 API 端点（统一 `/pmocker/config/*`）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/config/entityTypes` | 实体类型列表（含草稿，供配置页） |
| POST | `/config/entityType` | 新增实体类型（draft） |
| PUT | `/config/entityType/:id` | 更新实体类型 |
| POST | `/config/entityType/:id/copy` | 复制为 draft |
| POST | `/config/entityType/:id/transition` | 状态流转 |
| DELETE | `/config/entityType/:id` | 删除（仅 draft） |
| GET/POST/PUT/DELETE | `/config/fields...` | 字段定义 CRUD（按实体类型） |
| GET/POST/PUT/DELETE | `/config/dictionaries...` | 字典 + 明细 CRUD |
| GET/POST/PUT/DELETE | `/config/stateDefs...` | 状态流转定义 CRUD |
| GET/POST/PUT/DELETE | `/config/workflows...` | 工作流定义 CRUD |
| GET | `/config/seedEntities?entityType=&projectId=` | 业务种子实体列表 |
| POST/PUT/DELETE | `/config/seedEntity...` | 业务种子实体 CRUD |
| POST | `/config/export` | 导出 YAML 三件套到镜像源 |
| GET | `/config/stateDefs/public` | 已发布状态流转（前端 statusTransitions 读取） |

### 3.6 published 过滤生效链路

- `GetSchema`/`ListEntities`（eav.go）查询时**默认只返回 `status='published'`** 的配置
- 新增查询参数 `includeDraft=true` 供配置管理页预览全部
- loader 灌入时标记 `status='published'`

---

## 4. 前端设计

### 4.1 新增文件

- `gva/web/src/api/pmocker/config.js`
- `gva/web/src/view/pmocker/config/index.vue`（VerticalTabLayout 容器）
- `gva/web/src/view/pmocker/config/entityType.vue`
- `gva/web/src/view/pmocker/config/fieldDef.vue`
- `gva/web/src/view/pmocker/config/dictionary.vue`
- `gva/web/src/view/pmocker/config/stateDef.vue`
- `gva/web/src/view/pmocker/config/workflow.vue`
- `gva/web/src/view/pmocker/config/seedEntity.vue`
- `gva/web/src/view/pmocker/config/orgEntry.vue`（跳转 gva superAdmin）

### 4.2 子页组织（单菜单「初始配置」）

| 子页 | 功能要点 |
|------|---------|
| 实体类型 | 表格 CRUD + 复制 + 状态 tag + 流转按钮 |
| 字段定义 | 实体类型下拉筛选 + 字段表格 CRUD + 复制 + 流转 |
| 字典 | 字典列表 + 明细编辑（el-tree 或表格） |
| 状态流转 | 各模块状态定义表格（状态值/标签/样式/流转动作 JSON 编辑） |
| 工作流 | 工作流列表 + 定义 JSON 编辑 |
| 业务种子 | 项目/模块筛选 + 表格 + DynamicForm 编辑（完整 CRUD） |
| 组织权限 | 入口卡片跳转 gva superAdmin 各页面 |

### 4.3 状态流转交互

- 每个配置项行内显示状态 tag（复用 `statusOptions.js` 风格）
- 操作按钮：提交评审/发布/归档/恢复/删除/复制
- 按状态过滤显示（草稿/已发布/已归档）

### 4.4 statusTransitions.js 改造

- 启动/进入页面时从 `GET /config/stateDefs/public` 拉取状态流转配置
- `getTransitions(entityType)` 改为读全局配置（模块级缓存），**保留本地 fallback**（API 不可用或未配置时降级到内置定义）
- 保持现有列表页调用方式不变（`getTransitions` 签名兼容）

---

## 5. 测试策略

| 测试 | 说明 |
|------|------|
| `ConfigService` CRUD | 6 类配置对象增删改查 + 复制为 draft（testutil.NewMemoryDB） |
| `StateMachineService` | 状态机流转合法性：draft→reviewing→published→archived→restore，非法流转报错 |
| `ExportService` | 导出 YAML 与 loader 解析格式双向一致（schema.yaml/seed.yaml/menu.yaml） |
| published 过滤 | getSchema/ListEntities 只返回 published；includeDraft=true 返回全部 |
| 前端 | 状态流转交互、DynamicForm 读取 published schema（浏览器点测） |

---

## 6. 兼容性与 v1.1 衔接

- **loader 兼容**：现有 `loadSingleSchema`/`LoadSeed`/`LoadWorkflow` 灌入时显式设置 `status=published`
- **现网数据**：已有实例的元表加 `status` 列默认 `published`，零迁移成本
- **statusTransitions fallback**：API 未发布时前端降级到内置状态机，不影响现有页面
- **v1.1 升级**：配置页导出 YAML 三件套到 `images/pmbok6-hybrid/` → rebuild `.pmi` 镜像 → `pmocker upgrade` 应用到新实例；用户可在配置页调整种子数据后完成 v1.1 升级

---

## 7. 里程碑拆分（实施计划输入）

| 阶段 | 内容 |
|------|------|
| M13-A | 数据模型迁移（加 status 列 + pm_state_defs）+ 状态机 + ConfigService CRUD 后端 |
| M13-B | 前端配置页（7 子页）+ 状态流转交互 + statusTransitions 改读 API |
| M13-C | ExportService 导出 YAML + published 过滤改造 + 全量测试 + 端到端验证 |

---

## 8. 风险与依赖

| 风险 | 缓解 |
|------|------|
| 导出 YAML 与 loader 格式不一致 | ExportService 单测 + 双向校验测试 |
| 加 status 列影响现有查询 | 默认值 published 兼容；过滤仅影响配置读取，业务数据不受影响 |
| statusTransitions 改造影响现有列表 | 保留本地 fallback，渐进切换 |
| 工程量大（6 类对象 + 状态机 + 导出） | 分 3 个里程碑提交，每阶段独立验证 |
