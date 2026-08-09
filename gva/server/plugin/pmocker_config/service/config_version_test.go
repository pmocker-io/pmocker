package service

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestVersionSnapshotOnPublish(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "v1", Name: "版本测试", EntityType: "risk", Module: "risk", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "v1").First(&pkg)

	// 发布 → 生成版本快照
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatal(err)
	}
	versions, err := v.ListVersions(ctx, pkg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatalf("发布后应生成 1 个版本快照，got %d", len(versions))
	}
	if versions[0].SnapshotYAML != syncSeedYAML {
		t.Fatal("快照内容应与发布时的 seed_yaml 一致")
	}
	if versions[0].Version != 1 {
		t.Fatalf("快照版本 = %d, want 1", versions[0].Version)
	}
}

func TestRollbackRestoresSeedYAML(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	v := &ConfigVersionService{}

	// 创建并发布 v1
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "rb", Name: "回滚测试", EntityType: "risk", Module: "risk", SeedYAML: syncSeedYAML}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "rb").First(&pkg)
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatal(err)
	}

	// 修改 seed_yaml 并再次发布（v2）
	newSeed := "name: 风险管理V2\nmodules:\n  risk:\n    entity_type: risk\n    name: 风险管理V2\n    fields: []\n    states: []\n    transitions: []\n"
	if err := s.UpdateSeed(ctx, pkg.ID, newSeed); err != nil {
		t.Fatal(err)
	}
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatal(err)
	}
	db.First(&pkg, pkg.ID)
	if pkg.SeedYAML != newSeed {
		t.Fatal("v2 发布后 seed_yaml 应为新值")
	}

	// 回滚到 v1
	versions, _ := v.ListVersions(ctx, pkg.ID)
	if len(versions) != 2 {
		t.Fatalf("应有 2 个版本，got %d", len(versions))
	}
	v1 := versions[1] // 按时间倒序，v1 在后面
	if err := v.Rollback(ctx, pkg.ID, v1.ID); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	db.First(&pkg, pkg.ID)
	if pkg.SeedYAML != syncSeedYAML {
		t.Fatal("回滚后 seed_yaml 应恢复为 v1 快照")
	}
	if pkg.Status != "published" {
		t.Fatalf("回滚后状态 = %s, want published", pkg.Status)
	}
}
