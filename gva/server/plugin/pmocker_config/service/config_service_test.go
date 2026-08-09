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
