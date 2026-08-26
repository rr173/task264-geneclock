# BENZHI 评测说明

基于 Go 实现的合成生物基因调控时序复核后端服务，一款后端服务，完成调控线路与诱导阶段编排、多通道荧光读数归一化与元件开关状态推断，并对表达先后顺序约束做复核与行为版本发布。

## 启动

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go run ./cmd/geneclock --addr :8080 --db geneclock.db
```

## 自检（不启动长驻服务）

```bash
go run ./cmd/geneclock --smoke-test
```

`--smoke-test` 会真实创建 tetR-GFP 抑制回路、导入双通道读数、归一化基线、推断开关状态、复现 late_shutdown 顺序冲突、发布并冻结行为版本，关闭并重新打开数据库验证持久化，最后以 0 退出码结束。

## 构建门禁

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet   ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test  ./...
go run ./cmd/geneclock --smoke-test
```

## HTTP API（前缀 /api）

线路：`POST /api/circuits`、`GET /api/circuits`、`GET /api/circuits/{id}`、`POST /api/circuits/{id}/archive`、`POST|GET /api/circuits/{id}/elements`、`POST|GET /api/circuits/{id}/stages`
试验：`POST|GET /api/circuits/{id}/trials`、`POST /api/trials/{id}/transition`、`PATCH /api/trials/{id}/lag`
读数：`POST /api/trials/{id}/readings`、`GET /api/trials/{id}/readings`、`PATCH /api/readings/{id}/window`
分析：`POST /api/trials/{id}/normalize`、`GET /api/trials/{id}/normalized`、`POST /api/trials/{id}/infer`、`GET /api/trials/{id}/states`、`POST /api/trials/{id}/check-order`、`GET /api/trials/{id}/conflicts`
版本：`POST /api/trials/{id}/versions`、`GET /api/trials/{id}/versions`、`GET /api/versions/{id}`、`POST /api/versions/{id}/share`、`POST /api/versions/{id}/freeze`、`POST /api/versions/{id}/supersede`
统计：`GET /api/stats`、`GET /api/health`

## 持久化

SQLite（modernc.org/sqlite，CGO 无关）。表：circuits、elements、induction_stages、trials、readings、baselines、element_states、order_conflicts、behavior_versions。读数以 `(trial_id, seq, channel)` 幂等；行为版本冻结后绑定三份 JSON 快照且不可再流转。
