package seed

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultPassword = "Pmocker@2026"

var (
	deptIdMap = make(map[string]uint)
	posIdMap  = make(map[string]uint)
	userIdMap = make(map[string]uint)
)

func SeedOrgStructure(ctx context.Context) error {
	db := global.GVA_DB.WithContext(ctx)

	var deptCount int64
	db.Model(&system.SysDepartment{}).Count(&deptCount)
	if deptCount > 10 {
		return nil
	}

	if err := createDepartments(db); err != nil {
		return fmt.Errorf("createDepartments: %w", err)
	}
	if err := createPositions(db); err != nil {
		return fmt.Errorf("createPositions: %w", err)
	}
	if err := createAuthorities(db); err != nil {
		return fmt.Errorf("createAuthorities: %w", err)
	}
	if err := createUsers(db); err != nil {
		return fmt.Errorf("createUsers: %w", err)
	}
	if err := updateDeptLeaders(db); err != nil {
		return fmt.Errorf("updateDeptLeaders: %w", err)
	}

	return nil
}

type deptDef struct {
	name      string
	parentID  uint
	ancestors string
	leader    string
}

func createDepartments(db *gorm.DB) error {
	depts := []deptDef{
		{"集团总部", 0, "0", "pmo01"},
		{"智能排产系统研发部", 0, "0,1", "proj_a_pm"},
		{"项目管理组", 0, "", "proj_a_pm"},
		{"前端开发组", 0, "", "proj_a_fe"},
		{"后端开发组", 0, "", "proj_a_be"},
		{"质量测试组", 0, "", "proj_a_qa"},
		{"工程建设事业部", 0, "0,2", "proj_b_pm"},
		{"项目管理部", 0, "", "proj_b_pm"},
		{"土建工程部", 0, "", "proj_b_civil"},
		{"机电工程部", 0, "", "proj_b_mep"},
		{"安全造价部", 0, "", "proj_b_safety"},
		{"传感器研发中心", 0, "0,3", "proj_c_pm"},
		{"项目管理组", 0, "", "proj_c_pm"},
		{"结构设计组", 0, "", "proj_c_struct"},
		{"电子设计组", 0, "", "proj_c_elec"},
		{"工艺测试组", 0, "", "proj_c_process"},
	}

	statusTrue := true
	idMap := make(map[string]uint)

	for i, d := range depts {
		key := fmt.Sprintf("%s%d", d.name, i)
		dept := system.SysDepartment{
			Name:   d.name,
			Sort:   i,
			Status: &statusTrue,
		}
		switch i {
		case 0:
			dept.ParentId = 0
			dept.Ancestors = "0"
		case 1, 6, 11:
			parentKey := fmt.Sprintf("%s%d", "集团总部", 0)
			dept.ParentId = idMap[parentKey]
			parentID := idMap[parentKey]
			dept.Ancestors = fmt.Sprintf("0,%d", parentID)
		case 2, 3, 4, 5:
			parentKey := fmt.Sprintf("%s%d", "智能排产系统研发部", 1)
			dept.ParentId = idMap[parentKey]
			parentID := idMap[parentKey]
			dept.Ancestors = fmt.Sprintf("0,%d,%d", idMap[fmt.Sprintf("%s%d", "集团总部", 0)], parentID)
		case 7, 8, 9, 10:
			parentKey := fmt.Sprintf("%s%d", "工程建设事业部", 6)
			dept.ParentId = idMap[parentKey]
			parentID := idMap[parentKey]
			dept.Ancestors = fmt.Sprintf("0,%d,%d", idMap[fmt.Sprintf("%s%d", "集团总部", 0)], parentID)
		case 12, 13, 14, 15:
			parentKey := fmt.Sprintf("%s%d", "传感器研发中心", 11)
			dept.ParentId = idMap[parentKey]
			parentID := idMap[parentKey]
			dept.Ancestors = fmt.Sprintf("0,%d,%d", idMap[fmt.Sprintf("%s%d", "集团总部", 0)], parentID)
		}
		db.Create(&dept)
		idMap[key] = dept.ID
		deptIdMap[d.name+"_"+fmt.Sprintf("%d", i)] = dept.ID
	}
	return nil
}

func createPositions(db *gorm.DB) error {
	statusTrue := true
	positions := []struct {
		name, code string
		sort       int
	}{
		{"项目经理", "PM", 1},
		{"业务分析师", "BA", 2},
		{"前端开发工程师", "FE_DEV", 3},
		{"后端开发工程师", "BE_DEV", 4},
		{"测试工程师", "QA", 5},
		{"土建工程师", "CIVIL_ENG", 6},
		{"机电工程师", "MEP_ENG", 7},
		{"安全员", "SAFETY", 8},
		{"造价师", "QS", 9},
		{"结构工程师", "STRUCT_ENG", 10},
		{"电子工程师", "ELEC_ENG", 11},
		{"工艺工程师", "PROCESS_ENG", 12},
		{"PMO管理员", "PMO_ADMIN", 13},
		{"部门负责人", "DEPT_LEADER", 14},
		{"CCB成员", "CCB_MEMBER", 15},
	}
	for _, p := range positions {
		pos := system.SysPosition{Name: p.name, Code: p.code, Sort: p.sort, Status: &statusTrue}
		db.Create(&pos)
		posIdMap[p.code] = pos.ID
	}
	return nil
}

func createAuthorities(db *gorm.DB) error {
	parent888 := uint(888)
	authorities := []system.SysAuthority{
		{AuthorityId: 9001, AuthorityName: "PMO管理员", ParentId: &parent888, DataScope: 1, DefaultRouter: "dashboard"},
		{AuthorityId: 9002, AuthorityName: "项目经理", ParentId: &parent888, DataScope: 2, DefaultRouter: "dashboard"},
		{AuthorityId: 9003, AuthorityName: "团队成员", ParentId: &parent888, DataScope: 3, DefaultRouter: "dashboard"},
		{AuthorityId: 9004, AuthorityName: "干系人", ParentId: &parent888, DataScope: 4, DefaultRouter: "dashboard"},
	}
	for _, a := range authorities {
		db.Where("authority_id = ?", a.AuthorityId).FirstOrCreate(&a)
	}
	return nil
}

type userDef struct {
	username      string
	nickname      string
	deptNameIndex string
	posCode       string
	authId        uint
}

func createUsers(db *gorm.DB) error {
	users := []userDef{
		{"pmo01", "李PMO", "集团总部_0", "PMO_ADMIN", 9001},
		{"pmo02", "王PMO", "集团总部_0", "PMO_ADMIN", 9001},
		{"proj_a_pm", "张明", "项目管理组_2", "PM", 9002},
		{"proj_a_ba", "李娜", "项目管理组_2", "BA", 9003},
		{"proj_a_fe", "王强", "前端开发组_3", "FE_DEV", 9003},
		{"proj_a_be", "刘洋", "后端开发组_4", "BE_DEV", 9003},
		{"proj_a_qa", "陈静", "质量测试组_5", "QA", 9003},
		{"proj_b_pm", "赵刚", "项目管理部_7", "PM", 9002},
		{"proj_b_civil", "钱伟", "土建工程部_8", "CIVIL_ENG", 9003},
		{"proj_b_mep", "孙磊", "机电工程部_9", "MEP_ENG", 9003},
		{"proj_b_safety", "周梅", "安全造价部_10", "SAFETY", 9003},
		{"proj_b_qs", "吴芳", "安全造价部_10", "QS", 9003},
		{"proj_c_pm", "郑辉", "项目管理组_12", "PM", 9002},
		{"proj_c_struct", "冯雪", "结构设计组_13", "STRUCT_ENG", 9003},
		{"proj_c_elec", "褚晗", "电子设计组_14", "ELEC_ENG", 9003},
		{"proj_c_process", "卫鹏", "工艺测试组_15", "PROCESS_ENG", 9003},
		{"proj_c_test", "蒋琳", "工艺测试组_15", "QA", 9003},
	}

	pwdHash := utils.BcryptHash(defaultPassword)
	mustChange := false

	for _, u := range users {
		deptID := deptIdMap[u.deptNameIndex]
		user := system.SysUser{
			UUID:               uuid.New(),
			Username:           u.username,
			NickName:           u.nickname,
			Password:           pwdHash,
			AuthorityId:        u.authId,
			DeptId:             deptID,
			Enable:             1,
			MustChangePassword: mustChange,
		}
		db.Where("username = ?", u.username).FirstOrCreate(&user)
		userIdMap[u.username] = user.ID

		db.Create(&system.SysUserDepartment{SysUserId: user.ID, SysDepartmentId: deptID})

		if posID, ok := posIdMap[u.posCode]; ok {
			db.Create(&system.SysUserPosition{SysUserId: user.ID, SysPositionId: posID})
		}
		if u.username == "proj_a_pm" || u.username == "proj_b_pm" || u.username == "proj_c_pm" ||
			u.username == "proj_a_fe" || u.username == "proj_a_be" || u.username == "proj_a_qa" ||
			u.username == "proj_b_civil" || u.username == "proj_b_mep" || u.username == "proj_b_safety" ||
			u.username == "proj_c_struct" || u.username == "proj_c_elec" || u.username == "proj_c_process" ||
			u.username == "pmo01" {
			if posID, ok := posIdMap["DEPT_LEADER"]; ok {
				db.Create(&system.SysUserPosition{SysUserId: user.ID, SysPositionId: posID})
			}
		}
		if u.username == "pmo01" || u.username == "pmo02" ||
			u.username == "proj_a_pm" || u.username == "proj_b_pm" || u.username == "proj_c_pm" {
			if posID, ok := posIdMap["CCB_MEMBER"]; ok {
				db.Create(&system.SysUserPosition{SysUserId: user.ID, SysPositionId: posID})
			}
		}

		db.Create(&system.SysUserAuthority{SysUserId: user.ID, SysAuthorityAuthorityId: u.authId})
	}
	return nil
}

func updateDeptLeaders(db *gorm.DB) error {
	depts := []struct {
		key    string
		leader string
	}{
		{"集团总部_0", "pmo01"},
		{"智能排产系统研发部_1", "proj_a_pm"},
		{"项目管理组_2", "proj_a_pm"},
		{"前端开发组_3", "proj_a_fe"},
		{"后端开发组_4", "proj_a_be"},
		{"质量测试组_5", "proj_a_qa"},
		{"工程建设事业部_6", "proj_b_pm"},
		{"项目管理部_7", "proj_b_pm"},
		{"土建工程部_8", "proj_b_civil"},
		{"机电工程部_9", "proj_b_mep"},
		{"安全造价部_10", "proj_b_safety"},
		{"传感器研发中心_11", "proj_c_pm"},
		{"项目管理组_12", "proj_c_pm"},
		{"结构设计组_13", "proj_c_struct"},
		{"电子设计组_14", "proj_c_elec"},
		{"工艺测试组_15", "proj_c_process"},
	}
	for _, d := range depts {
		deptID, ok := deptIdMap[d.key]
		if !ok {
			continue
		}
		userID, ok := userIdMap[d.leader]
		if !ok {
			continue
		}
		db.Model(&system.SysDepartment{}).Where("id = ?", deptID).Update("leader_id", userID)
	}
	return nil
}
