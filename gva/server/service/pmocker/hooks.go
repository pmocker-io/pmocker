package pmocker

import (
	"context"
	"encoding/json"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// ScheduleBaselineHook 计划审批通过 → 生成计划基线
type ScheduleBaselineHook struct{}

func (h *ScheduleBaselineHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}
func (h *ScheduleBaselineHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" {
		return nil
	}
	_, err := (&BaselineService{}).CreateBaseline(entityID, "schedule", nil, userIDFromCtx(ctx))
	return err
}

// CostBaselineHook 成本审批通过 → 生成成本基线
type CostBaselineHook struct{}

func (h *CostBaselineHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}
func (h *CostBaselineHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" {
		return nil
	}
	_, err := (&BaselineService{}).CreateBaseline(entityID, "cost", nil, userIDFromCtx(ctx))
	return err
}

// ChangeApplyHook 变更批准 → 应用变更到目标实体 + 记录 change_logs
type ChangeApplyHook struct{}

func (h *ChangeApplyHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}
func (h *ChangeApplyHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "approve" && action != "close" {
		return nil
	}
	var change pmocker.PMEntity
	if err := global.GVA_DB.First(&change, entityID).Error; err != nil {
		return err
	}
	targetID := readAttrRef(entityID, "target_entity_id")
	if targetID == 0 {
		targetID = readAttrRef(entityID, "change_target_ref")
	}
	if targetID == 0 {
		return nil
	}
	fieldsJSON := readAttrString(entityID, "change_fields")
	if fieldsJSON == "" {
		return nil
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return err
	}
	userID := userIDFromCtx(ctx)
	changeReqID := entityID
	for field, newVal := range fields {
		oldVal := readAttrString(targetID, field)
		if oldVal == newVal {
			continue
		}
		if err := writeAttrString(targetID, field, newVal); err != nil {
			return err
		}
		if err := (&ChangeLogService{}).RecordChangeLog(targetID, field, oldVal, newVal, userID, &changeReqID); err != nil {
			return err
		}
	}
	return nil
}
