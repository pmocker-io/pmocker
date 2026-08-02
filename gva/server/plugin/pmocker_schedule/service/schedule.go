package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	deliverablesvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_deliverable/service"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

// ScheduleService 进度管理 service
type ScheduleService struct{}

func (s *ScheduleService) CreateTask(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "task", Title: title, Status: "planned", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *ScheduleService) ListTasks(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "task", offset, limit)
}

// GetTask 获取任务
func (s *ScheduleService) GetTask(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "task" {
		return nil, fmt.Errorf("entity %d is not a task", id)
	}
	return e, nil
}

// UpdateTask 更新任务
func (s *ScheduleService) UpdateTask(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "task" {
		return fmt.Errorf("not a task")
	}
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, e); err != nil {
		return err
	}
	// 兼容前端通过 updateTask 将状态改为完成：触发关联交付物自动检入（联动失败仅告警，不阻断）
	if isCompletedStatus(e.Status) {
		if latest, err := s.GetTask(ctx, e.ID); err == nil {
			s.checkInLinkedDeliverable(ctx, latest, 0)
		}
	}
	return nil
}

// Transition 任务状态流转。当流转到 completed/done 时，自动检入关联交付物（联动失败仅告警，不阻断任务流转）。
func (s *ScheduleService) Transition(ctx context.Context, taskID uint, status string, operatorID uint) error {
	e, err := s.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	e.Status = status
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, *e); err != nil {
		return err
	}
	if isCompletedStatus(status) {
		s.checkInLinkedDeliverable(ctx, e, operatorID)
	}
	return nil
}

// checkInLinkedDeliverable 任务完成时自动检入关联交付物。
// 读取 task.Attrs.deliverable_id，非 0 则调用 deliverable service CheckIn；
// 检入操作人优先取任务负责人 assignee_id，其次取 operatorID；
// 任何失败仅记录日志，不向上抛错以免阻塞任务状态流转。
func (s *ScheduleService) checkInLinkedDeliverable(ctx context.Context, task *eavtypes.Entity, operatorID uint) {
	if task == nil || task.Attrs == nil {
		return
	}
	deliverableID := attrToUint(task.Attrs["deliverable_id"])
	if deliverableID == 0 {
		return
	}
	userID := attrToUint(task.Attrs["assignee_id"])
	if userID == 0 {
		userID = operatorID
	}
	const versionNote = "任务完成自动检入"
	if err := deliverablesvc.DeliverableService.CheckIn(ctx, deliverableID, userID, versionNote, "", operatorID); err != nil {
		log.Printf("[schedule-linkage] 任务 %d 完成自动检入交付物 %d 失败: %v（不阻断任务流转）", task.ID, deliverableID, err)
	}
}

// isCompletedStatus 判断是否为完成态（兼容 completed/done 两种写法）
func isCompletedStatus(status string) bool {
	return status == "completed" || status == "done"
}

// attrToUint 将 EAV attr 值转换为 uint（支持 uint/int/int64/float64/string）
func attrToUint(v interface{}) uint {
	switch n := v.(type) {
	case uint:
		return n
	case int:
		return uint(n)
	case int64:
		return uint(n)
	case float64:
		return uint(n)
	case string:
		var x uint
		fmt.Sscanf(n, "%d", &x)
		return x
	}
	return 0
}

// DeleteTask 删除任务
func (s *ScheduleService) DeleteTask(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *ScheduleService) CreateMilestone(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "milestone", Title: title, Status: "planned", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *ScheduleService) ListMilestones(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "milestone", offset, limit)
}

// ComputeCPM 从项目任务计算关键路径
func (s *ScheduleService) ComputeCPM(ctx context.Context, projectID uint) ([]CPMResult, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "task", 0, 10000)
	if err != nil {
		return nil, err
	}
	tasks := make([]Task, 0, len(entities))
	for _, e := range entities {
		t := Task{ID: e.ID, Code: e.Title}
		if d, ok := e.Attrs["duration"].(float64); ok {
			t.Duration = int(d)
		}
		if preds, ok := e.Attrs["predecessors"]; ok {
			if b, err := json.Marshal(preds); err == nil {
				_ = json.Unmarshal(b, &t.Predecessors)
			}
		}
		tasks = append(tasks, t)
	}
	return CPM(tasks)
}

// SetBaseline 设置进度基线
func (s *ScheduleService) SetBaseline(ctx context.Context, projectID uint, snapshotJSON string) (uint, error) {
	bl := pmocker.PMBaseline{ProjectID: projectID, Type: "schedule", SnapshotJSON: snapshotJSON}
	if err := global.GVA_DB.WithContext(ctx).Create(&bl).Error; err != nil {
		return 0, err
	}
	return bl.ID, nil
}
