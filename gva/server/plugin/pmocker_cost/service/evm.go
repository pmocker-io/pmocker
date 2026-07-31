package cost

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
