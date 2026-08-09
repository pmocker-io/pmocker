package issue

import (
	"context"
	"fmt"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var IssueService = new(Service)

var IssueStatuses = []string{"open", "assigned", "in_progress", "resolved", "verified", "closed", "reopened"}

func (s *Service) CreateIssue(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "issue", Title: title, Status: "open", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) GetIssue(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "issue" {
		return nil, fmt.Errorf("entity %d is not an issue", id)
	}
	return e, nil
}

func (s *Service) ListIssues(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "issue", offset, limit)
}

func (s *Service) UpdateIssue(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "issue" {
		return fmt.Errorf("not an issue")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

func (s *Service) DeleteIssue(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *Service) AssignIssue(ctx context.Context, id uint, assigneeID uint) error {
	e, err := s.GetIssue(ctx, id)
	if err != nil {
		return err
	}
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["assignee"] = assigneeID
	e.Status = "assigned"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) ResolveIssue(ctx context.Context, id uint, resolution, rootCause string) error {
	e, err := s.GetIssue(ctx, id)
	if err != nil {
		return err
	}
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	e.Attrs["resolution"] = resolution
	e.Attrs["root_cause"] = rootCause
	e.Status = "resolved"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) CloseIssue(ctx context.Context, id uint) error {
	e, err := s.GetIssue(ctx, id)
	if err != nil {
		return err
	}
	e.Status = "closed"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) ReopenIssue(ctx context.Context, id uint) error {
	e, err := s.GetIssue(ctx, id)
	if err != nil {
		return err
	}
	e.Status = "reopened"
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func (s *Service) IssueBoard(ctx context.Context, projectID uint) (map[string][]eavtypes.Entity, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "issue", 0, 10000)
	if err != nil {
		return nil, err
	}
	board := make(map[string][]eavtypes.Entity)
	for _, status := range IssueStatuses {
		board[status] = []eavtypes.Entity{}
	}
	for _, e := range entities {
		status := e.Status
		if _, ok := board[status]; !ok {
			board[status] = []eavtypes.Entity{}
		}
		board[status] = append(board[status], e)
	}
	return board, nil
}

func (s *Service) IssueStats(ctx context.Context, projectID uint) (map[string]int, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "issue", 0, 10000)
	if err != nil {
		return nil, err
	}
	stats := make(map[string]int)
	for _, status := range IssueStatuses {
		stats[status] = 0
	}
	for _, e := range entities {
		stats[e.Status]++
	}
	return stats, nil
}

func BuildIssueBoard(entities []eavtypes.Entity) map[string][]eavtypes.Entity {
	board := make(map[string][]eavtypes.Entity)
	for _, status := range IssueStatuses {
		board[status] = []eavtypes.Entity{}
	}
	for _, e := range entities {
		status := e.Status
		if _, ok := board[status]; !ok {
			board[status] = []eavtypes.Entity{}
		}
		board[status] = append(board[status], e)
	}
	return board
}

func BuildIssueStats(entities []eavtypes.Entity) map[string]int {
	stats := make(map[string]int)
	for _, status := range IssueStatuses {
		stats[status] = 0
	}
	for _, e := range entities {
		stats[e.Status]++
	}
	return stats
}
