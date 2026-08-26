package constraint

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestCheckOrderLateShutdown(t *testing.T) {
	// 上游 sensor 撤除阶段仍未关闭（active），下游 gfp 已回落 → late_shutdown。
	timings := []ElementTiming{
		{
			ElementID:   "up",
			ElementName: "sensor",
			Upstream:    "",
			StatePerStage: map[string]model.RegulationState{
				"诱导激活": model.StateActive,
				"撤除恢复": model.StateActive, // 上游未及时关闭
			},
		},
		{
			ElementID:   "down",
			ElementName: "gfp",
			Upstream:    "up",
			StatePerStage: map[string]model.RegulationState{
				"诱导激活": model.StateActive,
				"撤除恢复": model.StateUninduced, // 下游已关闭
			},
		},
	}
	stages := []model.InductionStage{
		{Name: "诱导激活", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive},
		{Name: "撤除恢复", StartSeq: 5, EndSeq: 8, ExpectedState: model.StateUninduced},
	}
	res, err := CheckOrder("t1", timings, stages, 2)
	if err != nil {
		t.Fatalf("CheckOrder: %v", err)
	}
	found := false
	for _, c := range res.Conflicts {
		if c.Kind == KindLateShutdown {
			found = true
			if c.UpstreamElement != "sensor" || c.DownstreamElement != "gfp" {
				t.Errorf("冲突对象错误: %+v", c)
			}
		}
	}
	if !found {
		t.Errorf("应命中 late_shutdown, got %+v", res.Conflicts)
	}
}

func TestCheckOrderEarlyActivation(t *testing.T) {
	// 下游 active 而上游未激活 → early_activation。
	timings := []ElementTiming{
		{
			ElementID:   "up",
			ElementName: "a",
			Upstream:    "",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateUninduced},
		},
		{
			ElementID:   "down",
			ElementName: "b",
			Upstream:    "up",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateActive},
		},
	}
	stages := []model.InductionStage{
		{Name: "诱导激活", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive},
	}
	res, err := CheckOrder("t1", timings, stages, 0)
	if err != nil {
		t.Fatalf("CheckOrder: %v", err)
	}
	found := false
	for _, c := range res.Conflicts {
		if c.Kind == KindEarlyActivation {
			found = true
		}
	}
	if !found {
		t.Errorf("应命中 early_activation, got %+v", res.Conflicts)
	}
}

func TestCheckOrderNoConflict(t *testing.T) {
	// 上游先激活、下游后激活，撤除时上游先关闭 → 无冲突。
	timings := []ElementTiming{
		{
			ElementID:   "up",
			ElementName: "a",
			StatePerStage: map[string]model.RegulationState{
				"诱导激活": model.StateActive,
				"撤除恢复": model.StateUninduced,
			},
		},
		{
			ElementID:   "down",
			ElementName: "b",
			Upstream:    "up",
			StatePerStage: map[string]model.RegulationState{
				"诱导激活": model.StateActive,
				"撤除恢复": model.StateUninduced,
			},
		},
	}
	stages := []model.InductionStage{
		{Name: "诱导激活", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive},
		{Name: "撤除恢复", StartSeq: 5, EndSeq: 8, ExpectedState: model.StateUninduced},
	}
	res, err := CheckOrder("t1", timings, stages, 2)
	if err != nil {
		t.Fatalf("CheckOrder: %v", err)
	}
	if len(res.Conflicts) != 0 {
		t.Errorf("预期无冲突, got %+v", res.Conflicts)
	}
}

func TestCheckOrderDedup(t *testing.T) {
	// 同一 (stage,up,down,kind) 只保留一条：本例规则 1（early_activation）
	// 与规则 3（inversion）kind 不同，应保留 2 条；再验证完全重复时不翻倍。
	timings := []ElementTiming{
		{
			ElementID:   "up",
			ElementName: "a",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateRepressed},
		},
		{
			ElementID:   "down",
			ElementName: "b",
			Upstream:    "up",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateActive},
		},
	}
	stages := []model.InductionStage{
		{Name: "诱导激活", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive},
	}
	res, err := CheckOrder("t1", timings, stages, 0)
	if err != nil {
		t.Fatalf("CheckOrder: %v", err)
	}
	// early_activation 与 inversion 各 1 条。
	if len(res.Conflicts) != 2 {
		t.Fatalf("应有 2 条不同 kind 冲突, got %d: %+v", len(res.Conflicts), res.Conflicts)
	}
	// 同一 (stage,up,down,kind) 完全重复场景：制造两条同 kind 冲突去重为 1。
	dup := []ElementTiming{
		{
			ElementID:   "up",
			ElementName: "a",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateUninduced},
		},
		{
			ElementID:   "down",
			ElementName: "b",
			Upstream:    "up",
			StatePerStage: map[string]model.RegulationState{"诱导激活": model.StateActive},
		},
	}
	res2, err := CheckOrder("t1", dup, stages, 0)
	if err != nil {
		t.Fatalf("CheckOrder: %v", err)
	}
	for _, c := range res2.Conflicts {
		if c.Kind != KindEarlyActivation {
			t.Fatalf("仅应有 early_activation, got %+v", res2.Conflicts)
		}
	}
	if len(res2.Conflicts) != 1 {
		t.Fatalf("同 kind 重复应去重为 1, got %d: %+v", len(res2.Conflicts), res2.Conflicts)
	}
}
