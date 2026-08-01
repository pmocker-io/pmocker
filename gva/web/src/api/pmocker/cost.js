import service from '@/utils/request'

// @Summary 创建成本项
// @Router /cost/createItem [post]
export const createCostItem = (data) => {
  return service({ url: '/cost/createItem', method: 'post', data })
}

// @Summary 创建成本基线
// @Router /cost/baseline [post]
export const createCostBaseline = (data) => {
  return service({ url: '/cost/baseline', method: 'post', data })
}

// @Summary 获取成本项列表
// @Router /cost/listItems [get]
export const getCostItems = (params) => {
  return service({ url: '/cost/listItems', method: 'get', params })
}

// @Summary 挣值分析
// @Router /cost/evm [post]
export const analyzeCostEVM = (data) => {
  return service({ url: '/cost/evm', method: 'post', data })
}
