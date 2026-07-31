package eav

// SchemaDefinition 镜像 schema 层的 YAML 结构（每个插件的 pmocker/schema.yaml）
type SchemaDefinition struct {
	EntityType string    `yaml:"entity_type"`
	Module     string    `yaml:"module"`
	Fields     []FieldDef `yaml:"fields"`
	States     []string  `yaml:"states"`
	Workflow   string    `yaml:"workflow,omitempty"`
}

// SeedData 镜像 seed 层的 YAML 结构（每个插件的 pmocker/seed.yaml）
type SeedData struct {
	Roles        []map[string]interface{} `yaml:"roles"`
	Dictionaries []map[string]interface{} `yaml:"dictionaries"`
	EPS          []map[string]interface{} `yaml:"eps"`
}
