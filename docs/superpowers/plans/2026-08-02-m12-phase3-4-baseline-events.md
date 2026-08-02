# M12 Phase 3-4: 基线与偏差 + 业务事件引擎 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 激活 pmocker 的基线快照与偏差分析能力（T8/T9），并在现有工作流引擎上叠加 NodeHook 事件机制，实现审批通过自动生成基线、变更关闭自动应用、任务完成自动刷新项目完成度与红黄绿健康度（T10/T11）。

**Architecture:** 在 Phase 1-2 已激活的 EAV 数据层（pm_entities/pm_attrs + pm_time_entries/pm_cost_actuals/pm_change_logs）之上，新增 4 个 service：`BaselineService`（全量 JSON 快照 + 字段级 diff）、`VarianceService`（PV/EV/AC → SV/CV/SPI/CPI + 预警）、`ProgressService`（3 种完成度算法 + 健康度）、以及 `NodeHook` 机制（在 `WorkflowService.Execute`/`tryAdvanceAuto` 的状态转移点同步回调）。Hook 与现有 `AutoHandler` 并存，互不干扰；所有 hook 同步执行，与工作流事务串联，简单可控。

**Tech Stack:** Go 1.21+ / gin-vue-admin v3.0.0 / GORM / Vue 3 + Element Plus / echarts 5

## Global Constraints

- 依赖 Phase 1-2 已完成：`PMEntity.Priority` 字段、`PMBaseline.ChangeReqID` 字段、`PMTimeEntry`/`PMCostActual`/`PMApprovalRecord` model、`ChangeLogService.RecordChangeLog`、`RelationService`、`TaskLinkService` 均已存在
- 自定义代码在 `gva/server/model/pmocker`、`gva/server/service/pmocker`、`gva/server/api/v1/pmocker`、`gva/server/router/pmocker`
- EAV API 路由统一使用 `/pmocker/` 前缀（由上层 RouterGroup 注册，子路由用 `private.Group("xxx")`）
- service 通过 `pmocker.ServiceGroupApp` 嵌入聚合；api 通过 `apiGroup`（router 包内 `var apiGroup = api.ApiGroupApp`）引用；写操作挂 `middleware.OperationRecord()`
- 前端 HTTP 使用 `@/utils/request`；文件名 kebab-case，组件名 PascalCase
- `NodeHook` 定义在 gva service 层（`service/pmocker/workflow.go`），不改动 `pkg/pmocker/workflow` 包
- 自动生成的基线 `createdBy` 取自工作流执行人（通过 context 传递），缺失时为 0（系统）

## File Structure

### 后端新增/修改文件

| 文件 | 职责 | 操作 |
|------|------|------|
| gva/server/service/pmocker/attr_helper.go | EAV 属性读写工具函数（全 service 共享） | 新增 |
| gva/server/service/pmocker/baseline.go | 基线快照生成 + 字段级对比 service | 新增 |
| gva/server/service/pmocker/variance.go | PV/EV/AC 偏差计算 + 预警 service | 新增 |
| gva/server/service/pmocker/progress.go | 3 种完成度算法 + 健康度 service | 新增 |
| gva/server/service/pmocker/hooks.go | 3 个内置 NodeHook（schedule/cost/change） | 新增 |
| gva/server/service/pmocker/workflow.go | 新增 NodeHook 接口/注册表/转移点回调 | 修改 |
| gva/server/service/pmocker/enter.go | ServiceGroup 嵌入 3 个新 service | 修改 |
| gva/server/api/v1/pmocker/baseline.go | 基线 API handler | 新增 |
| gva/server/api/v1/pmocker/variance.go | 偏差 API handler | 新增 |
| gva/server/api/v1/pmocker/progress.go | 完成度 API handler | 新增 |
| gva/server/api/v1/pmocker/enter.go | ApiGroup 嵌入 3 个新 API | 修改 |
| gva/server/router/pmocker/business.go | 追加 baseline/variance/progress 路由 | 修改 |
| gva/server/plugin/pmocker_core/initialize/init.go | 注册 4 个 NodeHook | 修改 |

### 前端新增文件

| 文件 | 职责 | 操作 |
|------|------|------|
| gva/web/src/api/pmocker/baseline.js | 基线 API | 新增 |
| gva/web/src/api/pmocker/variance.js | 偏差 API | 新增 |
| gva/web/src/api/pmocker/progress.js | 完成度 API | 新增 |
| gva/web/src/view/pmocker/cost/baseline.vue | 基线列表 + 对比 diff 页 | 新增 |
| gva/web/src/view/pmocker/cost/variance.vue | SPI/CPI 仪表盘 + 偏差柱状图 | 新增 |

---

## Task 1: 基线快照管理（spec T8）

**Files:**
- Create: `gva/server/service/pmocker/attr_helper.go`
- Create: `gva/server/service/pmocker/baseline.go`
- Create: `gva/server/api/v1/pmocker/baseline.go`
- Modify: `gva/server/service/pmocker/enter.go`
- Modify: `gva/server/api/v1/pmocker/enter.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/baseline.js`
- Create: `gva/web/src/view/pmocker/cost/baseline.vue`

**Interfaces:**
- Consumes: `pmocker.PMBaseline`（含 `ChangeReqID`，Phase 1-2 已加）、`pmocker.PMEntity`、`pmocker.PMAttr`
- Produces:
  - `BaselineService.CreateBaseline(projectID uint, baselineType string, changeReqID *uint, createdBy uint) (uint, error)`
  - `BaselineService.CompareBaseline(baselineID uint) ([]BaselineDiff, error)`
  - `BaselineService.ListBaselines(projectID uint, baselineType string) ([]pmocker.PMBaseline, error)`
  - 共享属性工具：`readAttrString`/`readAttrDecimal`/`readAttrInt`/`readAttrRef`/`writeAttrString`/`writeAttrDecimal`（后续 Task 3/4 复用）
  - REST: `POST /pmocker/baseline/create`、`GET /pmocker/baseline/list`、`GET /pmocker/baseline/compare`

- [ ] **Step 1: 创建 attr_helper.go（共享属性读写工具）**

创建 `gva/server/service/pmocker/attr_helper.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// readAttrString 读取实体的字符串类属性（ValString/ValDate/ValDateTime/ValJSON 任一）
func readAttrString(entityID uint, key string) string {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValDate != nil {
		return *attr.ValDate
	}
	if attr.ValDateTime != nil {
		return *attr.ValDateTime
	}
	if attr.ValJSON != nil {
		return *attr.ValJSON
	}
	return ""
}

// readAttrDecimal 读取实体的数值属性
func readAttrDecimal(entityID uint, key string) float64 {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValDecimal != nil {
		return *attr.ValDecimal
	}
	return 0
}

// readAttrInt 读取实体的整数属性
func readAttrInt(entityID uint, key string) int64 {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValInt != nil {
		return *attr.ValInt
	}
	return 0
}

// readAttrRef 读取实体的引用属性（优先 ValRef，回退 ValInt）
func readAttrRef(entityID uint, key string) uint {
	var attr pmocker.PMAttr
	global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValRef != nil {
		return *attr.ValRef
	}
	if attr.ValInt != nil {
		return uint(*attr.ValInt)
	}
	return 0
}

// writeAttrString 写入/更新字符串类属性
func writeAttrString(entityID uint, key, value string) error {
	v := value
	return upsertAttr(entityID, key, &v, nil, nil)
}

// writeAttrDecimal 写入/更新数值属性
func writeAttrDecimal(entityID uint, key string, value float64) error {
	v := value
	return upsertAttr(entityID, key, nil, &v, nil)
}

// upsertAttr 新建或更新属性（按 entity_id+field_key 唯一索引）
func upsertAttr(entityID uint, key string, str *string, dec *float64, intv *int64) error {
	var attr pmocker.PMAttr
	err := global.GVA_DB.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr).Error
	if err != nil {
		attr = pmocker.PMAttr{EntityID: entityID, FieldKey: key, ValString: str, ValDecimal: dec, ValInt: intv}
		return global.GVA_DB.Create(&attr).Error
	}
	updates := map[string]interface{}{}
	if str != nil {
		updates["val_string"] = *str
	}
	if dec != nil {
		updates["val_decimal"] = *dec
	}
	if intv != nil {
		updates["val_int"] = *intv
	}
	if len(updates) == 0 {
		return nil
	}
	return global.GVA_DB.Model(&pmocker.PMAttr{}).Where("id = ?", attr.ID).Updates(updates).Error
}

// attrValueString 把任意 PMAttr 的非空值统一序列化为字符串（用于快照与 diff）
func attrValueString(attr pmocker.PMAttr) string {
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValDecimal != nil {
		return strconv.FormatFloat(*attr.ValDecimal, 'f', -1, 64)
	}
	if attr.ValInt != nil {
		return strconv.FormatInt(*attr.ValInt, 10)
	}
	if attr.ValDate != nil {
		return *attr.ValDate
	}
	if attr.ValDateTime != nil {
		return *attr.ValDateTime
	}
	if attr.ValBool != nil {
		if *attr.ValBool {
			return "true"
		}
		return "false"
	}
	if attr.ValJSON != nil {
		return *attr.ValJSON
	}
	if attr.ValRef != nil {
		return strconv.FormatUint(uint64(*attr.ValRef), 10)
	}
	return ""
}
```

> 注意：需 `import "strconv"`。

- [ ] **Step 2: 创建 baseline.go service**

创建 `gva/server/service/pmocker/baseline.go`：

```go
package pmocker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type BaselineService struct{}

// 基线类型 → 需快照的实体类型集合
var baselineEntityTypes = map[string][]string{
	"schedule": {"task"},
	"cost":     {"task", "cost_item"},
	"scope":    {"scope_item"},
}

// snapshotEntity 快照中的单个实体（实体字段 + 属性字典）
type snapshotEntity struct {
	EntityID   uint              `json:"entityId"`
	EntityType string            `json:"entityType"`
	Title      string            `json:"title"`
	Status     string            `json:"status"`
	OwnerID    *uint             `json:"ownerId,omitempty"`
	Attrs      map[string]string `json:"attrs"`
}

// baselineSnapshot 快照根结构
type baselineSnapshot struct {
	ProjectID uint             `json:"projectId"`
	Type      string           `json:"type"`
	CreatedAt string           `json:"createdAt"`
	Entities  []snapshotEntity `json:"entities"`
}

// BaselineDiff 基线对比的字段级差异项
type BaselineDiff struct {
	EntityType   string `json:"entityType"`
	EntityID     uint   `json:"entityId"`
	EntityTitle  string `json:"entityTitle"`
	FieldKey     string `json:"fieldKey"`
	BaselineVal  string `json:"baselineVal"`
	CurrentVal   string `json:"currentVal"`
	Change       string `json:"change"` // added / removed / modified
}

// CreateBaseline 序列化项目相关实体的 attrs 为 JSON 快照，写入 pm_baselines
func (s *BaselineService) CreateBaseline(projectID uint, baselineType string, changeReqID *uint, createdBy uint) (uint, error) {
	types, ok := baselineEntityTypes[baselineType]
	if !ok {
		return 0, fmt.Errorf("unsupported baseline type: %s", baselineType)
	}

	var entities []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type IN ?", projectID, types).
		Order("entity_type, id").Find(&entities).Error; err != nil {
		return 0, err
	}

	snap := baselineSnapshot{
		ProjectID: projectID,
		Type:      baselineType,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Entities:  make([]snapshotEntity, 0, len(entities)),
	}

	for _, e := range entities {
		var attrs []pmocker.PMAttr
		global.GVA_DB.Where("entity_id = ?", e.ID).Find(&attrs)
		attrMap := make(map[string]string, len(attrs))
		for _, a := range attrs {
			attrMap[a.FieldKey] = attrValueString(a)
		}
		snap.Entities = append(snap.Entities, snapshotEntity{
			EntityID:   e.ID,
			EntityType: e.EntityType,
			Title:      e.Title,
			Status:     e.Status,
			OwnerID:    e.OwnerID,
			Attrs:      attrMap,
		})
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return 0, err
	}

	bl := pmocker.PMBaseline{
		ProjectID:    projectID,
		Type:         baselineType,
		SnapshotJSON: string(data),
		ChangeReqID:  changeReqID,
	}
	if err := global.GVA_DB.Create(&bl).Error; err != nil {
		return 0, err
	}

	// 把最新基线回写到项目实体的 baselineId（便于前端展示当前基线）
	global.GVA_DB.Model(&pmocker.PMEntity{}).Where("id = ?", projectID).
		Update("baseline_id", bl.ID)
	return bl.ID, nil
}

// ListBaselines 查询项目基线列表（可按类型过滤）
func (s *BaselineService) ListBaselines(projectID uint, baselineType string) ([]pmocker.PMBaseline, error) {
	var list []pmocker.PMBaseline
	db := global.GVA_DB.Where("project_id = ?", projectID)
	if baselineType != "" {
		db = db.Where("type = ?", baselineType)
	}
	err := db.Order("created_at DESC").Find(&list).Error
	return list, err
}

// CompareBaseline 加载基线快照并与当前数据逐字段对比，返回 diff
func (s *BaselineService) CompareBaseline(baselineID uint) ([]BaselineDiff, error) {
	var bl pmocker.PMBaseline
	if err := global.GVA_DB.First(&bl, baselineID).Error; err != nil {
		return nil, err
	}

	var snap baselineSnapshot
	if err := json.Unmarshal([]byte(bl.SnapshotJSON), &snap); err != nil {
		return nil, fmt.Errorf("invalid snapshot json: %w", err)
	}

	// 当前数据：按 entityID → field → value
	current := make(map[uint]map[string]string)
	currentMeta := make(map[uint]snapshotEntity)
	for _, se := range snap.Entities {
		var attrs []pmocker.PMAttr
		global.GVA_DB.Where("entity_id = ?", se.EntityID).Find(&attrs)
		m := make(map[string]string, len(attrs))
		for _, a := range attrs {
			m[a.FieldKey] = attrValueString(a)
		}
		current[se.EntityID] = m
		// 元信息（title/status）取当前实体
		var e pmocker.PMEntity
		global.GVA_DB.First(&e, se.EntityID)
		currentMeta[se.EntityID] = snapshotEntity{
			EntityID: e.ID, EntityType: e.EntityType, Title: e.Title, Status: e.Status,
		}
	}

	diffs := make([]BaselineDiff, 0)
	for _, se := range snap.Entities {
		curAttrs := current[se.EntityID]
		meta := currentMeta[se.EntityID]
		title := meta.Title
		if title == "" {
			title = se.Title
		}
		// 1) 基线有、当前可能改的字段
		for k, oldV := range se.Attrs {
			newV, exists := curAttrs[k]
			if !exists {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, oldV, "", "removed"})
				continue
			}
			if oldV != newV {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, oldV, newV, "modified"})
			}
		}
		// 2) 当前新增的字段
		for k, newV := range curAttrs {
			if _, exists := se.Attrs[k]; !exists {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, "", newV, "added"})
			}
		}
	}
	return diffs, nil
}
```

- [ ] **Step 3: 在 ServiceGroup 嵌入 BaselineService**

修改 `gva/server/service/pmocker/enter.go`，在 `ServiceGroup` 中新增 `BaselineService`：

```go
package pmocker

// ServiceGroup 聚合所有 PMocker service
type ServiceGroup struct {
	EAVService
	WorkflowService
	RBACService
	RelationService
	TaskLinkService
	ChangeLogService
	CostLinkService
	TimeEntryService
	CostActualService
	BaselineService
}

// ServiceGroupApp 全局 service 入口
var ServiceGroupApp = new(ServiceGroup)
```

> 说明：`RelationService` 等前缀项为 Phase 1-2 已嵌入；若 Phase 1-2 未按此名嵌入，仅保留 `BaselineService` 新增项即可。本计划假设 Phase 1-2 已完成嵌入。

- [ ] **Step 4: 创建 baseline API handler**

创建 `gva/server/api/v1/pmocker/baseline.go`：

```go
package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaselineApi struct{}

// parseUintParam 字符串→uint（缺失/非法返回 0）
func parseUintParam(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

// Create 创建基线快照
func (a *BaselineApi) Create(c *gin.Context) {
	type createReq struct {
		ProjectID    uint   `json:"projectId" binding:"required"`
		Type         string `json:"type" binding:"required"`
		ChangeReqID  *uint  `json:"changeReqId"`
	}
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	id, err := service.BaselineService.CreateBaseline(req.ProjectID, req.Type, req.ChangeReqID, userID)
	if err != nil {
		global.GVA_LOG.Error("创建基线失败", zap.Error(err))
		response.FailWithMessage("创建基线失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"baselineId": id}, "创建成功", c)
}

// List 查询基线列表
func (a *BaselineApi) List(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	baselineType := c.Query("type")
	list, err := service.BaselineService.ListBaselines(projectID, baselineType)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// Compare 基线对比
func (a *BaselineApi) Compare(c *gin.Context) {
	baselineID := parseUintParam(c.Query("baselineId"))
	diffs, err := service.BaselineService.CompareBaseline(baselineID)
	if err != nil {
		response.FailWithMessage("对比失败: "+err.Error(), c)
		return
	}
	response.OkWithData(diffs, c)
}
```

- [ ] **Step 5: 在 ApiGroup 嵌入 BaselineApi**

修改 `gva/server/api/v1/pmocker/enter.go`：

```go
package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"

// ApiGroup 聚合所有 PMocker API
type ApiGroup struct {
	EAVApi
	BaselineApi
}

// ApiGroupApp 全局 API 入口
var ApiGroupApp = new(ApiGroup)

// service PMocker service 引用
var service = pmocker.ServiceGroupApp
```

- [ ] **Step 6: 追加 baseline 路由**

修改 `gva/server/router/pmocker/business.go`，在 `BusinessRouter` 的 `InitBusiness(public, private)` 方法内追加（与 Phase 1-2 的 relation/taskLink 等路由并列；若该方法签名不同，按 eav.go 的 `InitEAV(public, private)` 签名对齐）：

```go
// 基线管理（写操作记录操作日志）
{
	bl := private.Group("baseline").Use(middleware.OperationRecord())
	bl.POST("create", apiGroup.BaselineApi.Create)
}
// 基线查询（读操作不记录）
{
	bl := private.Group("baseline")
	bl.GET("list", apiGroup.BaselineApi.List)
	bl.GET("compare", apiGroup.BaselineApi.Compare)
}
```

> 需要 `import "github.com/flipped-aurora/gin-vue-admin/server/middleware"`。`apiGroup` 已在 `router/pmocker/enter.go` 定义为 `var apiGroup = api.ApiGroupApp`。

- [ ] **Step 7: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过，无错误

- [ ] **Step 8: 创建前端 baseline.js API**

创建 `gva/web/src/api/pmocker/baseline.js`：

```javascript
import service from '@/utils/request'

export const createBaseline = (data) => {
  return service({ url: '/pmocker/baseline/create', method: 'post', data })
}

export const listBaselines = (params) => {
  return service({ url: '/pmocker/baseline/list', method: 'get', params })
}

export const compareBaseline = (params) => {
  return service({ url: '/pmocker/baseline/compare', method: 'get', params })
}
```

- [ ] **Step 9: 创建前端 baseline.vue 基线对比页**

创建 `gva/web/src/view/pmocker/cost/baseline.vue`：

```vue
<template>
  <div class="baseline-page">
    <el-card shadow="never">
      <el-form :inline="true" size="small">
        <el-form-item label="项目ID">
          <el-input v-model="projectId" style="width: 120px" />
        </el-form-item>
        <el-form-item label="基线类型">
          <el-select v-model="baselineType" style="width: 140px" clearable>
            <el-option label="计划基线" value="schedule" />
            <el-option label="成本基线" value="cost" />
            <el-option label="范围基线" value="scope" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadList">查询</el-button>
          <el-button type="success" @click="createBaseline">生成基线</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="baselines" size="small" border highlight-current-row @row-click="selectBaseline">
        <el-table-column prop="ID" label="ID" width="80" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">{{ typeText(row.type) }}</template>
        </el-table-column>
        <el-table-column prop="CreatedAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click.stop="compare(row)">对比</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card v-if="diffs.length" shadow="never" style="margin-top: 12px" header="基线对比差异">
      <el-table :data="diffs" size="small" border>
        <el-table-column prop="entityTitle" label="实体" width="160" />
        <el-table-column prop="entityType" label="类型" width="100" />
        <el-table-column prop="fieldKey" label="字段" width="140" />
        <el-table-column prop="baselineVal" label="基线值" />
        <el-table-column prop="currentVal" label="当前值" />
        <el-table-column prop="change" label="变化" width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.change)" size="small">{{ row.change }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { listBaselines, createBaseline, compareBaseline } from '@/api/pmocker/baseline'

const projectId = ref('')
const baselineType = ref('')
const baselines = ref([])
const diffs = ref([])

const typeText = (t) => ({ schedule: '计划基线', cost: '成本基线', scope: '范围基线' }[t] || t)
const tagType = (c) => ({ added: 'success', removed: 'danger', modified: 'warning' }[c] || 'info')

const loadList = async () => {
  if (!projectId.value) return ElMessage.warning('请输入项目ID')
  const res = await listBaselines({ projectId: projectId.value, type: baselineType.value })
  if (res.code === 0) baselines.value = res.data || []
}

const createBaseline = async () => {
  if (!projectId.value || !baselineType.value) return ElMessage.warning('请输入项目ID并选择类型')
  const res = await createBaseline({ projectId: Number(projectId.value), type: baselineType.value })
  if (res.code === 0) {
    ElMessage.success('基线已生成')
    loadList()
  }
}

const selectBaseline = (row) => { compare(row) }

const compare = async (row) => {
  const res = await compareBaseline({ baselineId: row.ID })
  if (res.code === 0) {
    diffs.value = res.data || []
    if (!diffs.value.length) ElMessage.info('无差异')
  }
}
</script>
```

- [ ] **Step 10: 提交**

```bash
git add gva/server/service/pmocker/attr_helper.go gva/server/service/pmocker/baseline.go gva/server/api/v1/pmocker/baseline.go gva/server/service/pmocker/enter.go gva/server/api/v1/pmocker/enter.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/baseline.js gva/web/src/view/pmocker/cost/baseline.vue
git commit -m "feat(pmocker): 基线快照管理(T8) 全量JSON快照+字段级diff+对比页面"
```

---

## Task 2: 偏差分析与预警（spec T9）

**Files:**
- Create: `gva/server/service/pmocker/variance.go`
- Create: `gva/server/api/v1/pmocker/variance.go`
- Modify: `gva/server/service/pmocker/enter.go`
- Modify: `gva/server/api/v1/pmocker/enter.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/variance.js`
- Create: `gva/web/src/view/pmocker/cost/variance.vue`

**Interfaces:**
- Consumes: `pmocker.PMEntity`（task/cost_item/risk）、`pmocker.PMCostActual`、共享属性工具（Task 1 产出）、`BaselineService`（可选读取最新成本基线）
- Produces:
  - `VarianceService.CalcVariance(projectID uint) (*VarianceReport, error)`
  - `VarianceService.GetAlerts(projectID uint) ([]VarianceAlert, error)`
  - REST: `GET /pmocker/variance/calc`、`GET /pmocker/variance/alerts`

- [ ] **Step 1: 创建 variance.go service**

创建 `gva/server/service/pmocker/variance.go`：

```go
package pmocker

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type VarianceService struct{}

// VarianceReport 挣值管理偏差报告
type VarianceReport struct {
	ProjectID uint    `json:"projectId"`
	PV        float64 `json:"pv"`  // Planned Value 计划价值（任务预算之和）
	EV        float64 `json:"ev"`  // Earned Value 挣值（预算×进度%）
	AC        float64 `json:"ac"`  // Actual Cost 实际成本
	SV        float64 `json:"sv"`  // Schedule Variance 进度偏差 = EV - PV
	CV        float64 `json:"cv"`  // Cost Variance 成本偏差 = EV - AC
	SPI       float64 `json:"spi"` // Schedule Performance Index = EV / PV
	CPI       float64 `json:"cpi"` // Cost Performance Index = EV / AC
	CalcAt    string  `json:"calcAt"`
}

// VarianceAlert 预警项
type VarianceAlert struct {
	Type     string `json:"type"`     // overdue / over_budget / high_risk
	Severity string `json:"severity"` // warning / critical
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	EntityID uint   `json:"entityId"`
}

// CalcVariance 基于 PV/EV/AC 计算 SV/CV/SPI/CPI
func (s *VarianceService) CalcVariance(projectID uint) (*VarianceReport, error) {
	// 1) PV = sum(task.budget_cost)；EV = sum(task.budget_cost * progress/100)
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return nil, err
	}
	pv, ev := 0.0, 0.0
	for _, t := range tasks {
		budget := readAttrDecimal(t.ID, "budget_cost")
		if budget <= 0 {
			// 回退：estimated_hours * owner.hourly_rate（由 Phase 1-2 CostLinkService 同步）
			hours := readAttrDecimal(t.ID, "estimated_hours")
			if t.OwnerID != nil {
				rate := readAttrDecimal(*t.OwnerID, "hourly_rate")
				budget = hours * rate
			}
		}
		pv += budget
		progress := readAttrDecimal(t.ID, "progress")
		ev += budget * progress / 100.0
	}

	// 2) AC = sum(pm_cost_actuals.amount where status=confirmed)
	var acStruct = struct{ Total float64 }{}
	global.GVA_DB.Model(&pmocker.PMCostActual{}).
		Where("project_id = ? AND status = ?", projectID, "confirmed").
		Select("COALESCE(SUM(amount),0) as total").Scan(&acStruct)
	ac := acStruct.Total

	rpt := &VarianceReport{
		ProjectID: projectID,
		PV:        pv,
		EV:        ev,
		AC:        ac,
		SV:        ev - pv,
		CV:        ev - ac,
		CalcAt:    time.Now().Format("2006-01-02 15:04:05"),
	}
	if pv > 0 {
		rpt.SPI = ev / pv
	}
	if ac > 0 {
		rpt.CPI = ev / ac
	}
	return rpt, nil
}

// GetAlerts 检查超期任务、超支成本、高风险项
func (s *VarianceService) GetAlerts(projectID uint) ([]VarianceAlert, error) {
	alerts := make([]VarianceAlert, 0)
	today := time.Now().Format("2006-01-02")

	// 1) 超期任务：end_date < today 且 status != done
	var tasks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks)
	for _, t := range tasks {
		if t.Status == "done" {
			continue
		}
		end := readAttrString(t.ID, "end_date")
		if end != "" && end < today {
			alerts = append(alerts, VarianceAlert{
				Type: "overdue", Severity: "critical",
				Title: "任务超期: " + t.Title,
				Detail: fmt.Sprintf("计划结束 %s，当前状态 %s", end, t.Status),
				EntityID: t.ID,
			})
		}
	}

	// 2) 超支：CPI < 1 视为成本超支
	rpt, err := s.CalcVariance(projectID)
	if err == nil && rpt != nil {
		if rpt.PV > 0 && rpt.SPI < 0.9 {
			alerts = append(alerts, VarianceAlert{
				Type: "schedule_off", Severity: severityOf(rpt.SPI),
				Title: "进度偏差预警",
				Detail: fmt.Sprintf("SPI=%.2f（SV=%.2f），进度落后", rpt.SPI, rpt.SV),
			})
		}
		if rpt.AC > 0 && rpt.CPI < 0.9 {
			alerts = append(alerts, VarianceAlert{
				Type: "over_budget", Severity: severityOf(rpt.CPI),
				Title: "成本超支预警",
				Detail: fmt.Sprintf("CPI=%.2f（CV=%.2f），实际成本超出挣值", rpt.CPI, rpt.CV),
			})
		}
	}

	// 3) 高风险项：severity=high
	var risks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "risk").Find(&risks)
	for _, r := range risks {
		sev := readAttrString(r.ID, "severity")
		if sev == "high" || sev == "高" {
			alerts = append(alerts, VarianceAlert{
				Type: "high_risk", Severity: "critical",
				Title: "高风险: " + r.Title,
				Detail: readAttrString(r.ID, "description"),
				EntityID: r.ID,
			})
		}
	}
	return alerts, nil
}

func severityOf(idx float64) string {
	if idx < 0.7 {
		return "critical"
	}
	return "warning"
}
```

- [ ] **Step 2: 在 ServiceGroup 嵌入 VarianceService**

修改 `gva/server/service/pmocker/enter.go`，在 `ServiceGroup` 中新增 `VarianceService`：

```go
type ServiceGroup struct {
	EAVService
	WorkflowService
	RBACService
	RelationService
	TaskLinkService
	ChangeLogService
	CostLinkService
	TimeEntryService
	CostActualService
	BaselineService
	VarianceService
}
```

- [ ] **Step 3: 创建 variance API handler**

创建 `gva/server/api/v1/pmocker/variance.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type VarianceApi struct{}

// Calc 计算偏差
func (a *VarianceApi) Calc(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	rpt, err := service.VarianceService.CalcVariance(projectID)
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	response.OkWithData(rpt, c)
}

// Alerts 查询预警
func (a *VarianceApi) Alerts(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	alerts, err := service.VarianceService.GetAlerts(projectID)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(alerts, c)
}
```

- [ ] **Step 4: 在 ApiGroup 嵌入 VarianceApi**

修改 `gva/server/api/v1/pmocker/enter.go`，在 `ApiGroup` 中新增 `VarianceApi`：

```go
type ApiGroup struct {
	EAVApi
	BaselineApi
	VarianceApi
}
```

- [ ] **Step 5: 追加 variance 路由**

修改 `gva/server/router/pmocker/business.go`，追加：

```go
// 偏差分析（读操作）
{
	vr := private.Group("variance")
	vr.GET("calc", apiGroup.VarianceApi.Calc)
	vr.GET("alerts", apiGroup.VarianceApi.Alerts)
}
```

- [ ] **Step 6: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 7: 创建前端 variance.js API**

创建 `gva/web/src/api/pmocker/variance.js`：

```javascript
import service from '@/utils/request'

export const calcVariance = (params) => {
  return service({ url: '/pmocker/variance/calc', method: 'get', params })
}

export const getAlerts = (params) => {
  return service({ url: '/pmocker/variance/alerts', method: 'get', params })
}
```

- [ ] **Step 8: 创建前端 variance.vue 偏差图表**

创建 `gva/web/src/view/pmocker/cost/variance.vue`：

```vue
<template>
  <div class="variance-page">
    <el-card shadow="never">
      <el-form :inline="true" size="small">
        <el-form-item label="项目ID">
          <el-input v-model="projectId" style="width: 120px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAll">计算偏差</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="8">
        <el-card shadow="never" header="SPI 进度绩效">
          <div ref="spiRef" style="height: 240px" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" header="CPI 成本绩效">
          <div ref="cpiRef" style="height: 240px" />
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never" header="偏差（EV-PV/AC）">
          <div ref="barRef" style="height: 240px" />
        </el-card>
      </el-col>
    </el-row>

    <el-card v-if="alerts.length" shadow="never" style="margin-top: 12px" header="预警列表">
      <el-table :data="alerts" size="small" border>
        <el-table-column prop="type" label="类型" width="120" />
        <el-table-column prop="severity" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="row.severity === 'critical' ? 'danger' : 'warning'" size="small">{{ row.severity }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" />
        <el-table-column prop="detail" label="详情" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, nextTick, onBeforeUnmount } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { calcVariance, getAlerts } from '@/api/pmocker/variance'

const projectId = ref('')
const spiRef = ref(null)
const cpiRef = ref(null)
const barRef = ref(null)
const alerts = ref([])
let charts = []

const renderGauge = (dom, value, name) => {
  if (!dom) return
  const inst = echarts.init(dom)
  charts.push(inst)
  inst.setOption({
    series: [{
      type: 'gauge', min: 0, max: 2, splitNumber: 8,
      progress: { show: true, width: 18 },
      axisLine: { lineStyle: { width: 18 } },
      pointer: { width: 5 },
      detail: { valueAnimation: true, formatter: '{value}', fontSize: 20, offsetCenter: [0, '70%'] },
      data: [{ value: Number(value.toFixed(2)), name }]
    }]
  })
}

const renderBar = (dom, sv, cv) => {
  if (!dom) return
  const inst = echarts.init(dom)
  charts.push(inst)
  inst.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: ['进度偏差 SV', '成本偏差 CV'] },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar', barWidth: 40,
      data: [
        { value: Number(sv.toFixed(2)), itemStyle: { color: sv >= 0 ? '#67C23A' : '#F56C6C' } },
        { value: Number(cv.toFixed(2)), itemStyle: { color: cv >= 0 ? '#67C23A' : '#F56C6C' } }
      ]
    }]
  })
}

const loadAll = async () => {
  if (!projectId.value) return ElMessage.warning('请输入项目ID')
  charts.forEach(c => c.dispose()); charts = []
  const [vr, ar] = await Promise.all([
    calcVariance({ projectId: projectId.value }),
    getAlerts({ projectId: projectId.value })
  ])
  if (vr.code === 0 && vr.data) {
    await nextTick()
    renderGauge(spiRef.value, vr.data.spi, 'SPI')
    renderGauge(cpiRef.value, vr.data.cpi, 'CPI')
    renderBar(barRef.value, vr.data.sv, vr.data.cv)
  }
  if (ar.code === 0) alerts.value = ar.data || []
}

onBeforeUnmount(() => { charts.forEach(c => c.dispose()); charts = [] })
</script>
```

- [ ] **Step 9: 提交**

```bash
git add gva/server/service/pmocker/variance.go gva/server/api/v1/pmocker/variance.go gva/server/service/pmocker/enter.go gva/server/api/v1/pmocker/enter.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/variance.js gva/web/src/view/pmocker/cost/variance.vue
git commit -m "feat(pmocker): 偏差分析与预警(T9) PV/EV/AC挣值+SPI/CPI仪表盘+超期超支预警"
```

---

## Task 3: 工作流 NodeHook 事件引擎（spec T10）

**Files:**
- Modify: `gva/server/service/pmocker/workflow.go`
- Create: `gva/server/service/pmocker/hooks.go`
- Modify: `gva/server/plugin/pmocker_core/initialize/init.go`

**Interfaces:**
- Consumes: `WorkflowService.Execute`/`tryAdvanceAuto`（现有）、`BaselineService.CreateBaseline`（Task 1）、`ChangeLogService.RecordChangeLog`（Phase 1-2）、共享属性工具（Task 1）
- Produces:
  - `NodeHook` 接口：`OnEnter(ctx, entityID, nodeName) error` + `OnLeave(ctx, entityID, nodeName, action) error`
  - `WorkflowService.RegisterNodeHook(workflowCode, nodeName string, hook NodeHook)`
  - 3 个内置 hook：`ScheduleBaselineHook`、`CostBaselineHook`、`ChangeApplyHook`
  - context 工具：`hookUserIDKey`、`userIDFromCtx(ctx) uint`

- [ ] **Step 1: 在 workflow.go 新增 NodeHook 接口、注册表与 context 工具**

修改 `gva/server/service/pmocker/workflow.go`，在文件 import 块之后、`WorkflowService` 结构体之前新增：

```go
// NodeHook 节点事件钩子：在工作流节点进入/离开时同步回调。
// key 为 workflowCode + "." + nodeName，与 AutoHandler 并存，互不干扰。
type NodeHook interface {
	OnEnter(ctx context.Context, entityID uint, nodeName string) error
	OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error
}

// hookUserIDKey 用于通过 context 传递执行人 ID 给 hook
type hookCtxKey string

const hookUserIDKey hookCtxKey = "pmocker.hook.userID"

// userIDFromCtx 从 context 读取执行人 ID（缺失返回 0=系统）
func userIDFromCtx(ctx context.Context) uint {
	if v, ok := ctx.Value(hookUserIDKey).(uint); ok {
		return v
	}
	return 0
}
```

修改 `WorkflowService` 结构体，新增 `hooks` 字段：

```go
type WorkflowService struct {
	mu       sync.RWMutex
	handlers map[string]workflow.AutoHandler
	hooks    map[string]NodeHook
}
```

- [ ] **Step 2: 实现 RegisterNodeHook 与触发助手**

在 `workflow.go` 的 `RegisterAutoHandler` 方法之后新增：

```go
// RegisterNodeHook 注册节点事件钩子（幂等覆盖）。
// key = workflowCode + "." + nodeName，例如 "plan_approval.approve"
func (s *WorkflowService) RegisterNodeHook(workflowCode, nodeName string, hook NodeHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hooks == nil {
		s.hooks = make(map[string]NodeHook)
	}
	s.hooks[workflowCode+"."+nodeName] = hook
}

// fireOnLeave 离开节点时回调（忽略未注册的 key，忽略 nil error 以保证主流程不中断）
func (s *WorkflowService) fireOnLeave(ctx context.Context, workflowCode, nodeName, action string, entityID uint) error {
	s.mu.RLock()
	h, ok := s.hooks[workflowCode+"."+nodeName]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return h.OnLeave(ctx, entityID, nodeName, action)
}

// fireOnEnter 进入节点时回调
func (s *WorkflowService) fireOnEnter(ctx context.Context, workflowCode, nodeName string, entityID uint) error {
	s.mu.RLock()
	h, ok := s.hooks[workflowCode+"."+nodeName]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return h.OnEnter(ctx, entityID, nodeName)
}
```

- [ ] **Step 3: 在 Execute 转移点触发 hook**

修改 `Execute` 方法，在「应用状态转移」DB 更新成功之后、`tryAdvanceAuto` 调用之前，插入 hook 触发。定位现有代码：

```go
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// 3) 若目标节点是 auto，调度 handler 并继续链式推进
	if targetNode != nil && targetNode.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(ctx, instanceID, &wf, 0)
	}
	return nil
```

改为：

```go
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// 3) 触发 NodeHook：离开当前节点（带 action）+ 进入目标节点
	hookCtx := context.WithValue(ctx, hookUserIDKey, userID)
	fromNode := inst.CurrentNode
	if err := s.fireOnLeave(hookCtx, inst.WorkflowCode, fromNode, action, inst.EntityID); err != nil {
		return fmt.Errorf("onLeave hook failed on node %s: %w", fromNode, err)
	}
	if err := s.fireOnEnter(hookCtx, inst.WorkflowCode, target.To, inst.EntityID); err != nil {
		return fmt.Errorf("onEnter hook failed on node %s: %w", target.To, err)
	}
	// 4) 若目标节点是 auto，调度 handler 并继续链式推进
	if targetNode != nil && targetNode.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(hookCtx, instanceID, &wf, 0)
	}
	return nil
```

- [ ] **Step 4: 在 tryAdvanceAuto 链式推进时也触发 hook**

修改 `tryAdvanceAuto` 方法，在 C 段「应用下一次转移」DB 更新成功之后、D 段递归之前，插入 hook 触发。定位现有代码：

```go
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// D) 若下一个节点仍是 auto，递归推进
	if next != nil && next.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(ctx, instanceID, wf, depth+1)
	}
	return nil
```

改为：

```go
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// D) 触发 NodeHook：离开 auto 节点 + 进入下一节点
	leaveAction := autoTrans.On
	if leaveAction == "" {
		leaveAction = "done"
	}
	if err := s.fireOnLeave(ctx, inst.WorkflowCode, node.Code, leaveAction, inst.EntityID); err != nil {
		return fmt.Errorf("onLeave hook failed on auto node %s: %w", node.Code, err)
	}
	if next != nil {
		if err := s.fireOnEnter(ctx, inst.WorkflowCode, next.Code, inst.EntityID); err != nil {
			return fmt.Errorf("onEnter hook failed on node %s: %w", next.Code, err)
		}
	}
	// E) 若下一个节点仍是 auto，递归推进
	if next != nil && next.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(ctx, instanceID, wf, depth+1)
	}
	return nil
```

- [ ] **Step 5: 编译验证 workflow.go 改造**

Run: `cd gva/server && go build ./...`
Expected: 编译通过（`context`、`fmt` 已在 import 中）

- [ ] **Step 6: 创建 hooks.go（3 个内置 hook）**

创建 `gva/server/service/pmocker/hooks.go`：

```go
package pmocker

import (
	"context"
	"encoding/json"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// ScheduleBaselineHook 计划审批通过 → 生成计划基线
// 注册 key: workflowCode=plan_approval, nodeName=approve
type ScheduleBaselineHook struct{}

func (h *ScheduleBaselineHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}

func (h *ScheduleBaselineHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" {
		return nil
	}
	// entityID 即项目（eps_node）ID
	_, err := (&BaselineService{}).CreateBaseline(entityID, "schedule", nil, userIDFromCtx(ctx))
	return err
}

// CostBaselineHook 成本审批通过 → 生成成本基线
// 注册 key: workflowCode=cost_approval, nodeName=approve
type CostBaselineHook struct{}

func (h *CostBaselineHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}

func (h *CostBaselineHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" {
		return nil
	}
	_, err := (&BaselineService{}).CreateBaseline(entityID, "cost", nil, userIDFromCtx(ctx))
	return err
}

// ChangeApplyHook 变更关闭/批准 → 应用变更到目标实体 + 记录 change_logs
// 注册 key: workflowCode=change_request, nodeName=review（或 close）
type ChangeApplyHook struct{}

func (h *ChangeApplyHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}

func (h *ChangeApplyHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" && action != "close" {
		return nil
	}
	// entityID = change_request 实体 ID
	var change pmocker.PMEntity
	if err := global.GVA_DB.First(&change, entityID).Error; err != nil {
		return err
	}
	targetID := readAttrRef(entityID, "target_entity_id")
	if targetID == 0 {
		// 回退：change_target_ref
		targetID = readAttrRef(entityID, "change_target_ref")
	}
	if targetID == 0 {
		return nil // 无目标实体，跳过
	}
	fieldsJSON := readAttrString(entityID, "change_fields")
	if fieldsJSON == "" {
		return nil
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return err
	}
	userID := userIDFromCtx(ctx)
	changeReqID := entityID
	for field, newVal := range fields {
		oldVal := readAttrString(targetID, field)
		if oldVal == newVal {
			continue
		}
		if err := writeAttrString(targetID, field, newVal); err != nil {
			return err
		}
		if err := (&ChangeLogService{}).RecordChangeLog(targetID, field, oldVal, newVal, userID, &changeReqID); err != nil {
			return err
		}
	}
	return nil
}
```

> 说明：`ChangeLogService.RecordChangeLog(entityID, fieldKey, oldValue, newValue, changedBy, changeReqID)` 签名来自 Phase 1-2。变更请求实体需有 `target_entity_id`（ValRef）和 `change_fields`（ValJSON，形如 `{"estimated_hours":"120","end_date":"2026-05-20"}`）两个属性。

- [ ] **Step 7: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 8: 在插件 InitPMocker 中注册 hook**

修改 `gva/server/plugin/pmocker_core/initialize/init.go`，在初始化函数中（组织/业务种子数据加载之后）追加 hook 注册：

```go
import (
	pmockerSvc "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

// ... 在初始化函数末尾 ...

// 注册 NodeHook（workflowCode.nodeName 必须与工作流 YAML 定义一致）
wf := &pmockerSvc.ServiceGroupApp.WorkflowService
wf.RegisterNodeHook("plan_approval", "approve", &pmockerSvc.ScheduleBaselineHook{})
wf.RegisterNodeHook("cost_approval", "approve", &pmockerSvc.CostBaselineHook{})
wf.RegisterNodeHook("change_request", "review", &pmockerSvc.ChangeApplyHook{})
// task_complete_hook 在 Task 4 progress.go 中定义，此处一并注册
wf.RegisterNodeHook("task_workflow", "complete", &pmockerSvc.TaskCompleteHook{})
global.GVA_LOG.Info("pmocker NodeHook 已注册: plan_approval/cost_approval/change_request/task_workflow")
```

> 注意：`TaskCompleteHook` 在 Task 4 定义；若 Task 4 尚未实现，先注释掉该行，Task 4 完成后取消注释。`workflowCode`/`nodeName` 必须与各插件工作流 YAML 中的 `code` 和节点 `code` 完全一致。

- [ ] **Step 9: 提交**

```bash
git add gva/server/service/pmocker/workflow.go gva/server/service/pmocker/hooks.go gva/server/plugin/pmocker_core/initialize/init.go
git commit -m "feat(pmocker): 工作流NodeHook事件引擎(T10) 节点进出回调+3内置hook(基线/变更)"
```

---

## Task 4: 项目完成度自动汇总（spec T11）

**Files:**
- Create: `gva/server/service/pmocker/progress.go`
- Modify: `gva/server/service/pmocker/enter.go`
- Create: `gva/server/api/v1/pmocker/progress.go`
- Modify: `gva/server/api/v1/pmocker/enter.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/progress.js`

**Interfaces:**
- Consumes: `pmocker.PMEntity`（task/scope_item/risk）、`pmocker.PMWBSNode`、`VarianceService.CalcVariance`（Task 2）、共享属性工具（Task 1）、`NodeHook` 机制（Task 3）
- Produces:
  - `ProgressService.CalcByHours(projectID) (float64, error)`
  - `ProgressService.CalcByWBS(projectID) (float64, error)`
  - `ProgressService.CalcByCount(projectID) (float64, error)`
  - `ProgressService.CalcProjectProgress(projectID) (float64, error)`（按项目 `progress_algo` 配置派发）
  - `ProgressService.CalcHealthStatus(projectID) (string, error)`（red/yellow/green）
  - `TaskCompleteHook`（实现 `NodeHook`，任务完成→刷新项目进度+健康度）
  - REST: `GET /pmocker/progress/get?projectId=xxx`

- [ ] **Step 1: 创建 progress.go service（3 种算法 + 健康度 + hook）**

创建 `gva/server/service/pmocker/progress.go`：

```go
package pmocker

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ProgressService struct{}

// CalcByHours 工时加权平均（对标 MS Project）
// percent = sum(estimated_hours * progress/100) / sum(estimated_hours) * 100
func (s *ProgressService) CalcByHours(projectID uint) (float64, error) {
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	totalH, doneH := 0.0, 0.0
	for _, t := range tasks {
		h := readAttrDecimal(t.ID, "estimated_hours")
		p := readAttrDecimal(t.ID, "progress")
		totalH += h
		doneH += h * p / 100.0
	}
	if totalH == 0 {
		return 0, nil
	}
	return doneH / totalH * 100, nil
}

// CalcByWBS WBS 层级加权（对标 PMBOK，自底向上）
// 叶子节点进度 = 其下任务进度均值；父节点 = 子节点按 weight 加权平均
func (s *ProgressService) CalcByWBS(projectID uint) (float64, error) {
	var nodes []pmocker.PMWBSNode
	if err := global.GVA_DB.Where("project_id = ?", projectID).Find(&nodes).Error; err != nil {
		return 0, err
	}
	if len(nodes) == 0 {
		return 0, nil
	}
	childrenMap := map[uint][]pmocker.PMWBSNode{}
	var roots []pmocker.PMWBSNode
	for _, n := range nodes {
		if n.ParentID != nil {
			childrenMap[*n.ParentID] = append(childrenMap[*n.ParentID], n)
		} else {
			roots = append(roots, n)
		}
	}
	if len(roots) == 0 {
		return 0, nil
	}

	var calc func(n pmocker.PMWBSNode) float64
	calc = func(n pmocker.PMWBSNode) float64 {
		kids := childrenMap[n.ID]
		if len(kids) == 0 {
			// 叶子：聚合挂在该 scope_item 下的任务（task.parent_id = wbs.entity_id）
			var tasks []pmocker.PMEntity
			global.GVA_DB.Where("parent_id = ? AND entity_type = ?", n.EntityID, "task").Find(&tasks)
			if len(tasks) == 0 {
				return 0
			}
			sum := 0.0
			for _, t := range tasks {
				sum += readAttrDecimal(t.ID, "progress")
			}
			return sum / float64(len(tasks))
		}
		totalWeight, weighted := 0.0, 0.0
		for _, k := range kids {
			w := readAttrDecimal(k.EntityID, "weight")
			if w <= 0 {
				w = 1.0
			}
			weighted += calc(k) * w
			totalWeight += w
		}
		if totalWeight == 0 {
			return 0
		}
		return weighted / totalWeight
	}

	sum, cnt := 0.0, 0.0
	for _, r := range roots {
		sum += calc(r)
		cnt++
	}
	if cnt == 0 {
		return 0, nil
	}
	return sum / cnt, nil
}

// CalcByCount 任务数简单平均
// percent = count(done) / count(all) * 100
func (s *ProgressService) CalcByCount(projectID uint) (float64, error) {
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	done := 0
	for _, t := range tasks {
		if t.Status == "done" {
			done++
		}
	}
	return float64(done) / float64(len(tasks)) * 100, nil
}

// CalcProjectProgress 统一入口：读取项目 progress_algo 配置派发
func (s *ProgressService) CalcProjectProgress(projectID uint) (float64, error) {
	algo := readAttrString(projectID, "progress_algo")
	if algo == "" {
		algo = "hours"
	}
	switch algo {
	case "hours":
		return s.CalcByHours(projectID)
	case "wbs":
		return s.CalcByWBS(projectID)
	case "count":
		return s.CalcByCount(projectID)
	default:
		return s.CalcByHours(projectID)
	}
}

// CalcHealthStatus 基于进度偏差/成本偏差/风险数计算红黄绿健康度
func (s *ProgressService) CalcHealthStatus(projectID uint) (string, error) {
	rpt, err := (&VarianceService{}).CalcVariance(projectID)
	if err != nil {
		return "green", err
	}
	spi, cpi := 1.0, 1.0
	if rpt != nil {
		if rpt.PV > 0 {
			spi = rpt.SPI
		}
		if rpt.AC > 0 {
			cpi = rpt.CPI
		}
	}

	var risks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "risk").Find(&risks)
	highRisks := 0
	for _, r := range risks {
		sev := readAttrString(r.ID, "severity")
		if sev == "high" || sev == "高" {
			highRisks++
		}
	}

	if spi < 0.7 || cpi < 0.7 || highRisks >= 3 {
		return "red", nil
	}
	if spi < 0.9 || cpi < 0.9 || highRisks >= 1 {
		return "yellow", nil
	}
	return "green", nil
}

// TaskCompleteHook 任务完成 → 刷新项目进度 + 健康度
// 注册 key: workflowCode=task_workflow, nodeName=complete
type TaskCompleteHook struct{}

func (h *TaskCompleteHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}

func (h *TaskCompleteHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "complete" && action != "approve" && action != "done" {
		return nil
	}
	// entityID = task 实体；解析所属项目
	var task pmocker.PMEntity
	if err := global.GVA_DB.First(&task, entityID).Error; err != nil {
		return err
	}
	if task.ID == 0 || task.ProjectID == 0 {
		return nil
	}
	projectID := task.ProjectID
	percent, err := (&ProgressService{}).CalcProjectProgress(projectID)
	if err != nil {
		return fmt.Errorf("calc project progress failed: %w", err)
	}
	if err := writeAttrDecimal(projectID, "progress", percent); err != nil {
		return err
	}
	health, _ := (&ProgressService{}).CalcHealthStatus(projectID)
	if err := writeAttrString(projectID, "health_status", health); err != nil {
		return err
	}
	return nil
}
```

- [ ] **Step 2: 在 ServiceGroup 嵌入 ProgressService**

修改 `gva/server/service/pmocker/enter.go`，在 `ServiceGroup` 中新增 `ProgressService`：

```go
type ServiceGroup struct {
	EAVService
	WorkflowService
	RBACService
	RelationService
	TaskLinkService
	ChangeLogService
	CostLinkService
	TimeEntryService
	CostActualService
	BaselineService
	VarianceService
	ProgressService
}
```

- [ ] **Step 3: 编译验证（含 hook 注册）**

确认 Task 3 Step 8 中 `TaskCompleteHook` 注册行已取消注释，然后：

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 4: 创建 progress API handler**

创建 `gva/server/api/v1/pmocker/progress.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type ProgressApi struct{}

// Get 查询项目完成度 + 健康度
func (a *ProgressApi) Get(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	percent, err := service.ProgressService.CalcProjectProgress(projectID)
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	health, _ := service.ProgressService.CalcHealthStatus(projectID)
	algo := readAttrStringPublic(projectID, "progress_algo")
	if algo == "" {
		algo = "hours"
	}
	response.OkWithData(gin.H{
		"projectId":     projectID,
		"percent":       percent,
		"algo":          algo,
		"healthStatus":  health,
	}, c)
}
```

> 说明：handler 中需要读取属性，但 `readAttrString` 定义在 service 包（小写不可导出）。在 `api/v1/pmocker/progress.go` 中通过 `service.ProgressService` 暴露一个公开方法。在 Step 1 的 `progress.go` 末尾补一个公开方法：

在 `gva/server/service/pmocker/progress.go` 末尾追加：

```go
// GetProjectAlgo 返回项目的完成度算法配置（供 API 层调用）
func (s *ProgressService) GetProjectAlgo(projectID uint) string {
	algo := readAttrString(projectID, "progress_algo")
	if algo == "" {
		return "hours"
	}
	return algo
}
```

并将 `progress.go` API handler 中的 `algo := readAttrStringPublic(...)` 改为：

```go
	algo := service.ProgressService.GetProjectAlgo(projectID)
```

> 因此 Step 4 的 handler 最终为：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type ProgressApi struct{}

// Get 查询项目完成度 + 健康度
func (a *ProgressApi) Get(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	percent, err := service.ProgressService.CalcProjectProgress(projectID)
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	health, _ := service.ProgressService.CalcHealthStatus(projectID)
	algo := service.ProgressService.GetProjectAlgo(projectID)
	response.OkWithData(gin.H{
		"projectId":    projectID,
		"percent":      percent,
		"algo":         algo,
		"healthStatus": health,
	}, c)
}
```

- [ ] **Step 5: 在 ApiGroup 嵌入 ProgressApi**

修改 `gva/server/api/v1/pmocker/enter.go`：

```go
type ApiGroup struct {
	EAVApi
	BaselineApi
	VarianceApi
	ProgressApi
}
```

- [ ] **Step 6: 追加 progress 路由**

修改 `gva/server/router/pmocker/business.go`，追加：

```go
// 完成度（读操作）
{
	pg := private.Group("progress")
	pg.GET("get", apiGroup.ProgressApi.Get)
}
```

- [ ] **Step 7: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 8: 创建前端 progress.js API**

创建 `gva/web/src/api/pmocker/progress.js`：

```javascript
import service from '@/utils/request'

export const getProgress = (params) => {
  return service({ url: '/pmocker/progress/get', method: 'get', params })
}
```

- [ ] **Step 9: 端到端验证（任务完成→进度刷新）**

1. 启动后端：`cd gva/server && go run .`
2. 选一个项目（如 PROJ_A），通过 `GET /pmocker/progress/get?projectId=<id>` 记录当前 percent
3. 将项目下一个未完成任务的状态推进为 `done`，并触发其 task_workflow 的 `complete` 转移
4. 再次调用 `GET /pmocker/progress/get?projectId=<id>`，验证 percent 上升、`healthStatus` 为 green/yellow/red 之一

Expected: 项目 `progress` 与 `health_status` 属性被 `TaskCompleteHook` 自动刷新

- [ ] **Step 10: 提交**

```bash
git add gva/server/service/pmocker/progress.go gva/server/service/pmocker/enter.go gva/server/api/v1/pmocker/progress.go gva/server/api/v1/pmocker/enter.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/progress.js
git commit -m "feat(pmocker): 项目完成度自动汇总(T11) 3算法+健康度+task_complete_hook"
```

---

## Self-Review

### 1. Spec 覆盖检查

| Spec 要求（Phase 3-4） | 对应 Task | 状态 |
|------------------------|----------|------|
| T8 基线快照管理：审批通过自动生成计划/成本/范围基线 | Task 1（CreateBaseline）+ Task 3（ScheduleBaselineHook/CostBaselineHook） | ✅ |
| T8 基线列表 + 对比 | Task 1（ListBaselines/CompareBaseline + baseline.vue） | ✅ |
| T9 实际vs计划vs基线偏差图表 | Task 2（VarianceReport + variance.vue echarts） | ✅ |
| T9 SPI/CPI/进度偏差/成本偏差 | Task 2（PV/EV/AC → SV/CV/SPI/CPI） | ✅ |
| T9 超期/超支预警 | Task 2（GetAlerts：overdue/over_budget/high_risk） | ✅ |
| T10 NodeHook 接口 + 注册表 | Task 3（NodeHook 接口 + RegisterNodeHook + hooks map） | ✅ |
| T10 审批通过→生成基线 | Task 3（ScheduleBaselineHook/CostBaselineHook.OnLeave） | ✅ |
| T10 变更关闭→应用变更 + RecordChangeLog | Task 3（ChangeApplyHook.OnLeave） | ✅ |
| T10 在各插件 InitPMocker 注册 hook | Task 3 Step 8（init.go 注册） | ✅ |
| T11 3 种完成度算法 | Task 4（CalcByHours/CalcByWBS/CalcByCount） | ✅ |
| T11 统一入口按配置派发 | Task 4（CalcProjectProgress 读 progress_algo） | ✅ |
| T11 红黄绿健康度 | Task 4（CalcHealthStatus 基于 SPI/CPI/风险数） | ✅ |
| T11 任务完成→刷新项目进度（NodeHook） | Task 4（TaskCompleteHook.OnLeave） | ✅ |
| T11 API GET /pmocker/progress/get | Task 4 Step 6 | ✅ |
| spec 3.4 事件引擎架构（OnEnter/OnLeave + key=code.node） | Task 3 Step 1-2 | ✅ |
| spec 3.5 基线快照机制（全量 JSON + ChangeReqID） | Task 1（baselineSnapshot + PMBaseline.ChangeReqID） | ✅ |
| spec 3.6 3 种算法 + 健康度 | Task 4 | ✅ |

**覆盖完整，无遗漏。**

### 2. 占位符扫描

- 无 TBD/TODO/「后续实现」 ✅
- 所有 Step 含实际 Go/Vue 代码 ✅
- 每个 Task 有编译验证 Run + Expected ✅
- Task 4 Step 4 中先给出含 `readAttrStringPublic` 的版本，**紧接着在同一 Step 内修正为 `GetProjectAlgo` 公开方法**，最终 handler 代码自洽可编译 ✅（已消除中间态）

### 3. 类型一致性检查

- `NodeHook` 接口在 Task 3 定义（`OnEnter(ctx, entityID, nodeName)` + `OnLeave(ctx, entityID, nodeName, action)`），Task 4 `TaskCompleteHook` 与 Task 3 三个 hook 均实现该接口 ✅
- `RegisterNodeHook(workflowCode, nodeName, hook NodeHook)` 在 Task 3 定义，Task 3/Task 4 注册调用签名一致 ✅
- `BaselineService.CreateBaseline(projectID, baselineType, changeReqID *uint, createdBy uint) (uint, error)` 在 Task 1 定义，Task 3 `ScheduleBaselineHook`/`CostBaselineHook` 调用签名一致（`(&BaselineService{}).CreateBaseline(entityID, "schedule", nil, userID)`）✅
- `ChangeLogService.RecordChangeLog(entityID, fieldKey, oldValue, newValue, changedBy, changeReqID *uint)` 来自 Phase 1-2，Task 3 `ChangeApplyHook` 调用签名一致 ✅
- `VarianceService.CalcVariance(projectID) (*VarianceReport, error)` 在 Task 2 定义，Task 4 `CalcHealthStatus` 调用 `(&VarianceService{}).CalcVariance(projectID)` 签名一致 ✅
- 共享属性工具 `readAttrString`/`readAttrDecimal`/`readAttrRef`/`writeAttrString`/`writeAttrDecimal`/`attrValueString` 在 Task 1 `attr_helper.go` 定义，Task 2/3/4 均复用，签名一致 ✅
- `parseUintParam` 在 Task 1 `baseline.go`（api 包）定义，Task 2/Task 4 同包复用，无重复定义 ✅
- `hookUserIDKey`/`userIDFromCtx` 在 Task 3 `workflow.go`（service 包）定义，Task 3 `hooks.go` 与 Task 4 `progress.go` 同包复用 ✅
- `PMBaseline.ChangeReqID *uint` 依赖 Phase 1-2 已加字段，Task 1 CreateBaseline 直接赋值 ✅
- `PMWBSNode.EntityID/ParentID/ProjectID` 字段来自 `specialized.go`，Task 4 CalcByWBS 使用一致 ✅
- Service 嵌入：Task 1 加 `BaselineService`、Task 2 加 `VarianceService`、Task 4 加 `ProgressService`，三个 Step 的 `ServiceGroup` 定义为累积式一致 ✅
- Api 嵌入：Task 1 加 `BaselineApi`、Task 2 加 `VarianceApi`、Task 4 加 `ProgressApi`，累积式一致 ✅
- 路由路径：`/pmocker/baseline/*`、`/pmocker/variance/*`、`/pmocker/progress/*` 与前端 `baseline.js`/`variance.js`/`progress.js` URL 一致 ✅

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-02-m12-phase3-4-baseline-events.md`.

后续计划：
- Phase 5-6（T12-T16 报告+个人工作台）：`2026-08-02-m12-phase5-6-reports-workbench.md`（待编写）

Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
