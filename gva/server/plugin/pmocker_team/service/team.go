package team

import (
	"context"
	"fmt"

	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
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

// ratingDefaultScore 评级 → 默认评分（0-100），用于评分节点未手动打分时的兜底
var ratingDefaultScore = map[string]float64{
	"excellent":         90,
	"good":              80,
	"satisfactory":      70,
	"needs_improvement": 60,
	"unsatisfactory":    50,
}

// ScoreCalcHandler 工作流 auto 节点处理器：绩效评估自动评分
// handler 名：pmocker.team.score_calc
// 逻辑：若 score 未填，则按 rating 推断默认分回写；rating 也未填则保留原值不报错，流程照常推进。
func ScoreCalcHandler(ctx context.Context, entityID uint) error {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("get performance_review entity %d: %w", entityID, err)
	}
	if e.EntityType != EntityTypePerformanceReview {
		return fmt.Errorf("entity %d is %q, expect %s", entityID, e.EntityType, EntityTypePerformanceReview)
	}
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	// 已人工评分则不动（EAV 存整数值时读回 int64，小数为 float64）
	if score, ok := e.Attrs["score"]; ok {
		if v, ok := score.(float64); ok && v > 0 {
			return nil
		}
		if v, ok := score.(int64); ok && v > 0 {
			return nil
		}
		if v, ok := score.(int); ok && v > 0 {
			return nil
		}
	}
	// 按评级推断默认分
	if rating, ok := e.Attrs["rating"].(string); ok {
		if def, ok := ratingDefaultScore[rating]; ok {
			e.Attrs["score"] = def
			return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
		}
	}
	return nil
}
