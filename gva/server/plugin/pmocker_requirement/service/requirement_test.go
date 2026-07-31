package requirement

import (
	"testing"
)

func TestTraceRowSerialization(t *testing.T) {
	row := TraceRow{
		RequirementID:    1,
		RequirementTitle: "登录功能",
		Status:           "approved",
		DeliverableIDs:   []uint{10, 11},
	}
	if row.RequirementID != 1 {
		t.Errorf("id = %d", row.RequirementID)
	}
	if len(row.DeliverableIDs) != 2 {
		t.Errorf("deliverable count = %d", len(row.DeliverableIDs))
	}
}
