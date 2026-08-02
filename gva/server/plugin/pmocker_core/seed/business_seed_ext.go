package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	pmockerSvc "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"gorm.io/gorm"
)

// extendAllProjectsData 对所有项目派生扩展数据（V2-4-b/c/d/e/g）
func extendAllProjectsData(_ context.Context, rt *runtimeCtx) error {
	for _, proj := range rt.allProjects {
		if err := extendProjectData(rt, proj); err != nil {
			return fmt.Errorf("扩展项目 %s 数据: %w", proj.Title, err)
		}
	}
	// V2-4-f: 归档 PROJ_A（最接近完成的项目）+ 生成结项报告快照
	if err := archiveOneProject(rt); err != nil {
		return fmt.Errorf("归档项目失败: %w", err)
	}
	return nil
}

// extendProjectData 单项目扩展：milestone/baselines/time_entries/cost_actuals/workflows/relations/reports
func extendProjectData(rt *runtimeCtx, proj *pmocker.PMEntity) error {
	db := rt.db
	projectID := proj.ID

	// 1. 里程碑（每项目 5 个，从关键任务派生）
	if err := createMilestones(db, projectID, rt.leader); err != nil {
		return fmt.Errorf("里程碑: %w", err)
	}

	// 2. 基线（每项目 schedule + cost，共 6 个）
	blSvc := &pmockerSvc.BaselineService{}
	if _, err := blSvc.CreateBaseline(projectID, "schedule", nil, rt.leader); err != nil {
		return fmt.Errorf("计划基线: %w", err)
	}
	if _, err := blSvc.CreateBaseline(projectID, "cost", nil, rt.leader); err != nil {
		return fmt.Errorf("成本基线: %w", err)
	}

	// 3. 工时登记（每项目 ≥10，从已完成任务派生）
	if err := createTimeEntries(db, projectID, rt); err != nil {
		return fmt.Errorf("工时: %w", err)
	}

	// 4. 实际成本（每项目 ≥10，从 cost_item 派生）
	if err := createCostActuals(db, projectID, rt.leader); err != nil {
		return fmt.Errorf("实际成本: %w", err)
	}

	// 5. 工作流实例 + 审批记录（每项目 2 个）
	if err := createWorkflowInstances(db, projectID, rt); err != nil {
		return fmt.Errorf("工作流: %w", err)
	}

	// 6. 跨模块关联（每项目 5 个）
	if err := createRelations(db, projectID); err != nil {
		return fmt.Errorf("关联: %w", err)
	}

	// 7. 报告快照（dashboard×2 + pmo×1）
	rptSvc := &pmockerSvc.ReportService{}
	if err := rptSvc.GenerateReportSnapshot(projectID, "dashboard", "2026-06", rt.leader); err != nil {
		return fmt.Errorf("仪表盘快照(2026-06): %w", err)
	}
	if err := rptSvc.GenerateReportSnapshot(projectID, "dashboard", "2026-07", rt.leader); err != nil {
		return fmt.Errorf("仪表盘快照(2026-07): %w", err)
	}
	if err := rptSvc.GenerateReportSnapshot(projectID, "pmo", "2026-07", rt.leader); err != nil {
		return fmt.Errorf("PMO看板快照: %w", err)
	}
	return nil
}

// createMilestones 从任务列表派生 5 个里程碑（第一项/最后项/中间3项关键任务）
func createMilestones(db *gorm.DB, projectID uint, leader uint) error {
	var tasks []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "task").Order("id ASC").Find(&tasks)
	if len(tasks) == 0 {
		return nil
	}

	// 选 5 个索引：第一/最后/25%/50%/75%
	idxSet := map[int]bool{0: true, len(tasks) - 1: true}
	for _, pct := range []int{25, 50, 75} {
		idx := pct * len(tasks) / 100
		if idx > 0 && idx < len(tasks)-1 {
			idxSet[idx] = true
		}
	}

	for idx := range idxSet {
		t := tasks[idx]
		endDate := readAttrStr(db, t.ID, "end_date")
		m := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: "milestone",
			Title:      t.Title + "【里程碑】",
			Status:     t.Status,
			OwnerID:    t.OwnerID,
			Priority:   1,
			CreatedBy:  leader,
		}
		db.Create(m)
		createAttrStr(db, m.ID, "end_date", endDate)
		createAttrStr(db, m.ID, "milestone_type", "key")
	}
	return nil
}

// createTimeEntries 从已完成任务派生工时登记（每项目 ≥10）
// 注意：progress 字段在 pm_attrs 表而非 pm_entities 主表，需在循环中通过 readAttrInt 读取
func createTimeEntries(db *gorm.DB, projectID uint, rt *runtimeCtx) error {
	var tasks []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks)
	if len(tasks) == 0 {
		return nil
	}

	// 项目团队成员列表
	var members []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "team_member").Find(&members)
	if len(members) == 0 {
		return nil
	}

	count := 0
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, t := range tasks {
		estHours := readAttrDec(db, t.ID, "estimated_hours")
		progress := readAttrInt(db, t.ID, "progress")
		if estHours <= 0 || progress <= 0 {
			continue
		}

		// 任务负责人对应的 team_member 实体
		var member pmocker.PMEntity
		if t.OwnerID != nil {
			db.Where("project_id = ? AND entity_type = ? AND owner_id = ?", projectID, "team_member", *t.OwnerID).First(&member)
		}
		if member.ID == 0 {
			member = members[0]
		}

		rate := readAttrDec(db, member.ID, "hourly_rate")
		if rate <= 0 {
			rate = 100.0
		}

		// 按进度比例登记工时
		hours := estHours * float64(progress) / 100.0
		if hours < 4 {
			hours = 4
		}
		endDate := readAttrStr(db, t.ID, "end_date")
		if endDate == "" {
			endDate = "2026-06-15"
		}

		te := &pmocker.PMTimeEntry{
			ProjectID:   projectID,
			TaskID:      t.ID,
			MemberID:    member.ID,
			UserID:      rt.leader,
			Date:        endDate,
			Hours:       hours,
			HourlyRate:  rate,
			Cost:        hours * rate,
			Description: t.Title + " 进度工时",
			Status:      "approved",
			ApproverID:  &rt.leader,
			ApprovedAt:  &now,
		}
		db.Create(te)
		count++

		// 若任务工时较大，拆分为 2 条记录（更贴近真实场景）
		if estHours >= 80 && count < 15 {
			te2 := &pmocker.PMTimeEntry{
				ProjectID:   projectID,
				TaskID:      t.ID,
				MemberID:    member.ID,
				UserID:      rt.leader,
				Date:        endDate,
				Hours:       hours * 0.3,
				HourlyRate:  rate,
				Cost:        hours * 0.3 * rate,
				Description: t.Title + " 加班工时",
				Status:      "approved",
				ApproverID:  &rt.leader,
				ApprovedAt:  &now,
			}
			db.Create(te2)
			count++
		}
		if count >= 12 {
			break
		}
	}
	return nil
}

// createCostActuals 从 cost_item 派生实际成本（每项目 ≥10）
func createCostActuals(db *gorm.DB, projectID uint, leader uint) error {
	var costItems []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "cost_item").Find(&costItems)
	if len(costItems) == 0 {
		return nil
	}

	// 项目进度比例（用于派生实际成本占预算比例）
	var tasks []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks)
	totalProgress := 0
	for _, t := range tasks {
		totalProgress += int(readAttrInt(db, t.ID, "progress"))
	}
	progressPct := 50
	if len(tasks) > 0 {
		progressPct = totalProgress / len(tasks)
	}

	for _, c := range costItems {
		pv := readAttrDec(db, c.ID, "planned_value")
		if pv <= 0 {
			continue
		}
		costType := readAttrStr(db, c.ID, "cost_type")
		if costType == "" {
			costType = "other"
		}
		actual := pv * float64(progressPct) / 100.0
		ca := &pmocker.PMCostActual{
			ProjectID:   projectID,
			CostItemID:  &c.ID,
			CostType:    costType,
			Amount:      actual,
			Date:        "2026-06-30",
			Source:      "manual",
			Description: c.Title + " 实际执行",
			Status:      "confirmed",
		}
		db.Create(&ca)
	}
	return nil
}

// createWorkflowInstances 创建 2 个工作流实例 + 2 个审批记录
func createWorkflowInstances(db *gorm.DB, projectID uint, rt *runtimeCtx) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	// 查 PMO 审批人（pmo01）
	pmoID := rt.userIdMap["pmo01"]
	if pmoID == 0 {
		pmoID = rt.leader
	}
	pmoName := rt.userNick["pmo01"]
	if pmoName == "" {
		pmoName = "PMO"
	}

	wfDefs := []struct {
		code, name, node string
	}{
		{"schedule_baseline", "计划基线审批", "pm_approval"},
		{"cost_baseline", "成本基线审批", "fm_approval"},
	}

	for _, w := range wfDefs {
		// 工作流实例（已完成）
		inst := &pmocker.PMWorkflowInstance{
			EntityID:     projectID,
			WorkflowCode: w.code,
			CurrentNode:  "end",
			Status:       "completed",
		}
		if err := db.Create(inst).Error; err != nil {
			return err
		}

		// 审批记录
		sig := fmt.Sprintf("%s-%d-%s", pmoName, pmoID, now)
		rec := &pmocker.PMApprovalRecord{
			ProjectID:      projectID,
			EntityID:       projectID,
			EntityType:     "eps_node",
			WorkflowInstID: &inst.ID,
			NodeName:       w.node,
			ApproverID:     pmoID,
			ApproverName:   pmoName,
			Action:         "approve",
			Comment:        w.name + " 通过",
			Signature:      sig,
			ActedAt:        now,
		}
		if err := db.Create(rec).Error; err != nil {
			return err
		}
	}
	return nil
}

// createRelations 创建 5 个跨模块关联
func createRelations(db *gorm.DB, projectID uint) error {
	// 查找各模块首条实体
	firstEntity := func(entityType string) uint {
		var e pmocker.PMEntity
		db.Where("project_id = ? AND entity_type = ?", projectID, entityType).Order("id ASC").First(&e)
		return e.ID
	}

	reqID := firstEntity("requirement")
	taskID := firstEntity("task")
	issueID := firstEntity("issue")
	riskID := firstEntity("risk")
	changeID := firstEntity("change_request")
	deliverableID := firstEntity("deliverable")
	baselineID := firstEntity("baseline")
	_ = baselineID // baseline 是单独表，不在 pm_entities

	// 查 pm_baselines 表首个基线
	var bl pmocker.PMBaseline
	db.Where("project_id = ?", projectID).Order("id ASC").First(&bl)
	var baselineEntityID uint
	if bl.ID > 0 {
		// 基线作为虚拟实体，用项目 ID 关联
		baselineEntityID = projectID
	}

	rels := []pmocker.PMRelation{}
	if reqID > 0 && taskID > 0 {
		rels = append(rels, pmocker.PMRelation{SrcID: reqID, DstID: taskID, RelationType: "decomposes"})
	}
	if issueID > 0 && taskID > 0 {
		rels = append(rels, pmocker.PMRelation{SrcID: issueID, DstID: taskID, RelationType: "relates_to"})
	}
	if changeID > 0 && baselineEntityID > 0 {
		rels = append(rels, pmocker.PMRelation{SrcID: changeID, DstID: baselineEntityID, RelationType: "changes"})
	}
	if riskID > 0 && taskID > 0 {
		rels = append(rels, pmocker.PMRelation{SrcID: riskID, DstID: taskID, RelationType: "impacts"})
	}
	if deliverableID > 0 && taskID > 0 {
		rels = append(rels, pmocker.PMRelation{SrcID: deliverableID, DstID: taskID, RelationType: "delivers"})
	}

	for _, r := range rels {
		if err := db.Create(&r).Error; err != nil {
			return err
		}
	}
	return nil
}

// archiveOneProject 归档 PROJ_A + 生成结项报告快照（V2-4-f）
// 直接 SQL 标记归档（绕过 validateArchivable 严格校验），同时生成 close 报告快照
func archiveOneProject(rt *runtimeCtx) error {
	db := rt.db
	if len(rt.allProjects) == 0 {
		return nil
	}

	// 选第一个项目（PROJ_A）归档
	proj := rt.allProjects[0]
	projectID := proj.ID

	// 1. 直接更新项目状态为 archived
	now := time.Now().Format("2006-01-02 15:04:05")
	if err := db.Model(&pmocker.PMEntity{}).Where("id = ?", projectID).
		Updates(map[string]interface{}{"status": "archived", "updated_at": now}).Error; err != nil {
		return fmt.Errorf("更新项目状态: %w", err)
	}

	// 2. 关联实体标记 archived
	entityTypes := []string{"task", "issue", "risk", "requirement", "change_request", "deliverable",
		"team_member", "cost_item", "scope_item", "milestone"}
	for _, et := range entityTypes {
		db.Model(&pmocker.PMEntity{}).
			Where("project_id = ? AND entity_type = ? AND status != ?", projectID, et, "archived").
			Update("status", "archived")
	}

	// 3. 生成结项报告快照
	rptSvc := &pmockerSvc.ReportService{}
	if err := rptSvc.GenerateReportSnapshot(projectID, "close", "close", rt.leader); err != nil {
		return fmt.Errorf("结项报告快照: %w", err)
	}

	// 4. 更新内存中的项目状态
	proj.Status = "archived"
	return nil
}

// 内部 helper：从 attrs 读字符串/小数/整数（避免依赖 service 包未导出函数）
func readAttrStr(db *gorm.DB, entityID uint, key string) string {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValDate != nil {
		return *attr.ValDate
	}
	if attr.ValDateTime != nil {
		return *attr.ValDateTime
	}
	return ""
}

func readAttrDec(db *gorm.DB, entityID uint, key string) float64 {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValDecimal != nil {
		return *attr.ValDecimal
	}
	return 0
}

func readAttrInt(db *gorm.DB, entityID uint, key string) int64 {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValInt != nil {
		return *attr.ValInt
	}
	return 0
}
