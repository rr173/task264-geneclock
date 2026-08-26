package service

import (
	"encoding/json"
	"testing"

	"task264-geneclock/internal/model"
)

func TestPublishVersionIncludesFreshCircuitElements(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "snap", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "sensor", Kind: model.ElementSensor}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "on", Inducer: "a", StartSeq: 1, EndSeq: 2, ExpectedState: model.StateActive}); err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatal(err)
	}
	vals := []model.ReadingInput{
		{Seq: 1, Channel: "sensor", RawValue: 100, Unit: "RFU"}, {Seq: 2, Channel: "sensor", RawValue: 900, Unit: "RFU"},
		{Seq: 1, Channel: "gfp", RawValue: 80, Unit: "RFU"}, {Seq: 2, Channel: "gfp", RawValue: 700, Unit: "RFU"},
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
	v, err := app.PublishVersion(tr.ID, &model.VersionInput{Name: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	var snap struct {
		Elements []model.Element `json:"elements"`
	}
	if err := json.Unmarshal([]byte(v.CircuitSnapshot), &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.Elements) < 2 {
		t.Fatalf("snapshot must include element added after trial, got %d elements", len(snap.Elements))
	}
}
