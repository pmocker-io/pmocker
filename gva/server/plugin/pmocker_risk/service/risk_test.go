package risk

import "testing"

func TestCalcRiskLevel(t *testing.T) {
	cases := []struct {
		p, i  int
		want  RiskLevel
		score int
	}{
		{1, 1, RiskLow, 1},
		{2, 2, RiskMedium, 4},
		{3, 4, RiskHigh, 12},
		{5, 5, RiskCritical, 25},
	}
	for _, c := range cases {
		level, score := CalcRiskLevel(c.p, c.i)
		if level != c.want || score != c.score {
			t.Errorf("Calc(%d,%d) = %s/%d, want %s/%d", c.p, c.i, level, score, c.want, c.score)
		}
	}
}

func TestRiskMatrix(t *testing.T) {
	risks := []RiskInput{
		{ID: 1, Probability: 5, Impact: 5},
		{ID: 2, Probability: 1, Impact: 1},
	}
	m := RiskMatrix(risks)
	if m[4][4].Count != 1 {
		t.Errorf("critical cell count = %d, want 1", m[4][4].Count)
	}
	if m[0][0].Count != 1 {
		t.Errorf("low cell count = %d, want 1", m[0][0].Count)
	}
	if m[4][4].Level != RiskCritical {
		t.Errorf("cell level = %s, want critical", m[4][4].Level)
	}
}
