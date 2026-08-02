package pmocker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type BaselineService struct{}

var baselineEntityTypes = map[string][]string{
	"schedule": {"task"},
	"cost":     {"task", "cost_item"},
	"scope":    {"scope_item"},
}

type snapshotEntity struct {
	EntityID   uint              `json:"entityId"`
	EntityType string            `json:"entityType"`
	Title      string            `json:"title"`
	Status     string            `json:"status"`
	OwnerID    *uint             `json:"ownerId,omitempty"`
	Attrs      map[string]string `json:"attrs"`
}

type baselineSnapshot struct {
	ProjectID uint             `json:"projectId"`
	Type      string           `json:"type"`
	CreatedAt string           `json:"createdAt"`
	Entities  []snapshotEntity `json:"entities"`
}

type BaselineDiff struct {
	EntityType  string `json:"entityType"`
	EntityID    uint   `json:"entityId"`
	EntityTitle string `json:"entityTitle"`
	FieldKey    string `json:"fieldKey"`
	BaselineVal string `json:"baselineVal"`
	CurrentVal  string `json:"currentVal"`
	Change      string `json:"change"`
}

func (s *BaselineService) CreateBaseline(projectID uint, baselineType string, changeReqID *uint, createdBy uint) (uint, error) {
	types, ok := baselineEntityTypes[baselineType]
	if !ok {
		return 0, fmt.Errorf("unsupported baseline type: %s", baselineType)
	}

	var entities []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type IN ?", projectID, types).
		Order("entity_type, id").Find(&entities).Error; err != nil {
		return 0, err
	}

	snap := baselineSnapshot{
		ProjectID: projectID,
		Type:      baselineType,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Entities:  make([]snapshotEntity, 0, len(entities)),
	}

	for _, e := range entities {
		var attrs []pmocker.PMAttr
		global.GVA_DB.Where("entity_id = ?", e.ID).Find(&attrs)
		attrMap := make(map[string]string, len(attrs))
		for _, a := range attrs {
			attrMap[a.FieldKey] = attrValueString(a)
		}
		snap.Entities = append(snap.Entities, snapshotEntity{
			EntityID: e.ID, EntityType: e.EntityType, Title: e.Title,
			Status: e.Status, OwnerID: e.OwnerID, Attrs: attrMap,
		})
	}

	data, err := json.Marshal(snap)
	if err != nil {
		return 0, err
	}

	bl := pmocker.PMBaseline{
		ProjectID:    projectID,
		Type:         baselineType,
		SnapshotJSON: string(data),
		ChangeReqID:  changeReqID,
	}
	if err := global.GVA_DB.Create(&bl).Error; err != nil {
		return 0, err
	}
	global.GVA_DB.Model(&pmocker.PMEntity{}).Where("id = ?", projectID).Update("baseline_id", bl.ID)
	return bl.ID, nil
}

func (s *BaselineService) ListBaselines(projectID uint, baselineType string) ([]pmocker.PMBaseline, error) {
	var list []pmocker.PMBaseline
	db := global.GVA_DB.Where("project_id = ?", projectID)
	if baselineType != "" {
		db = db.Where("type = ?", baselineType)
	}
	err := db.Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *BaselineService) CompareBaseline(baselineID uint) ([]BaselineDiff, error) {
	var bl pmocker.PMBaseline
	if err := global.GVA_DB.First(&bl, baselineID).Error; err != nil {
		return nil, err
	}

	var snap baselineSnapshot
	if err := json.Unmarshal([]byte(bl.SnapshotJSON), &snap); err != nil {
		return nil, fmt.Errorf("invalid snapshot json: %w", err)
	}

	current := make(map[uint]map[string]string)
	currentMeta := make(map[uint]snapshotEntity)
	for _, se := range snap.Entities {
		var attrs []pmocker.PMAttr
		global.GVA_DB.Where("entity_id = ?", se.EntityID).Find(&attrs)
		m := make(map[string]string, len(attrs))
		for _, a := range attrs {
			m[a.FieldKey] = attrValueString(a)
		}
		current[se.EntityID] = m
		var e pmocker.PMEntity
		global.GVA_DB.First(&e, se.EntityID)
		currentMeta[se.EntityID] = snapshotEntity{
			EntityID: e.ID, EntityType: e.EntityType, Title: e.Title, Status: e.Status,
		}
	}

	diffs := make([]BaselineDiff, 0)
	for _, se := range snap.Entities {
		curAttrs := current[se.EntityID]
		meta := currentMeta[se.EntityID]
		title := meta.Title
		if title == "" {
			title = se.Title
		}
		for k, oldV := range se.Attrs {
			newV, exists := curAttrs[k]
			if !exists {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, oldV, "", "removed"})
				continue
			}
			if oldV != newV {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, oldV, newV, "modified"})
			}
		}
		for k, newV := range curAttrs {
			if _, exists := se.Attrs[k]; !exists {
				diffs = append(diffs, BaselineDiff{se.EntityType, se.EntityID, title, k, "", newV, "added"})
			}
		}
	}
	return diffs, nil
}
