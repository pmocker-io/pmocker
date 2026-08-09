/**
 * 核心字段配置表
 *
 * 设计原则：
 * - L1 通用核心（_universal）：所有实体类型共有的标识/管理字段，自动合并
 * - L2 模块核心：按 entity_type 配置的关键业务字段
 * - 未列出的字段自动归入"扩展属性"区动态渲染
 *
 * 可配置能力：
 * - 修改此文件即可调整各模块的核心字段，无需改组件代码
 * - 支持新增 entity_type 配置
 * - 每个字段可选配置 colSpan（栅格宽度，默认自动：text/json=24，其余=12）
 */

// ── 人员字段集合 ──
// 这些字段在 DynamicForm 中自动渲染为"用户选择器"（el-select + 用户列表）
// 值存储为用户 ID（int），显示为用户昵称
export const userFieldKeys = new Set([
  'assignee',          // 问题/变更/需求 经办人
  'reviewer',          // 交付物 审阅人
  'reviewer_id',       // 绩效 评审人
  'owner_id',          // 任务/风险/里程碑/交付物 负责人
  'reported_by',       // 问题 报告人
  'requested_by',      // 变更/需求 提出人
  'verified_by',       // 问题 验证人
  'validated_by',      // 变更 验证人
  'implemented_by',    // 变更 实施人
  'impact_analyzed_by',// 变更 影响分析人
  'ccb_decision_maker',// 变更 CCB决策人
  'changed_by',        // 变更日志 变更人
  'member_id',         // 培训/绩效 团队成员
  'manager_id',        // EPS 项目经理
  'sponsor_id',        // EPS 发起人
  'raci_responsible',  // 范围 RACI-R
  'raci_accountable',  // 范围 RACI-A
  'reporting_to',      // 团队成员 汇报对象
  'stakeholder_id',    // 需求 干系人
  'assignee_id',       // 任务 实施人
  'checked_out_by',    // 交付物 检出人
  'created_by',        // 通用 创建人
])

/**
 * 判断字段是否为人员字段
 */
export function isUserField(fieldKey) {
  return userFieldKeys.has(fieldKey)
}

export const coreFieldsConfig = {
  // L1: 通用核心字段 - 自动应用到所有实体类型
  _universal: ['code', 'description'],

  // L2: 模块核心字段（不含通用核心）
  cost_item: [
    'budget_type', 'planned_value', 'earned_value', 'actual_cost', 'bac',
    'cv', 'sv', 'cpi', 'spi', 'eac', 'etc', 'vac', 'tcpi',
    'estimation_method', 'contingency_reserve', 'management_reserve',
    'variance_reason'
  ],
  task: [
    'start_date', 'end_date', 'duration', 'progress', 'assignee_id',
    'dependency_type', 'lead_lag', 'task_type',
    'is_critical_path', 'total_float', 'free_float',
    'baseline_start', 'baseline_finish', 'actual_start', 'actual_finish'
  ],
  milestone: [
    'due_date', 'is_critical', 'owner_id', 'deliverable_id',
    'baseline_date', 'actual_date'
  ],
  risk: [
    'category', 'probability', 'impact', 'risk_score', 'risk_level',
    'response_strategy', 'owner_id', 'due_date',
    'expected_monetary_value', 'impact_amount', 'contingency_reserve',
    'opportunity_strategy', 'analysis_type', 'response_cost'
  ],
  requirement: [
    'priority', 'source', 'category', 'assignee',
    'moscow_priority', 'requirement_type', 'verification_method',
    'story_points', 'agile_level',
    'requested_by', 'requested_date', 'stakeholder_id'
  ],
  scope_item: [
    'wbs_code', 'is_work_package', 'acceptance_status',
    'raci_responsible', 'raci_accountable',
    'assumptions', 'constraints', 'exclusions'
  ],
  issue: [
    'priority', 'severity', 'category', 'assignee', 'reported_by',
    'due_date', 'resolution_type', 'related_risk_id',
    'verified_by', 'reopen_count'
  ],
  eps_node: [
    'name', 'type', 'governance_type', 'lifecycle_phase',
    'priority', 'health_status', 'methodology',
    'manager_id', 'sponsor_id', 'start_date', 'end_date', 'budget'
  ],
  deliverable: [
    'name', 'category', 'version', 'owner_id',
    'planned_date', 'actual_date',
    'review_status', 'security_classification', 'defect_count'
  ],
  change_request: [
    'title', 'type', 'priority', 'change_nature',
    'requested_by', 'requested_date',
    'is_emergency', 'affected_baselines',
    'ccb_decision', 'ccb_decision_date', 'ccb_decision_maker'
  ],
  team_member: [
    'full_name', 'role', 'department', 'allocation_percent',
    'skill_level', 'start_date', 'end_date', 'status',
    'reporting_to', 'hourly_rate', 'billable', 'communication_mode'
  ],
  team_role: [
    'name', 'authority_level', 'raci_default',
    'min_experience_years', 'max_headcount', 'is_active'
  ],
  training_record: [
    'member_id', 'course_name', 'training_type',
    'start_date', 'end_date', 'duration_hours',
    'certification_obtained', 'effectiveness_score'
  ],
  performance_review: [
    'member_id', 'review_period', 'review_type',
    'rating', 'score', 'reviewer_id', 'review_date', 'status'
  ]
}

/**
 * 获取指定实体类型的核心字段列表（含通用核心）
 * @param {string} entityType 实体类型
 * @returns {string[]} 核心字段 field_key 数组
 */
export function getCoreFieldKeys(entityType) {
  const universal = coreFieldsConfig._universal || []
  const moduleFields = coreFieldsConfig[entityType] || []
  return [...universal, ...moduleFields]
}

/**
 * 判断字段是否为核心字段
 * @param {string} entityType 实体类型
 * @param {string} fieldKey 字段名
 * @returns {boolean}
 */
export function isCoreField(entityType, fieldKey) {
  return getCoreFieldKeys(entityType).includes(fieldKey)
}
