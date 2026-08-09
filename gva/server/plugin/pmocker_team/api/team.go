package team

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	teamsvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team/service"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

type Api struct{}

// ---- 通用辅助方法 ----

func (a *Api) create(c *gin.Context, entityType string) {
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
	id, err := ServiceGroupApp.Create(c.Request.Context(), entityType, req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.Delete(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) update(c *gin.Context, entityType string) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	e.EntityType = entityType
	if err := ServiceGroupApp.Update(c.Request.Context(), entityType, e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) find(c *gin.Context, entityType string) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	e, err := ServiceGroupApp.Get(c.Request.Context(), entityType, uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) list(c *gin.Context, entityType string) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	// 同时兼容两种分页参数：offset/limit（后端约定）和 page/pageSize（前端约定）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if c.Query("page") != "" {
		offset = (page - 1) * pageSize
		limit = pageSize
	}
	list, total, err := ServiceGroupApp.List(c.Request.Context(), entityType, uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) transition(c *gin.Context, entityType string) {
	var req struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.Transition(c.Request.Context(), entityType, req.ID, req.Status); err != nil {
		response.FailWithMessage("状态流转失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("状态流转成功", c)
}

// ---- 团队成员 ----

func (a *Api) CreateMember(c *gin.Context)      { a.create(c, teamsvc.EntityTypeTeamMember) }
func (a *Api) DeleteMember(c *gin.Context)      { a.delete(c) }
func (a *Api) UpdateMember(c *gin.Context)      { a.update(c, teamsvc.EntityTypeTeamMember) }
func (a *Api) FindMember(c *gin.Context)        { a.find(c, teamsvc.EntityTypeTeamMember) }
func (a *Api) ListMember(c *gin.Context)        { a.list(c, teamsvc.EntityTypeTeamMember) }
func (a *Api) TransitionMember(c *gin.Context)  { a.transition(c, teamsvc.EntityTypeTeamMember) }

// ---- 角色定义 ----

func (a *Api) CreateRole(c *gin.Context)        { a.create(c, teamsvc.EntityTypeTeamRole) }
func (a *Api) DeleteRole(c *gin.Context)        { a.delete(c) }
func (a *Api) UpdateRole(c *gin.Context)        { a.update(c, teamsvc.EntityTypeTeamRole) }
func (a *Api) FindRole(c *gin.Context)          { a.find(c, teamsvc.EntityTypeTeamRole) }
func (a *Api) ListRole(c *gin.Context)          { a.list(c, teamsvc.EntityTypeTeamRole) }
func (a *Api) TransitionRole(c *gin.Context)    { a.transition(c, teamsvc.EntityTypeTeamRole) }

// ---- 培训记录 ----

func (a *Api) CreateTraining(c *gin.Context)      { a.create(c, teamsvc.EntityTypeTrainingRecord) }
func (a *Api) DeleteTraining(c *gin.Context)      { a.delete(c) }
func (a *Api) UpdateTraining(c *gin.Context)      { a.update(c, teamsvc.EntityTypeTrainingRecord) }
func (a *Api) FindTraining(c *gin.Context)        { a.find(c, teamsvc.EntityTypeTrainingRecord) }
func (a *Api) ListTraining(c *gin.Context)        { a.list(c, teamsvc.EntityTypeTrainingRecord) }
func (a *Api) TransitionTraining(c *gin.Context)  { a.transition(c, teamsvc.EntityTypeTrainingRecord) }

// ---- 绩效评估 ----

func (a *Api) CreatePerformance(c *gin.Context)      { a.create(c, teamsvc.EntityTypePerformanceReview) }
func (a *Api) DeletePerformance(c *gin.Context)      { a.delete(c) }
func (a *Api) UpdatePerformance(c *gin.Context)      { a.update(c, teamsvc.EntityTypePerformanceReview) }
func (a *Api) FindPerformance(c *gin.Context)        { a.find(c, teamsvc.EntityTypePerformanceReview) }
func (a *Api) ListPerformance(c *gin.Context)        { a.list(c, teamsvc.EntityTypePerformanceReview) }
func (a *Api) TransitionPerformance(c *gin.Context)  { a.transition(c, teamsvc.EntityTypePerformanceReview) }
