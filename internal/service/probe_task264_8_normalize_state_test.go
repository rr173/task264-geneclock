package service

import (
	"errors"
	"testing"

	"task264-geneclock/internal/model"
)

func TestNormalizeBaselineRequiresCollectingTrial(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "normalize-state", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter}); err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.NormalizeBaseline(tr.ID)
	if !errors.Is(err, model.ErrTrialNotReady) {
		t.Fatalf("preparing trial must reject normalize, got %v", err)
	}
}
