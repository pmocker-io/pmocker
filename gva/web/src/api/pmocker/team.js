import service from '@/utils/request'

// ---- 团队成员 ----
export const createMember = (data) => service({ url: '/pmocker/team/member/create', method: 'post', data })
export const deleteMember = (params) => service({ url: '/pmocker/team/member/delete', method: 'delete', params })
export const updateMember = (data) => service({ url: '/pmocker/team/member/update', method: 'put', data })
export const findMember = (params) => service({ url: '/pmocker/team/member/find', method: 'get', params })
export const listMember = (params) => service({ url: '/pmocker/team/member/list', method: 'get', params })
export const transitionMember = (data) => service({ url: '/pmocker/team/member/transition', method: 'post', data })

// ---- 角色定义 ----
export const createRole = (data) => service({ url: '/pmocker/team/role/create', method: 'post', data })
export const deleteRole = (params) => service({ url: '/pmocker/team/role/delete', method: 'delete', params })
export const updateRole = (data) => service({ url: '/pmocker/team/role/update', method: 'put', data })
export const findRole = (params) => service({ url: '/pmocker/team/role/find', method: 'get', params })
export const listRole = (params) => service({ url: '/pmocker/team/role/list', method: 'get', params })
export const transitionRole = (data) => service({ url: '/pmocker/team/role/transition', method: 'post', data })

// ---- 培训记录 ----
export const createTraining = (data) => service({ url: '/pmocker/team/training/create', method: 'post', data })
export const deleteTraining = (params) => service({ url: '/pmocker/team/training/delete', method: 'delete', params })
export const updateTraining = (data) => service({ url: '/pmocker/team/training/update', method: 'put', data })
export const findTraining = (params) => service({ url: '/pmocker/team/training/find', method: 'get', params })
export const listTraining = (params) => service({ url: '/pmocker/team/training/list', method: 'get', params })
export const transitionTraining = (data) => service({ url: '/pmocker/team/training/transition', method: 'post', data })

// ---- 绩效评估 ----
export const createPerformance = (data) => service({ url: '/pmocker/team/performance/create', method: 'post', data })
export const deletePerformance = (params) => service({ url: '/pmocker/team/performance/delete', method: 'delete', params })
export const updatePerformance = (data) => service({ url: '/pmocker/team/performance/update', method: 'put', data })
export const findPerformance = (params) => service({ url: '/pmocker/team/performance/find', method: 'get', params })
export const listPerformance = (params) => service({ url: '/pmocker/team/performance/list', method: 'get', params })
export const transitionPerformance = (data) => service({ url: '/pmocker/team/performance/transition', method: 'post', data })
