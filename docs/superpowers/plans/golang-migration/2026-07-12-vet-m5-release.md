# M5 — 首次发布

> **父计划**：`2026-07-12-python-to-go-cli.md`
> **依赖**：M4 完成（CI 全绿、文档切换完毕）。
> **并行性**：串行。

---

## M5.1 — 打 tag 发布
- tag：`v0.1.0`（semver；触发 `.github/workflows/release.yml` 的 `v*` 规则，见 M1.2）。
  - 注：早期误用 `vet/v0.1.0` 前缀导致 goreleaser `not semver` 失败，已统一改为 `vX.Y.Z`。
- goreleaser 出 6 平台 Release 资产：
  - `vet_X.Y.Z_darwin_{amd64,arm64}.tar.gz`
  - `vet_X.Y.Z_linux_{amd64,arm64}.tar.gz`
  - `vet_X.Y.Z_windows_{amd64,arm64}.zip`
- 校验：GitHub Release 页面含 6 平台二进制 + `checksums.txt`（sha256）。
- 发布方式（二选一，均已验证）：
  - 自动化：`make release VERSION=0.1.x` → `git tag` + `git push` → 触发 GitHub Action 全自动构建发布（**推荐，免人工**）。
  - 本地直发：`make release-api`（需 `.env` 中 `GITHUB_TOKEN`）→ 直接 `goreleaser release --clean`。

## M5.2 — `cmd/vet/install.sh` 一键安装/更新
- 脚本始终解析 **latest** Release（`/releases/latest`），永远指向最新版本，重跑即更新。
- 安装：`curl -fsSL https://raw.githubusercontent.com/buhaiqing/ve-skills/main/cmd/vet/install.sh | bash`
  （默认装到 `/usr/local/bin/vet`；可用 `INSTALL_DIR=/path bash install.sh` 自定义，目录需预建以保证可写、避免 sudo）。
- 用法示例：`vet validate --root .`、`vet check eval --json`、`vet gcl run ...`（详见 `cmd/vet/README.md`）。

---

## 退出标准（M5）
- [x] GitHub Release `v0.1.0` 含 6 平台资产（run 29195410724 success）
- [x] GitHub Release `v0.1.1` 含 6 平台资产（run 29196098814 success，tag 触发全自动化）
- [x] `install.sh` 一键安装 + 始终指向 latest，实测下载安装 `v0.1.1` 成功（`vet 0.1.1`，frontmatter 29/29 通过）
- [x] `make release` / `make release-api` 两条发布路径均可用
- [x] 代码已 force-push 至 `origin/main`（`6d69585`）；原 82 个 origin/main 提交已备份至 `backup/pre-go-migration-main`
- [x] 安装命令经实测可下载执行（见 M5.2）

## 状态
✅ **完成**（2026-07-12）。`v0.1.0` 与 `v0.1.1` 均通过 GitHub Action 全自动发布，6 平台资产齐全。
所有 golang-migration 计划（M1–M5）已交付。
