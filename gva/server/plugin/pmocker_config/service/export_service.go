package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gopkg.in/yaml.v3"
)

// ExportService 导出 published 配置包为 YAML（schema.yaml/seed.yaml/menu.yaml）
type ExportService struct{}

// Export 导出所有 published 配置包的 seed_yaml 聚合到 destDir
func (s *ExportService) Export(ctx context.Context, destDir string) error {
	db := global.GVA_DB.WithContext(ctx)

	// 1. 查询所有 published 配置包
	var pkgs []pmocker.PMConfigPackage
	if err := db.Where("status = ?", "published").Order("id").Find(&pkgs).Error; err != nil {
		return err
	}

	// 2. 聚合每个配置包的 seed_yaml 到 schema（实体类型/字段/状态/流转）
	schemas := make([]map[string]interface{}, 0, len(pkgs))
	allSeeds := make([]string, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.SeedYAML == "" {
			continue
		}
		seed, err := ParseSeedYAML([]byte(pkg.SeedYAML))
		if err != nil {
			return err
		}
		allSeeds = append(allSeeds, pkg.SeedYAML)
		schemas = append(schemas, map[string]interface{}{
			"entity_type": seed.EntityType,
			"module":      seed.Module,
			"name":        seed.Name,
			"fields":      seed.Fields,
			"states":      seed.States,
			"transitions": seed.Transitions,
		})
	}

	// 3. 写 schema.yaml（配置包聚合定义）
	schemaDoc := map[string]interface{}{"entity_types": schemas}
	schemaBytes, err := yaml.Marshal(schemaDoc)
	if err != nil {
		return err
	}

	// 4. 写 seed.yaml（项目种子聚合）
	seedDoc := map[string]interface{}{"config_packages": allSeeds}
	seedBytes, err := yaml.Marshal(seedDoc)
	if err != nil {
		return err
	}

	// 5. 写文件
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "schema.yaml"), schemaBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "seed.yaml"), seedBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "menu.yaml"), []byte("menus: []\n"), 0644); err != nil {
		return err
	}
	return nil
}
