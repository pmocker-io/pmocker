import service from '@/utils/request'

// @Summary 创建变更请求
// @Router /change/create [post]
export const createChange = (data) => {
  return service({ url: '/pmocker/change/create', method: 'post', data })
}

// @Summary 删除变更
// @Router /change/delete [delete]
export const deleteChange = (params) => {
  return service({ url: '/pmocker/change/delete', method: 'delete', params })
}

// @Summary 更新变更
// @Router /change/update [put]
export const updateChange = (data) => {
  return service({ url: '/pmocker/change/update', method: 'put', data })
}

// @Summary 查询变更详情
// @Router /change/find [get]
export const findChange = (params) => {
  return service({ url: '/pmocker/change/find', method: 'get', params })
}

// @Summary 获取变更列表
// @Router /change/list [get]
export const getChangeList = (params) => {
  return service({ url: '/pmocker/change/list', method: 'get', params })
}

// @Summary 变更影响分析
// @Router /change/analyze [post]
export const analyzeChange = (data) => {
  return service({ url: '/pmocker/change/analyze', method: 'post', data })
}

// @Summary CCB 评审
// @Router /change/ccbReview [post]
export const ccbReviewChange = (data) => {
  return service({ url: '/pmocker/change/ccbReview', method: 'post', data })
}

// @Summary 批准变更
// @Router /change/approve [post]
export const approveChange = (data) => {
  return service({ url: '/pmocker/change/approve', method: 'post', data })
}

// @Summary 驳回变更
// @Router /change/reject [post]
export const rejectChange = (data) => {
  return service({ url: '/pmocker/change/reject', method: 'post', data })
}

// @Summary 实施变更
// @Router /change/implement [post]
export const implementChange = (data) => {
  return service({ url: '/pmocker/change/implement', method: 'post', data })
}

// @Summary 验证变更
// @Router /change/verify [post]
export const verifyChange = (data) => {
  return service({ url: '/pmocker/change/verify', method: 'post', data })
}

// @Summary 关闭变更
// @Router /change/close [post]
export const closeChange = (data) => {
  return service({ url: '/pmocker/change/close', method: 'post', data })
}

// @Summary 获取变更日志
// @Router /change/listLogs [get]
export const getChangeLogs = (params) => {
  return service({ url: '/pmocker/change/listLogs', method: 'get', params })
}

// @Summary 获取影响报告
// @Router /change/impactReport [get]
export const getChangeImpactReport = (params) => {
  return service({ url: '/pmocker/change/impactReport', method: 'get', params })
}

// @Summary 获取 CCB 统计
// @Router /change/ccbStats [get]
export const getChangeCCBStats = (params) => {
  return service({ url: '/pmocker/change/ccbStats', method: 'get', params })
}
