package normalize

import (
	"testing"
)

func TestFold(t *testing.T) {
	cases := []struct {
		raw, base, want float64
	}{
		{200, 100, 2.0},
		{150, 100, 1.5},
		{50, 100, 0.5},
		{0, 100, 0},
		{333, 100, 3.33},
	}
	for _, c := range cases {
		if got := fold(c.raw, c.base); got != c.want {
			t.Errorf("fold(%v,%v)=%v want %v", c.raw, c.base, got, c.want)
		}
	}
	if fold(100, 0) != 0 {
		t.Error("零基线应返回 0")
	}
}
