基于 Go 实现的合成生物基因调控时序复核 Web 项目，一款后端服务，完成调控线路与诱导阶段编排、多通道荧光读数归一化与元件开关状态推断，并对表达先后顺序约束做复核与行为版本发布。

# BENZHI 评测说明：task264-geneclock

合成生物基因调控时序复核台（全栈 Web 应用，纯 Go 后端，无外部服务依赖）。

## 构建与运行

```bash
# 构建（CGO 关闭，纯 Go SQLite 驱动）
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
# 静态检查
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
# 单元测试
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
# 端到端自检（--smoke-test 契约）
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/geneclock --smoke-test
# 常驻服务
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/geneclock --addr :8080 --db geneclock.db
```

## --smoke-test 契约

不以长驻服务方式运行，而是：

1. 创建调控线路（tetR-GFP 抑制回路）。
2. 添加上游感应元件 sensor 与下游报告元件 gfp，添加「诱导激活」与「撤除恢复」两个诱导阶段。
3. 创建试验并进入采集中，按通道导入 16 条荧光读数。
4. 归一化报告基线（诱导前窗口均值），推断 2 元件 × 2 阶段共 4 条开关状态。
5. 执行顺序约束检查：撤除阶段上游 sensor 平均 fold ≈ 2.18 仍未关闭，命中 late_shutdown 冲突。
6. 标记一个饱和读数后复检冲突仍复现，试验发布，生成行为版本并冻结，验证冻结版本不可再流转。
7. 通过统计接口验证持久化恢复（线路/试验/版本均可见）。

任一步失败则以非零退出码结束并输出 `smoke-test FAILED`；全部通过输出 `smoke-test PASSED` 且退出码 0。

## 关键 API

| 能力 | 入口 |
|---|---|
| 创建线路 | POST /api/circuits |
| 添加元件 | POST /api/circuits/{id}/elements |
| 添加诱导阶段 | POST /api/circuits/{id}/stages |
| 创建试验 | POST /api/circuits/{id}/trials |
| 导入读数 | POST /api/trials/{id}/readings |
| 归一化 | POST /api/trials/{id}/normalize |
| 推断状态 | POST /api/trials/{id}/infer |
| 顺序检查 | POST /api/trials/{id}/check-order |
| 发布版本 | POST /api/trials/{id}/versions |
| 冻结版本 | POST /api/versions/{id}/freeze |
| 统计 | GET /api/stats |

## Docker

```bash
docker build -f Dockerfile -t task264-geneclock .
docker run --rm task264-geneclock --smoke-test
# 双架构
bash build_benzhi_docker.sh task264-geneclock linux/arm64
bash build_benzhi_docker.sh task264-geneclock linux/amd64
```

## 版本锁

- Go 1.26.3（GOTOOLCHAIN=local，go.mod `go 1.26.3`）
- SQLite 3.46.1（modernc.org/sqlite v1.52.0，纯 Go，CGO_ENABLED=0 可构建）
- component-versions.json 与 go.mod / Dockerfile 完全一致
