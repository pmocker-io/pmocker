package service

import (
	"context"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// TestPublishEmptySeedRejected 空 seed_yaml 发布应被拒绝，且不产生版本快照
func TestPublishEmptySeedRejected(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "empty", Name: "空包", EntityType: "task", Module: "schedule", SeedYAML: ""}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "empty").First(&pkg)

	if err := s.Publish(ctx, pkg.ID); err == nil {
		t.Fatal("空 seed_yaml 发布应报错")
	} else if !strings.Contains(err.Error(), "为空") {
		t.Fatalf("错误应提示 seed 为空，got: %v", err)
	}
	versions, _ := v.ListVersions(ctx, pkg.ID)
	if len(versions) != 0 {
		t.Fatalf("空包发布失败后不应生成版本快照，got %d", len(versions))
	}
	db.First(&pkg, pkg.ID)
	if pkg.Status != "draft" {
		t.Fatalf("发布失败后状态应保持 draft，got %s", pkg.Status)
	}
	if pkg.Version != 0 {
		t.Fatalf("发布失败后版本应保持 0，got %d", pkg.Version)
	}
}

// TestPublishNoModulesRejected 缺 modules 的 seed 发布应被拒绝
func TestPublishNoModulesRejected(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	noModules := "name: 无模块包\n"
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "nomod", Name: "无模块", EntityType: "task", Module: "schedule", SeedYAML: noModules}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "nomod").First(&pkg)

	if err := s.Publish(ctx, pkg.ID); err == nil {
		t.Fatal("缺 modules 的 seed 发布应报错")
	} else if !strings.Contains(err.Error(), "modules") {
		t.Fatalf("错误应提示缺少 modules，got: %v", err)
	}
}

// TestPublishInvalidYAMLRejected 非法 YAML 发布应被拒绝
func TestPublishInvalidYAMLRejected(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	badYAML := "modules: [unclosed"
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "bad", Name: "坏YAML", EntityType: "task", Module: "schedule", SeedYAML: badYAML}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "bad").First(&pkg)

	if err := s.Publish(ctx, pkg.ID); err == nil {
		t.Fatal("非法 YAML 发布应报错")
	}
}

// TestPublishArchivedRejected archived 配置包不可发布
func TestPublishArchivedRejected(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{})
	s := &ConfigPackageService{}

	pkg := pmocker.PMConfigPackage{Code: "arch", Name: "已归档", EntityType: "task", Module: "schedule", Status: "archived", SeedYAML: syncSeedYAML}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(ctx, pkg.ID); err == nil {
		t.Fatal("archived 配置包发布应报错")
	}
}

// TestPublishSyncFailureRollsBack Sync 同步失败（project_code 引用缺失）时，发布整体回滚且无快照
func TestPublishSyncFailureRollsBack(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	// 业务模块引用不存在的 project_code
	seed := `
name: 引用缺失
modules:
  schedule:
    entity_type: task
    name: 进度管理
    fields: []
    states: []
    transitions: []
    projects:
      - project_code: PROJ_MISSING
        entities:
          task:
            - {title: 幽灵任务, status: planned}
`
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "refmiss", Name: "引用缺失", EntityType: "task", Module: "schedule", SeedYAML: seed}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "refmiss").First(&pkg)

	if err := s.Publish(ctx, pkg.ID); err == nil {
		t.Fatal("project_code 引用缺失 EPS 时应发布失败")
	} else if !strings.Contains(err.Error(), "PROJ_MISSING") {
		t.Fatalf("错误应包含缺失的项目编码，got: %v", err)
	}
	versions, _ := v.ListVersions(ctx, pkg.ID)
	if len(versions) != 0 {
		t.Fatalf("发布失败后不应生成版本快照，got %d", len(versions))
	}
	// 已同步的实体类型也应回滚（事务）
	var count int64
	db.Model(&pmocker.PMEntityType{}).Where("type_code = ?", "task").Count(&count)
	if count != 0 {
		t.Fatalf("Sync 失败后实体类型不应残留，count=%d", count)
	}
	db.First(&pkg, pkg.ID)
	if pkg.Status != "draft" {
		t.Fatalf("发布失败后状态应保持 draft，got %s", pkg.Status)
	}
}

// TestRollbackWrongPackageVersion 回滚不属于该配置包的版本应被拒绝
func TestRollbackWrongPackageVersion(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	// 包 A
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "pkA", Name: "包A", EntityType: "risk", Module: "risk", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkA pmocker.PMConfigPackage
	db.Where("code = ?", "pkA").First(&pkA)
	if err := s.Publish(ctx, pkA.ID); err != nil {
		t.Fatal(err)
	}
	// 包 B（不发布，仅创建）
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "pkB", Name: "包B", EntityType: "task", Module: "schedule", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkB pmocker.PMConfigPackage
	db.Where("code = ?", "pkB").First(&pkB)

	// 用 B 的 ID 回滚 A 的版本
	versions, _ := v.ListVersions(ctx, pkA.ID)
	if len(versions) != 1 {
		t.Fatalf("A 应有 1 个版本，got %d", len(versions))
	}
	if err := v.Rollback(ctx, pkB.ID, versions[0].ID); err == nil {
		t.Fatal("回滚不属于该配置包的版本应报错")
	}
}

// TestRollbackNonexistentVersion 回滚不存在的版本 ID 应报错
func TestRollbackNonexistentVersion(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "vx", Name: "版本X", EntityType: "risk", Module: "risk", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "vx").First(&pkg)

	if err := v.Rollback(ctx, pkg.ID, 99999); err == nil {
		t.Fatal("回滚不存在的版本 ID 应报错")
	}
}

// TestRollbackCorruptSnapshot 快照 YAML 损坏时回滚失败且 seed_yaml 不变
func TestRollbackCorruptSnapshot(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "corrupt", Name: "坏快照", EntityType: "risk", Module: "risk", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "corrupt").First(&pkg)

	// 手工插入一个损坏快照版本
	corruptVer := pmocker.PMConfigVersion{PackageID: pkg.ID, Version: 9, SnapshotYAML: "modules: [broken", Flag: 0}
	if err := db.Create(&corruptVer).Error; err != nil {
		t.Fatal(err)
	}
	if err := v.Rollback(ctx, pkg.ID, corruptVer.ID); err == nil {
		t.Fatal("损坏快照回滚应报错")
	}
	db.First(&pkg, pkg.ID)
	if pkg.SeedYAML != syncSeedYAML {
		t.Fatal("回滚失败后 seed_yaml 不应被修改")
	}
}

// TestRollbackProjectCodeMissing 回滚时 project_code 引用缺失 EPS 应失败且事务回滚
func TestRollbackProjectCodeMissing(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	// 先建 EPS 项目 PROJ_A
	proj := pmocker.PMEntity{ProjectID: 0, EntityType: "eps_node", Title: "EPS回滚源", Status: "active"}
	if err := db.Create(&proj).Error; err != nil {
		t.Fatal(err)
	}
	if err := setAttr(db, proj.ID, "code", "PROJ_A"); err != nil {
		t.Fatal(err)
	}

	// 发布 v1（引用 PROJ_A，正常）
	seedV1 := `
name: 回滚引用测试
modules:
  schedule:
    entity_type: task
    name: 进度管理
    fields: []
    states: []
    transitions: []
    projects:
      - project_code: PROJ_A
        entities:
          task:
            - {title: 关联任务, status: planned}
`
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "rbmiss", Name: "回滚引用缺失", EntityType: "task", Module: "schedule", SeedYAML: seedV1}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "rbmiss").First(&pkg)
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatal(err)
	}
	versions, _ := v.ListVersions(ctx, pkg.ID)
	if len(versions) != 1 {
		t.Fatalf("v1 应有 1 个快照，got %d", len(versions))
	}

	// 删除 EPS 项目，使回滚时 project_code 解析失败
	if err := db.Delete(&proj).Error; err != nil {
		t.Fatal(err)
	}

	if err := v.Rollback(ctx, pkg.ID, versions[0].ID); err == nil {
		t.Fatal("回滚时 project_code 引用缺失 EPS 应报错")
	}
	// seed_yaml 不应被回滚修改
	db.First(&pkg, pkg.ID)
	if pkg.SeedYAML != seedV1 {
		t.Fatal("回滚失败后 seed_yaml 不应被修改")
	}
}
