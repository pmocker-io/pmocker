// build.go 构建默认的 PMBOK6-Hybrid 镜像。
// 从 gva/server/plugin/pmocker_*/pmocker/ 目录收集元数据，
// 组装成 .pmi 镜像文件（manifest.json + config.json + 3 层 tar）。
//
// 运行方式：在 images/pmbok6-hybrid 目录下执行 `go run .`
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

func main() {
	pluginDirs := []string{
		"requirement", "scope", "schedule", "cost", "risk",
		"issue", "eps", "deliverable", "change",
	}

	schemaFiles := make(map[string][]byte)
	pluginFiles := make(map[string][]byte)

	gvaPluginBase := filepath.Join("..", "..", "gva", "server", "plugin")
	for _, mod := range pluginDirs {
		pmockerDir := filepath.Join(gvaPluginBase, "pmocker_"+mod, "pmocker")

		// schema 聚合层：按模块名重命名
		schemaPath := filepath.Join(pmockerDir, "schema.yaml")
		if data, err := os.ReadFile(schemaPath); err == nil {
			schemaFiles[mod+".yaml"] = data
		} else {
			fmt.Fprintf(os.Stderr, "warn: 跳过 %s schema: %v\n", mod, err)
		}

		// plugins 层：保留原始路径结构
		for _, name := range []string{"manifest.yaml", "seed.yaml", "menu.yaml", "api.yaml"} {
			p := filepath.Join(pmockerDir, name)
			if data, err := os.ReadFile(p); err == nil {
				pluginFiles["pmocker_"+mod+"/pmocker/"+name] = data
			}
		}
	}

	// schema 层
	_, schemaTar, err := oci.CreateLayerFromFiles(schemaFiles, oci.LayerTypeSchema)
	if err != nil {
		log.Fatalf("create schema layer: %v", err)
	}
	schemaLayer := oci.NewLayerData(schemaTar, oci.LayerTypeSchema)

	// plugins 层
	_, pluginTar, err := oci.CreateLayerFromFiles(pluginFiles, oci.LayerTypePlugins)
	if err != nil {
		log.Fatalf("create plugins layer: %v", err)
	}
	pluginLayer := oci.NewLayerData(pluginTar, oci.LayerTypePlugins)

	// theme 层（占位）
	themeFiles := map[string][]byte{
		"theme.css": []byte(":root { --pm-primary: #409EFF; }\n"),
	}
	_, themeTar, _ := oci.CreateLayerFromFiles(themeFiles, oci.LayerTypeTheme)
	themeLayer := oci.NewLayerData(themeTar, oci.LayerTypeTheme)

	// config
	cfg := oci.NewConfig("PMBOK6-Hybrid", "1.0.0", pluginDirs)

	// 输出
	outPath := filepath.Join("pmbok6-hybrid.pmi")
	if err := oci.BuildImage(outPath, cfg, []oci.LayerData{schemaLayer, pluginLayer, themeLayer}); err != nil {
		log.Fatalf("build image: %v", err)
	}
	fmt.Printf("已生成镜像: %s\n", outPath)
	fmt.Printf("  - schema 文件数: %d\n", len(schemaFiles))
	fmt.Printf("  - plugin 文件数: %d\n", len(pluginFiles))
}
