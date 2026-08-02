package schedule

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

// Task 用于 CPM 计算的任务
type Task struct {
	ID           uint      `json:"id"`
	Code         string    `json:"code"`
	Duration     int       `json:"duration"`     // 工期(天)
	Predecessors []TaskDep `json:"predecessors"` // 前置任务
}

// TaskDep 任务依赖
type TaskDep struct {
	TaskID   uint   `json:"taskId"`
	LinkType string `json:"linkType"` // FS/SS/FF/SF
	Lag      int    `json:"lag"`      // 滞后天数
}

// CPMResult CPM 计算结果
type CPMResult struct {
	TaskID      uint   `json:"taskId"`
	EarlyStart  int    `json:"earlyStart"`  // 最早开始(天)
	EarlyFinish int    `json:"earlyFinish"` // 最早完成
	LateStart   int    `json:"lateStart"`   // 最晚开始
	LateFinish  int    `json:"lateFinish"`  // 最晚完成
	TotalFloat  int    `json:"totalFloat"`  // 总浮动
	FreeFloat   int    `json:"freeFloat"`   // 自由浮动
	IsCritical  bool   `json:"isCritical"`  // 是否关键路径
}

// CPM 计算关键路径（简化版，仅支持 FS 关系）
// 步骤：① 拓扑排序；② 正向计算 ES/EF；③ 反向计算 LF/LS；④ 计算浮动
func CPM(tasks []Task) ([]CPMResult, error) {
	if len(tasks) == 0 {
		return nil, nil
	}
	// 构建邻接表
	idx := make(map[uint]int, len(tasks))
	for i, t := range tasks {
		idx[t.ID] = i
	}
	// 校验前置引用存在
	for _, t := range tasks {
		for _, d := range t.Predecessors {
			if _, ok := idx[d.TaskID]; !ok {
				return nil, fmt.Errorf("task %d: predecessor %d not in task list", t.ID, d.TaskID)
			}
		}
	}
	// 拓扑排序（Kahn 算法）
	indeg := make([]int, len(tasks))
	adj := make([][]int, len(tasks))
	for i, t := range tasks {
		for _, d := range t.Predecessors {
			adj[idx[d.TaskID]] = append(adj[idx[d.TaskID]], i)
			indeg[i]++
		}
	}
	queue := []int{}
	for i := range tasks {
		if indeg[i] == 0 {
			queue = append(queue, i)
		}
	}
	topo := []int{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		topo = append(topo, u)
		for _, v := range adj[u] {
			indeg[v]--
			if indeg[v] == 0 {
				queue = append(queue, v)
			}
		}
	}
	if len(topo) != len(tasks) {
		return nil, fmt.Errorf("cycle detected in task dependencies")
	}
	// 正向计算 ES/EF
	es := make([]int, len(tasks))
	ef := make([]int, len(tasks))
	for _, u := range topo {
		t := tasks[u]
		es[u] = 0
		for _, d := range t.Predecessors {
			predIdx := idx[d.TaskID]
			// 仅支持 FS 关系（前置完成→后置开始）
			finish := ef[predIdx] + d.Lag
			if finish > es[u] {
				es[u] = finish
			}
		}
		ef[u] = es[u] + t.Duration
	}
	// 项目总工期
	projectEnd := 0
	for _, f := range ef {
		if f > projectEnd {
			projectEnd = f
		}
	}
	// 反向计算 LF/LS
	// LF[u] = min(后置任务 v 的 ES[v] - lag)（FS 关系：前置完成才能后置开始）
	// 无后继任务时 LF = projectEnd
	lf := make([]int, len(tasks))
	ls := make([]int, len(tasks))
	for i := range tasks {
		lf[i] = projectEnd
	}
	for i := len(topo) - 1; i >= 0; i-- {
		u := topo[i]
		ls[u] = lf[u] - tasks[u].Duration
		for _, v := range adj[u] {
			// 后置任务的 ES - Lag = 当前任务的 LF（FS 关系）
			lag := 0
			for _, d := range tasks[v].Predecessors {
				if idx[d.TaskID] == u {
					lag = d.Lag
					break
				}
			}
			if es[v]-lag < lf[u] {
				lf[u] = es[v] - lag
				ls[u] = lf[u] - tasks[u].Duration
			}
		}
	}
	// 计算结果
	results := make([]CPMResult, len(tasks))
	for i, t := range tasks {
		total := lf[i] - ef[i]
		// 自由浮动：min(后置 ES) - 当前 EF
		ff := projectEnd
		if len(adj[i]) == 0 {
			ff = 0
		} else {
			for _, v := range adj[i] {
				lag := 0
				for _, d := range tasks[v].Predecessors {
					if d.TaskID == t.ID {
						lag = d.Lag
						break
					}
				}
				if es[v]-lag-ef[i] < ff {
					ff = es[v] - lag - ef[i]
				}
			}
		}
		if ff < 0 {
			ff = 0
		}
		results[i] = CPMResult{
			TaskID:      t.ID,
			EarlyStart:  es[i],
			EarlyFinish: ef[i],
			LateStart:   ls[i],
			LateFinish:  lf[i],
			TotalFloat:  total,
			FreeFloat:   ff,
			IsCritical:  total == 0,
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].EarlyStart < results[j].EarlyStart
	})
	return results, nil
}

// FormatDate 将天数偏移转为日期（基于 baseDate）
func FormatDate(baseDate time.Time, days int) string {
	return baseDate.AddDate(0, 0, days).Format("2006-01-02")
}

// CPMHandler 工作流 auto 节点处理器：读取项目下所有 task，执行 CPM，写回关键路径/浮动等派生字段。
// handler 名：pmocker.schedule.cpm
func CPMHandler(ctx context.Context, taskEntityID uint) error {
	// 通过 task 实体反查 ProjectID
	root, err := pmservice.ServiceGroupApp.GetEntity(ctx, taskEntityID)
	if err != nil {
		return fmt.Errorf("get task entity %d: %w", taskEntityID, err)
	}
	if root.EntityType != "task" {
		return fmt.Errorf("entity %d is %q, expect task", taskEntityID, root.EntityType)
	}
	projectID := root.ProjectID
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "task", 0, 10000)
	if err != nil {
		return fmt.Errorf("list tasks for project %d: %w", projectID, err)
	}
	// 构造 CPM 输入
	tasks := make([]Task, 0, len(entities))
	for _, e := range entities {
		t := Task{ID: e.ID, Code: e.Title}
		if d, ok := e.Attrs["duration"]; ok {
			t.Duration = asInt(d)
		}
		if t.Duration <= 0 {
			t.Duration = 1
		}
		if preds, ok := e.Attrs["predecessors"]; ok {
			if b, mErr := json.Marshal(preds); mErr == nil {
				_ = json.Unmarshal(b, &t.Predecessors)
			} else if arr, ok := preds.([]interface{}); ok {
				for _, it := range arr {
					if m, ok := it.(map[string]interface{}); ok {
						dep := TaskDep{
							TaskID:   uint(asInt(m["taskId"])),
							LinkType: asStr(m["linkType"]),
							Lag:      asInt(m["lag"]),
						}
						if dep.LinkType == "" {
							dep.LinkType = "FS"
						}
						t.Predecessors = append(t.Predecessors, dep)
					}
				}
			}
		}
		tasks = append(tasks, t)
	}
	results, err := CPM(tasks)
	if err != nil {
		return fmt.Errorf("cpm compute: %w", err)
	}
	// 按 TaskID 回写字段
	resByID := make(map[uint]CPMResult, len(results))
	for _, r := range results {
		resByID[r.TaskID] = r
	}
	for _, e := range entities {
		r, ok := resByID[e.ID]
		if !ok {
			continue
		}
		if e.Attrs == nil {
			e.Attrs = map[string]interface{}{}
		}
		e.Attrs["is_critical_path"] = r.IsCritical
		e.Attrs["total_float"] = float64(r.TotalFloat)
		e.Attrs["free_float"] = float64(r.FreeFloat)
		// 若任务有 start_date，基于偏移推导 earliest/latest 日期
		if sd, ok := e.Attrs["start_date"].(string); ok && sd != "" {
			if base, pErr := time.Parse("2006-01-02", sd); pErr == nil {
				e.Attrs["baseline_start"] = FormatDate(base, r.EarlyStart)
				e.Attrs["baseline_finish"] = FormatDate(base, r.EarlyFinish)
			}
		}
		if upErr := pmservice.ServiceGroupApp.UpdateEntity(ctx, e); upErr != nil {
			return fmt.Errorf("update task %d: %w", e.ID, upErr)
		}
	}
	return nil
}

func asInt(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case uint:
		return int(n)
	case float64:
		return int(n)
	case string:
		x := 0
		fmt.Sscanf(n, "%d", &x)
		return x
	}
	return 0
}

func asStr(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
