package instance

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type VolumeManager struct {
	base string
}

func NewVolumeManager(base string) *VolumeManager {
	return &VolumeManager{base: base}
}

func (v *VolumeManager) Create() (string, error) {
	volumeID := uuid.New().String()
	volPath := v.Path(volumeID)
	if err := os.MkdirAll(volPath, 0755); err != nil {
		return "", fmt.Errorf("create volume dir: %w", err)
	}
	for _, sub := range []string{"uploads"} {
		if err := os.MkdirAll(filepath.Join(volPath, sub), 0755); err != nil {
			return "", err
		}
	}
	return volumeID, nil
}

func (v *VolumeManager) Path(volumeID string) string {
	return filepath.Join(v.base, volumeID)
}

func (v *VolumeManager) SystemDBPath(volumeID string) string {
	return filepath.Join(v.Path(volumeID), "system.db")
}

func (v *VolumeManager) ProjectDBPath(volumeID string) string {
	return filepath.Join(v.Path(volumeID), "project.db")
}

func (v *VolumeManager) ConfigPath(volumeID string) string {
	return filepath.Join(v.Path(volumeID), "config.yaml")
}

func (v *VolumeManager) UploadsPath(volumeID string) string {
	return filepath.Join(v.Path(volumeID), "uploads")
}

func (v *VolumeManager) Remove(volumeID string) error {
	return os.RemoveAll(v.Path(volumeID))
}
