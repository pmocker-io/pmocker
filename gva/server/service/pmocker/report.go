package pmocker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ReportService struct{}

// GenerateReportSnapshot 生成报告快照存入 pm_report_snapshots
// reportType: dashboard/pmo/close
// period: 如 "2026-06"（月报）或 "close"（结项）
func (s *ReportService) GenerateReportSnapshot(projectID uint, reportType string, period string, generatedBy uint) error {
	db := global.GVA_DB

	var snapshotData interface{}
	switch reportType {
	case "dashboard":
		ds := &DashboardService{}
		dash, err := ds.GetProjectDashboard(projectID)
		if err != nil {
			return err
		}
		snapshotData = dash
	case "pmo":
		ps := &PMODashboardService{}
		card, err := ps.GetProjectCard(projectID)
		if err != nil {
			return err
		}
		snapshotData = card
	case "close":
		as := &ArchiveService{}
		report, err := as.GetCloseReport(projectID)
		if err != nil {
			return err
		}
		snapshotData = report
	default:
		return fmt.Errorf("不支持的报告类型: %s", reportType)
	}

	jsonBytes, err := json.Marshal(snapshotData)
	if err != nil {
		return err
	}

	snapshot := pmocker.PMReportSnapshot{
		ProjectID:    projectID,
		ReportType:   reportType,
		Period:       period,
		SnapshotJSON: string(jsonBytes),
		GeneratedBy:  generatedBy,
	}
	return db.Create(&snapshot).Error
}

// GetReportSnapshots 查询项目的报告快照列表
func (s *ReportService) GetReportSnapshots(projectID uint, reportType string) ([]pmocker.PMReportSnapshot, error) {
	var snapshots []pmocker.PMReportSnapshot
	db := global.GVA_DB.Where("project_id = ?", projectID)
	if reportType != "" {
		db = db.Where("report_type = ?", reportType)
	}
	err := db.Order("created_at DESC").Find(&snapshots).Error
	return snapshots, err
}

// formatPeriod 格式化当前月份为 period 字符串
func formatPeriod(t time.Time) string {
	return t.Format("2006-01")
}
