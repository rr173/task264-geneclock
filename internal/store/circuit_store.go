package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"task264-geneclock/internal/model"
)

const timeLayout = "2006-01-02T15:04:05.000Z07:00"

func nowStr() string { return time.Now().UTC().Format(timeLayout) }

func parseTime(s string) (time.Time, error) {
	return time.Parse(timeLayout, s)
}

// CircuitStore 线路及元件/诱导阶段的仓储。
type CircuitStore struct{ db *sql.DB }

// CreateCircuit 创建线路。
func (c *CircuitStore) CreateCircuit(circ *model.Circuit) error {
	_, err := c.db.Exec(
		`INSERT INTO circuits(id,name,description,status,created_at,updated_at) VALUES(?,?,?,?,?,?)`,
		circ.ID, circ.Name, circ.Description, circ.Status, nowStr(), nowStr())
	if err != nil {
		return fmt.Errorf("insert circuit: %w", err)
	}
	return nil
}

// GetCircuit 读取线路。
func (c *CircuitStore) GetCircuit(id string) (*model.Circuit, error) {
	row := c.db.QueryRow(`SELECT id,name,description,status,created_at,updated_at FROM circuits WHERE id=?`, id)
	var circ model.Circuit
	var created, updated string
	if err := row.Scan(&circ.ID, &circ.Name, &circ.Description, &circ.Status, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	circ.CreatedAt, _ = parseTime(created)
	circ.UpdatedAt, _ = parseTime(updated)
	return &circ, nil
}

// ListCircuits 列出全部线路。
func (c *CircuitStore) ListCircuits() ([]model.Circuit, error) {
	rows, err := c.db.Query(`SELECT id,name,description,status,created_at,updated_at FROM circuits ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Circuit
	for rows.Next() {
		var circ model.Circuit
		var created, updated string
		if err := rows.Scan(&circ.ID, &circ.Name, &circ.Description, &circ.Status, &created, &updated); err != nil {
			return nil, err
		}
		circ.CreatedAt, _ = parseTime(created)
		circ.UpdatedAt, _ = parseTime(updated)
		out = append(out, circ)
	}
	return out, rows.Err()
}

// UpdateCircuitStatus 更新线路状态并刷新 updated_at。
func (c *CircuitStore) UpdateCircuitStatus(id string, status model.CircuitStatus) error {
	res, err := c.db.Exec(`UPDATE circuits SET status=?, updated_at=? WHERE id=?`, status, nowStr(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// AddElement 添加元件；同名元件返回重复错误。
func (c *CircuitStore) AddElement(e *model.Element) error {
	_, err := c.db.Exec(
		`INSERT INTO elements(id,circuit_id,name,kind,upstream,basal_level,created_at) VALUES(?,?,?,?,?,?,?)`,
		e.ID, e.CircuitID, e.Name, e.Kind, e.Upstream, e.BasalLevel, nowStr())
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert element: %w", err)
	}
	return nil
}

// ListElements 列出线路全部元件。
func (c *CircuitStore) ListElements(circuitID string) ([]model.Element, error) {
	rows, err := c.db.Query(
		`SELECT id,circuit_id,name,kind,upstream,basal_level,created_at FROM elements WHERE circuit_id=? ORDER BY created_at`,
		circuitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Element
	for rows.Next() {
		var e model.Element
		var created string
		if err := rows.Scan(&e.ID, &e.CircuitID, &e.Name, &e.Kind, &e.Upstream, &e.BasalLevel, &created); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = parseTime(created)
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetElement 读取单个元件。
func (c *CircuitStore) GetElement(id string) (*model.Element, error) {
	row := c.db.QueryRow(`SELECT id,circuit_id,name,kind,upstream,basal_level,created_at FROM elements WHERE id=?`, id)
	var e model.Element
	var created string
	if err := row.Scan(&e.ID, &e.CircuitID, &e.Name, &e.Kind, &e.Upstream, &e.BasalLevel, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	e.CreatedAt, _ = parseTime(created)
	return &e, nil
}

// AddStage 添加诱导阶段；区间重叠返回冲突错误。
func (c *CircuitStore) AddStage(s *model.InductionStage) error {
	_, err := c.db.Exec(
		`INSERT INTO induction_stages(id,circuit_id,name,inducer,start_seq,end_seq,expected_state,created_at) VALUES(?,?,?,?,?,?,?,?)`,
		s.ID, s.CircuitID, s.Name, s.Inducer, s.StartSeq, s.EndSeq, s.ExpectedState, nowStr())
	if err != nil {
		return fmt.Errorf("insert stage: %w", err)
	}
	return nil
}

// ListStages 列出线路全部诱导阶段。
func (c *CircuitStore) ListStages(circuitID string) ([]model.InductionStage, error) {
	rows, err := c.db.Query(
		`SELECT id,circuit_id,name,inducer,start_seq,end_seq,expected_state,created_at FROM induction_stages WHERE circuit_id=? ORDER BY start_seq`,
		circuitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.InductionStage
	for rows.Next() {
		var s model.InductionStage
		var created string
		if err := rows.Scan(&s.ID, &s.CircuitID, &s.Name, &s.Inducer, &s.StartSeq, &s.EndSeq, &s.ExpectedState, &created); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = parseTime(created)
		out = append(out, s)
	}
	return out, rows.Err()
}

// UpdateStage 更新诱导阶段（区间端点/期望状态）。
func (c *CircuitStore) UpdateStage(id string, s *model.InductionStage) error {
	res, err := c.db.Exec(
		`UPDATE induction_stages SET name=?,inducer=?,start_seq=?,end_seq=?,expected_state=? WHERE id=?`,
		s.Name, s.Inducer, s.StartSeq, s.EndSeq, s.ExpectedState, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "UNIQUE") || contains(err.Error(), "unique constraint"))
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
