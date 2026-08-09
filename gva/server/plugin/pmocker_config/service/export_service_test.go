package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

func TestExportConfigPackages(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	// 一个 published 配置包 + 一个 draft 配置包
	db.Create(&pmocker.PMConfigPackage{
		Code: "requirement", Name: "需求管理", EntityType: "requirement", Module: "requirement",
		Status: "published", Version: 1, SeedYAML: testSeedYAML,
	})
	db.Create(&pmocker.PMConfigPackage{
		Code: "draft-pkg", Name: "草稿包", EntityType: "risk", Module: "risk",
		Status: "draft", Version: 0, SeedYAML: "entity_type: risk\nmodule: risk\nname: 草稿\n",
	})

	s := &ExportService{}
	dir := t.TempDir()
	if err := s.Export(context.Background(), dir); err != nil {
		t.Fatalf("Export: %v", err)
	}

	// 只导出 published 配置包
	b, err := os.ReadFile(filepath.Join(dir, "schema.yaml"))
	if err != nil {
		t.Fatalf("schema.yaml 未生成: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("schema.yaml 为空")
	}

	// seed.yaml 也应生成
	if _, err := os.Stat(filepath.Join(dir, "seed.yaml")); err != nil {
		t.Fatalf("seed.yaml 未生成: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "menu.yaml")); err != nil {
		t.Fatalf("menu.yaml 未生成: %v", err)
	}
}
