package deliverable

import (
	"context"
	"fmt"
	"sort"
	"strings"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var DeliverableService = new(Service)

type DeliverableStatus string

const (
	StatusDraft      DeliverableStatus = "draft"
	StatusSubmitted  DeliverableStatus = "submitted"
	StatusReviewing  DeliverableStatus = "reviewing"
	StatusAccepted   DeliverableStatus = "accepted"
	StatusRejected   DeliverableStatus = "rejected"
	StatusBaselined  DeliverableStatus = "baselined"
	StatusObsolete   DeliverableStatus = "obsolete"
)

type VersionStatus string

const (
	VersionActive   VersionStatus = "active"
	VersionArchived VersionStatus = "archived"
)

type DeliverableCategory string

const (
	CategoryDocument DeliverableCategory = "document"
	CategorySoftware DeliverableCategory = "software"
	CategoryHardware DeliverableCategory = "hardware"
	CategoryService  DeliverableCategory = "service"
	CategoryTraining DeliverableCategory = "training"
)

type DeliverableStats struct {
	ByStatus   map[string]int            `json:"byStatus"`
	ByCategory map[string]int            `json:"byCategory"`
	Total      int                       `json:"total"`
}

type TraceItem struct {
	DeliverableID   uint     `json:"deliverableId"`
	DeliverableName string   `json:"deliverableName"`
	DeliverableCode string   `json:"deliverableCode"`
	Version         string   `json:"version"`
	Category        string   `json:"category"`
	Status          string   `json:"status"`
	RequirementIDs  []string `json:"requirementIds"`
}

type TraceabilityReport struct {
	ProjectID          uint                   `json:"projectId"`
	Deliverables       []TraceItem            `json:"deliverables"`
	AllRequirementIDs  []string               `json:"allRequirementIds"`
	CoveredReqIDs      []string               `json:"coveredReqIds"`
	UncoveredReqIDs    []string               `json:"uncoveredReqIds"`
	DeliverableCount   int                    `json:"deliverableCount"`
	TotalReqCount      int                    `json:"totalReqCount"`
	CoveredReqCount    int                    `json:"coveredReqCount"`
	CoverageRate       float64                `json:"coverageRate"`
}

type CreateVersionResult struct {
	VersionID uint   `json:"versionId"`
	Version   string `json:"version"`
}

func (s *Service) CreateDeliverable(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	if attrs == nil {
		attrs = map[string]interface{}{}
	}
	if _, ok := attrs["version"]; !ok {
		attrs["version"] = "1.0"
	}
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "deliverable", Title: title, Status: string(StatusDraft), CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) GetDeliverable(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "deliverable" {
		return nil, fmt.Errorf("entity %d is not a deliverable", id)
	}
	return e, nil
}

func (s *Service) ListDeliverables(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "deliverable", offset, limit)
}

func (s *Service) UpdateDeliverable(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "deliverable" {
		return fmt.Errorf("not a deliverable")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

func (s *Service) DeleteDeliverable(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *Service) SubmitForReview(ctx context.Context, id uint) error {
	e, err := s.GetDeliverable(ctx, id)
	if err != nil {
		return err
	}
	e.Status = string(StatusSubmitted)
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) AcceptDeliverable(ctx context.Context, id uint, reviewerComment string) error {
	e, err := s.GetDeliverable(ctx, id)
	if err != nil {
		return err
	}
	e.Status = string(StatusAccepted)
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["reviewer_comment"] = reviewerComment
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) RejectDeliverable(ctx context.Context, id uint, reason string) error {
	e, err := s.GetDeliverable(ctx, id)
	if err != nil {
		return err
	}
	e.Status = string(StatusRejected)
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["reject_reason"] = reason
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) CreateVersion(ctx context.Context, deliverableID uint, version, changeLog string, creatorID uint, fileRef string) (CreateVersionResult, error) {
	deliverable, err := s.GetDeliverable(ctx, deliverableID)
	if err != nil {
		return CreateVersionResult{}, err
	}

	versions, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, deliverable.ProjectID, "deliverable_version", 0, 10000)
	if err != nil {
		return CreateVersionResult{}, err
	}

	for _, v := range versions {
		if v.Attrs != nil && toUint(v.Attrs["deliverable_id"]) == deliverableID {
			if v.Status == string(VersionActive) {
				v.Status = string(VersionArchived)
				if v.Attrs == nil {
					v.Attrs = map[string]interface{}{}
				}
				v.Attrs["is_current"] = false
				if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, v); err != nil {
					return CreateVersionResult{}, err
				}
			}
		}
	}

	attrs := map[string]interface{}{
		"deliverable_id": deliverableID,
		"version":        version,
		"change_log":     changeLog,
		"created_by":     creatorID,
		"file_ref":       fileRef,
		"is_current":     true,
	}

	versionID, err := pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: deliverable.ProjectID, EntityType: "deliverable_version", Title: fmt.Sprintf("v%s", version), Status: string(VersionActive), CreatedBy: creatorID, Attrs: attrs,
	})
	if err != nil {
		return CreateVersionResult{}, err
	}

	deliverable.Attrs["version"] = version
	if fileRef != "" {
		deliverable.Attrs["file_ref"] = fileRef
	}
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, *deliverable); err != nil {
		return CreateVersionResult{}, err
	}

	return CreateVersionResult{VersionID: versionID, Version: version}, nil
}

func (s *Service) ListVersions(ctx context.Context, deliverableID uint) ([]eavtypes.Entity, error) {
	e, err := s.GetDeliverable(ctx, deliverableID)
	if err != nil {
		return nil, err
	}

	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, e.ProjectID, "deliverable_version", 0, 10000)
	if err != nil {
		return nil, err
	}

	var result []eavtypes.Entity
	for _, ent := range entities {
		if ent.Attrs != nil && toUint(ent.Attrs["deliverable_id"]) == deliverableID {
			result = append(result, ent)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		vi := toString(result[i].Attrs["version"])
		vj := toString(result[j].Attrs["version"])
		return compareVersion(vi, vj) > 0
	})

	return result, nil
}

func (s *Service) BaselineDeliverable(ctx context.Context, id uint) error {
	e, err := s.GetDeliverable(ctx, id)
	if err != nil {
		return err
	}
	e.Status = string(StatusBaselined)
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["baselined"] = true
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) TraceabilityReport(ctx context.Context, projectID uint) (*TraceabilityReport, error) {
	deliverables, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "deliverable", 0, 10000)
	if err != nil {
		return nil, err
	}

	coverSet := make(map[string]bool)
	items := make([]TraceItem, 0, len(deliverables))

	for _, d := range deliverables {
		item := TraceItem{
			DeliverableID:   d.ID,
			DeliverableName: d.Title,
			Status:          d.Status,
		}
		if d.Attrs != nil {
			item.DeliverableCode = toString(d.Attrs["code"])
			item.Version = toString(d.Attrs["version"])
			item.Category = toString(d.Attrs["category"])
			relStr := toString(d.Attrs["related_requirements"])
			if relStr != "" {
				reqs := strings.Split(strings.ReplaceAll(relStr, " ", ""), ",")
				unique := make([]string, 0, len(reqs))
				for _, r := range reqs {
					r = strings.TrimSpace(r)
					if r != "" {
						unique = append(unique, r)
						coverSet[r] = true
					}
				}
				item.RequirementIDs = unique
			}
		}
		items = append(items, item)
	}

	allSet := make(map[string]bool)
	for _, it := range items {
		for _, r := range it.RequirementIDs {
			allSet[r] = true
		}
	}

	allIDs := make([]string, 0, len(allSet))
	for r := range allSet {
		allIDs = append(allIDs, r)
	}
	sort.Strings(allIDs)

	coveredIDs := make([]string, 0, len(coverSet))
	for r := range coverSet {
		coveredIDs = append(coveredIDs, r)
	}
	sort.Strings(coveredIDs)

	var uncoveredIDs []string
	for _, r := range allIDs {
		if !coverSet[r] {
			uncoveredIDs = append(uncoveredIDs, r)
		}
	}

	coverageRate := 0.0
	if len(allIDs) > 0 {
		coverageRate = float64(len(coveredIDs)) / float64(len(allIDs)) * 100
	}

	return &TraceabilityReport{
		ProjectID:         projectID,
		Deliverables:      items,
		AllRequirementIDs: allIDs,
		CoveredReqIDs:     coveredIDs,
		UncoveredReqIDs:   uncoveredIDs,
		DeliverableCount:  len(items),
		TotalReqCount:     len(allIDs),
		CoveredReqCount:   len(coveredIDs),
		CoverageRate:      coverageRate,
	}, nil
}

func (s *Service) DeliverableStats(ctx context.Context, projectID uint) (*DeliverableStats, error) {
	deliverables, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "deliverable", 0, 10000)
	if err != nil {
		return nil, err
	}

	stats := &DeliverableStats{
		ByStatus:   make(map[string]int),
		ByCategory: make(map[string]int),
		Total:      len(deliverables),
	}

	for _, d := range deliverables {
		stats.ByStatus[d.Status]++
		if d.Attrs != nil {
			cat := toString(d.Attrs["category"])
			if cat != "" {
				stats.ByCategory[cat]++
			}
		}
	}

	return stats, nil
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

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		return fmt.Sprintf("%v", v)
	}
}

func compareVersion(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	maxLen := len(ap)
	if len(bp) > maxLen {
		maxLen = len(bp)
	}
	for i := 0; i < maxLen; i++ {
		an, bn := 0, 0
		if i < len(ap) {
			fmt.Sscanf(ap[i], "%d", &an)
		}
		if i < len(bp) {
			fmt.Sscanf(bp[i], "%d", &bn)
		}
		if an != bn {
			return an - bn
		}
	}
	return 0
}
