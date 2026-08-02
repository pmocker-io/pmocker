package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TimeEntryApi struct{}

func puint(s string) uint {
	u, _ := strconv.ParseUint(s, 10, 32)
	return uint(u)
}

func puintp(s string) *uint {
	if s == "" {
		return nil
	}
	u := puint(s)
	return &u
}

func (a *TimeEntryApi) CreateTimeEntry(c *gin.Context) {
	var e pmocker.PMTimeEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.CreateTimeEntry(e); err != nil {
		global.GVA_LOG.Error("创建失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *TimeEntryApi) UpdateTimeEntry(c *gin.Context) {
	var e pmocker.PMTimeEntry
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.UpdateTimeEntry(e); err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *TimeEntryApi) DeleteTimeEntry(c *gin.Context) {
	id := puint(c.Query("id"))
	if err := service.DeleteTimeEntry(id); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *TimeEntryApi) FindTimeEntry(c *gin.Context) {
	id := puint(c.Query("id"))
	e, err := service.GetTimeEntry(id)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(e, c)
}

func (a *TimeEntryApi) ListTimeEntries(c *gin.Context) {
	var pageInfo request.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	projectID := puint(c.Query("projectId"))
	userID := puintp(c.Query("userId"))
	status := c.Query("status")
	list, total, err := service.ListTimeEntries(projectID, userID, status, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize}, "成功", c)
}

func (a *TimeEntryApi) SubmitTimeEntry(c *gin.Context) {
	id := puint(c.Query("id"))
	if err := service.SubmitTimeEntry(id); err != nil {
		response.FailWithMessage("提交失败", c)
		return
	}
	response.OkWithMessage("提交成功", c)
}

func (a *TimeEntryApi) ApproveTimeEntry(c *gin.Context) {
	id := puint(c.Query("id"))
	approverID := puint(c.Query("approverId"))
	if err := service.ApproveTimeEntry(id, approverID); err != nil {
		response.FailWithMessage("审批失败", c)
		return
	}
	response.OkWithMessage("审批通过", c)
}

func (a *TimeEntryApi) RejectTimeEntry(c *gin.Context) {
	id := puint(c.Query("id"))
	approverID := puint(c.Query("approverId"))
	if err := service.RejectTimeEntry(id, approverID); err != nil {
		response.FailWithMessage("驳回失败", c)
		return
	}
	response.OkWithMessage("已驳回", c)
}

func (a *TimeEntryApi) UtilizationTimeEntry(c *gin.Context) {
	projectID := puint(c.Query("projectId"))
	userID := puint(c.Query("userId"))
	rate, err := service.CalcUtilization(projectID, userID)
	if err != nil {
		response.FailWithMessage("计算失败", c)
		return
	}
	response.OkWithData(map[string]float64{"utilization_rate": rate}, c)
}
