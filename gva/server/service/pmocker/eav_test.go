package pmocker

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	global.GVA_DB = db
	// 自动迁移 PMocker 表
	db.AutoMigrate(
		&pmocker.PMEntityType{},
		&pmocker.PMFieldDef{},
		&pmocker.PMRelationType{},
		&pmocker.PMEntity{},
		&pmocker.PMAttr{},
		&pmocker.PMRelation{},
	)
}

func TestEAVCreateAndGet(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	svc := &EAVService{}

	// 注册 schema
	err := svc.RegisterEntityType(ctx, eavtypes.EntityType{
		TypeCode:   "test_requirement",
		ModuleCode: "requirement",
		Name:       "测试需求",
	})
	if err != nil {
		t.Fatalf("RegisterEntityType failed: %v", err)
	}
	err = svc.RegisterFieldDef(ctx, eavtypes.FieldDef{
		EntityType:  "test_requirement",
		FieldKey:    "priority",
		FieldLabel:  "优先级",
		DataType:    eavtypes.DataTypeEnum,
		OptionsJSON: `["P0","P1","P2","P3"]`,
	})
	if err != nil {
		t.Fatalf("RegisterFieldDef failed: %v", err)
	}

	// 创建实体
	priority := "P1"
	id, err := svc.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  1,
		EntityType: "test_requirement",
		Title:      "测试需求1",
		Status:     "draft",
		Attrs: map[string]interface{}{
			"priority": priority,
		},
	})
	if err != nil {
		t.Fatalf("CreateEntity failed: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateEntity returned 0 ID")
	}

	// 查询实体
	entity, err := svc.GetEntity(ctx, id)
	if err != nil {
		t.Fatalf("GetEntity failed: %v", err)
	}
	if entity.Title != "测试需求1" {
		t.Errorf("expected title 测试需求1, got %s", entity.Title)
	}
	if entity.Attrs["priority"] != "P1" {
		t.Errorf("expected priority P1, got %v", entity.Attrs["priority"])
	}

	// 清理
	svc.DeleteEntity(ctx, id)
}
