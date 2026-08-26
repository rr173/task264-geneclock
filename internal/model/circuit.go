package model

import "time"

// Circuit 调控线路：一组调控元件与诱导阶段的集合，是复核台的第一层主体。
type Circuit struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      CircuitStatus  `json:"status"`
	Elements    []Element      `json:"elements,omitempty"`
	Stages      []InductionStage `json:"stages,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// Element 调控元件：启动子/抑制/报告/感应，带表达方向与期望基态。
type Element struct {
	ID        string      `json:"id"`
	CircuitID string      `json:"circuit_id"`
	Name      string      `json:"name"`
	Kind      ElementKind `json:"kind"`
	// Upstream 上游元件 ID；空串表示无上游（输入源）。
	Upstream string `json:"upstream"`
	// BasalLevel 未诱导时的期望基态（0 表示低表达，1 表示高表达）。
	BasalLevel int       `json:"basal_level"`
	CreatedAt  time.Time `json:"created_at"`
}

// InductionStage 诱导阶段：描述诱导剂从 start_seq 到 end_seq 的施加区间。
type InductionStage struct {
	ID        string    `json:"id"`
	CircuitID string    `json:"circuit_id"`
	Name      string    `json:"name"`
	Inducer   string    `json:"inducer"`
	// StartSeq 起始读数序号（含）。
	StartSeq int `json:"start_seq"`
	// EndSeq 结束读数序号（含）。
	EndSeq int `json:"end_seq"`
	// ExpectedState 该阶段预期达到的调控状态。
	ExpectedState RegulationState `json:"expected_state"`
	CreatedAt     time.Time       `json:"created_at"`
}

// CircuitInput 创建线路的入参。
type CircuitInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ElementInput 添加元件的入参。
type ElementInput struct {
	Name      string      `json:"name"`
	Kind      ElementKind `json:"kind"`
	Upstream  string      `json:"upstream"`
	BasalLevel int        `json:"basal_level"`
}

// StageInput 添加诱导阶段的入参。
type StageInput struct {
	Name          string          `json:"name"`
	Inducer       string          `json:"inducer"`
	StartSeq      int             `json:"start_seq"`
	EndSeq        int             `json:"end_seq"`
	ExpectedState RegulationState `json:"expected_state"`
}

// Validate 校验元件输入。
func (e *ElementInput) Validate() error {
	if e.Name == "" {
		return ErrValidation
	}
	switch e.Kind {
	case ElementPromoter, ElementRepressor, ElementReporter, ElementSensor:
	default:
		return ErrValidation
	}
	if e.BasalLevel != 0 && e.BasalLevel != 1 {
		return ErrValidation
	}
	return nil
}

// Validate 校验诱导阶段输入。
func (s *StageInput) Validate() error {
	if s.Name == "" || s.Inducer == "" {
		return ErrValidation
	}
	if s.StartSeq < 1 || s.EndSeq < s.StartSeq {
		return ErrValidation
	}
	switch s.ExpectedState {
	case StateUninduced, StateActive, StateRepressed:
	default:
		return ErrValidation
	}
	return nil
}

// Overlaps 判断两个诱导阶段区间是否重叠。
func (s *InductionStage) Overlaps(o *InductionStage) bool {
	return s.StartSeq <= o.EndSeq && o.StartSeq <= s.EndSeq
}
