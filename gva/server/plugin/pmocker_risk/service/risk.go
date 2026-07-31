package risk

import (
	"context"
	"fmt"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var RiskService = new(Service)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// CalcRiskLevel 根据概率(P)和影响(I)计算风险等级
// P/I 均为 1-5，乘积 1-25
// 1-3: low, 4-9: medium, 10-15: high, 16-25: critical
func CalcRiskLevel(probability, impact int) (RiskLevel, int) {
	score := probability * impact
	switch {
	case score <= 3:
		return RiskLow, score
	case score <= 9:
		return RiskMedium, score
	case score <= 15:
		return RiskHigh, score
	default:
		return RiskCritical, score
	}
}

// MatrixCell 风险矩阵单元格
type MatrixCell struct {
	Probability int       `json:"probability"`
	Impact      int       `json:"impact"`
	Level       RiskLevel `json:"level"`
	Score       int       `json:"score"`
	Count       int       `json:"count"`
}

// RiskMatrix 5x5 风险矩阵
func RiskMatrix(risks []RiskInput) [5][5]MatrixCell {
	var matrix [5][5]MatrixCell
	for p := 1; p <= 5; p++ {
		for i := 1; i <= 5; i++ {
			level, score := CalcRiskLevel(p, i)
			matrix[p-1][i-1] = MatrixCell{
				Probability: p,
				Impact:      i,
				Level:       level,
				Score:       score,
				Count:       0,
			}
		}
	}
	for _, r := range risks {
		if r.Probability >= 1 && r.Probability <= 5 && r.Impact >= 1 && r.Impact <= 5 {
			matrix[r.Probability-1][r.Impact-1].Count++
		}
	}
	return matrix
}

// RiskInput 风险输入
type RiskInput struct {
	ID          uint `json:"id"`
	Probability int  `json:"probability"`
	Impact      int  `json:"impact"`
}

// CreateRisk 创建风险
func (s *Service) CreateRisk(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	// 自动计算风险等级
	if p, ok := attrs["probability"]; ok {
		if i, ok := attrs["impact"]; ok {
			prob := toInt(p)
			imp := toInt(i)
			level, score := CalcRiskLevel(prob, imp)
			attrs["risk_score"] = score
			attrs["risk_level"] = string(level)
		}
	}
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "risk", Title: title, Status: "identified", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) GetRisk(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "risk" {
		return nil, fmt.Errorf("entity %d is not a risk", id)
	}
	return e, nil
}

func (s *Service) ListRisks(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "risk", offset, limit)
}

func (s *Service) UpdateRisk(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "risk" {
		return fmt.Errorf("not a risk")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

func (s *Service) DeleteRisk(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

// AssessRisk 评估风险（重新计算风险分值和等级）
func (s *Service) AssessRisk(ctx context.Context, id uint, probability, impact int) error {
	e, err := s.GetRisk(ctx, id)
	if err != nil {
		return err
	}
	level, score := CalcRiskLevel(probability, impact)
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["probability"] = probability
	e.Attrs["impact"] = impact
	e.Attrs["risk_score"] = score
	e.Attrs["risk_level"] = string(level)
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

// ProjectMatrix 生成项目风险矩阵
func (s *Service) ProjectMatrix(ctx context.Context, projectID uint) ([5][5]MatrixCell, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "risk", 0, 10000)
	if err != nil {
		return [5][5]MatrixCell{}, err
	}
	inputs := make([]RiskInput, 0, len(entities))
	for _, e := range entities {
		inputs = append(inputs, RiskInput{
			ID:          e.ID,
			Probability: toInt(e.Attrs["probability"]),
			Impact:      toInt(e.Attrs["impact"]),
		})
	}
	return RiskMatrix(inputs), nil
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}
