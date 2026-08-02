package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMEntity 实体主表（所有模块共用，靠 EntityType 区分）
type PMEntity struct {
	global.GVA_MODEL
	ProjectID  uint   `json:"projectId" gorm:"index;not null;comment:EPS项目节点ID"`
	EntityType string `json:"entityType" gorm:"size:64;index;not null;comment:实体类型"`
	ParentID   *uint  `json:"parentId" gorm:"index;comment:父节点ID"`
	Title      string `json:"title" gorm:"size:255;not null;comment:标题"`
	Status     string `json:"status" gorm:"size:32;index;comment:状态机当前态"`
	OwnerID    *uint  `json:"ownerId" gorm:"index;comment:责任人ID"`
	BaselineID *uint  `json:"baselineId" gorm:"index;comment:当前基线ID"`
	Seq        int    `json:"seq" gorm:"default:0;comment:排序"`
	Priority   int    `json:"priority" gorm:"default:2;index;comment:优先级0紧急1高2中3低"`
	CreatedBy  uint   `json:"createdBy" gorm:"comment:创建人"`
}

func (PMEntity) TableName() string { return "pm_entities" }

// PMAttr 属性表（EAV 动态字段值）
type PMAttr struct {
	global.GVA_MODEL
	EntityID    uint     `json:"entityId" gorm:"uniqueIndex:idx_attr;not null;comment:实体ID"`
	FieldKey    string   `json:"fieldKey" gorm:"size:64;uniqueIndex:idx_attr;not null;comment:字段键"`
	ValString   *string  `json:"valString" gorm:"type:text;comment:字符串值"`
	ValInt      *int64   `json:"valInt" gorm:"comment:整数值"`
	ValDecimal  *float64 `json:"valDecimal" gorm:"type:decimal(18,4);comment:小数值"`
	ValDate     *string  `json:"valDate" gorm:"size:10;comment:日期值"`
	ValDateTime *string  `json:"valDateTime" gorm:"size:19;comment:日期时间值"`
	ValBool     *bool    `json:"valBool" gorm:"comment:布尔值"`
	ValJSON     *string  `json:"valJson" gorm:"type:text;comment:JSON值"`
	ValRef      *uint    `json:"valRef" gorm:"index;comment:引用实体ID"`
}

func (PMAttr) TableName() string { return "pm_attrs" }

// PMRelation 实体关系表
type PMRelation struct {
	global.GVA_MODEL
	SrcID        uint   `json:"srcId" gorm:"uniqueIndex:idx_rel;not null;comment:源实体ID"`
	DstID        uint   `json:"dstId" gorm:"uniqueIndex:idx_rel;not null;comment:目标实体ID"`
	RelationType string `json:"relationType" gorm:"size:64;uniqueIndex:idx_rel;not null;comment:关系类型"`
}

func (PMRelation) TableName() string { return "pm_relations" }
