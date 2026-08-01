// Package instance 管理 PMSystem 实例生命周期。
// store.go 实现 SQLite 实例注册表。
package instance

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Instance 实例元数据
type Instance struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	ImageDigest string     `json:"imageDigest"`
	ImageRef    string     `json:"imageRef"`
	Port        int        `json:"port"`
	VolumeID    string     `json:"volumeId"`
	PID         int        `json:"pid"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
	StoppedAt   *time.Time `json:"stoppedAt,omitempty"`
}

// Store 实例注册表接口
type Store interface {
	Create(inst *Instance) error
	GetByID(id string) (*Instance, error)
	GetByName(name string) (*Instance, error)
	List(includeStopped bool) ([]*Instance, error)
	Update(inst *Instance) error
	Delete(id string) error
	Close() error
}

// SQLiteStore SQLite 实现
type SQLiteStore struct {
	db *sql.DB
}

// NewStore 创建实例注册表
func NewStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %s: %w", dbPath, err)
	}
	// 限制并发连接，避免 SQLite 写锁冲突
	db.SetMaxOpenConns(1)
	s := &SQLiteStore{db: db}
	// 设置 busy_timeout，避免写锁竞争时立即返回错误
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy_timeout: %w", err)
	}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS instances (
  id           TEXT PRIMARY KEY,
  name         TEXT UNIQUE NOT NULL,
  image_digest TEXT NOT NULL,
  image_ref    TEXT,
  port         INTEGER NOT NULL,
  volume_id    TEXT NOT NULL,
  pid          INTEGER DEFAULT 0,
  status       TEXT DEFAULT 'stopped',
  created_at   DATETIME NOT NULL,
  started_at   DATETIME,
  stopped_at   DATETIME
)`)
	return err
}

func (s *SQLiteStore) Create(inst *Instance) error {
	_, err := s.db.Exec(
		`INSERT INTO instances (id, name, image_digest, image_ref, port, volume_id, pid, status, created_at, started_at, stopped_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		inst.ID, inst.Name, inst.ImageDigest, inst.ImageRef, inst.Port,
		inst.VolumeID, inst.PID, inst.Status, inst.CreatedAt, inst.StartedAt, inst.StoppedAt,
	)
	if err != nil {
		return fmt.Errorf("create instance: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetByID(id string) (*Instance, error) {
	return s.getOne("SELECT id, name, image_digest, image_ref, port, volume_id, pid, status, created_at, started_at, stopped_at FROM instances WHERE id = ?", id)
}

func (s *SQLiteStore) GetByName(name string) (*Instance, error) {
	return s.getOne("SELECT id, name, image_digest, image_ref, port, volume_id, pid, status, created_at, started_at, stopped_at FROM instances WHERE name = ?", name)
}

func (s *SQLiteStore) GetByPID(pid int) (*Instance, error) {
	return s.getOne("SELECT id, name, image_digest, image_ref, port, volume_id, pid, status, created_at, started_at, stopped_at FROM instances WHERE pid = ?", pid)
}

func (s *SQLiteStore) List(includeStopped bool) ([]*Instance, error) {
	q := "SELECT id, name, image_digest, image_ref, port, volume_id, pid, status, created_at, started_at, stopped_at FROM instances"
	if !includeStopped {
		q += " WHERE status = 'running'"
	}
	q += " ORDER BY created_at DESC"
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*Instance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, inst)
	}
	return list, nil
}

func (s *SQLiteStore) Update(inst *Instance) error {
	_, err := s.db.Exec(
		`UPDATE instances SET name=?, image_digest=?, image_ref=?, port=?, volume_id=?, pid=?, status=?, started_at=?, stopped_at=? WHERE id=?`,
		inst.Name, inst.ImageDigest, inst.ImageRef, inst.Port, inst.VolumeID,
		inst.PID, inst.Status, inst.StartedAt, inst.StoppedAt, inst.ID,
	)
	return err
}

func (s *SQLiteStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM instances WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) getOne(query string, args ...interface{}) (*Instance, error) {
	row := s.db.QueryRow(query, args...)
	inst, err := scanInstance(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("instance not found")
	}
	return inst, err
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanInstance(s scannable) (*Instance, error) {
	inst := &Instance{}
	var startedAt, stoppedAt sql.NullTime
	err := s.Scan(
		&inst.ID, &inst.Name, &inst.ImageDigest, &inst.ImageRef,
		&inst.Port, &inst.VolumeID, &inst.PID, &inst.Status,
		&inst.CreatedAt, &startedAt, &stoppedAt,
	)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		inst.StartedAt = &startedAt.Time
	}
	if stoppedAt.Valid {
		inst.StoppedAt = &stoppedAt.Time
	}
	return inst, nil
}
