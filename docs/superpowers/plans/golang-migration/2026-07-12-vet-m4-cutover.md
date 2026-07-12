# M4 — 切换 + 清理

> **父计划**：`2026-07-12-python-to-go-cli.md`
> **依赖**：M2 + M3 全部完成（`vet` 功能等价替换所有 Python 脚本）。
> **并行性**：串行（依赖前序全完成）。

---

## M4.1 — CI 改调 `vet`
- 文件：`.github/workflows/validate.yml`
- 变更：将 `python3 scripts/validate_local.py`、`check_*.py`、`gcl_*.py` 调用替换为对应 `vet` 子命令。
- 前置：CI 环境需有 `vet` 二进制。方案：在 validate.yml 加一步下载/构建 `vet`（从 tag 发布资产或 `go build ./cmd/vet`）。
- 验证：PR 触发 CI 全绿，行为与旧 Python 一致。

## M4.2 — 旧 `scripts/*.py` 处理
- 默认：**保留并标 deprecated**（文件头加 `# DEPRECATED: superseded by cmd/vet — run \`vet <subcommand>\``）。
- 过渡期后删除（待用户拍板，不在本里程碑强删）。
- `gcl_critic_stub.py` 已被 M3.1 吸收，可删（其能力在 `--critic-stdin`）。
- `*_test.py` 已转 Go `_test.go`，可删。

## M4.3 — 文档指引切换
- AGENTS.md：将"运行 `python3 scripts/validate_local.py`"等指引改为 `vet validate` / `vet check ...`（含 GCL 节补"AI 翻译前先 `codegraph impact`"）。
- `ve-skill-generator/SKILL.md` / `references/`：同步命令引用。
- 旧计划文档 `2026-05-27-finops-aiops-optimization.md` 的 61 复选框：本里程碑顺带核实并勾（如仍过时则单列待办，不阻塞）。
- CodeGraph：本里程碑改动后 `codegraph sync`。

---

## 退出标准（M4）
- [ ] CI 全绿（用 `vet`）
- [ ] 文档引用一致，无孤儿 `scripts/*.py` 调用（Grep 确认）
- [ ] CodeGraph 含最终结构
- [ ] 独立 commit
