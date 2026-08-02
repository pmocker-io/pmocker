package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMTimeEntry 工时登记表
type PMTimeEntry struct {
	global.GVA_MODEL
	ProjectID   uint    `json:"projectId" gorm:"index;not null;comment:项目ID"`
	TaskID      uint    `json:"taskId" gorm:"index;not null;comment:任务实体ID"`
	MemberID    uint    `json:"memberId" gorm:"index;not null;comment:团队成员实体ID"`
	UserID      uint    `json:"userId" gorm:"index;not null;comment:sys_users用户ID"`
	Date        string  `json:"date" gorm:"size:10;index;not null;comment:日期YYYY-MM-DD"`
	Hours       float64 `json:"hours" gorm:"type:decimal(8,2);not null;comment:工时"`
	HourlyRate  float64 `json:"hourlyRate" gorm:"type:decimal(12,2);comment:时薪快照"`
	Cost        float64 `json:"cost" gorm:"type:decimal(14,2);comment:成本(hours*rate)"`
	Description string  `json:"description" gorm:"size:500;comment:工作描述"`
	Status      string  `json:"status" gorm:"size:16;index;default:draft;comment:draft/submitted/approved/rejected"`
	ApproverID  *uint   `json:"approverId" gorm:"index;comment:审批人ID"`
	ApprovedAt  *string `json:"approvedAt" gorm:"size:19;comment:审批时间"`
}

func (PMTimeEntry) TableName() string { return "pm_time_entries" }

// PMCostActual 实际成本执行表
type PMCostActual struct {
	global.GVA_MODEL
	ProjectID   uint    `json:"projectId" gorm:"index;not null;comment:项目ID"`
	TaskID      *uint   `json:"taskId" gorm:"index;comment:关联任务ID"`
	CostItemID  *uint   `json:"costItemId" gorm:"index;comment:关联成本项ID"`
	CostType    string  `json:"costType" gorm:"size:32;index;comment:labor/material/equipment/travel/other"`
	Amount      float64 `json:"amount" gorm:"type:decimal(14,2);not null;comment:金额"`
	Date        string  `json:"date" gorm:"size:10;index;not null;comment:发生日期"`
	Source      string  `json:"source" gorm:"size:32;comment:manual/time_entry/invoice"`
	RefID       *uint   `json:"refId" gorm:"index;comment:来源记录ID(如time_entry.id)"`
	Description string  `json:"description" gorm:"size:500;comment:描述"`
	Status      string  `json:"status" gorm:"size:16;index;default:pending;comment:pending/confirmed"`
}

func (PMCostActual) TableName() string { return "pm_cost_actuals" }

// PMApprovalRecord 审批签审记录表
type PMApprovalRecord struct {
	global.GVA_MODEL
	ProjectID       uint   `json:"projectId" gorm:"index;not null;comment:项目ID"`
	EntityID        uint   `json:"entityId" gorm:"index;not null;comment:被审批实体ID"`
	EntityType      string `json:"entityType" gorm:"size:64;index;comment:实体类型"`
	WorkflowInstID  *uint  `json:"workflowInstId" gorm:"index;comment:工作流实例ID"`
	NodeName        string `json:"nodeName" gorm:"size:64;comment:审批节点名"`
	ApproverID      uint   `json:"approverId" gorm:"not null;comment:审批人sys_users ID"`
	ApproverName    string `json:"approverName" gorm:"size:64;comment:审批人姓名快照"`
	Action          string `json:"action" gorm:"size:16;not null;comment:approve/reject/withdraw"`
	Comment         string `json:"comment" gorm:"type:text;comment:审批意见"`
	Signature       string `json:"signature" gorm:"size:128;comment:电子签名hash"`
	ActedAt         string `json:"actedAt" gorm:"size:19;comment:审批时间"`
}

func (PMApprovalRecord) TableName() string { return "pm_approval_records" }

// PMReportSnapshot 报告快照表（里程碑存档）
type PMReportSnapshot struct {
	global.GVA_MODEL
	ProjectID    uint   `json:"projectId" gorm:"index;not null;comment:项目ID"`
	ReportType   string `json:"reportType" gorm:"size:32;index;not null;comment:dashboard/pmo/close"`
	Period       string `json:"period" gorm:"size:10;index;comment:报告周期如2026-06或close"`
	SnapshotJSON string `json:"snapshotJson" gorm:"type:text;comment:快照JSON"`
	GeneratedBy  uint   `json:"generatedBy" gorm:"comment:生成人"`
}

func (PMReportSnapshot) TableName() string { return "pm_report_snapshots" }
