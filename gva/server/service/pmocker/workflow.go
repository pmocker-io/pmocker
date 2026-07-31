package pmocker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/pmocker-io/pmocker/pkg/pmocker/workflow"
)

// WorkflowService 实现 workflow.Engine 接口
type WorkflowService struct{}

// LoadDefinition 加载工作流定义到 DB
func (s *WorkflowService) LoadDefinition(ctx context.Context, def workflow.Definition) error {
	defJSON, _ := json.Marshal(def)
	return global.GVA_DB.WithContext(ctx).
		Where(pmocker.PMWorkflowDef{Code: def.Code}).
		Assign(pmocker.PMWorkflowDef{
			Name:           def.Name,
			EntityType:     def.EntityType,
			Trigger:        def.Trigger,
			DefinitionJSON: string(defJSON),
		}).FirstOrCreate(&pmocker.PMWorkflowDef{Code: def.Code}).Error
}

// Start 启动工作流实例
func (s *WorkflowService) Start(ctx context.Context, entityID uint, defCode string) (uint, error) {
	var def pmocker.PMWorkflowDef
	if err := global.GVA_DB.WithContext(ctx).Where("code = ?", defCode).First(&def).Error; err != nil {
		return 0, fmt.Errorf("workflow def %s not found: %w", defCode, err)
	}
	var wf workflow.Definition
	if err := json.Unmarshal([]byte(def.DefinitionJSON), &wf); err != nil {
		return 0, fmt.Errorf("invalid workflow definition: %w", err)
	}
	var startNode *workflow.Node
	for i := range wf.Nodes {
		if wf.Nodes[i].Type == workflow.NodeStart {
			startNode = &wf.Nodes[i]
			break
		}
	}
	if startNode == nil {
		return 0, fmt.Errorf("no start node in workflow %s", defCode)
	}
	inst := pmocker.PMWorkflowInstance{
		EntityID:     entityID,
		WorkflowCode: defCode,
		CurrentNode:  startNode.Code,
		Status:       "running",
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&inst).Error; err != nil {
		return 0, err
	}
	return inst.ID, nil
}

// Execute 执行转移
func (s *WorkflowService) Execute(ctx context.Context, instanceID uint, action string, userID uint) error {
	var inst pmocker.PMWorkflowInstance
	if err := global.GVA_DB.WithContext(ctx).First(&inst, instanceID).Error; err != nil {
		return err
	}
	if inst.Status != "running" {
		return fmt.Errorf("workflow instance %d is not running", instanceID)
	}
	var def pmocker.PMWorkflowDef
	if err := global.GVA_DB.WithContext(ctx).Where("code = ?", inst.WorkflowCode).First(&def).Error; err != nil {
		return err
	}
	var wf workflow.Definition
	if err := json.Unmarshal([]byte(def.DefinitionJSON), &wf); err != nil {
		return err
	}
	for _, t := range wf.Transitions {
		if t.From == inst.CurrentNode && (t.On == "" || t.On == action) {
			update := map[string]interface{}{"current_node": t.To}
			if t.Status != "" {
				update["status"] = t.Status
			}
			for _, n := range wf.Nodes {
				if n.Code == t.To && n.Type == workflow.NodeEnd {
					update["status"] = "completed"
					break
				}
			}
			return global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
				Where("id = ?", instanceID).Updates(update).Error
		}
	}
	return fmt.Errorf("no transition from %s on action %s", inst.CurrentNode, action)
}

// GetCurrentNode 获取当前节点
func (s *WorkflowService) GetCurrentNode(ctx context.Context, instanceID uint) (*workflow.Node, error) {
	var inst pmocker.PMWorkflowInstance
	if err := global.GVA_DB.WithContext(ctx).First(&inst, instanceID).Error; err != nil {
		return nil, err
	}
	var def pmocker.PMWorkflowDef
	if err := global.GVA_DB.WithContext(ctx).Where("code = ?", inst.WorkflowCode).First(&def).Error; err != nil {
		return nil, err
	}
	var wf workflow.Definition
	if err := json.Unmarshal([]byte(def.DefinitionJSON), &wf); err != nil {
		return nil, err
	}
	for i := range wf.Nodes {
		if wf.Nodes[i].Code == inst.CurrentNode {
			return &wf.Nodes[i], nil
		}
	}
	return nil, fmt.Errorf("node %s not found", inst.CurrentNode)
}

var _ workflow.Engine = (*WorkflowService)(nil)
