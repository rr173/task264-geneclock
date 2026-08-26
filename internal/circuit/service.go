// Package circuit 管理调控线路：元件录入、诱导阶段编排与重叠校验。
package circuit

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/store"
)

// Service 线路领域服务。
type Service struct {
	circuits *store.CircuitStore
}

// NewService 构造线路服务。
func NewService(circuits *store.CircuitStore) *Service {
	return &Service{circuits: circuits}
}

// CreateCircuit 创建线路并返回完整实体。
func (s *Service) CreateCircuit(in *model.CircuitInput) (*model.Circuit, error) {
	if in.Name == "" {
		return nil, model.ErrValidation
	}
	now := time.Now().UTC()
	circ := &model.Circuit{
		ID:          uuid.NewString(),
		Name:        in.Name,
		Description: in.Description,
		Status:      model.CircuitActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.circuits.CreateCircuit(circ); err != nil {
		return nil, err
	}
	return circ, nil
}

// Get 读取线路。
func (s *Service) Get(id string) (*model.Circuit, error) { return s.circuits.GetCircuit(id) }

// List 列出全部线路。
func (s *Service) List() ([]model.Circuit, error) { return s.circuits.ListCircuits() }

// Archive 封存线路：封存后不允许再添加元件/阶段。
func (s *Service) Archive(id string) (*model.Circuit, error) {
	circ, err := s.circuits.GetCircuit(id)
	if err != nil {
		return nil, err
	}
	if circ.Status == model.CircuitArchived {
		return nil, model.ErrCircuitArchived
	}
	if err := s.circuits.UpdateCircuitStatus(id, model.CircuitArchived); err != nil {
		return nil, err
	}
	circ.Status = model.CircuitArchived
	return circ, nil
}

// AddElement 添加元件：校验线路存在且未封存、元件输入合法、上游存在。
func (s *Service) AddElement(circuitID string, in *model.ElementInput) (*model.Element, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	circ, err := s.circuits.GetCircuit(circuitID)
	if err != nil {
		return nil, err
	}
	if circ.Status == model.CircuitArchived {
		return nil, model.ErrCircuitArchived
	}
	if in.Upstream != "" {
		if _, err := s.circuits.GetElement(in.Upstream); err != nil {
			return nil, model.ErrUnknownElement
		}
	}
	e := &model.Element{
		ID:         uuid.NewString(),
		CircuitID:  circuitID,
		Name:       in.Name,
		Kind:       in.Kind,
		Upstream:   in.Upstream,
		BasalLevel: in.BasalLevel,
		CreatedAt:  time.Now().UTC(),
	}
	if err := s.circuits.AddElement(e); err != nil {
		return nil, err
	}
	return e, nil
}

// ListElements 列出线路元件。
func (s *Service) ListElements(circuitID string) ([]model.Element, error) {
	return s.circuits.ListElements(circuitID)
}

// AddStage 添加诱导阶段：与既有阶段区间重叠时拒绝。
func (s *Service) AddStage(circuitID string, in *model.StageInput) (*model.InductionStage, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	circ, err := s.circuits.GetCircuit(circuitID)
	if err != nil {
		return nil, err
	}
	if circ.Status == model.CircuitArchived {
		return nil, model.ErrCircuitArchived
	}
	existing, err := s.circuits.ListStages(circuitID)
	if err != nil {
		return nil, err
	}
	st := &model.InductionStage{
		ID:            uuid.NewString(),
		CircuitID:     circuitID,
		Name:          in.Name,
		Inducer:       in.Inducer,
		StartSeq:      in.StartSeq,
		EndSeq:        in.EndSeq,
		ExpectedState: in.ExpectedState,
		CreatedAt:     time.Now().UTC(),
	}
	for _, ex := range existing {
		if st.Overlaps(&ex) {
			return nil, model.ErrStageOverlap
		}
	}
	if err := s.circuits.AddStage(st); err != nil {
		return nil, err
	}
	return st, nil
}

// ListStages 列出线路诱导阶段。
func (s *Service) ListStages(circuitID string) ([]model.InductionStage, error) {
	return s.circuits.ListStages(circuitID)
}

// StageForSeq 找到覆盖某读数序号的诱导阶段（含边界）。
func (s *Service) StageForSeq(circuitID string, seq int) (*model.InductionStage, error) {
	stages, err := s.circuits.ListStages(circuitID)
	if err != nil {
		return nil, err
	}
	for i := range stages {
		if seq >= stages[i].StartSeq && seq <= stages[i].EndSeq {
			return &stages[i], nil
		}
	}
	return nil, model.ErrNotFound
}

// Snapshot 序列化线路定义快照（元件 + 阶段），供版本冻结使用。
func (s *Service) Snapshot(circuitID string) (string, error) {
	elements, err := s.circuits.ListElements(circuitID)
	if err != nil {
		return "", err
	}
	stages, err := s.circuits.ListStages(circuitID)
	if err != nil {
		return "", err
	}
	type snap struct {
		CircuitID string                   `json:"circuit_id"`
		Elements  []model.Element          `json:"elements"`
		Stages    []model.InductionStage   `json:"stages"`
		At        string                   `json:"at"`
	}
	data, err := json.Marshal(snap{
		CircuitID: circuitID,
		Elements:  elements,
		Stages:    stages,
		At:        time.Now().UTC().Format(time.RFC3339),
	})
	_ = circuitID
	if err != nil {
		return "", fmt.Errorf("snapshot: %w", err)
	}
	return string(data), nil
}
