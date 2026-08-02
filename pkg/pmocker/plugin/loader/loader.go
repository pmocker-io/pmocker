package loader

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/pmocker-io/pmocker/pkg/pmocker/workflow"
	"gopkg.in/yaml.v3"
)

// Loader 将插件元数据 YAML 灌入 EAV 元表和 gva 系统表
type Loader struct {
	EAV      eavtypes.EAVStore
	Workflow workflow.Engine
	MenuReg  func(menus ...system.SysBaseMenu)
	APIReg   func(apis ...system.SysApi)
	DictReg  func(dicts ...system.SysDictionary)
}

// NewFromGva 用 gva 已有的 service 和 plugin-tool/utils 构造 Loader
func NewFromGva(
	eav eavtypes.EAVStore,
	wf workflow.Engine,
	menuReg func(menus ...system.SysBaseMenu),
	apiReg func(apis ...system.SysApi),
	dictReg func(dicts ...system.SysDictionary),
) *Loader {
	return &Loader{EAV: eav, Workflow: wf, MenuReg: menuReg, APIReg: apiReg, DictReg: dictReg}
}

// LoadSchema 灌入 schema.yaml 到 EAV 元表。
// 当 entity_types 数组非空时按多实体模式注册，否则走单实体模式（向后兼容）。
func (l *Loader) LoadSchema(ctx context.Context, yamlBytes []byte) error {
	var s SchemaYaml
	if err := yaml.Unmarshal(yamlBytes, &s); err != nil {
		return fmt.Errorf("parse schema.yaml: %w", err)
	}
	if len(s.EntityTypes) > 0 {
		for _, et := range s.EntityTypes {
			if err := l.loadSingleSchema(ctx, et.EntityType, et.Module, et.Name, et.Icon, et.IconColor, et.Fields); err != nil {
				return err
			}
		}
		return nil
	}
	return l.loadSingleSchema(ctx, s.EntityType, s.Module, s.Name, s.Icon, s.IconColor, s.Fields)
}

// loadSingleSchema 注册单个实体类型及其字段定义
func (l *Loader) loadSingleSchema(ctx context.Context, typeCode, module, name, icon, iconColor string, fields []FieldYaml) error {
	if err := l.EAV.RegisterEntityType(ctx, eavtypes.EntityType{
		TypeCode:   typeCode,
		ModuleCode: module,
		Name:       name,
		Icon:       icon,
		IconColor:  iconColor,
	}); err != nil {
		return fmt.Errorf("register entity type %s: %w", typeCode, err)
	}
	for _, f := range fields {
		if err := l.EAV.RegisterFieldDef(ctx, eavtypes.FieldDef{
			EntityType:   typeCode,
			FieldKey:     f.FieldKey,
			FieldLabel:   f.FieldLabel,
			DataType:     eavtypes.DataType(f.DataType),
			OptionsJSON:  f.OptionsJSON,
			DefaultValue: f.DefaultValue,
			Validators:   f.Validators,
		}); err != nil {
			return fmt.Errorf("register field %s.%s: %w", typeCode, f.FieldKey, err)
		}
	}
	return nil
}

// LoadSeed 灌入 seed.yaml 中的字典到 gva（角色/EPS 在 M4 阶段扩展）
func (l *Loader) LoadSeed(ctx context.Context, yamlBytes []byte) error {
	var seed SeedYaml
	if err := yaml.Unmarshal(yamlBytes, &seed); err != nil {
		return fmt.Errorf("parse seed.yaml: %w", err)
	}
	if l.DictReg != nil && len(seed.Dictionaries) > 0 {
		dicts := make([]system.SysDictionary, 0, len(seed.Dictionaries))
		for _, d := range seed.Dictionaries {
			details := make([]system.SysDictionaryDetail, 0, len(d.Details))
			for _, dt := range d.Details {
				details = append(details, system.SysDictionaryDetail{
					Label:  dt.Label,
					Value:  dt.Value,
					Extend: dt.Extend,
				})
			}
			dicts = append(dicts, system.SysDictionary{
				Type:                 d.Type,
				Name:                 d.Name,
				SysDictionaryDetails: details,
			})
		}
		l.DictReg(dicts...)
	}
	// 加载 EPS 种子数据（支持顶层 eps 和 entities.eps 两种格式）
	// 幂等：按 name 去重，已存在的 eps_node 跳过，避免服务重启导致重复创建
	epsNodes := seed.EPS
	if len(epsNodes) == 0 {
		epsNodes = seed.Entities.EPS
	}
	if l.EAV != nil && len(epsNodes) > 0 {
		existing, _, err := l.EAV.ListEntities(ctx, 0, "eps_node", 0, 10000)
		if err != nil {
			return fmt.Errorf("list existing eps_node for idempotent check: %w", err)
		}
		existingNames := make(map[string]bool, len(existing))
		for _, e := range existing {
			if name, ok := e.Attrs["name"].(string); ok && name != "" {
				existingNames[name] = true
			}
		}
		for _, e := range epsNodes {
			if existingNames[e.Name] {
				continue
			}
			attrs := map[string]interface{}{
				"parent_path": e.ParentPath,
				"name":        e.Name,
				"type":        e.Type,
				"code":        e.Code,
				"sort":        e.Sort,
				"is_active":   e.IsActive,
				"status":      e.Status,
			}
			if _, err := l.EAV.CreateEntity(ctx, eavtypes.Entity{
				EntityType: "eps_node",
				Title:      e.Name,
				Status:     e.Status,
				Attrs:      attrs,
			}); err != nil {
				return fmt.Errorf("create eps_node seed %q: %w", e.Name, err)
			}
		}
	}
	return nil
}

// LoadMenu 灌入 menu.yaml 到 gva sys_base_menus
// 支持多顶级菜单：按 parent_name 分组，每组先注册父菜单再注册子菜单
func (l *Loader) LoadMenu(yamlBytes []byte) error {
	if l.MenuReg == nil {
		return nil
	}
	var m MenuYaml
	if err := yaml.Unmarshal(yamlBytes, &m); err != nil {
		return fmt.Errorf("parse menu.yaml: %w", err)
	}

	// 按 parent_name 分组菜单
	// 顶级菜单: ParentName 为空
	// 子菜单: ParentName 指向父菜单的 Name
	topMenus := make([]system.SysBaseMenu, 0)
	childrenMap := make(map[string][]system.SysBaseMenu) // parent_name -> children

	for _, me := range m.Menus {
		menu := system.SysBaseMenu{
			Path:      me.Path,
			Name:      me.Name,
			Hidden:    me.Hidden,
			Component: me.Component,
			Sort:      me.Sort,
			Meta:      system.Meta{Title: me.Title, Icon: me.Icon},
		}
		if me.ParentName == "" {
			// 顶级菜单
			topMenus = append(topMenus, menu)
		} else {
			// 子菜单
			childrenMap[me.ParentName] = append(childrenMap[me.ParentName], menu)
		}
	}

	// 对每个顶级菜单及其子菜单分组注册
	// RegisterMenus 函数会自动将第一个菜单作为父菜单，其余作为子菜单
	for _, topMenu := range topMenus {
		group := []system.SysBaseMenu{topMenu}
		// 添加该父菜单的子菜单
		if children, ok := childrenMap[topMenu.Name]; ok {
			group = append(group, children...)
		}
		l.MenuReg(group...)
	}

	// 处理跨插件的子菜单（parent_name 指向其他插件注册的菜单）
	// 这些子菜单没有在当前分组中找到父菜单
	for parentName, children := range childrenMap {
		found := false
		for _, topMenu := range topMenus {
			if topMenu.Name == parentName {
				found = true
				break
			}
		}
		if !found {
			// 跨插件的子菜单：直接注册子菜单，不创建临时父菜单
			// RegisterMenus 会根据 parentName 查找或创建父菜单
			for _, child := range children {
				// 查找已存在的父菜单（由其他插件注册），或创建临时引用
				l.MenuReg(system.SysBaseMenu{
					Name: parentName,
				}, child)
			}
		}
	}

	return nil
}

// LoadAPI 灌入 api.yaml 到 gva sys_apis
func (l *Loader) LoadAPI(yamlBytes []byte) error {
	if l.APIReg == nil {
		return nil
	}
	var a APIYaml
	if err := yaml.Unmarshal(yamlBytes, &a); err != nil {
		return fmt.Errorf("parse api.yaml: %w", err)
	}
	l.APIReg(a.APIs...)
	return nil
}

// LoadWorkflow 灌入单个工作流 YAML
func (l *Loader) LoadWorkflow(ctx context.Context, yamlBytes []byte) error {
	var def workflow.Definition
	if err := yaml.Unmarshal(yamlBytes, &def); err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}
	return l.Workflow.LoadDefinition(ctx, def)
}

// LoadWorkflowDir 从 embed.FS 加载目录下所有 .yaml 工作流
func (l *Loader) LoadWorkflowDir(ctx context.Context, fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		bytes, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if err := l.LoadWorkflow(ctx, bytes); err != nil {
			return fmt.Errorf("load workflow %s: %w", filepath.Base(path), err)
		}
		return nil
	})
}
