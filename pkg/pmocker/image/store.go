// Package image 管理 PMocker 本地镜像缓存（~/.pmocker/images/）。
// 镜像按 digest 寻址存储，同时维护 name:tag -> digest 的索引。
package image

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

// ImageInfo 镜像信息
type ImageInfo struct {
	Digest    string       `json:"digest"`
	Name      string       `json:"name"`
	Tag       string       `json:"tag"`
	Size      int64        `json:"size"`
	CreatedAt string       `json:"createdAt"`
	Config    oci.Config   `json:"config"`
	Manifest  oci.Manifest `json:"manifest"`
}

// indexFile 镜像索引文件（name:tag -> digest 映射）
type indexFile struct {
	Entries []indexEntry `json:"entries"`
}

type indexEntry struct {
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Digest    string `json:"digest"`
	CreatedAt string `json:"createdAt"`
}

// Store 镜像缓存存储
type Store struct {
	baseDir   string // ~/.pmocker/images
	indexPath string // ~/.pmocker/images/index.json
}

// NewStore 创建镜像存储，baseDir 通常为 ~/.pmocker/images
func NewStore(baseDir string) *Store {
	return &Store{
		baseDir:   baseDir,
		indexPath: filepath.Join(baseDir, "index.json"),
	}
}

// DefaultStoreDir 返回默认存储路径 ~/.pmocker/images
func DefaultStoreDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".pmocker", "images"), nil
}

// AddImage 将 .pmi 文件添加到本地缓存
func (s *Store) AddImage(pmiPath, name, tag string) (*ImageInfo, error) {
	reader, err := oci.OpenImage(pmiPath)
	if err != nil {
		return nil, err
	}
	manifest := reader.Manifest()
	// 用 manifest 中 config 的 digest 作为镜像 digest
	digest := manifest.Config.Digest
	if digest == "" {
		return nil, fmt.Errorf("image has no config digest")
	}

	// 创建 digest 目录
	imgDir := filepath.Join(s.baseDir, strings.ReplaceAll(digest, ":", "_"))
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", imgDir, err)
	}

	// 复制 .pmi 文件
	dstPath := filepath.Join(imgDir, "image.pmi")
	if err := copyFile(pmiPath, dstPath); err != nil {
		return nil, err
	}

	// 保存 manifest 和 config
	manBytes, _ := oci.JSONMarshal(manifest)
	os.WriteFile(filepath.Join(imgDir, "manifest.json"), manBytes, 0644)
	cfgBytes, _ := oci.JSONMarshal(reader.Config())
	os.WriteFile(filepath.Join(imgDir, "config.json"), cfgBytes, 0644)

	// 更新索引
	info := &ImageInfo{
		Digest:    digest,
		Name:      name,
		Tag:       tag,
		Size:      manifest.Config.Size,
		CreatedAt: manifest.Annotations["pmocker.image.created"],
		Config:    reader.Config(),
		Manifest:  manifest,
	}
	if err := s.updateIndex(name, tag, digest); err != nil {
		return nil, err
	}
	return info, nil
}

// ListImages 列出所有本地镜像
func (s *Store) ListImages() ([]ImageInfo, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	infos := make([]ImageInfo, 0, len(idx.Entries))
	for _, e := range idx.Entries {
		info, err := s.GetImage(e.Digest)
		if err != nil {
			continue
		}
		info.Name = e.Name
		info.Tag = e.Tag
		info.CreatedAt = e.CreatedAt
		infos = append(infos, *info)
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})
	return infos, nil
}

// GetImage 按 digest 获取镜像信息
func (s *Store) GetImage(digest string) (*ImageInfo, error) {
	imgDir := filepath.Join(s.baseDir, strings.ReplaceAll(digest, ":", "_"))
	manPath := filepath.Join(imgDir, "manifest.json")
	cfgPath := filepath.Join(imgDir, "config.json")
	manBytes, err := os.ReadFile(manPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var man oci.Manifest
	json.Unmarshal(manBytes, &man)
	var cfg oci.Config
	json.Unmarshal(cfgBytes, &cfg)
	// 计算 .pmi 文件大小
	pmiPath := filepath.Join(imgDir, "image.pmi")
	size := int64(0)
	if fi, err := os.Stat(pmiPath); err == nil {
		size = fi.Size()
	}
	return &ImageInfo{
		Digest:   digest,
		Size:     size,
		Config:   cfg,
		Manifest: man,
	}, nil
}

// ResolveImage 按名称和标签解析镜像
func (s *Store) ResolveImage(name, tag string) (*ImageInfo, error) {
	idx, err := s.loadIndex()
	if err != nil {
		return nil, err
	}
	for _, e := range idx.Entries {
		if e.Name == name && (tag == "" || e.Tag == tag) {
			return s.GetImage(e.Digest)
		}
	}
	return nil, fmt.Errorf("image %s:%s not found", name, tag)
}

// RemoveImage 删除本地镜像
func (s *Store) RemoveImage(digest string) error {
	imgDir := filepath.Join(s.baseDir, strings.ReplaceAll(digest, ":", "_"))
	if err := os.RemoveAll(imgDir); err != nil {
		return fmt.Errorf("remove %s: %w", imgDir, err)
	}
	return s.removeFromIndex(digest)
}

func (s *Store) loadIndex() (*indexFile, error) {
	data, err := os.ReadFile(s.indexPath)
	if os.IsNotExist(err) {
		return &indexFile{}, nil
	}
	if err != nil {
		return nil, err
	}
	var idx indexFile
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func (s *Store) saveIndex(idx *indexFile) error {
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.indexPath, data, 0644)
}

func (s *Store) updateIndex(name, tag, digest string) error {
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	// 移除同名旧条目
	entries := idx.Entries[:0]
	for _, e := range idx.Entries {
		if e.Name != name || e.Tag != tag {
			entries = append(entries, e)
		}
	}
	entries = append(entries, indexEntry{Name: name, Tag: tag, Digest: digest, CreatedAt: nowRFC3339()})
	idx.Entries = entries
	return s.saveIndex(idx)
}

func (s *Store) removeFromIndex(digest string) error {
	idx, err := s.loadIndex()
	if err != nil {
		return err
	}
	entries := idx.Entries[:0]
	for _, e := range idx.Entries {
		if e.Digest != digest {
			entries = append(entries, e)
		}
	}
	idx.Entries = entries
	return s.saveIndex(idx)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

func nowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}
