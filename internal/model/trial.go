package model

import "time"

// Trial 线路试验：一次对特定线路的诱导-观测实验，状态机
// preparing → collecting → pending_review → published → archived。
type Trial struct {
	ID           string      `json:"id"`
	CircuitID    string      `json:"circuit_id"`
	Name         string      `json:"name"`
	Status       TrialStatus `json:"status"`
	LagPoints    int         `json:"lag_points"` // 滞后窗口（读数点数）
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	PublishedAt  *time.Time  `json:"published_at,omitempty"`
}

// Reading 荧光读数：单通道单时间点的原始强度，属读数窗口。
type Reading struct {
	ID            string       `json:"id"`
	TrialID       string       `json:"trial_id"`
	Seq           int          `json:"seq"`     // 读数序号（幂等键的一部分）
	Channel       string       `json:"channel"` // 通道名（如 gfp/od600）
	RawValue      float64      `json:"raw_value"`
	Unit          string       `json:"unit"`
	WindowStatus  WindowStatus `json:"window_status"` // 读数窗口状态
	CreatedAt     time.Time    `json:"created_at"`
}

// ReadingInput 批量导入读数条目。
type ReadingInput struct {
	Seq      int     `json:"seq"`
	Channel  string  `json:"channel"`
	RawValue float64 `json:"raw_value"`
	Unit     string  `json:"unit"`
}

// Validate 校验读数输入：单位缺失即拒绝。
func (r *ReadingInput) Validate() error {
	if r.Seq < 1 || r.Channel == "" || r.Unit == "" {
		return ErrUnitMissing
	}
	if r.RawValue < 0 {
		return ErrValidation
	}
	return nil
}

// TrialInput 创建试验入参。
type TrialInput struct {
	Name      string `json:"name"`
	LagPoints int    `json:"lag_points"`
}

// Validate 校验试验入参。
func (t *TrialInput) Validate() error {
	if t.Name == "" {
		return ErrValidation
	}
	if t.LagPoints < MinLagPoints || t.LagPoints > MaxLagPoints {
		return ErrValidation
	}
	return nil
}

// WindowUpdate 读数窗口状态更新入参。
type WindowUpdate struct {
	Status WindowStatus `json:"status"`
}

// Validate 校验窗口状态。
func (w *WindowUpdate) Validate() error {
	switch w.Status {
	case WindowValid, WindowSat, WindowMissing, WindowExcluded:
		return nil
	default:
		return ErrValidation
	}
}
