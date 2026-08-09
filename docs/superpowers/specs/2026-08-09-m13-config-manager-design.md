# M13 初始配置管理模块（聚合配置包模型） 设计文档

> 版本：v2.0（推翻 v1.0「6 类对象各自 CRUD」，改为聚合配置包模型）
> 日期：2026-08-09
> 状态：方向已确认，待实施
> 需求记忆：`aiDoc/memory/business/active/pmocker-config.md`

---

## 1. 背景与方向修正

### 1.1 v1.0 偏差

v1.0 设计把配置拆成 6 类对象（实体类型/字段/字典/状态/工作流/业务种子）各自 CRUD。**这是错误理解**——用户期望的配置管理是**聚合配置包**：一条配置记录包含该模块的完整定义。

### 1.2 用户期望（修正后）

- **一条配置记录（配置包）** = 实体类型 + 实体的字段 + 初始值（种子数据，含项目）+ 状态定义 + 流转规则（含退回规则）
- **交互模式**：类似 gva 自动化代码管理——配置包列表 → 点击 → 进入可编辑的配置详情
- **每条配置记录**：状态管理（草稿/评审/发布/归档）+ 版本管理（快照/回滚）
- **EPS 项目**：可新增/修改（当前有 bug）
- **完整种子数据** = 项目(EPS树，含状态及种子) × 各模块 × 模块字段 × 字段种子 + 模块 × 状态定义 × 流转规则

---

## 2. 核心模型：配置包（Config Package）

### 2.1 概念

**配置包** 是 PMocker 系统配置的基本单元，对标 gva 自动化代码的"包"（SysAutoCodePackage + History）。

```
pm_config_packages（配置包表）
├── id
├── code                    -- 配置包编码（requirement/schedule/eps/...）
├── name                    -- 显示名（需求管理/进度管理/EPS...）
├── description
├── version                 -- 当前版本号（int，从 1 递增）
├── status                  -- draft / reviewing / published / archived
├── seed_yaml               -- TEXT：完整种子数据（YAML 真源）
├── entity_type             -- 本包对应的实体类型（eps_node 为 EPS 配置包）
├── module                  -- 所属模块
├── created_by / created_at / updated_at
```

```
pm_config_versions（配置包版本历史表）
├── id
├── package_id              -- → pm_config_packages.id
├── version                 -- 版本号
├── snapshot_yaml           -- 不可变快照（发布时的 seed_yaml 全量）
├── flag                    -- 0=发布 1=回滚
├── created_by / created_at
```

### 2.2 seed_yaml 结构（聚合所有信息）

```yaml
entity_type: requirement
module: requirement
name: 需求管理
fields:                              # 实体字段定义
  - {key: code, label: 需求编码, data_type: string}
  - {key: priority, label: 优先级, data_type: enum, options: [P0,P1,P2,P3], default: P2}
  - {key: source, label: 来源, data_type: enum, options: [客户,市场,内部,合规]}
states:                              # 状态定义
  - {status: draft, label: 草稿, tag_type: info}
  - {status: reviewing, label: 评审中, tag_type: warning}
  - {status: published, label: 已发布, tag_type: success}
transitions:                         # 流转规则（含退回）
  - {from: draft, to: reviewing, action: submit}
  - {from: reviewing, to: published, action: approve}
  - {from: reviewing, to: draft, action: reject, rollback: true}   # 退回
projects:                            # 项目级种子数据（引用 EPS 项目）
  - project_id: 3                    # → EPS 配置包的叶子项目
    entities:                        # 该项目下本模块的实体种子
      requirement:
        - {title: 排产算法, status: published, priority: P0}
        - {title: 可视化看板, status: draft, priority: P1}
      issue:
        - {title: 算法性能瓶颈, status: closed, severity: high}
```

### 2.3 配置包粒度

| 配置包 | entity_type | 说明 |
|--------|-------------|------|
| `eps` | eps_node | **独立 EPS 配置包**，描述树层级 |
| `requirement` | requirement | 需求管理 |
| `schedule` | task | 进度管理 |
| `risk` | risk | 风险管理 |
| `issue` | issue | 问题管理 |
| `change` | change_request | 变更管理 |
| `deliverable` | deliverable | 交付物管理 |
| `scope` | scope_item | 范围管理 |
| `cost` | cost_item | 成本管理 |
| `team` | team_member | 团队管理 |

> 每个业务模块一个配置包；EPS 树独立一个配置包。

### 2.4 EPS 树种子（树层级）

EPS 配置包的 seed_yaml `projects` 描述**树节点层级**（集团→事业部→项目集→项目）：

```yaml
entity_type: eps_node
module: eps
name: 组织EPS
fields:
  - {key: type, label: 节点类型, data_type: enum, options: [group,division,program,project,subproject]}
  - {key: code, label: 编码, data_type: string}
states: [...]
transitions: [...]
projects:
  - code: GROUP_HQ
    name: 集团总部
    type: group                  # 非底层 = 抽象容器（类比菜单 RouterHolder）
    children:
      - code: DIV_RND
        name: 智能排产研发部
        type: division           # 抽象容器
        children:
          - code: PROJ_A
            name: 智能排产系统研发
            type: project        # 叶子 = 基本单元项目
            status: active
            priority: 1
```

> 类比菜单管理：有子菜单 → RouterHolder（抽象容器）；无子菜单 → 具体 .vue 页面（基本单元项目）。业务模块配置包的种子通过 `project_id` 引用 EPS 配置包中叶子项目的实际实体 ID。

---

## 3. 核心流程

### 3.1 配置包生命周期

```
创建配置包 → 编辑 seed_yaml（前端编辑页，保存为 draft）
  → 提交评审（reviewing）
  → 发布（published）
      ├── 1. 校验 seed_yaml（YAML 合法 + 结构完整）
      ├── 2. 自动同步运行表（事务内）：
      │     ├── 实体类型 → pm_entity_types
      │     ├── 字段 → pm_field_defs
      │     ├── 状态/流转 → pm_state_defs
      │     ├── 项目(仅EPS包) → pm_entities(eps_node) + EPS树
      │     └── 业务实体种子 → pm_entities + pm_attrs
      ├── 3. 生成版本快照 → pm_config_versions
      └── 4. version++
  → 归档（archived）/ 回滚到历史版本（restore）
```

### 3.2 发布时自动同步 DB（核心）

**编辑 seed_yaml 不写运行 DB**，**发布时一次性同步**（事务内，失败整体回滚）：

| seed_yaml 段 | 同步目标 |
|--------------|----------|
| `fields` | `pm_field_defs`（status=published） |
| `states` + `transitions` | `pm_state_defs`（聚合为 status + actions_json） |
| `projects`（EPS 包） | `pm_entities`(eps_node) + `pm_eps_tree`（树层级） |
| `projects[].entities`（业务包） | `pm_entities` + `pm_attrs`（按 project_id 归属） |

**幂等**：按配置包 code + 版本去重；重新发布同版本覆盖，新版本增量。

### 3.3 版本管理

- 每次发布生成不可变快照（`pm_config_versions`）
- 支持查看历史版本、**回滚到任意版本**（flag=1，恢复 snapshot_yaml 到 seed_yaml 并重新发布同步）

---

## 4. 后端设计

### 4.1 新增表（model）

| 表 | model | 说明 |
|----|-------|------|
| `pm_config_packages` | `PMConfigPackage` | 配置包（含 seed_yaml TEXT） |
| `pm_config_versions` | `PMConfigVersion` | 版本快照 |

> 复用现有：`pm_entity_types`/`pm_field_defs`（已有 status 列）、`pm_state_defs`（已有）、`pm_entities`/`pm_attrs`。

### 4.2 Service

| Service | 方法 |
|---------|------|
| `ConfigPackageService` | `List/Create/Get/Update(seed_yaml)/Delete`、`CopyAsDraft` |
| `ConfigStateMachine` | `SubmitReview/Approve(发布→同步DB)/Archive/Restore(回滚)/Delete` |
| `SeedSyncService` | `SyncPackageToDB(ctx, package)`：seed_yaml → 运行表（核心） |
| `SeedParser` | `ParseSeedYAML(bytes) → ConfigPackageSeed`（YAML 解析为结构化） |

### 4.3 API（统一 `/pmocker/config/*`）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/pmocker/config/packages` | 配置包列表（含状态/版本） |
| POST | `/pmocker/config/package` | 新建配置包（draft） |
| GET | `/pmocker/config/package/:id` | 获取配置包详情（含 seed_yaml） |
| PUT | `/pmocker/config/package/:id` | 更新 seed_yaml（draft/reviewing 可编辑） |
| POST | `/pmocker/config/package/:id/copy` | 复制为 draft |
| POST | `/pmocker/config/package/:id/transition` | 状态流转（submit_review/approve/archive/restore） |
| DELETE | `/pmocker/config/package/:id` | 删除（仅 draft） |
| GET | `/pmocker/config/package/:id/versions` | 版本历史 |
| POST | `/pmocker/config/package/:id/rollback` | 回滚到指定版本 |
| POST | `/pmocker/config/export` | 导出所有 published 配置包为 YAML 到镜像源 |

> `approve`（发布）触发 `SeedSyncService.SyncPackageToDB`。

---

## 5. 前端设计

### 5.1 页面结构

```
初始配置（菜单）
├── 配置包列表（config/packageList.vue）   ← 主入口，对标 gva autoCode 包列表
│   ├── 表格：编码/名称/实体类型/状态/版本/更新时间/操作
│   ├── 操作：编辑/复制/提交评审/发布/归档/回滚/删除
│   └── 新建配置包按钮
└── 配置包编辑（config/packageEditor.vue）  ← 点击进入
    ├── 基本信息（code/name/entity_type/module）
    ├── seed_yaml 编辑
    │   ├── 表单模式：字段表格 + 状态表格 + 流转表格 + 项目种子树
    │   └── 或 YAML 文本编辑器（进阶）
    ├── 版本历史侧栏
    └── 保存（draft）/ 提交评审 / 发布
```

### 5.2 交互

- **配置包列表**：对标 gva 自动化代码包管理——列表 → 点击行 → 进入编辑
- **编辑页**：分区块编辑（基本信息 / 字段 / 状态流转 / 项目种子），保存为 draft
- **发布**：触发后端同步 DB，成功后状态变 published，version++
- **回滚**：从版本历史选择 → 确认 → 恢复到该版本 seed_yaml

### 5.3 EPS 项目编辑修复

- 修复 `eps/tree.vue`：`createEPSNode`/`updateEPSNode` 传参与后端 `CreateNode`/`UpdateNode` 期望对齐（`title`/`name` 映射）

---

## 6. 测试策略

| 测试 | 说明 |
|------|------|
| `SeedParser` | seed_yaml 解析为结构化（YAML 合法/结构完整/错误处理） |
| `SeedSyncService` | 发布同步：seed_yaml → 各运行表正确写入；幂等；事务回滚 |
| `ConfigStateMachine` | 状态机流转合法性 + 发布触发同步 |
| `ConfigPackageService` | 配置包 CRUD + 复制 |
| 版本回滚 | 发布 → 回滚 → 恢复 snapshot_yaml 并重新同步 |
| EPS 项目编辑 | 前端 title/name 对齐后新增/修改项目成功 |
| 端到端 | 创建配置包 → 编辑种子 → 发布 → 验证运行表生效 |

---

## 7. 与 v1.1 衔接

- 配置包发布后 seed_yaml 即"种子数据真源"
- 导出所有 published 配置包为 YAML → `images/pmbok6-hybrid/` → rebuild `.pmi` → v1.1 upgrade
- 用户可在配置包编辑页直接调整种子数据，发布后 DB 同步更新，再导出重建镜像

---

## 8. 兼容性与清理

- **保留复用**：元表 `status` 列、`pm_state_defs` 表、published 过滤（eav.go）、状态机概念
- **推翻**：`pmocker_config` 的 6 类对象 CRUD（ConfigService/StateMachineService/ExportService 改造或删除）、前端 7 子页（改为配置包列表+编辑）
- **现有业务种子**（3 项目）：迁移为对应配置包的 seed_yaml projects 段（EPS 包 + 各业务包）

---

## 9. 里程碑拆分

| 阶段 | 内容 |
|------|------|
| M13-A | 数据模型（pm_config_packages/versions）+ 状态机 + 配置包 CRUD + seed_yaml 解析 |
| M13-B | 发布同步 DB（SeedSyncService）+ 版本快照/回滚 |
| M13-C | 前端配置包列表 + 编辑页 + EPS 项目编辑修复 |
| M13-D | 端到端验证 + 清理旧 6 类 CRUD 实现 |

---

## 10. 风险与依赖

| 风险 | 缓解 |
|------|------|
| seed_yaml 结构与现有 schema/seed 格式差异 | SeedParser 单测 + 与 loader 结构对齐 |
| 发布同步覆盖已有运行数据 | 幂等 + 事务 + 版本快照可回滚 |
| EPS 树种子与业务种子 project_id 关联 | EPS 包先发布（先生成项目 ID），业务包引用 |
| 现有 3 项目业务种子迁移复杂 | 迁移为配置包 projects 段，端到端验证数据一致 |
