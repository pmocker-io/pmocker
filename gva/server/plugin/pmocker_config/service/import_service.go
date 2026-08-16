package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// ImportService 配置包导入（跨实例迁移：export 产物 seed.yaml → 本实例）
type ImportService struct{}

// seedEntry 单个配置包导出条目（与 ExportService 的 seed.yaml 格式对齐）
type seedEntry struct {
	Code     string `yaml:"code" json:"code"`
	SeedYAML string `yaml:"seed_yaml" json:"seedYaml"`
}

// seedDoc seed.yaml 根结构（config_packages 数组）
type seedDoc struct {
	ConfigPackages []seedEntry `yaml:"config_packages" json:"configPackages"`
}

// ImportResult 导入结果统计
type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// ImportYAML 解析 seed.yaml（config_packages 数组），逐项幂等导入配置包
// 已存在（code 相同）的跳过；非法 seed 记入 Failed 并继续，不整体中断。
func (s *ImportService) ImportYAML(ctx context.Context, yamlContent string) (*ImportResult, error) {
	if yamlContent == "" {
		return nil, errors.New("导入内容为空")
	}
	var doc seedDoc
	if err := yaml.Unmarshal([]byte(yamlContent), &doc); err != nil {
		return nil, fmt.Errorf("seed.yaml 解析失败: %w", err)
	}
	if len(doc.ConfigPackages) == 0 {
		return nil, errors.New("config_packages 为空，无可导入配置包")
	}

	res := &ImportResult{}
	pkgSvc := &ConfigPackageService{}
	var firstErr error
	for _, entry := range doc.ConfigPackages {
		if entry.Code == "" || entry.SeedYAML == "" {
			res.Failed++
			if firstErr == nil {
				firstErr = errors.New("存在 code 或 seed_yaml 为空的条目")
			}
			continue
		}
		// 校验 seed 可解析
		if _, err := ParseSeedYAML([]byte(entry.SeedYAML)); err != nil {
			res.Failed++
			if firstErr == nil {
				firstErr = fmt.Errorf("配置包 %s 的 seed 解析失败: %w", entry.Code, err)
			}
			continue
		}
		// LoadSeedPackage 幂等：code 已存在则跳过
		existed := packageExists(ctx, entry.Code)
		seedYAML := strings.TrimSpace(entry.SeedYAML)
		if err := pkgSvc.LoadSeedPackage(ctx, entry.Code, seedYAML); err != nil {
			res.Failed++
			if firstErr == nil {
				firstErr = fmt.Errorf("配置包 %s 导入失败: %w", entry.Code, err)
			}
			continue
		}
		if existed {
			res.Skipped++
		} else {
			res.Imported++
		}
	}
	if res.Failed > 0 && firstErr != nil {
		global.GVA_LOG.Warn("配置包导入部分失败", zap.Int("failed", res.Failed), zap.Int("imported", res.Imported), zap.Int("skipped", res.Skipped), zap.Error(firstErr))
	}
	return res, nil
}

// packageExists 判断配置包是否已存在（LoadSeedPackage 幂等跳过后用于统计）
func packageExists(ctx context.Context, code string) bool {
	// 复用 LoadSeedPackage 的幂等语义：入库后必然存在
	var count int64
	global.GVA_DB.WithContext(ctx).Table("pm_config_packages").Where("code = ?", code).Count(&count)
	return count > 0
}
