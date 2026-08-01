import service from '@/utils/request'

// @Summary 创建问题
// @Router /issue/create [post]
export const createIssue = (data) => {
  return service({ url: '/issue/create', method: 'post', data })
}

// @Summary 删除问题
// @Router /issue/delete [delete]
export const deleteIssue = (params) => {
  return service({ url: '/issue/delete', method: 'delete', params })
}

// @Summary 更新问题
// @Router /issue/update [put]
export const updateIssue = (data) => {
  return service({ url: '/issue/update', method: 'put', data })
}

// @Summary 查询问题详情
// @Router /issue/find [get]
export const findIssue = (params) => {
  return service({ url: '/issue/find', method: 'get', params })
}

// @Summary 获取问题列表
// @Router /issue/list [get]
export const getIssueList = (params) => {
  return service({ url: '/issue/list', method: 'get', params })
}

// @Summary 分配问题
// @Router /issue/assign [post]
export const assignIssue = (data) => {
  return service({ url: '/issue/assign', method: 'post', data })
}

// @Summary 解决问题
// @Router /issue/resolve [post]
export const resolveIssue = (data) => {
  return service({ url: '/issue/resolve', method: 'post', data })
}

// @Summary 关闭问题
// @Router /issue/close [post]
export const closeIssue = (data) => {
  return service({ url: '/issue/close', method: 'post', data })
}

// @Summary 重新打开问题
// @Router /issue/reopen [post]
export const reopenIssue = (data) => {
  return service({ url: '/issue/reopen', method: 'post', data })
}

// @Summary 获取看板数据
// @Router /issue/board [get]
export const getIssueBoard = (params) => {
  return service({ url: '/issue/board', method: 'get', params })
}

// @Summary 获取问题统计
// @Router /issue/stats [get]
export const getIssueStats = (params) => {
  return service({ url: '/issue/stats', method: 'get', params })
}
