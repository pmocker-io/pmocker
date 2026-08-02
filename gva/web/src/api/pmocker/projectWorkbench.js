import service from '@/utils/request'

export const getMyProjects = (params) => {
  return service({ url: '/pmocker/projectWorkbench/my', method: 'get', params })
}

export const getFocusedProjects = () => {
  return service({ url: '/pmocker/projectWorkbench/focused', method: 'get' })
}
