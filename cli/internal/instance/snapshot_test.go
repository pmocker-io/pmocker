package instance

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestSnapshotVolume 快照应包含 system.db、dist/、uploads/ 的文件
func TestSnapshotVolume(t *testing.T) {
	volPath := t.TempDir()

	// 构造数据卷
	if err := os.WriteFile(filepath.Join(volPath, "system.db"), []byte("fake-sqlite"), 0644); err != nil {
		t.Fatal(err)
	}
	distDir := filepath.Join(volPath, "dist", "assets")
	if err := os.MkdirAll(distDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "app.js"), []byte("console.log(1)"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(volPath, "dist", "index.html"), []byte("<html></html>"), 0644); err != nil {
		t.Fatal(err)
	}
	uploadsDir := filepath.Join(volPath, "uploads")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploadsDir, "a.pdf"), []byte("%PDF"), 0644); err != nil {
		t.Fatal(err)
	}
	// 干扰文件（不应被打包）
	if err := os.WriteFile(filepath.Join(volPath, "gva-server.log"), []byte("log"), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := SnapshotVolume(volPath, false)
	if err != nil {
		t.Fatalf("SnapshotVolume: %v", err)
	}

	// 解析 tar 校验
	files := map[string]bool{}
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		files[hdr.Name] = true
	}

	for _, want := range []string{"system.db", "dist/index.html", "dist/assets/app.js", "uploads/a.pdf"} {
		if !files[want] {
			t.Errorf("快照缺少 %s", want)
		}
	}
	if files["gva-server.log"] {
		t.Error("不应打包 gva-server.log")
	}
}

// TestSnapshotVolumeMissingDB system.db 缺失应报错
func TestSnapshotVolumeMissingDB(t *testing.T) {
	volPath := t.TempDir()
	if _, err := SnapshotVolume(volPath, false); err == nil {
		t.Fatal("缺 system.db 应报错")
	}
}

// TestSnapshotVolumeEmptySubdirs 缺 dist/uploads 目录应跳过而非报错
func TestSnapshotVolumeEmptySubdirs(t *testing.T) {
	volPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(volPath, "system.db"), []byte("db"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := SnapshotVolume(volPath, false); err != nil {
		t.Fatalf("缺 dist/uploads 不应报错: %v", err)
	}
}
