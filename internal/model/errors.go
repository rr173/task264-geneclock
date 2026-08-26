// Package model 定义合成生物基因调控时序复核台的领域实体、状态与错误类型。
package model

import "errors"

// 领域错误，HTTP 层据此映射状态码。
var (
	ErrNotFound         = errors.New("资源不存在")
	ErrConflict         = errors.New("资源状态冲突")
	ErrValidation       = errors.New("参数校验失败")
	ErrFrozen           = errors.New("冻结版本不可修改")
	ErrDuplicate        = errors.New("唯一键重复")
	ErrBadTransition    = errors.New("非法状态流转")
	ErrUnitMissing      = errors.New("读数单位缺失")
	ErrUnknownElement   = errors.New("未知调控元件")
	ErrStageOverlap     = errors.New("诱导阶段时间区间重叠")
	ErrCircuitArchived  = errors.New("线路已封存")
	ErrTrialNotReady    = errors.New("试验状态不允许该操作")
	ErrBaselineNotReady = errors.New("基线尚未归一化")
	ErrNoStates         = errors.New("尚未推断调控状态")
	ErrNoConflicts      = errors.New("尚未执行顺序约束检查")
)

// ElementKind 调控元件类型。
type ElementKind string

const (
	ElementPromoter  ElementKind = "promoter"  // 启动子：驱动下游表达
	ElementRepressor ElementKind = "repressor" // 抑制元件：阻遏下游表达
	ElementReporter  ElementKind = "reporter"  // 报告元件：输出可观测荧光
	ElementSensor    ElementKind = "sensor"    // 感应元件：感知诱导剂浓度
)

// TrialStatus 线路试验状态机。
type TrialStatus string

const (
	TrialPreparing     TrialStatus = "preparing"     // 准备
	TrialCollecting    TrialStatus = "collecting"    // 采集中
	TrialPendingReview TrialStatus = "pending_review" // 待复核
	TrialPublished     TrialStatus = "published"     // 已发布
	TrialArchived      TrialStatus = "archived"      // 封存
)

// WindowStatus 读数窗口状态机。
type WindowStatus string

const (
	WindowPending  WindowStatus = "pending"   // 待校正
	WindowValid    WindowStatus = "valid"     // 有效
	WindowSat      WindowStatus = "saturated" // 饱和
	WindowMissing  WindowStatus = "missing"   // 缺失
	WindowExcluded WindowStatus = "excluded"  // 排除
)

// RegulationState 调控状态机。
type RegulationState string

const (
	StateUninduced   RegulationState = "uninduced"    // 未诱导
	StateActive      RegulationState = "active"       // 激活
	StateRepressed   RegulationState = "repressed"    // 抑制
	StateOrderFlaw   RegulationState = "order_conflict" // 顺序冲突
	StateConfirmed   RegulationState = "confirmed"    // 确认
)

// VersionStatus 行为版本状态机。
type VersionStatus string

const (
	VersionDraft     VersionStatus = "draft"     // 草稿
	VersionShared    VersionStatus = "shared"    // 共享
	VersionFrozen    VersionStatus = "frozen"    // 冻结
	VersionSuperseded VersionStatus = "superseded" // 替代
)

// CircuitStatus 线路生命周期。
type CircuitStatus string

const (
	CircuitActive   CircuitStatus = "active"   // 在用
	CircuitArchived CircuitStatus = "archived" // 封存
)

// 状态流转合法性表。
var trialTransitions = map[TrialStatus]map[TrialStatus]bool{
	TrialPreparing: {TrialCollecting: true},
	TrialCollecting: {TrialPendingReview: true, TrialArchived: true},
	TrialPendingReview: {TrialPublished: true, TrialArchived: true},
	TrialPublished: {TrialArchived: true},
}

// CanTransitionTrial 判断试验状态流转是否合法。
func CanTransitionTrial(from, to TrialStatus) bool {
	if from == to {
		return true
	}
	return trialTransitions[from][to]
}

var windowTransitions = map[WindowStatus]map[WindowStatus]bool{
	WindowPending: {WindowValid: true, WindowSat: true, WindowMissing: true, WindowExcluded: true},
	WindowValid:   {WindowSat: true, WindowMissing: true, WindowExcluded: true},
	WindowSat:     {WindowValid: true, WindowExcluded: true},
	WindowMissing: {WindowValid: true, WindowExcluded: true},
}

// CanTransitionWindow 判断读数窗口状态流转是否合法。
func CanTransitionWindow(from, to WindowStatus) bool {
	if from == to {
		return true
	}
	return windowTransitions[from][to]
}

var versionTransitions = map[VersionStatus]map[VersionStatus]bool{
	VersionDraft:      {VersionShared: true, VersionSuperseded: true},
	VersionShared:     {VersionFrozen: true, VersionSuperseded: true},
	VersionFrozen:     {VersionSuperseded: true},
}

// CanTransitionVersion 判断行为版本状态流转是否合法。
func CanTransitionVersion(from, to VersionStatus) bool {
	if from == to {
		return true
	}
	return versionTransitions[from][to]
}

// 归一化后用于判定开关状态的阈值常量。
const (
	// ActivationFold 归一化 fold-change 达到该值判为激活。
	ActivationFold = 2.0
	// RepressionFold 归一化 fold-change 低于该值判为抑制。
	RepressionFold = 0.5
	// MinLagPoints 滞后窗口最小点数。
	MinLagPoints = 1
	// MaxLagPoints 滞后窗口最大点数。
	MaxLagPoints = 12
)
