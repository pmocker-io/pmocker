package instance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
)

// RunOptions pmocker run 选项
type RunOptions struct {
	ImageRef      string
	Name          string
	Port          int
	AdminPassword string
}

// Manager 实例生命周期管理器
type Manager struct {
	store      Store
	volumes    *VolumeManager
	imageStore *image.Store
	binPath    string
}

// NewManager 创建实例管理器
func NewManager(store Store, vols *VolumeManager, imgStore *image.Store, binPath string) *Manager {
	return &Manager{
		store:      store,
		volumes:    vols,
		imageStore: imgStore,
		binPath:    binPath,
	}
}

// Run 启动新实例
func (m *Manager) Run(opts RunOptions) (*Instance, error) {
	// 1. 解析镜像
	var digest string
	if opts.ImageRef != "" {
		name, tag := parseRef(opts.ImageRef)
		info, err := m.imageStore.ResolveImage(name, tag)
		if err != nil {
			// 可能是 .pmi 文件，尝试导入
			if _, err := os.Stat(opts.ImageRef); err == nil {
				imported, err := m.imageStore.AddImage(opts.ImageRef, name, tag)
				if err != nil {
					return nil, fmt.Errorf("import image: %w", err)
				}
				digest = imported.Digest
			} else {
				return nil, fmt.Errorf("resolve image %s: %w", opts.ImageRef, err)
			}
		} else {
			digest = info.Digest
		}
	} else {
		return nil, fmt.Errorf("image is required")
	}

	// 2. 创建数据卷
	volumeID, err := m.volumes.Create()
	if err != nil {
		return nil, fmt.Errorf("create volume: %w", err)
	}

	// 3. 生成实例名称（如未指定）
	name := opts.Name
	if name == "" {
		name = "pms-" + volumeID[:8]
	}

	// 4. 创建实例记录
	now := time.Now()
	inst := &Instance{
		ID:          uuid.New().String(),
		Name:        name,
		ImageDigest: digest,
		ImageRef:    opts.ImageRef,
		Port:        opts.Port,
		VolumeID:    volumeID,
		PID:         0,
		Status:      "stopped",
		CreatedAt:   now,
	}

	// 5. 生成 config.yaml
	if err := GenerateConfig(m.volumes, inst); err != nil {
		m.volumes.Remove(volumeID)
		return nil, fmt.Errorf("generate config: %w", err)
	}

	// 5.5 复制前端 dist 到数据卷（如果存在）
	if err := m.copyFrontendDist(volumeID); err != nil {
		// dist 不存在不阻止启动，只是没有前端页面
		fmt.Printf("warning: copy frontend dist: %v\n", err)
	}

	// 6. fork gva-server
	pid, err := m.startProcess(inst, opts.AdminPassword)
	if err != nil {
		return nil, fmt.Errorf("start process: %w (volume kept at %s for debugging)", err, m.volumes.Path(volumeID))
	}

	inst.PID = pid
	inst.Status = "running"
	inst.StartedAt = &now

	// 7. 写入实例记录
	if err := m.store.Create(inst); err != nil {
		m.stopProcess(pid)
		m.volumes.Remove(volumeID)
		return nil, fmt.Errorf("save instance: %w", err)
	}

	return inst, nil
}

// Stop 停止实例
func (m *Manager) Stop(idOrName string) error {
	inst, err := m.Get(idOrName)
	if err != nil {
		return err
	}
	if inst.Status != "running" {
		return fmt.Errorf("instance %s is not running", inst.Name)
	}
	if err := m.stopProcess(inst.PID); err != nil {
		return err
	}
	now := time.Now()
	inst.PID = 0
	inst.Status = "stopped"
	inst.StoppedAt = &now
	return m.store.Update(inst)
}

// Start 重启已停止实例
func (m *Manager) Start(idOrName string) error {
	inst, err := m.Get(idOrName)
	if err != nil {
		return err
	}
	if inst.Status == "running" {
		return fmt.Errorf("instance %s is already running", inst.Name)
	}
	pid, err := m.startProcess(inst, "")
	if err != nil {
		return fmt.Errorf("start process: %w", err)
	}
	now := time.Now()
	inst.PID = pid
	inst.Status = "running"
	inst.StartedAt = &now
	return m.store.Update(inst)
}

// Remove 删除实例
func (m *Manager) Remove(idOrName string, removeVolume bool) error {
	inst, err := m.Get(idOrName)
	if err != nil {
		return err
	}
	if inst.Status == "running" {
		if err := m.Stop(idOrName); err != nil {
			return err
		}
	}
	if err := m.store.Delete(inst.ID); err != nil {
		return err
	}
	if removeVolume {
		return m.volumes.Remove(inst.VolumeID)
	}
	return nil
}

// Get 按 ID 或名称查找实例
func (m *Manager) Get(idOrName string) (*Instance, error) {
	inst, err := m.store.GetByID(idOrName)
	if err == nil {
		return inst, nil
	}
	return m.store.GetByName(idOrName)
}

// startProcess fork gva-server 进程
func (m *Manager) startProcess(inst *Instance, adminPassword string) (int, error) {
	cfgPath := m.volumes.ConfigPath(inst.VolumeID)
	volPath := m.volumes.Path(inst.VolumeID)
	logPath := filepath.Join(volPath, "gva-server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return 0, fmt.Errorf("open log file: %w", err)
	}
	cmd := exec.Command(m.binPath)
	cmd.Dir = volPath // 工作目录设为数据卷，使 ./dist 能被找到
	cmd.Env = append(os.Environ(),
		"GVA_CONFIG="+cfgPath,
		"GIN_MODE=release",
		"PMOCKER_AUTO_INIT=1",
	)
	if adminPassword != "" {
		cmd.Env = append(cmd.Env, "PMOCKER_ADMIN_PASSWORD="+adminPassword)
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// detach 进程，避免父进程退出时被杀
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return 0, err
	}
	pid := cmd.Process.Pid
	// 释放资源，让进程独立运行
	_ = cmd.Process.Release()
	logFile.Close()
	time.Sleep(2 * time.Second)
	if !processExists(pid) {
		return 0, fmt.Errorf("gva-server exited immediately, check log: %s", logPath)
	}
	return pid, nil
}

// copyFrontendDist 复制前端 dist 到数据卷
func (m *Manager) copyFrontendDist(volumeID string) error {
	// dist 与 gva-server 二进制同级目录
	srcDist := filepath.Join(filepath.Dir(m.binPath), "dist")
	if _, err := os.Stat(srcDist); err != nil {
		return fmt.Errorf("dist not found at %s", srcDist)
	}
	dstDist := filepath.Join(m.volumes.Path(volumeID), "dist")
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

// stopProcess 停止进程
func (m *Manager) stopProcess(pid int) error {
	if pid <= 0 {
		return nil
	}
	if !processExists(pid) {
		return nil
	}
	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		_ = proc.Kill()
	}
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()
	select {
	case <-done:
		return nil
	case <-time.After(30 * time.Second):
		_ = proc.Kill()
		<-done
		return fmt.Errorf("process killed after timeout")
	}
}

// processExists 检查进程是否存活
func processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	if runtime.GOOS == "windows" {
		out, _ := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/NH", "/FO", "CSV").Output()
		return strings.Contains(string(out), strconv.Itoa(pid))
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

// parseRef 从 name:tag 中提取 name 和 tag
func parseRef(ref string) (name, tag string) {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			return ref[:i], ref[i+1:]
		}
	}
	return ref, "latest"
}

// pmockerRoot 返回 PMocker 根目录。
// 支持通过 PMOCKER_HOME 环境变量自定义，默认为 ~/.pmocker
func pmockerRoot() (string, error) {
	if home := os.Getenv("PMOCKER_HOME"); home != "" {
		return home, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pmocker"), nil
}

// DefaultInstancesDir 返回默认实例目录 ~/.pmocker/instances
func DefaultInstancesDir() (string, error) {
	root, err := pmockerRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "instances"), nil
}

// DefaultVolumesDir 返回默认数据卷目录 ~/.pmocker/volumes
func DefaultVolumesDir() (string, error) {
	root, err := pmockerRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "volumes"), nil
}

// DefaultBinDir 返回默认二进制目录 ~/.pmocker/bin
func DefaultBinDir() (string, error) {
	root, err := pmockerRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "bin"), nil
}

// InitDefaultStore 初始化默认实例存储
func InitDefaultStore() (*SQLiteStore, error) {
	dir, err := DefaultInstancesDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return NewStore(filepath.Join(dir, "instances.db"))
}

// InitDefaultVolumes 初始化默认数据卷管理器
func InitDefaultVolumes() (*VolumeManager, error) {
	dir, err := DefaultVolumesDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return NewVolumeManager(dir), nil
}
