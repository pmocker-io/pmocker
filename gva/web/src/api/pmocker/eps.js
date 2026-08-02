import service from '@/utils/request'

// @Summary 创建 EPS 节点
// @Router /eps/createNode [post]
export const createEPSNode = (data) => {
  return service({ url: '/pmocker/eps/createNode', method: 'post', data })
}

// @Summary 更新 EPS 节点
// @Router /eps/updateNode [put]
export const updateEPSNode = (data) => {
  return service({ url: '/pmocker/eps/updateNode', method: 'put', data })
}

// @Summary 删除 EPS 节点
// @Router /eps/deleteNode [delete]
export const deleteEPSNode = (params) => {
  return service({ url: '/pmocker/eps/deleteNode', method: 'delete', params })
}

// @Summary 移动 EPS 节点
// @Router /eps/moveNode [post]
export const moveEPSNode = (data) => {
  return service({ url: '/pmocker/eps/moveNode', method: 'post', data })
}

// @Summary 添加成员
// @Router /eps/addMember [post]
export const addEPSMember = (data) => {
  return service({ url: '/pmocker/eps/addMember', method: 'post', data })
}

// @Summary 移除成员
// @Router /eps/removeMember [delete]
export const removeEPSMember = (params) => {
  return service({ url: '/pmocker/eps/removeMember', method: 'delete', params })
}

// @Summary 获取 EPS 节点列表
// @Router /eps/listNodes [get]
export const getEPSNodes = (params) => {
  return service({ url: '/pmocker/eps/listNodes', method: 'get', params })
}

// @Summary 获取成员列表
// @Router /eps/listMembers [get]
export const getEPSMembers = (params) => {
  return service({ url: '/pmocker/eps/listMembers', method: 'get', params })
}

// @Summary 查询 EPS 节点详情
// @Router /eps/find [get]
export const findEPSNode = (params) => {
  return service({ url: '/pmocker/eps/find', method: 'get', params })
}

// @Summary 获取 EPS 树结构（用于项目选择器/仪表盘）
// @Router /eps/tree [get]
export const getEPSTree = (params) => {
  return service({ url: '/pmocker/eps/tree', method: 'get', params })
}
