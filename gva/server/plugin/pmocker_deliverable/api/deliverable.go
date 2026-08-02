package deliverable

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
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
	id, err := ServiceGroupApp.CreateDeliverable(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
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
	if err := ServiceGroupApp.DeleteDeliverable(c.Request.Context(), uint(id)); err != nil {
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
	if err := ServiceGroupApp.UpdateDeliverable(c.Request.Context(), e, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// CheckOut 检出交付物（排他锁定）
func (a *Api) CheckOut(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if req.ID == 0 {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.CheckOut(c.Request.Context(), req.ID, utils.GetUserID(c)); err != nil {
		response.FailWithMessage("检出失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("检出成功", c)
}

// CheckIn 检入交付物（解锁并可选记录版本）
func (a *Api) CheckIn(c *gin.Context) {
	var req struct {
		ID          uint   `json:"id"`
		VersionNote string `json:"versionNote"`
		FileRef     string `json:"fileRef"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if req.ID == 0 {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.CheckIn(c.Request.Context(), req.ID, utils.GetUserID(c), req.VersionNote, req.FileRef, utils.GetUserAuthorityId(c)); err != nil {
		response.FailWithMessage("检入失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("检入成功", c)
}

func (a *Api) Find(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	e, err := ServiceGroupApp.GetDeliverable(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := ServiceGroupApp.ListDeliverables(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) SubmitReview(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.SubmitForReview(c.Request.Context(), req.ID); err != nil {
		response.FailWithMessage("提交评审失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已提交评审", c)
}

func (a *Api) Accept(c *gin.Context) {
	var req struct {
		ID              uint   `json:"id"`
		ReviewerComment string `json:"reviewerComment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.AcceptDeliverable(c.Request.Context(), req.ID, req.ReviewerComment); err != nil {
		response.FailWithMessage("验收失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("验收通过", c)
}

func (a *Api) Reject(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.RejectDeliverable(c.Request.Context(), req.ID, req.Reason); err != nil {
		response.FailWithMessage("驳回失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已驳回", c)
}

func (a *Api) CreateVersion(c *gin.Context) {
	var req struct {
		DeliverableID uint   `json:"deliverableId"`
		Version       string `json:"version"`
		ChangeLog     string `json:"changeLog"`
		CreatorID     uint   `json:"creatorId"`
		FileRef       string `json:"fileRef"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	result, err := ServiceGroupApp.CreateVersion(c.Request.Context(), req.DeliverableID, req.Version, req.ChangeLog, req.CreatorID, req.FileRef)
	if err != nil {
		response.FailWithMessage("创建版本失败: "+err.Error(), c)
		return
	}
	response.OkWithData(result, c)
}

func (a *Api) ListVersions(c *gin.Context) {
	deliverableID, _ := strconv.ParseUint(c.DefaultQuery("deliverableId", "0"), 10, 64)
	list, err := ServiceGroupApp.ListVersions(c.Request.Context(), uint(deliverableID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"list": list}, c)
}

func (a *Api) Baseline(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.BaselineDeliverable(c.Request.Context(), req.ID); err != nil {
		response.FailWithMessage("基线化失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已基线化", c)
}

func (a *Api) TraceReport(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	report, err := ServiceGroupApp.TraceabilityReport(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("生成追溯报告失败: "+err.Error(), c)
		return
	}
	response.OkWithData(report, c)
}

func (a *Api) Stats(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	stats, err := ServiceGroupApp.DeliverableStats(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("统计失败: "+err.Error(), c)
		return
	}
	response.OkWithData(stats, c)
}
