package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ChangeLogService struct{}

func (s *ChangeLogService) RecordChangeLog(entityID uint, fieldKey, oldValue, newValue string, changedBy uint, changeReqID *uint) error {
	if oldValue == newValue {
		return nil
	}
	return global.GVA_DB.Create(&pmocker.PMChangeLog{
		EntityID:        entityID,
		FieldKey:        fieldKey,
		OldValue:        oldValue,
		NewValue:        newValue,
		ChangedBy:       changedBy,
		ChangeRequestID: changeReqID,
	}).Error
}

func (s *ChangeLogService) ListChangeLogs(entityID uint) ([]pmocker.PMChangeLog, error) {
	var logs []pmocker.PMChangeLog
	err := global.GVA_DB.Where("entity_id = ?", entityID).Order("created_at DESC").Find(&logs).Error
	return logs, err
}
