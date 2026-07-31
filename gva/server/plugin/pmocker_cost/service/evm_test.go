package cost

import (
	"math"
	"testing"
)

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestEVM(t *testing.T) {
	// BAC=1000, PV=600, EV=500, AC=550
	r := EVM(EVMInput{BAC: 1000, PV: 600, EV: 500, AC: 550})
	if !approxEqual(r.CV, -50) {
		t.Errorf("CV = %f, want -50", r.CV)
	}
	if !approxEqual(r.SV, -100) {
		t.Errorf("SV = %f, want -100", r.SV)
	}
	if !approxEqual(r.CPI, 500.0/550) {
		t.Errorf("CPI = %f, want %f", r.CPI, 500.0/550)
	}
	if !approxEqual(r.SPI, 500.0/600) {
		t.Errorf("SPI = %f, want %f", r.SPI, 500.0/600)
	}
	expectedEAC := 1000 / (500.0 / 550)
	if !approxEqual(r.EAC, expectedEAC) {
		t.Errorf("EAC = %f, want %f", r.EAC, expectedEAC)
	}
	if !approxEqual(r.VAC, 1000-expectedEAC) {
		t.Errorf("VAC = %f, want %f", r.VAC, 1000-expectedEAC)
	}
}

func TestEVMZeroAC(t *testing.T) {
	// AC=0 时不进行除法计算，CPI/EAC 应为 0
	r := EVM(EVMInput{BAC: 1000, PV: 600, EV: 500, AC: 0})
	if r.CPI != 0 {
		t.Errorf("CPI = %f, want 0", r.CPI)
	}
	if r.EAC != 0 {
		t.Errorf("EAC = %f, want 0", r.EAC)
	}
}
