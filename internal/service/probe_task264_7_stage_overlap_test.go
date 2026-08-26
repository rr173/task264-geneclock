package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

func TestAdjacentStagesOverlapRejected(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "overlap", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "a", Inducer: "x", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{Name: "b", Inducer: "y", StartSeq: 4, EndSeq: 8, ExpectedState: model.StateUninduced}); err != model.ErrStageOverlap {
		t.Fatalf("shared endpoint stages must overlap, got %v", err)
	}
}
