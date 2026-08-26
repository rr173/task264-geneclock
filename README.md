# task264-geneclock 合成生物基因调控时序复核台

基于 Go 实现的合成生物基因调控时序复核 Web 项目，一款后端服务，完成调控线路与诱导阶段编排、多通道荧光读数归一化与元件开关状态推断，并对表达先后顺序约束做复核与行为版本发布。

## 业务背景

合成生物研究者构建基因调控线路（启动子/抑制元件/报告基因），在诱导剂切换后观察各元件荧光表达时序。本服务校验线路是否遵守预期的表达先后关系：上游元件应先于下游激活、先于下游关闭。若下游信号先于上游关闭（或上游在撤除后未及时关闭），系统标记顺序冲突，供研究者复核并发布行为版本。

## 核心闭环

1. 创建调控线路，录入元件（promoter/repressor/reporter/sensor）与诱导阶段（诱导剂、起始/结束读数序号、期望状态）。
2. 创建线路试验，进入采集中后按通道批量导入荧光读数（幂等：同一 trial+seq+channel 拒绝重复）。
3. 归一化：以「诱导前窗口」为报告基线计算各通道 fold-change。
4. 推断元件开关状态：fold-change ≥ 2.0 判激活、≤ 0.5 判抑制，否则未诱导。
5. 顺序约束检查：检测 early_activation（下游先于上游激活）、late_shutdown（撤除后上游未及时关闭）、inversion（顺序反转）。
6. 研究者标记饱和/缺失/排除读数窗口、调整滞后窗口后复检；发布行为版本（草稿→共享→冻结→替代）。

## 标准命令

```bash
# 构建
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
# 静态检查
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
# 测试
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
# 端到端自检（创建数据→推断→冲突→发布冻结版本→统计验证持久化）
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/geneclock --smoke-test
# 启动服务
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/geneclock --addr :8080 --db geneclock.db
```

## API 入口（前缀 /api）

| 能力 | 入口 | 说明 |
|---|---|---|
| 创建线路 | POST /api/circuits | 线路名+描述 |
| 线路列表/详情 | GET /api/circuits, GET /api/circuits/{id} | 详情含元件与阶段 |
| 封存线路 | POST /api/circuits/{id}/archive | 封存后不可增改 |
| 添加/列出元件 | POST|GET /api/circuits/{id}/elements | 校验未知上游 |
| 添加/列出诱导阶段 | POST|GET /api/circuits/{id}/stages | 拒绝区间重叠 |
| 创建/列出试验 | POST|GET /api/circuits/{id}/trials | preparing→…→published |
| 试验状态流转 | POST /api/trials/{id}/transition | 合法流转表校验 |
| 调整滞后窗口 | PATCH /api/trials/{id}/lag | 1..12 点 |
| 导入读数 | POST /api/trials/{id}/readings | 幂等，缺单位拒绝 |
| 读数列表 | GET /api/trials/{id}/readings | |
| 标记读数窗口 | PATCH /api/readings/{id}/window | valid/saturated/missing/excluded |
| 归一化基线 | POST /api/trials/{id}/normalize | 返回各通道基线 |
| 归一化结果 | GET /api/trials/{id}/normalized | fold-change 列表 |
| 推断开关状态 | POST /api/trials/{id}/infer | 逐元件×阶段 |
| 状态列表 | GET /api/trials/{id}/states | |
| 顺序约束检查 | POST /api/trials/{id}/check-order | 生成冲突清单 |
| 冲突列表 | GET /api/trials/{id}/conflicts | |
| 发布行为版本 | POST /api/trials/{id}/versions | 绑定三份快照 |
| 版本列表/详情 | GET /api/trials/{id}/versions, GET /api/versions/{id} | |
| 共享/冻结/替代版本 | POST /api/versions/{id}/share\|freeze\|supersede | 状态机校验 |
| 统计/健康 | GET /api/stats, GET /api/health | |

## 持久化

SQLite（modernc.org/sqlite 纯 Go 驱动，CGO 无关），库文件默认 `geneclock.db`。表：circuits、elements、induction_stages、trials、readings、baselines、element_states、order_conflicts、behavior_versions。关闭后重开同一数据库，全部线路/试验/推断/冲突/版本均恢复；`--smoke-test` 最后一步通过统计验证持久化。

## 状态机

- 线路试验：preparing → collecting → pending_review → published → archived
- 读数窗口：pending → valid / saturated / missing → excluded
- 调控状态：uninduced → active / repressed → order_conflict → confirmed
- 行为版本：draft → shared → frozen → superseded（frozen 不可再流转）

## 模块责任

- internal/circuit：线路、元件、诱导阶段定义与重叠/上游校验
- internal/sampling：读数幂等导入与窗口状态流转
- internal/normalize：诱导前窗口基线归一化
- internal/infer：fold-change 阈值状态推断
- internal/constraint：表达先后顺序约束与冲突判定
- internal/version：行为版本快照、共享、冻结、替代
- internal/service：用例编排
- internal/store：SQLite 迁移与仓储
- internal/httpapi：REST 层
