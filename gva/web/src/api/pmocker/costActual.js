import service from '@/utils/request'

export const createCostActual = (data) => service({ url: '/pmocker/costActual/create', method: 'post', data })
export const updateCostActual = (data) => service({ url: '/pmocker/costActual/update', method: 'put', data })
export const deleteCostActual = (params) => service({ url: '/pmocker/costActual/delete', method: 'delete', params })
export const findCostActual = (params) => service({ url: '/pmocker/costActual/find', method: 'get', params })
export const listCostActuals = (params) => service({ url: '/pmocker/costActual/list', method: 'get', params })
export const confirmCostActual = (params) => service({ url: '/pmocker/costActual/confirm', method: 'post', params })
