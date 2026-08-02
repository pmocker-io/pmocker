package pmocker

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ProgressService struct{}

func (s *ProgressService) CalcByHours(projectID uint) (float64, error) {
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	totalH, doneH := 0.0, 0.0
	for _, t := range tasks {
		h := readAttrDecimal(t.ID, "estimated_hours")
		p := readAttrDecimal(t.ID, "progress")
		totalH += h
		doneH += h * p / 100.0
	}
	if totalH == 0 {
		return 0, nil
	}
	return doneH / totalH * 100, nil
}

func (s *ProgressService) CalcByWBS(projectID uint) (float64, error) {
	var nodes []pmocker.PMWBSNode
	if err := global.GVA_DB.Where("project_id = ?", projectID).Find(&nodes).Error; err != nil {
		return 0, err
	}
	if len(nodes) == 0 {
		return 0, nil
	}
	childrenMap := map[uint][]pmocker.PMWBSNode{}
	var roots []pmocker.PMWBSNode
	for _, n := range nodes {
		if n.ParentID != nil {
			childrenMap[*n.ParentID] = append(childrenMap[*n.ParentID], n)
		} else {
			roots = append(roots, n)
		}
	}
	if len(roots) == 0 {
		return 0, nil
	}

	var calc func(n pmocker.PMWBSNode) float64
	calc = func(n pmocker.PMWBSNode) float64 {
		kids := childrenMap[n.ID]
		if len(kids) == 0 {
			var tasks []pmocker.PMEntity
			global.GVA_DB.Where("parent_id = ? AND entity_type = ?", n.EntityID, "task").Find(&tasks)
			if len(tasks) == 0 {
				return 0
			}
			sum := 0.0
			for _, t := range tasks {
				sum += readAttrDecimal(t.ID, "progress")
			}
			return sum / float64(len(tasks))
		}
		totalWeight, weighted := 0.0, 0.0
		for _, k := range kids {
			w := readAttrDecimal(k.EntityID, "weight")
			if w <= 0 {
				w = 1.0
			}
			weighted += calc(k) * w
			totalWeight += w
		}
		if totalWeight == 0 {
			return 0
		}
		return weighted / totalWeight
	}
	sum, cnt := 0.0, 0.0
	for _, r := range roots {
		sum += calc(r)
		cnt++
	}
	if cnt == 0 {
		return 0, nil
	}
	return sum / cnt, nil
}

func (s *ProgressService) CalcByCount(projectID uint) (float64, error) {
	var tasks []pmocker.PMEntity
	if err := global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks).Error; err != nil {
		return 0, err
	}
	if len(tasks) == 0 {
		return 0, nil
	}
	done := 0
	for _, t := range tasks {
		if t.Status == "done" {
			done++
		}
	}
	return float64(done) / float64(len(tasks)) * 100, nil
}

func (s *ProgressService) CalcProjectProgress(projectID uint) (float64, error) {
	algo := readAttrString(projectID, "progress_algo")
	if algo == "" {
		algo = "hours"
	}
	switch algo {
	case "hours":
		return s.CalcByHours(projectID)
	case "wbs":
		return s.CalcByWBS(projectID)
	case "count":
		return s.CalcByCount(projectID)
	default:
		return s.CalcByHours(projectID)
	}
}

func (s *ProgressService) CalcHealthStatus(projectID uint) (string, error) {
	rpt, err := (&VarianceService{}).CalcVariance(projectID)
	if err != nil {
		return "green", err
	}
	spi, cpi := 1.0, 1.0
	if rpt != nil {
		if rpt.PV > 0 {
			spi = rpt.SPI
		}
		if rpt.AC > 0 {
			cpi = rpt.CPI
		}
	}
	var risks []pmocker.PMEntity
	global.GVA_DB.Where("project_id = ? AND entity_type = ?", projectID, "risk").Find(&risks)
	highRisks := 0
	for _, r := range risks {
		score := readAttrDecimal(r.ID, "score")
		if score >= 0.5 {
			highRisks++
		}
	}
	if spi < 0.7 || cpi < 0.7 || highRisks >= 3 {
		return "red", nil
	}
	if spi < 0.9 || cpi < 0.9 || highRisks >= 1 {
		return "yellow", nil
	}
	return "green", nil
}

func (s *ProgressService) GetProjectAlgo(projectID uint) string {
	algo := readAttrString(projectID, "progress_algo")
	if algo == "" {
		return "hours"
	}
	return algo
}

// TaskCompleteHook 任务完成 → 刷新项目进度 + 健康度
type TaskCompleteHook struct{}

func (h *TaskCompleteHook) OnEnter(ctx context.Context, entityID uint, nodeName string) error {
	return nil
}
func (h *TaskCompleteHook) OnLeave(ctx context.Context, entityID uint, nodeName string, action string) error {
	if action != "complete" && action != "approve" && action != "done" {
		return nil
	}
	var task pmocker.PMEntity
	if err := global.GVA_DB.First(&task, entityID).Error; err != nil {
		return err
	}
	if task.ID == 0 || task.ProjectID == 0 {
		return nil
	}
	projectID := task.ProjectID
	percent, err := (&ProgressService{}).CalcProjectProgress(projectID)
	if err != nil {
		return fmt.Errorf("calc project progress failed: %w", err)
	}
	if err := writeAttrDecimal(projectID, "progress", percent); err != nil {
		return err
	}
	health, _ := (&ProgressService{}).CalcHealthStatus(projectID)
	if err := writeAttrString(projectID, "health_status", health); err != nil {
		return err
	}
	return nil
}
