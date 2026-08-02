package cost

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var CostService = new(Service)

func (s *Service) CreateItem(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "cost_item", Title: title, Status: "planned", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) ListItems(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "cost_item", offset, limit)
}

// GetItem 获取成本项
func (s *Service) GetItem(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "cost_item" {
		return nil, fmt.Errorf("entity %d is not a cost_item", id)
	}
	return e, nil
}

// UpdateItem 更新成本项
func (s *Service) UpdateItem(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "cost_item" {
		return fmt.Errorf("not a cost_item")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

// DeleteItem 删除成本项
func (s *Service) DeleteItem(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

// ComputeEVM 计算项目级 EVM（汇总所有成本项）
func (s *Service) ComputeEVM(ctx context.Context, projectID uint) (EVMResult, error) {
	items, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "cost_item", 0, 10000)
	if err != nil {
		return EVMResult{}, err
	}
	var in EVMInput
	for _, it := range items {
		in.BAC += getFloat(it.Attrs["planned_value"])
		in.PV += getFloat(it.Attrs["planned_value"]) * getProgressRate(it.Attrs) // PV = 总预算 * 计划完成%
		in.EV += getFloat(it.Attrs["planned_value"]) * getActualRate(it.Attrs)   // EV = 总预算 * 实际完成%
		in.AC += getFloat(it.Attrs["actual_cost"])
	}
	// 简化：PV = BAC（假设项目按计划）
	in.PV = in.BAC
	return EVM(in), nil
}

func (s *Service) SetBaseline(ctx context.Context, projectID uint, snapshotJSON string) (uint, error) {
	bl := pmocker.PMBaseline{ProjectID: projectID, Type: "cost", SnapshotJSON: snapshotJSON}
	if err := global.GVA_DB.WithContext(ctx).Create(&bl).Error; err != nil {
		return 0, err
	}
	return bl.ID, nil
}

func getFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func getProgressRate(attrs map[string]interface{}) float64 {
	return 1.0 // 简化：所有项均按计划完成
}

func getActualRate(attrs map[string]interface{}) float64 {
	if ac, ok := attrs["actual_cost"]; ok {
		if pv, ok := attrs["planned_value"]; ok {
			if getFloat(pv) > 0 {
				return getFloat(ac) / getFloat(pv)
			}
		}
	}
	return 0
}
