// build.go 构建默认的 PMBOK6-Hybrid 镜像。
// 自动扫描 gva/server/plugin/pmocker_* 目录收集元数据，
// 组装成 .pmi 镜像文件（manifest.json + config.json + 3 层 tar）。
//
// 运行方式：在 images/pmbok6-hybrid 目录下执行 `go run .`
//
// 新增 PMocker 插件无需修改本文件——会自动发现 pmocker_* 目录（pmocker_core 除外）。
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

// pmockerCore 是核心插件（不参与 schema 聚合，无业务实体）
const pmockerCore = "pmocker_core"

func main() {
	gvaPluginBase := filepath.Join("..", "..", "gva", "server", "plugin")

	// 自动发现所有 pmocker_* 业务插件目录（排除 pmocker_core）
	pluginDirs := discoverPlugins(gvaPluginBase)

	schemaFiles := make(map[string][]byte)
	pluginFiles := make(map[string][]byte)

	for _, mod := range pluginDirs {
		pmockerDir := filepath.Join(gvaPluginBase, "pmocker_"+mod, "pmocker")

		// schema 聚合层：按模块名重命名
		schemaPath := filepath.Join(pmockerDir, "schema.yaml")
		if data, err := os.ReadFile(schemaPath); err == nil {
			schemaFiles[mod+".yaml"] = data
		} else {
			fmt.Fprintf(os.Stderr, "warn: 跳过 %s schema: %v\n", mod, err)
		}

		// plugins 层：保留原始路径结构（manifest/seed/menu/api + workflows 目录）
		for _, name := range []string{"manifest.yaml", "seed.yaml", "menu.yaml", "api.yaml"} {
			p := filepath.Join(pmockerDir, name)
			if data, err := os.ReadFile(p); err == nil {
				pluginFiles["pmocker_"+mod+"/pmocker/"+name] = data
			}
		}
		// workflows 目录（递归收集 *.yaml）
		workflowsDir := filepath.Join(pmockerDir, "workflows")
		if entries, err := os.ReadDir(workflowsDir); err == nil {
			for _, e := range entries {
				if e.IsDir() || filepath.Ext(e.Name()) != ".yaml" {
					continue
				}
				p := filepath.Join(workflowsDir, e.Name())
				if data, err := os.ReadFile(p); err == nil {
					pluginFiles["pmocker_"+mod+"/pmocker/workflows/"+e.Name()] = data
				}
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
	fmt.Printf("  - 发现插件: %v\n", pluginDirs)
}

// discoverPlugins 扫描 gva/server/plugin/ 目录，自动发现所有 pmocker_* 业务插件。
// 排除 pmocker_core（核心包，无业务实体，不参与 schema 聚合）。
// 要求每个业务插件目录下存在 pmocker/schema.yaml（确保是完整插件而非空目录）。
func discoverPlugins(gvaPluginBase string) []string {
	entries, err := os.ReadDir(gvaPluginBase)
	if err != nil {
		log.Fatalf("扫描插件目录失败 %s: %v", gvaPluginBase, err)
	}

	var mods []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "pmocker_") {
			continue
		}
		if name == pmockerCore {
			continue
		}
		// 验证是完整插件：pmocker 子目录 + schema.yaml 存在
		mod := strings.TrimPrefix(name, "pmocker_")
		schemaPath := filepath.Join(gvaPluginBase, name, "pmocker", "schema.yaml")
		if _, err := os.Stat(schemaPath); err != nil {
			fmt.Fprintf(os.Stderr, "warn: 跳过 %s（缺少 pmocker/schema.yaml）: %v\n", name, err)
			continue
		}
		mods = append(mods, mod)
	}

	sort.Strings(mods)
	if len(mods) == 0 {
		log.Fatal("未发现任何 pmocker_* 业务插件")
	}
	return mods
}
