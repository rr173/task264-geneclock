package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestRecheckOrderClearsStaleConflicts(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "stale-conflict", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	up, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "sensor", Kind: model.ElementSensor})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter, Upstream: up.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "on", Inducer: "a", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "off", Inducer: "w", StartSeq: 5, EndSeq: 8, ExpectedState: model.StateUninduced}); err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatal(err)
	}
	vals := []model.ReadingInput{
		{Seq: 1, Channel: "sensor", RawValue: 100, Unit: "RFU"}, {Seq: 2, Channel: "sensor", RawValue: 100, Unit: "RFU"},
		{Seq: 3, Channel: "sensor", RawValue: 900, Unit: "RFU"}, {Seq: 4, Channel: "sensor", RawValue: 900, Unit: "RFU"},
		{Seq: 5, Channel: "sensor", RawValue: 900, Unit: "RFU"}, {Seq: 6, Channel: "sensor", RawValue: 900, Unit: "RFU"},
		{Seq: 7, Channel: "sensor", RawValue: 900, Unit: "RFU"}, {Seq: 8, Channel: "sensor", RawValue: 900, Unit: "RFU"},
		{Seq: 1, Channel: "gfp", RawValue: 80, Unit: "RFU"}, {Seq: 2, Channel: "gfp", RawValue: 80, Unit: "RFU"},
		{Seq: 3, Channel: "gfp", RawValue: 700, Unit: "RFU"}, {Seq: 4, Channel: "gfp", RawValue: 700, Unit: "RFU"},
		{Seq: 5, Channel: "gfp", RawValue: 80, Unit: "RFU"}, {Seq: 6, Channel: "gfp", RawValue: 80, Unit: "RFU"},
		{Seq: 7, Channel: "gfp", RawValue: 80, Unit: "RFU"}, {Seq: 8, Channel: "gfp", RawValue: 80, Unit: "RFU"},
	}
	if _, err := app.Sampling.ImportReadings(tr.ID, vals); err != nil {
		t.Fatal(err)
	}
	if _, err := app.NormalizeBaseline(tr.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.InferRegulationStates(tr.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := app.CheckOrderConstraints(tr.ID); err != nil {
		t.Fatal(err)
	}
	first, err := app.ListConflicts(tr.ID)
	if err != nil || len(first) == 0 {
		t.Fatalf("first check need conflicts, got %d err=%v", len(first), err)
	}
	if _, err := app.CheckOrderConstraints(tr.ID); err != nil {
		t.Fatal(err)
	}
	second, err := app.ListConflicts(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != len(first) {
		t.Fatalf("recheck must replace conflicts not accumulate: first=%d second=%d", len(first), len(second))
	}
}
