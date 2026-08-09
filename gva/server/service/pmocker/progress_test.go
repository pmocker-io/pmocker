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

// TestCalcByCountCompletedStatus 验证任务完成度按 completed 状态统计（对齐 schema states）。
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
