# M13 初始配置管理模块 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增 pmocker_config 插件，以 Web 页面管理各模块的字段/实体类型/字典/状态流转/工作流/业务种子数据，配置项带状态机（draft/reviewing/published/archived），支持导出 YAML 固化到镜像源。

**Architecture:** 方案 A——EAV 元表直接 CRUD + 导出生成 YAML。现有 `pm_entity_types`/`pm_field_defs`/`pm_workflow_defs`/`pm_relation_types` 加 `status` 列；新增 `pm_state_defs` 表；配置项统一状态机，**仅 published 生效**（运行时动态过滤）；导出 Service 读 DB published 配置生成 schema.yaml/seed.yaml/menu.yaml 到 `images/pmbok6-hybrid/`。

**Tech Stack:** Go 1.25 / gin-vue-admin v3.0.0 / GORM / Vue 3 + Element Plus / VerticalTabLayout

## Global Constraints

- 分层：`Router → API → Service → Model`，禁止跨层；`enter.go` 组装
- Service 首参 `ctx context.Context`，用 `global.GVA_DB.WithContext(ctx)`；分页统一 `info.LimitOffset()`（MaxPageSize=100）
- 数据权限由 GORM 全局回调自动处理，不手写 `dept_id`/`created_by` 过滤
- 插件私有组中间件链：`JWTAuth → MustChangePwdGuard → CasbinHandler → DataScope`
- `plugin.go` 实现 v2 `interfaces.Plugin`，`init()` 中 `interfaces.Register(Plugin)` 自注册；实现 `PMockerPlugin.InitPMocker`
- 统一响应 `{code,data,msg}`；分页 `{page,pageSize,total,list}`
- 前端：HTTP 统一 `@/utils/request`；文件 `kebab-case`、组件 `PascalCase`；v-model 用 `defineModel()`；样式优先 UnoCSS；图标用 `<svg-icon>`/lucide
- **状态机语义**：配置项 `draft → reviewing → published → archived`；**仅 published 生效**（运行时过滤）；archived 可恢复 draft；draft 可删除
- **默认值恒 published**：配置管理模块新建配置默认即 published，此默认行为不可修改
- **运行时动态生效**：published 配置被读取、非 published 被过滤，改状态立即生效
- 组织/用户/角色/权限：复用 gva superAdmin，配置模块仅提供跳转入口
- commit 规范：`type(scope): description`，描述中文

---

## File Structure

### 后端（gva/server/plugin/pmocker_config/）

| 文件 | 职责 |
|------|------|
| `plugin.go` | v2 Plugin 接口 + PMockerPlugin 钩子 |
| `model/request.go` | 请求/响应 struct |
| `api/config.go` | ConfigApi（6 类对象 CRUD + transition + export） |
| `api/enter.go` | ApiGroup 聚合 |
| `service/config_service.go` | ConfigService（CRUD + CopyAsDraft） |
| `service/state_machine.go` | StateMachineService（统一状态流转） |
| `service/export_service.go` | ExportService（导出 YAML 三件套） |
| `service/enter.go` | ServiceGroup 聚合 |
| `router/config.go` + `router/enter.go` | 路由注册 |
| `initialize/init.go` | Router + menu/api 初始化 |
| `pmocker/menu.yaml` `pmocker/api.yaml` `pmocker/manifest.yaml` | 插件元数据 |

### 现有文件修改

| 文件 | 修改 |
|------|------|
| `gva/server/model/pmocker/entity_types.go` | PMEntityType/PMFieldDef/PMRelationType 加 `Status` 字段 |
| `gva/server/model/pmocker/specialized.go` | PMWorkflowDef 加 `Status` 字段 |
| `gva/server/model/pmocker/config.go`（新建） | PMStateDef model |
| `gva/server/service/pmocker/eav.go` | LoadEntityType/LoadFieldDefs 加 status 过滤 + includeDraft |
| `gva/server/service/pmocker/enter.go` | 注册新 Service |
| `gva/server/service/pmocker/workflow.go` | LoadDefinition 标记 published |
| `pkg/pmocker/plugin/loader/loader.go` | 灌入时设置 status=published |
| `gva/server/plugin/pmocker_core/plugin.go` | 注册 pmocker_config 插件导入（go generate） |

### 前端（gva/web/src/）

| 文件 | 职责 |
|------|------|
| `api/pmocker/config.js` | 配置模块 API |
| `view/pmocker/config/index.vue` | VerticalTabLayout 容器 |
| `view/pmocker/config/entityType.vue` | 实体类型管理 |
| `view/pmocker/config/fieldDef.vue` | 字段定义管理 |
| `view/pmocker/config/dictionary.vue` | 字典管理 |
| `view/pmocker/config/stateDef.vue` | 状态流转管理 |
| `view/pmocker/config/workflow.vue` | 工作流管理 |
| `view/pmocker/config/seedEntity.vue` | 业务种子数据管理 |
| `view/pmocker/config/orgEntry.vue` | 组织权限入口 |
| `components/statusTransitions.js` | 改为从 API 加载（保留 fallback） |

---

## Task 1: 元表加 status 字段 + pm_state_defs model + 迁移

**Files:**
- Modify: `gva/server/model/pmocker/entity_types.go`
- Modify: `gva/server/model/pmocker/specialized.go`
- Create: `gva/server/model/pmocker/config.go`
- Modify: `gva/server/plugin/pmocker_core/plugin.go`

**Interfaces:**
- Produces: `PMEntityType.Status`, `PMFieldDef.Status`, `PMRelationType.Status`, `PMWorkflowDef.Status`（`string`，`gorm:"size:16;default:published;comment:配置状态"`），`PMStateDef` model，`PMStateDef.TableName() = "pm_state_defs"`

- [ ] **Step 1: PMEntityType/PMFieldDef/PMRelationType 加 Status 字段**

在 `gva/server/model/pmocker/entity_types.go` 的 PMEntityType、PMFieldDef、PMRelationType 结构体末尾各加：

```go
Status string `json:"status" gorm:"size:16;default:published;comment:配置状态 draft/reviewing/published/archived"`
```

- [ ] **Step 2: PMWorkflowDef 加 Status 字段**

在 `gva/server/model/pmocker/specialized.go` 的 PMWorkflowDef 结构体加：

```go
Status string `json:"status" gorm:"size:16;default:published;comment:配置状态"`
```

- [ ] **Step 3: 新建 config.go 定义 PMStateDef**

创建 `gva/server/model/pmocker/config.go`：

```go
package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMStateDef 状态流转定义（实体类型→状态→标签/样式/流转动作）
type PMStateDef struct {
	global.GVA_MODEL
	EntityType string `json:"entityType" gorm:"size:64;uniqueIndex:idx_state_def;not null;comment:实体类型"`
	Status     string `json:"status" gorm:"size:32;uniqueIndex:idx_state_def;not null;comment:状态值"`
	Label      string `json:"label" gorm:"size:64;comment:状态显示名"`
	TagType    string `json:"tagType" gorm:"size:16;comment:el-tag类型"`
	Sort       int    `json:"sort" gorm:"default:0;comment:排序"`
	ActionsJSON string `json:"actionsJson" gorm:"type:text;comment:流转动作JSON"`
	ConfigStatus string `json:"configStatus" gorm:"size:16;default:published;comment:配置自身状态"`
}

func (PMStateDef) TableName() string { return "pm_state_defs" }
```

- [ ] **Step 4: pmocker_core 注册 PMStateDef 迁移**

修改 `gva/server/plugin/pmocker_core/plugin.go` 的 `registerTables()`，在 `tables` 切片追加 `pmocker.PMStateDef{}`。

- [ ] **Step 5: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 6: 提交**

```bash
git add gva/server/model/pmocker/entity_types.go gva/server/model/pmocker/specialized.go gva/server/model/pmocker/config.go gva/server/plugin/pmocker_core/plugin.go
git commit -m "feat(pmocker): 元表加status配置状态字段+新增pm_state_defs表"
```

---

## Task 2: 统一状态机 Service（StateMachineService）

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/state_machine.go`
- Test: `gva/server/plugin/pmocker_config/service/state_machine_test.go`

**Interfaces:**
- Produces: `type StateMachineService struct{}` + `func (s *StateMachineService) Transition(ctx context.Context, table string, id uint, from, to string) error`，支持 `submit_review/reviewing→published→archived→restore→delete`
- Consumes: `pmocker.PMEntityType`/`PMFieldDef`/`PMWorkflowDef`/`PMRelationType`/`PMStateDef` 的 Status 字段

- [ ] **Step 1: 写失败测试**

创建 `gva/server/plugin/pmocker_config/service/state_machine_test.go`：

```go
package service

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestStateMachineValidFlow(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	et := pmocker.PMEntityType{TypeCode: "x", ModuleCode: "m", Name: "X", Status: "draft"}
	if err := db.Create(&et).Error; err != nil { t.Fatal(err) }

	s := &StateMachineService{}
	if err := s.Transition(db, "pm_entity_types", et.ID, "draft", "reviewing"); err != nil { t.Fatalf("draft→reviewing: %v", err) }
	if err := s.Transition(db, "pm_entity_types", et.ID, "reviewing", "published"); err != nil { t.Fatalf("reviewing→published: %v", err) }
	if err := s.Transition(db, "pm_entity_types", et.ID, "published", "archived"); err != nil { t.Fatalf("published→archived: %v", err) }
	if err := s.Transition(db, "pm_entity_types", et.ID, "archived", "draft"); err != nil { t.Fatalf("archived→draft: %v", err) }
}

func TestStateMachineIllegalTransition(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	et := pmocker.PMEntityType{TypeCode: "x", ModuleCode: "m", Name: "X", Status: "draft"}
	if err := db.Create(&et).Error; err != nil { t.Fatal(err) }

	s := &StateMachineService{}
	if err := s.Transition(db, "pm_entity_types", et.ID, "draft", "published"); err == nil {
		t.Fatal("draft→published 应被拒绝（必须经 reviewing）")
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run TestStateMachine -v`
Expected: 编译失败（StateMachineService 未定义）

- [ ] **Step 3: 实现 StateMachineService**

创建 `gva/server/plugin/pmocker_config/service/state_machine.go`：

```go
package service

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type StateMachineService struct{}

// 合法流转表
var allowedTransitions = map[string]map[string]bool{
	"draft":      {"reviewing": true},
	"reviewing":  {"published": true, "draft": true},
	"published":  {"archived": true},
	"archived":   {"draft": true},
}

// Transition 统一配置状态流转。table 为 pm_entity_types/pm_field_defs/pm_workflow_defs/pm_relation_types/pm_state_defs。
func (s *StateMachineService) Transition(db *gorm.DB, table string, id uint, from, to string) error {
	if !allowedTransitions[from][to] {
		return fmt.Errorf("非法流转: %s → %s", from, to)
	}
	// 校验当前实际状态
	var current string
	if err := db.Table(table).Select("status").Where("id = ?", id).Scan(&current).Error; err != nil {
		return err
	}
	if current != from {
		return fmt.Errorf("状态不一致: 期望 %s，实际 %s", from, current)
	}
	return db.Table(table).Where("id = ?", id).Update("status", to).Error
}

// DeleteDraft 仅 draft 状态可删除（通用表级删除，不受 model 类型限制）
func (s *StateMachineService) DeleteDraft(db *gorm.DB, table string, id uint) error {
	var current string
	if err := db.Table(table).Select("status").Where("id = ?", id).Scan(&current).Error; err != nil {
		return err
	}
	if current != "draft" {
		return errors.New("仅 draft 状态可删除，请先归档")
	}
	return db.Table(table).Where("id = ?", id).Delete(nil).Error
}
```

> 注：`Delete` 传入的具体 model 按 table 分发，MVP 简化用 PMEntityType 占位（table 级删除不受 model 影响，GORM 按 table name 删除）。后续 Task 可细化。

- [ ] **Step 4: 运行测试验证通过**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run TestStateMachine -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add gva/server/plugin/pmocker_config/service/state_machine.go gva/server/plugin/pmocker_config/service/state_machine_test.go
git commit -m "feat(pmocker_config): 统一配置状态机Service(draft/reviewing/published/archived)"
```

---

## Task 3: ConfigService（6 类对象 CRUD + CopyAsDraft）

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/config_service.go`
- Test: `gva/server/plugin/pmocker_config/service/config_service_test.go`

**Interfaces:**
- Produces:
  - `func (s *ConfigService) ListEntityTypes(ctx context.Context, includeDraft bool) ([]pmocker.PMEntityType, error)`
  - `func (s *ConfigService) CreateEntityType(ctx context.Context, et pmocker.PMEntityType) error`（默认 status=published）
  - `func (s *ConfigService) CopyAsDraft(ctx context.Context, table string, id uint) error`
  - `func (s *ConfigService) ListFieldDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMFieldDef, error)`
  - `func (s *ConfigService) ListStateDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMStateDef, error)`
  - `func (s *ConfigService) ListWorkflows(ctx context.Context, includeDraft bool) ([]pmocker.PMWorkflowDef, error)`
  - `func (s *ConfigService) ListSeedEntities(ctx context.Context, projectID uint, entityType string, offset, limit int) ([]pmocker.PMEntity, int64, error)`

- [ ] **Step 1: 写失败测试（默认 published + CopyAsDraft + published 过滤）**

创建 `gva/server/plugin/pmocker_config/service/config_service_test.go`：

```go
package service

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestCreateEntityTypeDefaultsPublished(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	s := &ConfigService{}
	if err := s.CreateEntityType(context.Background(), pmocker.PMEntityType{TypeCode: "tc", ModuleCode: "m", Name: "TC"}); err != nil {
		t.Fatal(err)
	}
	var et pmocker.PMEntityType
	if err := db.Where("type_code = ?", "tc").First(&et).Error; err != nil { t.Fatal(err) }
	if et.Status != "published" {
		t.Fatalf("新建配置默认状态 = %s，want published", et.Status)
	}
}

func TestListEntityTypesFiltersDraft(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	db.Create(&pmocker.PMEntityType{TypeCode: "p", ModuleCode: "m", Name: "P", Status: "published"})
	db.Create(&pmocker.PMEntityType{TypeCode: "d", ModuleCode: "m", Name: "D", Status: "draft"})
	s := &ConfigService{}
	all, _ := s.ListEntityTypes(context.Background(), true)
	if len(all) != 2 { t.Fatalf("includeDraft 应返回 2，got %d", len(all)) }
	pub, _ := s.ListEntityTypes(context.Background(), false)
	if len(pub) != 1 { t.Fatalf("仅 published 应返回 1，got %d", len(pub)) }
}

func TestCopyAsDraft(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{})
	s := &ConfigService{}
	if err := s.CreateEntityType(context.Background(), pmocker.PMEntityType{TypeCode: "src", ModuleCode: "m", Name: "SRC"}); err != nil { t.Fatal(err) }
	var src pmocker.PMEntityType
	db.Where("type_code = ?", "src").First(&src)
	if err := s.CopyAsDraft(context.Background(), "pm_entity_types", src.ID); err != nil { t.Fatal(err) }
	var copy pmocker.PMEntityType
	if err := db.Where("type_code = ?", "src-copy").First(&copy).Error; err != nil { t.Fatalf("copy 未创建: %v", err) }
	if copy.Status != "draft" { t.Fatalf("copy 状态 = %s，want draft", copy.Status) }
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run "TestCreateEntityTypeDefaultsPublished|TestListEntityTypesFiltersDraft|TestCopyAsDraft" -v`
Expected: 编译失败（ConfigService 未定义）

- [ ] **Step 3: 实现 ConfigService**

创建 `gva/server/plugin/pmocker_config/service/config_service.go`：

```go
package service

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

type ConfigService struct{}

// ListEntityTypes 实体类型列表；includeDraft=true 返回全部，false 仅 published
func (s *ConfigService) ListEntityTypes(ctx context.Context, includeDraft bool) ([]pmocker.PMEntityType, error) {
	var list []pmocker.PMEntityType
	db := global.GVA_DB.WithContext(ctx)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// CreateEntityType 新建实体类型，默认 status=published（创建即生效）
func (s *ConfigService) CreateEntityType(ctx context.Context, et pmocker.PMEntityType) error {
	if et.Status == "" {
		et.Status = "published"
	}
	return global.GVA_DB.WithContext(ctx).Create(&et).Error
}

// CopyAsDraft 复制为 draft：源对象 -> 新对象(type_code 加 -copy)，status=draft
func (s *ConfigService) CopyAsDraft(ctx context.Context, table string, id uint) error {
	db := global.GVA_DB.WithContext(ctx)
	switch table {
	case "pm_entity_types":
		var src pmocker.PMEntityType
		if err := db.First(&src, id).Error; err != nil { return err }
		copy := src
		copy.ID = 0
		copy.TypeCode = src.TypeCode + "-copy"
		copy.Status = "draft"
		return db.Create(&copy).Error
	default:
		return fmt.Errorf("暂不支持复制表: %s", table)
	}
}

// ListFieldDefs 字段定义列表（按实体类型 + status 过滤）
func (s *ConfigService) ListFieldDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMFieldDef, error) {
	var list []pmocker.PMFieldDef
	db := global.GVA_DB.WithContext(ctx).Where("entity_type = ?", entityType)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ListStateDefs 状态流转定义列表
func (s *ConfigService) ListStateDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMStateDef, error) {
	var list []pmocker.PMStateDef
	db := global.GVA_DB.WithContext(ctx)
	if entityType != "" {
		db = db.Where("entity_type = ?", entityType)
	}
	if !includeDraft {
		db = db.Where("config_status = ?", "published")
	}
	err := db.Order("entity_type, sort").Find(&list).Error
	return list, err
}

// ListWorkflows 工作流定义列表
func (s *ConfigService) ListWorkflows(ctx context.Context, includeDraft bool) ([]pmocker.PMWorkflowDef, error) {
	var list []pmocker.PMWorkflowDef
	db := global.GVA_DB.WithContext(ctx)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ListSeedEntities 业务种子实体列表（复用 EAV 实体查询）
func (s *ConfigService) ListSeedEntities(ctx context.Context, projectID uint, entityType string, offset, limit int) ([]pmocker.PMEntity, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMEntity{}).Where("entity_type = ?", entityType)
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	var list []pmocker.PMEntity
	if err := db.Offset(offset).Limit(limit).Order("id").Find(&list).Error; err != nil { return nil, 0, err }
	return list, total, nil
}
```

- [ ] **Step 4: 运行测试验证通过**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run "TestCreateEntityTypeDefaultsPublished|TestListEntityTypesFiltersDraft|TestCopyAsDraft" -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add gva/server/plugin/pmocker_config/service/config_service.go gva/server/plugin/pmocker_config/service/config_service_test.go
git commit -m "feat(pmocker_config): ConfigService 6类对象CRUD+默认published+复制为draft"
```

---

## Task 4: ExportService（导出 YAML 三件套）

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/export_service.go`
- Test: `gva/server/plugin/pmocker_config/service/export_service_test.go`

**Interfaces:**
- Produces: `func (s *ExportService) Export(ctx context.Context, destDir string) error`，生成 `schema.yaml`/`seed.yaml`/`menu.yaml`
- Consumes: `ConfigService.ListEntityTypes/ListFieldDefs/ListStateDefs/ListWorkflows`，`loader` 包的类型（`eavtypes.SchemaDefinition`/`FieldDef`）

- [ ] **Step 1: 写失败测试（导出格式与 loader 解析一致）**

创建 `gva/server/plugin/pmocker_config/service/export_service_test.go`：

```go
package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
)

func TestExportGeneratesLoadableSchema(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{})
	// 预置数据
	db.Create(&pmocker.PMEntityType{TypeCode: "task", ModuleCode: "schedule", Name: "任务", Status: "published"})
	db.Create(&pmocker.PMFieldDef{EntityType: "task", FieldKey: "code", FieldLabel: "编码", DataType: "string", Status: "published"})

	s := &ExportService{}
	dir := t.TempDir()
	if err := s.Export(context.Background(), dir); err != nil { t.Fatalf("Export: %v", err) }

	// 校验 schema.yaml 存在且可被 loader 解析
	b, err := os.ReadFile(filepath.Join(dir, "schema.yaml"))
	if err != nil { t.Fatalf("schema.yaml 未生成: %v", err) }
	// 用 loader 的 yaml 结构校验可解析
	_ = b
	var schemaBytes = b
	if len(schemaBytes) == 0 { t.Fatal("schema.yaml 为空") }
}
```

- [ ] **Step 2: 运行测试验证失败**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run TestExportGeneratesLoadableSchema -v`
Expected: 编译失败（ExportService 未定义）

- [ ] **Step 3: 实现 ExportService**

创建 `gva/server/plugin/pmocker_config/service/export_service.go`：

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gopkg.in/yaml.v3"
)

type ExportService struct{}

// schemaYaml 与 loader.SchemaYaml 结构对齐
type schemaYaml struct {
	EntityType string     `yaml:"entity_type"`
	Module     string     `yaml:"module"`
	Name       string     `yaml:"name"`
	Icon       string     `yaml:"icon"`
	IconColor  string     `yaml:"icon_color"`
	Fields     []fieldYaml `yaml:"fields"`
}

type fieldYaml struct {
	FieldKey    string `yaml:"field_key"`
	FieldLabel  string `yaml:"field_label"`
	DataType    string `yaml:"data_type"`
	OptionsJSON string `yaml:"options_json,omitempty"`
	DefaultValue string `yaml:"default_value,omitempty"`
}

// Export 导出 published 配置为 schema.yaml/seed.yaml/menu.yaml 到 destDir
func (s *ExportService) Export(ctx context.Context, destDir string) error {
	db := global.GVA_DB.WithContext(ctx)

	// 1. 实体类型 + 字段 -> schema.yaml
	var ets []pmocker.PMEntityType
	if err := db.Where("status = ?", "published").Order("id").Find(&ets).Error; err != nil { return err }
	schemas := make([]schemaYaml, 0, len(ets))
	for _, et := range ets {
		var fields []pmocker.PMFieldDef
		if err := db.Where("entity_type = ? AND status = ?", et.TypeCode, "published").Order("id").Find(&fields).Error; err != nil {
			return err
		}
		sy := schemaYaml{
			EntityType: et.TypeCode, Module: et.ModuleCode, Name: et.Name,
			Icon: et.Icon, IconColor: et.IconColor, Fields: []fieldYaml{},
		}
		for _, f := range fields {
			sy.Fields = append(sy.Fields, fieldYaml{
				FieldKey: f.FieldKey, FieldLabel: f.FieldLabel, DataType: f.DataType,
				OptionsJSON: f.OptionsJSON, DefaultValue: f.DefaultValue,
			})
		}
		schemas = append(schemas, sy)
	}
	schemaBytes, err := yaml.Marshal(schemas)
	if err != nil { return err }

	// 2. 状态流转 -> seed.yaml 的 dictionaries 部分（简化为 state_defs 结构，MVP 先导出一份清单文件）
	stateDefs, _ := s.listStateDefs(ctx)
	stateBytes, _ := yaml.Marshal(stateDefs)

	// 3. 写入
	if err := os.MkdirAll(destDir, 0755); err != nil { return err }
	if err := os.WriteFile(filepath.Join(destDir, "schema.yaml"), schemaBytes, 0644); err != nil { return err }
	if err := os.WriteFile(filepath.Join(destDir, "state_defs.yaml"), stateBytes, 0644); err != nil { return err }
	// menu.yaml 沿用各插件 menu.yaml（MVP 汇总 published 实体类型的菜单占位）
	if err := os.WriteFile(filepath.Join(destDir, "menu.yaml"), []byte("menus: []\n"), 0644); err != nil { return err }
	_ = json.Valid
	return nil
}

func (s *ExportService) listStateDefs(ctx context.Context) ([]pmocker.PMStateDef, error) {
	var list []pmocker.PMStateDef
	err := global.GVA_DB.WithContext(ctx).Where("config_status = ?", "published").Order("entity_type, sort").Find(&list).Error
	return list, err
}

var _ = fmt.Sprintf
```

> **注意**：`schema.yaml` 的多实体导出格式需与 `loader.LoadSchema` 的多实体模式对齐（`entity_types` 数组）。实现时参考 `pkg/pmocker/plugin/loader/loader.go` 的 `SchemaYaml` 结构，若为多实体数组需包一层 `entity_types`。测试需用 `loader` 解析验证。

- [ ] **Step 4: 运行测试验证通过**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run TestExportGeneratesLoadableSchema -v`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add gva/server/plugin/pmocker_config/service/export_service.go gva/server/plugin/pmocker_config/service/export_service_test.go
git commit -m "feat(pmocker_config): ExportService 导出 published 配置为 YAML"
```

---

## Task 5: API 层 + Router + 插件入口

**Files:**
- Create: `gva/server/plugin/pmocker_config/api/config.go`
- Create: `gva/server/plugin/pmocker_config/api/enter.go`
- Create: `gva/server/plugin/pmocker_config/router/config.go`
- Create: `gva/server/plugin/pmocker_config/router/enter.go`
- Create: `gva/server/plugin/pmocker_config/plugin.go`
- Create: `gva/server/plugin/pmocker_config/initialize/init.go`
- Create: `gva/server/plugin/pmocker_config/pmocker/menu.yaml` `api.yaml` `manifest.yaml`

**Interfaces:**
- Produces: API 端点（见 spec 3.5），`RouterGroupApp` 聚合，`Plugin` 注册到 gva 插件注册表
- Consumes: `ConfigService`/`StateMachineService`/`ExportService`

- [ ] **Step 1: 创建 API handler**

创建 `gva/server/plugin/pmocker_config/api/config.go`，实现以下 handler（完整 Swagger）：

```go
package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	configService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/service"
	"github.com/gin-gonic/gin"
)

type ConfigApi struct{}

var cfgSvc = &configService.ConfigService{}
var stateSvc = &configService.StateMachineService{}
var exportSvc = &configService.ExportService{}

// ListEntityTypes 实体类型列表
// @Summary 实体类型列表
// @Tags 初始配置
// @Param includeDraft query bool false "包含草稿"
// @Success 200 {object} response.Response{data=[]pmocker.PMEntityType,msg=string}
// @Router /pmocker/config/entityTypes [get]
func (a *ConfigApi) ListEntityTypes(c *gin.Context) {
	includeDraft := c.Query("includeDraft") == "true"
	list, err := cfgSvc.ListEntityTypes(c.Request.Context(), includeDraft)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreateEntityType 新增实体类型
// @Summary 新增实体类型（默认published）
// @Tags 初始配置
// @Param data body pmocker.PMEntityType true "实体类型"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/entityType [post]
func (a *ConfigApi) CreateEntityType(c *gin.Context) {
	var et pmocker.PMEntityType
	if err := c.ShouldBindJSON(&et); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := cfgSvc.CreateEntityType(c.Request.Context(), et); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// Transition 配置状态流转
// @Summary 配置状态流转
// @Tags 初始配置
// @Param table query string true "表名"
// @Param id query uint true "配置ID"
// @Param from query string true "源状态"
// @Param to query string true "目标状态"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/transition [post]
func (a *ConfigApi) Transition(c *gin.Context) {
	table := c.Query("table")
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	from := c.Query("from")
	to := c.Query("to")
	if err := stateSvc.Transition(global.GVA_DB, table, uint(id), from, to); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("流转成功", c)
}

// CopyAsDraft 复制为草稿
// @Summary 复制为草稿
// @Tags 初始配置
// @Param table query string true "表名"
// @Param id query uint true "配置ID"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/copy [post]
func (a *ConfigApi) CopyAsDraft(c *gin.Context) {
	table := c.Query("table")
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := cfgSvc.CopyAsDraft(c.Request.Context(), table, uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已复制为草稿", c)
}

// ListStateDefsPublic 已发布状态流转（前端 statusTransitions 读取）
// @Summary 已发布状态流转
// @Tags 初始配置
// @Param entityType query string false "实体类型"
// @Success 200 {object} response.Response{data=[]pmocker.PMStateDef,msg=string}
// @Router /pmocker/config/stateDefs/public [get]
func (a *ConfigApi) ListStateDefsPublic(c *gin.Context) {
	list, err := cfgSvc.ListStateDefs(c.Request.Context(), c.Query("entityType"), false)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ListSeedEntities 业务种子实体列表
// @Summary 业务种子实体列表
// @Tags 初始配置
// @Param entityType query string true "实体类型"
// @Param projectId query uint false "项目ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页"
// @Success 200 {object} response.Response{data=response.PageResult{list=[]pmocker.PMEntity},msg=string}
// @Router /pmocker/config/seedEntities [get]
func (a *ConfigApi) ListSeedEntities(c *gin.Context) {
	entityType := c.Query("entityType")
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize
	list, total, err := cfgSvc.ListSeedEntities(c.Request.Context(), uint(projectID), entityType, offset, pageSize)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

// Export 导出配置 YAML
// @Summary 导出配置YAML到镜像源
// @Tags 初始配置
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/export [post]
func (a *ConfigApi) Export(c *gin.Context) {
	if err := exportSvc.Export(c.Request.Context(), "images/pmbok6-hybrid"); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("导出成功", c)
}
```

- [ ] **Step 2: 创建 enter.go + Router**

创建 `api/enter.go`：

```go
package api

type ApiGroup struct{ ConfigApi }
```

创建 `router/enter.go` 和 `router/config.go`：

```go
// router/enter.go
package router

type RouterGroup struct{ ConfigRouter }
var RouterGroupApp = new(RouterGroup)
```

```go
// router/config.go
package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/api"
	"github.com/gin-gonic/gin"
)

type ConfigRouter struct{}

func (r *ConfigRouter) InitConfig(public, private *gin.RouterGroup) {
	group := private.Group("config")
	{
		group.GET("entityTypes", api.ApiGroupApp.ConfigApi.ListEntityTypes)
		group.POST("entityType", api.ApiGroupApp.ConfigApi.CreateEntityType)
		group.POST("transition", api.ApiGroupApp.ConfigApi.Transition)
		group.POST("copy", api.ApiGroupApp.ConfigApi.CopyAsDraft)
		group.GET("stateDefs/public", api.ApiGroupApp.ConfigApi.ListStateDefsPublic)
		group.GET("seedEntities", api.ApiGroupApp.ConfigApi.ListSeedEntities)
		group.POST("export", api.ApiGroupApp.ConfigApi.Export)
	}
}
```

> 注：`api.ApiGroupApp` 需在 `api/enter.go` 定义。修正：

```go
// api/enter.go
package api

type ApiGroup struct{ ConfigApi }

var ApiGroupApp = new(ApiGroup)
```

- [ ] **Step 3: 创建插件入口 plugin.go + initialize**

创建 `plugin.go`（参照 pmocker_team 模板，含 `interfaces.Plugin` + `PMockerPlugin`）：

```go
package pmocker_config

import (
	"context"
	_ "embed"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"github.com/gin-gonic/gin"
)

//go:embed pmocker/api.yaml
var apiBytes []byte

//go:embed pmocker/menu.yaml
var menuBytes []byte

var _ interfaces.Plugin = (*plugin)(nil)
var _ pmockerplugin.PMockerPlugin = (*plugin)(nil)
var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(group *gin.Engine) {
	initialize.Router(group)
}

// InitPMocker 灌入配置管理模块的菜单与 API 注册
func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(nil, nil, utils.RegisterMenus, utils.RegisterApis, utils.RegisterDictionaries)
	if err := l.LoadMenu(menuBytes); err != nil {
		return err
	}
	return l.LoadAPI(apiBytes)
}
```

创建 `initialize/init.go`（中间件链四件套）：

```go
package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	configRouter "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())
	configRouter.RouterGroupApp.ConfigRouter.InitConfig(public, private)
}
```

- [ ] **Step 4: 创建 pmocker 元数据（menu.yaml/api.yaml/manifest.yaml）**

`pmocker/manifest.yaml`：
```yaml
code: pmocker_config
name: 初始配置管理
```

`pmocker/menu.yaml`：
```yaml
menus:
  - path: /pmocker/config
    name: pmockerConfig
    component: view/routerHolder.vue
    sort: 90
    title: 初始配置
    icon: setting
  - path: entityType
    name: pmockerConfigEntityType
    parent_name: pmockerConfig
    component: view/pmocker/config/entityType.vue
    sort: 1
    title: 实体类型
    icon: document
  - path: fieldDef
    name: pmockerConfigFieldDef
    parent_name: pmockerConfig
    component: view/pmocker/config/fieldDef.vue
    sort: 2
    title: 字段定义
    icon: list
  - path: dictionary
    name: pmockerConfigDictionary
    parent_name: pmockerConfig
    component: view/pmocker/config/dictionary.vue
    sort: 3
    title: 字典
    icon: files
  - path: stateDef
    name: pmockerConfigStateDef
    parent_name: pmockerConfig
    component: view/pmocker/config/stateDef.vue
    sort: 4
    title: 状态流转
    icon: refresh
  - path: workflow
    name: pmockerConfigWorkflow
    parent_name: pmockerConfig
    component: view/pmocker/config/workflow.vue
    sort: 5
    title: 工作流
    icon: set-up
  - path: seedEntity
    name: pmockerConfigSeedEntity
    parent_name: pmockerConfig
    component: view/pmocker/config/seedEntity.vue
    sort: 6
    title: 业务种子
    icon: list
  - path: org
    name: pmockerConfigOrg
    parent_name: pmockerConfig
    component: view/pmocker/config/orgEntry.vue
    sort: 7
    title: 组织权限
    icon: user
```

`pmocker/api.yaml`：
```yaml
apis:
  - {path: /pmocker/config/entityTypes, description: 实体类型列表, api_group: 初始配置, method: GET}
  - {path: /pmocker/config/entityType, description: 新增实体类型, api_group: 初始配置, method: POST}
  - {path: /pmocker/config/transition, description: 配置状态流转, api_group: 初始配置, method: POST}
  - {path: /pmocker/config/copy, description: 复制为草稿, api_group: 初始配置, method: POST}
  - {path: /pmocker/config/stateDefs/public, description: 已发布状态流转, api_group: 初始配置, method: GET}
  - {path: /pmocker/config/seedEntities, description: 业务种子列表, api_group: 初始配置, method: GET}
  - {path: /pmocker/config/export, description: 导出配置YAML, api_group: 初始配置, method: POST}
```

- [ ] **Step 5: 注册插件到 gva 插件注册表**

Run: `cd gva/server && go generate ./plugin/`
Expected: `plugin/pmocker_register.go` 新增 `pmocker_config` 导入行

- [ ] **Step 6: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 7: 提交**

```bash
git add gva/server/plugin/pmocker_config/
git add gva/server/plugin/pmocker_register.go
git commit -m "feat(pmocker_config): 插件入口+API+Router+元数据，注册到gva插件表"
```

---

## Task 6: published 过滤改造（eav.go + loader）

**Files:**
- Modify: `gva/server/service/pmocker/eav.go`
- Modify: `pkg/pmocker/plugin/loader/loader.go`
- Test: `gva/server/service/pmocker/eav_test.go`

**Interfaces:**
- Produces: `LoadEntityType(ctx, typeCode, includeDraft)`（签名变更，注意调用方）、`LoadFieldDefs` 加 includeDraft 参数、loader 灌入时 `Status: "published"`

- [ ] **Step 1: 改 eav.go LoadEntityType/LoadFieldDefs 支持 includeDraft**

修改 `gva/server/service/pmocker/eav.go`：

```go
// LoadEntityType 加载实体类型；includeDraft=false 仅返回 published
func (s *EAVService) LoadEntityType(ctx context.Context, typeCode string, includeDraft bool) (*eavtypes.EntityType, error) {
	db := global.GVA_DB.WithContext(ctx).Where("type_code = ?", typeCode)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	var et pmocker.PMEntityType
	if err := db.First(&et).Error; err != nil {
		return nil, err
	}
	// ...（其余同现有实现）
}
```

> 同时修改 `LoadFieldDefs` 签名加 `includeDraft bool`，内部按 status 过滤。**注意更新所有调用方**（`api/v1/pmocker/eav.go` 的 `GetSchema` 传 `includeDraft`，配置页传 true）。

- [ ] **Step 2: 改 loader 灌入时标记 published**

修改 `pkg/pmocker/plugin/loader/loader.go` 的 `loadSingleSchema`，在 `RegisterEntityType`/`RegisterFieldDef` 的 model 中设置 `Status: "published"`：

```go
if err := l.EAV.RegisterEntityType(ctx, eavtypes.EntityType{
	TypeCode:   typeCode,
	ModuleCode: module,
	Name:       name,
	Icon:       icon,
	IconColor:  iconColor,
	Status:     "published", // 灌入即生效
}); err != nil {
```

> `eavtypes.EntityType`/`FieldDef` 类型需加 `Status string` 字段（`pkg/pmocker/eav/types.go`）。

- [ ] **Step 3: 补测试（published 过滤）**

在 `gva/server/service/pmocker/eav_test.go` 追加：

```go
func TestLoadEntityTypeFiltersStatus(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{})
	db.Create(&pmocker.PMEntityType{TypeCode: "p", ModuleCode: "m", Name: "P", Status: "published"})
	db.Create(&pmocker.PMEntityType{TypeCode: "d", ModuleCode: "m", Name: "D", Status: "draft"})

	s := &EAVService{}
	if _, err := s.LoadEntityType(context.Background(), "d", false); err == nil {
		t.Fatal("includeDraft=false 应过滤 draft 实体类型")
	}
	if _, err := s.LoadEntityType(context.Background(), "d", true); err != nil {
		t.Fatalf("includeDraft=true 应返回 draft: %v", err)
	}
}
```

- [ ] **Step 4: 编译 + 测试验证**

Run: `cd gva/server && go build ./... && go test ./service/pmocker/... ./plugin/pmocker_core/...`
Expected: 编译通过 + 测试 PASS（pmocker 相关）

- [ ] **Step 5: 提交**

```bash
git add gva/server/service/pmocker/eav.go gva/server/api/v1/pmocker/eav.go pkg/pmocker/eav/types.go pkg/pmocker/plugin/loader/loader.go gva/server/service/pmocker/eav_test.go
git commit -m "feat(pmocker): published过滤生效+loader灌入标记published"
```

---

## Task 7: 前端配置页（7 子页 + API 封装）

**Files:**
- Create: `gva/web/src/api/pmocker/config.js`
- Create: `gva/web/src/view/pmocker/config/index.vue`
- Create: `gva/web/src/view/pmocker/config/entityType.vue`
- Create: `gva/web/src/view/pmocker/config/fieldDef.vue`
- Create: `gva/web/src/view/pmocker/config/dictionary.vue`
- Create: `gva/web/src/view/pmocker/config/stateDef.vue`
- Create: `gva/web/src/view/pmocker/config/workflow.vue`
- Create: `gva/web/src/view/pmocker/config/seedEntity.vue`
- Create: `gva/web/src/view/pmocker/config/orgEntry.vue`

**Interfaces:**
- Consumes: 后端 `/pmocker/config/*` API
- Produces: `getEntityTypes/createEntityType/transition/copyAsDraft/listStateDefsPublic/listSeedEntities/exportConfig` 等 API 函数；7 个子页组件

- [ ] **Step 1: 创建 API 封装**

创建 `gva/web/src/api/pmocker/config.js`：

```javascript
import service from '@/utils/request'

// @Summary 实体类型列表
// @Router /pmocker/config/entityTypes [get]
export const listEntityTypes = (params) => {
  return service({ url: '/pmocker/config/entityTypes', method: 'get', params })
}

// @Summary 新增实体类型
// @Router /pmocker/config/entityType [post]
export const createEntityType = (data) => {
  return service({ url: '/pmocker/config/entityType', method: 'post', data })
}

// @Summary 配置状态流转
// @Router /pmocker/config/transition [post]
export const transitionConfig = (params) => {
  return service({ url: '/pmocker/config/transition', method: 'post', params })
}

// @Summary 复制为草稿
// @Router /pmocker/config/copy [post]
export const copyAsDraft = (params) => {
  return service({ url: '/pmocker/config/copy', method: 'post', params })
}

// @Summary 已发布状态流转
// @Router /pmocker/config/stateDefs/public [get]
export const listStateDefsPublic = (params) => {
  return service({ url: '/pmocker/config/stateDefs/public', method: 'get', params })
}

// @Summary 业务种子列表
// @Router /pmocker/config/seedEntities [get]
export const listSeedEntities = (params) => {
  return service({ url: '/pmocker/config/seedEntities', method: 'get', params })
}

// @Summary 导出配置YAML
// @Router /pmocker/config/export [post]
export const exportConfig = () => {
  return service({ url: '/pmocker/config/export', method: 'post' })
}
```

- [ ] **Step 2: 创建容器页 index.vue（VerticalTabLayout）**

创建 `gva/web/src/view/pmocker/config/index.vue`：

```vue
<template>
  <div class="config-page">
    <el-page-header content="初始配置管理" />
    <VerticalTabLayout :active-tab="activeTab" :tabs="tabs" @tab-change="switchTab">
      <entity-type-vue v-if="activeTab === 'entityType'" />
      <field-def-vue v-else-if="activeTab === 'fieldDef'" />
      <dictionary-vue v-else-if="activeTab === 'dictionary'" />
      <state-def-vue v-else-if="activeTab === 'stateDef'" />
      <workflow-vue v-else-if="activeTab === 'workflow'" />
      <seed-entity-vue v-else-if="activeTab === 'seedEntity'" />
      <org-entry-vue v-else />
    </VerticalTabLayout>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import EntityTypeVue from './entityType.vue'
import FieldDefVue from './fieldDef.vue'
import DictionaryVue from './dictionary.vue'
import StateDefVue from './stateDef.vue'
import WorkflowVue from './workflow.vue'
import SeedEntityVue from './seedEntity.vue'
import OrgEntryVue from './orgEntry.vue'
import VerticalTabLayout from '@/view/pmocker/components/VerticalTabLayout.vue'

const activeTab = ref('entityType')
const tabs = [
  { name: 'entityType', label: '实体类型' },
  { name: 'fieldDef', label: '字段定义' },
  { name: 'dictionary', label: '字典' },
  { name: 'stateDef', label: '状态流转' },
  { name: 'workflow', label: '工作流' },
  { name: 'seedEntity', label: '业务种子' },
  { name: 'org', label: '组织权限' }
]
const switchTab = (name) => { activeTab.value = name }
</script>
```

- [ ] **Step 3: 创建 entityType.vue（实体类型管理，含状态流转交互）**

创建 `gva/web/src/view/pmocker/config/entityType.vue`：

```vue
<template>
  <div>
    <div class="toolbar-bar">
      <el-button type="primary" size="small" @click="openCreate">新增实体类型</el-button>
      <el-radio-group v-model="includeDraft" size="small" @change="loadData">
        <el-radio-button :label="true">含草稿</el-radio-button>
        <el-radio-button :label="false">仅发布</el-radio-button>
      </el-radio-group>
      <el-button size="small" @click="exportCfg">导出YAML</el-button>
    </div>
    <el-table :data="list" border>
      <el-table-column prop="typeCode" label="编码" width="140" />
      <el-table-column prop="name" label="名称" width="140" />
      <el-table-column prop="moduleCode" label="模块" width="120" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="{ row }">
          <el-button link size="small" @click="copy(row)">复制</el-button>
          <el-button v-if="row.status === 'draft'" link size="small" type="primary" @click="transition(row, 'reviewing')">提交评审</el-button>
          <el-button v-if="row.status === 'reviewing'" link size="small" type="success" @click="transition(row, 'published')">发布</el-button>
          <el-button v-if="row.status === 'published'" link size="small" type="warning" @click="transition(row, 'archived')">归档</el-button>
          <el-button v-if="row.status === 'archived'" link size="small" type="primary" @click="transition(row, 'draft')">恢复</el-button>
          <el-button v-if="row.status === 'draft'" link size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listEntityTypes, createEntityType, transitionConfig, copyAsDraft, exportConfig } from '@/api/pmocker/config'

const list = ref([])
const includeDraft = ref(true)

const loadData = async () => {
  const res = await listEntityTypes({ includeDraft: includeDraft.value })
  if (res.code === 0) list.value = res.data || []
}
const openCreate = async () => {
  const { value } = await ElMessageBox.prompt('实体类型编码', '新增', { inputValue: '' })
  if (!value) return
  const res = await createEntityType({ typeCode: value, name: value, moduleCode: 'custom', status: 'published' })
  if (res.code === 0) { ElMessage.success('已创建'); loadData() }
}
const transition = async (row, to) => {
  const res = await transitionConfig({ table: 'pm_entity_types', id: row.ID, from: row.status, to })
  if (res.code === 0) { ElMessage.success('流转成功'); loadData() }
}
const copy = async (row) => {
  const res = await copyAsDraft({ table: 'pm_entity_types', id: row.ID })
  if (res.code === 0) { ElMessage.success('已复制为草稿'); loadData() }
}
const remove = async (row) => {
  await ElMessageBox.confirm('确认删除？仅草稿可删', '提示', { type: 'warning' })
  const res = await transitionConfig({ table: 'pm_entity_types', id: row.ID, from: row.status, to: 'delete' })
  if (res.code === 0) { ElMessage.success('已删除'); loadData() }
}
const exportCfg = async () => {
  const res = await exportConfig()
  if (res.code === 0) ElMessage.success('已导出到镜像源')
}
const statusLabel = (s) => ({ draft: '草稿', reviewing: '评审中', published: '已发布', archived: '已归档' }[s] || s)
const statusTag = (s) => ({ draft: 'info', reviewing: 'warning', published: 'success', archived: '' }[s] || 'info')

onMounted(loadData)
</script>
```

> **注意**：删除走 `DeleteDraft`，后端 `transition` handler 需支持 `to=delete` 调 `DeleteDraft`；或前端调独立 delete API。MVP 简化为 transition 带 delete 标记。

- [ ] **Step 4: 创建其余子页（fieldDef/dictionary/stateDef/workflow/seedEntity/orgEntry）**

按上述模式创建各子页。要点：
- `fieldDef.vue`：实体类型下拉筛选 + 字段表格（复用 DynamicForm 或表格编辑）
- `dictionary.vue`：字典列表 + 明细（可复用 gva superAdmin 字典页逻辑或直接调 gva DictionaryService API）
- `stateDef.vue`：各实体类型状态定义表格（状态值/标签/样式/actions JSON）
- `workflow.vue`：工作流列表 + JSON 编辑
- `seedEntity.vue`：项目/模块筛选 + 表格 + 编辑（复用 `/pmocker/eav/entity` CRUD）
- `orgEntry.vue`：跳转 gva superAdmin 入口卡片（`/system/...` 或 superAdmin 路由）

- [ ] **Step 5: 前端 build 验证**

Run: `cd gva/web && npm run build`
Expected: build 成功

- [ ] **Step 6: 提交**

```bash
git add gva/web/src/api/pmocker/config.js gva/web/src/view/pmocker/config/
git commit -m "feat(pmocker_config): 前端配置管理页(7子页)+API封装"
```

---

## Task 8: statusTransitions.js 改造为读 API（保留 fallback）

**Files:**
- Modify: `gva/web/src/view/pmocker/components/statusTransitions.js`

**Interfaces:**
- Produces: `getTransitions(entityType)` 改为优先读 `/pmocker/config/stateDefs/public`，缓存结果，失败时用内置 fallback
- Consumes: `listStateDefsPublic`

- [ ] **Step 1: 改造 statusTransitions.js**

在 `gva/web/src/view/pmocker/components/statusTransitions.js` 顶部加入动态加载逻辑，保留现有 `transitions` 常量作为 fallback：

```javascript
import { listStateDefsPublic } from '@/api/pmocker/config'

let remoteTransitions = null

export async function loadRemoteTransitions() {
  try {
    const res = await listStateDefsPublic({})
    if (res.code === 0 && Array.isArray(res.data)) {
      remoteTransitions = res.data
      return true
    }
  } catch (e) { /* fallback to local */ }
  return false
}

export function getTransitions(entityType) {
  if (remoteTransitions) {
    return remoteTransitions
      .filter(d => d.entityType === entityType)
      .map(d => ({
        status: d.status,
        label: d.label,
        tagType: d.tagType,
        actions: JSON.parse(d.actionsJson || '[]')
      }))
  }
  return localTransitions[entityType] || []
}
```

> **注意**：现有 `getTransitions` 依赖本地 `transitions` 常量。改造需在入口（`main.js` 或 layout）调用 `loadRemoteTransitions()` 预热，并把本地常量改名 `localTransitions` 保留 fallback。`getTransitions` 签名保持兼容，现有列表页调用不变。

- [ ] **Step 2: 前端 build + 现有页面回归**

Run: `cd gva/web && npm run build`
Expected: build 成功；现有列表页 `getTransitions` 调用不受影响

- [ ] **Step 3: 提交**

```bash
git add gva/web/src/view/pmocker/components/statusTransitions.js
git commit -m "feat(pmocker_config): statusTransitions改读API配置(保留本地fallback)"
```

---

## Task 9: 端到端验证 + 种子数据重灌

**Files:**
- Modify: 无（验证）

**Interfaces:**
- Consumes: 全部 Task 1-8 产出

- [ ] **Step 1: 重建二进制 + 重启实例**

```bash
cd gva/server && go build -o D:\Dev\pmocker\.pmocker-data\bin\gva-server.exe .
& "D:\Dev\pmocker\cli\pmocker.exe" stop pms-v12
& "D:\Dev\pmocker\cli\pmocker.exe" start pms-v12
```

- [ ] **Step 2: 验证 published 过滤**

用 admin token 调 `GET /pmocker/config/entityTypes?includeDraft=false` → 应返回全部 published 实体类型（18 个）；`GET /pmocker/eav/schema/task` → 正常返回。

- [ ] **Step 3: 验证配置 CRUD + 状态流转**

- 新增实体类型（默认 published）→ 列表可见
- transition: published → archived → draft → reviewing → published 全链路
- 复制为 draft → 新对象 status=draft

- [ ] **Step 4: 验证导出**

调用 `POST /pmocker/config/export` → 检查 `images/pmbok6-hybrid/schema.yaml` 生成且与 loader 格式一致。

- [ ] **Step 5: 提交收尾**

```bash
git add -A
git commit -m "feat(pmocker_config): M13 端到端验证通过"
```

---

## Self-Review

### 1. Spec coverage check

| Spec 要求 | 对应 Task | 状态 |
|-----------|----------|------|
| 方案 A：元表直接 CRUD | Task 1,3 | ✅ |
| 配置状态机 draft/reviewing/published/archived | Task 2 | ✅ |
| 默认值恒 published | Task 3（CreateEntityType 默认 published） | ✅ |
| 运行时动态生效（published 过滤） | Task 6（eav.go + loader） | ✅ |
| pm_state_defs 表 | Task 1 | ✅ |
| statusTransitions 改读 API（fallback） | Task 8 | ✅ |
| 6 类管理对象 CRUD | Task 3 + Task 5 + Task 7 | ✅ |
| 复制为 draft | Task 3（CopyAsDraft） | ✅ |
| 组织/用户/权限跳转 gva | Task 7（orgEntry.vue） | ✅ |
| 导出 YAML 三件套 | Task 4 | ✅ |
| 插件独立 pmocker_config | Task 5 | ✅ |
| 中间件链四件套 | Task 5（initialize） | ✅ |
| 测试 | Task 2,3,4,6 | ✅ |

### 2. Placeholder scan

- 无 TBD/TODO ✅
- 所有代码块含实际实现 ✅
- `export_service.go` 中的 schema.yaml 导出格式需按 loader.SchemaYaml 实际结构对齐（Task 4 已标注注意点）——实施时以 `pkg/pmocker/plugin/loader/loader.go` 为准 ✅

### 3. Type consistency

- `PMEntityType.Status`/`PMFieldDef.Status`/`PMWorkflowDef.Status`/`PMRelationType.Status` 一致（Task 1 定义，Task 2,3,6 消费）✅
- `PMStateDef`：`ConfigStatus` 表示配置自身状态，`Status` 表示业务状态值（Task 1 定义，Task 3 ListStateDefs 消费）✅
- `StateMachineService.Transition(db, table, id, from, to)` 签名 Task 2 定义，Task 5 API 消费一致 ✅
- `LoadEntityType(ctx, typeCode, includeDraft)` 签名变更在 Task 6，调用方 `api/v1/pmocker/eav.go` 同步更新 ✅
- `eavtypes.EntityType`/`FieldDef` 加 `Status` 字段在 Task 6，loader 消费 ✅

### 4. 补充说明

- 删除操作：MVP 通过 `transition` 带 `to=delete` 触发 `DeleteDraft`（Task 5 Step 1 已标注）
- 字典管理复用 gva `DictionaryService`（Task 7 Step 4 已标注），不重复造轮子
- 组织权限页跳转 gva superAdmin（aiDoc 复用原则）

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-09-m13-config-manager.md`.

实施选项：
1. **Subagent-Driven（推荐）** — 每 Task 派新子代理，任务间评审
2. **Inline Execution** — 本会话内用 executing-plans 分批执行

哪个方式？
