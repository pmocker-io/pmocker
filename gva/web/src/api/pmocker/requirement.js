import service from '@/utils/request'

// @Summary 创建需求
// @Router /requirement/create [post]
export const createRequirement = (data) => {
  return service({ url: '/requirement/create', method: 'post', data })
}

// @Summary 删除需求
// @Router /requirement/delete [delete]
export const deleteRequirement = (params) => {
  return service({ url: '/requirement/delete', method: 'delete', params })
}

// @Summary 更新需求
// @Router /requirement/update [put]
export const updateRequirement = (data) => {
  return service({ url: '/requirement/update', method: 'put', data })
}

// @Summary 查询需求详情
// @Router /requirement/find [get]
export const findRequirement = (params) => {
  return service({ url: '/requirement/find', method: 'get', params })
}

// @Summary 获取需求列表
// @Router /requirement/list [get]
export const getRequirementList = (params) => {
  return service({ url: '/requirement/list', method: 'get', params })
}

// @Summary 提交需求评审
// @Router /requirement/submitReview [post]
export const submitRequirementReview = (data) => {
  return service({ url: '/requirement/submitReview', method: 'post', data })
}

// @Summary 批准需求
// @Router /requirement/approve [post]
export const approveRequirement = (data) => {
  return service({ url: '/requirement/approve', method: 'post', data })
}

// @Summary 驳回需求
// @Router /requirement/reject [post]
export const rejectRequirement = (data) => {
  return service({ url: '/requirement/reject', method: 'post', data })
}

// @Summary 获取需求追踪矩阵
// @Router /requirement/traceMatrix [get]
export const getRequirementTraceMatrix = (params) => {
  return service({ url: '/requirement/traceMatrix', method: 'get', params })
}
