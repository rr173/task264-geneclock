package model

import "time"

// Baseline 报告基线：每个通道归一化所用的参考强度与 fold-change 结果。
type Baseline struct {
	ID            string    `json:"id"`
	TrialID       string    `json:"trial_id"`
	Channel       string    `json:"channel"`
	BaselineValue float64   `json:"baseline_value"` // 基线强度（报告基准）
	CreatedAt     time.Time `json:"created_at"`
}

// NormalizedReading 归一化后的读数：原始强度 / 基线 = fold change。
type NormalizedReading struct {
	ReadingID     string  `json:"reading_id"`
	Seq           int     `json:"seq"`
	Channel       string  `json:"channel"`
	RawValue      float64 `json:"raw_value"`
	BaselineValue float64 `json:"baseline_value"`
	FoldChange    float64 `json:"fold_change"`
	WindowStatus  WindowStatus `json:"window_status"`
}

// ElementState 单个元件的调控状态推断结果。
type ElementState struct {
	TrialID      string         `json:"trial_id"`
	ElementID    string         `json:"element_id"`
	ElementName  string         `json:"element_name"`
	StageName    string         `json:"stage_name"`
	StageSeq     int            `json:"stage_seq"` // 阶段起始序号
	State        RegulationState `json:"state"`
	FoldChange   float64        `json:"fold_change"`
	Reason       string         `json:"reason"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// OrderConflict 表达先后顺序冲突。
type OrderConflict struct {
	ID              string    `json:"id"`
	TrialID         string    `json:"trial_id"`
	StageName       string    `json:"stage_name"`
	UpstreamElement string    `json:"upstream_element"`
	DownstreamElement string  `json:"downstream_element"`
	// Kind 冲突类别：late_shutdown（上游未及时关闭）/ early_activation（下游提前激活）/ inversion（顺序反转）。
	Kind      string    `json:"kind"`
	Detail    string    `json:"detail"`
	CreatedAt time.Time `json:"created_at"`
}

// StateInferenceResult 一次推断的汇总。
type StateInferenceResult struct {
	TrialID    string         `json:"trial_id"`
	States     []ElementState `json:"states"`
	Generated  time.Time      `json:"generated_at"`
}

// ConflictCheckResult 一次顺序检查的汇总。
type ConflictCheckResult struct {
	TrialID   string           `json:"trial_id"`
	Conflicts []OrderConflict  `json:"conflicts"`
	CheckedAt time.Time        `json:"checked_at"`
}
