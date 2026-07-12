# M1 — 脚手架 + 发布管线（vet）

> **父计划**：`2026-07-12-python-to-go-cli.md`
> **目标**：建立 `cmd/vet/` Go module、goreleaser + GitHub Actions 发布管线，用占位命令 `vet version` 跑通 5 平台交叉编译与 Release 流程。
> **依赖**：无（M0 已完成：仓库已 `codegraph init`，AGENTS.md 规则就位）。
> **并行性**：M1 内部**串行**（M1.1 module 是 M1.2 goreleaser 的构建依赖）。

---

## M1.1 — 建立 `cmd/vet/` Go module

### 交付物
- `cmd/vet/go.mod`：`module github.com/buhaiqing/ve-skills/cmd/vet`，Go 1.22+，零第三方依赖。
- `cmd/vet/main.go`：子命令路由骨架 + 占位 `version` 子命令。
- `cmd/vet/README.md`：工具说明（安装/用法留 M5 补完）。

### main.go 骨架要求
- 使用标准库 `flag` 包做子命令路由（`flag.NewFlagSet` 每子命令一个）。
- 顶层命令：`vet version`、`vet check`、`vet gcl`、`vet validate`。
- `vet version` 输出：版本号（常量 `version = "0.0.0-dev"`）+ commit + 构建时间（可从 `runtime/debug.ReadBuildInfo` 取，可选）。
- 未知/缺子命令 → 打印 usage 到 stderr，exit 2。

### 等价判据
- `go build ./...` 在 darwin/linux/windows 均通过（交叉编译 `GOOS=windows go build` 等）。
- `go test ./...` 通过（此时无测试，空过）。
- `./vet version` 输出含版本字符串，exit 0。

### 验证
```bash
cd cmd/vet
go build -o /tmp/vet . && /tmp/vet version
GOOS=windows GOARCH=amd64 go build -o /tmp/vet.exe . && echo "windows ok"
```

---

## M1.2 — goreleaser + GitHub Actions 发布管线

### 交付物
- `.goreleaser.yaml`：build 5 平台（darwin/arm64、darwin/amd64、linux/arm64、linux/amd64、windows/amd64），binary 名 `vet`，含 `version`/`commit`/`date` ldflags。
- `.github/workflows/release.yml`：on `push.tags: ["vet/v*"]`，步骤：checkout（fetch-depth 0）→ setup-go → goreleaser `--clean`，发 GitHub Release + 资产。
- 可选：`.github/workflows/test.yml` 或并入现有 validate.yml：PR 时 `cd cmd/vet && go build ./... && go test ./...`。

### 要求
- goreleaser 配置 `archives` 含 `checksum`。
- tag 格式 `vet/v0.1.0` → Release 名 `vet-v0.1.0`。
- 不提交 token；用 `GITHUB_TOKEN`（Actions 默认提供）。

### 验证（不实际发 Release）
```bash
cd cmd/vet
goreleaser build --snapshot --clean   # 本地出 5 平台二进制到 dist/
goreleaser check                       # 配置校验
```
- 退出：本地 `goreleaser build --snapshot` 成功生成 5 平台产物。

---

## M1.3 — CodeGraph 收录 `cmd/vet/`

### 交付物
- `cmd/vet/` 初始结构已入 CodeGraph 图谱。

### 步骤
```bash
codegraph sync --quiet
codegraph query vet        # 应命中 main.go 的 version 等相关符号
codegraph status           # Files 数应较 M0 的 37 增加
```

### 等价判据
- `codegraph query vet` 返回非空（main.go 符号已索引）。
- `codegraph status` 的 Files 计数 > 37（含新增 go 文件）。

---

## 退出标准（M1 整体）
- [ ] `go build ./...` + 三平台交叉编译通过
- [ ] `./vet version` 正常输出
- [ ] `goreleaser build --snapshot` 出 5 平台二进制
- [ ] `goreleaser check` 通过
- [ ] CodeGraph 含 `cmd/vet/` 节点
- [ ] 独立 commit（M1.1/M1.2/M1.3 可分 commit，建议 M1.1+1.2 合并一 commit、M1.3 一 commit）

---

## 备注
- 本里程碑不动任何 Python 逻辑，仅搭骨架。真实功能迁移在 M2/M3。
- `vet` 名与 `go vet` 命令同名但作用域不同（本工具仅作 ve-skills 仓库子命令），已在计划中确认采用。
