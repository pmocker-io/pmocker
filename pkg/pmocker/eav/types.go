// Package eav 定义 PMocker 的 EAV（Entity-Attribute-Value）数据模型类型。
// 这些是纯 Go 类型，不依赖 GORM 或 gin-vue-admin，由 gva 层实现 GORM 绑定。
package eav

import "time"

// DataType EAV 字段支持的数据类型
type DataType string

const (
	DataTypeString   DataType = "string"
	DataTypeText     DataType = "text"
	DataTypeInt      DataType = "int"
	DataTypeDecimal  DataType = "decimal"
	DataTypeDate     DataType = "date"
	DataTypeDateTime DataType = "datetime"
	DataTypeBool     DataType = "bool"
	DataTypeEnum     DataType = "enum"
	DataTypeRef      DataType = "ref"
	DataTypeJSON     DataType = "json"
)

// EntityType 实体类型定义（对应 pm_entity_types 表）
type EntityType struct {
	TypeCode   string `json:"type_code" yaml:"type_code"`
	ModuleCode string `json:"module_code" yaml:"module_code"`
	Name       string `json:"name" yaml:"name"`
	Icon       string `json:"icon" yaml:"icon"`
	IconColor  string `json:"icon_color" yaml:"icon_color"`
	Status     string `json:"status" yaml:"status"`
}

// FieldDef 字段定义（对应 pm_field_defs 表）
type FieldDef struct {
	EntityType    string   `json:"entity_type" yaml:"entity_type"`
	FieldKey      string   `json:"field_key" yaml:"field_key"`
	FieldLabel    string   `json:"field_label" yaml:"field_label"`
	DataType      DataType `json:"data_type" yaml:"data_type"`
	OptionsJSON   string   `json:"options_json,omitempty" yaml:"options_json,omitempty"`
	DefaultValue  string   `json:"default_value,omitempty" yaml:"default_value,omitempty"`
	Validators    string   `json:"validators,omitempty" yaml:"validators,omitempty"`
	Status        string   `json:"status" yaml:"status"`
}

// RelationType 关系类型定义（对应 pm_relation_types 表）
type RelationType struct {
	TypeCode    string `json:"type_code" yaml:"type_code"`
	Name        string `json:"name" yaml:"name"`
	SrcType     string `json:"src_type" yaml:"src_type"`
	DstType     string `json:"dst_type" yaml:"dst_type"`
	Directional bool   `json:"directional" yaml:"directional"`
}

// Entity 实体数据（业务层通用结构）
type Entity struct {
	ID           uint                   `json:"id"`
	ProjectID    uint                   `json:"project_id"`
	EntityType   string                 `json:"entity_type"`
	ParentID     *uint                  `json:"parent_id,omitempty"`
	Title        string                 `json:"title"`
	Status       string                 `json:"status"`
	OwnerID      *uint                  `json:"owner_id,omitempty"`
	OwnerName    string                 `json:"ownerName,omitempty"`
	BaselineID   *uint                  `json:"baseline_id,omitempty"`
	Seq          int                    `json:"seq"`
	CreatedBy    uint                   `json:"created_by,omitempty"`
	CreatedByName string                `json:"createdByName,omitempty"`
	Attrs        map[string]interface{} `json:"attrs,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}
