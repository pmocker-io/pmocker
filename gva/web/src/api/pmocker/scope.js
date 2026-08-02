import service from '@/utils/request'

// @Summary 创建范围项
// @Router /scope/createItem [post]
export const createScopeItem = (data) => {
  return service({ url: '/pmocker/scope/createItem', method: 'post', data })
}

// @Summary 构建 WBS
// @Router /scope/buildWBS [post]
export const buildScopeWBS = (data) => {
  return service({ url: '/pmocker/scope/buildWBS', method: 'post', data })
}

// @Summary 创建范围基线
// @Router /scope/baseline [post]
export const createScopeBaseline = (data) => {
  return service({ url: '/pmocker/scope/baseline', method: 'post', data })
}

// @Summary 获取范围项列表
// @Router /scope/listItems [get]
export const getScopeItems = (params) => {
  return service({ url: '/pmocker/scope/listItems', method: 'get', params })
}

// @Summary 获取 WBS 树
// @Router /scope/getWBS [get]
export const getScopeWBS = (params) => {
  return service({ url: '/pmocker/scope/getWBS', method: 'get', params })
}

// @Summary 更新范围项
// @Router /scope/update [put]
export const updateScopeItem = (data) => {
  return service({ url: '/pmocker/scope/update', method: 'put', data })
}

// @Summary 删除范围项
// @Router /scope/delete [delete]
export const deleteScopeItem = (params) => {
  return service({ url: '/pmocker/scope/delete', method: 'delete', params })
}

// @Summary 查询范围项详情
// @Router /scope/find [get]
export const findScopeItem = (params) => {
  return service({ url: '/pmocker/scope/find', method: 'get', params })
}
