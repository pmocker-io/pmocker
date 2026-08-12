package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestSqliteDSNConcurrencyPragmas 验证 Dsn() 生成的连接默认启用了
// busy_timeout(5000) 与 journal_mode(WAL) 并发写防护参数。
func TestSqliteDSNConcurrencyPragmas(t *testing.T) {
	dir := t.TempDir()
	s := Sqlite{}
	s.Path = dir
	s.Dbname = "concurrent"

	dsn := s.Dsn()
	if dsn == "" {
		t.Fatal("Dsn() 不应为空")
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	// busy_timeout 应为 5000ms
	var busy int
	if err := db.Raw("PRAGMA busy_timeout").Scan(&busy).Error; err != nil {
		t.Fatalf("读取 busy_timeout 失败: %v", err)
	}
	if busy != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busy)
	}

	// journal_mode 应为 wal
	var journal string
	if err := db.Raw("PRAGMA journal_mode").Scan(&journal).Error; err != nil {
		t.Fatalf("读取 journal_mode 失败: %v", err)
	}
	if journal != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journal)
	}

	// 触发一次写（WAL 模式下应产生 -wal 文件）
	if err := db.Exec("CREATE TABLE t(id INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	walFile := filepath.Join(dir, s.Dbname+".db-wal")
	if _, err := os.Stat(walFile); err != nil {
		t.Fatalf("WAL 模式应产生 -wal 文件: %v", err)
	}
}
