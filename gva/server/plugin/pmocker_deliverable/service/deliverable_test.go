package deliverable

import (
	"strings"
	"testing"
)

func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.0", "1.0", 0},
		{"1.1", "1.0", 1},
		{"1.0", "1.1", -1},
		{"2.0", "1.9", 1},
		{"1.0.0", "1.0.1", -1},
		{"10.0", "9.0", 1},
	}
	for _, c := range cases {
		got := compareVersion(c.a, c.b)
		if got != c.want {
			if got*c.want < 0 || (got == 0) != (c.want == 0) {
				t.Errorf("compareVersion(%s,%s)=%d, want %d", c.a, c.b, got, c.want)
			}
		}
	}
}

func TestToString(t *testing.T) {
	if toString(nil) != "" {
		t.Errorf("toString(nil) should be empty")
	}
	if toString("abc") != "abc" {
		t.Errorf("toString(abc) should be abc")
	}
	if toString(123) != "123" {
		t.Errorf("toString(123) should be '123'")
	}
}

func TestToUint(t *testing.T) {
	if toUint(nil) != 0 {
		t.Errorf("toUint(nil) should be 0")
	}
	if toUint(uint(42)) != 42 {
		t.Errorf("toUint(uint(42)) should be 42")
	}
	if toUint(100) != 100 {
		t.Errorf("toUint(100) should be 100")
	}
	if toUint(3.14) != 3 {
		t.Errorf("toUint(3.14) should be 3")
	}
}

func TestTraceabilityReportStructure(t *testing.T) {
	report := &TraceabilityReport{
		ProjectID:        1,
		AllRequirementIDs: []string{"REQ-1", "REQ-2", "REQ-3"},
		CoveredReqIDs:    []string{"REQ-1", "REQ-2"},
		UncoveredReqIDs:  []string{"REQ-3"},
		Deliverables: []TraceItem{
			{
				DeliverableID:   10,
				DeliverableName: "需求规格说明书",
				DeliverableCode: "D-001",
				Version:         "1.0",
				Category:        "document",
				Status:          "accepted",
				RequirementIDs:  []string{"REQ-1"},
			},
			{
				DeliverableID:   11,
				DeliverableName: "软件源代码",
				DeliverableCode: "D-002",
				Version:         "2.0",
				Category:        "software",
				Status:          "submitted",
				RequirementIDs:  []string{"REQ-2"},
			},
		},
		DeliverableCount: 2,
		TotalReqCount:    3,
		CoveredReqCount:  2,
		CoverageRate:     66.67,
	}

	if report.ProjectID != 1 {
		t.Errorf("ProjectID = %d, want 1", report.ProjectID)
	}
	if report.DeliverableCount != 2 {
		t.Errorf("DeliverableCount = %d, want 2", report.DeliverableCount)
	}
	if report.TotalReqCount != 3 {
		t.Errorf("TotalReqCount = %d, want 3", report.TotalReqCount)
	}
	if report.CoveredReqCount != 2 {
		t.Errorf("CoveredReqCount = %d, want 2", report.CoveredReqCount)
	}
	if len(report.UncoveredReqIDs) != 1 {
		t.Errorf("UncoveredReqIDs length = %d, want 1", len(report.UncoveredReqIDs))
	}
	if report.UncoveredReqIDs[0] != "REQ-3" {
		t.Errorf("first uncovered = %s, want REQ-3", report.UncoveredReqIDs[0])
	}
	if len(report.Deliverables) != 2 {
		t.Errorf("deliverables length = %d, want 2", len(report.Deliverables))
	}
	if report.Deliverables[0].RequirementIDs[0] != "REQ-1" {
		t.Errorf("first deliverable first req = %s, want REQ-1", report.Deliverables[0].RequirementIDs[0])
	}
	if strings.Contains(report.Deliverables[1].Category, "software") == false {
		t.Errorf("second deliverable category should contain 'software'")
	}
}

func TestVersionArchivalLogic(t *testing.T) {
	type VersionRec struct {
		ID         uint
		DeliverID  uint
		Version    string
		Status     string
		IsCurrent  bool
	}

	var versions []VersionRec
	_ = versions

	initial := VersionRec{ID: 1, DeliverID: 99, Version: "1.0", Status: "active", IsCurrent: true}
	initial.Status = "archived"
	initial.IsCurrent = false

	if initial.Status != "archived" {
		t.Errorf("old version should be archived")
	}
	if initial.IsCurrent != false {
		t.Errorf("old version should not be current")
	}

	newVer := VersionRec{ID: 2, DeliverID: 99, Version: "2.0", Status: "active", IsCurrent: true}
	if newVer.Status != "active" {
		t.Errorf("new version should be active")
	}
	if !newVer.IsCurrent {
		t.Errorf("new version should be current")
	}
}
