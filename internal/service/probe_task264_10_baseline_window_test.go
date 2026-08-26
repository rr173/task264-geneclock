package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestBaselineUsesPreInductionWindowOnly(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "baseline", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "on", Inducer: "a", StartSeq: 3, EndSeq: 5, ExpectedState: model.StateActive}); err != nil {
		t.Fatal(err)
	}
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t", LagPoints: 2})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatal(err)
	}
	vals := []model.ReadingInput{
		{Seq: 1, Channel: "gfp", RawValue: 100, Unit: "RFU"}, {Seq: 2, Channel: "gfp", RawValue: 100, Unit: "RFU"},
		{Seq: 3, Channel: "gfp", RawValue: 5000, Unit: "RFU"}, {Seq: 4, Channel: "gfp", RawValue: 5000, Unit: "RFU"},
	}
	if _, err := app.Sampling.ImportReadings(tr.ID, vals); err != nil {
		t.Fatal(err)
	}
	baselines, err := app.NormalizeBaseline(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	base := baselines["gfp"]
	if base > 150 {
		t.Fatalf("baseline must come from pre-induction window only, got %v", base)
	}
	res, err := app.InferRegulationStates(tr.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range res.States {
		if st.StageName == "on" && st.State != model.StateActive {
			t.Fatalf("with correct baseline induction should be active, got %s fold=%v", st.State, st.FoldChange)
		}
	}
}
