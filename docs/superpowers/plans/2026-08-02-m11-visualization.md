# M11 专业可视化与跨模块联动 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补全 10 模块中唯一明确缺失的核心可视化——进度管理甘特图（对标 MS Project 进度引擎），并交付三类专业能力闭环：交付物检入/检出 Check-in/Out 排他编辑（对标青翼/华田 PLM）、变更影响分析 diff 可视化（对标青翼 ECN/ECR）、跨模块联动（任务-交付物绑定、投入度→成本核算、RACI 联动）。M10 已完成混合表单适配，已有 7 个专业可视化页面（cost/curve、schedule/critical、risk/matrix、scope/wbs、eps/tree、eps/members、issue/board），M11 不重做这些页面。

**Architecture:** 甘特图采用 echarts custom series（自定义系列）实现，无新依赖（项目已装 echarts）。时间轴用 xAxis type=time，任务条用 renderItem 绘制矩形，依赖连线用 markLine 或 renderItem 贝塞尔曲线，关键路径任务条高亮红色。交付物检入检出在后端 deliverable service 新增 CheckIn/CheckOut 方法 + 状态机锁定（checked_out_by 字段），前端 versions.vue 增加检入检出按钮。变更 diff 复用 change/impactReport API，前端用 el-diff 或双栏对比组件展示字段级差异。跨模块联动通过 ref 字段 + 聚合查询实现，不引入新表。

**Tech Stack:** Vue 3 + Element Plus + echarts 5.x（custom series）+ gin-vue-admin v3.0.0 + EAV + Go 1.22+

## Global Constraints

- 甘特图必须用 echarts custom series，禁止引入 dhtmlx-gantt/frappe-gantt 等新依赖（保持依赖精简）
- 甘特图数据复用 `getScheduleTasks` API（已有），不新增查询接口；依赖关系复用 task 的 `dependency_type` + 关联字段
- 交付物检入检出必须实现排他编辑：同一交付物同一时间只能有一人 checkout，checkout 后他人不可 checkout 也不可 update
- 变更 diff 必须字段级对比（基线版本 vs 当前版本），展示 field_label + 旧值 + 新值，禁止整对象 JSON diff
- 跨模块联动用 ref 字段 + 后端聚合查询，禁止前端 N+1 查询
- 所有新增 API 路由前缀统一 `/api/pmocker/<mod>/`，写操作挂 OperationRecord
- 前端 HTTP 用 `@/utils/request`，文件名 kebab-case，组件名 PascalCase
- 改动后需 `--rebuild` 重建前后端才能在实例中生效

## 优先级总览

| 批次 | Task | 内容 | 对标 | 优先级 |
|------|------|------|------|------|
| P0 | T1 | schedule/gantt.vue 真甘特图（echarts custom series） | MSP 进度引擎 | 极高 |
| P1 | T2 | deliverable 检入检出 Check-in/Out 排他编辑 | QY/HT PLM | 高 |
| P1 | T3 | change 变更 diff 可视化（字段级对比） | QY ECN/ECR | 高 |
| P2 | T4 | 跨模块联动：任务-交付物绑定 | 需求文档 2.4① | 中 |
| P2 | T5 | 跨模块联动：投入度→成本核算 | 需求文档 2.4⑩ | 中 |
| P2 | T6 | 跨模块联动：RACI 矩阵联动（范围↔团队） | 需求文档 2.4⑨ | 中 |
| P3 | T7 | 端到端验证 + 重建镜像 | — | — |

---

## Task 1: schedule/gantt.vue 真甘特图（P0）

**Problem:** [schedule/gantt.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/schedule/gantt.vue) 名为「甘特图」实为 el-table 表格，无时间轴、无任务条、无依赖连线、无关键路径高亮，与需求文档 2.3 节「进度管理 ★ MS Project 进度引擎：甘特图视图」严重不符。这是 10 模块中唯一明确缺失的核心可视化。

**Goal:** 用 echarts custom series 实现专业甘特图：时间轴（xAxis type=time）+ 任务条（按 start_date/end_date 绘制矩形）+ 进度填充（progress%）+ 依赖连线（FS/SS/FF/SF）+ 关键路径高亮（is_critical_path 红色）+ 里程碑菱形标记 + 缩放/拖拽。保留现有表格作为下方明细列表。

**Files:**
- Modify: `gva/web/src/view/pmocker/schedule/gantt.vue`（重写 template + script）
- Reuse: `gva/web/src/api/pmocker/schedule.js` 的 `getScheduleTasks`（无需新增 API）

**Interfaces:**
- Consumes: `getScheduleTasks(params)` 返回 `{list: [{ID, title, status, attrs: {start_date, end_date, duration, progress, is_critical_path, total_float, dependency_type, ...}}]}`
- Produces: 甘特图组件 `PmockerScheduleGantt`，echarts 实例渲染时间轴+任务条

- [ ] **Step 1: 确认 getScheduleTasks 返回数据结构**

  Read: `gva/server/plugin/pmocker_schedule/service/schedule.go` 的 List 方法
  确认返回的 task.attrs 包含 start_date/end_date/progress/is_critical_path/dependency_type 字段（M9 已补全这些字段）
  若 dependency 关联字段缺失（如 predecessor_id），需确认 task schema 是否有前置任务引用字段

- [ ] **Step 2: 重写 gantt.vue template（甘特图 + 明细表格双区）**

```vue
<template>
  <div>
    <div class="gva-btn-list">
      <el-button type="primary" @click="openTaskDialog(null)">
        <svg-icon icon="lucide:plus" /> 新增任务
      </el-button>
      <el-button-group class="ml-2">
        <el-button :type="viewMode==='day'?'primary':''" @click="setViewMode('day')">日</el-button>
        <el-button :type="viewMode==='week'?'primary':''" @click="setViewMode('week')">周</el-button>
        <el-button :type="viewMode==='month'?'primary':''" @click="setViewMode('month')">月</el-button>
      </el-button-group>
    </div>
    <div class="gva-table-box">
      <div ref="chartRef" style="width: 100%; height: 500px" />
      <el-empty v-if="!tableData.length" description="暂无任务数据" />
    </div>
    <!-- 下方保留原明细表格 -->
    <el-table :data="tableData" row-key="ID" class="mt-4">
      <!-- 保留原有表格列 -->
    </el-table>
    <!-- 保留原 Dialog 表单 -->
  </div>
</template>
```

- [ ] **Step 3: 实现 echarts custom series 甘特图渲染**

```javascript
import * as echarts from 'echarts'

const renderGantt = (tasks) => {
  if (chartInstance) chartInstance.dispose()
  if (!chartRef.value) return
  chartInstance = echarts.init(chartRef.value)

  // 任务条数据：每个 task 一个 custom series item
  const minTime = Math.min(...tasks.map(t => new Date(t.attrs.start_date).getTime()))
  const maxTime = Math.max(...tasks.map(t => new Date(t.attrs.end_date).getTime()))
  const pad = (maxTime - minTime) * 0.1 // 时间轴留白

  const series = [{
    type: 'custom',
    renderItem: (params, api) => {
      const categoryIdx = api.value(0) // y 轴任务索引
      const start = api.coord([api.value(1), categoryIdx])
      const end = api.coord([api.value(2), categoryIdx])
      const progress = api.value(3) // 进度 0-100
      const isCritical = api.value(4) // 关键路径
      const barHeight = api.size([0, 1])[1] * 0.6

      const rectShape = echarts.graphic.clipRectByRect(
        { x: start[0], y: start[1] - barHeight / 2, width: end[0] - start[0], height: barHeight },
        { x: params.coordSys.x, y: params.coordSys.y, width: params.coordSys.width, height: params.coordSys.height }
      )
      if (!rectShape) return
      // 进度填充
      const progressWidth = rectShape.width * (progress / 100)
      return {
        type: 'group',
        children: [
          { type: 'rect', transition: ['shape'], shape: rectShape, style: { fill: isCritical ? '#ffe6e6' : '#e6f7ff', stroke: isCritical ? '#ff4d4f' : '#1890ff' } },
          { type: 'rect', shape: { x: rectShape.x, y: rectShape.y, width: progressWidth, height: rectShape.height }, style: { fill: isCritical ? '#ff4d4f' : '#1890ff', opacity: 0.6 } },
          { type: 'text', style: { text: api.value(5), x: rectShape.x + 4, y: rectShape.y + rectShape.height / 2, fill: '#333', fontSize: 11 } }
        ]
      }
    },
    encode: { x: [1, 2], y: 0 },
    data: tasks.map((t, idx) => ({
      name: t.title,
      value: [idx, new Date(t.attrs.start_date).getTime(), new Date(t.attrs.end_date).getTime(), t.attrs.progress || 0, t.attrs.is_critical_path ? 1 : 0, t.title]
    }))
  }]

  // 依赖连线（FS 为主）：用 markLine 或单独 series
  // 里程碑用 scatter 菱形

  chartInstance.setOption({
    tooltip: { formatter: (p) => { /* 显示任务详情 */ } },
    grid: { left: 200, right: 40, top: 40, bottom: 60 },
    xAxis: { type: 'time', min: minTime - pad, max: maxTime + pad },
    yAxis: { type: 'category', data: tasks.map(t => t.title), inverse: true },
    dataZoom: [{ type: 'slider', xAxisIndex: 0, filterMode: 'weakFilter' }, { type: 'inside', xAxisIndex: 0 }],
    series
  })
}
```

- [ ] **Step 4: 实现依赖连线**

  遍历 tasks，对有 predecessor 的任务，用 echarts graphic 画贝塞尔曲线从前置任务条末端连到当前任务条前端。不同 dependency_type（FS/SS/FF/SF）用不同颜色/线型。
  若 task schema 无 predecessor_id 字段，则在 Step 1 确认后，用 dependency_type + 关联字段（可能需查询 task 的 ref 字段）。

- [ ] **Step 5: 实现里程碑菱形标记**

  里程碑（task_type=milestone 或 duration=0）用 echarts scatter symbol=diamond 绘制在 due_date 位置。

- [ ] **Step 6: 实现视图切换（日/周/月）**

  `setViewMode(mode)` 调整 xAxis 的 minInterval：
  - day: 24*3600*1000
  - week: 7*24*3600*1000
  - month: 30*24*3600*1000

- [ ] **Step 7: onMounted 加载数据 + 窗口 resize 自适应**

```javascript
onMounted(() => { loadData() ; window.addEventListener('resize', handleResize) })
onBeforeUnmount(() => { window.removeEventListener('resize', handleResize) ; chartInstance?.dispose() })
const handleResize = () => chartInstance?.resize()
```

- [ ] **Step 8: 保留原有 CRUD Dialog（DynamicForm）**

  甘特图上方的「新增任务」按钮 + 下方明细表格的「编辑」按钮，复用原有 openTaskDialog + DynamicForm(entity-type="task")。

- [ ] **Step 9: 前端编译验证**

  Run: `cd gva/web && npm run build`
  Expected: build success

- [ ] **Step 10: 浏览器实测**

  Run: `./cli/pmocker.exe run -n pms-dev -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080 -f --rebuild`
  访问进度管理-甘特图，新建 3 个任务（含依赖关系、1 个关键路径、1 个里程碑）
  Expected: 甘特图渲染任务条 + 进度填充 + 依赖连线 + 关键路径红色 + 里程碑菱形 + 缩放可用

- [ ] **Step 11: Commit**

  Run: `git add gva/web/src/view/pmocker/schedule/gantt.vue && git commit -m "feat(schedule): 实现专业甘特图（echarts custom series 时间轴+任务条+依赖连线+关键路径高亮）"`

---

## Task 2: deliverable 检入检出 Check-in/Out 排他编辑（P1）

**Problem:** [deliverable/versions.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/deliverable/versions.vue) 只有版本列表表格，无检入检出（Check-in/Out）排他编辑功能。需求文档 2.3 节「交付物管理 ★ 青翼PLM + 华天PLM：检入/检出 Check-in/Out 排他编辑」明确要求。当前 [deliverable.js](file:///d:/Dev/pmocker/gva/web/src/api/pmocker/deliverable.js) 无 checkin/checkout API。

**Goal:** 实现交付物检入检出排他编辑：checkout 后交付物被当前用户锁定（他人不可 checkout 也不可 update），checkin 时可填写版本说明并解锁。状态：available（可检出）→ checked_out（已检出）→ available（检入后）。

**Files:**
- Modify: `gva/server/plugin/pmocker_deliverable/service/deliverable.go`（新增 CheckOut/CheckIn 方法）
- Modify: `gva/server/plugin/pmocker_deliverable/api/deliverable.go`（新增 CheckOut/CheckIn handler）
- Modify: `gva/server/plugin/pmocker_deliverable/router/deliverable.go`（新增路由）
- Modify: `gva/web/src/api/pmocker/deliverable.js`（新增 checkOut/checkIn 方法）
- Modify: `gva/web/src/view/pmocker/deliverable/versions.vue`（新增检入检出按钮 + 锁定状态显示）

**Interfaces:**
- Produces:
  - 后端 `CheckOut(ctx, deliverableID, userID) error` —— 设置 checked_out_by=userID, checked_out_at=now，状态→checked_out
  - 后端 `CheckIn(ctx, deliverableID, userID, versionNote string) error` —— 清除锁定，状态→available，创建新版本记录
  - 前端 `checkOut(data)` / `checkIn(data)` API
- Consumes: deliverable schema 需有 checked_out_by(ref→sys_users)、checked_out_at(datetime) 字段（若缺需在 schema.yaml 补充）

- [ ] **Step 1: 确认 deliverable schema 是否有锁定字段**

  Read: `gva/server/plugin/pmocker_deliverable/pmocker/schema.yaml`
  确认是否有 `checked_out_by`(ref→sys_users) + `checked_out_at`(datetime) + `lock_status`(enum) 字段。若缺，补充到 schema.yaml。

- [ ] **Step 2: 后端 service 新增 CheckOut/CheckIn 方法**

  Read: `gva/server/plugin/pmocker_deliverable/service/deliverable.go`
  CheckOut 逻辑：查当前 lock_status，若已 checked_out 且 checked_out_by != 当前用户 → 返回 error「交付物已被 XXX 检出，无法编辑」；否则设置锁定。
  CheckIn 逻辑：校验 checked_out_by == 当前用户，清除锁定，调用 createVersion 记录版本。

- [ ] **Step 3: 后端 api + router 新增路由**

  - POST `/deliverable/checkOut` → CheckOut（挂 OperationRecord）
  - POST `/deliverable/checkIn` → CheckIn（挂 OperationRecord）

- [ ] **Step 4: 前端 deliverable.js 新增 API**

```javascript
export const checkOutDeliverable = (data) => service({ url: '/pmocker/deliverable/checkOut', method: 'post', data })
export const checkInDeliverable = (data) => service({ url: '/pmocker/deliverable/checkIn', method: 'post', data })
```

- [ ] **Step 5: 前端 versions.vue 增加检入检出按钮 + 锁定状态**

  表格新增「锁定状态」列：available 显示「可用」绿色标签，checked_out 显示「已检出（XXX）」红色标签。
  操作列：available 状态显示「检出」按钮；checked_out 且是当前用户显示「检入」按钮，否则禁用并 tooltip 提示「已被 XXX 检出」。
  检入时弹 Dialog 填写版本说明。

- [ ] **Step 6: Update 方法增加锁定校验**

  后端 deliverable update 方法：若 lock_status=checked_out 且 checked_out_by != 当前用户 → 拒绝更新。

- [ ] **Step 7: 编译 + 浏览器实测**

  Run: `cd gva/server && go build ./...` + `cd gva/web && npm run build`
  Run: `./cli/pmocker.exe run -n pms-dev -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080 -f --rebuild`
  测试：用户A检出→用户B尝试检出失败→用户A检入→用户B可检出

- [ ] **Step 8: Commit**

  Run: `git add gva/server/plugin/pmocker_deliverable/ gva/web/src/{api,view}/pmocker/deliverable* && git commit -m "feat(deliverable): 实现检入检出 Check-in/Out 排他编辑（锁定状态+版本记录）"`

---

## Task 3: change 变更 diff 可视化（P1）

**Problem:** [change/impact.vue](file:///d:/Dev/pmocker/gva/web/src/view/pmocker/change/impact.vue) 只有影响分析表格，无变更前后对比 diff 可视化。需求文档 2.3 节「变更管理 ★ 青翼PLM ECN/ECR：变更前后对比：基线对比、版本差异 diff（与「交付物管理」联动）」明确要求。当前 [change.js](file:///d:/Dev/pmocker/gva/web/src/api/pmocker/change.js) 有 `getChangeImpactReport` 但无专用 diff API。

**Goal:** 实现变更请求的基线版本 vs 当前版本字段级 diff 可视化：展示 field_label + 旧值 + 新值 + 变化标记，支持交付物版本 diff 联动。

**Files:**
- Modify: `gva/server/plugin/pmocker_change/service/change.go`（新增 GetDiff 方法，对比基线 snapshot vs 当前 attrs）
- Modify: `gva/server/plugin/pmocker_change/api/change.go`（新增 GetDiff handler）
- Modify: `gva/server/plugin/pmocker_change/router/change.go`（新增路由）
- Modify: `gva/web/src/api/pmocker/change.js`（新增 getChangeDiff 方法）
- Modify: `gva/web/src/view/pmocker/change/impact.vue`（新增 diff 对比面板）

**Interfaces:**
- Produces:
  - 后端 `GetDiff(ctx, changeID) ([]FieldDiff, error)` 返回 `[{field_key, field_label, old_value, new_value, changed: bool}]`
  - 前端 `getChangeDiff(params)` API
- Consumes: change_request 需有 baseline_snapshot(json) 字段存储变更前的基线快照（若缺需在 schema 补充，或在审批通过时自动生成 snapshot）

- [ ] **Step 1: 确认 change_request schema 是否有 baseline_snapshot 字段**

  Read: `gva/server/plugin/pmocker_change/pmocker/schema.yaml`
  若无 baseline_snapshot(json)，补充字段，并在审批通过时自动 snapshot 当前 attrs。

- [ ] **Step 2: 后端 service 新增 GetDiff 方法**

  对比 baseline_snapshot (json) vs 当前 entity attrs，逐字段比较，返回 FieldDiff 列表。
  对 ref 字段需解析显示名（如 requester_id → 用户名）。

- [ ] **Step 3: 后端 api + router 新增路由**

  GET `/change/diff/:id` → GetDiff

- [ ] **Step 4: 前端 change.js 新增 API**

```javascript
export const getChangeDiff = (params) => service({ url: '/pmocker/change/diff/' + params.id, method: 'get' })
```

- [ ] **Step 5: 前端 impact.vue 新增 diff 对比面板**

  双栏对比：左栏「基线值」右栏「当前值」，变化字段高亮黄色，新增字段绿色，删除字段红色。用 el-table 三列（字段名/旧值/新值）+ 行状态色。

- [ ] **Step 6: 编译 + 浏览器实测**

  测试：创建变更请求→修改关联实体字段→查看 diff→确认变化的字段高亮

- [ ] **Step 7: Commit**

  Run: `git add gva/server/plugin/pmocker_change/ gva/web/src/{api,view}/pmocker/change* && git commit -m "feat(change): 实现变更前后字段级 diff 可视化（基线对比+高亮变化）"`

---

## Task 4: 跨模块联动——任务-交付物绑定（P2）

**Problem:** 需求文档 2.4①「任务-交付物绑定：WBS 任务直接关联设计文件/图纸/零部件，任务完成自动触发交付物签入」未实现。当前 schedule task 与 deliverable 无关联。

**Goal:** 在 task schema 增加 deliverable_id(ref→deliverable) 字段（M9 已有），前端 schedule/gantt.vue 和 deliverable/list.vue 增加绑定入口，任务完成（status=done）时自动触发关联交付物 checkin。

**Files:**
- Modify: `gva/server/plugin/pmocker_schedule/service/schedule.go`（Transition 方法：task→done 时触发关联 deliverable checkin）
- Modify: `gva/web/src/view/pmocker/schedule/gantt.vue`（任务表单显示关联交付物，DynamicForm 已渲染 deliverable_id）

- [ ] **Step 1: 确认 task schema 有 deliverable_id(ref→deliverable) 字段**

  Read: `gva/server/plugin/pmocker_schedule/pmocker/schema.yaml`，确认 M9 已补 deliverable_id。若无需补。

- [ ] **Step 2: 后端 schedule Transition 增加联动逻辑**

  task status→done 时，若 attrs.deliverable_id 非空，调用 deliverable service CheckIn 自动检入。

- [ ] **Step 3: 前端 gantt.vue 任务表单显示关联交付物名**

  DynamicForm 已渲染 deliverable_id（ref 字段），但显示 ID 不友好。改为 select 远程搜索交付物名称。

- [ ] **Step 4: 编译 + 测试 + Commit**

  Run: `git commit -m "feat(linkage): 任务完成自动触发关联交付物检入（任务-交付物绑定联动）"`

---

## Task 5: 跨模块联动——投入度→成本核算（P2）

**Problem:** 需求文档 2.4⑩「技能矩阵 + 投入度：团队成员技能等级 + 多项目投入度分摊，支撑资源分配决策」+ team 模块「时薪×工时→成本核算联动」未实现。

**Goal:** team_member 的 hourly_rate × allocation_percent × 周期 → 自动汇总到 cost_item 的 actual_cost。在 team/member.vue 显示成员成本贡献，在 cost/budget.vue 显示按成员分摊的成本。

**Files:**
- Modify: `gva/server/plugin/pmocker_team/service/team.go`（新增 GetCostContribution 方法）
- Modify: `gva/server/plugin/pmocker_cost/service/cost.go`（新增按成员聚合 actual_cost 查询）
- Modify: `gva/web/src/view/pmocker/team/member.vue`（表格增加「成本贡献」列）
- Modify: `gva/web/src/view/pmocker/cost/budget.vue`（增加按成员分摊视图）

- [ ] **Step 1: 后端 team service 新增 GetCostContribution**

  按成员聚合：hourly_rate × 估算工时（基于 allocation_percent × 周期）= 成本贡献

- [ ] **Step 2: 后端 cost service 新增按成员聚合查询**

- [ ] **Step 3: 前端 member.vue 增加成本贡献列**

- [ ] **Step 4: 前端 budget.vue 增加按成员分摊视图**

- [ ] **Step 5: 编译 + 测试 + Commit**

  Run: `git commit -m "feat(linkage): 团队投入度→成本核算联动（成员成本贡献+按成员分摊）"`

---

## Task 6: 跨模块联动——RACI 矩阵联动（P2）

**Problem:** 需求文档 2.4⑨「RACI 矩阵：范围项/任务显式标注 R/A/C/I 四角色，与团队管理角色定义联动」未实现。scope_item 有 raci_responsible/raci_accountable 字段（M9 已补），但前端无可视化 RACI 矩阵，且未与 team_role 联动。

**Goal:** 在 scope/wbs.vue 增加 RACI 矩阵视图（按 WBS 节点 × 团队角色，标注 R/A/C/I），raci 字段 select 选项从 team_role 动态加载。

**Files:**
- Modify: `gva/web/src/view/pmocker/scope/wbs.vue`（增加 RACI 矩阵 tab）
- Modify: `gva/web/src/api/pmocker/scope.js`（若需新增 RACI 聚合查询）
- Reuse: `gva/web/src/api/pmocker/team.js` 的 `listRole`（加载角色选项）

- [ ] **Step 1: 前端 wbs.vue 增加 RACI 矩阵 tab**

  el-tabs 切换「WBS 树」/「RACI 矩阵」。RACI 矩阵用 el-table：行=WBS 节点，列=团队角色，单元格=R/A/C/I 选择器。

- [ ] **Step 2: RACI 字段 select 选项从 team_role 动态加载**

  onMounted 调 listRole，填充 raci_responsible/raci_accountable 的选项。

- [ ] **Step 3: 编译 + 测试 + Commit**

  Run: `git commit -m "feat(linkage): RACI 矩阵联动（WBS×团队角色，选项动态加载）"`

---

## Task 7: 端到端验证 + 重建镜像（P3）

- [ ] **Step 1: 重建默认镜像**

  Run: `cd images/pmbok6-hybrid && go run .`
  Expected: 镜像构建成功，包含所有 M11 改动

- [ ] **Step 2: 启动实例全流程验证**

  Run: `./cli/pmocker.exe run -n pms-dev -i images/pmbok6-hybrid/pmbok6-hybrid.pmi -p 8080 -f --rebuild`
  验证：
  - 甘特图：任务条+依赖+关键路径+里程碑+缩放
  - 交付物：检出→他人不可编辑→检入→版本记录
  - 变更 diff：基线对比+字段高亮
  - 任务-交付物联动：任务完成触发交付物检入
  - 投入度-成本联动：成员成本贡献显示
  - RACI 矩阵：WBS×角色矩阵渲染
  - 已有 7 个可视化页面回归无破坏

- [ ] **Step 3: Commit + Push**

  Run: `git add -A && git commit -m "chore(m11): 重建默认镜像含 M11 可视化与联动" && git push`

---

## 验收标准

1. **甘特图**：schedule/gantt.vue 渲染时间轴+任务条+进度填充+依赖连线+关键路径高亮+里程碑菱形+缩放，对标 MS Project
2. **检入检出**：交付物 checkout 后他人不可 checkout/update，checkin 生成版本记录，对标青翼/华田 PLM
3. **变更 diff**：字段级对比基线 vs 当前，变化字段高亮，对标青翼 ECN/ECR
4. **任务-交付物联动**：任务完成自动触发关联交付物检入
5. **投入度-成本联动**：成员成本贡献显示，按成员分摊成本视图
6. **RACI 联动**：WBS×团队角色矩阵，选项从 team_role 动态加载
7. **无回归**：M10 表单适配 + 已有 7 个可视化页面不受影响
8. **依赖精简**：无新前端依赖（甘特图用已有 echarts）

---

## 执行建议

- **P0 优先**：T1 甘特图是 M11 最核心交付，对标 MS Project，单独完成后即可显著提升进度模块实用性
- **T1 技术方案**：echarts custom series 是关键， renderItem 绘制矩形任务条。若 custom series 实现复杂度高，可降级用 markLine + scatter 组合，但 custom series 效果最佳
- **T2/T3 可并行**：交付物检入检出与变更 diff 独立，可并行开发
- **T4/T5/T6 跨模块联动**：依赖前三个 Task 完成，且需后端聚合查询支持，建议放在 P2 批次
- **T7 必须重建镜像**：M11 改动涉及后端 service + schema，需重建镜像才能在实例中生效
- **与 M10 衔接**：M10 收尾（T6+T7）完成后立即开 M11，复用同一实例 pms-dev
