/**
 * 各模块状态流转配置
 *
 * 定义每个实体类型的状态分组及每个状态下可执行的操作
 * 用于列表页面的按状态分组展示和批量流转操作
 *
 * 格式：
 * transitions[entityType] = [
 *   {
 *     status: 'draft',           // 状态值
 *     label: '草稿',              // 状态显示名
 *     tagType: 'info',           // el-tag 类型
 *     actions: [                 // 该状态下可执行的操作
 *       { label: '提交评审', target: 'reviewing', apiFn: 'submitReview', type: 'primary' },
 *     ]
 *   }
 * ]
 */

export const transitions = {
  // 问题管理
  issue: [
    { status: 'open', label: '待处理', tagType: 'info', actions: [
      { label: '分配', target: 'assigned', apiFn: 'assignIssue', type: 'primary' },
      { label: '关闭', target: 'closed', apiFn: 'closeIssue', type: 'success' },
    ]},
    { status: 'assigned', label: '已分配', tagType: 'warning', actions: [
      { label: '开始处理', target: 'in_progress', apiFn: 'updateIssue', type: 'primary' },
      { label: '关闭', target: 'closed', apiFn: 'closeIssue', type: 'success' },
    ]},
    { status: 'in_progress', label: '处理中', tagType: 'warning', actions: [
      { label: '解决', target: 'resolved', apiFn: 'resolveIssue', type: 'success' },
      { label: '关闭', target: 'closed', apiFn: 'closeIssue', type: 'success' },
    ]},
    { status: 'resolved', label: '已解决', tagType: 'success', actions: [
      { label: '验证', target: 'verified', apiFn: 'updateIssue', type: 'primary' },
      { label: '重开', target: 'reopened', apiFn: 'reopenIssue', type: 'danger' },
    ]},
    { status: 'verified', label: '已验证', tagType: 'success', actions: [
      { label: '关闭', target: 'closed', apiFn: 'closeIssue', type: 'success' },
      { label: '重开', target: 'reopened', apiFn: 'reopenIssue', type: 'danger' },
    ]},
    { status: 'closed', label: '已关闭', tagType: '', actions: [
      { label: '重开', target: 'reopened', apiFn: 'reopenIssue', type: 'warning' },
    ]},
    { status: 'reopened', label: '已重开', tagType: 'danger', actions: [
      { label: '解决', target: 'resolved', apiFn: 'resolveIssue', type: 'success' },
      { label: '关闭', target: 'closed', apiFn: 'closeIssue', type: 'success' },
    ]},
  ],

  // 任务管理
  task: [
    { status: 'planned', label: '计划中', tagType: 'info', actions: [
      { label: '开始', target: 'in_progress', apiFn: 'transitionTask', type: 'primary' },
      { label: '取消', target: 'cancelled', apiFn: 'transitionTask', type: 'danger' },
    ]},
    { status: 'in_progress', label: '进行中', tagType: 'warning', actions: [
      { label: '完成', target: 'completed', apiFn: 'transitionTask', type: 'success' },
      { label: '暂停', target: 'on_hold', apiFn: 'transitionTask', type: 'warning' },
    ]},
    { status: 'on_hold', label: '暂停', tagType: 'danger', actions: [
      { label: '恢复', target: 'in_progress', apiFn: 'transitionTask', type: 'primary' },
      { label: '取消', target: 'cancelled', apiFn: 'transitionTask', type: 'danger' },
    ]},
    { status: 'completed', label: '已完成', tagType: 'success', actions: [] },
    { status: 'cancelled', label: '已取消', tagType: '', actions: [] },
  ],

  // 变更管理
  change_request: [
    // draft: 后端 CreateChangeRequest 强制设为 draft，需经"提交"进入 submitted
    { status: 'draft', label: '草稿', tagType: 'info', actions: [
      { label: '提交', target: 'submitted', apiFn: 'updateChange', type: 'primary' },
    ]},
    { status: 'submitted', label: '已提交', tagType: 'info', actions: [
      { label: '影响分析', target: 'analyzing', apiFn: 'analyzeChange', type: 'primary' },
    ]},
    { status: 'analyzing', label: '分析中', tagType: 'warning', actions: [
      { label: '提交CCB', target: 'ccb_review', apiFn: 'ccbReviewChange', type: 'primary' },
      { label: '驳回', target: 'rejected', apiFn: 'rejectChange', type: 'danger' },
    ]},
    { status: 'ccb_review', label: 'CCB审批', tagType: 'warning', actions: [
      { label: '批准', target: 'approved', apiFn: 'approveChange', type: 'success' },
      { label: '驳回', target: 'rejected', apiFn: 'rejectChange', type: 'danger' },
    ]},
    { status: 'approved', label: '已批准', tagType: 'success', actions: [
      { label: '开始实施', target: 'implementing', apiFn: 'implementChange', type: 'primary' },
    ]},
    { status: 'rejected', label: '已驳回', tagType: 'danger', actions: [] },
    { status: 'implementing', label: '实施中', tagType: 'primary', actions: [
      // 状态机 verify_pass 后 status=verified（非 verifying）
      { label: '验证', target: 'verified', apiFn: 'verifyChange', type: 'warning' },
    ]},
    { status: 'verified', label: '已验证', tagType: 'success', actions: [
      { label: '关闭', target: 'closed', apiFn: 'closeChange', type: 'success' },
    ]},
    { status: 'closed', label: '已关闭', tagType: '', actions: [] },
    { status: 'cancelled', label: '已取消', tagType: '', actions: [] },
  ],

  // 风险管理 - 状态对齐 schema: identified, assessed, responding, closed, realized
  risk: [
    { status: 'identified', label: '已识别', tagType: 'info', actions: [
      { label: '评估', target: 'assessed', apiFn: 'updateRisk', type: 'primary' },
      { label: '关闭', target: 'closed', apiFn: 'updateRisk', type: 'success' },
    ]},
    { status: 'assessed', label: '已评估', tagType: 'warning', actions: [
      { label: '响应', target: 'responding', apiFn: 'updateRisk', type: 'primary' },
      { label: '关闭', target: 'closed', apiFn: 'updateRisk', type: 'success' },
    ]},
    { status: 'responding', label: '响应中', tagType: 'primary', actions: [
      { label: '关闭', target: 'closed', apiFn: 'updateRisk', type: 'success' },
      { label: '已发生', target: 'realized', apiFn: 'updateRisk', type: 'danger' },
    ]},
    { status: 'closed', label: '已关闭', tagType: 'success', actions: [] },
    { status: 'realized', label: '已发生', tagType: 'danger', actions: [
      { label: '关闭', target: 'closed', apiFn: 'updateRisk', type: 'success' },
    ]},
  ],

  // 需求管理 - 状态对齐 schema: draft, reviewing, approved, rejected, fulfilled
  requirement: [
    { status: 'draft', label: '草稿', tagType: 'info', actions: [
      { label: '提交评审', target: 'reviewing', apiFn: 'submitRequirementReview', type: 'primary' },
    ]},
    { status: 'reviewing', label: '评审中', tagType: 'warning', actions: [
      { label: '批准', target: 'approved', apiFn: 'approveRequirement', type: 'success' },
      { label: '驳回', target: 'rejected', apiFn: 'rejectRequirement', type: 'danger' },
    ]},
    { status: 'approved', label: '已批准', tagType: 'success', actions: [
      { label: '标记实现', target: 'fulfilled', apiFn: 'updateRequirement', type: 'primary' },
    ]},
    { status: 'rejected', label: '已驳回', tagType: 'danger', actions: [
      { label: '重新提交', target: 'reviewing', apiFn: 'submitRequirementReview', type: 'warning' },
    ]},
    { status: 'fulfilled', label: '已实现', tagType: 'success', actions: [] },
  ],

  // 交付物管理 - 状态对齐 schema: draft, submitted, reviewing, accepted, rejected, baselined, obsolete
  deliverable: [
    { status: 'draft', label: '草稿', tagType: 'info', actions: [
      { label: '提交', target: 'submitted', apiFn: 'submitDeliverableReview', type: 'primary' },
    ]},
    { status: 'submitted', label: '已提交', tagType: 'info', actions: [
      { label: '评审', target: 'reviewing', apiFn: 'updateDeliverable', type: 'primary' },
    ]},
    { status: 'reviewing', label: '评审中', tagType: 'warning', actions: [
      { label: '接收', target: 'accepted', apiFn: 'acceptDeliverable', type: 'success' },
      { label: '驳回', target: 'rejected', apiFn: 'rejectDeliverable', type: 'danger' },
    ]},
    { status: 'accepted', label: '已接收', tagType: 'success', actions: [
      { label: '基线', target: 'baselined', apiFn: 'updateDeliverable', type: 'primary' },
    ]},
    { status: 'rejected', label: '已驳回', tagType: 'danger', actions: [
      { label: '重新提交', target: 'submitted', apiFn: 'submitDeliverableReview', type: 'warning' },
    ]},
    { status: 'baselined', label: '已基线', tagType: 'success', actions: [
      { label: '废弃', target: 'obsolete', apiFn: 'updateDeliverable', type: 'danger' },
    ]},
    { status: 'obsolete', label: '已废弃', tagType: '', actions: [] },
  ],
}

/**
 * 获取指定实体类型的状态流转配置
 */
export function getTransitions(entityType) {
  return transitions[entityType] || []
}

/**
 * 获取状态分组的列表数据
 * @param {Array} list - 原始列表数据
 * @param {string} entityType - 实体类型
 * @returns {Array} 分组后的数据 [{ status, label, tagType, actions, items }]
 */
export function groupByStatus(list, entityType) {
  const config = getTransitions(entityType)
  return config.map(group => ({
    ...group,
    items: list.filter(item => item.status === group.status),
  })).filter(group => group.items.length > 0)
}
