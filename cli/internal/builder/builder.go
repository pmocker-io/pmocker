// Package builder 管理 gva server 二进制的自动构建。
package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Builder gva server 二进制构建器
type Builder struct {
	binPath      string // ~/.pmocker/bin/gva-server
	mcpBinPath   string // ~/.pmocker/bin/gva-mcp
	gvaServerDir string // gva/server 源码路径
	gvaWebDir    string // gva/web 源码路径
}

// NewBuilder 创建构建器
func NewBuilder(pmockerHome, gvaServerDir, gvaWebDir string) *Builder {
	binName := "gva-server"
	mcpBinName := "gva-mcp"
	if runtime.GOOS == "windows" {
		binName += ".exe"
		mcpBinName += ".exe"
	}
	return &Builder{
		binPath:      filepath.Join(pmockerHome, "bin", binName),
		mcpBinPath:   filepath.Join(pmockerHome, "bin", mcpBinName),
		gvaServerDir: gvaServerDir,
		gvaWebDir:    gvaWebDir,
	}
}

// McpBinPath 返回 MCP 二进制路径
func (b *Builder) McpBinPath() string {
	return b.mcpBinPath
}

// Ensure 确保二进制存在，不存在则自动构建
func (b *Builder) Ensure() error {
	serverExists := fileExists(b.binPath)
	mcpExists := fileExists(b.mcpBinPath)
	if serverExists && mcpExists {
		return nil
	}
	fmt.Println("首次运行，正在构建 gva 二进制...")
	if !serverExists {
		if err := b.buildServer(); err != nil {
			return fmt.Errorf("build gva server: %w", err)
		}
	}
	if !mcpExists {
		if err := b.buildMCPServer(); err != nil {
			return fmt.Errorf("build gva mcp: %w", err)
		}
	}
	if err := b.buildWeb(); err != nil {
		return fmt.Errorf("build gva web: %w", err)
	}
	fmt.Println("构建完成")
	return nil
}

// BinPath 返回二进制路径
func (b *Builder) BinPath() string {
	return b.binPath
}

func (b *Builder) buildServer() error {
	if err := os.MkdirAll(filepath.Dir(b.binPath), 0755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", b.binPath, ".")
	cmd.Dir = b.gvaServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (b *Builder) buildMCPServer() error {
	if err := os.MkdirAll(filepath.Dir(b.mcpBinPath), 0755); err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", b.mcpBinPath, "./cmd/mcp")
	cmd.Dir = b.gvaServerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (b *Builder) buildWeb() error {
	nodeModules := filepath.Join(b.gvaWebDir, "node_modules")
	if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
		fmt.Println("安装前端依赖...")
		cmd := exec.Command("npm", "install")
		cmd.Dir = b.gvaWebDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npm install: %w", err)
		}
	}
	fmt.Println("构建前端...")
	cmd := exec.Command("npm", "run", "build")
	cmd.Dir = b.gvaWebDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm run build: %w", err)
	}
	// 复制 dist 到 binPath 同级目录
	srcDist := filepath.Join(b.gvaWebDir, "dist")
	dstDist := filepath.Join(filepath.Dir(b.binPath), "dist")
	fmt.Println("复制前端 dist 到", dstDist)
	return copyDir(srcDist, dstDist)
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
