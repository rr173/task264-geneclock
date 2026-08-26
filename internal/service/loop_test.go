package service

import (
	"testing"

	"task264-geneclock/internal/model"
)

// 端到端集成测试：完整业务闭环（与 --smoke-test 对齐）。
func TestFullLoop(t *testing.T) {
	app := newTestApp(t)

	// 创建线路
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{
		Name: "测试回路", Description: "集成测试",
	})
	if err != nil {
		t.Fatalf("创建线路: %v", err)
	}

	// 添加元件
	up, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{
		Name: "sensor", Kind: model.ElementSensor,
	})
	if err != nil {
		t.Fatalf("添加上游: %v", err)
	}
	if _, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{
		Name: "gfp", Kind: model.ElementReporter, Upstream: up.ID,
	}); err != nil {
		t.Fatalf("添加下游: %v", err)
	}

	// 添加阶段
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{
		Name: "激活", Inducer: "aTc", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive,
	}); err != nil {
		t.Fatalf("添加激活阶段: %v", err)
	}
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{
		Name: "撤除", Inducer: "wash", StartSeq: 5, EndSeq: 8, ExpectedState: model.StateUninduced,
	}); err != nil {
		t.Fatalf("添加撤除阶段: %v", err)
	}

	// 阶段重叠校验
	if _, err := app.Circuits.AddStage(circ.ID, &model.StageInput{
		Name: "重叠", Inducer: "x", StartSeq: 3, EndSeq: 6, ExpectedState: model.StateActive,
	}); err != model.ErrStageOverlap {
		t.Fatalf("应拒绝重叠阶段, got %v", err)
	}

	// 创建试验
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "t1", LagPoints: 2})
	if err != nil {
		t.Fatalf("创建试验: %v", err)
	}
	if _, err := app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		t.Fatalf("采集: %v", err)
	}

	// 导入读数
	readings := []model.ReadingInput{
		{Seq: 1, Channel: "sensor", RawValue: 100, Unit: "RFU"},
		{Seq: 2, Channel: "sensor", RawValue: 120, Unit: "RFU"},
		{Seq: 3, Channel: "sensor", RawValue: 800, Unit: "RFU"},
		{Seq: 4, Channel: "sensor", RawValue: 1500, Unit: "RFU"},
		{Seq: 5, Channel: "sensor", RawValue: 1400, Unit: "RFU"},
		{Seq: 6, Channel: "sensor", RawValue: 1000, Unit: "RFU"},
		{Seq: 7, Channel: "sensor", RawValue: 300, Unit: "RFU"},
		{Seq: 8, Channel: "sensor", RawValue: 180, Unit: "RFU"},
		{Seq: 1, Channel: "gfp", RawValue: 60, Unit: "RFU"},
		{Seq: 2, Channel: "gfp", RawValue: 70, Unit: "RFU"},
		{Seq: 3, Channel: "gfp", RawValue: 600, Unit: "RFU"},
		{Seq: 4, Channel: "gfp", RawValue: 1200, Unit: "RFU"},
		{Seq: 5, Channel: "gfp", RawValue: 150, Unit: "RFU"},
		{Seq: 6, Channel: "gfp", RawValue: 100, Unit: "RFU"},
		{Seq: 7, Channel: "gfp", RawValue: 80, Unit: "RFU"},
		{Seq: 8, Channel: "gfp", RawValue: 60, Unit: "RFU"},
	}
	if _, err := app.Sampling.ImportReadings(tr.ID, readings); err != nil {
		t.Fatalf("导入读数: %v", err)
	}

	// 幂等：重复导入同 (seq,channel) 拒绝。
	if _, err := app.Sampling.ImportReadings(tr.ID, []model.ReadingInput{
		{Seq: 1, Channel: "sensor", RawValue: 100, Unit: "RFU"},
	}); err != model.ErrDuplicate {
		t.Fatalf("重复读数应拒绝, got %v", err)
	}

	// 归一化
	baselines, err := app.NormalizeBaseline(tr.ID)
	if err != nil {
		t.Fatalf("归一化: %v", err)
	}
	if len(baselines) != 2 {
		t.Fatalf("应有两通道基线, got %v", baselines)
	}

	// 推断
	res, err := app.InferRegulationStates(tr.ID)
	if err != nil {
		t.Fatalf("推断: %v", err)
	}
	if len(res.States) != 4 {
		t.Fatalf("应有 4 条状态, got %d", len(res.States))
	}

	// 顺序检查：撤除阶段上游未关闭 → late_shutdown
	chk, err := app.CheckOrderConstraints(tr.ID)
	if err != nil {
		t.Fatalf("顺序检查: %v", err)
	}
	if len(chk.Conflicts) == 0 {
		t.Fatal("应检测到顺序冲突")
	}

	// 版本发布与冻结
	v, err := app.PublishVersion(tr.ID, &model.VersionInput{Name: "v1"})
	if err != nil {
		t.Fatalf("发布版本: %v", err)
	}
	if _, err := app.ShareVersion(v.ID); err != nil {
		t.Fatalf("共享版本: %v", err)
	}
	if _, err := app.FreezeVersion(v.ID); err != nil {
		t.Fatalf("冻结版本: %v", err)
	}
	if _, err := app.ShareVersion(v.ID); err == nil {
		t.Fatal("冻结版本不应可再共享")
	}

	// 统计
	stats, err := app.Stats()
	if err != nil {
		t.Fatalf("统计: %v", err)
	}
	if stats.Circuits != 1 || stats.Trials != 1 || stats.PublishedVersions != 1 {
		t.Fatalf("统计异常: %+v", stats)
	}
}

// newTestApp 构造内存 SQLite 测试应用。
func newTestApp(t *testing.T) *App {
	t.Helper()
	st, err := openMemoryStore()
	if err != nil {
		t.Fatalf("打开内存库: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st)
}
