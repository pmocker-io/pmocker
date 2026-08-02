package cost

import (
	"context"
	"fmt"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

// EVMInput 挣值计算输入
type EVMInput struct {
	BAC float64 // 完工预算 Budget at Completion
	PV  float64 // 计划价值 Planned Value（截止当前时点应完成的预算）
	EV  float64 // 挣值 Earned Value（已完成的预算价值）
	AC  float64 // 实际成本 Actual Cost（已发生的实际成本）
}

// EVMResult 挣值计算结果
type EVMResult struct {
	CV   float64 `json:"cv"`   // 成本偏差 = EV - AC
	SV   float64 `json:"sv"`   // 进度偏差 = EV - PV
	CPI  float64 `json:"cpi"`  // 成本绩效指数 = EV / AC
	SPI  float64 `json:"spi"`  // 进度绩效指数 = EV / PV
	EAC  float64 `json:"eac"`  // 完工估算 = BAC / CPI
	ETC  float64 `json:"etc"`  // 完工尚需估算 = EAC - AC
	VAC  float64 `json:"vac"`  // 完工偏差 = BAC - EAC
	TCPI float64 `json:"tcpi"` // 完工尚需绩效指数 = (BAC - EV) / (BAC - AC)
}

// EVM 计算挣值指标
func EVM(in EVMInput) EVMResult {
	r := EVMResult{
		CV: in.EV - in.AC,
		SV: in.EV - in.PV,
	}
	if in.AC != 0 {
		r.CPI = in.EV / in.AC
		r.EAC = in.BAC / r.CPI
		r.ETC = r.EAC - in.AC
	}
	if in.PV != 0 {
		r.SPI = in.EV / in.PV
	}
	r.VAC = in.BAC - r.EAC
	if in.BAC != in.AC {
		r.TCPI = (in.BAC - in.EV) / (in.BAC - in.AC)
	}
	return r
}

// EVMHandler 工作流 auto 节点处理器：读取 cost_item 实体，计算 EVM 派生字段并回写。
// handler 名：pmocker.cost.evm_calc
func EVMHandler(ctx context.Context, entityID uint) error {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, entityID)
	if err != nil {
		return fmt.Errorf("get cost entity %d: %w", entityID, err)
	}
	if e.EntityType != "cost_item" {
		return fmt.Errorf("entity %d is %q, expect cost_item", entityID, e.EntityType)
	}
	if e.Attrs == nil {
		e.Attrs = map[string]interface{}{}
	}
	pv := asFloat(e.Attrs["planned_value"])
	ev := asFloat(e.Attrs["earned_value"])
	ac := asFloat(e.Attrs["actual_cost"])
	bac := asFloat(e.Attrs["bac"])
	if bac == 0 {
		bac = pv // BAC 默认等于 PV（总预算）
	}
	r := EVM(EVMInput{BAC: bac, PV: pv, EV: ev, AC: ac})
	e.Attrs["bac"] = bac
	e.Attrs["cv"] = round2(r.CV)
	e.Attrs["sv"] = round2(r.SV)
	e.Attrs["cpi"] = round2(r.CPI)
	e.Attrs["spi"] = round2(r.SPI)
	e.Attrs["eac"] = round2(r.EAC)
	e.Attrs["etc"] = round2(r.ETC)
	e.Attrs["vac"] = round2(r.VAC)
	e.Attrs["tcpi"] = round2(r.TCPI)
	_ = eavtypes.Entity{}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, *e)
}

func asFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case uint:
		return float64(n)
	case string:
		f := 0.0
		fmt.Sscanf(n, "%f", &f)
		return f
	}
	return 0
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
