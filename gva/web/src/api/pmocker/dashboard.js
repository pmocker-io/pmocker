import service from '@/utils/request'

export const getDashboard = (params) => {
  return service({ url: '/pmocker/dashboard/get', method: 'get', params })
}

export const generateSnapshot = (data) => {
  return service({ url: '/pmocker/report/snapshot', method: 'post', data })
}

export const listSnapshots = (params) => {
  return service({ url: '/pmocker/report/list', method: 'get', params })
}
