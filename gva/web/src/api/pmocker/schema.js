import service from '@/utils/request'

// @Summary 获取实体类型 schema（字段定义）
// @Router /eav/schema/:entityType [get]
export const getSchema = (entityType) => {
  return service({ url: `/pmocker/eav/schema/${entityType}`, method: 'get' })
}

// @Summary 创建实体
// @Router /eav/entity [post]
export const createEntity = (data) => {
  return service({ url: '/pmocker/eav/entity', method: 'post', data })
}

// @Summary 更新实体
// @Router /eav/entity [put]
export const updateEntity = (data) => {
  return service({ url: '/pmocker/eav/entity', method: 'put', data })
}

// @Summary 删除实体
// @Router /eav/entity/:id [delete]
export const deleteEntity = (id) => {
  return service({ url: `/pmocker/eav/entity/${id}`, method: 'delete' })
}

// @Summary 获取实体详情
// @Router /eav/entity/:id [get]
export const getEntity = (id) => {
  return service({ url: `/pmocker/eav/entity/${id}`, method: 'get' })
}

// @Summary 列出实体
// @Router /eav/entities [get]
export const listEntities = (params) => {
  return service({ url: '/pmocker/eav/entities', method: 'get', params })
}
