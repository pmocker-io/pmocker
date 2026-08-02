# M12 业务流闭环与数据联动 设计文档

> 版本：v1.0
> 日期：2026-08-02
> 状态：待评审（差距分析+方案架构，评审后进入 brainstorm 技术选型，再写详细计划）
> 对标：MS Project / 禅道 / 华天PLM 商业产品入门水平

***

## 1. 背景与问题陈述

### 1.1 当前状态

M1-M11 完成了 pmocker 的骨架：CLI、镜像、EAV 引擎、工作流引擎、10 模块 CRUD、前端表单适配、专业可视化（甘特图/检入检出/变更 diff/跨模块联动）。

但调研证实：**当前 10 模块是孤立 CRUD，零跨模块业务调用，专用表只有表结构无业务逻辑**。停留在 Demo 级别，远未达到商业项目管理软件入门水平。

### 1.2 用户给出的真实业务流

**单个项目全流程**：创建项目→添加团队→创建任务设前后置关系→指派责任人→评估工时成本→计划/成本/团队审批→自动生成计划基线/成本基线/资源负荷→项目执行（实际进展+工时+费用）→交付物上传审核归档→任务完成自动刷新完成%和偏差→记录问题/风险/变更/需求→重新更新计划→追踪管理→PDCA循环→定期报告→项目结束归档结项

**PMO 全流程**：授权立项→关注 EPS 整体进展→审核批准计划/成本/风险/交付物→查看运行报告→批准结项

### 1.3 目标

M12 让 10 模块从孤立 CRUD 升级为**真实业务流闭环**，达到商业产品入门水平，并通过 3 个项目全流程数据流转测试验证。

***

## 2. 差距分析报告

### 2.1 专用表激活状态（需求文档 7.3 节）

| 专用表                     | 用途               | 表结构   | 业务逻辑              | 差距                             |
| ----------------------- | ---------------- | ----- | ----------------- | ------------------------------ |
| `pm_wbs_nodes`          | WBS 物化路径树        | ✅ 已定义 | ⚠️ 部分使用（scope 模块） | 路径查询/权限继承未验证                   |
| `pm_eps_tree`           | EPS 物化路径树        | ✅ 已定义 | ⚠️ 部分使用（eps 模块）   | 权限继承未实现                        |
| `pm_task_links`         | 任务依赖 FS/SS/FF/SF | ✅ 已定义 | ❌ 未使用             | CPM 无真实依赖数据，甘特图依赖连线无数据源        |
| `pm_baselines`          | 基线快照             | ✅ 已定义 | ❌ 未使用             | 无快照生成、无偏差计算、无基线对比              |
| `pm_change_logs`        | 字段级变更日志          | ✅ 已定义 | ❌ 未使用             | 无 old/new\_value 记录、无审计追溯      |
| `pm_deliverable_files`  | 交付物文件元数据         | ✅ 已定义 | ⚠️ 部分使用           | Check-in/Out 用了 attrs 字段，未用专用表 |
| `pm_workflow_instances` | 工作流实例            | ✅ 已定义 | ⚠️ 部分使用           | 状态流转有，但不触发业务逻辑                 |

### 2.2 跨模块联动状态（需求文档 2.4 节 11 项能力）

| # | 能力           | M11 状态                             | M12 目标                            |
| - | ------------ | ---------------------------------- | --------------------------------- |
| ① | 任务-交付物绑定     | ⚠️ 前端有 deliverable\_id 字段，任务完成触发检入 | 强化：专用表关联 + 甘特图可视化 + 自动检入闭环        |
| ② | 文档生命周期+检入检出  | ✅ M11 已实现                          | 保持，迁移到 pm\_deliverable\_files 专用表 |
| ③ | ECN/ECR 变更闭环 | ⚠️ 有 diff UI，无变更执行跟踪               | 补全：变更执行→闭环验证→影响实体更新               |
| ④ | 电子签名+审签记录    | ❌ 未实现                              | 新增：审批操作记录操作者+时间戳+内容               |
| ⑤ | BOM 多视图      | v1 不要求                             | 跳过                                |
| ⑥ | 基于文档状态的权限    | ❌ 未实现                              | 新增：草稿/评审/发布/归档态权限控制               |
| ⑦ | 基线管理         | ❌ 未实现                              | **核心**：3 类基线快照+偏差+变更控制            |
| ⑧ | 工作流引擎        | ✅ 状态机有                             | **核心**：增加节点 hook 触发业务逻辑           |
| ⑨ | RACI 矩阵      | ✅ M11 已实现                          | 保持                                |
| ⑩ | 技能矩阵+投入度     | ⚠️ M11 有成本贡献列                      | 强化：工时登记→利用率→成本核算闭环                |
| ⑪ | 绩效闭环         | ✅ 已实现                              | 保持                                |

### 2.3 业务流闭环缺口（对照用户业务流）

| 业务流环节        | 当前                         | M12 目标                      |
| ------------ | -------------------------- | --------------------------- |
| 创建项目         | eps/tree 有 CRUD            | 增加项目创建向导（项目→团队→计划）          |
| 添加团队成员       | team\_member 无 project\_id | team\_member 关联 project\_id |
| 任务前后置关系      | pm\_task\_links 空壳         | 激活：FS/SS/FF/SF 编辑+CPM 真实计算  |
| 指派责任人        | task 有 owner\_id 字段        | 前端选择器+团队成员联动                |
| 评估工时成本       | 无联动                        | 任务工时×成员费率→成本预算              |
| 计划/成本/团队审批   | 工作流只转移状态                   | 审批通过→自动生成基线                 |
| 生成基线/资源负荷    | pm\_baselines 空壳           | 激活：快照生成+资源负荷计算              |
| 实际进展+工时+费用   | 无工时单/费用登记                  | 新增 pm\_time\_entries + 费用登记 |
| 交付物上传审核归档    | M11 有检入检出                  | 补全：上传→评审→发布→归档全生命周期         |
| 任务完成刷新完成%+偏差 | 无自动汇总                      | 事件引擎：任务完成→项目进度汇总+偏差         |
| 问题/风险/变更/需求  | 各模块孤立 CRUD                 | pm\_relations 关联+变更影响基线     |
| 重新更新计划       | 无计划修订                      | 基线变更走变更控制                   |
| 追踪管理 PDCA    | 无闭环                        | 状态流转闭环+再评估                  |
| 定期报告         | 无报告系统                      | 项目仪表盘+运行报告                  |
| 结项归档         | 无归档流程                      | 归档状态机+结项报告                  |

***

## 3. 方案架构

### 3.1 总体架构

```
┌─────────────────────────────────────────────────────────┐
│                    前端业务流层                          │
│  项目向导 │ 工时单 │ 审批中心 │ 仪表盘 │ PMO看板 │ 结项报告 │
├─────────────────────────────────────────────────────────┤
│                    业务事件引擎                          │
│  工作流节点Hook │ 状态流转触发器 │ 自动汇总计算器         │
├─────────────────────────────────────────────────────────┤
│                    跨模块联动层                          │
│  pm_relations │ pm_task_links │ 基线偏差 │ 成本聚合      │
├─────────────────────────────────────────────────────────┤
│                    EAV 数据层（已有）                     │
│  pm_entities │ pm_attrs │ pm_field_defs │ pm_entity_types│
├─────────────────────────────────────────────────────────┤
│                    专用表层（激活）                       │
│ pm_baselines │ pm_change_logs │ pm_time_entries(新)      │
│ pm_deliverable_files │ pm_workflow_instances             │
└─────────────────────────────────────────────────────────┘
```

### 3.2 pm\_entities 表字段增强（优先级）

在现有 `pm_entities` 表增加 `priority` 列（快速查询字段，不放入 attrs JSON），对标 MS Project 任务优先级、禅道项目/任务优先级：

```sql
ALTER TABLE pm_entities ADD COLUMN priority INT DEFAULT 2;
-- priority: 0=紧急(P0) 1=高(P1) 2=中(P2) 3=低(P3)
-- 项目(EPS node)和任务(task)均使用此字段
-- 高优先级(P0/P1)的项目和任务进入"我关注"子视图
```

| 优先级   | 数值 | 说明            | "我关注"可见性    |
| ----- | -- | ------------- | ----------- |
| P0 紧急 | 0  | 公司级关键项目/阻塞性任务 | 决策/审批人员默认可见 |
| P1 高  | 1  | 重要项目/关键路径任务   | 决策/审批人员默认可见 |
| P2 中  | 2  | 常规项目/普通任务     | 仅分配人可见      |
| P3 低  | 3  | 低优先级/可选任务     | 仅分配人可见      |

**"我关注"子视图可见性规则**：

* `PMO_ADMIN` 角色：可见所有 P0/P1 项目和任务

* `DEPT_LEADER` 岗位：可见本部门及子级下 P0/P1 项目和任务

* `CCB_MEMBER` 岗位：可见所有 P0/P1 变更相关任务

* 其他角色：仅可见自己负责的 P0/P1 任务

### 3.3 新增表设计

#### 3.3.1 pm\_time\_entries（工时登记表）— 新增

```sql
CREATE TABLE pm_time_entries (
  id            BIGINT PRIMARY KEY,
  project_id    BIGINT NOT NULL,          -- → pm_entities（EPS项目节点）
  task_id       BIGINT,                   -- → pm_entities（关联任务，可为空=项目级工时）
  member_id     BIGINT NOT NULL,          -- → pm_entities（team_member）
  user_id       BIGINT NOT NULL,          -- → sys_users（填报人）
  work_date     DATE NOT NULL,            -- 工作日期
  hours         DECIMAL(5,2) NOT NULL,    -- 工时（小时）
  cost          DECIMAL(12,2),            -- 成本（hours × member.hourly_rate，自动计算）
  description   TEXT,                     -- 工作内容描述
  status        VARCHAR(32) DEFAULT 'submitted', -- submitted/approved/rejected
  approved_by   BIGINT,                   -- 审批人 → sys_users
  approved_at   TIMESTAMP,
  created_at    TIMESTAMP,
  updated_at    TIMESTAMP
);
```

#### 3.3.2 pm\_cost\_actuals（实际成本执行表）— 新增

```sql
CREATE TABLE pm_cost_actuals (
  id            BIGINT PRIMARY KEY,
  project_id    BIGINT NOT NULL,
  task_id       BIGINT,                   -- 关联任务（可为空=项目级费用）
  cost_type     VARCHAR(32) NOT NULL,     -- labor/material/equipment/other
  amount        DECIMAL(12,2) NOT NULL,   -- 金额
  incurred_date DATE NOT NULL,            -- 发生日期
  description   TEXT,
  source        VARCHAR(32),              -- manual/time_entry/deliverable
  source_id     BIGINT,                   -- 来源实体ID（如 time_entry_id）
  created_at    TIMESTAMP
);
```

#### 3.3.3 pm\_approval\_records（审批签审记录表）— 新增

```sql
CREATE TABLE pm_approval_records (
  id            BIGINT PRIMARY KEY,
  workflow_instance_id BIGINT NOT NULL,   -- → pm_workflow_instances
  node_code     VARCHAR(64) NOT NULL,     -- 审批节点
  approver_id   BIGINT NOT NULL,          -- → sys_users
  action        VARCHAR(32) NOT NULL,     -- approve/reject/withdraw
  comment       TEXT,                     -- 审批意见
  signature     VARCHAR(128),             -- 电子签名（用户名+时间戳哈希）
  acted_at      TIMESTAMP NOT NULL
);
```

### 3.4 事件引擎架构

工作流节点 hook 机制：在现有 `WorkflowService.Execute` 基础上，增加节点进入/离开回调。

```go
// 节点事件钩子
type NodeHook interface {
    OnEnter(ctx context.Context, instance *WorkflowInstance, node *NodeDef) error
    OnLeave(ctx context.Context, instance *WorkflowInstance, node *NodeDef, action string) error
}

// 注册表（类似 AutoHandler）
hooks map[string]NodeHook  // key: workflow_code + "." + node_code

// 内置 hook 示例
- "plan_approval.approve".OnLeave → 生成计划基线快照
- "cost_approval.approve".OnLeave → 生成成本基线快照
- "task_done.complete".OnLeave → 刷新项目完成度+触发交付物检入
- "change_request.closed".OnLeave → 应用变更到目标实体+记录 change_logs
```

### 3.5 基线快照机制

```go
type Baseline struct {
    ID           uint
    ProjectID    uint
    Type         string   // scope/schedule/cost
    SnapshotJSON string   // 完整快照（所有相关实体的 attrs）
    CreatedBy    uint
    CreatedAt    time.Time
    ChangeReqID  uint     // 关联的变更请求（基线变更必须走变更控制）
}

// 生成快照
func CreateBaseline(ctx, projectID, baselineType, changeReqID) error {
    // 1. 查询项目下所有相关实体（task/cost_item/scope_item）
    // 2. 序列化 attrs 为 JSON 快照
    // 3. 写入 pm_baselines
}

// 偏差计算
func CalcVariance(ctx, projectID, baselineID) (*VarianceReport, error) {
    // 1. 加载基线快照
    // 2. 查询当前实际值
    // 3. 逐项对比：SV(进度偏差)=EV-PV, CV(成本偏差)=EV-AC
    // 4. 返回偏差报告
}
```

### 3.6 项目完成度汇总算法（3 种可选）

用户决策：3 种算法全部提供，下拉可选。3 个测试项目各覆盖 1 种。

```go
// 算法1: 工时加权平均（项目A使用，对标 MS Project）
func CalcByHours(ctx, projectID) (percent, error) {
    tasks := ListTasksByProject(projectID)
    totalHours, completedHours := 0, 0
    for _, t := range tasks {
        hours := t.attrs["estimated_hours"]
        progress := t.attrs["progress"] // 0-100
        totalHours += hours
        completedHours += hours * progress / 100
    }
    return completedHours / totalHours * 100
}

// 算法2: WBS 层级加权（项目B使用，对标 PMBOK）
func CalcByWBS(ctx, projectID) (percent, error) {
    // 自底向上：叶子节点进度→父节点加权平均→项目根节点
    wbsTree := LoadWBSTree(projectID)
    return calcNodeProgress(wbsTree.Root)
}

// 算法3: 任务数简单平均（项目C使用）
func CalcByCount(ctx, projectID) (percent, error) {
    tasks := ListTasksByProject(projectID)
    completed := countByStatus(tasks, "done")
    return completed / len(tasks) * 100
}

// 统一入口：按项目配置的算法类型计算
func CalcProjectProgress(ctx, projectID) (percent, error) {
    algo := GetProjectAlgoConfig(projectID) // hours/wbs/count
    switch algo {
    case "hours": return CalcByHours(ctx, projectID)
    case "wbs":   return CalcByWBS(ctx, projectID)
    case "count": return CalcByCount(ctx, projectID)
    }
}
```

### 3.7 模块数据流关系图

#### 3.7.1 10 模块业务关联矩阵

| 源模块                               | → 目标模块               | 关联字段/表                                            | 数据流说明               | 触发时机      |
| --------------------------------- | -------------------- | ------------------------------------------------- | ------------------- | --------- |
| EPS（项目）                           | 所有模块                 | `pm_entities.project_id`                          | 项目创建→所有实体归属项目       | 创建任何实体时   |
| Team（团队）                          | Schedule（进度）         | `task.owner_id` → `team_member.user_id`           | 团队成员指派为任务责任人        | 任务指派      |
| Team（团队）                          | Cost（成本）             | `team_member.hourly_rate` × `time_entry.hours`    | 时薪×工时→人工成本          | 工时登记审批    |
| Team（团队）                          | EPS（健康度）             | `utilization_rate` → 项目健康度                        | 利用率→项目资源健康度 RAG     | 利用率计算     |
| Schedule（进度）                      | Deliverable（交付物）     | `task.deliverable_id` + `pm_relations`            | 任务完成→自动检入关联交付物      | 任务状态→done |
| Schedule（进度）                      | Cost（成本）             | `task.estimated_hours` × `member.rate`            | 任务工时估算→成本预算         | 计划制定      |
| Schedule（进度）                      | Baseline             | `pm_baselines`(type=schedule)                     | 任务快照→计划基线           | 审批通过      |
| Schedule（进度）                      | EPS（项目）              | `CalcProjectProgress()`                           | 任务进度→项目完成%          | 任务完成      |
| Requirement（需求）                   | Schedule（任务）         | `pm_relations`(type=decomposes)                   | 需求→分解为任务            | 需求批准后     |
| Requirement（需求）                   | Change（变更）           | `pm_relations`(type=changes)                      | 需求变更→变更请求           | 需求变更时     |
| Scope（WBS）                        | Schedule（进度）         | `scope_item.id` ← `task.parent_id`                | WBS 节点→任务归属         | WBS 分解后   |
| Scope（WBS）                        | Deliverable          | `pm_relations`(type=delivers)                     | WBS 节点→交付物          | 交付物定义     |
| Scope（WBS）                        | Team（RACI）           | `scope_item.raci_*` → `team_role`                 | WBS×角色 RACI 矩阵      | RACI 配置   |
| Cost（成本）                          | Baseline             | `pm_baselines`(type=cost)                         | 成本快照→成本基线           | 审批通过      |
| Cost（成本）                          | Schedule             | `cost_item.task_id`                               | 成本项关联任务             | 成本登记      |
| Change（变更）                        | 任意实体                 | `change_request.target_entity` + `pm_change_logs` | 变更→应用到目标实体+记录日志     | 变更关闭      |
| Change（变更）                        | Baseline             | `pm_baselines.change_req_id`                      | 基线变更→新基线            | 变更批准      |
| Issue（问题）                         | Schedule/Deliverable | `pm_relations`(type=relates\_to)                  | 问题关联任务/交付物          | 问题创建      |
| Issue（问题）                         | Change（变更）           | `pm_relations`(type=triggers)                     | 问题触发变更              | 问题升级      |
| Risk（风险）                          | Schedule/Change      | `pm_relations`(type=impacts)                      | 风险关联任务/触发变更         | 风险识别      |
| Risk（风险）                          | EPS（健康度）             | 风险数+严重度→项目健康度                                     | 风险→项目风险健康度 RAG      | 风险评估      |
| Workflow                          | 所有模块                 | `NodeHook.OnLeave`                                | 审批通过→触发业务逻辑         | 工作流流转     |
| Schedule/Issue/Change/Deliverable | 个人任务中心(T15)          | `owner_id`/`assignee`/`reviewer` 聚合               | 各模块任务按当前用户聚合为个人待办   | 用户登录查看    |
| EPS（项目）                           | 项目工作台(T16)           | `creator_id`/`leader_id`/`team_member.user_id` 聚合 | 项目按用户角色（创建/负责/参与）聚合 | 用户登录查看    |

#### 3.6.2 数据流向图（核心业务流）

```
                        ┌─────────────┐
                        │  EPS 项目树  │
                        │ (project_id) │
                        └──────┬──────┘
                               │ 归属
          ┌────────────────────┼────────────────────┐
          ▼                    ▼                    ▼
   ┌────────────┐      ┌────────────┐      ┌────────────┐
   │ Team 团队   │      │ Scope WBS  │      │ Requirement│
   │ (member)   │      │ (wbs_node) │      │ (需求)      │
   └─────┬──────┘      └─────┬──────┘      └─────┬──────┘
         │ 指派                │ 分解               │ 分解
         ▼                    ▼                    ▼
   ┌─────────────────────────────────────────────────────┐
   │              Schedule 进度管理（task）                │
   │   owner_id ← Team    parent_id ← Scope   ← Requirement
   │   deliverable_id → Deliverable   estimated_hours → Cost
   └──────┬──────────────┬──────────────┬───────────────┘
          │ 依赖           │ 完成触发       │ 工时
          ▼              ▼              ▼
   ┌────────────┐  ┌────────────┐  ┌────────────┐
   │pm_task_links│  │Deliverable │  │Time Entries│
   │ (FS/SS/FF)  │  │ (检入检出)  │  │ (工时→成本) │
   └──────┬─────┘  └─────┬──────┘  └─────┬──────┘
          │ CPM          │ 归档           │ 转化
          ▼              ▼              ▼
   ┌─────────────────────────────────────────────┐
   │            Cost 成本管理                      │
   │  budget(预算) ← task.estimated_hours         │
   │  actual(实际) ← time_entries + cost_actuals  │
   │  EVM: PV/EV/AC/SPI/CPI                       │
   └──────────────────┬──────────────────────────┘
                      │ 快照
                      ▼
   ┌─────────────────────────────────────────────┐
   │           Baseline 基线管理                   │
   │  schedule_baseline / cost_baseline           │
   │  ← 审批通过自动生成（NodeHook）                │
   │  → 偏差计算（实际 vs 计划 vs 基线）            │
   └─────────────────────────────────────────────┘

   ┌──────────┐  ┌──────────┐  ┌──────────┐
   │ Issue 问题│  │ Risk 风险 │  │Change 变更│
   └─────┬────┘  └─────┬────┘  └─────┬────┘
         │             │             │
         └──────┬──────┴──────┬──────┘
                ▼             ▼
          pm_relations   pm_change_logs
          (实体关联)      (变更审计)
                │             │
                ▼             ▼
         关联到 task/     变更应用→
         deliverable     记录 old/new

┌─────────────────────────────────────────────────────┐
│              个人工作台（聚合视图层 T15/T16）           │
│                                                      │
│  个人任务中心(T15)              项目工作台(T16)        │
│  ├─ 项目任务(schedule,owner)   ├─ 我创建的项目        │
│  ├─ 问题任务(issue,assignee)   ├─ 我负责的项目(PM)    │
│  ├─ 变更任务(change,assignee)  └─ 我参与的项目(成员)  │
│  ├─ 交付物任务(deliverable,                          │
│  │  reviewer/checked_out_by)  按状态分组：            │
│  └─ 审批待办(workflow)        立项中/进行中/已归档/   │
│                                已暂停                 │
│  按状态分组：                                         │
│  待开始/进行中/已完成/已逾期                           │
│  统计：任务总数/完成率/逾期数  统计：项目数/健康度分布  │
└─────────────────────────────────────────────────────┘
```

### 3.8 项目管理业务模型（完整抽象）

#### 3.7.1 三层业务模型

```
┌─────────────────────────────────────────────────────────────┐
│ 第一层：组织架构层（gva 内置，复用不重建）                    │
│                                                              │
│  sys_departments（部门树）  sys_positions（岗位）             │
│       ↓ 多对多               ↓ 多对多                         │
│  sys_users（用户）─── sys_user_departments / positions        │
│       ↓ 多对多                                                │
│  sys_authorities（角色：PMO管理员/项目经理/团队成员/干系人）    │
│  sys_authority_departments（角色-部门数据权限范围）            │
│                                                              │
│  用途：审批流路由基于部门负责人/角色；数据权限基于部门范围      │
├─────────────────────────────────────────────────────────────┤
│ 第二层：项目管理层（EPS + 基线 + 审批）                       │
│                                                              │
│  pm_entities(entity_type=eps_node)  ← 项目节点，关联组织      │
│  pm_baselines                       ← 计划/成本/范围基线快照  │
│  pm_approval_records                ← 审批签审记录            │
│  pm_workflow_instances              ← 工作流实例              │
│                                                              │
│  用途：项目级数据隔离、基线管理、审批闭环                      │
├─────────────────────────────────────────────────────────────┤
│ 第三层：执行层（10 模块业务实体）                             │
│                                                              │
│  团队: team_member(→user_id,project_id)                      │
│  进度: task(→project_id,owner_id,deliverable_id)             │
│        pm_task_links(FS/SS/FF/SF 依赖)                       │
│  成本: cost_item(→project_id,task_id)                        │
│        pm_time_entries(工时→成本)                             │
│        pm_cost_actuals(实际费用)                              │
│  范围: scope_item(WBS树,→project_id)                         │
│  交付物: deliverable(→project_id,task_id)                    │
│          pm_deliverable_files(文件+检入检出)                  │
│  需求/风险/问题/变更: 各实体(→project_id)                     │
│  关联: pm_relations(实体间通用关联)                           │
│  审计: pm_change_logs(字段级变更日志)                         │
│                                                              │
│  用途：业务数据存储与跨模块联动                                │
├─────────────────────────────────────────────────────────────┤
│ 第四层：个人工作台聚合层（只读视图，不创建新数据）             │
│                                                              │
│  个人任务中心(T15): 聚合 task/issue/change/deliverable        │
│    ← owner_id/assignee/reviewer = 当前用户                   │
│    按状态分组：待开始/进行中/已完成/已逾期                     │
│                                                              │
│  项目工作台(T16): 聚合 eps_node                              │
│    ← creator_id/leader_id/team_member.user_id = 当前用户     │
│    分组：我创建的/我负责的/我参与的                           │
│    按状态分组：立项中/进行中/已归档/已暂停                     │
│                                                              │
│  用途：个人效率视角，"我要做什么"+"我负责什么"                │
└─────────────────────────────────────────────────────────────┘
```

#### 3.8.2 核心业务实体关系（ER 模型）

```
sys_users ──┬── sys_user_departments ── sys_departments（组织树）
             └── sys_user_positions ─── sys_positions（岗位）
                  │
                  │ user_id
                  ▼
pm_entities(eps_node) ──project_id──→ 所有业务实体
     │                                    │
     │ leader_id                          │ owner_id
     ▼                                    ▼
team_member ────────────────────→ task ────→ deliverable
     │                              │  ↑         │
     │ hourly_rate                  │  │deliverable_id
     ▼                              │  │         ▼
pm_time_entries ──→ cost ──────────┘  │  pm_deliverable_files
     │                  ↑              │
     │ hours×rate       │task_id       │ parent_id
     ▼                  │              ▼
pm_cost_actuals ──→ cost_item    pm_task_links
                                     │
                     pm_relations ───┤(通用关联)
                     (req↔task,      │
                      issue↔deliver, │
                      risk↔task,     ▼
                      change↔baseline) pm_baselines
                                        │ change_req_id
                                        ▼
                                  pm_change_logs
```

#### 3.7.3 审批流与组织架构的关联机制

审批流节点通过以下方式关联组织架构（gva 内置表）：

| 审批节点     | 审批人来源       | gva 表/字段                                                  | 示例        |
| -------- | ----------- | --------------------------------------------------------- | --------- |
| 项目经理审批   | 任务责任人或项目负责人 | `sys_users`（通过 team\_member.user\_id）                     | 项目A 张明审批  |
| 部门负责人审批  | 项目所属部门的负责人  | `sys_departments.leader_id`                               | 研发部部长审批   |
| PMO 审批   | PMO 管理员角色   | `sys_authorities`(authorityId=PMO) + `sys_user_authority` | PMO 管理员审批 |
| CCB 变更审批 | CCB 成员角色组   | `sys_authorities` + 部门范围                                  | 变更控制委员会审批 |
| 交付物评审    | 评审人角色（按部门）  | `sys_departments` + `sys_positions`(评审岗)                  | 工艺评审组审批   |

**审批人解析逻辑**：

```
1. 角色路由：approvers="role:pmo_admin" → 查 sys_user_authority 找该角色用户
2. 部门路由：approvers="dept_leader:{project.dept_id}" → 查 sys_departments.leader_id
3. 岗位路由：approvers="position:tech_lead" → 查 sys_user_positions 找该岗位用户
4. 组合路由：approvers="role:pm AND dept:{project.dept_id}" → 角色与部门交集
```

***

## 4. M12 五阶段任务分解

### Phase 1 - 数据骨架激活

| Task | 内容                          | 关键交付                                                                                                                       |
| ---- | --------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| T1   | 组织架构初始化 + 项目维度数据隔离 + 项目创建向导 | 初始化 sys\_departments/positions/users（3 组织 17 用户）；team\_member 关联 project\_id+user\_id；创建项目向导（项目→团队→计划三步）；EPS 节点关联 dept\_id |
| T2   | 激活 pm\_relations（实体间关联）     | 关联 CRUD API + 前端关联选择器（需求↔任务、问题↔交付物、变更↔基线）                                                                                  |
| T3   | 激活 pm\_task\_links（任务依赖）    | 依赖编辑 API + 甘特图依赖连线 + CPM 基于真实依赖计算                                                                                          |
| T4   | 激活 pm\_change\_logs（变更审计）   | 字段级 old/new\_value 自动记录 + 审计追溯 UI                                                                                          |

### Phase 2 - 计划-成本-资源联动

| Task | 内容                      | 关键交付                         |
| ---- | ----------------------- | ---------------------------- |
| T5   | 任务指派 + 工时估算 + 成本预算联动    | 任务指派团队成员→工时估算×费率→成本预算自动生成    |
| T6   | 工时登记（pm\_time\_entries） | 工时单 CRUD + 审批 + 利用率计算        |
| T7   | 成本执行（pm\_cost\_actuals） | 实际费用登记 + 工时→成本自动转化 + EVM 实际值 |

### Phase 3 - 基线与偏差

| Task | 内容      | 关键交付                                |
| ---- | ------- | ----------------------------------- |
| T8   | 基线快照管理  | 审批通过自动生成计划/成本/范围基线；基线列表+对比          |
| T9   | 偏差分析与预警 | 实际vs计划vs基线偏差图表；SPI/CPI/进度偏差；超期/超支预警 |

### Phase 4 - 业务事件引擎

| Task | 内容         | 关键交付                                |
| ---- | ---------- | ----------------------------------- |
| T10  | 工作流节点 hook | NodeHook 接口+注册表；审批通过→生成基线；变更关闭→应用变更 |
| T11  | 项目完成度自动汇总  | 任务完成→刷新项目进度；红黄绿灯健康度                 |

### Phase 5 - 报告与结项

| Task | 内容           | 关键交付                                |
| ---- | ------------ | ----------------------------------- |
| T12  | 项目仪表盘 + 运行报告 | 项目概览页（进度/成本/风险/问题/资源汇总）；定期报告快照      |
| T13  | EPS PMO 看板   | 跨项目组合视图；项目健康度 RAG；资源负荷汇总            |
| T14  | 结项归档流程       | 归档状态机；结项报告生成（任务/问题/风险/需求/资源/变更汇总统计） |

### Phase 6 - 个人工作台（任务中心+项目工作台）

| Task | 内容           | 关键交付                                                                                                    |
| ---- | ------------ | ------------------------------------------------------------------------------------------------------- |
| T15  | 任务管理（个人任务中心） | 聚合项目任务/问题任务/变更任务/交付物任务；按状态分组（待开始/进行中/已完成/已逾期）；**含"我关注的任务"子视图（高优先级任务，决策/审批人员默认可见）**；任务处理进度+统计数据          |
| T16  | 项目管理（项目工作台）  | 我创建的/我负责的/我参与的项目；**含"我关注的项目"子视图（高优先级项目，决策/审批人员默认可见）**；按状态分组（立项中/进行中/已归档/已暂停）；项目卡片（健康度/进度%/成本偏差/风险数/优先级） |

***

## 5. 期初数据规划（种子数据）

### 5.0 组织架构设计（gva 内置表，审批流基础）

种子数据必须先初始化 gva 内置组织架构表（`sys_departments` / `sys_positions` / `sys_users` / `sys_user_departments` / `sys_user_positions` / `sys_authorities`），3 个项目对应 3 个独立组织（顶级部门），每个组织下设子部门，用户分配到部门+岗位+角色，审批流基于此路由。

#### 5.0.1 部门树（sys\_departments，3 组织 × 子部门）

```
0. 集团总部（顶级，PMO 所在）
├── 1. 智能排产系统研发部（项目A 组织）
│   ├── 1.1 项目管理组
│   ├── 1.2 前端开发组
│   ├── 1.3 后端开发组
│   └── 1.4 质量测试组
├── 2. 工程建设事业部（项目B 组织）
│   ├── 2.1 项目管理部
│   ├── 2.2 土建工程部
│   ├── 2.3 机电工程部
│   └── 2.4 安全造价部
└── 3. 传感器研发中心（项目C 组织）
    ├── 3.1 项目管理组
    ├── 3.2 结构设计组
    ├── 3.3 电子设计组
    └── 3.4 工艺测试组
```

#### 5.0.2 岗位定义（sys\_positions）

| 岗位编码         | 岗位名称    | 所属组织 | 备注         |
| ------------ | ------- | ---- | ---------- |
| PM           | 项目经理    | 通用   | 各组织项目管理组/部 |
| BA           | 业务分析师   | 研发   | 需求分析       |
| FE\_DEV      | 前端开发工程师 | 研发   | 前端开发组      |
| BE\_DEV      | 后端开发工程师 | 研发   | 后端开发组      |
| QA           | 测试工程师   | 研发   | 质量测试组      |
| CIVIL\_ENG   | 土建工程师   | 工程   | 土建工程部      |
| MEP\_ENG     | 机电工程师   | 工程   | 机电工程部      |
| SAFETY       | 安全员     | 工程   | 安全造价部      |
| QS           | 造价师     | 工程   | 安全造价部      |
| STRUCT\_ENG  | 结构工程师   | 研制   | 结构设计组      |
| ELEC\_ENG    | 电子工程师   | 研制   | 电子设计组      |
| PROCESS\_ENG | 工艺工程师   | 研制   | 工艺测试组      |
| TEST\_ENG    | 测试工程师   | 研制   | 工艺测试组      |
| PMO\_ADMIN   | PMO 管理员 | 集团   | 集团总部，跨项目   |
| DEPT\_LEADER | 部门负责人   | 通用   | 各子部门负责人    |
| CCB\_MEMBER  | CCB 成员  | 通用   | 变更控制委员会    |

#### 5.0.3 用户清单（sys\_users，3 项目 × 5 人 + PMO 2 人 = 17 人）

| 用户名              | 昵称    | 主部门             | 岗位           | 角色     | 所属项目 |
| ---------------- | ----- | --------------- | ------------ | ------ | ---- |
| admin            | 超级管理员 | 集团总部            | —            | 超级管理员  | 全局   |
| pmo01            | 李PMO  | 集团总部            | PMO\_ADMIN   | PMO管理员 | 跨项目  |
| pmo02            | 王PMO  | 集团总部            | PMO\_ADMIN   | PMO管理员 | 跨项目  |
| proj\_a\_pm      | 张明    | 1.智能排产研发部/项目管理组 | PM           | 项目经理   | 项目A  |
| proj\_a\_ba      | 李娜    | 1.智能排产研发部/项目管理组 | BA           | 团队成员   | 项目A  |
| proj\_a\_fe      | 王强    | 1.智能排产研发部/前端开发组 | FE\_DEV      | 团队成员   | 项目A  |
| proj\_a\_be      | 刘洋    | 1.智能排产研发部/后端开发组 | BE\_DEV      | 团队成员   | 项目A  |
| proj\_a\_qa      | 陈静    | 1.智能排产研发部/质量测试组 | QA           | 团队成员   | 项目A  |
| proj\_b\_pm      | 赵刚    | 2.工程建设事业部/项目管理部 | PM           | 项目经理   | 项目B  |
| proj\_b\_civil   | 钱伟    | 2.工程建设事业部/土建工程部 | CIVIL\_ENG   | 团队成员   | 项目B  |
| proj\_b\_mep     | 孙磊    | 2.工程建设事业部/机电工程部 | MEP\_ENG     | 团队成员   | 项目B  |
| proj\_b\_safety  | 周梅    | 2.工程建设事业部/安全造价部 | SAFETY       | 团队成员   | 项目B  |
| proj\_b\_qs      | 吴芳    | 2.工程建设事业部/安全造价部 | QS           | 团队成员   | 项目B  |
| proj\_c\_pm      | 郑辉    | 3.传感器研发中心/项目管理组 | PM           | 项目经理   | 项目C  |
| proj\_c\_struct  | 冯雪    | 3.传感器研发中心/结构设计组 | STRUCT\_ENG  | 团队成员   | 项目C  |
| proj\_c\_elec    | 褚晗    | 3.传感器研发中心/电子设计组 | ELEC\_ENG    | 团队成员   | 项目C  |
| proj\_c\_process | 卫鹏    | 3.传感器研发中心/工艺测试组 | PROCESS\_ENG | 团队成员   | 项目C  |
| proj\_c\_test    | 蒋琳    | 3.传感器研发中心/工艺测试组 | TEST\_ENG    | 团队成员   | 项目C  |

> 每个用户的 `sys_users.dept_id` 设为主部门，`sys_user_departments` 维护多部门归属，`sys_user_positions` 维护岗位，`sys_user_authority` 维护角色。密码统一初始化，`must_change_password=true`。

#### 5.0.4 角色（sys\_authorities，4 内置角色）

| 角色ID       | 角色名     | 数据范围    | 说明               |
| ---------- | ------- | ------- | ---------------- |
| 888        | 超级管理员   | 1全部     | admin            |
| PMO\_ADMIN | PMO 管理员 | 1全部     | 跨项目管理、立项/结项审批    |
| PM         | 项目经理    | 2本部门及子级 | 管理本项目，计划/成本/团队审批 |
| TEAM       | 团队成员    | 3本部门    | 执行任务、登记工时        |
| VIEWER     | 干系人     | 4仅本人    | 只读查看             |

#### 5.0.5 审批流路由示例（基于组织架构）

| 审批节点      | 审批人解析                                            | 对应用户          |
| --------- | ------------------------------------------------ | ------------- |
| 项目A 计划审批  | `role:PM AND dept:1.*` → 项目A 的项目经理               | 张明            |
| 项目A 成本审批  | `position:DEPT_LEADER AND dept:1` → 研发部负责人       | （设张明兼任或另设负责人） |
| 变更 CCB 审批 | `position:CCB_MEMBER` → CCB 成员组                  | PMO + 各 PM    |
| 项目A 结项审批  | `role:PMO_ADMIN` → PMO 管理员                       | 李PMO/王PMO     |
| 交付物工艺评审   | `position:PROCESS_ENG AND dept:3.*` → 传感器工艺工程师   | 卫鹏            |
| 工时审批      | `role:PM AND dept:{member.dept_id}` → 成员所属项目的 PM | 对应项目 PM       |

### 5.1 三个测试项目场景

#### 项目 A：智能排产系统研发（软件研发，敏捷混合）

* **健康度**：绿色（正常）

* **优先级**：P1（高）— 重要项目，进入"我关注"

* **周期**：2026-01-06 \~ 2026-06-30（6个月）

* **团队角色（5+，引用 sys\_users）**：PM 张明(proj\_a\_pm)、BA 李娜(proj\_a\_ba)、前端Dev 王强(proj\_a\_fe)、后端Dev 刘洋(proj\_a\_be)、QA 陈静(proj\_a\_qa)

* **任务（15+）**：需求调研→架构设计→DB设计→前端框架搭建→后端框架搭建→用户管理→排产算法→可视化看板→集成测试→性能测试→UAT→部署上线→培训→文档交付→结项

  * 依赖关系：架构设计→DB设计/前端框架/后端框架；后端框架→排产算法→用户管理；前端框架→可视化看板；集成测试←前端+后端；部署上线←UAT←集成测试

  * 关键路径：需求调研→架构设计→后端框架→排产算法→集成测试→UAT→部署上线

* **成本（10+）**：5人×6月薪资（PM 25k/月、BA 18k、前端 20k、后端 22k、QA 16k）+ 服务器 2万 + 工具许可 1.5万 + 培训 0.5万

* **基线（5+）**：初始计划基线、初始成本基线、范围基线、修订计划基线（需求变更后）、修订成本基线

* **需求（10+）**：排产算法、可视化看板、用户管理、权限控制、数据导入、报表导出、告警通知、移动端适配、API开放、多工厂支持

* **问题（10+）**：算法性能瓶颈、前端内存泄漏、DB死锁、接口超时、兼容性问题、数据一致性、权限遗漏、报表慢查询、告警误报、移动端布局错乱

* **风险（10+）**：算法精度不达标、核心人员离职、需求蔓延、技术栈升级、第三方依赖、数据迁移、性能压力、安全漏洞、进度延期、成本超支

* **变更（10+）**：排产规则变更、UI改版、增加移动端、DB结构调整、API重构、权限模型调整、报表引擎替换、告警规则修改、部署方案变更、验收标准调整

* **交付物（20+）**：需求规格说明书、架构设计文档、DB设计文档、API文档、前端代码、后端代码、排产算法库、测试计划、测试用例、测试报告、UAT报告、部署手册、用户手册、培训材料、运维手册、源码包、安装包、\_release notes、项目总结、结项报告

#### 项目 B：新建厂房工程（工程建设，瀑布）

* **健康度**：红色（超支延期）

* **周期**：2025-10-01 \~ 2026-09-30（12个月，已延期2个月）

* **团队角色（5+，引用 sys\_users）**：项目经理 赵刚(proj\_b\_pm)、土建工程师 钱伟(proj\_b\_civil)、机电工程师 孙磊(proj\_b\_mep)、安全员 周梅(proj\_b\_safety)、造价师 吴芳(proj\_b\_qs)

* **任务（15+）**：地质勘察→方案设计→施工图设计→报建审批→基础施工→主体结构→二次结构→机电安装→装修工程→室外工程→消防验收→竣工验收→设备调试→搬迁入驻→结项

  * 关键路径长，主体结构延期导致整体延期

* **成本（10+）**：材料费（混凝土500万、钢筋800万、砖石200万、装修材料300万）、人工费、设备租赁、设计费、监理费、检测费、安全措施费、管理费

* **基线偏差大**：计划成本 3500万 vs 实际 4200万（超支 20%），进度延期 2 个月

#### 项目 C：新型传感器研制（PLM 硬件研发）

* **健康度**：黄色（有变更，可控）

* **优先级**：P1（高）— 重要研制项目，进入"我关注"

* **周期**：2026-02-01 \~ 2026-08-31（7个月）

* **团队角色（5+，引用 sys\_users）**：PM 郑辉(proj\_c\_pm)、结构工程师 冯雪(proj\_c\_struct)、电子工程师 褚晗(proj\_c\_elec)、工艺工程师 卫鹏(proj\_c\_process)、测试工程师 蒋琳(proj\_c\_test)

* **任务（15+）**：需求分析→总体方案→电路设计→PCB设计→结构设计→元器件采购→PCB打样→样机焊接→结构加工→样机装配→功能测试→环境测试→可靠性测试→小批量试产→设计定型

  * 电路设计变更导致 PCB 返工

* **交付物（20+）**：需求规格书、总体方案书、电路原理图、PCB Gerber、结构3D图、BOM表、工艺路线、测试大纲、功能测试报告、环境测试报告、可靠性测试报告、试产总结、定型报告、设计文件包、图纸归档

* **ECN/ECR 变更多**：电路改版 ECN、结构优化 ECN、BOM 变更 ECR、工艺调整 ECR

### 5.2 数据关联矩阵

```
项目A（研发）
├── team_member(5) ──→ task.owner_id(15) ──→ time_entry(60+) ──→ cost_actual(labor)
├── requirement(10) ──relation──→ task(15) ──relation──→ deliverable(20)
├── risk(10) ──relation──→ task / change_request
├── issue(10) ──relation──→ task / deliverable
├── change_request(10) ──→ baseline变更 ──→ change_logs
├── baseline(5): 计划×2 + 成本×2 + 范围×1
├── task_links: 15+ 依赖关系（FS为主，含1个SS）
├── approval_records: 计划审批/成本审批/变更审批/交付物审批
├── dashboard_report(2): 月度报告×2
└── 结项报告(1) + 归档

项目B（工程）— 同结构，数据体现超支延期
项目C（研制）— 同结构，数据体现ECN变更
```

### 5.3 数据量汇总

| 数据类型    | 项目A | 项目B | 项目C | 合计  | 最低要求      |
| ------- | --- | --- | --- | --- | --------- |
| 任务      | 15  | 16  | 15  | 46  | ≥45（15×3） |
| 成本数据    | 12  | 14  | 11  | 37  | ≥30（10×3） |
| 基线      | 5   | 6   | 5   | 16  | ≥15（5×3）  |
| 团队角色    | 5   | 5   | 5   | 15  | ≥15（5×3）  |
| 问题      | 12  | 11  | 10  | 33  | ≥30（10×3） |
| 风险      | 10  | 11  | 10  | 31  | ≥30（10×3） |
| 变更      | 10  | 11  | 10  | 31  | ≥30（10×3） |
| 需求      | 10  | 10  | 10  | 30  | ≥30（10×3） |
| 交付物     | 20  | 21  | 20  | 61  | ≥60（20×3） |
| 工时登记    | 60  | 70  | 55  | 185 | —         |
| 审批记录    | 8   | 10  | 9   | 27  | 若干        |
| 项目仪表盘报告 | 2   | 2   | 2   | 6   | ≥6（2×3）   |
| PMO看板报告 | 1   | 1   | 1   | 3   | ≥3（1×3）   |
| 归档      | 是   | 是   | 是   | 3   | 全部归档      |

### 5.4 种子数据生成方式

新建 `gva/server/plugin/pmocker_core/seed/business_flow_seed.go`，通过 `InitPMocker()` 钩子灌入，**按依赖顺序初始化**：

```
1. 组织架构（gva 内置表，最先初始化）
   → sys_authorities（4 角色）
   → sys_positions（16 岗位）
   → sys_departments（集团总部 + 3 组织 × 4 子部门 = 13 部门）
   → sys_users（17 用户，含密码哈希）
   → sys_user_authority / sys_user_departments / sys_user_positions（关联）
   → sys_authority_departments（角色-部门数据权限）

2. 项目层（EPS）
   → pm_entities(eps_node) × 3，每个关联主部门 dept_id

3. 执行层（10 模块业务实体）
   → team_member × 15（关联 project_id + sys_user_id）
   → scope_item(WBS) × N
   → requirement × 30 + pm_relations(需求↔任务)
   → task × 46 + pm_task_links(依赖) + owner_id 指派
   → cost_item × 37 + 成本预算(工时×费率)
   → deliverable × 61 + pm_deliverable_files
   → issue × 33 + pm_relations(问题↔任务/交付物)
   → risk × 31 + pm_relations(风险↔任务)
   → change_request × 31 + pm_change_logs

4. 基线与审批
   → pm_workflow_instances(模拟流转)
   → pm_approval_records(审批签审，基于组织架构路由)
   → pm_baselines × 16(审批通过自动生成)
   → pm_time_entries × 185 + pm_cost_actuals

5. 报告与归档
   → 项目仪表盘快照 × 6
   → PMO 看板快照 × 3
   → 3 项目归档 + 结项报告 × 3
```

数据时间线符合真实场景（项目B 已进入执行后期，项目C 中期，项目A 后期）。

***

## 6. 全流程数据流转测试方案

### 6.1 测试路径（单项目全流程）

以**项目A 智能排产系统研发**为主测试路径：

```
Step 1: 创建项目A（项目向导）
  → EPS 节点创建 + 项目基本信息

Step 2: 添加团队（5角色）
  → team_member × 5，关联 project_id

Step 3: 创建计划任务（15+任务，设前后置关系）
  → task × 15 + pm_task_links（依赖关系）
  → CPM 计算关键路径

Step 4: 指派责任人 + 工时估算
  → task.owner_id = team_member.user_id
  → task.estimated_hours 填写
  → 成本预算自动生成（工时×费率）

Step 5: 提交计划/成本/团队审批
  → 工作流启动
  → 审批通过 → 事件引擎触发 → 生成计划基线+成本基线

Step 6: 项目执行
  → 任务实际开始/完成时间填写
  → 工时登记（pm_time_entries）→ 审批 → 成本执行
  → 费用登记（pm_cost_actuals）
  → 交付物上传 → 评审 → 发布 → 归档

Step 7: 任务完成 → 自动汇总
  → 项目完成% 刷新
  → 偏差计算（实际vs计划vs基线）
  → 关联交付物自动检入

Step 8: 记录问题/风险/变更/需求
  → 创建 issue/risk/change/requirement
  → pm_relations 关联到 task/deliverable
  → 变更审批通过 → 应用变更 → 记录 change_logs
  → 基线变更 → 生成新基线

Step 9: 重新更新计划
  → 任务调整 → 新基线
  → 偏差重新计算

Step 10: 追踪管理（PDCA）
  → 问题解决→验证→关闭
  → 风险应对→跟踪→再评估
  → 状态流转闭环

Step 11: 定期报告
  → 项目仪表盘快照 × 2

Step 12: PMO 看板
  → EPS 跨项目汇总

Step 13: 结项归档
  → 归档状态机
  → 结项报告生成（汇总统计）
  → 项目状态→archived

Step 14: 个人任务中心验证（以 proj_a_be 刘洋登录）
  → 查看个人待办：聚合项目任务(owner) + 问题任务(assignee) + 变更任务 + 交付物任务(reviewer)
  → 按状态分组：待开始/进行中/已完成/已逾期
  → 统计：任务总数/完成率/逾期数

Step 15: 项目工作台验证（分别以 PM/成员/PMO 登录）
  → PM 张明：查看"我负责的项目"=项目A
  → 成员 刘洋：查看"我参与的项目"=项目A
  → PMO 李PMO：查看"我创建的项目"=3个项目
  → 按状态分组：进行中/已归档（项目A/B/C 已归档）
  → 项目卡片：健康度/进度%/成本偏差/风险数
```

### 6.2 测试路径（PMO 全流程）

```
Step 1: PMO 授权立项（3个项目）
Step 2: 持续关注 EPS 整体进展（PMO 看板）
Step 3: 审核批准项目计划/成本/风险/交付物
Step 4: 查看运行报告（各项目仪表盘）
Step 5: 批准项目结项（3个项目全部归档）
```

### 6.3 自动化测试策略

* **数据完整性测试**：验证所有实体关联（pm\_relations/pm\_task\_links）完整

* **业务事件测试**：验证审批通过→基线生成、任务完成→进度汇总

* **偏差计算测试**：验证 SV/CV/SPI/CPI 计算正确性

* **报告测试**：验证仪表盘数据与底层实体一致

* **归档测试**：验证归档后数据只读、结项报告完整

***

## 7. 验收标准

### 7.1 功能验收

| #  | 验收项     | 标准                                                                                |
| -- | ------- | --------------------------------------------------------------------------------- |
| 1  | 项目维度隔离  | 所有实体含 project\_id，前端按项目筛选                                                         |
| 2  | 项目创建向导  | 三步向导（项目→团队→计划）可走通                                                                 |
| 3  | 实体间关联   | pm\_relations CRUD 可用，前端关联选择器可用                                                   |
| 4  | 任务依赖    | pm\_task\_links CRUD 可用，甘特图显示依赖连线，CPM 基于真实依赖                                      |
| 5  | 变更审计    | pm\_change\_logs 自动记录字段级 old/new\_value                                           |
| 6  | 任务指派联动  | 指派成员→工时估算→成本预算自动生成                                                                |
| 7  | 工时登记    | 工时单 CRUD+审批，利用率自动计算                                                               |
| 8  | 成本执行    | 费用登记+工时→成本转化，EVM 实际值正确                                                            |
| 9  | 基线快照    | 审批通过自动生成基线，基线列表+对比可用                                                              |
| 10 | 偏差分析    | 实际vs计划vs基线偏差图表，SPI/CPI 正确                                                         |
| 11 | 事件引擎    | 审批→基线、任务完成→进度汇总+交付物检入                                                             |
| 12 | 项目完成度   | 任务完成→项目进度自动刷新，红黄绿灯                                                                |
| 13 | 项目仪表盘   | 进度/成本/风险/问题/资源汇总页                                                                 |
| 14 | PMO 看板  | 跨项目组合视图，健康度 RAG                                                                   |
| 15 | 结项归档    | 归档状态机，结项报告生成                                                                      |
| 16 | 个人任务中心  | 聚合项目/问题/变更/交付物任务，按状态分组（待开始/进行中/已完成/已逾期），统计任务总数/完成率/逾期数                            |
| 17 | 项目工作台   | 我创建的/我负责的/我参与的项目，按状态分组（立项中/进行中/已归档/已暂停），项目卡片含健康度/进度%/成本偏差/风险数/优先级                 |
| 18 | 优先级字段   | 项目和任务均含 priority 字段（P0紧急/P1高/P2中/P3低），pm\_entities 表独立列存储                         |
| 19 | 我关注的子视图 | 任务中心和项目工作台含"我关注的项目/任务"子视图，展示 P0/P1 高优先级项；PMO\_ADMIN/DEPT\_LEADER/CCB\_MEMBER 默认可见 |

### 7.2 数据验收（3 项目全流程）

| #  | 验收项      | 最低标准                           |
| -- | -------- | ------------------------------ |
| 1  | 项目数      | ≥ 3                            |
| 2  | 每项目任务数   | ≥ 15                           |
| 3  | 每项目成本数据  | ≥ 10                           |
| 4  | 计划+成本基线  | ≥ 5                            |
| 5  | 每项目团队角色  | ≥ 5                            |
| 6  | 每项目问题    | ≥ 10                           |
| 7  | 每项目风险    | ≥ 10                           |
| 8  | 每项目变更    | ≥ 10                           |
| 9  | 每项目需求    | ≥ 10                           |
| 10 | 每项目交付物   | ≥ 20                           |
| 11 | 审批流      | 若干（含计划/成本/变更/交付物审批）            |
| 12 | 项目仪表盘报告  | 每项目 ≥ 2                        |
| 13 | PMO 看板报告 | 每项目 ≥ 1                        |
| 14 | 归档       | 3 个项目全部归档                      |
| 15 | 数据真实性    | 时间/成本/人员负荷符合真实场景               |
| 16 | 组织架构     | 3 个独立组织（部门树），每个 ≥4 子部门，共 13 部门 |
| 17 | 系统用户     | ≥17 个 sys\_users 用户，关联部门+岗位+角色 |
| 18 | 岗位定义     | ≥16 个 sys\_positions 岗位        |
| 19 | 审批路由     | 审批人基于角色/部门/岗位真实解析，非硬编码         |
| 20 | 数据权限     | PM 角色仅见本部门及子级数据，PMO 见全部        |

### 7.3 业务流验收

| # | 验收项     | 标准                                                    |
| - | ------- | ----------------------------------------------------- |
| 1 | 单项目全流程  | 创建→团队→计划→审批→基线→执行→交付物→完成汇总→问题/风险/变更→更新计划→追踪→报告→归档 全闭环 |
| 2 | PMO 全流程 | 立项→EPS监控→审批→报告→结项 全闭环                                 |
| 3 | 跨模块联动   | 任务-交付物、任务-成本、变更-基线、问题-任务 关联真实生效                       |
| 4 | 事件触发    | 审批通过→基线、任务完成→汇总 自动触发无需人工                              |
| 5 | 偏差预警    | 超期/超支自动标红预警                                           |

***

## 8. 技术选型决策（已确认）

| # | 技术点                | 决策                        | 理由/行业对标                                                                              |
| - | ------------------ | ------------------------- | ------------------------------------------------------------------------------------ |
| 1 | 工时表设计              | **独立表 pm\_time\_entries** | 对标禅道工时单、MS Project 资源工作表；独立表支持高效 SUM/GROUP BY 聚合、独立审批流、直接成本计算                        |
| 2 | 事件引擎架构             | **NodeHook 同步回调**         | 对比观察者/事件总线，同步回调事务内完成、简单可控、与现有 AutoHandler 一致；耗时操作可后续异步化                              |
| 3 | 基线快照机制             | **全量 JSON 快照**            | 对标 MS Project 基线（Baseline1-11 存完整快照）；种子数据最大 46 任务，JSON 可控；对比直接                       |
| 4 | 项目完成度算法            | **3 种算法全部提供，下拉可选**        | 对标 MS Project（工时加权）+ PMBOK（WBS 层级）+ 简单平均；3 个测试项目各覆盖 1 种：项目A=工时加权、项目B=WBS层级、项目C=任务数平均 |
| 5 | pm\_relations 通用关联 | **统一 pm\_relations 表**    | 对标 PLM 的关联关系表（如华天 PLM link 表）；统一表灵活可扩展任意关联类型，relation\_type 区分                       |
| 6 | 种子数据生成             | **混合：Go 基础+YAML 业务**      | 组织架构/角色用 Go（类型安全）；业务数据用 YAML（数据与代码分离，非开发可编辑）；Go 解析 YAML 加载                           |
| 7 | 报告快照               | **混合：实时+里程碑快照**           | 对标禅道仪表盘（实时）+ MS Project 基准报告（快照）；日常实时查询，月报/结项生成快照存档                                  |
| 8 | 归档机制               | **状态标记+软删除**              | 对标禅道/华天 PLM 归档实践；项目 status=archived，关联实体软删除标记，只读不可编辑，全局过滤                            |

### 8.1 行业最佳实践对标（表设计）

| 表                      | 行业对标                             | 关键设计点                                                                       |
| ---------------------- | -------------------------------- | --------------------------------------------------------------------------- |
| pm\_time\_entries      | 禅道工时单、MS Project 资源工作表           | 按 date+task+member 维度登记，审批状态流转，cost 自动计算                                    |
| pm\_cost\_actuals      | MS Project 实际成本、PMBOK 挣值管理       | cost\_type 分类（labor/material/equipment），source 追溯（manual/time\_entry）       |
| pm\_baselines          | MS Project Baseline1-11、PMBOK 基准 | type 区分（schedule/cost/scope），snapshot\_json 全量快照，change\_req\_id 关联变更       |
| pm\_task\_links        | MS Project 前置任务、PMP 依赖关系         | link\_type（FS/SS/FF/SF），lag 滞后量，CPM 基于 FS 计算                                |
| pm\_relations          | 华天 PLM link 表、PMBOK 追溯矩阵         | relation\_type 区分（decomposes/relates\_to/triggers/impacts/delivers/changes） |
| pm\_change\_logs       | 华天 PLM ECN 审计、ITIL 变更管理          | entity\_id+field\_key+old\_value+new\_value，change\_req\_id 关联              |
| pm\_approval\_records  | 禅道审批记录、ISO 9001 审签记录             | approver\_id+action+comment+signature（电子签名）+acted\_at                       |
| pm\_deliverable\_files | 华天 PLM 文档管理、PMBOS 交付物            | lock\_status+checked\_out\_by+checked\_out\_at，版本控制                         |
| sys\_departments       | gva 内置（对标钉钉/企微组织架构）              | ParentId+Ancestors 物化路径，LeaderId 负责人                                        |
| sys\_positions         | gva 内置（对标岗位职级体系）                 | Code 编码，与角色正交                                                               |

***

## 9. 风险与依赖

| 风险            | 影响       | 缓解                               |
| ------------- | -------- | -------------------------------- |
| 工程量大（14 Task） | 周期长      | 分 Phase 提交，每 Phase 可独立验证         |
| 事件引擎改造工作流     | 可能影响已有功能 | NodeHook 可选注册，不影响已有 auto handler |
| 种子数据量大        | 灌入慢/调试难  | 分项目生成，支持单独重建                     |
| 基线快照 JSON 过大  | 性能问题     | 仅快照关键 attrs，非全量                  |
| 前端改造面广        | UI 回归风险  | 复用 DynamicForm，新增组件独立            |

***

## 10. 下一步

1. **本文档评审**：用户审阅差距分析+方案架构+期初数据+测试方案+验收标准
2. **Brainstorm 技术选型**：针对第 8 节 8 个技术点深入讨论
3. **编写详细计划文档**：task-by-task 带 checkbox 步骤的计划（docs/superpowers/plans/2026-08-02-m12-business-flow\.md）
4. **执行 M12**：按 Phase 1→5 顺序实施

