// Package store 提供 SQLite 持久化：建表迁移与各实体的读写。
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store 封装 SQLite 连接与全部仓储。
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // modernc sqlite 单写者
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close 关闭数据库。
func (s *Store) Close() error { return s.db.Close() }

// DB 暴露底层连接（供业务包做事务）。
func (s *Store) DB() *sql.DB { return s.db }

// NewCircuitStore 构造线路仓储。
func NewCircuitStore(db *sql.DB) *CircuitStore { return &CircuitStore{db: db} }

// NewTrialStore 构造试验仓储。
func NewTrialStore(db *sql.DB) *TrialStore { return &TrialStore{db: db} }

// NewAnalysisStore 构造分析仓储。
func NewAnalysisStore(db *sql.DB) *AnalysisStore { return &AnalysisStore{db: db} }

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS circuits (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS elements (
			id TEXT PRIMARY KEY,
			circuit_id TEXT NOT NULL REFERENCES circuits(id),
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			upstream TEXT NOT NULL DEFAULT '',
			basal_level INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			UNIQUE(circuit_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS induction_stages (
			id TEXT PRIMARY KEY,
			circuit_id TEXT NOT NULL REFERENCES circuits(id),
			name TEXT NOT NULL,
			inducer TEXT NOT NULL,
			start_seq INTEGER NOT NULL,
			end_seq INTEGER NOT NULL,
			expected_state TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS trials (
			id TEXT PRIMARY KEY,
			circuit_id TEXT NOT NULL REFERENCES circuits(id),
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			lag_points INTEGER NOT NULL DEFAULT 3,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			published_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS readings (
			id TEXT PRIMARY KEY,
			trial_id TEXT NOT NULL REFERENCES trials(id),
			seq INTEGER NOT NULL,
			channel TEXT NOT NULL,
			raw_value REAL NOT NULL,
			unit TEXT NOT NULL,
			window_status TEXT NOT NULL DEFAULT 'pending',
			created_at TEXT NOT NULL,
			UNIQUE(trial_id, seq, channel)
		)`,
		`CREATE TABLE IF NOT EXISTS baselines (
			id TEXT PRIMARY KEY,
			trial_id TEXT NOT NULL REFERENCES trials(id),
			channel TEXT NOT NULL,
			baseline_value REAL NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(trial_id, channel)
		)`,
		`CREATE TABLE IF NOT EXISTS element_states (
			trial_id TEXT NOT NULL,
			element_id TEXT NOT NULL,
			element_name TEXT NOT NULL,
			stage_name TEXT NOT NULL,
			stage_seq INTEGER NOT NULL,
			state TEXT NOT NULL,
			fold_change REAL NOT NULL,
			reason TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY(trial_id, element_id, stage_name)
		)`,
		`CREATE TABLE IF NOT EXISTS order_conflicts (
			id TEXT PRIMARY KEY,
			trial_id TEXT NOT NULL REFERENCES trials(id),
			stage_name TEXT NOT NULL,
			upstream_element TEXT NOT NULL,
			downstream_element TEXT NOT NULL,
			kind TEXT NOT NULL,
			detail TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS behavior_versions (
			id TEXT PRIMARY KEY,
			trial_id TEXT NOT NULL REFERENCES trials(id),
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			circuit_snapshot TEXT NOT NULL,
			state_snapshot TEXT NOT NULL,
			conflicts_snapshot TEXT NOT NULL,
			created_at TEXT NOT NULL,
			frozen_at TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
