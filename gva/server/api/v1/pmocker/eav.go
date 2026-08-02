package pmocker

import (
	"fmt"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

// EAVApi EAV 相关 API
type EAVApi struct{}

// CreateEntity 创建实体
func (a *EAVApi) CreateEntity(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := service.CreateEntity(c.Request.Context(), e)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

// GetEntity 获取实体
func (a *EAVApi) GetEntity(c *gin.Context) {
	id := parseUint(c.Param("id"))
	entity, err := service.GetEntity(c.Request.Context(), id)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(entity, c)
}

// ListEntities 列出实体
func (a *EAVApi) ListEntities(c *gin.Context) {
	projectID := parseUint(c.Query("projectId"))
	typeCode := c.Query("entityType")
	offset := parseInt(c.DefaultQuery("offset", "0"))
	limit := parseInt(c.DefaultQuery("limit", "20"))
	entities, total, err := service.ListEntities(c.Request.Context(), projectID, typeCode, offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": entities, "total": total}, "查询成功", c)
}

// RegisterSchema 注册 schema（元数据）
func (a *EAVApi) RegisterSchema(c *gin.Context) {
	var def eavtypes.SchemaDefinition
	if err := c.ShouldBindJSON(&def); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := service.RegisterEntityType(c.Request.Context(), eavtypes.EntityType{
		TypeCode:   def.EntityType,
		ModuleCode: def.Module,
		Name:       def.EntityType,
	}); err != nil {
		response.FailWithMessage("注册实体类型失败: "+err.Error(), c)
		return
	}
	for _, f := range def.Fields {
		if err := service.RegisterFieldDef(c.Request.Context(), f); err != nil {
			response.FailWithMessage("注册字段失败: "+err.Error(), c)
			return
		}
	}
	response.OkWithMessage("Schema 注册成功", c)
}

// GetSchema 获取实体类型的字段定义
func (a *EAVApi) GetSchema(c *gin.Context) {
	entityType := c.Param("entityType")
	et, err := service.LoadEntityType(c.Request.Context(), entityType)
	if err != nil {
		response.FailWithMessage("实体类型不存在: "+err.Error(), c)
		return
	}
	fields, err := service.LoadFieldDefs(c.Request.Context(), entityType)
	if err != nil {
		response.FailWithMessage("加载字段定义失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"entity_type": et, "fields": fields}, c)
}

// UpdateEntity 更新实体
func (a *EAVApi) UpdateEntity(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := service.UpdateEntity(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// DeleteEntity 删除实体
func (a *EAVApi) DeleteEntity(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := service.DeleteEntity(c.Request.Context(), id); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// parseUint 辅助函数：字符串转 uint
func parseUint(s string) uint {
	n, _ := strconv.ParseUint(s, 10, 64)
	return uint(n)
}

// parseInt 辅助函数：字符串转 int
func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// 确保 fmt 被使用（防止 import 报错）
var _ = fmt.Sprintf
