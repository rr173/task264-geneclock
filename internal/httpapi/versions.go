package httpapi

import (
	"net/http"

	"task264-geneclock/internal/model"
)

func (s *Server) handlePublishVersion(w http.ResponseWriter, r *http.Request) {
	var in model.VersionInput
	if !readBody(w, r, &in) {
		return
	}
	v, err := s.app.PublishVersion(pathID(r, "id"), &in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (s *Server) handleListVersions(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.ListVersions(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.app.GetVersion(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleShareVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.app.ShareVersion(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleFreezeVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.app.FreezeVersion(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (s *Server) handleSupersedeVersion(w http.ResponseWriter, r *http.Request) {
	v, err := s.app.SupersedeVersion(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, v)
}
