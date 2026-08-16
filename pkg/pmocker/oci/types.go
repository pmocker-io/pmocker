// Package oci 实现 PMocker 镜像格式（简化 OCI）。
// .pmi 文件本质是 tar 归档，包含 manifest.json、config.json 和若干 layer tar。
package oci

import "encoding/json"

// Manifest 镜像清单（manifest.json）
type Manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	Config        Descriptor        `json:"config"`
	Layers        []LayerDescriptor `json:"layers"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// Descriptor 内容寻址描述符
type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"` // sha256:xxx
	Size      int64  `json:"size"`
}

// LayerDescriptor 带类型标注的层描述符
type LayerDescriptor struct {
	Descriptor
	Annotations map[string]string `json:"annotations,omitempty"`
}

// Config 镜像配置（config.json）
type Config struct {
	Methodology   string              `json:"methodology"`
	Version       string              `json:"version"`
	GVAVersion    string              `json:"gvaVersion"`
	Entrypoint    []string            `json:"entrypoint"`
	ExposedPorts  map[string]struct{} `json:"exposedPorts"`
	DBDriver      string              `json:"dbDriver"`
	Modules       []string            `json:"modules"`
	DefaultAdmin  DefaultAdmin        `json:"defaultAdmin"`
}

// DefaultAdmin 默认管理员配置
type DefaultAdmin struct {
	Username    string `json:"username"`
	PasswordEnv string `json:"passwordEnv"`
}

// LayerType 层类型枚举
type LayerType string

const (
	LayerTypePlugins LayerType = "plugins"
	LayerTypeSchema   LayerType = "schema"
	LayerTypeTheme    LayerType = "theme"
	LayerTypeAssets   LayerType = "assets"
	// LayerTypeData 实例数据层：commit/export 时打包的数据卷快照（sqlite + dist + uploads）
	LayerTypeData LayerType = "data"
)

// DigestPrefix SHA256 摘要前缀
const DigestPrefix = "sha256:"

// 层的 MediaType 常量
const (
	MediaTypeManifest   = "application/vnd.pmocker.manifest.v1+json"
	MediaTypeConfig     = "application/vnd.pmocker.config.v1+json"
	MediaTypeLayerPrefix = "application/vnd.pmocker.layer."
)

// MediaTypeString 为 LayerDescriptor 添加 mediaType 字段
func (l LayerDescriptor) MediaTypeString() string {
	return MediaTypeLayerPrefix + string(l.Annotations["pmocker.layer.type"]) + ".v1.tar"
}

// NewManifest 创建默认 Manifest
func NewManifest(cfg Config, cfgDigest string, cfgSize int64, layers []LayerDescriptor) Manifest {
	return Manifest{
		SchemaVersion: 2,
		MediaType:     MediaTypeManifest,
		Config: Descriptor{
			MediaType: MediaTypeConfig,
			Digest:    cfgDigest,
			Size:      cfgSize,
		},
		Layers:      layers,
		Annotations: map[string]string{},
	}
}

// NewConfig 创建默认 Config
func NewConfig(methodology, version string, modules []string) Config {
	return Config{
		Methodology:  methodology,
		Version:      version,
		GVAVersion:   "v3.0.0",
		Entrypoint:   []string{"./gva-server", "--config=config.yaml"},
		ExposedPorts: map[string]struct{}{"8080/tcp": {}},
		DBDriver:     "sqlite",
		Modules:      modules,
		DefaultAdmin: DefaultAdmin{Username: "admin", PasswordEnv: "PMOCKER_ADMIN_PASSWORD"},
	}
}

// JSONMarshal 便捷序列化
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
