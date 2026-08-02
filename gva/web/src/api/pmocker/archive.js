import service from '@/utils/request'

export const archiveProject = (data) => {
  return service({ url: '/pmocker/project/archive', method: 'post', data })
}

export const getCloseReport = (params) => {
  return service({ url: '/pmocker/project/closeReport', method: 'get', params })
}
