package change

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"gorm.io/gorm"
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

// FieldDiff 变更前后字段级差异
type FieldDiff struct {
	FieldKey   string `json:"fieldKey"`
	FieldLabel string `json:"fieldLabel"`
	OldValue   string `json:"oldValue"`
	NewValue   string `json:"newValue"`
	Changed    bool   `json:"changed"`
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
	// 提交评审时冻结基线快照，作为变更前后 diff 的对比基准
	if err := s.snapshotBaseline(cr); err != nil {
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

// snapshotBaseline 将当前 attrs（排除 baseline_snapshot 自身）序列化为 JSON 存入 baseline_snapshot 字段。
// 在提交评审时调用，冻结变更请求的基线状态用于后续 diff 对比。
func (s *Service) snapshotBaseline(cr *eavtypes.Entity) error {
	if cr.Attrs == nil {
		cr.Attrs = map[string]interface{}{}
	}
	snap := make(map[string]interface{}, len(cr.Attrs))
	for k, v := range cr.Attrs {
		if k == "baseline_snapshot" {
			continue
		}
		snap[k] = v
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("marshal baseline snapshot: %w", err)
	}
	cr.Attrs["baseline_snapshot"] = string(b)
	return nil
}

// GetDiff 对比 change_request 的基线快照(baseline_snapshot)与当前 attrs，返回字段级差异列表。
// old_value 为空表示新增字段，new_value 为空表示删除字段，changed=true 表示值发生变化。
// ref 字段暂以 ID 形式返回（解析显示名需跨模块查询，此处保持轻量）。
func (s *Service) GetDiff(ctx context.Context, changeID uint) ([]FieldDiff, error) {
	cr, err := s.GetChangeRequest(ctx, changeID)
	if err != nil {
		return nil, err
	}

	// 加载字段定义，用于字段标签与排序（加载失败时降级为以字段键作为标签）
	defs, _ := pmservice.ServiceGroupApp.LoadFieldDefs(ctx, "change_request")
	labelMap := make(map[string]string, len(defs))
	defOrder := make([]string, 0, len(defs))
	for _, d := range defs {
		labelMap[d.FieldKey] = d.FieldLabel
		defOrder = append(defOrder, d.FieldKey)
	}

	// 解析基线快照（存储为 JSON 字符串）
	baseline := map[string]interface{}{}
	if cr.Attrs != nil {
		if snap, ok := cr.Attrs["baseline_snapshot"].(string); ok && snap != "" {
			_ = json.Unmarshal([]byte(snap), &baseline)
		}
	}
	current := cr.Attrs
	if current == nil {
		current = map[string]interface{}{}
	}

	// 收集所有字段键（排除 baseline_snapshot 自身）
	keySet := make(map[string]bool)
	for k := range baseline {
		if k != "baseline_snapshot" {
			keySet[k] = true
		}
	}
	for k := range current {
		if k != "baseline_snapshot" {
			keySet[k] = true
		}
	}

	// 按 schema 定义顺序排列，未定义字段追加在末尾（按字母序）
	ordered := make([]string, 0, len(keySet))
	for _, k := range defOrder {
		if keySet[k] {
			ordered = append(ordered, k)
			delete(keySet, k)
		}
	}
	extras := make([]string, 0, len(keySet))
	for k := range keySet {
		extras = append(extras, k)
	}
	sort.Strings(extras)
	ordered = append(ordered, extras...)

	diffs := make([]FieldDiff, 0, len(ordered))
	for _, k := range ordered {
		oldV := formatDiffValue(baseline[k])
		newV := formatDiffValue(current[k])
		diffs = append(diffs, FieldDiff{
			FieldKey:   k,
			FieldLabel: labelOf(labelMap, k),
			OldValue:   oldV,
			NewValue:   newV,
			Changed:    oldV != newV,
		})
	}
	return diffs, nil
}

// labelOf 取字段标签，缺省回退为字段键
func labelOf(labelMap map[string]string, key string) string {
	if l, ok := labelMap[key]; ok && l != "" {
		return l
	}
	return key
}

// formatDiffValue 将任意属性值格式化为可读字符串，nil/空值统一为空串以便 diff 比较
func formatDiffValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case map[string]interface{}:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	case []interface{}:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", val)
	}
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
	var projectID uint
	if changeID > 0 {
		cr, err := s.GetChangeRequest(ctx, changeID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "is not a change_request") {
				return []eavtypes.Entity{}, nil
			}
			return nil, err
		}
		projectID = cr.ProjectID
	}
	logs, _, listErr := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "change_log", 0, 10000)
	if listErr != nil {
		return nil, listErr
	}
	if changeID == 0 {
		return logs, nil
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
