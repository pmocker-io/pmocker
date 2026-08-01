import service from '@/utils/request'

// @Summary 创建范围项
// @Router /scope/createItem [post]
export const createScopeItem = (data) => {
  return service({ url: '/scope/createItem', method: 'post', data })
}

// @Summary 构建 WBS
// @Router /scope/buildWBS [post]
export const buildScopeWBS = (data) => {
  return service({ url: '/scope/buildWBS', method: 'post', data })
}

// @Summary 创建范围基线
// @Router /scope/baseline [post]
export const createScopeBaseline = (data) => {
  return service({ url: '/scope/baseline', method: 'post', data })
}

// @Summary 获取范围项列表
// @Router /scope/listItems [get]
export const getScopeItems = (params) => {
  return service({ url: '/scope/listItems', method: 'get', params })
}

// @Summary 获取 WBS 树
// @Router /scope/getWBS [get]
export const getScopeWBS = (params) => {
  return service({ url: '/scope/getWBS', method: 'get', params })
}
