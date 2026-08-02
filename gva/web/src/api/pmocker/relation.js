import service from '@/utils/request'

export const createRelation = (data) => {
  return service({ url: '/pmocker/relation/create', method: 'post', data })
}
export const deleteRelation = (params) => {
  return service({ url: '/pmocker/relation/delete', method: 'delete', params })
}
export const listRelations = (params) => {
  return service({ url: '/pmocker/relation/list', method: 'get', params })
}
