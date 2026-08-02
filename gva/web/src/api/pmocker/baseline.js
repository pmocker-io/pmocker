import service from '@/utils/request'
export const createBaseline = (data) => service({ url: '/pmocker/baseline/create', method: 'post', data })
export const listBaselines = (params) => service({ url: '/pmocker/baseline/list', method: 'get', params })
export const compareBaseline = (params) => service({ url: '/pmocker/baseline/compare', method: 'get', params })
