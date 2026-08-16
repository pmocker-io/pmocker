package instance

import (
	"archive/tar"
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// SnapshotVolume 把实例数据卷打包为 tar 快照（data 层用）。
// 包含：system.db（WAL checkpoint 后）、dist/、uploads/。
// 输出 tar 字节供 oci.CreateLayerFromFiles / 直接写文件使用。
//
// sqliteCheckpoint=true 时先对 system.db 执行 PRAGMA wal_checkpoint(TRUNCATE)，
// 把 -wal 内容合并回主库，确保快照一致性（实例运行中也安全，WAL 支持多进程）。
func SnapshotVolume(volumePath string, sqliteCheckpoint bool) ([]byte, error) {
	if sqliteCheckpoint {
		if err := checkpointSQLite(filepath.Join(volumePath, "system.db")); err != nil {
			return nil, fmt.Errorf("sqlite checkpoint: %w", err)
		}
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	addFile := func(rel string) error {
		full := filepath.Join(volumePath, rel)
		info, err := os.Stat(full)
		if err != nil {
			if os.IsNotExist(err) {
				return nil // 缺失文件跳过（如 uploads 可能为空但目录在）
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(full)
		if err != nil {
			return err
		}
		hdr := &tar.Header{Name: filepath.ToSlash(rel), Mode: 0644, Size: int64(len(data))}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	}

	// system.db（必选）
	if _, err := os.Stat(filepath.Join(volumePath, "system.db")); os.IsNotExist(err) {
		tw.Close()
		return nil, fmt.Errorf("system.db 不存在，无法快照: %s", volumePath)
	}
	if err := addFile("system.db"); err != nil {
		tw.Close()
		return nil, err
	}
	// dist/（前端产物，整体打入）
	if err := addTree(tw, filepath.Join(volumePath, "dist"), "dist"); err != nil {
		tw.Close()
		return nil, err
	}
	// uploads/（交付物文件）
	if err := addTree(tw, filepath.Join(volumePath, "uploads"), "uploads"); err != nil {
		tw.Close()
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// checkpointSQLite 对 sqlite 执行 wal_checkpoint(TRUNCATE) 合并 -wal 到主库。
// 若库非 WAL 模式或不可打开，返回 nil（尽力而为，不强失败）。
func checkpointSQLite(dbPath string) error {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return fmt.Errorf("system.db 不存在: %s", dbPath)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	var result string
	if err := db.QueryRow("PRAGMA wal_checkpoint(TRUNCATE)").Scan(&result); err != nil {
		// 库非 WAL 模式时 PRAGMA 返回单行单列，Scan 可能因列数不符报错，忽略
		return nil
	}
	_ = result
	return nil
}

// addTree 递归把目录下所有文件加入 tar（保留相对前缀 prefix）。
func addTree(tw *tar.Writer, dir, prefix string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(filepath.Join(prefix, rel))
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(data))}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
}
