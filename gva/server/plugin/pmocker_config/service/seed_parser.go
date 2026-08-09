package service

import (
	"errors"

	"gopkg.in/yaml.v3"
)

// FieldSeed 字段定义种子
type FieldSeed struct {
	Key      string   `yaml:"key" json:"key"`
	Label    string   `yaml:"label" json:"label"`
	DataType string   `yaml:"data_type" json:"dataType"`
	Options  []string `yaml:"options" json:"options"`
	Default  string   `yaml:"default" json:"default"`
}

// StateSeed 状态定义种子
type StateSeed struct {
	Status  string `yaml:"status" json:"status"`
	Label   string `yaml:"label" json:"label"`
	TagType string `yaml:"tag_type" json:"tagType"`
}

// TransitionSeed 流转规则种子（含退回）
type TransitionSeed struct {
	From     string `yaml:"from" json:"from"`
	To       string `yaml:"to" json:"to"`
	Action   string `yaml:"action" json:"action"`
	Rollback bool   `yaml:"rollback" json:"rollback"`
}

// ProjectSeed 项目种子（EPS 树节点或业务项目引用）
type ProjectSeed struct {
	Code        string                 `yaml:"code" json:"code"`
	Name        string                 `yaml:"name" json:"name"`
	Type        string                 `yaml:"type" json:"type"`
	ProjectID   uint                   `yaml:"project_id" json:"projectId"`
	ProjectCode string                 `yaml:"project_code" json:"projectCode"`
	Status      string                 `yaml:"status" json:"status"`
	Priority    int                    `yaml:"priority" json:"priority"`
	Children    []ProjectSeed          `yaml:"children" json:"children"`
	Entities    map[string][]map[string]interface{} `yaml:"entities" json:"entities"`
}

// ConfigPackageSeed 配置包种子（聚合所有模块）
// 一个配置包 = 所有模块的完整种子数据集合
type ConfigPackageSeed struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Modules     map[string]ModuleSeed  `yaml:"modules" json:"modules"`
}

// ModuleSeed 单个模块的种子定义
type ModuleSeed struct {
	EntityType  string           `yaml:"entity_type" json:"entityType"`
	Name        string           `yaml:"name" json:"name"`
	Fields      []FieldSeed      `yaml:"fields" json:"fields"`
	States      []StateSeed      `yaml:"states" json:"states"`
	Transitions []TransitionSeed `yaml:"transitions" json:"transitions"`
	Projects    []ProjectSeed    `yaml:"projects" json:"projects"`
}

// ParseSeedYAML 解析配置包种子 YAML 为结构化（多模块聚合）
func ParseSeedYAML(bytes []byte) (*ConfigPackageSeed, error) {
	if len(bytes) == 0 {
		return nil, errors.New("seed_yaml 为空")
	}
	var seed ConfigPackageSeed
	if err := yaml.Unmarshal(bytes, &seed); err != nil {
		return nil, err
	}
	if len(seed.Modules) == 0 {
		return nil, errors.New("seed_yaml 缺少 modules（至少一个模块）")
	}
	return &seed, nil
}
