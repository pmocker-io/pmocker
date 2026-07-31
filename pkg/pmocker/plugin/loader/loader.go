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

// LoadSchema 灌入 schema.yaml 到 EAV 元表
func (l *Loader) LoadSchema(ctx context.Context, yamlBytes []byte) error {
	var s SchemaYaml
	if err := yaml.Unmarshal(yamlBytes, &s); err != nil {
		return fmt.Errorf("parse schema.yaml: %w", err)
	}
	if err := l.EAV.RegisterEntityType(ctx, eavtypes.EntityType{
		TypeCode:   s.EntityType,
		ModuleCode: s.Module,
		Name:       s.Name,
		Icon:       s.Icon,
		IconColor:  s.IconColor,
	}); err != nil {
		return fmt.Errorf("register entity type %s: %w", s.EntityType, err)
	}
	for _, f := range s.Fields {
		if err := l.EAV.RegisterFieldDef(ctx, eavtypes.FieldDef{
			EntityType:   s.EntityType,
			FieldKey:     f.FieldKey,
			FieldLabel:   f.FieldLabel,
			DataType:     eavtypes.DataType(f.DataType),
			OptionsJSON:  f.OptionsJSON,
			DefaultValue: f.DefaultValue,
			Validators:   f.Validators,
		}); err != nil {
			return fmt.Errorf("register field %s.%s: %w", s.EntityType, f.FieldKey, err)
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
	return nil
}

// LoadMenu 灌入 menu.yaml 到 gva sys_base_menus
func (l *Loader) LoadMenu(yamlBytes []byte) error {
	if l.MenuReg == nil {
		return nil
	}
	var m MenuYaml
	if err := yaml.Unmarshal(yamlBytes, &m); err != nil {
		return fmt.Errorf("parse menu.yaml: %w", err)
	}
	menus := make([]system.SysBaseMenu, 0, len(m.Menus))
	for _, me := range m.Menus {
		menus = append(menus, system.SysBaseMenu{
			Path:      me.Path,
			Name:      me.Name,
			Hidden:    me.Hidden,
			Component: me.Component,
			Sort:      me.Sort,
			Meta:      system.Meta{Title: me.Title, Icon: me.Icon},
		})
	}
	l.MenuReg(menus...)
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
