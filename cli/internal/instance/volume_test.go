package instance

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVolumeManager(t *testing.T) {
	tmpDir := t.TempDir()
	vm := NewVolumeManager(tmpDir)

	volID, err := vm.Create()
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if volID == "" {
		t.Fatal("empty volumeID")
	}

	volPath := vm.Path(volID)
	if _, err := os.Stat(volPath); os.IsNotExist(err) {
		t.Error("volume dir not created")
	}

	uploadsPath := vm.UploadsPath(volID)
	if _, err := os.Stat(uploadsPath); os.IsNotExist(err) {
		t.Error("uploads dir not created")
	}

	sysDB := vm.SystemDBPath(volID)
	if filepath.Base(sysDB) != "system.db" {
		t.Errorf("systemDB base = %s", filepath.Base(sysDB))
	}

	if err := vm.Remove(volID); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(volPath); !os.IsNotExist(err) {
		t.Error("volume dir still exists after remove")
	}
}
