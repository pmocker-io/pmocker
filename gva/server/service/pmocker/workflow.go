package pmocker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/pmocker-io/pmocker/pkg/pmocker/workflow"
)

// NodeHook 节点事件钩子
type NodeHook interface {
	OnEnter(ctx context.Context, entityID uint, nodeName string) error
	OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error
}

type hookCtxKey string

const hookUserIDKey hookCtxKey = "pmocker.hook.userID"

func userIDFromCtx(ctx context.Context) uint {
	if v, ok := ctx.Value(hookUserIDKey).(uint); ok {
		return v
	}
	return 0
}

// WorkflowService 实现 workflow.Engine 接口
type WorkflowService struct {
	mu       sync.RWMutex
	handlers map[string]workflow.AutoHandler
	hooks    map[string]NodeHook
}

// RegisterAutoHandler 注册 NodeAuto 节点处理器（幂等覆盖）。
func (s *WorkflowService) RegisterAutoHandler(name string, h workflow.AutoHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.handlers == nil {
		s.handlers = make(map[string]workflow.AutoHandler)
	}
	s.handlers[name] = h
}

// RegisterNodeHook 注册节点事件钩子（按 workflowCode.nodeName 索引，幂等覆盖）。
func (s *WorkflowService) RegisterNodeHook(workflowCode, nodeName string, hook NodeHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.hooks == nil {
		s.hooks = make(map[string]NodeHook)
	}
	s.hooks[workflowCode+"."+nodeName] = hook
}

func (s *WorkflowService) fireOnLeave(ctx context.Context, workflowCode, nodeName, action string, entityID uint) error {
	s.mu.RLock()
	h, ok := s.hooks[workflowCode+"."+nodeName]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return h.OnLeave(ctx, entityID, nodeName, action)
}

func (s *WorkflowService) fireOnEnter(ctx context.Context, workflowCode, nodeName string, entityID uint) error {
	s.mu.RLock()
	h, ok := s.hooks[workflowCode+"."+nodeName]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return h.OnEnter(ctx, entityID, nodeName)
}

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
	instID := inst.ID
	// 若 start 之后立即进入 auto 节点（通过 action 空转移），尝试自动链式推进
	if startNode.Type == workflow.NodeStart {
		_ = s.tryAdvanceAuto(ctx, instID, &wf, 0)
	}
	return instID, nil
}

// Execute 执行转移；若目标节点为 NodeAuto，则自动调度 Handler 并继续链式推进直到非 auto 节点。
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
	// 1) 定位转移
	var target *workflow.Transition
	for i := range wf.Transitions {
		t := &wf.Transitions[i]
		if t.From == inst.CurrentNode && (t.On == "" || t.On == action) {
			target = t
			break
		}
	}
	if target == nil {
		return fmt.Errorf("no transition from %s on action %s", inst.CurrentNode, action)
	}
	// 2) 应用状态转移
	update := map[string]interface{}{"current_node": target.To}
	if target.Status != "" {
		update["status"] = target.Status
	}
	targetNode := findNode(&wf, target.To)
	if targetNode != nil && targetNode.Type == workflow.NodeEnd {
		update["status"] = "completed"
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// 3) 触发 NodeHook
	hookCtx := context.WithValue(ctx, hookUserIDKey, userID)
	fromNode := inst.CurrentNode
	if err := s.fireOnLeave(hookCtx, inst.WorkflowCode, fromNode, action, inst.EntityID); err != nil {
		return fmt.Errorf("onLeave hook failed on node %s: %w", fromNode, err)
	}
	if err := s.fireOnEnter(hookCtx, inst.WorkflowCode, target.To, inst.EntityID); err != nil {
		return fmt.Errorf("onEnter hook failed on node %s: %w", target.To, err)
	}
	// 4) 若目标节点是 auto，调度 handler 并继续链式推进
	if targetNode != nil && targetNode.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(hookCtx, instanceID, &wf, 0)
	}
	return nil
}

// tryAdvanceAuto 从当前实例节点出发：若当前节点为 NodeAuto 则执行 handler → 走首个 on=="" 或默认 on 转移 → 循环直到非 auto 节点。
// depth 用于防止循环配置导致的死循环（上限 16）。
func (s *WorkflowService) tryAdvanceAuto(ctx context.Context, instanceID uint, wf *workflow.Definition, depth int) error {
	if depth > 16 {
		return fmt.Errorf("auto node chain exceeded max depth 16 for instance %d", instanceID)
	}
	var inst pmocker.PMWorkflowInstance
	if err := global.GVA_DB.WithContext(ctx).First(&inst, instanceID).Error; err != nil {
		return err
	}
	if inst.Status != "running" {
		return nil
	}
	node := findNode(wf, inst.CurrentNode)
	if node == nil || node.Type != workflow.NodeAuto {
		return nil
	}
	// A) 调度 handler（若配置）
	if node.Handler != "" {
		s.mu.RLock()
		h, ok := s.handlers[node.Handler]
		s.mu.RUnlock()
		if !ok {
			return fmt.Errorf("auto handler %q not registered (node=%s instance=%d)", node.Handler, node.Code, instanceID)
		}
		if err := h(ctx, inst.EntityID); err != nil {
			return fmt.Errorf("auto handler %q failed: %w", node.Handler, err)
		}
	}
	// B) 寻找出边：优先 on=="" 空转移；否则取该节点的第一条出边
	var autoTrans *workflow.Transition
	var firstTrans *workflow.Transition
	for i := range wf.Transitions {
		t := &wf.Transitions[i]
		if t.From != node.Code {
			continue
		}
		if firstTrans == nil {
			firstTrans = t
		}
		if t.On == "" {
			autoTrans = t
			break
		}
	}
	if autoTrans == nil {
		// 空转移不存在时，使用第一条出边的 On 作为隐式 action（例如 "done" / "start_work"）
		autoTrans = firstTrans
	}
	if autoTrans == nil {
		return fmt.Errorf("no outgoing transition from auto node %s (instance=%d)", node.Code, instanceID)
	}
	// C) 应用下一次转移
	update := map[string]interface{}{"current_node": autoTrans.To}
	if autoTrans.Status != "" {
		update["status"] = autoTrans.Status
	}
	next := findNode(wf, autoTrans.To)
	if next != nil && next.Type == workflow.NodeEnd {
		update["status"] = "completed"
	}
	if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWorkflowInstance{}).
		Where("id = ?", instanceID).Updates(update).Error; err != nil {
		return err
	}
	// D) 触发 NodeHook
	leaveAction := autoTrans.On
	if leaveAction == "" {
		leaveAction = "done"
	}
	if err := s.fireOnLeave(ctx, inst.WorkflowCode, node.Code, leaveAction, inst.EntityID); err != nil {
		return fmt.Errorf("onLeave hook failed on auto node %s: %w", node.Code, err)
	}
	if next != nil {
		if err := s.fireOnEnter(ctx, inst.WorkflowCode, next.Code, inst.EntityID); err != nil {
			return fmt.Errorf("onEnter hook failed on node %s: %w", next.Code, err)
		}
	}
	// E) 若下一个节点仍是 auto，递归推进
	if next != nil && next.Type == workflow.NodeAuto {
		return s.tryAdvanceAuto(ctx, instanceID, wf, depth+1)
	}
	return nil
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

func findNode(wf *workflow.Definition, code string) *workflow.Node {
	for i := range wf.Nodes {
		if wf.Nodes[i].Code == code {
			return &wf.Nodes[i]
		}
	}
	return nil
}

var _ workflow.Engine = (*WorkflowService)(nil)
