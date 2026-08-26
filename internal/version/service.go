// Package version 管理行为版本：快照绑定、共享、冻结与替代。
package version

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/store"
)

// Service 行为版本服务。
type Service struct {
	analysis *store.AnalysisStore
}

// NewService 构造版本服务。
func NewService(analysis *store.AnalysisStore) *Service {
	return &Service{analysis: analysis}
}

// SnapshotBundle 冻结版本所需的三个快照。
type SnapshotBundle struct {
	Circuit   string
	States    string
	Conflicts string
}

// Create 创建草稿版本：绑定试验当前推断与冲突快照。
func (s *Service) Create(trialID string, in *model.VersionInput, bundle SnapshotBundle) (*model.BehaviorVersion, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	v := &model.BehaviorVersion{
		ID:                uuid.NewString(),
		TrialID:           trialID,
		Name:              in.Name,
		Status:            model.VersionDraft,
		CircuitSnapshot:   bundle.Circuit,
		StateSnapshot:     bundle.States,
		ConflictsSnapshot: bundle.Conflicts,
		CreatedAt:         now,
	}
	if err := s.analysis.UpsertVersion(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Share 共享版本：draft → shared。
func (s *Service) Share(id string) (*model.BehaviorVersion, error) {
	v, err := s.analysis.GetVersion(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionVersion(v.Status, model.VersionShared) {
		return nil, model.ErrBadTransition
	}
	if err := s.analysis.UpdateVersionStatus(id, model.VersionShared, nil); err != nil {
		return nil, err
	}
	v.Status = model.VersionShared
	return v, nil
}

// Freeze 冻结版本：shared → frozen；冻结后不可修改。
func (s *Service) Freeze(id string) (*model.BehaviorVersion, error) {
	v, err := s.analysis.GetVersion(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionVersion(v.Status, model.VersionFrozen) {
		return nil, model.ErrBadTransition
	}
	now := time.Now().UTC()
	nowStr := now.Format("2006-01-02T15:04:05.000Z07:00")
	if err := s.analysis.UpdateVersionStatus(id, model.VersionFrozen, &nowStr); err != nil {
		return nil, err
	}
	v.Status = model.VersionFrozen
	v.FrozenAt = &now
	return v, nil
}

// Supersede 替代版本：frozen/shared → superseded。
func (s *Service) Supersede(id string) (*model.BehaviorVersion, error) {
	v, err := s.analysis.GetVersion(id)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionVersion(v.Status, model.VersionSuperseded) {
		return nil, model.ErrBadTransition
	}
	if err := s.analysis.UpdateVersionStatus(id, model.VersionSuperseded, nil); err != nil {
		return nil, err
	}
	v.Status = model.VersionSuperseded
	return v, nil
}

// Get 读取版本。
func (s *Service) Get(id string) (*model.BehaviorVersion, error) {
	return s.analysis.GetVersion(id)
}

// List 列出试验全部版本。
func (s *Service) List(trialID string) ([]model.BehaviorVersion, error) {
	return s.analysis.ListVersions(trialID)
}

// MarshalSnapshots 将结构体快照序列化为 JSON 字符串。
func MarshalSnapshots(circuit any, states any, conflicts any) (SnapshotBundle, error) {
	c, err := json.Marshal(circuit)
	if err != nil {
		return SnapshotBundle{}, err
	}
	st, err := json.Marshal(states)
	if err != nil {
		return SnapshotBundle{}, err
	}
	cf, err := json.Marshal(conflicts)
	if err != nil {
		return SnapshotBundle{}, err
	}
	return SnapshotBundle{Circuit: string(c), States: string(st), Conflicts: string(cf)}, nil
}
