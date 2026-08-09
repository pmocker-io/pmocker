/**
 * 各模块实体状态选项定义
 *
 * 用于编辑对话框中的状态下拉选择器和列表中的状态标签渲染
 */

export const statusOptions = {
  // 问题管理
  issue: [
    { value: 'open', label: '待处理', type: 'info' },
    { value: 'assigned', label: '已分配', type: 'warning' },
    { value: 'in_progress', label: '处理中', type: 'warning' },
    { value: 'resolved', label: '已解决', type: 'success' },
    { value: 'verified', label: '已验证', type: 'success' },
    { value: 'closed', label: '已关闭', type: '' },
    { value: 'reopened', label: '已重开', type: 'danger' },
  ],
  // 风险管理
  risk: [
    { value: 'identified', label: '已识别', type: 'info' },
    { value: 'analyzed', label: '已分析', type: 'warning' },
    { value: 'responded', label: '已响应', type: 'primary' },
    { value: 'closed', label: '已关闭', type: 'success' },
    { value: 'opportunity', label: '机会', type: 'success' },
  ],
  // 进度管理-任务
  task: [
    { value: 'planned', label: '计划中', type: 'info' },
    { value: 'in_progress', label: '进行中', type: 'warning' },
    { value: 'completed', label: '已完成', type: 'success' },
    { value: 'on_hold', label: '暂停', type: 'danger' },
    { value: 'cancelled', label: '已取消', type: '' },
  ],
  // 进度管理-里程碑
  milestone: [
    { value: 'planned', label: '计划中', type: 'info' },
    { value: 'reached', label: '已达成', type: 'success' },
    { value: 'missed', label: '已逾期', type: 'danger' },
  ],
  // 变更管理
  change_request: [
    { value: 'submitted', label: '已提交', type: 'info' },
    { value: 'analyzing', label: '分析中', type: 'warning' },
    { value: 'ccb_review', label: 'CCB审批', type: 'warning' },
    { value: 'approved', label: '已批准', type: 'success' },
    { value: 'rejected', label: '已驳回', type: 'danger' },
    { value: 'implementing', label: '实施中', type: 'primary' },
    { value: 'verifying', label: '验证中', type: 'warning' },
    { value: 'closed', label: '已关闭', type: '' },
  ],
  // 交付物管理
  deliverable: [
    { value: 'draft', label: '草稿', type: 'info' },
    { value: 'reviewing', label: '评审中', type: 'warning' },
    { value: 'accepted', label: '已接收', type: 'success' },
    { value: 'rejected', label: '已驳回', type: 'danger' },
    { value: 'baselined', label: '已基线', type: 'success' },
    { value: 'checked_out', label: '已检出', type: 'warning' },
  ],
  // 需求管理
  requirement: [
    { value: 'draft', label: '草稿', type: 'info' },
    { value: 'reviewing', label: '评审中', type: 'warning' },
    { value: 'approved', label: '已批准', type: 'success' },
    { value: 'rejected', label: '已驳回', type: 'danger' },
    { value: 'implemented', label: '已实现', type: 'success' },
  ],
  // 成本管理
  cost_item: [
    { value: 'planned', label: '计划中', type: 'info' },
    { value: 'committed', label: '已承诺', type: 'warning' },
    { value: 'actual', label: '实际', type: 'primary' },
    { value: 'closed', label: '已关闭', type: '' },
  ],
  // 团队管理-成员
  team_member: [
    { value: 'candidate', label: '候选', type: 'info' },
    { value: 'active', label: '在职', type: 'success' },
    { value: 'on_leave', label: '休假', type: 'warning' },
    { value: 'terminated', label: '离职', type: 'danger' },
  ],
  // 团队管理-角色
  team_role: [
    { value: 'draft', label: '草稿', type: 'info' },
    { value: 'active', label: '启用', type: 'success' },
    { value: 'inactive', label: '停用', type: 'danger' },
  ],
  // 团队管理-培训
  training_record: [
    { value: 'planned', label: '计划中', type: 'info' },
    { value: 'in_progress', label: '进行中', type: 'warning' },
    { value: 'completed', label: '已完成', type: 'success' },
    { value: 'cancelled', label: '已取消', type: 'danger' },
  ],
  // 团队管理-绩效
  performance_review: [
    { value: 'draft', label: '草稿', type: 'info' },
    { value: 'self_review', label: '自评中', type: 'warning' },
    { value: 'reviewing', label: '评审中', type: 'warning' },
    { value: 'completed', label: '已完成', type: 'success' },
  ],
  // 范围管理
  scope_item: [
    { value: 'draft', label: '草稿', type: 'info' },
    { value: 'active', label: '活跃', type: 'success' },
    { value: 'baselined', label: '已基线', type: 'primary' },
    { value: 'closed', label: '已关闭', type: '' },
  ],
  // EPS节点
  eps_node: [
    { value: 'planning', label: '规划中', type: 'info' },
    { value: 'active', label: '活跃', type: 'success' },
    { value: 'on_hold', label: '暂停', type: 'warning' },
    { value: 'closed', label: '已关闭', type: '' },
    { value: 'archived', label: '已归档', type: 'info' },
  ],
}

/**
 * 获取指定实体类型的状态选项列表
 */
export function getStatusOptions(entityType) {
  return statusOptions[entityType] || []
}

/**
 * 获取状态的显示标签
 */
export function getStatusLabel(entityType, status) {
  const opts = statusOptions[entityType] || []
  const found = opts.find(s => s.value === status)
  return found ? found.label : status
}

/**
 * 获取状态的标签类型（用于 el-tag type 属性）
 */
export function getStatusType(entityType, status) {
  const opts = statusOptions[entityType] || []
  const found = opts.find(s => s.value === status)
  return found ? found.type : 'info'
}
