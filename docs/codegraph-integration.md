# CodeGraph Integration — 代码变动即时同步

> 配套 `AGENTS.md` §CodeGraph Integration。`CodeGraph` 维护仓库知识图谱，使 AI 能检索符号、调用链与影响面；并为 Python→Go 翻译提供语义基底（翻译前用 `codegraph callees` / `codegraph impact <symbol>` 查依赖，避免漏迁被调方）。完整内容在此维护；AGENTS.md 只保留摘要。

---

## 两层触发

1. **Skill 运行时层（对用户透明）**：GCL 运行 / 代码改动经 Skill 执行且变更落盘后，运行时自动 `codegraph sync --quiet`（挂在 GCL `### Trace` 写出之后 / Skill `recover` 步骤末尾）。用户无感知。依赖运行时 PATH 含 `codegraph` 二进制。
2. **贡献者层（改 skill 规格 / 脚本的人）**：任何代码、脚本或规格变动后，**第一时间** `codegraph sync`（仓库未 `init` 时先 `codegraph init`）。sync 须覆盖 `cmd/vet/` 子目录（Go 工具同样入图）。可选：装 git post-commit hook 自动 `codegraph sync --quiet`（本地，不提交）。

## 规则

- 新克隆仓库若未索引：先 `codegraph init`，再把 `.codegraph/` 加入 `.gitignore`（本地索引，禁止提交）。
- 每次变动后 `codegraph status` 确认新增/变更文件已入图；翻译类任务前用 `codegraph query <新符号>` 验证可达。
- `codegraph` 为本仓库**已验证存在的工具**（v1.1.6，`~/.local/bin/codegraph`），非假设。
