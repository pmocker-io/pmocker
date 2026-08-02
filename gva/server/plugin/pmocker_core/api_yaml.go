package pmocker_core

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gopkg.in/yaml.v3"
)

// apiYaml 对应 pmocker/api.yaml 的结构
type apiYaml struct {
	APIs []system.SysApi `yaml:"apis"`
}

// parseApiYaml 解析 api.yaml 字节为 SysApi 列表
func parseApiYaml(data []byte) []system.SysApi {
	var a apiYaml
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil
	}
	return a.APIs
}
