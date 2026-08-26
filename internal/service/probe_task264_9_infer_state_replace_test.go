package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestInferClearsConflictsWithStates(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "double-state", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "on", Inducer: "a", StartSeq: 1, EndSeq: 2, ExpectedState: model.StateActive}); err != nil {
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
		{Seq: 1, Channel: "gfp", RawValue: 100, Unit: "RFU"}, {Seq: 2, Channel: "gfp", RawValue: 900, Unit: "RFU"},
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
	if _, err := app.InferRegulationStates(tr.ID); err != nil {
		t.Fatal(err)
	}
	states, err := app.ListElementStates(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 {
		t.Fatalf("re-infer must replace old states, got %d rows", len(states))
	}
}
