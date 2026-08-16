package team

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/glebarez/sqlite"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"gorm.io/gorm"
)

func setupTeamTestDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	global.GVA_DB = db
	db.AutoMigrate(
		&pmocker.PMEntityType{},
		&pmocker.PMFieldDef{},
		&pmocker.PMEntity{},
		&pmocker.PMAttr{},
	)
	svc := &pmservice.EAVService{}
	ctx := context.Background()
	if err := svc.RegisterEntityType(ctx, eavtypes.EntityType{TypeCode: EntityTypePerformanceReview, ModuleCode: "team", Name: "绩效评估"}); err != nil {
		t.Fatal(err)
	}
	for _, f := range []struct{ key, dataType string }{
		{"rating", "enum"}, {"score", "decimal"}, {"member_id", "user"},
	} {
		if err := svc.RegisterFieldDef(ctx, eavtypes.FieldDef{
			EntityType: EntityTypePerformanceReview, FieldKey: f.key, FieldLabel: f.key, DataType: eavtypes.DataType(f.dataType),
		}); err != nil {
			t.Fatal(err)
		}
	}
}

// TestScoreCalcFromRating rating 已填 → handler 按评级推断默认分回写
func TestScoreCalcFromRating(t *testing.T) {
	setupTeamTestDB(t)
	ctx := context.Background()
	svc := &pmservice.EAVService{}

	id, err := svc.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: 1, EntityType: EntityTypePerformanceReview, Title: "季度评估", Status: "in_review",
		Attrs: map[string]interface{}{"rating": "good", "member_id": 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ScoreCalcHandler(ctx, id); err != nil {
		t.Fatalf("ScoreCalcHandler: %v", err)
	}
	e, err := svc.GetEntity(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := e.Attrs["score"].(int64); !ok || v != 80 {
		t.Fatalf("score = %v (type %T), want 80", e.Attrs["score"], e.Attrs["score"])
	}
}

// TestScoreCalcPreservesManualScore 已人工评分 → handler 不改写
func TestScoreCalcPreservesManualScore(t *testing.T) {
	setupTeamTestDB(t)
	ctx := context.Background()
	svc := &pmservice.EAVService{}

	id, err := svc.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: 1, EntityType: EntityTypePerformanceReview, Title: "年度评估", Status: "in_review",
		Attrs: map[string]interface{}{"rating": "excellent", "score": 95.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ScoreCalcHandler(ctx, id); err != nil {
		t.Fatalf("ScoreCalcHandler: %v", err)
	}
	e, _ := svc.GetEntity(ctx, id)
	if v, ok := e.Attrs["score"].(int64); !ok || v != 95 {
		t.Fatalf("人工评分被覆盖: score = %v (type %T), want 95", e.Attrs["score"], e.Attrs["score"])
	}
}

// TestScoreCalcNoRatingNoError rating/score 均未填 → 不报错、流程照常
func TestScoreCalcNoRatingNoError(t *testing.T) {
	setupTeamTestDB(t)
	ctx := context.Background()
	svc := &pmservice.EAVService{}

	id, err := svc.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: 1, EntityType: EntityTypePerformanceReview, Title: "试用期评估", Status: "in_review",
		Attrs: map[string]interface{}{"member_id": 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ScoreCalcHandler(ctx, id); err != nil {
		t.Fatalf("ScoreCalcHandler 对未填评分不应报错: %v", err)
	}
	e, _ := svc.GetEntity(ctx, id)
	if e.Attrs["score"] != nil {
		t.Fatalf("未填评级时不应写入 score: %v", e.Attrs["score"])
	}
}

// TestScoreCalcWrongEntityType 非 performance_review 实体 → 报错
func TestScoreCalcWrongEntityType(t *testing.T) {
	setupTeamTestDB(t)
	ctx := context.Background()
	svc := &pmservice.EAVService{}

	if err := svc.RegisterEntityType(ctx, eavtypes.EntityType{TypeCode: "team_member", ModuleCode: "team", Name: "成员"}); err != nil {
		t.Fatal(err)
	}
	id, err := svc.CreateEntity(ctx, eavtypes.Entity{ProjectID: 1, EntityType: "team_member", Title: "张三", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if err := ScoreCalcHandler(ctx, id); err == nil {
		t.Fatal("非 performance_review 实体应报错")
	}
}
