# M5 — 首次发布

> **父计划**：`2026-07-12-python-to-go-cli.md`
> **依赖**：M4 完成（CI 全绿、文档切换完毕）。
> **并行性**：串行。

---

## M5.1 — 打 tag 发布
- tag：`vet/v0.1.0`（触发 `.github/workflows/release.yml`，见 M1.2）。
- goreleaser 出 5 平台 Release 资产：darwin/arm64、darwin/amd64、linux/arm64、linux/amd64、windows/amd64。
- 校验：GitHub Release 页面含 5 平台二进制 + checksum。

## M5.2 — `cmd/vet/README.md` 完善
- 安装：curl 一行命令（从 Release 下载对应平台二进制到 `/usr/local/bin/vet`）。
- 用法：每个子命令示例（`vet validate`、`vet check eval --json`、`vet gcl run ...`）。
- CodeGraph 辅助翻译示例：演示"翻译前 `codegraph impact <fn>` 查依赖"。

---

## 退出标准（M5）
- [ ] GitHub Release `vet-v0.1.0` 含 5 平台资产
- [ ] 安装命令经实测可下载执行
- [ ] README 完整
- [ ] 独立 commit（tag 本身即发布点）
