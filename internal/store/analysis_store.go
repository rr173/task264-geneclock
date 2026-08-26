package store

import (
	"database/sql"
	"errors"
	"fmt"

	"task264-geneclock/internal/model"
)

// AnalysisStore 调控状态推断与顺序冲突的仓储。
type AnalysisStore struct{ db *sql.DB }

// SaveElementStates 覆盖保存某试验的全部元件状态。
func (a *AnalysisStore) SaveElementStates(trialID string, states []model.ElementState) error {
	for _, st := range states {
		if _, err := a.db.Exec(
			`INSERT INTO element_states(trial_id,element_id,element_name,stage_name,stage_seq,state,fold_change,reason,updated_at)
			 VALUES(?,?,?,?,?,?,?,?,?)`,
			st.TrialID, st.ElementID, st.ElementName, st.StageName, st.StageSeq,
			st.State, st.FoldChange, st.Reason, nowStr()); err != nil {
			return fmt.Errorf("insert element_state: %w", err)
		}
	}
	return nil
}

// ListElementStates 列出试验全部元件状态。
func (a *AnalysisStore) ListElementStates(trialID string) ([]model.ElementState, error) {
	rows, err := a.db.Query(
		`SELECT trial_id,element_id,element_name,stage_name,stage_seq,state,fold_change,reason,updated_at
		 FROM element_states WHERE trial_id=? ORDER BY stage_seq,element_name`,
		trialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.ElementState
	for rows.Next() {
		var st model.ElementState
		var updated string
		if err := rows.Scan(&st.TrialID, &st.ElementID, &st.ElementName, &st.StageName, &st.StageSeq,
			&st.State, &st.FoldChange, &st.Reason, &updated); err != nil {
			return nil, err
		}
		st.UpdatedAt, _ = parseTime(updated)
		out = append(out, st)
	}
	return out, rows.Err()
}

// HasElementStates 判断试验是否已有状态推断结果。
func (a *AnalysisStore) HasElementStates(trialID string) (bool, error) {
	var n int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM element_states WHERE trial_id=?`, trialID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// ClearConflicts 清空试验冲突清单。
func (a *AnalysisStore) ClearConflicts(trialID string) error {
	_, err := a.db.Exec(`DELETE FROM order_conflicts WHERE trial_id=?`, trialID)
	return err
}

// SaveConflicts 保存顺序冲突清单。
func (a *AnalysisStore) SaveConflicts(conflicts []model.OrderConflict) error {
	for _, c := range conflicts {
		if _, err := a.db.Exec(
			`INSERT INTO order_conflicts(id,trial_id,stage_name,upstream_element,downstream_element,kind,detail,created_at)
			 VALUES(?,?,?,?,?,?,?,?)`,
			c.ID, c.TrialID, c.StageName, c.UpstreamElement, c.DownstreamElement, c.Kind, c.Detail, nowStr()); err != nil {
			return fmt.Errorf("insert conflict: %w", err)
		}
	}
	return nil
}

// ListConflicts 列出试验全部顺序冲突。
func (a *AnalysisStore) ListConflicts(trialID string) ([]model.OrderConflict, error) {
	rows, err := a.db.Query(
		`SELECT id,trial_id,stage_name,upstream_element,downstream_element,kind,detail,created_at
		 FROM order_conflicts WHERE trial_id=? ORDER BY created_at`,
		trialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.OrderConflict
	for rows.Next() {
		var c model.OrderConflict
		var created string
		if err := rows.Scan(&c.ID, &c.TrialID, &c.StageName, &c.UpstreamElement, &c.DownstreamElement,
			&c.Kind, &c.Detail, &created); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = parseTime(created)
		out = append(out, c)
	}
	return out, rows.Err()
}

// HasConflicts 判断试验是否已执行顺序检查。
func (a *AnalysisStore) HasConflicts(trialID string) (bool, error) {
	var n int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM order_conflicts WHERE trial_id=?`, trialID).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountConflicts 统计试验未处理冲突数（有冲突行即计）。
func (a *AnalysisStore) CountConflicts(trialID string) (int, error) {
	var n int
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM order_conflicts WHERE trial_id=?`, trialID).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// UpsertVersion 创建或更新行为版本。
func (a *AnalysisStore) UpsertVersion(v *model.BehaviorVersion) error {
	frozen := ""
	if v.FrozenAt != nil {
		frozen = v.FrozenAt.UTC().Format(timeLayout)
	}
	_, err := a.db.Exec(
		`INSERT INTO behavior_versions(id,trial_id,name,status,circuit_snapshot,state_snapshot,conflicts_snapshot,created_at,frozen_at)
		 VALUES(?,?,?,?,?,?,?,?,?)
		 ON CONFLICT(id) DO UPDATE SET status=excluded.status, frozen_at=excluded.frozen_at`,
		v.ID, v.TrialID, v.Name, v.Status, v.CircuitSnapshot, v.StateSnapshot, v.ConflictsSnapshot,
		nowStr(), frozen)
	if err != nil {
		return fmt.Errorf("upsert version: %w", err)
	}
	return nil
}

// GetVersion 读取行为版本。
func (a *AnalysisStore) GetVersion(id string) (*model.BehaviorVersion, error) {
	row := a.db.QueryRow(
		`SELECT id,trial_id,name,status,circuit_snapshot,state_snapshot,conflicts_snapshot,created_at,frozen_at
		 FROM behavior_versions WHERE id=?`, id)
	var v model.BehaviorVersion
	var created string
	var frozen sql.NullString
	if err := row.Scan(&v.ID, &v.TrialID, &v.Name, &v.Status, &v.CircuitSnapshot, &v.StateSnapshot,
		&v.ConflictsSnapshot, &created, &frozen); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, err
	}
	v.CreatedAt, _ = parseTime(created)
	if frozen.Valid {
		if ft, err := parseTime(frozen.String); err == nil {
			v.FrozenAt = &ft
		}
	}
	return &v, nil
}

// ListVersions 列出试验全部行为版本。
func (a *AnalysisStore) ListVersions(trialID string) ([]model.BehaviorVersion, error) {
	rows, err := a.db.Query(
		`SELECT id,trial_id,name,status,circuit_snapshot,state_snapshot,conflicts_snapshot,created_at,frozen_at
		 FROM behavior_versions WHERE trial_id=? ORDER BY created_at`,
		trialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.BehaviorVersion
	for rows.Next() {
		var v model.BehaviorVersion
		var created string
		var frozen sql.NullString
		if err := rows.Scan(&v.ID, &v.TrialID, &v.Name, &v.Status, &v.CircuitSnapshot, &v.StateSnapshot,
			&v.ConflictsSnapshot, &created, &frozen); err != nil {
			return nil, err
		}
		v.CreatedAt, _ = parseTime(created)
		if frozen.Valid {
			if ft, err := parseTime(frozen.String); err == nil {
				v.FrozenAt = &ft
			}
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// UpdateVersionStatus 更新版本状态。
func (a *AnalysisStore) UpdateVersionStatus(id string, status model.VersionStatus, frozenAt *string) error {
	_, err := a.db.Exec(`UPDATE behavior_versions SET status=?, frozen_at=? WHERE id=?`, status, frozenAt, id)
	if err != nil {
		return err
	}
	return nil
}

// Stats 统计概览。
func (a *AnalysisStore) Stats() (*model.Stats, error) {
	st := &model.Stats{}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM circuits`).Scan(&st.Circuits); err != nil {
		return nil, err
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM trials`).Scan(&st.Trials); err != nil {
		return nil, err
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM trials WHERE status=?`, model.TrialPendingReview).Scan(&st.PendingTrials); err != nil {
		return nil, err
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM behavior_versions WHERE status=?`, model.VersionFrozen).Scan(&st.PublishedVersions); err != nil {
		return nil, err
	}
	if err := a.db.QueryRow(`SELECT COUNT(*) FROM order_conflicts`).Scan(&st.OpenConflicts); err != nil {
		return nil, err
	}
	return st, nil
}
