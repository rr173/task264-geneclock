package httpapi

import "net/http"

func (s *Server) handleNormalize(w http.ResponseWriter, r *http.Request) {
	baselines, err := s.app.NormalizeBaseline(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"baselines": baselines})
}

func (s *Server) handleListNormalized(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.ListNormalized(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleInfer(w http.ResponseWriter, r *http.Request) {
	res, err := s.app.InferRegulationStates(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleListStates(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.ListElementStates(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleCheckOrder(w http.ResponseWriter, r *http.Request) {
	res, err := s.app.CheckOrderConstraints(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) handleListConflicts(w http.ResponseWriter, r *http.Request) {
	list, err := s.app.ListConflicts(pathID(r, "id"))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}
