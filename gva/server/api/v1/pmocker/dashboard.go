package pmocker

import (
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardApi struct{}

// Get 项目仪表盘
// GET /pmocker/dashboard/get?projectId=xxx
func (a *DashboardApi) Get(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	if projectID == 0 {
		response.FailWithMessage("projectId 不能为空", c)
		return
	}
	dash, err := service.DashboardService.GetProjectDashboard(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("获取项目仪表盘失败", zap.Error(err))
		response.FailWithMessage("获取仪表盘失败", c)
		return
	}
	response.OkWithData(dash, c)
}

// Snapshot 生成报告快照
// POST /pmocker/report/snapshot  body: {projectId, reportType, period}
func (a *DashboardApi) Snapshot(c *gin.Context) {
	var req struct {
		ProjectID  uint   `json:"projectId"`
		ReportType string `json:"reportType"`
		Period     string `json:"period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Period == "" {
		req.Period = apiFormatPeriod(time.Now())
	}
	userID := utils.GetUserID(c)
	if err := service.ReportService.GenerateReportSnapshot(req.ProjectID, req.ReportType, req.Period, userID); err != nil {
		global.GVA_LOG.Error("生成报告快照失败", zap.Error(err))
		response.FailWithMessage("生成快照失败", c)
		return
	}
	response.OkWithMessage("快照已生成", c)
}

// ListSnapshots 查询报告快照列表
// GET /pmocker/report/list?projectId=xxx&reportType=dashboard
func (a *DashboardApi) ListSnapshots(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	reportType := c.Query("reportType")
	snapshots, err := service.ReportService.GetReportSnapshots(uint(projectID), reportType)
	if err != nil {
		response.FailWithMessage("查询快照失败", c)
		return
	}
	response.OkWithData(snapshots, c)
}

// apiFormatPeriod API 包内格式化当前月份为 period 字符串
func apiFormatPeriod(t time.Time) string {
	return t.Format("2006-01")
}
