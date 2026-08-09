# M13 初始配置管理（聚合配置包模型）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 推翻 v1.0「6 类对象各自 CRUD」，重建为**聚合配置包模型**——一条配置包记录 = 实体类型 + 字段 + 种子数据(含项目) + 状态定义 + 流转规则，对标 gva autoCode 包管理，带状态机与版本管理，发布时自动同步 DB。

**Architecture:** 新增 `pm_config_packages`/`pm_config_versions` 表存配置包与版本快照；seed_yaml（YAML 真源）聚合全部配置；发布时 `SeedSyncService` 校验 seed_yaml 并事务内同步到 `pm_entity_types`/`pm_field_defs`/`pm_state_defs`/`pm_entities`/`pm_attrs`；每模块一包 + 独立 EPS 配置包（树层级）。

**Tech Stack:** Go 1.25 / gin-vue-admin v3.0.0 / GORM / Vue 3 + Element Plus / yaml.v3

## Global Constraints

- 分层：`Router → API → Service → Model`；`enter.go` 组装
- Service 首参 `ctx context.Context`，用 `global.GVA_DB.WithContext(ctx)`
- 插件私有组中间件链：`JWTAuth → MustChangePwdGuard → CasbinHandler → DataScope`
- `plugin.go` 实现 v2 `interfaces.Plugin`，`init()` 自注册；实现 `PMockerPlugin.InitPMocker`
- 统一响应 `{code,data,msg}`；前端 HTTP 统一 `@/utils/request`；`defineModel()`；UnoCSS；`<svg-icon>`
- 配置包状态机：`draft → reviewing → published → archived`；**发布时同步 DB**（编辑保存不写运行表）
- 每模块一配置包 + 独立 EPS 配置包（树层级：非底层=容器，叶子=项目）
- 版本管理：发布生成不可变快照，支持回滚
- commit：`type(scope): description`，中文描述

---

## File Structure

### 后端（gva/server/plugin/pmocker_config/）

| 文件 | 职责 |
|------|------|
| `model/config.go`（新建） | PMConfigPackage / PMConfigVersion model |
| `service/seed_parser.go`（新建） | seed_yaml → 结构化解析 |
| `service/seed_sync.go`（新建） | 发布时同步 DB（核心） |
| `service/config_package.go`（新建） | 配置包 CRUD + 复制 |
| `service/state_machine.go`（改造） | 状态机（作用于配置包） |
| `service/export_service.go`（改造） | 导出 published 配置包为 YAML |
| `api/config.go`（重写） | ConfigApi（packages 端点） |
| `router/config.go`（重写） | 路由 |
| `plugin.go` / `initialize/init.go` | 插件入口（保留） |

### 现有文件修改

| 文件 | 修改 |
|------|------|
| `gva/server/model/pmocker/config.go` | 已有 PMStateDef（保留） |
| `gva/server/plugin/pmocker_core/plugin.go` | registerTables 追加 PMConfigPackage/PMConfigVersion |
| `gva/web/src/view/pmocker/eps/tree.vue` | EPS 项目新增/修改传参修复 |

### 前端（gva/web/src/）

| 文件 | 职责 |
|------|------|
| `api/pmocker/config.js`（重写） | 配置包 API |
| `view/pmocker/config/packageList.vue`（新建） | 配置包列表（主入口） |
| `view/pmocker/config/packageEditor.vue`（新建） | 配置包编辑（字段/状态/流转/项目种子） |
| `view/pmocker/config/index.vue`（重写） | 初始配置菜单容器 |

---

## Task 1: 数据模型（pm_config_packages / pm_config_versions）

**Files:**
- Create: `gva/server/model/pmocker/config_package.go`
- Modify: `gva/server/plugin/pmocker_core/plugin.go`

**Interfaces:**
- Produces: `PMConfigPackage`（`TableName() = "pm_config_packages"`）、`PMConfigVersion`（`TableName() = "pm_config_versions"`）

- [ ] **Step 1: 创建 model**

创建 `gva/server/model/pmocker/config_package.go`：

```go
package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMConfigPackage 配置包：聚合实体类型+字段+种子+状态+流转
type PMConfigPackage struct {
	global.GVA_MODEL
	Code        string `json:"code" gorm:"size:64;uniqueIndex;not null;comment:配置包编码"`
	Name        string `json:"name" gorm:"size:128;comment:显示名"`
	Description string `json:"description" gorm:"type:text;comment:描述"`
	EntityType  string `json:"entityType" gorm:"size:64;index;comment:实体类型"`
	Module      string `json:"module" gorm:"size:32;index;comment:所属模块"`
	Version     int    `json:"version" gorm:"default:1;comment:当前版本号"`
	Status      string `json:"status" gorm:"size:16;default:draft;comment:draft/reviewing/published/archived"`
	SeedYAML    string `json:"seedYaml" gorm:"type:text;comment:种子数据YAML真源"`
}

func (PMConfigPackage) TableName() string { return "pm_config_packages" }

// PMConfigVersion 配置包版本快照（发布时生成，不可变）
type PMConfigVersion struct {
	global.GVA_MODEL
	PackageID    uint   `json:"packageId" gorm:"index;not null;comment:配置包ID"`
	Version      int    `json:"version" gorm:"comment:版本号"`
	SnapshotYAML string `json:"snapshotYaml" gorm:"type:text;comment:版本快照YAML"`
	Flag         int    `json:"flag" gorm:"default:0;comment:0=发布 1=回滚"`
}

func (PMConfigVersion) TableName() string { return "pm_config_versions" }
```

- [ ] **Step 2: 注册迁移**

修改 `gva/server/plugin/pmocker_core/plugin.go` 的 `registerTables()`，`tables` 追加 `pmocker.PMConfigPackage{}`、`pmocker.PMConfigVersion{}`。

- [ ] **Step 3: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 4: 提交**

```bash
git add gva/server/model/pmocker/config_package.go gva/server/plugin/pmocker_core/plugin.go
git commit -m "feat(pmocker): 新增pm_config_packages/pm_config_versions配置包模型"
```

---

## Task 2: SeedParser（seed_yaml → 结构化）

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/seed_parser.go`
- Test: `gva/server/plugin/pmocker_config/service/seed_parser_test.go`

**Interfaces:**
- Produces:
  - `type ConfigPackageSeed struct { EntityType, Module, Name string; Fields []FieldSeed; States []StateSeed; Transitions []TransitionSeed; Projects []ProjectSeed }`
  - `type FieldSeed struct { Key, Label, DataType string; Options []string; Default string }`
  - `type StateSeed struct { Status, Label, TagType string }`
  - `type TransitionSeed struct { From, To, Action string; Rollback bool }`
  - `type ProjectSeed struct { Code, Name, Type string; ProjectID uint; Status string; Priority int; Children []ProjectSeed; Entities map[string][]map[string]interface{} }`
  - `func ParseSeedYAML(bytes []byte) (*ConfigPackageSeed, error)`

- [ ] **Step 1: 写失败测试**

创建 `gva/server/plugin/pmocker_config/service/seed_parser_test.go`，验证解析完整 seed_yaml（字段/状态/流转/项目树）。

- [ ] **Step 2: 运行测试验证失败**

Run: `cd gva/server && go test ./plugin/pmocker_config/service/... -run TestParseSeedYAML -v`
Expected: 编译失败（未定义）

- [ ] **Step 3: 实现 ParseSeedYAML**

用 `gopkg.in/yaml.v3` 解析，结构见 Interfaces。注意 `Entities map[string][]map[string]interface{}`（每模块实体数组）。

- [ ] **Step 4: 测试通过 + 提交**

---

## Task 3: ConfigPackageService（配置包 CRUD + 复制）

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/config_package.go`
- Test: `gva/server/plugin/pmocker_config/service/config_package_test.go`

**Interfaces:**
- Produces:
  - `func (s *ConfigPackageService) List(ctx, includeDraft bool) ([]PMConfigPackage, error)`
  - `func (s *ConfigPackageService) Create(ctx, pkg PMConfigPackage) error`（默认 draft）
  - `func (s *ConfigPackageService) Get(ctx, id uint) (*PMConfigPackage, error)`
  - `func (s *ConfigPackageService) UpdateSeed(ctx, id uint, seedYAML string) error`（draft/reviewing 可改）
  - `func (s *ConfigPackageService) CopyAsDraft(ctx, id uint) error`（code 加 -copy）
  - `func (s *ConfigPackageService) Delete(ctx, id uint) error`（仅 draft）

- [ ] **Step 1: 写失败测试**（CRUD + 复制 + 仅 draft 可删）
- [ ] **Step 2: 验证失败**
- [ ] **Step 3: 实现**
- [ ] **Step 4: 测试通过 + 提交**

---

## Task 4: 状态机 + 发布同步（核心）

**Files:**
- Modify: `gva/server/plugin/pmocker_config/service/state_machine.go`
- Create: `gva/server/plugin/pmocker_config/service/seed_sync.go`
- Test: `gva/server/plugin/pmocker_config/service/seed_sync_test.go`

**Interfaces:**
- Produces:
  - `func (s *SeedSyncService) Sync(ctx, pkg *PMConfigPackage) error` — 发布时同步 DB（事务）
  - `func (s *SeedSyncService) SyncEntityType(ctx, db, seed) error`
  - `func (s *SeedSyncService) SyncFields(ctx, db, seed) error`
  - `func (s *SeedSyncService) SyncStates(ctx, db, seed) error`
  - `func (s *SeedSyncService) SyncProjects(ctx, db, seed) error`（EPS 包建树 + 业务包建实体）
- `Transition` 改造：`approve`（to=published）时触发 `SeedSyncService.Sync`

- [ ] **Step 1: 写失败测试**（发布同步：seed_yaml → pm_entity_types/pm_field_defs/pm_state_defs/pm_entities 正确写入 + 幂等）

```go
func TestSyncEntityTypeAndFields(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	seed, _ := ParseSeedYAML([]byte(testSeedYAML))
	s := &SeedSyncService{}
	if err := s.SyncEntityType(db, seed); err != nil { t.Fatal(err) }
	if err := s.SyncFields(db, seed); err != nil { t.Fatal(err) }
	var et pmocker.PMEntityType
	if err := db.Where("type_code = ?", "requirement").First(&et).Error; err != nil { t.Fatal(err) }
	if et.Status != "published" { t.Fatalf("同步后实体类型应 published") }
	var fd pmocker.PMFieldDef
	if err := db.Where("entity_type = ? AND field_key = ?", "requirement", "priority").First(&fd).Error; err != nil { t.Fatal(err) }
	// 幂等：重复同步不重复创建
	_ = s.SyncFields(db, seed)
	var count int64
	db.Model(&pmocker.PMFieldDef{}).Where("entity_type = ?", "requirement").Count(&count)
	if count != 2 { t.Fatalf("幂等失败 count=%d", count) }
}
```

- [ ] **Step 2: 验证失败**
- [ ] **Step 3: 实现 seed_sync.go**（Sync* 方法，upsert 幂等）
- [ ] **Step 4: 改造 state_machine.go**：`Transition` 增加 `publish` 分支调用 Sync
- [ ] **Step 5: 测试通过 + 提交**

---

## Task 5: 版本快照 + 回滚

**Files:**
- Create: `gva/server/plugin/pmocker_config/service/config_version.go`
- Test: `gva/server/plugin/pmocker_config/service/config_version_test.go`

**Interfaces:**
- Produces:
  - `func (s *ConfigVersionService) Snapshot(ctx, pkg) error`（发布时记录版本）
  - `func (s *ConfigVersionService) ListVersions(ctx, packageID) ([]PMConfigVersion, error)`
  - `func (s *ConfigVersionService) Rollback(ctx, packageID, version) error`（恢复 snapshot → seed_yaml → 重新发布同步，flag=1）

- [ ] **Step 1: 写失败测试**（发布记录版本 → 修改 → 回滚 → seed_yaml 恢复）
- [ ] **Step 2: 验证失败**
- [ ] **Step 3: 实现**
- [ ] **Step 4: 测试通过 + 提交**

---

## Task 6: API 层 + Router 重写

**Files:**
- Rewrite: `gva/server/plugin/pmocker_config/api/config.go`
- Rewrite: `gva/server/plugin/pmocker_config/router/config.go`
- Modify: `gva/server/plugin/pmocker_config/pmocker/api.yaml` `menu.yaml`

**Interfaces:**
- Produces: `/pmocker/config/packages`、`/package`、`/package/:id`、`/package/:id/copy`、`/package/:id/transition`、`/package/:id/versions`、`/package/:id/rollback`、`/export`

- [ ] **Step 1: 重写 api/config.go**（packages 端点，完整 Swagger）
- [ ] **Step 2: 重写 router/config.go**
- [ ] **Step 3: 更新 api.yaml/menu.yaml**（menu 改为 配置包列表 + 编辑页）
- [ ] **Step 4: 编译 + 提交**

---

## Task 7: 前端配置包列表 + 编辑页

**Files:**
- Rewrite: `gva/web/src/api/pmocker/config.js`
- Create: `gva/web/src/view/pmocker/config/packageList.vue`
- Create: `gva/web/src/view/pmocker/config/packageEditor.vue`
- Rewrite: `gva/web/src/view/pmocker/config/index.vue`
- Delete: 旧 7 子页（entityType/fieldDef/dictionary/stateDef/workflow/seedEntity/orgEntry）

**Interfaces:**
- Consumes: `/pmocker/config/packages` 等 API

- [ ] **Step 1: 重写 config.js API**
- [ ] **Step 2: packageList.vue**（列表 + 操作按钮 + 新建）
- [ ] **Step 3: packageEditor.vue**（基本信息 + seed_yaml 区块编辑：字段表格/状态表格/流转表格/项目种子；版本侧栏）
- [ ] **Step 4: 重写 index.vue**（VerticalTabLayout 或直接包列表为主页）
- [ ] **Step 5: 删除旧 7 子页 + build 验证 + 提交**

---

## Task 8: EPS 项目编辑修复

**Files:**
- Modify: `gva/web/src/view/pmocker/eps/tree.vue`

- [ ] **Step 1: 对齐传参**

修复 `handleSave`：`createEPSNode` 传 `{projectId, name: form.title, parentId, attrs, status}`；`updateEPSNode` 传 `{ID, title, attrs, status}`（与后端期望一致）。同时确认后端 `CreateNode` 读 `name`、`UpdateNode` 读 `Entity{title}`。

- [ ] **Step 2: build + 浏览器验证新增/修改项目成功**

---

## Task 9: 端到端验证 + 清理

- [ ] **Step 1: 重建二进制 + 重启实例**
- [ ] **Step 2: 创建配置包 → 编辑 seed → 发布 → 验证运行表生效**
- [ ] **Step 3: 版本回滚验证**
- [ ] **Step 4: EPS 新增/修改项目验证**
- [ ] **Step 5: 清理旧 6 类 CRUD 残留（前端 7 子页删除、后端冗余方法）+ 提交**

---

## Self-Review

### 1. Spec coverage

| Spec 要求 | Task |
|-----------|------|
| 配置包 = 实体+字段+种子+状态+流转 | Task 2 (seed_parser) |
| 每模块一包 + EPS 包 | Task 1,2 |
| seed_yaml YAML 存储 | Task 2 |
| 发布时同步 DB | Task 4 |
| 状态机 | Task 4 |
| 版本快照 + 回滚 | Task 5 |
| 配置包列表→编辑（autoCode 模式） | Task 7 |
| EPS 项目新增/修改 | Task 8 |
| 推翻 6 类 CRUD | Task 9 |

### 2. Placeholder scan

- 无 TBD/TODO ✅
- 代码块含实际实现 ✅

### 3. Type consistency

- `PMConfigPackage.SeedYAML`/`PMConfigVersion.SnapshotYAML` 一致（Task 1 定义，Task 4/5 消费）✅
- `ConfigPackageSeed` 结构 Task 2 定义，Task 4 Sync 消费 ✅
- `Transition` 改造：approve → Sync（Task 4）✅

### 4. 执行顺序依赖

- Task 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9
- Task 4 依赖 Task 2（parser）、Task 3（pkg CRUD）
- Task 5 依赖 Task 4（Sync）

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-09-m13-config-manager.md`.

实施选项：
1. **Subagent-Driven（推荐）** — 每 Task 派新子代理，任务间评审
2. **Inline Execution** — 本会话内分批执行
