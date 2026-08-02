import service from '@/utils/request'

export const createTimeEntry = (data) => service({ url: '/pmocker/timeEntry/create', method: 'post', data })
export const updateTimeEntry = (data) => service({ url: '/pmocker/timeEntry/update', method: 'put', data })
export const deleteTimeEntry = (params) => service({ url: '/pmocker/timeEntry/delete', method: 'delete', params })
export const findTimeEntry = (params) => service({ url: '/pmocker/timeEntry/find', method: 'get', params })
export const listTimeEntries = (params) => service({ url: '/pmocker/timeEntry/list', method: 'get', params })
export const submitTimeEntry = (params) => service({ url: '/pmocker/timeEntry/submit', method: 'post', params })
export const approveTimeEntry = (params) => service({ url: '/pmocker/timeEntry/approve', method: 'post', params })
export const rejectTimeEntry = (params) => service({ url: '/pmocker/timeEntry/reject', method: 'post', params })
export const getUtilization = (params) => service({ url: '/pmocker/timeEntry/utilization', method: 'get', params })
