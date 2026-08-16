package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// TestImportYAML 导入 seed.yaml（config_packages 数组）→ 幂等创建 draft 配置包
func TestImportYAML(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	is := &ImportService{}

	seedDoc := buildExportSeedDoc()

	res, err := is.ImportYAML(ctx, seedDoc)
	if err != nil {
		t.Fatalf("ImportYAML: %v", err)
	}
	if res.Imported != 1 {
		t.Fatalf("Imported = %d, want 1", res.Imported)
	}
	if res.Skipped != 0 {
		t.Fatalf("Skipped = %d, want 0", res.Skipped)
	}

	var pkg pmocker.PMConfigPackage
	if err := db.Where("code = ?", "imported_pkg").First(&pkg).Error; err != nil {
		t.Fatalf("导入的配置包不存在: %v", err)
	}
	if pkg.Status != "draft" {
		t.Fatalf("导入配置包状态 = %s, want draft", pkg.Status)
	}
	if pkg.SeedYAML != strings.TrimSpace(syncSeedYAML) {
		t.Fatal("导入的 seed_yaml 应与源一致（trim 后）")
	}

	// 幂等：重复导入跳过
	res2, err := is.ImportYAML(ctx, seedDoc)
	if err != nil {
		t.Fatalf("重复导入: %v", err)
	}
	if res2.Skipped != 1 {
		t.Fatalf("重复导入 Skipped = %d, want 1", res2.Skipped)
	}
	var count int64
	db.Model(&pmocker.PMConfigPackage{}).Count(&count)
	if count != 1 {
		t.Fatalf("幂等失败 count=%d, want 1", count)
	}
}

// buildExportSeedDoc 模拟 export_service 的 seed.yaml 输出格式（config_packages 对象数组）
func buildExportSeedDoc() string {
	// yaml 里 seed_yaml 用字面块 `|` 保留换行
	return "config_packages:\n  - code: imported_pkg\n    seed_yaml: |\n" + indentBlock(syncSeedYAML, "      ")
}

func indentBlock(s, pad string) string {
	out := ""
	for _, line := range splitLines(s) {
		out += pad + line + "\n"
	}
	return out
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, ch := range s {
		if ch == '\n' {
			lines = append(lines, cur)
			cur = ""
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// TestImportYAMLBadDoc 非法 YAML 导入应报错
func TestImportYAMLBadDoc(t *testing.T) {
	ctx := context.Background()
	testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	is := &ImportService{}

	if _, err := is.ImportYAML(ctx, "config_packages: [broken"); err == nil {
		t.Fatal("非法 YAML 导入应报错")
	}
}

// TestImportYAMLNoPackages 空 config_packages 导入应报错
func TestImportYAMLNoPackages(t *testing.T) {
	ctx := context.Background()
	testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	is := &ImportService{}

	if _, err := is.ImportYAML(ctx, "config_packages: []"); err == nil {
		t.Fatal("空 config_packages 导入应报错")
	}
}

// TestImportExportRoundTrip 导出→导入闭环：ExportService 产物能被 ImportYAML 原样还原
func TestImportExportRoundTrip(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewMemoryDB(t, &pmocker.PMConfigPackage{})
	is := &ImportService{}
	es := &ExportService{}

	// 建一个 published 配置包并导出
	if err := db.Create(&pmocker.PMConfigPackage{
		Code: "roundtrip", Name: "闭环测试", EntityType: "requirement", Module: "requirement",
		Status: "published", Version: 1, SeedYAML: syncSeedYAML,
	}).Error; err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := es.Export(ctx, dir); err != nil {
		t.Fatalf("Export: %v", err)
	}
	seedYAML, err := os.ReadFile(filepath.Join(dir, "seed.yaml"))
	if err != nil {
		t.Fatalf("seed.yaml 未生成: %v", err)
	}

	// 导入到新库（用新 MemoryDB 模拟目标实例）
	db2 := testutil.NewMemoryDBWithoutGlobal(t, &pmocker.PMConfigPackage{})
	global.GVA_DB = db2
	res, err := is.ImportYAML(ctx, string(seedYAML))
	if err != nil {
		t.Fatalf("ImportYAML(export 产物): %v\n=== seed.yaml 内容 ===\n%s", err, string(seedYAML))
	}
	if res.Imported != 1 {
		t.Fatalf("Imported = %d, want 1", res.Imported)
	}
	var pkg pmocker.PMConfigPackage
	if err := db2.Where("code = ?", "roundtrip").First(&pkg).Error; err != nil {
		t.Fatalf("导出→导入还原失败: %v", err)
	}
	if pkg.SeedYAML != strings.TrimSpace(syncSeedYAML) {
		t.Fatal("导入的 seed_yaml 应与源一致（trim 后）")
	}
}
