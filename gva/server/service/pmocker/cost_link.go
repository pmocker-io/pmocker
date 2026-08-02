package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type CostLinkService struct{}

func (s *CostLinkService) SyncTaskCostBudget(taskID uint) error {
	db := global.GVA_DB

	var task pmocker.PMEntity
	if err := db.First(&task, taskID).Error; err != nil {
		return err
	}
	if task.OwnerID == nil {
		return nil
	}

	var hoursAttr pmocker.PMAttr
	estimatedHours := 0.0
	if err := db.Where("entity_id = ? AND field_key = ?", taskID, "estimated_hours").
		First(&hoursAttr).Error; err == nil && hoursAttr.ValDecimal != nil {
		estimatedHours = *hoursAttr.ValDecimal
	}

	var rateAttr pmocker.PMAttr
	hourlyRate := 0.0
	if err := db.Where("entity_id = ? AND field_key = ?", *task.OwnerID, "hourly_rate").
		First(&rateAttr).Error; err == nil && rateAttr.ValDecimal != nil {
		hourlyRate = *rateAttr.ValDecimal
	}

	if estimatedHours == 0 || hourlyRate == 0 {
		return nil
	}

	budget := estimatedHours * hourlyRate

	var existing pmocker.PMAttr
	err := db.Where("entity_id = ? AND field_key = ?", taskID, "budget_cost").
		First(&existing).Error
	if err != nil {
		budgetVal := budget
		db.Create(&pmocker.PMAttr{
			EntityID: taskID, FieldKey: "budget_cost", ValDecimal: &budgetVal,
		})
	} else {
		budgetVal := budget
		db.Model(&pmocker.PMAttr{}).Where("id = ?", existing.ID).
			Update("val_decimal", budgetVal)
	}
	return nil
}
