package service

import (
	"errors"
	"testing"

	"task264-geneclock/internal/model"
)

func TestImportReadingsRejectedAfterCollecting(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "import-state", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialPendingReview); err != nil {
		t.Fatal(err)
	}
	_, err = app.Sampling.ImportReadings(tr.ID, []model.ReadingInput{
		{Seq: 1, Channel: "gfp", RawValue: 100, Unit: "RFU"},
	})
	if !errors.Is(err, model.ErrTrialNotReady) {
		t.Fatalf("pending_review must reject import, got %v", err)
	}
}
