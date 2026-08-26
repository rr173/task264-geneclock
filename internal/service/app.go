// Package service 编排业务包：组合线路/采样/归一化/推断/约束/版本服务，暴露应用级用例。
package service

import (
	"time"

	"task264-geneclock/internal/circuit"
	"task264-geneclock/internal/infer"
	"task264-geneclock/internal/normalize"
	"task264-geneclock/internal/sampling"
	"task264-geneclock/internal/store"
	"task264-geneclock/internal/version"
)

// App 应用编排根。
type App struct {
	Store      *store.Store
	Circuits   *circuit.Service
	Sampling   *sampling.Service
	Normalize  *normalize.Service
	Infer      *infer.Service
	Versions   *version.Service
	Analysis   *store.AnalysisStore
	TrialStore *store.TrialStore
	CircuitStore *store.CircuitStore
}

// New 组装全部服务。
func New(st *store.Store) *App {
	circuitStore := store.NewCircuitStore(st.DB())
	trialStore := store.NewTrialStore(st.DB())
	analysisStore := store.NewAnalysisStore(st.DB())
	return &App{
		Store:        st,
		Circuits:     circuit.NewService(circuitStore),
		Sampling:     sampling.NewService(trialStore),
		Normalize:    normalize.NewService(trialStore),
		Infer:        infer.NewService(trialStore),
		Versions:     version.NewService(analysisStore),
		Analysis:     analysisStore,
		TrialStore:   trialStore,
		CircuitStore: circuitStore,
	}
}

// Now 统一时间源。
func Now() time.Time { return time.Now().UTC() }
