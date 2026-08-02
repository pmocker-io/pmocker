import service from '@/utils/request'

export const listChangeLogs = (params) => {
  return service({ url: '/pmocker/changeLog/list', method: 'get', params })
}
