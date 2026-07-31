package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// PMEntityType 实体类型定义表（元表）
type PMEntityType struct {
	global.GVA_MODEL
	TypeCode   string `json:"typeCode" gorm:"uniqueIndex;size:64;comment:实体类型编码"`
	ModuleCode string `json:"moduleCode" gorm:"size:32;index;comment:所属模块"`
	Name       string `json:"name" gorm:"size:128;comment:名称"`
	Icon       string `json:"icon" gorm:"size:64;comment:图标"`
	IconColor  string `json:"iconColor" gorm:"size:16;comment:图标颜色"`
}

func (PMEntityType) TableName() string { return "pm_entity_types" }

// PMFieldDef 字段定义表（元表）
type PMFieldDef struct {
	global.GVA_MODEL
	EntityType    string `json:"entityType" gorm:"size:64;uniqueIndex:idx_field_def;comment:实体类型"`
	FieldKey      string `json:"fieldKey" gorm:"size:64;uniqueIndex:idx_field_def;comment:字段键"`
	FieldLabel    string `json:"fieldLabel" gorm:"size:128;comment:字段标签"`
	DataType      string `json:"dataType" gorm:"size:32;comment:数据类型"`
	OptionsJSON   string `json:"optionsJson" gorm:"type:text;comment:选项JSON"`
	DefaultValue  string `json:"defaultValue" gorm:"type:text;comment:默认值"`
	Validators    string `json:"validators" gorm:"type:text;comment:校验规则JSON"`
}

func (PMFieldDef) TableName() string { return "pm_field_defs" }

// PMRelationType 关系类型定义表（元表）
type PMRelationType struct {
	global.GVA_MODEL
	TypeCode    string `json:"typeCode" gorm:"uniqueIndex;size:64;comment:关系类型编码"`
	Name        string `json:"name" gorm:"size:128;comment:名称"`
	SrcType     string `json:"srcType" gorm:"size:64;comment:源实体类型"`
	DstType     string `json:"dstType" gorm:"size:64;comment:目标实体类型"`
	Directional bool   `json:"directional" gorm:"default:true;comment:是否有向"`
}

func (PMRelationType) TableName() string { return "pm_relation_types" }

// PMFieldVersion 字段版本追踪表
type PMFieldVersion struct {
	global.GVA_MODEL
	FieldDefID   uint   `json:"fieldDefId" gorm:"index;comment:字段定义ID"`
	ImageDigest  string `json:"imageDigest" gorm:"size:80;index;comment:镜像SHA256"`
	SnapshotJSON string `json:"snapshotJson" gorm:"type:text;comment:字段定义快照"`
}

func (PMFieldVersion) TableName() string { return "pm_field_versions" }
