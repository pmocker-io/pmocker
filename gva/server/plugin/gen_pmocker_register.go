//go:build ignore

// gen_pmocker_register.go 自动扫描 plugin/ 目录下的 pmocker_* 子目录，
// 读取各插件的 pmocker/manifest.yaml 获取模块中文名，生成空白导入文件 pmocker_register.go。
//
// 运行方式：在 gva/server/plugin/ 目录下执行 `go run gen_pmocker_register.go`
// 或在 gva/server/ 目录下执行 `go generate ./plugin/`（pmocker_register.go 中有 go:generate 指令）。
//
// 新增 PMocker 插件后运行本生成器即可自动更新空白导入，无需手动编辑 pmocker_register.go。
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// manifest 对应 pmocker/manifest.yaml 的部分字段
type manifest struct {
	Code string `yaml:"code"`
	Name string `yaml:"name"`
}

const pluginBase = "github.com/flipped-aurora/gin-vue-admin/server/plugin"

func main() {
	// 生成器位于 gva/server/plugin/ 目录下
	entries, err := os.ReadDir(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "扫描当前目录失败:", err)
		os.Exit(1)
	}

	// 发现所有 pmocker_* 目录
	// pmocker_core 无 manifest.yaml，硬编码处理；业务插件要求 pmocker/manifest.yaml
	type pluginInfo struct {
		ImportPath string
		Name       string // 中文名
		DirName    string // pmocker_xxx
	}

	var core = pluginInfo{
		ImportPath: pluginBase + "/pmocker_core",
		Name:       "核心插件",
		DirName:    "pmocker_core",
	}
	var business []pluginInfo

	for _, e := range entries {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "pmocker_") {
			continue
		}
		if e.Name() == "pmocker_core" {
			continue // core 已硬编码
		}

		// 读取 manifest.yaml 获取中文名
		manifestPath := filepath.Join(e.Name(), "pmocker", "manifest.yaml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: 跳过 %s（缺少 pmocker/manifest.yaml）: %v\n", e.Name(), err)
			continue
		}
		var m manifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			fmt.Fprintf(os.Stderr, "warn: 跳过 %s（manifest.yaml 解析失败）: %v\n", e.Name(), err)
			continue
		}

		info := pluginInfo{
			ImportPath: pluginBase + "/" + e.Name(),
			Name:       m.Name,
			DirName:    e.Name(),
		}
		business = append(business, info)
	}

	if len(business) == 0 {
		fmt.Fprintln(os.Stderr, "错误: 未发现任何 pmocker_* 业务插件")
		os.Exit(1)
	}

	// 按目录名字母序排列业务插件
	sort.Slice(business, func(i, j int) bool {
		return business[i].DirName < business[j].DirName
	})

	// 生成文件内容
	var buf strings.Builder
	buf.WriteString("// 此文件由 gen_pmocker_register.go 自动生成，请勿手动编辑。\n")
	buf.WriteString("// 新增 PMocker 插件后运行: go generate ./plugin/\n")
	buf.WriteString("//go:generate go run gen_pmocker_register.go\n\n")
	buf.WriteString("package plugin\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t// 空白导入 PMocker 插件，触发各自的 init() 注册到 gva 插件注册表\n")
	buf.WriteString("\t// 核心插件（最先导入，负责初始化钩子和表）\n")
	buf.WriteString(fmt.Sprintf("\t_ \"%s\"\n\n", core.ImportPath))
	buf.WriteString("\t// 业务插件（按目录名字母序自动生成）\n")

	// 计算最长 ImportPath 用于对齐注释
	maxLen := 0
	for _, b := range business {
		if len(b.ImportPath) > maxLen {
			maxLen = len(b.ImportPath)
		}
	}

	for _, b := range business {
		pad := strings.Repeat(" ", maxLen-len(b.ImportPath))
		buf.WriteString(fmt.Sprintf("\t_ \"%s\"%s  // %s\n", b.ImportPath, pad, b.Name))
	}
	buf.WriteString(")\n")

	// 覆写 pmocker_register.go
	outPath := "pmocker_register.go"
	if err := os.WriteFile(outPath, []byte(buf.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "写入 pmocker_register.go 失败:", err)
		os.Exit(1)
	}

	fmt.Printf("已生成 %s（%d 核心插件 + %d 业务插件）\n", outPath, 1, len(business))
	for _, b := range business {
		fmt.Printf("  - %s: %s\n", b.DirName, b.Name)
	}
}
