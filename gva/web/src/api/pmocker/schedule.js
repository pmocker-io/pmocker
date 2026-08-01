import service from '@/utils/request'

// @Summary 创建任务
// @Router /schedule/createTask [post]
export const createScheduleTask = (data) => {
  return service({ url: '/pmocker/schedule/createTask', method: 'post', data })
}

// @Summary 创建里程碑
// @Router /schedule/createMilestone [post]
export const createScheduleMilestone = (data) => {
  return service({ url: '/pmocker/schedule/createMilestone', method: 'post', data })
}

// @Summary 创建进度基线
// @Router /schedule/baseline [post]
export const createScheduleBaseline = (data) => {
  return service({ url: '/pmocker/schedule/baseline', method: 'post', data })
}

// @Summary 获取任务列表
// @Router /schedule/listTasks [get]
export const getScheduleTasks = (params) => {
  return service({ url: '/pmocker/schedule/listTasks', method: 'get', params })
}

// @Summary 获取里程碑列表
// @Router /schedule/listMilestones [get]
export const getScheduleMilestones = (params) => {
  return service({ url: '/pmocker/schedule/listMilestones', method: 'get', params })
}

// @Summary CPM 关键路径分析
// @Router /schedule/cpm [post]
export const analyzeScheduleCPM = (data) => {
  return service({ url: '/pmocker/schedule/cpm', method: 'post', data })
}
