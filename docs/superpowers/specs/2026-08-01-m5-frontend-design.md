# PMocker M5 前端页面阶段设计文档

## 目标

为 M3 已实现的 9 个项目管理后端模块构建完整的前端页面，遵循 `gva/aiDoc/` 规范，使用 Element Plus + UnoCSS 语义 token + 项目已有图表库依赖，交付可运行的 MVP 前端。

## 范围

* 9 个项目管理模块共 23 个前端页面

* 9 个 API 封装文件

* 同步更新 9 个 menu.yaml 的 component 字段

* 不含新增 npm 依赖（全部用项目已有库）

## 目录结构

```
gva/web/src/
├── api/pmocker/              # API 封装层（9 文件）
│   ├── requirement.js        # create/delete/update/find/list + submitReview/approve/reject + traceMatrix
│   ├── scope.js              # createItem/buildWBS/baseline + listItems/getWBS
│   ├── schedule.js           # createTask/createMilestone/baseline + listTasks/listMilestones + cpm
│   ├── cost.js               # createItem/baseline + listItems + evm
│   ├── risk.js               # create/delete/update + find/list + assess + matrix
│   ├── issue.js              # create/delete/update + find/list + assign/resolve/close/reopen + board/stats
│   ├── eps.js                # createNode/updateNode/deleteNode/moveNode + addMember/removeMember + listNodes/listMembers
│   ├── deliverable.js        # create/delete/update + submitReview/accept/reject + createVersion/baseline + listVersions/traceReport/stats
│   └── change.js             # create/delete/update + analyze/ccbReview/approve/reject/implement/verify/close + listLogs/impactReport/ccbStats
├── view/pmocker/             # 页面层（9 模块，23 页面）
│   ├── requirement/
│   │   ├── list.vue          # 需求列表 + CRUD + 审批流（submitReview→approve/reject）
│   │   └── matrix.vue        # 需求追踪矩阵（el-table 二维表）
│   ├── scope/
│   │   ├── wbs.vue           # WBS 树形分解（el-tree + 自定义节点 + 拖拽排序）
│   │   └── baseline.vue      # 范围基线列表 + baseline 操作
│   ├── schedule/
│   │   ├── gantt.vue         # 甘特图（el-table + 日期条 + echarts 自定义）
│   │   ├── critical.vue      # CPM 关键路径网络图（echarts graph）
│   │   └── milestone.vue     # 里程碑列表 + timeline
│   ├── cost/
│   │   ├── budget.vue        # 预算管理列表 + createItem + baseline
│   │   ├── evm.vue           # EVM 挣值分析（vue-echarts 折线图：PV/EV/AC）
│   │   └── curve.vue         # S 曲线（vue-echarts 面积图）
│   ├── risk/
│   │   ├── register.vue      # 风险登记册（列表 + CRUD + assess 评估）
│   │   └── matrix.vue        # 5×5 风险热力图（echarts heatmap）
│   ├── issue/
│   │   ├── list.vue          # 问题列表 + CRUD + 状态流转（assign→resolve→close→reopen）
│   │   └── board.vue         # 看板视图（vuedraggable + el-card，4 列：待处理/处理中/已解决/已关闭）
│   ├── eps/
│   │   ├── tree.vue          # EPS 组织树（el-tree + 新增/编辑/删除/移动节点）
│   │   └── members.vue       # 成员管理（选中 EPS 节点 → 添加/移除成员）
│   ├── deliverable/
│   │   ├── list.vue          # 交付物列表 + CRUD + 审批流（submitReview→accept/reject）
│   │   ├── versions.vue      # 版本追踪时间线（el-timeline + createVersion/baseline）
│   │   └── trace.vue         # 追溯视图（交付物→需求→范围项关联表）
│   └── change/
│       ├── list.vue          # 变更请求列表 + CRUD + 状态流转
│       ├── ccb.vue           # CCB 审批台（el-steps 流程 + approve/reject/implement/verify/close）
│       ├── log.vue           # 变更日志（el-table + listLogs）
│       └── impact.vue        # 影响分析报告（analyze + impactReport 可视化）
```

## 后端 API 矩阵

每个 API 封装文件遵循 aiDoc 规范：

* 统一走 `@/utils/request`

* JSDoc 补全接口说明

* 函数命名：`动词 + 名词`（如 `createRequirement`, `getRequirementList`）

### API 端点清单

**requirement** (`/requirement`):

* `POST /requirement/create` → createRequirement

* `DELETE /requirement/delete` → deleteRequirement

* `PUT /requirement/update` → updateRequirement

* `GET /requirement/find` → findRequirement

* `GET /requirement/list` → getRequirementList

* `POST /requirement/submitReview` → submitRequirementReview

* `POST /requirement/approve` → approveRequirement

* `POST /requirement/reject` → rejectRequirement

* `GET /requirement/traceMatrix` → getRequirementTraceMatrix

**scope** (`/scope`):

* `POST /scope/createItem` → createScopeItem

* `POST /scope/buildWBS` → buildScopeWBS

* `POST /scope/baseline` → createScopeBaseline

* `GET /scope/listItems` → getScopeItems

* `GET /scope/getWBS` → getScopeWBS

**schedule** (`/schedule`):

* `POST /schedule/createTask` → createScheduleTask

* `POST /schedule/createMilestone` → createScheduleMilestone

* `POST /schedule/baseline` → createScheduleBaseline

* `GET /schedule/listTasks` → getScheduleTasks

* `GET /schedule/listMilestones` → getScheduleMilestones

* `POST /schedule/cpm` → analyzeScheduleCPM

**cost** (`/cost`):

* `POST /cost/createItem` → createCostItem

* `POST /cost/baseline` → createCostBaseline

* `GET /cost/listItems` → getCostItems

* `POST /cost/evm` → analyzeCostEVM

**risk** (`/risk`):

* `POST /risk/create` → createRisk

* `DELETE /risk/delete` → deleteRisk

* `PUT /risk/update` → updateRisk

* `GET /risk/find` → findRisk

* `GET /risk/list` → getRiskList

* `POST /risk/assess` → assessRisk

* `GET /risk/matrix` → getRiskMatrix

**issue** (`/issue`):

* `POST /issue/create` → createIssue

* `DELETE /issue/delete` → deleteIssue

* `PUT /issue/update` → updateIssue

* `GET /issue/find` → findIssue

* `GET /issue/list` → getIssueList

* `POST /issue/assign` → assignIssue

* `POST /issue/resolve` → resolveIssue

* `POST /issue/close` → closeIssue

* `POST /issue/reopen` → reopenIssue

* `GET /issue/board` → getIssueBoard

* `GET /issue/stats` → getIssueStats

**eps** (`/eps`):

* `POST /eps/createNode` → createEPSNode

* `PUT /eps/updateNode` → updateEPSNode

* `DELETE /eps/deleteNode` → deleteEPSNode

* `POST /eps/moveNode` → moveEPSNode

* `POST /eps/addMember` → addEPSMember

* `DELETE /eps/removeMember` → removeEPSMember

* `GET /eps/listNodes` → getEPSNodes

* `GET /eps/listMembers` → getEPSMembers

**deliverable** (`/deliverable`):

* `POST /deliverable/create` → createDeliverable

* `DELETE /deliverable/delete` → deleteDeliverable

* `PUT /deliverable/update` → updateDeliverable

* `GET /deliverable/find` → findDeliverable

* `GET /deliverable/list` → getDeliverableList

* `POST /deliverable/submitReview` → submitDeliverableReview

* `POST /deliverable/accept` → acceptDeliverable

* `POST /deliverable/reject` → rejectDeliverable

* `POST /deliverable/createVersion` → createDeliverableVersion

* `POST /deliverable/baseline` → createDeliverableBaseline

* `GET /deliverable/listVersions` → getDeliverableVersions

* `GET /deliverable/traceReport` → getDeliverableTraceReport

* `GET /deliverable/stats` → getDeliverableStats

**change** (`/change`):

* `POST /change/create` → createChange

* `DELETE /change/delete` → deleteChange

* `PUT /change/update` → updateChange

* `GET /change/find` → findChange

* `GET /change/list` → getChangeList

* `POST /change/analyze` → analyzeChange

* `POST /change/ccbReview` → ccbReviewChange

* `POST /change/approve` → approveChange

* `POST /change/reject` → rejectChange

* `POST /change/implement` → implementChange

* `POST /change/verify` → verifyChange

* `POST /change/close` → closeChange

* `GET /change/listLogs` → getChangeLogs

* `GET /change/impactReport` → getChangeImpactReport

* `GET /change/ccbStats` → getChangeCCBStats

## 技术选型

### 通用页面（列表 + CRUD）

* Element Plus：el-table、el-form、el-dialog、el-button

* 查询区：`gva-search-box` + `el-form :inline="true"`

* 表格区：`gva-table-box` + `el-table`

* 弹窗表单：el-dialog + el-form

* 分页：el-pagination

### 特殊视图

| 视图       | 实现方案                                 | 依赖                    |
| -------- | ------------------------------------ | --------------------- |
| 需求追踪矩阵   | el-table 二维表，行=需求，列=范围项，单元格=关联状态     | Element Plus          |
| WBS 树形分解 | el-tree + 自定义节点插槽 + el-tree-v2（大数据量） | Element Plus          |
| 甘特图      | el-table + 自定义列渲染日期条                 | Element Plus          |
| CPM 关键路径 | echarts graph 类型，节点=任务，边=依赖关系，高亮关键路径 | echarts + vue-echarts |
| 里程碑      | el-timeline + el-card                | Element Plus          |
| EVM 挣值图  | vue-echarts 折线图，3 系列（PV/EV/AC）+ 数据缩放 | vue-echarts           |
| S 曲线     | vue-echarts 面积图                      | vue-echarts           |
| 风险矩阵     | echarts heatmap，5×5 网格，颜色=风险等级       | echarts               |
| 问题看板     | vuedraggable + el-card，4 列拖拽         | vuedraggable          |
| EPS 组织树  | el-tree + 树形操作                       | Element Plus          |
| 成员管理     | el-transfer + el-table               | Element Plus          |
| 版本时间线    | el-timeline + el-tag                 | Element Plus          |
| 追溯视图     | el-table 层级展开                        | Element Plus          |
| CCB 审批台  | el-steps + el-card + 审批表单            | Element Plus          |
| 变更日志     | el-table + el-tag 状态                 | Element Plus          |
| 影响分析     | el-descriptions + echarts 柱状图        | echarts               |

## aiDoc 规范遵循

### 前端约束

* HTTP 请求统一走 `@/utils/request`

* 全局状态用 Pinia（本阶段无需新增 store）

* 路由由后端 menu.yaml 驱动，前端通过异步路由加载

* 文件名 kebab-case，组件名 PascalCase，变量名 camelCase

### 组件写法

* `v-model` 一律用 `defineModel()`

* 有限枚举 prop 补 validator

* 图标用全局 `<svg-icon icon="lucide:xxx" />`

### 样式规范

* 业务页面用 Element Plus + UnoCSS 语义 token

* 避免硬编码颜色（`color: #xxx`）、字号（`font-size: xxpx`）

* 避免内联样式

* 主题相关用 CSS 变量

### API 封装

* 统一走 `service`（`@/utils/request`）

* JSDoc 风格接口说明

* 函数命名：`动词 + 名词`

### 工具函数复用

* HTTP 请求：`@/utils/request`

* 字典数据：`@/utils/dictionary`

* UUID：`CreateUUID`

* 按钮权限：`useBtnAuth`

* 命名转换：`@/utils/stringFun`

* 跨组件通信：`@/utils/bus`（事件总线）

## menu.yaml 更新

所有 9 个 menu.yaml 的 component 字段从 `plugin/pmocker_<mod>/view/xxx` 改为 `view/pmocker/<mod>/xxx`：

| 模块          | 旧 component                                | 新 component                         |
| ----------- | ------------------------------------------ | ----------------------------------- |
| requirement | `plugin/pmocker_requirement/view/list`     | `view/pmocker/requirement/list`     |
| requirement | `plugin/pmocker_requirement/view/matrix`   | `view/pmocker/requirement/matrix`   |
| scope       | `plugin/pmocker_scope/view/wbs`            | `view/pmocker/scope/wbs`            |
| scope       | `plugin/pmocker_scope/view/baseline`       | `view/pmocker/scope/baseline`       |
| schedule    | `plugin/pmocker_schedule/view/gantt`       | `view/pmocker/schedule/gantt`       |
| schedule    | `plugin/pmocker_schedule/view/critical`    | `view/pmocker/schedule/critical`    |
| schedule    | `plugin/pmocker_schedule/view/milestone`   | `view/pmocker/schedule/milestone`   |
| cost        | `plugin/pmocker_cost/view/budget`          | `view/pmocker/cost/budget`          |
| cost        | `plugin/pmocker_cost/view/evm`             | `view/pmocker/cost/evm`             |
| cost        | `plugin/pmocker_cost/view/curve`           | `view/pmocker/cost/curve`           |
| risk        | `plugin/pmocker_risk/view/register`        | `view/pmocker/risk/register`        |
| risk        | `plugin/pmocker_risk/view/matrix`          | `view/pmocker/risk/matrix`          |
| issue       | `plugin/pmocker_issue/view/list`           | `view/pmocker/issue/list`           |
| issue       | `plugin/pmocker_issue/view/board`          | `view/pmocker/issue/board`          |
| eps         | `plugin/pmocker_eps/view/tree`             | `view/pmocker/eps/tree`             |
| eps         | `plugin/pmocker_eps/view/members`          | `view/pmocker/eps/members`          |
| deliverable | `plugin/pmocker_deliverable/view/list`     | `view/pmocker/deliverable/list`     |
| deliverable | `plugin/pmocker_deliverable/view/versions` | `view/pmocker/deliverable/versions` |
| deliverable | `plugin/pmocker_deliverable/view/trace`    | `view/pmocker/deliverable/trace`    |
| change      | `plugin/pmocker_change/view/list`          | `view/pmocker/change/list`          |
| change      | `plugin/pmocker_change/view/ccb`           | `view/pmocker/change/ccb`           |
| change      | `plugin/pmocker_change/view/log`           | `view/pmocker/change/log`           |
| change      | `plugin/pmocker_change/view/impact`        | `view/pmocker/change/impact`        |

## Task 分解策略

按模块分 9 个 Task，每个 Task 包含：API 封装 + 页面组件 + menu.yaml 更新。
Task 内步骤：API 封装 → 列表页 → 特殊视图页 → 提交。

| Task | 模块          | 页面数 | 复杂度                  |
| ---- | ----------- | --- | -------------------- |
| 1    | requirement | 2   | 中（列表+审批流+追踪矩阵）       |
| 2    | scope       | 2   | 中（WBS 树+基线）          |
| 3    | schedule    | 3   | 高（甘特图+CPM 网络图+里程碑）   |
| 4    | cost        | 3   | 高（EVM 折线+S 曲线+预算）    |
| 5    | risk        | 2   | 中（登记册+热力图）           |
| 6    | issue       | 2   | 中（列表+看板拖拽）           |
| 7    | eps         | 2   | 中（组织树+成员管理）          |
| 8    | deliverable | 3   | 中（列表+版本时间线+追溯）       |
| 9    | change      | 4   | 高（列表+CCB 流程+日志+影响分析） |

## 验证策略

* 前端编译：`cd gva/web && npm run build` 无错误

* 页面可访问：启动后端 + 前端，菜单导航到各页面可加载

* API 调用：列表页能拉取数据（即使空列表不报错）

* aiDoc 规范自查：无硬编码颜色、无内联样式、API 走 request.js

