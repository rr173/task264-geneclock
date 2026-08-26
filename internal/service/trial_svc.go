package service

import (
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
)

// CreateTrial 在指定线路下创建试验。
func (a *App) CreateTrial(circuitID string, in *model.TrialInput) (*model.Trial, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	circ, err := a.CircuitStore.GetCircuit(circuitID)
	if err != nil {
		return nil, err
	}
	if circ.Status == model.CircuitArchived {
		return nil, model.ErrCircuitArchived
	}
	now := time.Now().UTC()
	tr := &model.Trial{
		ID:        uuid.NewString(),
		CircuitID: circuitID,
		Name:      in.Name,
		Status:    model.TrialPreparing,
		LagPoints: in.LagPoints,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.TrialStore.CreateTrial(tr); err != nil {
		return nil, err
	}
	return tr, nil
}

// GetTrial 读取试验。
func (a *App) GetTrial(id string) (*model.Trial, error) { return a.TrialStore.GetTrial(id) }

// ListTrials 列出线路试验。
func (a *App) ListTrials(circuitID string) ([]model.Trial, error) {
	return a.TrialStore.ListTrials(circuitID)
}

// TransitionTrial 状态流转：preparing→collecting→pending_review→published/archived。
func (a *App) TransitionTrial(id string, to model.TrialStatus) (*model.Trial, error) {
	tr, err := a.TrialStore.GetTrial(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionTrial(tr.Status, to) {
		return nil, model.ErrBadTransition
	}
	var published *string
	if to == model.TrialPublished {
		s := time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")
		published = &s
	}
	if err := a.TrialStore.UpdateTrialStatus(id, to, published); err != nil {
		return nil, err
	}
	tr.Status = to
	tr.UpdatedAt = time.Now().UTC()
	if to == model.TrialPublished {
		pt := time.Now().UTC()
		tr.PublishedAt = &pt
	}
	return tr, nil
}

// SetLagWindow 调整滞后窗口（读数点数），影响顺序约束判定。
func (a *App) SetLagWindow(id string, lag int) (*model.Trial, error) {
	if lag < model.MinLagPoints || lag > model.MaxLagPoints {
		return nil, model.ErrValidation
	}
	tr, err := a.TrialStore.GetTrial(id)
	if err != nil {
		return nil, err
	}
	if err := a.TrialStore.UpdateTrialLag(id, lag); err != nil {
		return nil, err
	}
	tr.LagPoints = lag
	return tr, nil
}

// Stats 统计概览。
func (a *App) Stats() (*model.Stats, error) { return a.Analysis.Stats() }
