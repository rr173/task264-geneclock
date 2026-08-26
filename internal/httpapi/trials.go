package httpapi

import (
	"net/http"

	"task264-geneclock/internal/model"
)

func (s *Server) handleCreateTrial(w http.ResponseWriter, r *http.Request) {
	var in model.TrialInput
	if !readBody(w, r, &in) {
		return
	}
	tr, err := s.app.CreateTrial(pathID(r, "id"), &in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tr)
}

func (s *Server) handleListTrials(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.ListTrials(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetTrial(w http.ResponseWriter, r *http.Request) {
	tr, err := s.app.GetTrial(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tr)
}

type transitionInput struct {
	Status model.TrialStatus `json:"status"`
}

func (s *Server) handleTransitionTrial(w http.ResponseWriter, r *http.Request) {
	var in transitionInput
	if !readBody(w, r, &in) {
		return
	}
	tr, err := s.app.TransitionTrial(pathID(r, "id"), in.Status)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tr)
}

type lagInput struct {
	LagPoints int `json:"lag_points"`
}

func (s *Server) handleSetLag(w http.ResponseWriter, r *http.Request) {
	var in lagInput
	if !readBody(w, r, &in) {
		return
	}
	tr, err := s.app.SetLagWindow(pathID(r, "id"), in.LagPoints)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tr)
}
