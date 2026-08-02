import service from '@/utils/request'

// @Summary 创建交付物
// @Router /deliverable/create [post]
export const createDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/create', method: 'post', data })
}

// @Summary 删除交付物
// @Router /deliverable/delete [delete]
export const deleteDeliverable = (params) => {
  return service({ url: '/pmocker/deliverable/delete', method: 'delete', params })
}

// @Summary 更新交付物
// @Router /deliverable/update [put]
export const updateDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/update', method: 'put', data })
}

// @Summary 查询交付物详情
// @Router /deliverable/find [get]
export const findDeliverable = (params) => {
  return service({ url: '/pmocker/deliverable/find', method: 'get', params })
}

// @Summary 获取交付物列表
// @Router /deliverable/list [get]
export const getDeliverableList = (params) => {
  return service({ url: '/pmocker/deliverable/list', method: 'get', params })
}

// @Summary 提交交付物评审
// @Router /deliverable/submitReview [post]
export const submitDeliverableReview = (data) => {
  return service({ url: '/pmocker/deliverable/submitReview', method: 'post', data })
}

// @Summary 接收交付物
// @Router /deliverable/accept [post]
export const acceptDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/accept', method: 'post', data })
}

// @Summary 驳回交付物
// @Router /deliverable/reject [post]
export const rejectDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/reject', method: 'post', data })
}

// @Summary 创建新版本
// @Router /deliverable/createVersion [post]
export const createDeliverableVersion = (data) => {
  return service({ url: '/pmocker/deliverable/createVersion', method: 'post', data })
}

// @Summary 创建基线
// @Router /deliverable/baseline [post]
export const createDeliverableBaseline = (data) => {
  return service({ url: '/pmocker/deliverable/baseline', method: 'post', data })
}

// @Summary 检出交付物（排他锁定）
// @Router /deliverable/checkOut [post]
export const checkOutDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/checkOut', method: 'post', data })
}

// @Summary 检入交付物（解锁并可选记录版本）
// @Router /deliverable/checkIn [post]
export const checkInDeliverable = (data) => {
  return service({ url: '/pmocker/deliverable/checkIn', method: 'post', data })
}

// @Summary 获取版本列表
// @Router /deliverable/listVersions [get]
export const getDeliverableVersions = (params) => {
  return service({ url: '/pmocker/deliverable/listVersions', method: 'get', params })
}

// @Summary 获取追溯报告
// @Router /deliverable/traceReport [get]
export const getDeliverableTraceReport = (params) => {
  return service({ url: '/pmocker/deliverable/traceReport', method: 'get', params })
}

// @Summary 获取交付物统计
// @Router /deliverable/stats [get]
export const getDeliverableStats = (params) => {
  return service({ url: '/pmocker/deliverable/stats', method: 'get', params })
}
