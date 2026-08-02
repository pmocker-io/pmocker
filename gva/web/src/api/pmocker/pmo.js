import service from '@/utils/request'

export const getPMODashboard = () => {
  return service({ url: '/pmocker/pmo/dashboard', method: 'get' })
}
