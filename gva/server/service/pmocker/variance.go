package pmocker

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type VarianceService struct{}

type VarianceReport struct {
	ProjectID uint    `json:"projectId"`
	PV        float64 `json:"pv"`
	EV        float64 `json:"ev"`
	AC        float64 `json:"ac"`
	SV        float64 `json:"sv"`
	CV        float64 `json:"cv"`
	SPI       float64 `json:"spi"`
	CPI       float64 `json:"cpi"`
	CalcAt    string  `json:"calcAt"`
}

type VarianceAlert struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	EntityID uint   `json:"entityId"`
}

func (s *VarianceService) CalcVariance(projectID uint) (*VarianceReport, error) {
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return nil, err
	}
	pv, ev := 0.0, 0.0
	for _, t := range tasks {
		budget := readAttrDecimal(t.ID, "budget_cost")
		if budget <= 0 {
			hours := readAttrDecimal(t.ID, "estimated_hours")
			if t.OwnerID != nil {
				rate := readAttrDecimal(*t.OwnerID, "hourly_rate")
				budget = hours * rate
			}
		}
		pv += budget
		progress := readAttrDecimal(t.ID, "progress")
		ev += budget * progress / 100.0
	}

	var acStruct = struct{ Total float64 }{}
	global.GVA_DB.Model(&pmocker.PMCostActual{}).
		Where("project_id = ? AND status = ?", projectID, "confirmed").
		Select("COALESCE(SUM(amount),0) as total").Scan(&acStruct)

	rpt := &VarianceReport{
		ProjectID: projectID, PV: pv, EV: ev, AC: acStruct.Total,
		SV: ev - pv, CV: ev - acStruct.Total,
		CalcAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	if pv > 0 {
		rpt.SPI = ev / pv
	}
	if acStruct.Total > 0 {
		rpt.CPI = ev / acStruct.Total
	}
	return rpt, nil
}

func (s *VarianceService) GetAlerts(projectID uint) ([]VarianceAlert, error) {
	alerts := make([]VarianceAlert, 0)
	today := time.Now().Format("2006-01-02")

	var tasks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks)
	for _, t := range tasks {
		if t.Status == "done" {
			continue
		}
		end := readAttrString(t.ID, "end_date")
		if end != "" && end < today {
			alerts = append(alerts, VarianceAlert{
				Type: "overdue", Severity: "critical",
				Title:    "任务超期: " + t.Title,
				Detail:   fmt.Sprintf("计划结束 %s，当前状态 %s", end, t.Status),
				EntityID: t.ID,
			})
		}
	}

	rpt, err := s.CalcVariance(projectID)
	if err == nil && rpt != nil {
		if rpt.PV > 0 && rpt.SPI < 0.9 {
			alerts = append(alerts, VarianceAlert{
				Type: "schedule_off", Severity: severityOf(rpt.SPI),
				Title:  "进度偏差预警",
				Detail: fmt.Sprintf("SPI=%.2f（SV=%.2f），进度落后", rpt.SPI, rpt.SV),
			})
		}
		if rpt.AC > 0 && rpt.CPI < 0.9 {
			alerts = append(alerts, VarianceAlert{
				Type: "over_budget", Severity: severityOf(rpt.CPI),
				Title:  "成本超支预警",
				Detail: fmt.Sprintf("CPI=%.2f（CV=%.2f），实际成本超出挣值", rpt.CPI, rpt.CV),
			})
		}
	}

	var risks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "risk").Find(&risks)
	for _, r := range risks {
		score := readAttrDecimal(r.ID, "score")
		if score >= 0.5 {
			alerts = append(alerts, VarianceAlert{
				Type: "high_risk", Severity: "critical",
				Title:    "高风险: " + r.Title,
				Detail:   readAttrString(r.ID, "response"),
				EntityID: r.ID,
			})
		}
	}
	return alerts, nil
}

func severityOf(idx float64) string {
	if idx < 0.7 {
		return "critical"
	}
	return "warning"
}
