package service

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

const syncSeedYAML = `
entity_type: requirement
module: requirement
name: 需求管理
fields:
  - {key: code, label: 需求编码, data_type: string}
  - {key: priority, label: 优先级, data_type: enum, options: [P0,P1,P2,P3], default: P2}
states:
  - {status: draft, label: 草稿, tag_type: info}
  - {status: reviewing, label: 评审中, tag_type: warning}
  - {status: published, label: 已发布, tag_type: success}
transitions:
  - {from: draft, to: reviewing, action: submit}
  - {from: reviewing, to: published, action: approve}
projects:
  - project_id: 3
    entities:
      requirement:
        - {title: 排产算法, status: published, priority: P0}
`

func TestSyncEntityTypeAndFields(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	seed, err := ParseSeedYAML([]byte(syncSeedYAML))
	if err != nil {
		t.Fatal(err)
	}
	s := &SeedSyncService{}
	if err := s.Sync(ctx, db, seed); err != nil {
		t.Fatal(err)
	}

	// 实体类型同步
	var et pmocker.PMEntityType
	if err := db.Where("type_code = ?", "requirement").First(&et).Error; err != nil {
		t.Fatalf("实体类型未同步: %v", err)
	}
	if et.Status != "published" {
		t.Fatalf("实体类型 status = %s, want published", et.Status)
	}

	// 字段同步
	var fd pmocker.PMFieldDef
	if err := db.Where("entity_type = ? AND field_key = ?", "requirement", "priority").First(&fd).Error; err != nil {
		t.Fatalf("字段未同步: %v", err)
	}
	if fd.DefaultValue != "P2" {
		t.Fatalf("字段默认值 = %s, want P2", fd.DefaultValue)
	}

	// 状态流转同步到 pm_state_defs
	var sd pmocker.PMStateDef
	if err := db.Where("entity_type = ? AND status = ?", "requirement", "reviewing").First(&sd).Error; err != nil {
		t.Fatalf("状态定义未同步: %v", err)
	}
	if sd.Label != "评审中" {
		t.Fatalf("状态 label = %s, want 评审中", sd.Label)
	}

	// 幂等：重复同步不重复
	_ = s.Sync(ctx, db, seed)
	var etCount int64
	db.Model(&pmocker.PMEntityType{}).Where("type_code = ?", "requirement").Count(&etCount)
	if etCount != 1 {
		t.Fatalf("实体类型幂等失败 count=%d", etCount)
	}
	var fdCount int64
	db.Model(&pmocker.PMFieldDef{}).Where("entity_type = ?", "requirement").Count(&fdCount)
	if fdCount != 2 {
		t.Fatalf("字段幂等失败 count=%d", fdCount)
	}
}

func TestSyncProjectEntities(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	seed, err := ParseSeedYAML([]byte(syncSeedYAML))
	if err != nil {
		t.Fatal(err)
	}
	s := &SeedSyncService{}
	if err := s.Sync(ctx, db, seed); err != nil {
		t.Fatal(err)
	}

	// 项目实体种子同步
	var ent pmocker.PMEntity
	if err := db.Where("project_id = ? AND entity_type = ? AND title = ?", 3, "requirement", "排产算法").First(&ent).Error; err != nil {
		t.Fatalf("项目实体种子未同步: %v", err)
	}
	// priority attr
	var attr pmocker.PMAttr
	if err := db.Where("entity_id = ? AND field_key = ?", ent.ID, "priority").First(&attr).Error; err != nil {
		t.Fatalf("实体 attr 未同步: %v", err)
	}
	if attr.ValString == nil || *attr.ValString != "P0" {
		t.Fatalf("priority attr = %+v, want P0", attr)
	}
}

func TestPublishPackageSyncsDB(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{}, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{}, &pmocker.PMEntity{}, &pmocker.PMAttr{})
	s := &ConfigPackageService{}
	if err := s.Create(ctx, pmocker.PMConfigPackage{
		Code: "requirement", Name: "需求管理", EntityType: "requirement", Module: "requirement",
		SeedYAML: syncSeedYAML,
	}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "requirement").First(&pkg)

	// 发布（draft → reviewing → published）
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatalf("Publish: %v", err)
	}
	var updated pmocker.PMConfigPackage
	db.First(&updated, pkg.ID)
	if updated.Status != "published" {
		t.Fatalf("发布后状态 = %s, want published", updated.Status)
	}
	// 运行表同步
	var et pmocker.PMEntityType
	if err := db.Where("type_code = ?", "requirement").First(&et).Error; err != nil {
		t.Fatalf("发布后实体类型未同步: %v", err)
	}
}
