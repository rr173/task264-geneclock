package store

import (
	"database/sql"
	"errors"
	"fmt"

	"task264-geneclock/internal/model"
)

// TrialStore 线路试验与荧光读数的仓储。
type TrialStore struct{ db *sql.DB }

// CreateTrial 创建试验。
func (t *TrialStore) CreateTrial(tr *model.Trial) error {
	_, err := t.db.Exec(
		`INSERT INTO trials(id,circuit_id,name,status,lag_points,created_at,updated_at,published_at) VALUES(?,?,?,?,?,?,?,?)`,
		tr.ID, tr.CircuitID, tr.Name, tr.Status, tr.LagPoints, nowStr(), nowStr(), nil)
	if err != nil {
		return fmt.Errorf("insert trial: %w", err)
	}
	return nil
}

// GetTrial 读取试验。
func (t *TrialStore) GetTrial(id string) (*model.Trial, error) {
	row := t.db.QueryRow(`SELECT id,circuit_id,name,status,lag_points,created_at,updated_at,published_at FROM trials WHERE id=?`, id)
	var tr model.Trial
	var created, updated string
	var published sql.NullString
	if err := row.Scan(&tr.ID, &tr.CircuitID, &tr.Name, &tr.Status, &tr.LagPoints, &created, &updated, &published); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	tr.CreatedAt, _ = parseTime(created)
	tr.UpdatedAt, _ = parseTime(updated)
	if published.Valid {
		if pt, err := parseTime(published.String); err == nil {
			tr.PublishedAt = &pt
		}
	}
	return &tr, nil
}

// ListTrials 列出线路下全部试验。
func (t *TrialStore) ListTrials(circuitID string) ([]model.Trial, error) {
	rows, err := t.db.Query(
		`SELECT id,circuit_id,name,status,lag_points,created_at,updated_at,published_at FROM trials WHERE circuit_id=? ORDER BY created_at`,
		circuitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Trial
	for rows.Next() {
		var tr model.Trial
		var created, updated string
		var published sql.NullString
		if err := rows.Scan(&tr.ID, &tr.CircuitID, &tr.Name, &tr.Status, &tr.LagPoints, &created, &updated, &published); err != nil {
			return nil, err
		}
		tr.CreatedAt, _ = parseTime(created)
		tr.UpdatedAt, _ = parseTime(updated)
		if published.Valid {
			if pt, err := parseTime(published.String); err == nil {
				tr.PublishedAt = &pt
			}
		}
		out = append(out, tr)
	}
	return out, rows.Err()
}

// UpdateTrialStatus 更新试验状态。
func (t *TrialStore) UpdateTrialStatus(id string, status model.TrialStatus, publishedAt *string) error {
	_, err := t.db.Exec(`UPDATE trials SET status=?, updated_at=?, published_at=? WHERE id=?`,
		status, nowStr(), publishedAt, id)
	if err != nil {
		return err
	}
	return nil
}

// UpdateTrialLag 更新滞后窗口。
func (t *TrialStore) UpdateTrialLag(id string, lag int) error {
	_, err := t.db.Exec(`UPDATE trials SET lag_points=?, updated_at=? WHERE id=?`, lag, nowStr(), id)
	if err != nil {
		return err
	}
	return nil
}

// AddReading 幂等写入读数：同一 (trial, seq, channel) 重复返回重复错误。
func (t *TrialStore) AddReading(r *model.Reading) error {
	_, err := t.db.Exec(
		`INSERT INTO readings(id,trial_id,seq,channel,raw_value,unit,window_status,created_at) VALUES(?,?,?,?,?,?,?,?)`,
		r.ID, r.TrialID, r.Seq, r.Channel, r.RawValue, r.Unit, r.WindowStatus, nowStr())
	if err != nil {
		if isUniqueViolation(err) {
			return model.ErrDuplicate
		}
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}

// ListReadings 列出试验全部读数。
func (t *TrialStore) ListReadings(trialID string) ([]model.Reading, error) {
	rows, err := t.db.Query(
		`SELECT id,trial_id,seq,channel,raw_value,unit,window_status,created_at FROM readings WHERE trial_id=? ORDER BY seq,channel`,
		trialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Reading
	for rows.Next() {
		var r model.Reading
		var created string
		if err := rows.Scan(&r.ID, &r.TrialID, &r.Seq, &r.Channel, &r.RawValue, &r.Unit, &r.WindowStatus, &created); err != nil {
			return nil, err
		}
		r.CreatedAt, _ = parseTime(created)
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetReading 读取单个读数。
func (t *TrialStore) GetReading(id string) (*model.Reading, error) {
	row := t.db.QueryRow(`SELECT id,trial_id,seq,channel,raw_value,unit,window_status,created_at FROM readings WHERE id=?`, id)
	var r model.Reading
	var created string
	if err := row.Scan(&r.ID, &r.TrialID, &r.Seq, &r.Channel, &r.RawValue, &r.Unit, &r.WindowStatus, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	r.CreatedAt, _ = parseTime(created)
	return &r, nil
}

// UpdateWindowStatus 更新读数窗口状态。
func (t *TrialStore) UpdateWindowStatus(id string, status model.WindowStatus) error {
	res, err := t.db.Exec(`UPDATE readings SET window_status=? WHERE id=?`, status, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// SaveBaseline 幂等保存通道基线。
func (t *TrialStore) SaveBaseline(b *model.Baseline) error {
	_, err := t.db.Exec(
		`INSERT INTO baselines(id,trial_id,channel,baseline_value,created_at) VALUES(?,?,?,?,?)
		 ON CONFLICT(trial_id,channel) DO UPDATE SET baseline_value=excluded.baseline_value`,
		b.ID, b.TrialID, b.Channel, b.BaselineValue, nowStr())
	if err != nil {
		return fmt.Errorf("save baseline: %w", err)
	}
	return nil
}

// GetBaselines 读取试验全部通道基线。
func (t *TrialStore) GetBaselines(trialID string) ([]model.Baseline, error) {
	rows, err := t.db.Query(
		`SELECT id,trial_id,channel,baseline_value,created_at FROM baselines WHERE trial_id=? ORDER BY channel`,
		trialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Baseline
	for rows.Next() {
		var b model.Baseline
		var created string
		if err := rows.Scan(&b.ID, &b.TrialID, &b.Channel, &b.BaselineValue, &created); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = parseTime(created)
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBaselineByChannel 读取单通道基线。
func (t *TrialStore) GetBaselineByChannel(trialID, channel string) (*model.Baseline, error) {
	row := t.db.QueryRow(`SELECT id,trial_id,channel,baseline_value,created_at FROM baselines WHERE trial_id=? AND channel=?`,
		trialID, channel)
	var b model.Baseline
	var created string
	if err := row.Scan(&b.ID, &b.TrialID, &b.Channel, &b.BaselineValue, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	b.CreatedAt, _ = parseTime(created)
	return &b, nil
}
