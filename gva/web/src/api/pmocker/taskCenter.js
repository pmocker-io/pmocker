import service from '@/utils/request'

export const getMyTasks = (params) => {
  return service({ url: '/pmocker/taskCenter/my', method: 'get', params })
}

export const getFocusedTasks = () => {
  return service({ url: '/pmocker/taskCenter/focused', method: 'get' })
}

export const getTaskStats = () => {
  return service({ url: '/pmocker/taskCenter/stats', method: 'get' })
}
