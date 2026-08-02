package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RelationApi struct{}

func (a *RelationApi) Create(c *gin.Context) {
	var rel pmocker.PMRelation
	if err := c.ShouldBindJSON(&rel); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.CreateRelation(rel); err != nil {
		global.GVA_LOG.Error("创建关联失败", zap.Error(err))
		response.FailWithMessage("创建关联失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *RelationApi) Delete(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if err := service.DeleteRelation(uint(id)); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *RelationApi) List(c *gin.Context) {
	entityIDStr := c.Query("entityId")
	entityID, _ := strconv.ParseUint(entityIDStr, 10, 32)
	direction := c.DefaultQuery("direction", "both")
	rels, err := service.ListPMRelations(uint(entityID), direction)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(rels, c)
}

func (a *RelationApi) CreateTaskLink(c *gin.Context) {
	var link pmocker.PMTaskLink
	if err := c.ShouldBindJSON(&link); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.CreateTaskLink(link); err != nil {
		global.GVA_LOG.Error("创建任务依赖失败", zap.Error(err))
		response.FailWithMessage("创建任务依赖失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *RelationApi) DeleteTaskLink(c *gin.Context) {
	idStr := c.Query("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if err := service.DeleteTaskLink(uint(id)); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *RelationApi) ListTaskLinks(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	projectID, _ := strconv.ParseUint(projectIDStr, 10, 32)
	links, err := service.ListTaskLinks(uint(projectID))
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(links, c)
}

func (a *RelationApi) ListChangeLogs(c *gin.Context) {
	entityIDStr := c.Query("entityId")
	entityID, _ := strconv.ParseUint(entityIDStr, 10, 32)
	logs, err := service.ListChangeLogs(uint(entityID))
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(logs, c)
}
