package service

import (
	"errors"
	"testing"

	"task264-geneclock/internal/model"
)

func TestArchivedCircuitRejectsNewElements(t *testing.T) {
	app := newTestApp(t)
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{Name: "archived", Description: "probe"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Circuits.Archive(circ.ID); err != nil {
		t.Fatal(err)
	}
	_, err = app.Circuits.AddElement(circ.ID, &model.ElementInput{Name: "gfp", Kind: model.ElementReporter})
	if !errors.Is(err, model.ErrCircuitArchived) {
		t.Fatalf("archived circuit must reject new elements, got %v", err)
	}
}
