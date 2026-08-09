package change

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) Create(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Title     string                 `json:"title"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := ServiceGroupApp.CreateChangeRequest(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.DeleteChangeRequest(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) Update(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	e.EntityType = "change_request"
	if err := ServiceGroupApp.UpdateChangeRequest(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) Find(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	e, err := ServiceGroupApp.GetChangeRequest(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if c.Query("page") != "" {
		offset = (page - 1) * pageSize
		limit = pageSize
	}
	list, total, err := ServiceGroupApp.ListChangeRequests(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) Analyze(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	report, err := ServiceGroupApp.ImpactAnalysis(c.Request.Context(), req.ID)
	if err != nil {
		response.FailWithMessage("影响分析失败: "+err.Error(), c)
		return
	}
	response.OkWithData(report, c)
}

func (a *Api) CCBReview(c *gin.Context) {
	var req struct {
		ID     uint `json:"id"`
		UserID uint `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	instID, err := ServiceGroupApp.SubmitToCCB(c.Request.Context(), req.ID, req.UserID)
	if err != nil {
		response.FailWithMessage("提交CCB审批失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"instanceId": instID}, c)
}

func (a *Api) Approve(c *gin.Context) {
	var req struct {
		ID       uint   `json:"id"`
		Decision string `json:"decision"`
		UserID   uint   `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.ApproveChange(c.Request.Context(), req.ID, req.Decision, req.UserID); err != nil {
		response.FailWithMessage("批准失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已批准", c)
}

func (a *Api) Reject(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id"`
		Reason string `json:"reason"`
		UserID uint   `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.RejectChange(c.Request.Context(), req.ID, req.Reason, req.UserID); err != nil {
		response.FailWithMessage("拒绝失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已拒绝", c)
}

func (a *Api) Implement(c *gin.Context) {
	var req struct {
		ID            uint `json:"id"`
		ImplementerID uint `json:"implementerId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.StartImplementation(c.Request.Context(), req.ID, req.ImplementerID); err != nil {
		response.FailWithMessage("开始实施失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已开始实施", c)
}

func (a *Api) Verify(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id"`
		Result string `json:"result"`
		UserID uint   `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.VerifyChange(c.Request.Context(), req.ID, req.Result, req.UserID); err != nil {
		response.FailWithMessage("验证失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("验证完成", c)
}

func (a *Api) Close(c *gin.Context) {
	var req struct {
		ID     uint `json:"id"`
		UserID uint `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.CloseChange(c.Request.Context(), req.ID, req.UserID); err != nil {
		response.FailWithMessage("关闭失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已关闭", c)
}

func (a *Api) ListLogs(c *gin.Context) {
	changeID, _ := strconv.ParseUint(c.Query("changeId"), 10, 64)
	logs, err := ServiceGroupApp.ListChangeLogs(c.Request.Context(), uint(changeID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(logs, c)
}

func (a *Api) ImpactReport(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	report, err := ServiceGroupApp.ImpactAnalysis(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("影响分析失败: "+err.Error(), c)
		return
	}
	response.OkWithData(report, c)
}

// GetDiff 变更前后字段级 diff 对比
func (a *Api) GetDiff(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	diffs, err := ServiceGroupApp.GetDiff(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("diff 查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(diffs, c)
}

func (a *Api) CCBStats(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	stats, err := ServiceGroupApp.CCBStats(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("统计失败: "+err.Error(), c)
		return
	}
	response.OkWithData(stats, c)
}
