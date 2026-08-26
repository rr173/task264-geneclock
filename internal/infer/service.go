// Package infer 根据归一化 fold-change 推断调控元件开关状态。
package infer

import (
	"time"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/store"
)

// Service 状态推断服务。
type Service struct {
	trials *store.TrialStore
}

// NewService 构造推断服务。
func NewService(trials *store.TrialStore) *Service {
	return &Service{trials: trials}
}

// InferStates 对试验的每个元件在每个诱导阶段推断调控状态。
// 依据：通道归一化 fold-change 与激活/抑制阈值比较；阶段按期望状态给出 reason。
// lag 为滞后窗口：阶段起点后 lag 个读数点才计入统计（表达响应延迟）。
func (s *Service) InferStates(trialID string, elements []model.Element, stages []model.InductionStage,
	channelOf func(element model.Element) (string, bool), lag int) (*model.StateInferenceResult, error) {

	readings, err := s.trials.ListReadings(trialID)
	if err != nil {
		return nil, err
	}
	baselines, err := s.trials.GetBaselines(trialID)
	if err != nil {
		return nil, err
	}
	base := map[string]float64{}
	for _, b := range baselines {
		base[b.Channel] = b.BaselineValue
	}
	// channel -> seq -> fold
	folds := map[string]map[int]float64{}
	for _, r := range readings {
		bv, ok := base[r.Channel]
		if !ok || bv <= 0 {
			continue
		}
		if folds[r.Channel] == nil {
			folds[r.Channel] = map[int]float64{}
		}
		folds[r.Channel][r.Seq] = fold(r.RawValue, bv)
	}

	now := time.Now().UTC()
	var states []model.ElementState
	for _, elem := range elements {
		ch, ok := channelOf(elem)
		if !ok {
			continue
		}
		for _, st := range stages {
			avg, n := avgFold(folds[ch], st.StartSeq+lag, st.EndSeq)
			if n == 0 {
				continue
			}
			state, reason := classify(avg, st.ExpectedState)
			states = append(states, model.ElementState{
				TrialID:     trialID,
				ElementID:   elem.ID,
				ElementName: elem.Name,
				StageName:   st.Name,
				StageSeq:    st.StartSeq,
				State:       state,
				FoldChange:  avg,
				Reason:      reason,
				UpdatedAt:   now,
			})
		}
	}
	return &model.StateInferenceResult{TrialID: trialID, States: states, Generated: now}, nil
}

func avgFold(m map[int]float64, start, end int) (float64, int) {
	var sum float64
	var n int
	for seq, v := range m {
		if seq >= start && seq <= end {
			sum += v
			n++
		}
	}
	if n == 0 {
		return 0, 0
	}
	return sum / float64(n), n
}

// classify 结合阈值与期望状态给出判定。
func classify(avg float64, expected model.RegulationState) (model.RegulationState, string) {
	switch {
	case avg >= model.ActivationFold:
		if expected == model.StateRepressed {
			return model.StateOrderFlaw, "期望抑制但 fold-change 达激活阈值，疑似顺序冲突"
		}
		return model.StateActive, "fold-change 达到激活阈值"
	case avg <= model.RepressionFold:
		if expected == model.StateActive {
			return model.StateOrderFlaw, "期望激活但 fold-change 低于抑制阈值，疑似顺序冲突"
		}
		return model.StateRepressed, "fold-change 低于抑制阈值"
	default:
		return model.StateUninduced, "fold-change 处于未诱导区间"
	}
}

func fold(raw, base float64) float64 {
	if base == 0 {
		return 0
	}
	v := raw / base
	return float64(int(v*1000+0.5)) / 1000
}
