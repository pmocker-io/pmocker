package pmocker

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type TimeEntryService struct{}

func (s *TimeEntryService) CreateTimeEntry(entry pmocker.PMTimeEntry) error {
	entry.Status = "draft"
	entry.Cost = entry.Hours * entry.HourlyRate
	return global.GVA_DB.Create(&entry).Error
}

func (s *TimeEntryService) UpdateTimeEntry(entry pmocker.PMTimeEntry) error {
	entry.Cost = entry.Hours * entry.HourlyRate
	return global.GVA_DB.Save(&entry).Error
}

func (s *TimeEntryService) DeleteTimeEntry(id uint) error {
	return global.GVA_DB.Delete(&pmocker.PMTimeEntry{}, id).Error
}

func (s *TimeEntryService) GetTimeEntry(id uint) (pmocker.PMTimeEntry, error) {
	var e pmocker.PMTimeEntry
	err := global.GVA_DB.First(&e, id).Error
	return e, err
}

func (s *TimeEntryService) ListTimeEntries(projectID uint, userID *uint, status string, page, pageSize int) ([]pmocker.PMTimeEntry, int64, error) {
	var entries []pmocker.PMTimeEntry
	var total int64
	db := global.GVA_DB.Model(&pmocker.PMTimeEntry{}).Where("project_id = ?", projectID)
	if userID != nil {
		db = db.Where("user_id = ?", *userID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}
	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Order("date DESC").Offset(offset).Limit(pageSize).Find(&entries).Error
	return entries, total, err
}

func (s *TimeEntryService) SubmitTimeEntry(id uint) error {
	return global.GVA_DB.Model(&pmocker.PMTimeEntry{}).Where("id = ?", id).
		Update("status", "submitted").Error
}

func (s *TimeEntryService) ApproveTimeEntry(id uint, approverID uint) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	return global.GVA_DB.Model(&pmocker.PMTimeEntry{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "approved",
			"approver_id": approverID,
			"approved_at": now,
		}).Error
}

func (s *TimeEntryService) RejectTimeEntry(id uint, approverID uint) error {
	return global.GVA_DB.Model(&pmocker.PMTimeEntry{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      "rejected",
			"approver_id": approverID,
		}).Error
}

func (s *TimeEntryService) CalcUtilization(projectID uint, userID uint) (float64, error) {
	var totalHours float64
	global.GVA_DB.Model(&pmocker.PMTimeEntry{}).
		Where("project_id = ? AND user_id = ? AND status = ?", projectID, userID, "approved").
		Select("COALESCE(SUM(hours), 0)").Scan(&totalHours)
	plannedHours := 160.0
	if plannedHours == 0 {
		return 0, nil
	}
	return totalHours / plannedHours * 100, nil
}
