package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/global"

// PMWBSNode WBS 树结构
type PMWBSNode struct {
	global.GVA_MODEL
	ProjectID     uint   `json:"projectId" gorm:"index;not null;comment:项目ID"`
	ParentID      *uint  `json:"parentId" gorm:"index;comment:父节点ID"`
	Path          string `json:"path" gorm:"size:255;index;comment:物化路径如1.3.2"`
	Level         int    `json:"level" gorm:"comment:层级"`
	EntityID      uint   `json:"entityId" gorm:"index;comment:关联实体ID"`
	DeliverableID *uint  `json:"deliverableId" gorm:"index;comment:关联交付物ID"`
}

func (PMWBSNode) TableName() string { return "pm_wbs_nodes" }

// PMEPSTree EPS 树
type PMEPSTree struct {
	global.GVA_MODEL
	ParentID    *uint  `json:"parentId" gorm:"index;comment:父节点ID"`
	Path        string `json:"path" gorm:"size:255;index;comment:物化路径"`
	ProjectType string `json:"projectType" gorm:"size:32;comment:类型"`
	Name        string `json:"name" gorm:"size:128;comment:名称"`
}

func (PMEPSTree) TableName() string { return "pm_eps_tree" }

// PMTaskLink 任务依赖关系
type PMTaskLink struct {
	global.GVA_MODEL
	SrcTaskID uint   `json:"srcTaskId" gorm:"index;comment:前置任务ID"`
	DstTaskID uint   `json:"dstTaskId" gorm:"index;comment:后置任务ID"`
	LinkType  string `json:"linkType" gorm:"size:2;comment:FS/SS/FF/SF"`
	Lag       int    `json:"lag" gorm:"default:0;comment:滞后量"`
}

func (PMTaskLink) TableName() string { return "pm_task_links" }

// PMBaseline 基线快照
type PMBaseline struct {
	global.GVA_MODEL
	ProjectID    uint   `json:"projectId" gorm:"index;comment:项目ID"`
	Type         string `json:"type" gorm:"size:16;comment:scope/schedule/cost"`
	SnapshotJSON string `json:"snapshotJson" gorm:"type:text;comment:快照JSON"`
	ChangeReqID  *uint  `json:"changeReqId" gorm:"index;comment:变更请求ID"`
}

func (PMBaseline) TableName() string { return "pm_baselines" }

// PMChangeLog 变更日志
type PMChangeLog struct {
	global.GVA_MODEL
	EntityID         uint   `json:"entityId" gorm:"index;comment:实体ID"`
	FieldKey         string `json:"fieldKey" gorm:"size:64;comment:字段键"`
	OldValue         string `json:"oldValue" gorm:"type:text;comment:旧值"`
	NewValue         string `json:"newValue" gorm:"type:text;comment:新值"`
	ChangedBy        uint   `json:"changedBy" gorm:"comment:变更人"`
	ChangeRequestID  *uint  `json:"changeRequestId" gorm:"index;comment:变更请求ID"`
}

func (PMChangeLog) TableName() string { return "pm_change_logs" }

// PMDeliverableFile 交付物文件
type PMDeliverableFile struct {
	global.GVA_MODEL
	DeliverableID uint   `json:"deliverableId" gorm:"index;comment:交付物实体ID"`
	FilePath      string `json:"filePath" gorm:"size:512;comment:文件路径"`
	Checksum      string `json:"checksum" gorm:"size:64;comment:文件校验和"`
	Size          int64  `json:"size" gorm:"comment:文件大小"`
	Mime          string `json:"mime" gorm:"size:64;comment:MIME类型"`
	CheckedOutBy  *uint  `json:"checkedOutBy" gorm:"index;comment:检出人ID"`
	VersionMajor  int    `json:"versionMajor" gorm:"default:1;comment:主版本号"`
	VersionMinor  int    `json:"versionMinor" gorm:"default:0;comment:次版本号"`
}

func (PMDeliverableFile) TableName() string { return "pm_deliverable_files" }

// PMWorkflowDef 工作流定义
type PMWorkflowDef struct {
	global.GVA_MODEL
	Code           string `json:"code" gorm:"uniqueIndex;size:64;comment:工作流编码"`
	Name           string `json:"name" gorm:"size:128;comment:名称"`
	EntityType     string `json:"entityType" gorm:"size:64;comment:实体类型"`
	Trigger        string `json:"trigger" gorm:"size:32;comment:触发条件"`
	DefinitionJSON string `json:"definitionJson" gorm:"type:text;comment:定义JSON"`
}

func (PMWorkflowDef) TableName() string { return "pm_workflow_defs" }

// PMWorkflowInstance 工作流实例
type PMWorkflowInstance struct {
	global.GVA_MODEL
	EntityID     uint   `json:"entityId" gorm:"index;comment:实体ID"`
	WorkflowCode string `json:"workflowCode" gorm:"size:64;comment:工作流编码"`
	CurrentNode  string `json:"currentNode" gorm:"size:64;comment:当前节点"`
	Status       string `json:"status" gorm:"size:16;comment:running/completed/rejected"`
}

func (PMWorkflowInstance) TableName() string { return "pm_workflow_instances" }
