import service from '@/utils/request'

// @Summary 实体类型列表
// @Router /pmocker/config/entityTypes [get]
export const listEntityTypes = (params) => {
  return service({ url: '/pmocker/config/entityTypes', method: 'get', params })
}

// @Summary 新增实体类型
// @Router /pmocker/config/entityType [post]
export const createEntityType = (data) => {
  return service({ url: '/pmocker/config/entityType', method: 'post', data })
}

// @Summary 配置状态流转（to=delete 表示删除草稿）
// @Router /pmocker/config/transition [post]
export const transitionConfig = (params) => {
  return service({ url: '/pmocker/config/transition', method: 'post', params })
}

// @Summary 复制为草稿
// @Router /pmocker/config/copy [post]
export const copyAsDraft = (params) => {
  return service({ url: '/pmocker/config/copy', method: 'post', params })
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
