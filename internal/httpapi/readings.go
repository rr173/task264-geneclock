package httpapi

import (
	"net/http"

	"task264-geneclock/internal/model"
)

type readingsInput struct {
	Readings []model.ReadingInput `json:"readings"`
}

func (s *Server) handleImportReadings(w http.ResponseWriter, r *http.Request) {
	var in readingsInput
	if !readBody(w, r, &in) {
		return
	}
	if len(in.Readings) == 0 {
		writeErr(w, model.ErrValidation)
		return
	}
	list, err := s.app.Sampling.ImportReadings(pathID(r, "id"), in.Readings)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"imported": len(list),
		"readings": list,
	})
}

func (s *Server) handleListReadings(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.Sampling.ListReadings(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleUpdateWindow(w http.ResponseWriter, r *http.Request) {
	var in model.WindowUpdate
	if !readBody(w, r, &in) {
		return
	}
	rd, err := s.app.Sampling.UpdateWindow(pathID(r, "id"), &in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rd)
}
