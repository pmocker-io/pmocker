package pmocker

import (
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/internal/testutil"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// TestReadAttrDecimalCompatibleWithValInt 验证 readAttrDecimal 能读取 val_int 存储的整数值。
// 进度 progress 等整数值由 seed/writeAttrValue 写入 val_int 列，而进度算法按 decimal 读取。
func TestReadAttrDecimalCompatibleWithValInt(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMAttr{})

	eid := uint(1)
	i := int64(100)
	if err := db.Create(&pmocker.PMAttr{EntityID: eid, FieldKey: "progress", ValInt: &i}).Error; err != nil {
		t.Fatal(err)
	}
	if got := readAttrDecimal(eid, "progress"); got != 100 {
		t.Fatalf("readAttrDecimal(progress) = %v, want 100 (val_int 整数存储应被兼容读取)", got)
	}

	d := 70.5
	if err := db.Create(&pmocker.PMAttr{EntityID: eid, FieldKey: "ratio", ValDecimal: &d}).Error; err != nil {
		t.Fatal(err)
	}
	if got := readAttrDecimal(eid, "ratio"); got != 70.5 {
		t.Fatalf("readAttrDecimal(ratio) = %v, want 70.5 (val_decimal 小数存储)", got)
	}

	// 不存在的字段返回 0
	if got := readAttrDecimal(eid, "missing"); got != 0 {
		t.Fatalf("readAttrDecimal(missing) = %v, want 0", got)
	}
}

// TestCalcProjectProgressWBSFallsBackToHours 验证 wbs 算法在无 pm_wbs_nodes 时回退工时算法。
func TestCalcProjectProgressWBSFallsBackToHours(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntity{}, &pmocker.PMAttr{}, &pmocker.PMWBSNode{})

	// 项目指定 wbs 算法
	p := &pmocker.PMEntity{ProjectID: 0, EntityType: "eps_node", Title: "P", Status: "active"}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	if err := writeAttrString(p.ID, "progress_algo", "wbs"); err != nil {
		t.Fatal(err)
	}

	// 2 个任务：hours=100 progress=100, hours=100 progress=50
	for _, tt := range []struct {
		title string
		h, pr float64
	}{
		{"t1", 100, 100},
		{"t2", 100, 50},
	} {
		tk := &pmocker.PMEntity{ProjectID: p.ID, EntityType: "task", Title: tt.title, Status: "in_progress"}
		if err := db.Create(tk).Error; err != nil {
			t.Fatal(err)
		}
		if err := writeAttrDecimal(tk.ID, "estimated_hours", tt.h); err != nil {
			t.Fatal(err)
		}
		if err := writeAttrDecimal(tk.ID, "progress", tt.pr); err != nil {
			t.Fatal(err)
		}
	}

	s := &ProgressService{}
	got, err := s.CalcProjectProgress(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	// 工时加权：(100*100/100 + 100*50/100) / 200 * 100 = 75
	if got != 75 {
		t.Fatalf("CalcProjectProgress(wbs fallback) = %v, want 75", got)
	}
}

func TestCalcByCountCompletedStatus(t *testing.T) {
	db := testutil.NewMemoryDB(t, &pmocker.PMEntity{})

	// 3 个任务：2 completed + 1 planned
	for i, st := range []string{"completed", "completed", "planned"} {
		if err := db.Create(&pmocker.PMEntity{
			ProjectID: 1, EntityType: "task", Title: "t" + string(rune('0'+i)), Status: st,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}

	s := &ProgressService{}
	got, err := s.CalcByCount(1)
	if err != nil {
		t.Fatal(err)
	}
	if got != 66.66666666666666 && got != 66.66666666666667 {
		t.Fatalf("CalcByCount = %v, want ~66.67", got)
	}
}
