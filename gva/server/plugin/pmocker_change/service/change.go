package change

import (
	"context"
	"fmt"
	"time"

	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

type Service struct{}

var ChangeService = new(Service)

type ImpactReport struct {
	ScheduleDays    int            `json:"scheduleDays"`
	CostAmount      float64        `json:"costAmount"`
	RiskLevel       string         `json:"riskLevel"`
	AffectedCount   int            `json:"affectedCount"`
	AffectedEntities []AffectedEntity `json:"affectedEntities"`
	OverallScore    int            `json:"overallScore"`
}

type AffectedEntity struct {
	EntityType string `json:"entityType"`
	EntityID   uint   `json:"entityId"`
	EntityName string `json:"entityName"`
	ImpactDesc string `json:"impactDesc"`
}

type CCBStatsResult struct {
	ByStatus map[string]int            `json:"byStatus"`
	ByType   map[string]int            `json:"byType"`
	Matrix   map[string]map[string]int `json:"matrix"`
	Total    int                       `json:"total"`
	Pending  int                       `json:"pending"`
}

func (s *Service) CreateChangeRequest(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	if attrs == nil {
		attrs = map[string]interface{}{}
	}
	if attrs["requested_date"] == nil {
		attrs["requested_date"] = time.Now().Format("2006-01-02")
	}
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  projectID,
		EntityType: "change_request",
		Title:      title,
		Status:     "draft",
		CreatedBy:  creatorID,
		Attrs:      attrs,
	})
}

func (s *Service) GetChangeRequest(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "change_request" {
		return nil, fmt.Errorf("entity %d is not a change_request", id)
	}
	return e, nil
}

func (s *Service) ListChangeRequests(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "change_request", offset, limit)
}

func (s *Service) UpdateChangeRequest(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "change_request" {
		return fmt.Errorf("not a change_request")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

func (s *Service) DeleteChangeRequest(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *Service) ImpactAnalysis(ctx context.Context, changeID uint) (*ImpactReport, error) {
	cr, err := s.GetChangeRequest(ctx, changeID)
	if err != nil {
		return nil, err
	}

	report := &ImpactReport{
		AffectedEntities: make([]AffectedEntity, 0),
	}

	if cr.Attrs != nil {
		report.ScheduleDays = toInt(cr.Attrs["impact_schedule"])
		report.CostAmount = toFloat64(cr.Attrs["impact_cost"])

		rels, relErr := pmservice.ServiceGroupApp.ListRelations(ctx, changeID)
		if relErr == nil {
			for _, rel := range rels {
				dst, dstErr := pmservice.ServiceGroupApp.GetEntity(ctx, rel.DstID)
				name := ""
				if dstErr == nil {
					name = dst.Title
				}
				report.AffectedEntities = append(report.AffectedEntities, AffectedEntity{
					EntityType: rel.RelationType,
					EntityID:   rel.DstID,
					EntityName: name,
					ImpactDesc: fmt.Sprintf("关联关系: %s", rel.RelationType),
				})
			}
		}

		if scope, ok := cr.Attrs["impact_scope"].(string); ok && scope != "" {
			report.AffectedEntities = append(report.AffectedEntities, AffectedEntity{
				EntityType: "scope",
				EntityID:   0,
				EntityName: "范围影响",
				ImpactDesc: scope,
			})
		}
		if quality, ok := cr.Attrs["impact_quality"].(string); ok && quality != "" {
			report.AffectedEntities = append(report.AffectedEntities, AffectedEntity{
				EntityType: "quality",
				EntityID:   0,
				EntityName: "质量影响",
				ImpactDesc: quality,
			})
		}
	}

	report.AffectedCount = len(report.AffectedEntities)

	scheduleScore := 0
	switch {
	case report.ScheduleDays <= 0:
		scheduleScore = 0
	case report.ScheduleDays <= 3:
		scheduleScore = 1
	case report.ScheduleDays <= 7:
		scheduleScore = 2
	case report.ScheduleDays <= 14:
		scheduleScore = 3
	default:
		scheduleScore = 5
	}

	costScore := 0
	switch {
	case report.CostAmount <= 0:
		costScore = 0
	case report.CostAmount <= 1000:
		costScore = 1
	case report.CostAmount <= 10000:
		costScore = 2
	case report.CostAmount <= 50000:
		costScore = 3
	default:
		costScore = 5
	}

	countScore := 0
	switch {
	case report.AffectedCount <= 1:
		countScore = 0
	case report.AffectedCount <= 3:
		countScore = 1
	case report.AffectedCount <= 5:
		countScore = 2
	case report.AffectedCount <= 10:
		countScore = 3
	default:
		countScore = 5
	}

	report.OverallScore = scheduleScore + costScore + countScore
	switch {
	case report.OverallScore <= 2:
		report.RiskLevel = "low"
	case report.OverallScore <= 5:
		report.RiskLevel = "medium"
	case report.OverallScore <= 8:
		report.RiskLevel = "high"
	default:
		report.RiskLevel = "critical"
	}

	return report, nil
}

func (s *Service) SubmitToCCB(ctx context.Context, id, userID uint) (uint, error) {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return 0, err
	}
	cr.Status = "submitted"
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr); err != nil {
		return 0, err
	}
	instID, err := pmservice.ServiceGroupApp.Start(ctx, id, "change_request")
	if err != nil {
		return 0, err
	}
	cr.Status = "ccb_review"
	if updateErr := pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr); updateErr != nil {
		return instID, updateErr
	}
	return instID, nil
}

func (s *Service) ApproveChange(ctx context.Context, id uint, decision string, userID uint) error {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return err
	}
	if cr.Attrs == nil {
		cr.Attrs = map[string]interface{}{}
	}
	cr.Attrs["ccb_decision"] = decision
	cr.Status = "approved"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr)
}

func (s *Service) RejectChange(ctx context.Context, id uint, reason string, userID uint) error {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return err
	}
	if cr.Attrs == nil {
		cr.Attrs = map[string]interface{}{}
	}
	cr.Attrs["ccb_decision"] = "rejected: " + reason
	cr.Status = "rejected"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr)
}

func (s *Service) StartImplementation(ctx context.Context, id, implementerID uint) error {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return err
	}
	if cr.Attrs == nil {
		cr.Attrs = map[string]interface{}{}
	}
	cr.Attrs["implemented_by"] = implementerID
	cr.Attrs["implemented_date"] = time.Now().Format("2006-01-02")
	cr.Status = "implementing"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr)
}

func (s *Service) VerifyChange(ctx context.Context, id uint, result string, userID uint) error {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return err
	}
	if cr.Attrs == nil {
		cr.Attrs = map[string]interface{}{}
	}
	cr.Attrs["validation_result"] = result
	cr.Status = "verified"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr)
}

func (s *Service) CloseChange(ctx context.Context, id, userID uint) error {
	cr, err := s.GetChangeRequest(ctx, id)
	if err != nil {
		return err
	}
	cr.Status = "closed"
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, *cr); err != nil {
		return err
	}

	logAttrs := map[string]interface{}{
		"change_request_id": id,
		"entity_type":       "change_request",
		"entity_id":         id,
		"field_name":        "status",
		"old_value":         cr.Status,
		"new_value":         "closed",
		"changed_by":        userID,
		"change_reason":     "变更关闭",
	}
	_, logErr := pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  cr.ProjectID,
		EntityType: "change_log",
		Title:      fmt.Sprintf("变更#%d关闭日志", id),
		Status:     "recorded",
		CreatedBy:  userID,
		Attrs:      logAttrs,
	})
	return logErr
}

func (s *Service) ListChangeLogs(ctx context.Context, changeID uint) ([]eavtypes.Entity, error) {
	cr, err := s.GetChangeRequest(ctx, changeID)
	if err != nil {
		return nil, err
	}
	logs, _, listErr := pmservice.ServiceGroupApp.ListEntities(ctx, cr.ProjectID, "change_log", 0, 10000)
	if listErr != nil {
		return nil, listErr
	}
	filtered := make([]eavtypes.Entity, 0, len(logs))
	for _, log := range logs {
		if log.Attrs != nil {
			if rid, ok := log.Attrs["change_request_id"]; ok {
				if toUint(rid) == changeID {
					filtered = append(filtered, log)
				}
			}
		}
	}
	return filtered, nil
}

func (s *Service) CCBStats(ctx context.Context, projectID uint) (*CCBStatsResult, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "change_request", 0, 10000)
	if err != nil {
		return nil, err
	}

	result := &CCBStatsResult{
		ByStatus: make(map[string]int),
		ByType:   make(map[string]int),
		Matrix:   make(map[string]map[string]int),
	}

	statuses := []string{"draft", "submitted", "analyzing", "ccb_review", "approved", "rejected", "implementing", "verified", "closed", "cancelled"}
	types := []string{"scope", "schedule", "cost", "quality", "resource", "risk", "other"}

	for _, st := range statuses {
		result.ByStatus[st] = 0
		result.Matrix[st] = make(map[string]int)
		for _, tp := range types {
			result.Matrix[st][tp] = 0
		}
	}
	for _, tp := range types {
		result.ByType[tp] = 0
	}

	pendingStatuses := map[string]bool{"submitted": true, "analyzing": true, "ccb_review": true}

	for _, e := range entities {
		result.Total++
		if _, ok := result.ByStatus[e.Status]; ok {
			result.ByStatus[e.Status]++
		}
		if pendingStatuses[e.Status] {
			result.Pending++
		}
		if e.Attrs != nil {
			tp := ""
			if t, ok := e.Attrs["type"].(string); ok {
				tp = t
			}
			if tp != "" {
				if _, ok := result.ByType[tp]; ok {
					result.ByType[tp]++
				}
				if m, ok := result.Matrix[e.Status]; ok {
					if _, ok2 := m[tp]; ok2 {
						m[tp]++
					}
				}
			}
		}
	}

	return result, nil
}

func toInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float32:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}

func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return 0
}

func toUint(v interface{}) uint {
	switch n := v.(type) {
	case uint:
		return n
	case int:
		return uint(n)
	case int64:
		return uint(n)
	case float64:
		return uint(n)
	}
	return 0
}

func CalcImpactRisk(overallScore int) string {
	switch {
	case overallScore <= 2:
		return "low"
	case overallScore <= 5:
		return "medium"
	case overallScore <= 8:
		return "high"
	default:
		return "critical"
	}
}
