package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type RelationService struct{}

func (s *RelationService) CreateRelation(rel pmocker.PMRelation) error {
	return global.GVA_DB.Create(&rel).Error
}

func (s *RelationService) DeleteRelation(id uint) error {
	return global.GVA_DB.Delete(&pmocker.PMRelation{}, id).Error
}

func (s *RelationService) ListPMRelations(entityID uint, direction string) ([]pmocker.PMRelation, error) {
	var rels []pmocker.PMRelation
	db := global.GVA_DB.Model(&pmocker.PMRelation{})
	switch direction {
	case "out":
		db = db.Where("src_id = ?", entityID)
	case "in":
		db = db.Where("dst_id = ?", entityID)
	default:
		db = db.Where("src_id = ? OR dst_id = ?", entityID, entityID)
	}
	err := db.Find(&rels).Error
	return rels, err
}

func (s *RelationService) ListPMRelationsByType(entityID uint, relationType string) ([]pmocker.PMRelation, error) {
	var rels []pmocker.PMRelation
	err := global.GVA_DB.Where("(src_id = ? OR dst_id = ?) AND relation_type = ?", entityID, entityID, relationType).Find(&rels).Error
	return rels, err
}
