// Package httpapi 提供 RESTful HTTP 层，路由统一以 /api 开头。
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/service"
)

// Server HTTP 服务。
type Server struct {
	app *service.App
	mux *http.ServeMux
}

// New 构造 HTTP 服务并注册全部路由。
func New(app *service.App) *Server {
	s := &Server{app: app, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler 返回 http.Handler。
func (s *Server) Handler() http.Handler { return s.mux }

// routes 注册 /api 下全部路由。
func (s *Server) routes() {
	// 线路
	s.mux.HandleFunc("POST /api/circuits", s.handleCreateCircuit)
	s.mux.HandleFunc("GET /api/circuits", s.handleListCircuits)
	s.mux.HandleFunc("GET /api/circuits/{id}", s.handleGetCircuit)
	s.mux.HandleFunc("POST /api/circuits/{id}/archive", s.handleArchiveCircuit)
	s.mux.HandleFunc("POST /api/circuits/{id}/elements", s.handleAddElement)
	s.mux.HandleFunc("GET /api/circuits/{id}/elements", s.handleListElements)
	s.mux.HandleFunc("POST /api/circuits/{id}/stages", s.handleAddStage)
	s.mux.HandleFunc("GET /api/circuits/{id}/stages", s.handleListStages)
	// 试验
	s.mux.HandleFunc("POST /api/circuits/{id}/trials", s.handleCreateTrial)
	s.mux.HandleFunc("GET /api/circuits/{id}/trials", s.handleListTrials)
	s.mux.HandleFunc("GET /api/trials/{id}", s.handleGetTrial)
	s.mux.HandleFunc("POST /api/trials/{id}/transition", s.handleTransitionTrial)
	s.mux.HandleFunc("PATCH /api/trials/{id}/lag", s.handleSetLag)
	// 读数
	s.mux.HandleFunc("POST /api/trials/{id}/readings", s.handleImportReadings)
	s.mux.HandleFunc("GET /api/trials/{id}/readings", s.handleListReadings)
	s.mux.HandleFunc("PATCH /api/readings/{id}/window", s.handleUpdateWindow)
	// 分析
	s.mux.HandleFunc("POST /api/trials/{id}/normalize", s.handleNormalize)
	s.mux.HandleFunc("GET /api/trials/{id}/normalized", s.handleListNormalized)
	s.mux.HandleFunc("POST /api/trials/{id}/infer", s.handleInfer)
	s.mux.HandleFunc("GET /api/trials/{id}/states", s.handleListStates)
	s.mux.HandleFunc("POST /api/trials/{id}/check-order", s.handleCheckOrder)
	s.mux.HandleFunc("GET /api/trials/{id}/conflicts", s.handleListConflicts)
	// 版本
	s.mux.HandleFunc("POST /api/trials/{id}/versions", s.handlePublishVersion)
	s.mux.HandleFunc("GET /api/trials/{id}/versions", s.handleListVersions)
	s.mux.HandleFunc("GET /api/versions/{id}", s.handleGetVersion)
	s.mux.HandleFunc("POST /api/versions/{id}/share", s.handleShareVersion)
	s.mux.HandleFunc("POST /api/versions/{id}/freeze", s.handleFreezeVersion)
	s.mux.HandleFunc("POST /api/versions/{id}/supersede", s.handleSupersedeVersion)
	// 统计与健康
	s.mux.HandleFunc("GET /api/stats", s.handleStats)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
}

// writeJSON 统一 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// writeErr 错误响应映射。
func writeErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, model.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, model.ErrValidation), errors.Is(err, model.ErrUnitMissing),
		errors.Is(err, model.ErrUnknownElement), errors.Is(err, model.ErrStageOverlap):
		status = http.StatusBadRequest
	case errors.Is(err, model.ErrDuplicate):
		status = http.StatusConflict
	case errors.Is(err, model.ErrConflict), errors.Is(err, model.ErrFrozen),
		errors.Is(err, model.ErrBadTransition), errors.Is(err, model.ErrTrialNotReady),
		errors.Is(err, model.ErrCircuitArchived), errors.Is(err, model.ErrBaselineNotReady),
		errors.Is(err, model.ErrNoStates), errors.Is(err, model.ErrNoConflicts):
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// readBody 解析 JSON 请求体。
func readBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeErr(w, model.ErrValidation)
		return false
	}
	return true
}

// pathID 取路径参数。
func pathID(r *http.Request, key string) string { return r.PathValue(key) }

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	st, err := s.app.Stats()
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}
