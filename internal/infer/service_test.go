package infer

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestClassify(t *testing.T) {
	cases := []struct {
		avg      float64
		expected model.RegulationState
		want     model.RegulationState
	}{
		{3.0, model.StateActive, model.StateActive},     // 达到激活阈值
		{0.3, model.StateRepressed, model.StateRepressed}, // 低于抑制阈值
		{1.0, model.StateActive, model.StateUninduced},   // 中间区间
		{3.0, model.StateRepressed, model.StateOrderFlaw}, // 期望抑制却激活
		{0.3, model.StateActive, model.StateOrderFlaw},    // 期望激活却被抑制
	}
	for _, c := range cases {
		got, _ := classify(c.avg, c.expected)
		if got != c.want {
			t.Errorf("classify(%v,%s)=%s want %s", c.avg, c.expected, got, c.want)
		}
	}
}

func TestAvgFold(t *testing.T) {
	m := map[int]float64{1: 1.0, 2: 2.0, 3: 3.0, 5: 5.0}
	avg, n := avgFold(m, 2, 4)
	if n != 2 {
		t.Fatalf("n=%d want 2", n)
	}
	if avg != 2.5 {
		t.Fatalf("avg=%v want 2.5", avg)
	}
	if _, n := avgFold(m, 8, 10); n != 0 {
		t.Fatalf("空区间 n=%d want 0", n)
	}
}
