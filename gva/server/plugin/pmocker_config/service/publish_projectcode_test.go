package service

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// TestPublishWithProjectCode 完整复现：EPS项目 + 业务配置包(project_code) 发布后任务正确归属
func TestPublishWithProjectCode(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t,
		&pmocker.PMConfigPackage{}, &pmocker.PMConfigVersion{},
		&pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{},
		&pmocker.PMEntity{}, &pmocker.PMAttr{},
	)
	// 1. 建 EPS 项目（模拟真实：title=智能排产系统研发, code=PROJ_A）
	proj := pmocker.PMEntity{ProjectID: 0, EntityType: "eps_node", Title: "智能排产系统研发", Status: "active"}
	if err := db.Create(&proj).Error; err != nil {
		t.Fatal(err)
	}
	if err := setAttr(db, proj.ID, "code", "PROJ_A"); err != nil {
		t.Fatal(err)
	}

	// 2. 建 schedule 配置包（project_code: PROJ_A）
	seed := `
name: 进度配置
modules:
  schedule:
    entity_type: task
    name: 进度管理
    fields:
      - {key: code, label: 任务编号, data_type: string}
    states:
      - {status: planned, label: 计划中, tag_type: info}
    transitions: []
    projects:
      - project_code: PROJ_A
        entities:
          task:
            - {title: 需求调研, status: planned, priority: 2}
`
	s := &ConfigPackageService{}
	if err := s.Create(ctx, pmocker.PMConfigPackage{
		Code: "schedule", Name: "进度管理", EntityType: "task", Module: "schedule", SeedYAML: seed,
	}); err != nil {
		t.Fatal(err)
	}
	var pkg pmocker.PMConfigPackage
	db.Where("code = ?", "schedule").First(&pkg)

	// 3. 发布
	if err := s.Publish(ctx, pkg.ID); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// 4. 验证任务归属到 EPS 项目
	var task pmocker.PMEntity
	if err := db.Where("entity_type = ? AND title = ?", "task", "需求调研").First(&task).Error; err != nil {
		t.Fatalf("任务未创建: %v", err)
	}
	if task.ProjectID != proj.ID {
		t.Fatalf("任务 project_id = %d, want %d (project_code 应解析)", task.ProjectID, proj.ID)
	}
}
