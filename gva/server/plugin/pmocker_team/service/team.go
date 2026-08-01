package team

import (
	"context"
	"fmt"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var TeamService = new(Service)

// 4 个实体类型
const (
	EntityTypeTeamMember        = "team_member"
	EntityTypeTeamRole          = "team_role"
	EntityTypeTrainingRecord    = "training_record"
	EntityTypePerformanceReview = "performance_review"
)

// initialStatus 每个实体类型的初始状态
var initialStatus = map[string]string{
	EntityTypeTeamMember:        "candidate",
	EntityTypeTeamRole:          "active",
	EntityTypeTrainingRecord:    "planned",
	EntityTypePerformanceReview: "draft",
}

// validTypes 已注册的实体类型集合
var validTypes = map[string]bool{
	EntityTypeTeamMember:        true,
	EntityTypeTeamRole:          true,
	EntityTypeTrainingRecord:    true,
	EntityTypePerformanceReview: true,
}

// Create 创建团队管理实体
func (s *Service) Create(ctx context.Context, entityType string, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	if !validTypes[entityType] {
		return 0, fmt.Errorf("unsupported entity type: %s", entityType)
	}
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: entityType, Title: title, Status: initialStatus[entityType], CreatedBy: creatorID, Attrs: attrs,
	})
}

// Get 获取团队管理实体
func (s *Service) Get(ctx context.Context, entityType string, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != entityType {
		return nil, fmt.Errorf("entity %d is not a %s", id, entityType)
	}
	return e, nil
}

// List 列出团队管理实体
func (s *Service) List(ctx context.Context, entityType string, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, entityType, offset, limit)
}

// Update 更新团队管理实体
func (s *Service) Update(ctx context.Context, entityType string, e eavtypes.Entity) error {
	if e.EntityType != entityType {
		return fmt.Errorf("not a %s", entityType)
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

// Delete 删除团队管理实体
func (s *Service) Delete(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

// Transition 状态流转
func (s *Service) Transition(ctx context.Context, entityType string, id uint, status string) error {
	e, err := s.Get(ctx, entityType, id)
	if err != nil {
		return err
	}
	e.Status = status
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}
