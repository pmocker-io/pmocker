import service from '@/utils/request'
export const getProgress = (params) => service({ url: '/pmocker/progress/get', method: 'get', params })
