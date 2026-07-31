// Package diff 实现两个 .pmi 镜像的 schema/plugins/seed 差异对比。
package diff

import (
	"fmt"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"gopkg.in/yaml.v3"
)

// SchemaDiff schema 层差异
type SchemaDiff struct {
	AddedEntityTypes   []string      // 新增的实体类型
	RemovedEntityTypes []string      // 删除的实体类型
	FieldChanges       []FieldChange // 字段级变更
}

// FieldChange 字段变更
type FieldChange struct {
	EntityType  string `json:"entity_type"`
	FieldKey    string `json:"field_key"`
	ChangeType  string `json:"change_type"` // added / removed / modified
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
	ChangedAttr string `json:"changed_attr,omitempty"` // data_type / default_value / options_json / validators
}

// PluginDiff plugins 层差异
type PluginDiff struct {
	AddedModules   []string
	RemovedModules []string
}

// SeedDiff seed 层差异
type SeedDiff struct {
	AddedDictionaries   []string
	RemovedDictionaries []string
}

// DiffResult 完整 diff 结果
type DiffResult struct {
	Schema  SchemaDiff
	Plugins PluginDiff
	Seed    SeedDiff
}

// DiffSchemas 对比两个 SchemaYaml 的差异
func DiffSchemas(oldSchemas, newSchemas []loader.EntitySchemaYaml) *SchemaDiff {
	result := &SchemaDiff{
		FieldChanges: []FieldChange{},
	}
	oldMap := make(map[string]loader.EntitySchemaYaml)
	newMap := make(map[string]loader.EntitySchemaYaml)
	for _, s := range oldSchemas {
		oldMap[s.EntityType] = s
	}
	for _, s := range newSchemas {
		newMap[s.EntityType] = s
	}
	// 新增的实体类型
	for et := range newMap {
		if _, ok := oldMap[et]; !ok {
			result.AddedEntityTypes = append(result.AddedEntityTypes, et)
		}
	}
	// 删除的实体类型
	for et := range oldMap {
		if _, ok := newMap[et]; !ok {
			result.RemovedEntityTypes = append(result.RemovedEntityTypes, et)
		}
	}
	// 对比共有实体类型的字段
	for et, oldE := range oldMap {
		newE, ok := newMap[et]
		if !ok {
			continue
		}
		oldFields := fieldMap(oldE.Fields)
		newFields := fieldMap(newE.Fields)
		// 新增字段
		for fk := range newFields {
			if _, ok := oldFields[fk]; !ok {
				result.FieldChanges = append(result.FieldChanges, FieldChange{
					EntityType: et, FieldKey: fk, ChangeType: "added",
				})
			}
		}
		// 删除字段
		for fk := range oldFields {
			if _, ok := newFields[fk]; !ok {
				result.FieldChanges = append(result.FieldChanges, FieldChange{
					EntityType: et, FieldKey: fk, ChangeType: "removed",
				})
			}
		}
		// 修改字段
		for fk, oldF := range oldFields {
			newF, ok := newFields[fk]
			if !ok {
				continue
			}
			if oldF.DataType != newF.DataType {
				result.FieldChanges = append(result.FieldChanges, FieldChange{
					EntityType: et, FieldKey: fk, ChangeType: "modified",
					ChangedAttr: "data_type", OldValue: oldF.DataType, NewValue: newF.DataType,
				})
			}
			if oldF.DefaultValue != newF.DefaultValue {
				result.FieldChanges = append(result.FieldChanges, FieldChange{
					EntityType: et, FieldKey: fk, ChangeType: "modified",
					ChangedAttr: "default_value", OldValue: oldF.DefaultValue, NewValue: newF.DefaultValue,
				})
			}
			if oldF.OptionsJSON != newF.OptionsJSON {
				result.FieldChanges = append(result.FieldChanges, FieldChange{
					EntityType: et, FieldKey: fk, ChangeType: "modified",
					ChangedAttr: "options_json", OldValue: oldF.OptionsJSON, NewValue: newF.OptionsJSON,
				})
			}
		}
	}
	return result
}

func fieldMap(fields []loader.FieldYaml) map[string]loader.FieldYaml {
	m := make(map[string]loader.FieldYaml, len(fields))
	for _, f := range fields {
		m[f.FieldKey] = f
	}
	return m
}

// loadSchemasFromImage 从镜像中提取所有 schema.yaml 并解析
func loadSchemasFromImage(imgPath string) ([]loader.EntitySchemaYaml, error) {
	reader, err := oci.OpenImage(imgPath)
	if err != nil {
		return nil, err
	}
	files, err := reader.ExtractLayerFiles(oci.LayerTypeSchema)
	if err != nil {
		return nil, fmt.Errorf("extract schema layer: %w", err)
	}
	var allSchemas []loader.EntitySchemaYaml
	for _, data := range files {
		var s loader.SchemaYaml
		if err := yaml.Unmarshal(data, &s); err != nil {
			continue // 跳过无法解析的文件
		}
		if len(s.EntityTypes) > 0 {
			allSchemas = append(allSchemas, s.EntityTypes...)
		} else if s.EntityType != "" {
			allSchemas = append(allSchemas, loader.EntitySchemaYaml{
				EntityType: s.EntityType, Module: s.Module, Name: s.Name,
				Icon: s.Icon, IconColor: s.IconColor, Fields: s.Fields,
				States: s.States, Workflows: s.Workflows,
			})
		}
	}
	return allSchemas, nil
}

// DiffImages 对比两个 .pmi 镜像
func DiffImages(oldPath, newPath string) (*DiffResult, error) {
	oldSchemas, err := loadSchemasFromImage(oldPath)
	if err != nil {
		return nil, fmt.Errorf("load old schemas: %w", err)
	}
	newSchemas, err := loadSchemasFromImage(newPath)
	if err != nil {
		return nil, fmt.Errorf("load new schemas: %w", err)
	}
	result := &DiffResult{
		Schema: *DiffSchemas(oldSchemas, newSchemas),
	}
	return result, nil
}
