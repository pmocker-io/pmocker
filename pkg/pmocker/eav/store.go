package eav

import "context"

// EAVStore 定义 EAV 数据存储接口。
// gva 层实现此接口（用 GORM 操作 pm_entities/pm_attrs 等表）。
type EAVStore interface {
	// 元数据
	LoadEntityType(ctx context.Context, typeCode string) (*EntityType, error)
	LoadFieldDefs(ctx context.Context, typeCode string) ([]FieldDef, error)
	RegisterEntityType(ctx context.Context, et EntityType) error
	RegisterFieldDef(ctx context.Context, fd FieldDef) error

	// 实体 CRUD
	CreateEntity(ctx context.Context, e Entity) (uint, error)
	GetEntity(ctx context.Context, id uint) (*Entity, error)
	UpdateEntity(ctx context.Context, e Entity) error
	DeleteEntity(ctx context.Context, id uint) error
	ListEntities(ctx context.Context, projectID uint, typeCode string, offset, limit int) ([]Entity, int64, error)

	// 关系
	AddRelation(ctx context.Context, srcID, dstID uint, relType string) error
	ListRelations(ctx context.Context, entityID uint) ([]Relation, error)
}

// Relation 实体关系
type Relation struct {
	ID           uint   `json:"id"`
	SrcID        uint   `json:"src_id"`
	DstID        uint   `json:"dst_id"`
	RelationType string `json:"relation_type"`
}
