import service from '@/utils/request'
export const calcVariance = (params) => service({ url: '/pmocker/variance/calc', method: 'get', params })
export const getAlerts = (params) => service({ url: '/pmocker/variance/alerts', method: 'get', params })
