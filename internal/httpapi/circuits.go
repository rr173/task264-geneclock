package httpapi

import (
	"net/http"

	"task264-geneclock/internal/model"
)

func (s *Server) handleCreateCircuit(w http.ResponseWriter, r *http.Request) {
	var in model.CircuitInput
	if !readBody(w, r, &in) {
		return
	}
	circ, err := s.app.Circuits.CreateCircuit(&in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, circ)
}

func (s *Server) handleListCircuits(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Circuits.List()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetCircuit(w http.ResponseWriter, r *http.Request) {
	circ, err := s.app.Circuits.Get(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	elements, err := s.app.Circuits.ListElements(circ.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	stages, err := s.app.Circuits.ListStages(circ.ID)
	if err != nil {
		writeErr(w, err)
		return
	}
	circ.Elements = elements
	circ.Stages = stages
	writeJSON(w, http.StatusOK, circ)
}

func (s *Server) handleArchiveCircuit(w http.ResponseWriter, r *http.Request) {
	circ, err := s.app.Circuits.Archive(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, circ)
}

func (s *Server) handleAddElement(w http.ResponseWriter, r *http.Request) {
	var in model.ElementInput
	if !readBody(w, r, &in) {
		return
	}
	e, err := s.app.Circuits.AddElement(pathID(r, "id"), &in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleListElements(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Circuits.ListElements(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleAddStage(w http.ResponseWriter, r *http.Request) {
	var in model.StageInput
	if !readBody(w, r, &in) {
		return
	}
	st, err := s.app.Circuits.AddStage(pathID(r, "id"), &in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, st)
}

func (s *Server) handleListStages(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Circuits.ListStages(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
