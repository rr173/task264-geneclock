// Package normalize 负责报告基线归一化：以诱导前基线强度为分母计算 fold-change。
package normalize

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/store"
)

func sortBySeq(list []model.Reading) {
	sort.Slice(list, func(i, j int) bool { return list[i].Seq < list[j].Seq })
}

// Service 归一化领域服务。
type Service struct {
	trials *store.TrialStore
}

// NewService 构造归一化服务。
func NewService(trials *store.TrialStore) *Service {
	return &Service{trials: trials}
}

// ChannelBaselines 计算各通道报告基线：取该通道「最早 lag 个有效读数」的均值
// （诱导前窗口基础表达水平）。若已保存过基线则直接沿用。
func (s *Service) ChannelBaselines(trialID string, lag int) (map[string]float64, error) {
	readings, err := s.trials.ListReadings(trialID)
	if err != nil {
		return nil, err
	}
	baselines := map[string]float64{}
	if existing, err := s.trials.GetBaselines(trialID); err == nil {
		for _, b := range existing {
			baselines[b.Channel] = b.BaselineValue
		}
	}
	// 每通道按 seq 排序的诱导前窗口（最早 lag 个点）。
	byChannel := map[string][]model.Reading{}
	for _, r := range readings {
		if r.WindowStatus == model.WindowSat || r.WindowStatus == model.WindowExcluded {
			continue
		}
		byChannel[r.Channel] = append(byChannel[r.Channel], r)
	}
	for ch, list := range byChannel {
		sortBySeq(list)
		if lag < 1 {
			lag = 1
		}
		if lag > len(list) {
			lag = len(list)
		}
		var sum float64
		for i := 0; i < lag; i++ {
			sum += list[i].RawValue
		}
		mean := sum / float64(lag)
		if _, ok := baselines[ch]; !ok {
			baselines[ch] = mean
		}
	}
	return baselines, nil
}

// PersistBaselines 保存基线并返回每通道归一化结果。
func (s *Service) PersistBaselines(trialID string, lag int) (map[string]float64, error) {
	baselines, err := s.ChannelBaselines(trialID, lag)
	if err != nil {
		return nil, err
	}
	if len(baselines) == 0 {
		return nil, model.ErrBaselineNotReady
	}
	for ch, val := range baselines {
		b := &model.Baseline{
			ID:            uuid.NewString(),
			TrialID:       trialID,
			Channel:       ch,
			BaselineValue: val,
			CreatedAt:     time.Now().UTC(),
		}
		if err := s.trials.SaveBaseline(b); err != nil {
			return nil, err
		}
	}
	return baselines, nil
}

// NormalizeReadings 按保存的基线归一化全部读数。
func (s *Service) NormalizeReadings(trialID string) ([]model.NormalizedReading, error) {
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
	var out []model.NormalizedReading
	for _, r := range readings {
		bv, ok := base[r.Channel]
		if !ok || bv <= 0 {
			continue
		}
		out = append(out, model.NormalizedReading{
			ReadingID:     r.ID,
			Seq:           r.Seq,
			Channel:       r.Channel,
			RawValue:      r.RawValue,
			BaselineValue: bv,
			FoldChange:    fold(r.RawValue, bv),
			WindowStatus:  r.WindowStatus,
		})
	}
	return out, nil
}

func fold(raw, base float64) float64 {
	if base == 0 {
		return 0
	}
	v := raw / base
	// 保留 3 位小数，避免浮点噪声。
	return float64(int(v*1000+0.5)) / 1000
}

// AverageFoldByChannel 计算某通道在指定序号区间的平均 fold-change。
func (s *Service) AverageFoldByChannel(trialID, channel string, startSeq, endSeq int, lag int) (float64, error) {
	norm, err := s.NormalizeReadings(trialID)
	if err != nil {
		return 0, err
	}
	var sum float64
	var n int
	for _, r := range norm {
		if r.Channel != channel {
			continue
		}
		if r.WindowStatus == model.WindowSat || r.WindowStatus == model.WindowExcluded {
			continue
		}
		// 滞后窗口：阶段起点后 lag 个点才计入。
		if r.Seq < startSeq+lag || r.Seq > endSeq {
			continue
		}
		sum += r.FoldChange
		n++
	}
	if n == 0 {
		return 0, fmt.Errorf("%w: 通道 %s 区间内无有效读数", model.ErrValidation, channel)
	}
	return sum / float64(n), nil
}
