// Package constraint 检验表达先后顺序约束：上游元件必须先于下游关闭，
// 下游报告信号不得先于上游元件切换到关闭/抑制状态。
package constraint

import (
	"time"

	"github.com/google/uuid"

	"task264-geneclock/internal/model"
)

// Kind 常量：顺序冲突类别。
const (
	KindLateShutdown     = "late_shutdown"      // 上游未及时关闭
	KindEarlyActivation  = "early_activation"   // 下游提前激活
	KindInversion        = "inversion"          // 顺序反转
)

// ElementTiming 一次推断中某元件的时序信息。
type ElementTiming struct {
	ElementID   string
	ElementName string
	Upstream    string
	// StatePerStage stage name -> 推断状态。
	StatePerStage map[string]model.RegulationState
}

// CheckOrder 检查给定推断状态下元件间的先后约束。
// 规则：
//  1. 若元件 A 是元件 B 的上游（B.Upstream == A.ID），在任一阶段，B 状态为 active 而 A 状态为
//     uninduced/repressed 时，判 early_activation（下游先于上游激活）。
//  2. 诱导剂撤除阶段（期望 uninduced）若 A 仍 active（滞后窗口后仍未关闭），判 late_shutdown。
//  3. 若同一阶段内下游状态先于上游翻转（上游 active 而下游 repressed 且期望相反），判 inversion。
func CheckOrder(trialID string, timings []ElementTiming,
	stages []model.InductionStage, lag int) (*model.ConflictCheckResult, error) {

	now := time.Now().UTC()
	var conflicts []model.OrderConflict

	byID := map[string]*ElementTiming{}
	for i := range timings {
		t := timings[i]
		byID[t.ElementID] = &t
	}

	for i := range timings {
		cur := &timings[i]
		if cur.Upstream == "" {
			continue
		}
		up, ok := byID[cur.Upstream]
		if !ok {
			continue
		}
		for _, st := range stages {
			downState := cur.StatePerStage[st.Name]
			upState := up.StatePerStage[st.Name]
			// 规则 1：下游激活但上游未激活 → early_activation。
			if downState == model.StateActive && upState != model.StateActive {
				conflicts = append(conflicts, model.OrderConflict{
					ID:                uuid.NewString(),
					TrialID:           trialID,
					StageName:         st.Name,
					UpstreamElement:   up.ElementName,
					DownstreamElement: cur.ElementName,
					Kind:              KindEarlyActivation,
					Detail:            "下游元件在阶段 " + st.Name + " 已激活，而上游元件 " + up.ElementName + " 尚未激活（期望先上游后下游）",
					CreatedAt:         now,
				})
			}
			// 规则 2：撤除阶段上游仍 active → late_shutdown。
			if st.ExpectedState == model.StateUninduced && upState == model.StateActive {
				conflicts = append(conflicts, model.OrderConflict{
					ID:                uuid.NewString(),
					TrialID:           trialID,
					StageName:         st.Name,
					UpstreamElement:   up.ElementName,
					DownstreamElement: cur.ElementName,
					Kind:              KindLateShutdown,
					Detail:            "撤除阶段 " + st.Name + " 上游元件 " + up.ElementName + " 在滞后窗口后仍未关闭",
					CreatedAt:         now,
				})
			}
			// 规则 3：期望 active 但上游被抑制而下游激活 → inversion。
			if st.ExpectedState == model.StateActive && upState == model.StateRepressed && downState == model.StateActive {
				conflicts = append(conflicts, model.OrderConflict{
					ID:                uuid.NewString(),
					TrialID:           trialID,
					StageName:         st.Name,
					UpstreamElement:   up.ElementName,
					DownstreamElement: cur.ElementName,
					Kind:              KindInversion,
					Detail:            "阶段 " + st.Name + " 上游被抑制而下游激活，表达顺序反转",
					CreatedAt:         now,
				})
			}
		}
	}

	// 去重：同 (stage, up, down, kind) 只保留一条。
	seen := map[string]bool{}
	dedup := conflicts[:0]
	for _, c := range conflicts {
		key := c.StageName + "|" + c.UpstreamElement + "|" + c.DownstreamElement + "|" + c.Kind
		if seen[key] {
			continue
		}
		seen[key] = true
		dedup = append(dedup, c)
	}

	return &model.ConflictCheckResult{
		TrialID:   trialID,
		Conflicts: dedup,
		CheckedAt: now,
	}, nil
}
