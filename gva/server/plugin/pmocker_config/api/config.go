package api

import (
	"os"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	configService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/service"
	"github.com/gin-gonic/gin"
)

// ConfigApi 初始配置管理 API（配置包模型）
type ConfigApi struct{}

var pkgSvc = &configService.ConfigPackageService{}
var verSvc = &configService.ConfigVersionService{}
var exportSvc = &configService.ExportService{}

// ListPackages 配置包列表
// @Summary 配置包列表
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param includeDraft query bool false "包含草稿"
// @Success 200 {object} response.Response{data=[]pmocker.PMConfigPackage,msg=string}
// @Router /pmocker/config/packages [get]
func (a *ConfigApi) ListPackages(c *gin.Context) {
	includeDraft := c.Query("includeDraft") == "true"
	list, err := pkgSvc.List(c.Request.Context(), includeDraft)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// CreatePackage 新建配置包（draft）
// @Summary 新建配置包
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param data body pmocker.PMConfigPackage true "配置包"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package [post]
func (a *ConfigApi) CreatePackage(c *gin.Context) {
	var pkg pmocker.PMConfigPackage
	if err := c.ShouldBindJSON(&pkg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pkgSvc.Create(c.Request.Context(), pkg); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// GetPackage 获取配置包详情
// @Summary 获取配置包详情
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{data=pmocker.PMConfigPackage,msg=string}
// @Router /pmocker/config/package/{id} [get]
func (a *ConfigApi) GetPackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	pkg, err := pkgSvc.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(pkg, c)
}

// UpdatePackageSeed 更新配置包 seed_yaml
// @Summary 更新配置包种子数据
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Param data body object true "seedYaml"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id} [put]
func (a *ConfigApi) UpdatePackageSeed(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req struct {
		SeedYAML string `json:"seedYaml"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pkgSvc.UpdateSeed(c.Request.Context(), uint(id), req.SeedYAML); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// CopyPackage 复制为草稿
// @Summary 复制配置包为草稿
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id}/copy [post]
func (a *ConfigApi) CopyPackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := pkgSvc.CopyAsDraft(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已复制为草稿", c)
}

// PublishPackage 发布配置包（同步 DB）
// @Summary 发布配置包
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id}/publish [post]
func (a *ConfigApi) PublishPackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := pkgSvc.Publish(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("发布失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("发布成功", c)
}

// TransitionPackage 配置包状态流转（archive/restore）
// @Summary 配置包状态流转
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Param to query string true "目标状态 archived/restore"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id}/transition [post]
func (a *ConfigApi) TransitionPackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	to := c.Query("to")
	switch to {
	case "archived":
		if err := pkgSvc.Archive(c.Request.Context(), uint(id)); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	case "restore":
		if err := pkgSvc.Restore(c.Request.Context(), uint(id)); err != nil {
			response.FailWithMessage(err.Error(), c)
			return
		}
	default:
		response.FailWithMessage("不支持的流转目标: "+to, c)
		return
	}
	response.OkWithMessage("流转成功", c)
}

// DeletePackage 删除配置包（仅 draft）
// @Summary 删除配置包
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id} [delete]
func (a *ConfigApi) DeletePackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := pkgSvc.Delete(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// ListPackageVersions 配置包版本历史
// @Summary 配置包版本历史
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{data=[]pmocker.PMConfigVersion,msg=string}
// @Router /pmocker/config/package/{id}/versions [get]
func (a *ConfigApi) ListPackageVersions(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versions, err := verSvc.ListVersions(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(versions, c)
}

// RollbackPackage 回滚到指定版本
// @Summary 配置包回滚
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Param versionId query int true "版本ID"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id}/rollback [post]
func (a *ConfigApi) RollbackPackage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	versionID, _ := strconv.ParseUint(c.Query("versionId"), 10, 64)
	if err := verSvc.Rollback(c.Request.Context(), uint(id), uint(versionID)); err != nil {
		response.FailWithMessage("回滚失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("回滚成功", c)
}

// ListStateDefsPublic 已发布状态流转（前端 statusTransitions 读取）
// @Summary 已发布状态流转
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param entityType query string false "实体类型"
// @Success 200 {object} response.Response{data=[]pmocker.PMStateDef,msg=string}
// @Router /pmocker/config/stateDefs/public [get]
func (a *ConfigApi) ListStateDefsPublic(c *gin.Context) {
	list, err := pkgSvc.ListStateDefsPublic(c.Request.Context(), c.Query("entityType"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

// Export 导出配置 YAML 到镜像源
// @Summary 导出配置YAML到镜像源
// @Tags 初始配置
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/export [post]
func (a *ConfigApi) Export(c *gin.Context) {
	destDir := os.Getenv("PMOCKER_EXPORT_DIR")
	if destDir == "" {
		destDir = "images/pmbok6-hybrid"
	}
	if err := exportSvc.Export(c.Request.Context(), destDir); err != nil {
		response.FailWithMessage("导出失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("导出成功: "+destDir, c)
}

// GetPackageSeedStruct 获取配置包结构化种子（前端层级编辑器）
// @Summary 获取配置包结构化种子
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Success 200 {object} response.Response{data=object,msg=string}
// @Router /pmocker/config/package/{id}/seed [get]
func (a *ConfigApi) GetPackageSeedStruct(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	seed, err := pkgSvc.GetSeedStruct(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithData(seed, c)
}

// UpdatePackageSeedStruct 保存配置包结构化种子（前端层级编辑器）
// @Summary 保存配置包结构化种子
// @Tags 初始配置
// @Security ApiKeyAuth
// @Param id path int true "配置包ID"
// @Param data body object true "结构化 seed"
// @Success 200 {object} response.Response{msg=string}
// @Router /pmocker/config/package/{id}/seed [put]
func (a *ConfigApi) UpdatePackageSeedStruct(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var seed configService.ConfigPackageSeed
	if err := c.ShouldBindJSON(&seed); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := pkgSvc.UpdateSeedStruct(c.Request.Context(), uint(id), &seed); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已保存", c)
}

var _ = global.GVA_DB
