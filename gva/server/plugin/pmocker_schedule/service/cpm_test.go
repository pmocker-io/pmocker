package schedule

import "testing"

func TestCPMSimpleChain(t *testing.T) {
	// A(3) -> B(2) -> C(4)，关键路径 A→B→C，总工期 9
	tasks := []Task{
		{ID: 1, Code: "A", Duration: 3},
		{ID: 2, Code: "B", Duration: 2, Predecessors: []TaskDep{{TaskID: 1, LinkType: "FS"}}},
		{ID: 3, Code: "C", Duration: 4, Predecessors: []TaskDep{{TaskID: 2, LinkType: "FS"}}},
	}
	res, err := CPM(tasks)
	if err != nil {
		t.Fatalf("CPM: %v", err)
	}
	if len(res) != 3 {
		t.Fatalf("expected 3 results, got %d", len(res))
	}
	// 所有任务浮动为 0
	for _, r := range res {
		if r.TotalFloat != 0 {
			t.Errorf("task %d total float = %d, want 0", r.TaskID, r.TotalFloat)
		}
		if !r.IsCritical {
			t.Errorf("task %d should be critical", r.TaskID)
		}
	}
	// 最早完成日期：A=3, B=5, C=9
	if res[0].EarlyFinish != 3 {
		t.Errorf("A EF = %d", res[0].EarlyFinish)
	}
	if res[1].EarlyFinish != 5 {
		t.Errorf("B EF = %d", res[1].EarlyFinish)
	}
	if res[2].EarlyFinish != 9 {
		t.Errorf("C EF = %d", res[2].EarlyFinish)
	}
}

func TestCPMParallel(t *testing.T) {
	// A(2) 同时启动 B(3) 和 C(1)，关键路径 A→B
	tasks := []Task{
		{ID: 1, Code: "A", Duration: 2},
		{ID: 2, Code: "B", Duration: 3, Predecessors: []TaskDep{{TaskID: 1, LinkType: "FS"}}},
		{ID: 3, Code: "C", Duration: 1, Predecessors: []TaskDep{{TaskID: 1, LinkType: "FS"}}},
	}
	res, err := CPM(tasks)
	if err != nil {
		t.Fatalf("CPM: %v", err)
	}
	// C 总浮动 = 2（项目工期 5 - A→C 路径工期 3）
	var cRes *CPMResult
	for i := range res {
		if res[i].TaskID == 3 {
			cRes = &res[i]
		}
	}
	if cRes == nil {
		t.Fatal("task C not found")
	}
	if cRes.TotalFloat != 2 {
		t.Errorf("C total float = %d, want 2", cRes.TotalFloat)
	}
	if cRes.IsCritical {
		t.Errorf("C should not be critical")
	}
}

func TestCPMCycle(t *testing.T) {
	tasks := []Task{
		{ID: 1, Code: "A", Duration: 2, Predecessors: []TaskDep{{TaskID: 2, LinkType: "FS"}}},
		{ID: 2, Code: "B", Duration: 2, Predecessors: []TaskDep{{TaskID: 1, LinkType: "FS"}}},
	}
	if _, err := CPM(tasks); err == nil {
		t.Error("expected cycle error, got nil")
	}
}
