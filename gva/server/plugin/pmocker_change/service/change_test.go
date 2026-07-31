package change

import "testing"

func TestCalcImpactRisk(t *testing.T) {
	cases := []struct {
		score int
		want  string
	}{
		{0, "low"},
		{2, "low"},
		{3, "medium"},
		{5, "medium"},
		{6, "high"},
		{8, "high"},
		{9, "critical"},
		{15, "critical"},
	}
	for _, c := range cases {
		got := CalcImpactRisk(c.score)
		if got != c.want {
			t.Errorf("CalcImpactRisk(%d) = %s, want %s", c.score, got, c.want)
		}
	}
}

func TestCCBStatsMatrix(t *testing.T) {
	statuses := []string{"draft", "submitted", "analyzing", "ccb_review", "approved", "rejected", "implementing", "verified", "closed", "cancelled"}
	types := []string{"scope", "schedule", "cost", "quality", "resource", "risk", "other"}

	matrix := make(map[string]map[string]int)
	for _, st := range statuses {
		matrix[st] = make(map[string]int)
		for _, tp := range types {
			matrix[st][tp] = 0
		}
	}

	matrix["ccb_review"]["scope"] = 3
	matrix["ccb_review"]["schedule"] = 2
	matrix["approved"]["cost"] = 1

	if matrix["ccb_review"]["scope"] != 3 {
		t.Errorf("matrix[ccb_review][scope] = %d, want 3", matrix["ccb_review"]["scope"])
	}
	if matrix["ccb_review"]["schedule"] != 2 {
		t.Errorf("matrix[ccb_review][schedule] = %d, want 2", matrix["ccb_review"]["schedule"])
	}
	if matrix["approved"]["cost"] != 1 {
		t.Errorf("matrix[approved][cost] = %d, want 1", matrix["approved"]["cost"])
	}
	if matrix["draft"]["other"] != 0 {
		t.Errorf("matrix[draft][other] = %d, want 0", matrix["draft"]["other"])
	}

	pendingStatuses := map[string]bool{"submitted": true, "analyzing": true, "ccb_review": true}
	pendingCount := 0
	for _, st := range statuses {
		if pendingStatuses[st] {
			for _, tp := range types {
				pendingCount += matrix[st][tp]
			}
		}
	}
	expectedPending := 3 + 2
	if pendingCount != expectedPending {
		t.Errorf("pendingCount = %d, want %d", pendingCount, expectedPending)
	}
}

func TestToInt(t *testing.T) {
	cases := []struct {
		v    interface{}
		want int
	}{
		{int(42), 42},
		{int32(42), 42},
		{int64(42), 42},
		{float32(42.7), 42},
		{float64(42.9), 42},
		{"invalid", 0},
		{nil, 0},
	}
	for _, c := range cases {
		got := toInt(c.v)
		if got != c.want {
			t.Errorf("toInt(%v) = %d, want %d", c.v, got, c.want)
		}
	}
}

func TestToFloat64(t *testing.T) {
	cases := []struct {
		v    interface{}
		want float64
	}{
		{float64(3.14), 3.14},
		{float32(2.5), 2.5},
		{int(10), 10.0},
		{int64(20), 20.0},
		{"invalid", 0},
	}
	for _, c := range cases {
		got := toFloat64(c.v)
		if got != c.want {
			t.Errorf("toFloat64(%v) = %f, want %f", c.v, got, c.want)
		}
	}
}
