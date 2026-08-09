package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMConfigPackage 配置包：聚合实体类型+字段+种子数据(含项目)+状态定义+流转规则
type PMConfigPackage struct {
	global.GVA_MODEL
	Code        string `json:"code" gorm:"size:64;uniqueIndex;not null;comment:配置包编码"`
	Name        string `json:"name" gorm:"size:128;comment:显示名"`
	Description string `json:"description" gorm:"type:text;comment:描述"`
	EntityType  string `json:"entityType" gorm:"size:64;index;comment:实体类型"`
	Module      string `json:"module" gorm:"size:32;index;comment:所属模块"`
	Version     int    `json:"version" gorm:"default:1;comment:当前版本号"`
	Status      string `json:"status" gorm:"size:16;default:draft;comment:draft/reviewing/published/archived"`
	SeedYAML    string `json:"seedYaml" gorm:"type:text;comment:种子数据YAML真源"`
}

func (PMConfigPackage) TableName() string { return "pm_config_packages" }

// PMConfigVersion 配置包版本快照（发布时生成，不可变）
type PMConfigVersion struct {
	global.GVA_MODEL
	PackageID    uint   `json:"packageId" gorm:"index;not null;comment:配置包ID"`
	Version      int    `json:"version" gorm:"comment:版本号"`
	SnapshotYAML string `json:"snapshotYaml" gorm:"type:text;comment:版本快照YAML"`
	Flag         int    `json:"flag" gorm:"default:0;comment:0=发布 1=回滚"`
}

func (PMConfigVersion) TableName() string { return "pm_config_versions" }
