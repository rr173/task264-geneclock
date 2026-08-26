package service

import (
	"task264-geneclock/internal/constraint"
	"task264-geneclock/internal/model"
)

// NormalizeBaseline 归一化并持久化基线，返回每通道基线值。
func (a *App) NormalizeBaseline(trialID string) (map[string]float64, error) {
	tr, err := a.TrialStore.GetTrial(trialID)
	if err != nil {
		return nil, err
	}
	if tr.Status == model.TrialPreparing {
		return nil, model.ErrTrialNotReady
	}
	return a.Normalize.PersistBaselines(trialID, tr.LagPoints)
}

// ListNormalized 查看归一化后的全部读数。
func (a *App) ListNormalized(trialID string) ([]model.NormalizedReading, error) {
	return a.Normalize.NormalizeReadings(trialID)
}

// InferRegulationStates 推断元件开关状态并持久化。
func (a *App) InferRegulationStates(trialID string) (*model.StateInferenceResult, error) {
	tr, err := a.TrialStore.GetTrial(trialID)
	if err != nil {
		return nil, err
	}
	circ, err := a.CircuitStore.GetCircuit(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	if circ.Status == model.CircuitArchived {
		return nil, model.ErrCircuitArchived
	}
	elements, err := a.CircuitStore.ListElements(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	stages, err := a.CircuitStore.ListStages(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	if len(elements) == 0 || len(stages) == 0 {
		return nil, model.ErrValidation
	}
	res, err := a.Infer.InferStates(trialID, elements, stages, channelOfElement, tr.LagPoints)
	if err != nil {
		return nil, err
	}
	if err := a.Analysis.SaveElementStates(trialID, res.States); err != nil {
		return nil, err
	}
	return res, nil
}

// CheckOrderConstraints 执行顺序约束检查并持久化冲突。
func (a *App) CheckOrderConstraints(trialID string) (*model.ConflictCheckResult, error) {
	tr, err := a.TrialStore.GetTrial(trialID)
	if err != nil {
		return nil, err
	}
	elements, err := a.CircuitStore.ListElements(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	stages, err := a.CircuitStore.ListStages(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	states, err := a.Analysis.ListElementStates(trialID)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, model.ErrNoStates
	}
	// 从持久化状态恢复各元件在各阶段的推断状态。
	stateByElem := map[string]map[string]model.RegulationState{}
	for _, st := range states {
		if stateByElem[st.ElementID] == nil {
			stateByElem[st.ElementID] = map[string]model.RegulationState{}
		}
		stateByElem[st.ElementID][st.StageName] = st.State
	}
	var timings []constraint.ElementTiming
	for _, e := range elements {
		timings = append(timings, constraint.ElementTiming{
			ElementID:     e.ID,
			ElementName:   e.Name,
			Upstream:      e.Upstream,
			StatePerStage: stateByElem[e.ID],
		})
	}
	res, err := constraint.CheckOrder(trialID, timings, stages, tr.LagPoints)
	if err != nil {
		return nil, err
	}
	if err := a.Analysis.ClearConflicts(trialID); err != nil {
		return nil, err
	}
	if err := a.Analysis.SaveConflicts(res.Conflicts); err != nil {
		return nil, err
	}
	return res, nil
}

// ListElementStates 查看试验的元件状态推断。
func (a *App) ListElementStates(trialID string) ([]model.ElementState, error) {
	return a.Analysis.ListElementStates(trialID)
}

// ListConflicts 查看试验的顺序冲突。
func (a *App) ListConflicts(trialID string) ([]model.OrderConflict, error) {
	return a.Analysis.ListConflicts(trialID)
}

// channelOfElement 决定某元件用哪个荧光通道判定状态：
// 报告/感应元件用同名的通道；抑制元件用其上游通道（观测下游被压制）。
func channelOfElement(e model.Element) (string, bool) {
	if e.Kind == model.ElementReporter || e.Kind == model.ElementSensor {
		return e.Name, true
	}
	if e.Kind == model.ElementPromoter {
		return e.Name, true
	}
	// 抑制元件没有独立报告通道，跳过（其状态由下游报告体现）。
	return "", false
}
