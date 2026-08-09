package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMStateDef 状态流转定义（实体类型→状态→标签/样式/流转动作）
type PMStateDef struct {
	global.GVA_MODEL
	EntityType   string `json:"entityType" gorm:"size:64;uniqueIndex:idx_state_def;not null;comment:实体类型"`
	Status       string `json:"status" gorm:"size:32;uniqueIndex:idx_state_def;not null;comment:状态值"`
	Label        string `json:"label" gorm:"size:64;comment:状态显示名"`
	TagType      string `json:"tagType" gorm:"size:16;comment:el-tag类型"`
	Sort         int    `json:"sort" gorm:"default:0;comment:排序"`
	ActionsJSON  string `json:"actionsJson" gorm:"type:text;comment:流转动作JSON"`
	ConfigStatus string `json:"configStatus" gorm:"size:16;default:published;comment:配置自身状态"`
}

func (PMStateDef) TableName() string { return "pm_state_defs" }
