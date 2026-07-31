package issue

import (
	"testing"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

func TestBuildIssueBoard(t *testing.T) {
	entities := []eavtypes.Entity{
		{ID: 1, Status: "open", Title: "Issue 1"},
		{ID: 2, Status: "open", Title: "Issue 2"},
		{ID: 3, Status: "assigned", Title: "Issue 3"},
		{ID: 4, Status: "in_progress", Title: "Issue 4"},
		{ID: 5, Status: "resolved", Title: "Issue 5"},
		{ID: 6, Status: "closed", Title: "Issue 6"},
	}
	board := BuildIssueBoard(entities)
	if len(board["open"]) != 2 {
		t.Errorf("open count = %d, want 2", len(board["open"]))
	}
	if len(board["assigned"]) != 1 {
		t.Errorf("assigned count = %d, want 1", len(board["assigned"]))
	}
	if len(board["in_progress"]) != 1 {
		t.Errorf("in_progress count = %d, want 1", len(board["in_progress"]))
	}
	if len(board["resolved"]) != 1 {
		t.Errorf("resolved count = %d, want 1", len(board["resolved"]))
	}
	if len(board["closed"]) != 1 {
		t.Errorf("closed count = %d, want 1", len(board["closed"]))
	}
	if len(board["verified"]) != 0 {
		t.Errorf("verified count = %d, want 0", len(board["verified"]))
	}
	if len(board["reopened"]) != 0 {
		t.Errorf("reopened count = %d, want 0", len(board["reopened"]))
	}
	if board["open"][0].ID != 1 {
		t.Errorf("first open issue id = %d, want 1", board["open"][0].ID)
	}
}

func TestBuildIssueStats(t *testing.T) {
	entities := []eavtypes.Entity{
		{ID: 1, Status: "open"},
		{ID: 2, Status: "open"},
		{ID: 3, Status: "assigned"},
		{ID: 4, Status: "in_progress"},
		{ID: 5, Status: "resolved"},
		{ID: 6, Status: "verified"},
		{ID: 7, Status: "closed"},
		{ID: 8, Status: "reopened"},
	}
	stats := BuildIssueStats(entities)
	if stats["open"] != 2 {
		t.Errorf("open stats = %d, want 2", stats["open"])
	}
	if stats["assigned"] != 1 {
		t.Errorf("assigned stats = %d, want 1", stats["assigned"])
	}
	if stats["in_progress"] != 1 {
		t.Errorf("in_progress stats = %d, want 1", stats["in_progress"])
	}
	if stats["resolved"] != 1 {
		t.Errorf("resolved stats = %d, want 1", stats["resolved"])
	}
	if stats["verified"] != 1 {
		t.Errorf("verified stats = %d, want 1", stats["verified"])
	}
	if stats["closed"] != 1 {
		t.Errorf("closed stats = %d, want 1", stats["closed"])
	}
	if stats["reopened"] != 1 {
		t.Errorf("reopened stats = %d, want 1", stats["reopened"])
	}
}
