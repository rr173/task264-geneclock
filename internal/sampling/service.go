// Package sampling 接收多通道荧光读数：幂等导入、窗口状态流转。
package sampling

import (
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/store"
)

// Service 采样领域服务。
type Service struct {
	trials *store.TrialStore
}

// NewService 构造采样服务。
func NewService(trials *store.TrialStore) *Service {
	return &Service{trials: trials}
}

// ImportReadings 批量导入读数：逐条幂等写入，(trial, seq, channel) 重复则整体失败。
func (s *Service) ImportReadings(trialID string, inputs []model.ReadingInput) ([]model.Reading, error) {
	tr, err := s.trials.GetTrial(trialID)
	if err != nil {
		return nil, err
	}
	if tr.Status != model.TrialCollecting && tr.Status != model.TrialPreparing {
		return nil, model.ErrTrialNotReady
	}
	var out []model.Reading
	seen := map[string]bool{}
	for _, in := range inputs {
		if err := in.Validate(); err != nil {
			return nil, err
		}
		key := in.Channel + ":" + itoa(in.Seq)
		if seen[key] {
			return nil, model.ErrDuplicate
		}
		seen[key] = true
		r := &model.Reading{
			ID:           uuid.NewString(),
			TrialID:      trialID,
			Seq:          in.Seq,
			Channel:      in.Channel,
			RawValue:     in.RawValue,
			Unit:         in.Unit,
			WindowStatus: model.WindowPending,
			CreatedAt:    time.Now().UTC(),
		}
		if err := s.trials.AddReading(r); err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, nil
}

// ListReadings 列出试验全部读数。
func (s *Service) ListReadings(trialID string) ([]model.Reading, error) {
	return s.trials.ListReadings(trialID)
}

// UpdateWindow 更新读数窗口状态；校验流转合法性（pending→valid/saturated/...）。
func (s *Service) UpdateWindow(readingID string, in *model.WindowUpdate) (*model.Reading, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	r, err := s.trials.GetReading(readingID)
	if err != nil {
		return nil, err
	}
	if !model.CanTransitionWindow(r.WindowStatus, in.Status) {
		return nil, model.ErrBadTransition
	}
	if err := s.trials.UpdateWindowStatus(readingID, in.Status); err != nil {
		return nil, err
	}
	r.WindowStatus = in.Status
	return r, nil
}

// Channels 汇总试验出现的全部通道。
func (s *Service) Channels(trialID string) ([]string, error) {
	readings, err := s.trials.ListReadings(trialID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []string
	for _, r := range readings {
		if !seen[r.Channel] {
			seen[r.Channel] = true
			out = append(out, r.Channel)
		}
	}
	return out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
