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

// @Summary 更新任务
export const updateTask = (data) => {
  return service({ url: '/pmocker/schedule/updateTask', method: 'post', data })
}

// @Summary 查询任务详情
export const findTask = (params) => {
  return service({ url: '/pmocker/schedule/findTask', method: 'get', params })
}

// @Summary 删除任务
export const deleteTask = (params) => {
  return service({ url: '/pmocker/schedule/deleteTask', method: 'delete', params })
}

// @Summary 任务状态流转
export const transitionTask = (data) => {
  return service({ url: '/pmocker/schedule/transitionTask', method: 'post', data })
}
