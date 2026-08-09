package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"gopkg.in/yaml.v3"
)

func TestExportGeneratesLoadableSchema(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntityType{}, &pmocker.PMFieldDef{}, &pmocker.PMStateDef{})
	db.Create(&pmocker.PMEntityType{TypeCode: "task", ModuleCode: "schedule", Name: "任务", Status: "published"})
	db.Create(&pmocker.PMEntityType{TypeCode: "draftType", ModuleCode: "m", Name: "草稿", Status: "draft"})
	db.Create(&pmocker.PMFieldDef{EntityType: "task", FieldKey: "code", FieldLabel: "编码", DataType: "string", Status: "published"})

	s := &ExportService{}
	dir := t.TempDir()
	if err := s.Export(context.Background(), dir); err != nil {
		t.Fatalf("Export: %v", err)
	}

	b, err := os.ReadFile(filepath.Join(dir, "schema.yaml"))
	if err != nil {
		t.Fatalf("schema.yaml 未生成: %v", err)
	}

	var schema loader.SchemaYaml
	if err := yaml.Unmarshal(b, &schema); err != nil {
		t.Fatalf("schema.yaml 无法被 loader 结构解析: %v", err)
	}
	// 只导出 published（草稿类型被过滤）
	if len(schema.EntityTypes) != 1 {
		t.Fatalf("应只导出 1 个 published 实体类型，got %d", len(schema.EntityTypes))
	}
	if schema.EntityTypes[0].EntityType != "task" {
		t.Fatalf("导出实体类型 = %s, want task", schema.EntityTypes[0].EntityType)
	}
	if len(schema.EntityTypes[0].Fields) != 1 {
		t.Fatalf("task 应导出 1 个字段，got %d", len(schema.EntityTypes[0].Fields))
	}

	// state_defs.yaml 与 menu.yaml 也应生成
	for _, f := range []string{"state_defs.yaml", "menu.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("%s 未生成: %v", f, err)
		}
	}
}
