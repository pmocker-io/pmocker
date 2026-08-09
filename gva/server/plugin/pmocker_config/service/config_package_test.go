package service

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestConfigPackageCRUD(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	ctx := context.Background()
	s := &ConfigPackageService{}

	// 创建（默认 draft）
	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "requirement", Name: "需求管理", EntityType: "requirement", Module: "requirement"}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	if err := db.Where("code = ?", "requirement").First(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	if pkg.Status != "draft" {
		t.Fatalf("新建配置包状态 = %s, want draft", pkg.Status)
	}
	if pkg.Version != 0 {
		t.Fatalf("新建配置包版本 = %d, want 0（未发布）", pkg.Version)
	}

	// 更新 seed_yaml（draft 可编辑）
	seed := "entity_type: requirement\nmodule: requirement\nname: 需求管理\n"
	if err := s.UpdateSeed(ctx, pkg.ID, seed); err != nil {
		t.Fatal(err)
	}
	var updated pmocker.PMConfigPackage
	db.First(&updated, pkg.ID)
	if updated.SeedYAML != seed {
		t.Fatal("UpdateSeed 未生效")
	}

	// 列表
	list, err := s.List(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("List count = %d, want 1", len(list))
	}

	// 删除（draft 可删）
	if err := s.Delete(ctx, pkg.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, pkg.ID); err == nil {
		t.Fatal("删除后应查不到")
	}
}

func TestConfigPackageCopyAsDraft(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	ctx := context.Background()
	s := &ConfigPackageService{}

	if err := s.Create(ctx, pmocker.PMConfigPackage{Code: "src", Name: "源包", EntityType: "task", Module: "schedule"}); err != nil {
		t.Fatal(err)
	}
	var src pmocker.PMConfigPackage
	db.Where("code = ?", "src").First(&src)

	if err := s.CopyAsDraft(ctx, src.ID); err != nil {
		t.Fatal(err)
	}
	var copy pmocker.PMConfigPackage
	if err := db.Where("code = ?", "src-copy").First(&copy).Error; err != nil {
		t.Fatalf("复制未创建: %v", err)
	}
	if copy.Status != "draft" {
		t.Fatalf("复制包状态 = %s, want draft", copy.Status)
	}
	if copy.EntityType != "task" {
		t.Fatalf("复制包 entityType = %s, want task", copy.EntityType)
	}
}

func TestLoadSeedPackageIdempotent(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	ctx := context.Background()
	s := &ConfigPackageService{}

	if err := s.LoadSeedPackage(ctx, "eps", testSeedYAML); err != nil {
		t.Fatal(err)
	}
	// 重复加载幂等
	if err := s.LoadSeedPackage(ctx, "eps", testSeedYAML); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&pmocker.PMConfigPackage{}).Count(&count)
	if count != 1 {
		t.Fatalf("幂等失败 count=%d", count)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "eps").First(&pkg)
	if pkg.Status != "draft" {
		t.Fatalf("导入配置包状态 = %s, want draft", pkg.Status)
	}
	if pkg.EntityType != "requirement" {
		t.Fatalf("EntityType = %s, want requirement", pkg.EntityType)
	}
}

func TestConfigPackageDeletePublishedRejected(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	ctx := context.Background()
	s := &ConfigPackageService{}

	pkg := pmocker.PMConfigPackage{Code: "p", Name: "已发布", EntityType: "risk", Module: "risk", Status: "published"}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, pkg.ID); err == nil {
		t.Fatal("published 配置包不应被删除")
	}
}
