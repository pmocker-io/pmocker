package schedule

import (
	"context"
	"encoding/json"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
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
