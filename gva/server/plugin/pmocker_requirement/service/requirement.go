package requirement

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

// Service 需求管理 service，封装 EAV 操作 + 业务算法
type Service struct{}

// RequirementService 单例
var RequirementService = new(Service)

// CreateRequirement 创建需求
func (s *Service) CreateRequirement(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmocker.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  projectID,
		EntityType: "requirement",
		Title:      title,
		Status:     "draft",
		CreatedBy:  creatorID,
		Attrs:      attrs,
	})
}

// GetRequirement 获取需求
func (s *Service) GetRequirement(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmocker.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "requirement" {
		return nil, fmt.Errorf("entity %d is not a requirement", id)
	}
	return e, nil
}

// ListRequirements 列出项目的需求
func (s *Service) ListRequirements(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmocker.ServiceGroupApp.ListEntities(ctx, projectID, "requirement", offset, limit)
}

// UpdateRequirement 更新需求
func (s *Service) UpdateRequirement(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "requirement" {
		return fmt.Errorf("not a requirement")
	}
	return pmocker.ServiceGroupApp.UpdateEntity(ctx, e)
}

// DeleteRequirement 删除需求
func (s *Service) DeleteRequirement(ctx context.Context, id uint) error {
	return pmocker.ServiceGroupApp.DeleteEntity(ctx, id)
}

// TraceMatrix 需求追踪矩阵（需求→交付物 关系）
func (s *Service) TraceMatrix(ctx context.Context, projectID uint) ([]TraceRow, error) {
	reqs, _, err := pmocker.ServiceGroupApp.ListEntities(ctx, projectID, "requirement", 0, 1000)
	if err != nil {
		return nil, err
	}
	rows := make([]TraceRow, 0, len(reqs))
	for _, r := range reqs {
		rels, err := pmocker.ServiceGroupApp.ListRelations(ctx, r.ID)
		if err != nil {
			continue
		}
		deliverableIDs := make([]uint, 0, len(rels))
		for _, rel := range rels {
			if rel.RelationType == "traced_to" {
				deliverableIDs = append(deliverableIDs, rel.DstID)
			}
		}
		rows = append(rows, TraceRow{
			RequirementID:    r.ID,
			RequirementTitle: r.Title,
			Status:           r.Status,
			DeliverableIDs:   deliverableIDs,
		})
	}
	return rows, nil
}

// TraceRow 追踪矩阵行
type TraceRow struct {
	RequirementID    uint   `json:"requirementId"`
	RequirementTitle string `json:"requirementTitle"`
	Status           string `json:"status"`
	DeliverableIDs   []uint `json:"deliverableIds"`
}

// SubmitReview 启动需求评审工作流
func (s *Service) SubmitReview(ctx context.Context, requirementID, userID uint) (uint, error) {
	return pmocker.ServiceGroupApp.Start(ctx, requirementID, "requirement_review")
}

// Approve 通过工作流转移
func (s *Service) Approve(ctx context.Context, instanceID, userID uint) error {
	return pmocker.ServiceGroupApp.Execute(ctx, instanceID, "approve", userID)
}

// Reject 驳回工作流转移
func (s *Service) Reject(ctx context.Context, instanceID, userID uint) error {
	return pmocker.ServiceGroupApp.Execute(ctx, instanceID, "reject", userID)
}
