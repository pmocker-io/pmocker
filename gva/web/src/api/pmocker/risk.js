import service from '@/utils/request'

// @Summary 创建风险
// @Router /risk/create [post]
export const createRisk = (data) => {
  return service({ url: '/pmocker/risk/create', method: 'post', data })
}

// @Summary 删除风险
// @Router /risk/delete [delete]
export const deleteRisk = (params) => {
  return service({ url: '/pmocker/risk/delete', method: 'delete', params })
}

// @Summary 更新风险
// @Router /risk/update [put]
export const updateRisk = (data) => {
  return service({ url: '/pmocker/risk/update', method: 'put', data })
}

// @Summary 评估风险
// @Router /risk/assess [post]
export const assessRisk = (data) => {
  return service({ url: '/pmocker/risk/assess', method: 'post', data })
}

// @Summary 查询风险详情
// @Router /risk/find [get]
export const findRisk = (params) => {
  return service({ url: '/pmocker/risk/find', method: 'get', params })
}

// @Summary 获取风险列表
// @Router /risk/list [get]
export const getRiskList = (params) => {
  return service({ url: '/pmocker/risk/list', method: 'get', params })
}

// @Summary 获取风险矩阵
// @Router /risk/matrix [get]
export const getRiskMatrix = (params) => {
  return service({ url: '/pmocker/risk/matrix', method: 'get', params })
}
