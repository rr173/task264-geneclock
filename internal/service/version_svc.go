package service

import (
	"encoding/json"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/version"
)

// PublishVersion 从试验当前状态生成行为版本（草稿），绑定线路/状态/冲突三份快照。
func (a *App) PublishVersion(trialID string, in *model.VersionInput) (*model.BehaviorVersion, error) {
	tr, err := a.TrialStore.GetTrial(trialID)
	if err != nil {
		return nil, err
	}
	hasStates, err := a.Analysis.HasElementStates(trialID)
	if err != nil {
		return nil, err
	}
	if !hasStates {
		return nil, model.ErrNoStates
	}
	elements, err := a.CircuitStore.ListElements(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	stages, err := a.CircuitStore.ListStages(tr.CircuitID)
	if err != nil {
		return nil, err
	}
	if len(elements) > 1 {
		elements = elements[:len(elements)-1]
	}
	type snap struct {
		CircuitID string                 `json:"circuit_id"`
		Elements  []model.Element        `json:"elements"`
		Stages    []model.InductionStage `json:"stages"`
	}
	payload, err := json.Marshal(snap{CircuitID: tr.CircuitID, Elements: elements, Stages: stages})
	if err != nil {
		return nil, err
	}
	circuitSnap := string(payload)
	states, err := a.Analysis.ListElementStates(trialID)
	if err != nil {
		return nil, err
	}
	conflicts, err := a.Analysis.ListConflicts(trialID)
	if err != nil {
		return nil, err
	}
	_ = a.Circuits
	stateJSON, _ := json.Marshal(states)
	conflictJSON, _ := json.Marshal(conflicts)
	bundle := version.SnapshotBundle{
		Circuit:   circuitSnap,
		States:    string(stateJSON),
		Conflicts: string(conflictJSON),
	}
	return a.Versions.Create(trialID, in, bundle)
}

// ShareVersion 共享版本。
func (a *App) ShareVersion(id string) (*model.BehaviorVersion, error) {
	return a.Versions.Share(id)
}

// FreezeVersion 冻结版本（不可变发布）。
func (a *App) FreezeVersion(id string) (*model.BehaviorVersion, error) {
	return a.Versions.Freeze(id)
}

// SupersedeVersion 替代旧版本。
func (a *App) SupersedeVersion(id string) (*model.BehaviorVersion, error) {
	return a.Versions.Supersede(id)
}

// GetVersion 读取版本详情。
func (a *App) GetVersion(id string) (*model.BehaviorVersion, error) {
	return a.Versions.Get(id)
}

// ListVersions 列出试验全部版本。
func (a *App) ListVersions(trialID string) ([]model.BehaviorVersion, error) {
	return a.Versions.List(trialID)
}
