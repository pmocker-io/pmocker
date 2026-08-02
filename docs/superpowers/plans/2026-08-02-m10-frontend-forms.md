# M10 前端表单适配 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 10 大模块的前端页面能展示和编辑 M9 新增的 286 个 EAV 动态字段，通过「核心字段配置化 + 扩展字段动态渲染」的混合表单架构，使前端表单与 M9 元数据层对齐，CRUD 闭环可用、工作流状态转移按钮齐全。

**Architecture:** 采用混合表单架构。核心字段（L1 通用 + L2 模块核心）由 [coreFields.js](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/components/coreFields.js) 配置驱动渲染，扩展字段（L3）从后端 schema API 动态拉取并折叠展示。核心组件 [DynamicForm.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/components/DynamicForm.vue) 通过 `v-model` 双向绑定父组件 `reactive(form)` 对象，直接引用 `props.modelValue.attrs` 利用 Vue 3 深层响应式，避免替换整个对象导致响应式链断裂。后端 EAV API 提供通用 entity CRUD + schema 查询，各业务模块保留专用路由（如 `/pmocker/team/member/create`）复用 EAV service 层。

**Tech Stack:** Vue 3 (Composition API) + Element Plus + echarts（已有可视化页面）+ gin-vue-admin v3.0.0 + EAV 通用 API

## Global Constraints

- 前端 HTTP 请求统一用 `@/utils/request`，文件名 kebab-case，组件名 PascalCase
- EAV API 路由前缀统一为 `/pmocker/eav/`，与其他 pmocker 业务插件路由风格一致（commit 02dd6d33 已修复）
- DynamicForm 通过 `entity-type` prop 区分实体类型，核心字段配置在 coreFields.js 中可调整无需改组件代码
- `attrs` 必须直接引用 `props.modelValue.attrs`，禁止用 `emit('update:modelValue', {...})` 替换整个对象（会断裂响应式链，bug 已修 a367b029）
- 控件类型映射：string/text→el-input、int/decimal→el-input-number、date/datetime→el-date-picker、bool→el-switch、enum→el-select、ref/json→el-input
- 栅格规则：text/json 全宽 span=24，其余 span=12
- 业务模块专用路由（team/schedule/cost 等）复用 EAV service 层，后端硬编码绑定 entity_type，前端无需传 entity_type

## 优先级总览

| 批次 | Task | 内容 | 状态 | 优先级 |
|------|------|------|------|------|
| P0 | T1 | DynamicForm 核心组件 + coreFields 配置 | ✅ 已完成 | 极高 |
| P0 | T2 | 后端 EAV API（schema + entity CRUD） | ✅ 已完成 | 极高 |
| P1 | T3 | 10 模块前端页面适配（引入 DynamicForm） | ✅ 已完成 | 高 |
| P1 | T4 | API 层（schema.js + 10 模块专用 API） | ✅ 已完成 | 高 |
| P1 | T5 | Team 模块 4 页面（member/role/training/performance） | ✅ 已完成 | 高 |
| P2 | T6 | 收尾：修复 member.vue createMember 未传 status | ⬜ 待执行 | 中 |
| P2 | T7 | 端到端回归验证（EAV entity CRUD + 10 模块流程） | ⬜ 待执行 | 高 |

---

## Task 1: DynamicForm 核心组件 + coreFields 配置 ✅

**Files:**
- Create: `gva/web/src/view/pmocker/components/DynamicForm.vue`
- Create: `gva/web/src/view/pmocker/components/coreFields.js`

**Interfaces:**
- Produces: `DynamicForm` 组件（props: `entityType: string`, `modelValue: {title, status, attrs}`）；`getCoreFieldKeys(entityType): string[]`、`isCoreField(entityType, fieldKey): boolean`

**完成内容：**

- [x] **Step 1: coreFields.js 配置 14 个实体类型核心字段**
  - L1 通用核心（`_universal`）：`code`, `description`
  - L2 模块核心：cost_item(17)/task(15)/milestone(6)/risk(14)/requirement(12)/scope_item(8)/issue(10)/eps_node(12)/deliverable(9)/change_request(11)/team_member(12)/team_role(6)/training_record(8)/performance_review(8)
  - 导出 `getCoreFieldKeys()`、`isCoreField()` 辅助函数

- [x] **Step 2: DynamicForm.vue 实现混合表单**
  - 核心字段区（el-divider「基本信息」）+ 扩展属性折叠区（el-collapse）
  - 控件类型映射 `getComponent()`：9 种 data_type → Element Plus 组件
  - 栅格 `getColSpan()`：text/json=24，其余=12
  - enum 选项解析 `parseOptions()`、默认值解析 `parseDefaultValue()`
  - onMounted 加载 schema + watch entityType 重新加载

- [x] **Step 3: attrs 深层响应式（修复 bug a367b029）**
  - `attrs` 改为 computed 直接引用 `props.modelValue.attrs`，初始化时 `props.modelValue.attrs = {}`
  - 禁止 `emit('update:modelValue', {...})` 替换整个对象

---

## Task 2: 后端 EAV API ✅

**Files:**
- Modify: `gva/server/router/pmocker/eav.go`
- Modify: `gva/server/api/v1/pmocker/eav.go`
- Modify: `gva/server/plugin/pmocker_core/plugin.go`（路由前缀修复）

**Interfaces:**
- Produces: 7 个 EAV 路由（POST/PUT/DELETE/GET entity + GET entities + GET/POST schema）

**完成内容：**

- [x] **Step 1: EAV 路由注册**（写操作挂 OperationRecord，读操作不挂）
  - POST `eav/entity` → CreateEntity
  - POST `eav/schema` → RegisterSchema
  - PUT `eav/entity` → UpdateEntity
  - DELETE `eav/entity/:id` → DeleteEntity
  - GET `eav/entity/:id` → GetEntity
  - GET `eav/entities` → ListEntities（支持 entityType 过滤）
  - GET `eav/schema/:entityType` → GetSchema

- [x] **Step 2: EAV API 实现**
  - CreateEntity/UpdateEntity 接收 `eavtypes.Entity`（含 entity_type/title/status/attrs）
  - ListEntities 支持 `projectId` + `entityType` + `offset` + `limit` 参数
  - GetSchema 返回 `{entity_type, fields}`

- [x] **Step 3: EAV 路由前缀一致性修复（commit 02dd6d33）**
  - 修改前：`group.Group("pmocker")` → 路由在 `/pmocker/eav/`（无 `/api` 前缀）
  - 修改后：`group.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")` → 路由在 `/api/pmocker/eav/`，与其他业务插件统一

---

## Task 3: 10 模块前端页面适配 ✅

**Files:**
- 27 个 .vue 文件分布在 `gva/web/src/view/pmocker/{requirement,scope,deliverable,schedule,cost,risk,issue,change,eps,team}/`

**完成内容：**

- [x] **Step 1: 10 模块页面全部存在**
  - requirement: list, matrix
  - scope: wbs, baseline
  - deliverable: list, trace, versions
  - schedule: gantt, critical, milestone
  - cost: budget, curve, evm
  - risk: matrix, register
  - issue: list, board
  - change: list, ccb, impact, log
  - eps: tree, members
  - team: member, role, training, performance

- [x] **Step 2: 代表性页面引入 DynamicForm**
  - [requirement/list.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/requirement/list.vue)：搜索+表格+分页+工作流状态转移按钮（提交评审/批准/驳回）+ DynamicForm(entity-type="requirement")
  - [team/member.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/team/member.vue)：表格+Dialog+DynamicForm(entity-type="team_member")
  - [schedule/gantt.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/schedule/gantt.vue)：表格+DynamicForm(entity-type="task")（注：名为甘特图实为表格，专业可视化留待 M11）

- [x] **Step 3: 已有 7 个专业可视化页面（无需 M10 处理）**
  - cost/curve.vue：echarts S 曲线（PV/EV/AC）
  - schedule/critical.vue：echarts graph 关键路径网络图
  - risk/matrix.vue：echarts 风险矩阵热力图
  - scope/wbs.vue：el-tree WBS 树
  - eps/tree.vue：el-tree EPS 树
  - eps/members.vue：el-tree/draggable
  - issue/board.vue：draggable 看板泳道

---

## Task 4: API 层 ✅

**Files:**
- Create: `gva/web/src/api/pmocker/schema.js`
- Existing: 10 模块专用 API（change/cost/deliverable/eps/issue/requirement/risk/schedule/scope/team.js）

**完成内容：**

- [x] **Step 1: schema.js 提供 EAV 通用 API**
  - getSchema(entityType) / createEntity(data) / updateEntity(data) / deleteEntity(id) / getEntity(id) / listEntities(params)

- [x] **Step 2: 10 模块专用 API 齐全**
  - 各模块 CRUD + 工作流状态转移 API（如 requirement 的 submitReview/approve/reject，change 的 ccbReview/approve/reject/implement/verify/close）

- [x] **Step 3: Vite 代理端口修复**
  - `.env.development`：VITE_SERVER_PORT 从 8888 改为 8080（与 pmocker run 默认端口一致）

---

## Task 5: Team 模块 4 页面 ✅

**Files:**
- Create: `gva/web/src/view/pmocker/team/{member,role,training,performance}.vue`
- Existing: `gva/web/src/api/pmocker/team.js`

**完成内容：**

- [x] **Step 1: 4 页面齐全**
  - member.vue：团队成员（投入度/技能等级/状态流候选→入职→在职→离职→已离职）
  - role.vue：角色定义（RACI 默认/权限级别/编制）
  - training.vue：培训记录（柯氏四级评估）
  - performance.vue：绩效评估（评级/360°/IDP）

- [x] **Step 2: team.js API 完整**
  - 4 实体 × CRUD = 20 个方法，走专用路由 `/pmocker/team/{member,role,training,performance}/{create,delete,update,find,list}`

- [x] **Step 3: team 后端复用 EAV service**
  - [team.go](file:///d:/Dev/pmocker/gva/server/plugin/pmocker_team/api/team.go) 通用 create/delete/update/find/list/transition 方法，硬编码绑定 EntityTypeTeamMember 等，前端无需传 entity_type

---

## Task 6: 收尾——修复 member.vue createMember 未传 status ⬜

**Problem:** [member.vue#L136](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/team/member.vue#L136) 的 `createMember({ title: form.title, attrs: form.attrs })` 未传 `status`，而 update（L134）传了 `status: form.status`。虽然后端 service 层有默认值兜底，但前端不一致且新增成员状态不可控。

**Goal:** createMember 调用补传 `status: form.status`，与 update 保持一致。

**Files:**
- Modify: `gva/web/src/view/pmocker/team/member.vue:136`

**Interfaces:**
- Consumes: team 后端 `create(c, entityType)` 接收 `{projectId, title, attrs, creatorId}`（status 由 service 默认值处理，但前端传 status 可作为意向值，若后端支持则采纳）

- [ ] **Step 1: 检查 team service 层 create 方法是否接收 status**

  Read: `gva/server/plugin/pmocker_team/service/team.go`
  确认 `Create(ctx, entityType, projectID, title, attrs, creatorID)` 签名。若不接收 status，则前端传 status 无效，需评估是否改 service 签名。若 service 用默认值 'candidate'，则前端可不传，仅补注释说明。

- [ ] **Step 2: 修复 member.vue createMember 调用**

  根据Step 1 结果：
  - 方案 A（service 不接收 status）：保持 createMember 不传 status，补注释 `// status 由后端 service 默认 'candidate'`，并在 form 初始化时确保 `status: 'candidate'`
  - 方案 B（service 可扩展接收 status）：改为 `createMember({ title: form.title, status: form.status, attrs: form.attrs })`

```javascript
// 修改前（member.vue L136）
await createMember({ title: form.title, attrs: form.attrs })

// 修改后（方案 B）
await createMember({ title: form.title, status: form.status, attrs: form.attrs })
```

- [ ] **Step 3: 同步检查 role/training/performance 三页 create 调用**

  Read: `gva/web/src/view/pmocker/team/{role,training,performance}.vue`
  确认三页 createXxx 调用是否同样缺失 status，统一修复。

- [ ] **Step 4: 前端编译验证**

  Run: `cd gva/web && npm run build`
  Expected: build success，无 lint error

- [ ] **Step 5: Commit**

  Run: `git add gva/web/src/view/pmocker/team/ && git commit -m "fix(pmocker): team 模块 create 调用补传 status 保持一致性"`

---

## Task 7: 端到端回归验证 ⬜

**Problem:** M10 完成后未做完整端到端验证。memory 中有两个遗留疑虑需澄清：
1. "EAV entity API 404" —— 路由检查完整（GET entity/:id 存在）且前缀已修复，需实测确认
2. "team createMember attrs 保存" —— 前端不传 entity_type 是正确的（后端硬编码绑定），需实测确认 attrs 真的落库

**Goal:** 通过 `pmocker run --rebuild` 启动实例，验证 EAV entity CRUD 全闭环 + 10 模块新建→填字段→保存→列表显示→编辑全流程。

**Files:**
- 无代码改动，纯验证任务

- [ ] **Step 1: 启动实例（强制重建确保最新代码）**

  Run: `./cli/pmocker.exe run -n pms-dev -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080 -f --rebuild`
  Expected: 实例启动成功，前端可访问 http://127.0.0.1:8080

- [ ] **Step 2: 验证 EAV schema API**

  浏览器登录后，打开任一模块（如需求管理-列表），点「新增需求」打开 Dialog。
  Expected: DynamicForm 渲染出核心字段（基本信息区）+ 扩展属性折叠区，字段数与 M9 schema 一致（requirement 约 21 字段）
  Console: 无 JS error

- [ ] **Step 3: 验证 EAV entity 创建（attrs 落库）**

  在需求 Dialog 填写：名称=「测试需求001」+ 核心字段 priority=high + moscow_priority=Must + 扩展字段若干 → 保存
  Expected: 保存成功，列表显示「测试需求001」+ priority=high + moscow_priority=Must
  验证 attrs 真的落库：刷新页面后字段值仍在

- [ ] **Step 4: 验证 EAV entity 更新**

  编辑「测试需求001」，修改 priority=low → 保存
  Expected: 列表更新为 priority=low

- [ ] **Step 5: 验证 EAV entity 详情查询（澄清 404 疑虑）**

  后端日志或浏览器 Network 面板检查：GET `/api/pmocker/eav/entity/:id` 返回 200（非 404）
  Expected: entity 详情含 attrs 完整字段

- [ ] **Step 6: 验证 team 模块 attrs 保存（澄清遗留疑虑）**

  打开团队管理-成员，点「新增成员」填：名称=「张三」+ role=PM + allocation_percent=80 + skill_level=senior → 保存
  Expected: 保存成功，列表显示张三 + role=PM + 投入度 80% + 技能等级 senior
  验证 attrs 落库：编辑张三，确认字段值仍在

- [ ] **Step 7: 抽查 3 个模块的新建→编辑闭环**

  抽查：成本管理-预算、风险管理-登记册、进度管理-甘特图
  每个模块：新建填字段→保存→列表显示→编辑修改→保存→确认更新
  Expected: 全部通过，字段值正确显示和保存

- [ ] **Step 8: 验证已有 7 个专业可视化页面可渲染**

  逐一访问：cost/curve（S曲线）、schedule/critical（关键路径图）、risk/matrix（风险矩阵）、scope/wbs（WBS树）、eps/tree（EPS树）、issue/board（看板）
  Expected: 图表/树/看板正常渲染，Console 无 error

- [ ] **Step 9: 验证工作流状态转移**

  在需求列表：草稿→提交评审→批准（或驳回）
  Expected: 状态正确流转，按钮按状态显示/隐藏

- [ ] **Step 10: 记录验证结果**

  若全部通过：M10 标记完成，更新 topics.md
  若有失败：记录失败现象 + 根因 + 修复方案，必要时开新 Task

---

## 验收标准

1. **混合表单架构**：DynamicForm + coreFields 可配置，核心字段区 + 扩展属性折叠区正常渲染
2. **10 模块页面**：全部存在且引入 DynamicForm，CRUD 闭环可用
3. **EAV API**：7 路由完整，entity CRUD 全闭环（创建/查询/更新/删除），schema 查询正常
4. **路由一致性**：EAV 路由在 `/api/pmocker/eav/`，与其他业务插件统一
5. **attrs 落库**：核心字段 + 扩展字段值能保存到 EAV 属性表，刷新后仍在
6. **team 模块**：4 页面齐全，createMember/updateMember attrs 正确保存
7. **已有可视化**：7 个专业页面（curve/critical/matrix/wbs/eps-tree/eps-members/board）正常渲染
8. **工作流**：状态转移按钮齐全，状态正确流转
9. **无回归**：M1-M9 既有功能不受影响
10. **Console 无 error**：浏览器 Console 无 JS error

---

## 执行建议

- **T6 + T7 顺序执行**：先修 member.vue status（T6），再端到端验证（T7），避免验证时发现 status 问题又回头改
- **T7 用 --rebuild**：确保前后端用最新代码，避免缓存导致误判
- **T7 失败处理**：若 EAV entity 404 复现，优先检查 [pmocker_core/plugin.go](file:///d:/Dev/pmocker/gva/server/plugin/pmocker_core/plugin.go) 路由注册和 Casbin 规则初始化
- **M10 完成后立即开 M11**：M11 聚焦 schedule/gantt.vue 真甘特图 + 交付物检入检出 + 变更 diff + 跨模块联动
