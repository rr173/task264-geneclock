package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestExcludedReadingsDoNotSkewInference(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "excluded-fold", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter}); err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatal(err)
	}
	readings := []model.ReadingInput{
		{Seq: 1, Channel: "gfp", RawValue: 100, Unit: "RFU"},
		{Seq: 2, Channel: "gfp", RawValue: 5000, Unit: "RFU"},
		{Seq: 3, Channel: "gfp", RawValue: 100, Unit: "RFU"},
		{Seq: 4, Channel: "gfp", RawValue: 100, Unit: "RFU"},
	}
	loaded, err := app.Sampling.ImportReadings(tr.ID, readings)
	if err != nil {
		t.Fatal(err)
	}
	for _, rd := range loaded {
		if rd.Seq == 2 {
			if _, err := app.Sampling.UpdateWindow(rd.ID, &model.WindowUpdate{Status: model.WindowExcluded}); err != nil {
				t.Fatal(err)
			}
		}
	}
	baselines, err := app.NormalizeBaseline(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if baselines["gfp"] > 150 {
		t.Fatalf("excluded saturated point must not skew baseline, got %v", baselines["gfp"])
	}
}
