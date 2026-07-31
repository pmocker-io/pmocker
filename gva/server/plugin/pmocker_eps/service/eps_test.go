package eps

import (
	"testing"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

func TestValidateCycle(t *testing.T) {
	svc := new(Service)

	cases := []struct {
		oldPath string
		newPath string
		wantErr bool
	}{
		{"/集团总部", "/集团总部/技术中心", true},
		{"/集团总部/技术中心", "/集团总部", false},
		{"/集团总部", "/集团总部", true},
		{"/集团总部/技术中心", "/集团总部/行政中心", false},
		{"/A/B/C", "/A/B", false},
		{"/A/B", "/A/B/C/D", true},
	}
	for _, c := range cases {
		err := svc.ValidateCycle(nil, 1, c.oldPath, c.newPath)
		hasErr := err != nil
		if hasErr != c.wantErr {
			t.Errorf("ValidateCycle(%s, %s) err=%v, wantErr=%v", c.oldPath, c.newPath, err, c.wantErr)
		}
	}
}

func TestBuildTreeFromEntities(t *testing.T) {
	type A = map[string]interface{}
	entities := []eavtypes.Entity{
		{ID: 1, Title: "集团总部", EntityType: "eps_node", Attrs: A{"parent_path": "/", "name": "集团总部", "code": "G1", "type": "group"}},
		{ID: 2, Title: "技术中心", EntityType: "eps_node", Attrs: A{"parent_path": "/集团总部", "name": "技术中心", "code": "T1", "type": "division"}},
		{ID: 3, Title: "研发部", EntityType: "eps_node", Attrs: A{"parent_path": "/集团总部/技术中心", "name": "研发部", "code": "R1", "type": "department"}},
		{ID: 4, Title: "行政中心", EntityType: "eps_node", Attrs: A{"parent_path": "/集团总部", "name": "行政中心", "code": "A1", "type": "division"}},
	}

	roots := buildTreeFromEntities(entities)
	if len(roots) != 1 {
		t.Fatalf("root count = %d, want 1", len(roots))
	}
	if roots[0].Name != "集团总部" || roots[0].ID != 1 {
		t.Errorf("root name/id = %s/%d, want 集团总部/1", roots[0].Name, roots[0].ID)
	}
	if len(roots[0].Children) != 2 {
		t.Fatalf("root children count = %d, want 2", len(roots[0].Children))
	}

	var tech, admin TreeNode
	for _, c := range roots[0].Children {
		if c.Name == "技术中心" {
			tech = c
		} else if c.Name == "行政中心" {
			admin = c
		}
	}
	if tech.ID != 2 {
		t.Errorf("tech id = %d, want 2", tech.ID)
	}
	if admin.ID != 4 {
		t.Errorf("admin id = %d, want 4", admin.ID)
	}
	if len(tech.Children) != 1 || tech.Children[0].ID != 3 {
		t.Errorf("tech children count=%d firstID=%v, want count=1 id=3", len(tech.Children), func() interface{} {
			if len(tech.Children) > 0 {
				return tech.Children[0].ID
			}
			return nil
		}())
	}
}
