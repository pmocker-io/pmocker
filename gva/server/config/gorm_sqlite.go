package config

import (
	"fmt"
	"path/filepath"
)

type Sqlite struct {
	GeneralDB `yaml:",inline" mapstructure:",squash"`
}

// Dsn 返回 SQLite 连接 DSN（追加并发写防护参数）
// busy_timeout: 写锁等待 5s，避免并发写立即 SQLITE_BUSY
// journal_mode(WAL): 读写并发不互斥，且异常退出不会残留 rollback journal
func (s *Sqlite) Dsn() string {
	return fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", filepath.Join(s.Path, s.Dbname+".db"))
}
