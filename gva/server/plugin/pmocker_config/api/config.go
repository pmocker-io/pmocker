package api

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	configService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/service"
	"github.com/gin-gonic/gin"
)

// ConfigApi 初始配置管理 API
type ConfigApi struct{}

var (
	cfgSvc    = configService.ConfigService{}
	stateSvc  = configService.StateMachineService{}
	exportSvc = configService.ExportService{}
)

// ListEntityTypes
// @Tags      PMockerConfig
// @Summary   实体类型列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     includeDraft query string false "是否包含草稿(true/false)"
// @Success   200  {object}  response.Response{data=[]pmocker.PMEntityType,msg=string}  "返回实体类型列表"
// @Router    /pmocker/config/entityTypes [get]
func (a *ConfigApi) ListEntityTypes(c *gin.Context) {
	includeDraft := c.Query("includeDraft") == "true"
	list, err := cfgSvc.ListEntityTypes(c.Request.Context(), includeDraft)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreateEntityType
// @Tags      PMockerConfig
// @Summary   新增实体类型
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     data  body      pmocker.PMEntityType  true  "实体类型定义"
// @Success   200   {object}  response.Response{msg=string}  "创建成功"
// @Router    /pmocker/config/entityType [post]
func (a *ConfigApi) CreateEntityType(c *gin.Context) {
	var et pmocker.PMEntityType
	if err := c.ShouldBindJSON(&et); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := cfgSvc.CreateEntityType(c.Request.Context(), et); err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// Transition
// @Tags      PMockerConfig
// @Summary   配置状态流转
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     table  query  string  true  "配置表名"
// @Param     id     query  integer  true  "配置ID"
// @Param     from   query  string  true  "源状态"
// @Param     to     query  string  true  "目标状态(delete 表示删除草稿)"
// @Success   200    {object}  response.Response{msg=string}  "状态流转成功"
// @Router    /pmocker/config/transition [post]
func (a *ConfigApi) Transition(c *gin.Context) {
	table := c.Query("table")
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	from := c.Query("from")
	to := c.Query("to")
	if to == "delete" {
		if err := stateSvc.DeleteDraft(global.GVA_DB, table, uint(id)); err != nil {
			response.FailWithMessage("删除失败: "+err.Error(), c)
			return
		}
		response.OkWithMessage("删除成功", c)
		return
	}
	if err := stateSvc.Transition(global.GVA_DB, table, uint(id), from, to); err != nil {
		response.FailWithMessage("状态流转失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("状态流转成功", c)
}

// CopyAsDraft
// @Tags      PMockerConfig
// @Summary   复制为草稿
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     table  query  string  true  "配置表名"
// @Param     id     query  integer  true  "配置ID"
// @Success   200    {object}  response.Response{msg=string}  "复制成功"
// @Router    /pmocker/config/copy [post]
func (a *ConfigApi) CopyAsDraft(c *gin.Context) {
	table := c.Query("table")
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := cfgSvc.CopyAsDraft(c.Request.Context(), table, uint(id)); err != nil {
		response.FailWithMessage("复制失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("复制成功", c)
}

// ListStateDefsPublic
// @Tags      PMockerConfig
// @Summary   已发布状态流转
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     entityType  query  string  false  "实体类型"
// @Success   200         {object}  response.Response{data=[]pmocker.PMStateDef,msg=string}  "返回已发布状态流转定义"
// @Router    /pmocker/config/stateDefs/public [get]
func (a *ConfigApi) ListStateDefsPublic(c *gin.Context) {
	entityType := c.Query("entityType")
	list, err := cfgSvc.ListStateDefs(c.Request.Context(), entityType, false)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// ListSeedEntities
// @Tags      PMockerConfig
// @Summary   业务种子列表
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Param     projectId   query  integer  false  "项目ID"
// @Param     entityType  query  string  true  "实体类型"
// @Param     page        query  integer  false  "页码"
// @Param     pageSize    query  integer  false  "每页数量"
// @Success   200         {object}  response.Response{data=gin.H{list=[]pmocker.PMEntity,total=int64},msg=string}  "返回列表与总数"
// @Router    /pmocker/config/seedEntities [get]
func (a *ConfigApi) ListSeedEntities(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	entityType := c.Query("entityType")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset := (page - 1) * pageSize
	list, total, err := cfgSvc.ListSeedEntities(c.Request.Context(), uint(projectID), entityType, offset, pageSize)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

// Export
// @Tags      PMockerConfig
// @Summary   导出配置YAML
// @Security  ApiKeyAuth
// @accept    application/json
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "导出成功"
// @Router    /pmocker/config/export [post]
func (a *ConfigApi) Export(c *gin.Context) {
	if err := exportSvc.Export(c.Request.Context(), "images/pmbok6-hybrid"); err != nil {
		response.FailWithMessage("导出失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("导出成功", c)
}
