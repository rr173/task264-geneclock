package model

import "time"

// BehaviorVersion 行为版本：绑定线路定义快照与状态推断结果，冻结后不可改。
type BehaviorVersion struct {
	ID            string        `json:"id"`
	TrialID       string        `json:"trial_id"`
	Name          string        `json:"name"`
	Status        VersionStatus `json:"status"`
	// CircuitSnapshot 冻结时的线路定义快照（元件+阶段 JSON）。
	CircuitSnapshot string `json:"circuit_snapshot"`
	// StateSnapshot 冻结时的状态推断结果 JSON。
	StateSnapshot string `json:"state_snapshot"`
	// ConflictsSnapshot 冻结时的冲突清单 JSON。
	ConflictsSnapshot string    `json:"conflicts_snapshot"`
	CreatedAt     time.Time     `json:"created_at"`
	FrozenAt      *time.Time    `json:"frozen_at,omitempty"`
}

// VersionInput 创建行为版本入参。
type VersionInput struct {
	Name string `json:"name"`
}

// Validate 校验版本入参。
func (v *VersionInput) Validate() error {
	if v.Name == "" {
		return ErrValidation
	}
	return nil
}

// Stats 复核台统计。
type Stats struct {
	Circuits       int `json:"circuits"`
	Trials         int `json:"trials"`
	PendingTrials  int `json:"pending_trials"`
	PublishedVersions int `json:"published_versions"`
	OpenConflicts  int `json:"open_conflicts"`
}
