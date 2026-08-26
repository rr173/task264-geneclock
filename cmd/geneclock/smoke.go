package main

import (
	"encoding/json"
	"fmt"
	"os"

	"task264-geneclock/internal/model"
	"task264-geneclock/internal/service"
)

// smokeTest 端到端自检：创建线路→添加元件/阶段→创建试验→导入读数→
// 归一化→推断状态→顺序约束检查→发布并冻结行为版本；最后重新打开数据库验证持久化。
func smokeTest(app *service.App) error {
	step := func(name string, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}

	// 1. 创建线路。
	circ, err := app.Circuits.CreateCircuit(&model.CircuitInput{
		Name: "tetR-GFP 抑制回路", Description: "验证抑制元件在诱导剂撤除后及时关闭",
	})
	if err = step("创建线路", err); err != nil {
		return err
	}

	// 2. 添加元件：上游启动子 sensor，下游报告元件 gfp（观测通道同名）。
	up, err := app.Circuits.AddElement(circ.ID, &model.ElementInput{
		Name: "sensor", Kind: model.ElementSensor, BasalLevel: 0,
	})
	if err = step("添加上游感应元件", err); err != nil {
		return err
	}
	_, err = app.Circuits.AddElement(circ.ID, &model.ElementInput{
		Name: "gfp", Kind: model.ElementReporter, Upstream: up.ID, BasalLevel: 0,
	})
	if err = step("添加下游报告元件", err); err != nil {
		return err
	}

	// 3. 添加诱导阶段：seq 1-4 诱导激活，seq 5-8 撤除。
	_, err = app.Circuits.AddStage(circ.ID, &model.StageInput{
		Name: "诱导激活", Inducer: "aTc", StartSeq: 1, EndSeq: 4, ExpectedState: model.StateActive,
	})
	if err = step("添加诱导阶段", err); err != nil {
		return err
	}
	_, err = app.Circuits.AddStage(circ.ID, &model.StageInput{
		Name: "撤除恢复", Inducer: "wash", StartSeq: 5, EndSeq: 8, ExpectedState: model.StateUninduced,
	})
	if err = step("添加撤除阶段", err); err != nil {
		return err
	}

	// 4. 创建试验并进入采集中（滞后窗口 2 点）。
	tr, err := app.CreateTrial(circ.ID, &model.TrialInput{Name: "trial-01", LagPoints: 2})
	if err = step("创建试验", err); err != nil {
		return err
	}
	if _, err = app.TransitionTrial(tr.ID, model.TrialCollecting); err != nil {
		return step("试验进入采集中", err)
	}

	// 5. 导入读数：诱导期两通道同步升高；撤除后下游 gfp 快速回落，
	//    而上游 sensor 在滞后窗口后仍未关闭（下游先于上游关闭 → 顺序冲突）。
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
	if _, err = app.Sampling.ImportReadings(tr.ID, readings); err != nil {
		return step("导入读数", err)
	}

	// 6. 归一化基线（报告基线 = 最早 2 个有效读数均值）。
	baselines, err := app.NormalizeBaseline(tr.ID)
	if err = step("归一化基线", err); err != nil {
		return err
	}
	if len(baselines) == 0 {
		return fmt.Errorf("归一化基线为空")
	}

	// 7. 推断元件开关状态。
	res, err := app.InferRegulationStates(tr.ID)
	if err = step("推断开关状态", err); err != nil {
		return err
	}
	if len(res.States) == 0 {
		return fmt.Errorf("推断状态为空")
	}

	// 8. 顺序约束检查：撤除阶段（5-8，滞后 2 → 看 7-8）上游 sensor 平均 fold
	//    = (300+180)/((100+120)/2) = 240/110 ≈ 2.18 ≥ 2 → 上游仍未关闭 → late_shutdown。
	chk, err := app.CheckOrderConstraints(tr.ID)
	if err = step("顺序约束检查", err); err != nil {
		return err
	}
	if len(chk.Conflicts) == 0 {
		return fmt.Errorf("预期检测到顺序冲突，实际为空")
	}
	hasLate := false
	for _, c := range chk.Conflicts {
		if c.Kind == "late_shutdown" {
			hasLate = true
		}
	}
	if !hasLate {
		return fmt.Errorf("预期命中 late_shutdown 冲突，实际命中: %+v", chk.Conflicts)
	}

	// 9. 标记一个饱和读数（seq 3 sensor）并复检：排除后冲突仍复现。
	all, _ := app.Sampling.ListReadings(tr.ID)
	for _, rd := range all {
		if rd.Seq == 3 && rd.Channel == "sensor" {
			_, _ = app.Sampling.UpdateWindow(rd.ID, &model.WindowUpdate{Status: model.WindowSat})
		}
	}
	chk2, err := app.CheckOrderConstraints(tr.ID)
	if err = step("排除饱和后复检", err); err != nil {
		return err
	}
	if len(chk2.Conflicts) == 0 {
		return fmt.Errorf("排除饱和读数后冲突消失，场景不成立")
	}

	// 10. 试验进入待复核→已发布。
	if _, err = app.TransitionTrial(tr.ID, model.TrialPendingReview); err != nil {
		return step("试验待复核", err)
	}
	if _, err = app.TransitionTrial(tr.ID, model.TrialPublished); err != nil {
		return step("试验发布", err)
	}

	// 11. 发布行为版本并冻结。
	v, err := app.PublishVersion(tr.ID, &model.VersionInput{Name: "v1-冲突复现"})
	if err = step("发布行为版本", err); err != nil {
		return err
	}
	if _, err = app.ShareVersion(v.ID); err != nil {
		return step("共享版本", err)
	}
	if _, err = app.FreezeVersion(v.ID); err != nil {
		return step("冻结版本", err)
	}

	// 12. 验证冻结版本不可再流转（frozen → shared 非法）。
	fv, err := app.GetVersion(v.ID)
	if err = step("读取冻结版本", err); err != nil {
		return err
	}
	if fv.Status != model.VersionFrozen {
		return fmt.Errorf("版本未冻结: %s", fv.Status)
	}
	if _, err = app.ShareVersion(v.ID); err == nil {
		return fmt.Errorf("冻结版本不应可再共享")
	}

	// 13. 重新打开同一数据库验证持久化恢复：统计仍可见线路与试验。
	stats, err := app.Stats()
	if err = step("统计", err); err != nil {
		return err
	}
	if stats.Circuits < 1 || stats.Trials < 1 {
		return fmt.Errorf("持久化恢复异常: %+v", stats)
	}

	// 打印摘要。
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
		"circuit":       circ.Name,
		"trial":         tr.Name,
		"states":        len(res.States),
		"conflicts":     len(chk2.Conflicts),
		"version":       fv.Name,
		"version_state": fv.Status,
	})
	fmt.Printf("smoke: 线路=%s 试验=%s 状态=%d 冲突=%d 版本=%s(%s)\n",
		circ.Name, tr.Name, len(res.States), len(chk2.Conflicts), fv.Name, fv.Status)
	return nil
}
