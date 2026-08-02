package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type CostActualService struct{}

func (s *CostActualService) CreateCostActual(actual pmocker.PMCostActual) error {
	if actual.Status == "" {
		actual.Status = "pending"
	}
	return global.GVA_DB.Create(&actual).Error
}

func (s *CostActualService) UpdateCostActual(actual pmocker.PMCostActual) error {
	return global.GVA_DB.Save(&actual).Error
}

func (s *CostActualService) DeleteCostActual(id uint) error {
	return global.GVA_DB.Delete(&pmocker.PMCostActual{}, id).Error
}

func (s *CostActualService) GetCostActual(id uint) (pmocker.PMCostActual, error) {
	var a pmocker.PMCostActual
	err := global.GVA_DB.First(&a, id).Error
	return a, err
}

func (s *CostActualService) ListCostActuals(projectID uint, costType string, page, pageSize int) ([]pmocker.PMCostActual, int64, error) {
	var list []pmocker.PMCostActual
	var total int64
	db := global.GVA_DB.Model(&pmocker.PMCostActual{}).Where("project_id = ?", projectID)
	if costType != "" {
		db = db.Where("cost_type = ?", costType)
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("date DESC").Offset(offset).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *CostActualService) ConfirmCostActual(id uint) error {
	return global.GVA_DB.Model(&pmocker.PMCostActual{}).Where("id = ?", id).
		Update("status", "confirmed").Error
}

func (s *CostActualService) ApproveTimeEntryToCost(entry *pmocker.PMTimeEntry) error {
	actual := pmocker.PMCostActual{
		ProjectID:   entry.ProjectID,
		TaskID:      &entry.TaskID,
		CostType:    "labor",
		Amount:      entry.Cost,
		Date:        entry.Date,
		Source:      "time_entry",
		RefID:       &entry.ID,
		Description: "工时自动转化: " + entry.Description,
		Status:      "confirmed",
	}
	return global.GVA_DB.Create(&actual).Error
}
