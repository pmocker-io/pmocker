import service from '@/utils/request'

// @Summary 配置包列表
// @Router /pmocker/config/packages [get]
export const listPackages = (params) => {
  return service({ url: '/pmocker/config/packages', method: 'get', params })
}

// @Summary 新建配置包
// @Router /pmocker/config/package [post]
export const createPackage = (data) => {
  return service({ url: '/pmocker/config/package', method: 'post', data })
}

// @Summary 配置包详情
// @Router /pmocker/config/package/:id [get]
export const getPackage = (id) => {
  return service({ url: `/pmocker/config/package/${id}`, method: 'get' })
}

// @Summary 更新配置包seed
// @Router /pmocker/config/package/:id [put]
export const updatePackageSeed = (id, data) => {
  return service({ url: `/pmocker/config/package/${id}`, method: 'put', data })
}

// @Summary 删除配置包
// @Router /pmocker/config/package/:id [delete]
export const deletePackage = (id) => {
  return service({ url: `/pmocker/config/package/${id}`, method: 'delete' })
}

// @Summary 复制配置包为草稿
// @Router /pmocker/config/package/:id/copy [post]
export const copyPackage = (id) => {
  return service({ url: `/pmocker/config/package/${id}/copy`, method: 'post' })
}

// @Summary 发布配置包
// @Router /pmocker/config/package/:id/publish [post]
export const publishPackage = (id) => {
  return service({ url: `/pmocker/config/package/${id}/publish`, method: 'post' })
}

// @Summary 配置包归档/恢复
// @Router /pmocker/config/package/:id/transition [post]
export const transitionPackage = (id, to) => {
  return service({ url: `/pmocker/config/package/${id}/transition`, method: 'post', params: { to } })
}

// @Summary 配置包版本历史
// @Router /pmocker/config/package/:id/versions [get]
export const listPackageVersions = (id) => {
  return service({ url: `/pmocker/config/package/${id}/versions`, method: 'get' })
}

// @Summary 配置包回滚
// @Router /pmocker/config/package/:id/rollback [post]
export const rollbackPackage = (id, versionId) => {
  return service({ url: `/pmocker/config/package/${id}/rollback`, method: 'post', params: { versionId } })
}

// @Summary 已发布状态流转
// @Router /pmocker/config/stateDefs/public [get]
export const listStateDefsPublic = (params) => {
  return service({ url: '/pmocker/config/stateDefs/public', method: 'get', params })
}

// @Summary 导出配置YAML
// @Router /pmocker/config/export [post]
export const exportConfig = () => {
  return service({ url: '/pmocker/config/export', method: 'post' })
}
