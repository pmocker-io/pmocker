package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gopkg.in/yaml.v3"
)

// ExportService 导出 published 配置为 YAML 三件套（schema.yaml/state_defs.yaml/menu.yaml）
type ExportService struct{}

// entitySchemaYaml 与 loader.EntitySchemaYaml 对齐（多实体模式）
type entitySchemaYaml struct {
	EntityType string      `yaml:"entity_type"`
	Module     string      `yaml:"module"`
	Name       string      `yaml:"name"`
	Icon       string      `yaml:"icon"`
	IconColor  string      `yaml:"icon_color"`
	Fields     []fieldYaml `yaml:"fields"`
	States     []string    `yaml:"states,omitempty"`
}

type schemaYaml struct {
	EntityTypes []entitySchemaYaml `yaml:"entity_types"`
}

type fieldYaml struct {
	FieldKey     string `yaml:"field_key"`
	FieldLabel   string `yaml:"field_label"`
	DataType     string `yaml:"data_type"`
	OptionsJSON  string `yaml:"options_json,omitempty"`
	DefaultValue string `yaml:"default_value,omitempty"`
	Validators   string `yaml:"validators,omitempty"`
}

// Export 导出 published 配置到 destDir（schema.yaml/state_defs.yaml/menu.yaml）
func (s *ExportService) Export(ctx context.Context, destDir string) error {
	db := global.GVA_DB.WithContext(ctx)

	// 1. 实体类型 + 字段 -> schema.yaml（仅 published）
	var ets []pmocker.PMEntityType
	if err := db.Where("status = ?", "published").Order("id").Find(&ets).Error; err != nil {
		return err
	}
	schema := schemaYaml{EntityTypes: []entitySchemaYaml{}}
	for _, et := range ets {
		var fields []pmocker.PMFieldDef
		if err := db.Where("entity_type = ? AND status = ?", et.TypeCode, "published").Order("id").Find(&fields).Error; err != nil {
			return err
		}
		es := entitySchemaYaml{
			EntityType: et.TypeCode, Module: et.ModuleCode, Name: et.Name,
			Icon: et.Icon, IconColor: et.IconColor, Fields: []fieldYaml{},
		}
		for _, f := range fields {
			es.Fields = append(es.Fields, fieldYaml{
				FieldKey: f.FieldKey, FieldLabel: f.FieldLabel, DataType: f.DataType,
				OptionsJSON: f.OptionsJSON, DefaultValue: f.DefaultValue, Validators: f.Validators,
			})
		}
		schema.EntityTypes = append(schema.EntityTypes, es)
	}
	schemaBytes, err := yaml.Marshal(schema)
	if err != nil {
		return err
	}

	// 2. 状态流转定义 -> state_defs.yaml（仅 published config_status）
	var stateDefs []pmocker.PMStateDef
	if err := db.Where("config_status = ?", "published").Order("entity_type, sort").Find(&stateDefs).Error; err != nil {
		return err
	}
	stateBytes, err := yaml.Marshal(stateDefs)
	if err != nil {
		return err
	}

	// 3. 写文件
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "schema.yaml"), schemaBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "state_defs.yaml"), stateBytes, 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destDir, "menu.yaml"), []byte("menus: []\n"), 0644); err != nil {
		return err
	}
	return nil
}
