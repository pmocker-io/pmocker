// Package loader 提供插件元数据 YAML 加载器，将 pmocker/ 子目录的
// schema/seed/menu/api/workflows 灌入 EAV 元表和 gva 系统表。
package loader

import "github.com/flipped-aurora/gin-vue-admin/server/model/system"

// ManifestYaml 插件清单
type ManifestYaml struct {
	Code         string   `yaml:"code"`
	Name         string   `yaml:"name"`
	Version      string   `yaml:"version"`
	Description  string   `yaml:"description"`
	Dependencies []string `yaml:"dependencies"`
}

// SchemaYaml 对应 pmocker/schema.yaml
type SchemaYaml struct {
	EntityType string      `yaml:"entity_type"`
	Module     string      `yaml:"module"`
	Name       string      `yaml:"name"`
	Icon       string      `yaml:"icon"`
	IconColor  string      `yaml:"icon_color"`
	Fields     []FieldYaml `yaml:"fields"`
	States     []string    `yaml:"states"`
	Workflows  []string    `yaml:"workflows"`
}

type FieldYaml struct {
	FieldKey     string `yaml:"field_key"`
	FieldLabel   string `yaml:"field_label"`
	DataType     string `yaml:"data_type"`
	OptionsJSON  string `yaml:"options_json"`
	DefaultValue string `yaml:"default_value"`
	Validators   string `yaml:"validators"`
}

// SeedYaml 对应 pmocker/seed.yaml
type SeedYaml struct {
	Roles        []map[string]interface{} `yaml:"roles"`
	Dictionaries []DictYaml               `yaml:"dictionaries"`
	EPS          []EPSYaml                `yaml:"eps"`
}

type DictYaml struct {
	Type    string           `yaml:"type"`
	Name    string           `yaml:"name"`
	Details []DictDetailYaml `yaml:"details"`
}

type DictDetailYaml struct {
	Label  string `yaml:"label"`
	Value  string `yaml:"value"`
	Extend string `yaml:"extend"`
}

type EPSYaml struct {
	ParentPath string `yaml:"parent_path"`
	Name       string `yaml:"name"`
	Type       string `yaml:"type"`
}

// MenuYaml 对应 pmocker/menu.yaml
type MenuYaml struct {
	Menus []MenuEntryYaml `yaml:"menus"`
}

type MenuEntryYaml struct {
	Path       string `yaml:"path"`
	Name       string `yaml:"name"`
	Hidden     bool   `yaml:"hidden"`
	Component  string `yaml:"component"`
	Sort       int    `yaml:"sort"`
	Title      string `yaml:"title"`
	Icon       string `yaml:"icon"`
	ParentName string `yaml:"parent_name"`
}

// APIYaml 对应 pmocker/api.yaml
type APIYaml struct {
	APIs []system.SysApi `yaml:"apis"`
}
