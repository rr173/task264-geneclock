package model

import "testing"

func TestCanTransitionTrial(t *testing.T) {
	cases := []struct {
		from, to TrialStatus
		want     bool
	}{
		{TrialPreparing, TrialCollecting, true},
		{TrialCollecting, TrialPendingReview, true},
		{TrialPendingReview, TrialPublished, true},
		{TrialPublished, TrialArchived, true},
		{TrialPreparing, TrialPublished, false},
		{TrialCollecting, TrialPreparing, false},
		{TrialArchived, TrialPendingReview, false},
	}
	for _, c := range cases {
		if got := CanTransitionTrial(c.from, c.to); got != c.want {
			t.Errorf("CanTransitionTrial(%s,%s)=%v want %v", c.from, c.to, got, c.want)
		}
	}
}

func TestCanTransitionWindow(t *testing.T) {
	if !CanTransitionWindow(WindowPending, WindowSat) {
		t.Error("pending->saturated 应允许")
	}
	if !CanTransitionWindow(WindowSat, WindowExcluded) {
		t.Error("saturated->excluded 应允许")
	}
	if CanTransitionWindow(WindowExcluded, WindowValid) {
		t.Error("excluded->valid 应禁止")
	}
	if !CanTransitionWindow(WindowMissing, WindowValid) {
		t.Error("missing->valid 应允许（补测）")
	}
}

func TestCanTransitionVersion(t *testing.T) {
	if !CanTransitionVersion(VersionDraft, VersionShared) {
		t.Error("draft->shared 应允许")
	}
	if !CanTransitionVersion(VersionShared, VersionFrozen) {
		t.Error("shared->frozen 应允许")
	}
	if !CanTransitionVersion(VersionFrozen, VersionSuperseded) {
		t.Error("frozen->superseded 应允许")
	}
	if CanTransitionVersion(VersionFrozen, VersionShared) {
		t.Error("frozen->shared 应禁止")
	}
}

func TestStageOverlap(t *testing.T) {
	a := &InductionStage{StartSeq: 1, EndSeq: 4}
	b := &InductionStage{StartSeq: 4, EndSeq: 6} // 端点相接视为重叠
	c := &InductionStage{StartSeq: 5, EndSeq: 6} // 不重叠
	d := &InductionStage{StartSeq: 2, EndSeq: 3} // 完全包含
	if !a.Overlaps(b) {
		t.Error("端点相接应重叠")
	}
	if a.Overlaps(c) {
		t.Error("5-6 与 1-4 不应重叠")
	}
	if !a.Overlaps(d) {
		t.Error("包含关系应重叠")
	}
}

func TestReadingInputValidate(t *testing.T) {
	ok := ReadingInput{Seq: 1, Channel: "gfp", RawValue: 100, Unit: "RFU"}
	if err := ok.Validate(); err != nil {
		t.Errorf("合法读数应通过: %v", err)
	}
	noUnit := ReadingInput{Seq: 1, Channel: "gfp", RawValue: 100}
	if err := noUnit.Validate(); err != ErrUnitMissing {
		t.Errorf("缺单位应返回 ErrUnitMissing, got %v", err)
	}
	neg := ReadingInput{Seq: 1, Channel: "gfp", RawValue: -1, Unit: "RFU"}
	if err := neg.Validate(); err != ErrValidation {
		t.Errorf("负值应返回 ErrValidation, got %v", err)
	}
}
