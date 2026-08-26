package sampling

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestItoa(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{5, "5"},
		{123, "123"},
		{-7, "-7"},
	}
	for _, c := range cases {
		if got := itoa(c.n); got != c.want {
			t.Errorf("itoa(%d)=%q want %q", c.n, got, c.want)
		}
	}
}

func TestWindowTransitions(t *testing.T) {
	if !model.CanTransitionWindow(model.WindowPending, model.WindowSat) {
		t.Error("pending->saturated 应允许")
	}
	if model.CanTransitionWindow(model.WindowExcluded, model.WindowValid) {
		t.Error("excluded->valid 应禁止")
	}
	if !model.CanTransitionWindow(model.WindowMissing, model.WindowExcluded) {
		t.Error("missing->excluded 应允许")
	}
}
