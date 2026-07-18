# CodeGraph Integration — 代码变动即时同步

> 配套 `AGENTS.md` §CodeGraph Integration。`CodeGraph` 维护仓库知识图谱，使 AI 能检索符号、调用链与影响面；并为 Python→Go 翻译提供语义基底（翻译前用 `codegraph callees` / `codegraph impact <symbol>` 查依赖，避免漏迁被调方）。完整内容在此维护；AGENTS.md 只保留摘要。

---

## 概述

CodeGraph (`codegraph` CLI, `~/.local/bin/codegraph`) 维护仓库知识图谱。本仓库已配置 MCP Server（`.mcp.json`），Agent 启动时自动获得 `codegraph_explore` 工具。

---

## MCP 配置（项目级，已配置）

```json
// .mcp.json
{
  "mcpServers": {
    "codegraph": {
      "type": "stdio",
      "command": "codegraph",
      "args": ["serve", "--mcp"],
      "description": "CodeGraph — AST-based code intelligence"
    }
  }
}
```

**Agent 使用方式**：直接调用 `codegraph_explore` MCP 工具（一个调用即可获取符号定义 + 调用链 + 影响面 + 源文件内容，等效于多次 grep/read）。CLI 等价命令：`codegraph explore <symbol>`。

---

## 两层触发

1. **Skill 运行时层（对用户透明）**：GCL 运行 / 代码改动经 Skill 执行且变更落盘后，运行时自动 `codegraph sync --quiet`（挂在 GCL `### Trace` 写出之后 / Skill `recover` 步骤末尾）。用户无感知。依赖运行时 PATH 含 `codegraph` 二进制。
2. **贡献者层（改 skill 规格 / 脚本的人）**：任何代码、脚本或规格变动后，**第一时间** `codegraph sync`（仓库未 `init` 时先 `codegraph init`）。sync 须覆盖 `cmd/vet/` 子目录（Go 工具同样入图）。可选：装 git post-commit hook 自动 `codegraph sync --quiet`（本地，不提交）。

---

## MANDATORY: 每次代码变更后必须 sync

> **铁律 — 不可打破。** 每次 Go 代码变更提交前，**必须**执行 `codegraph sync --quiet` 确保索引最新。

**Why**: 过期索引导致调用链分析和影响面判断错误。不 sync = 索引不可信 = 后续所有查询无效。

**强制执行**：
```bash
# 每次代码变更后、提交前（Agent 自动执行）
cd <repo-root> && codegraph sync --quiet
```

**验证**：
```bash
# 确认索引最新
codegraph status | grep -q "up to date" || { echo "INDEX STALE"; codegraph sync --quiet; }
```

---

## MANDATORY: CodeGraph MCP 优先于 grep/read

> **强规则：所有代码理解任务必须先用 `codegraph explore <symbol>`，再用 grep/read 补充。**

**为什么**：CodeGraph 的 AST + 调用图覆盖了 grep 无法到达的跳转表（接口实现、动态派送、跨包调用）。纯文本搜索会漏掉真实调用者。

**强制执行顺序**：
1. `codegraph explore <symbol>` → 获取符号定义 + 调用者 + 影响面 + 源文件内容（带行号）
2. 仅在 CodeGraph 索引缺失或不确定时用 grep/read 交叉验证
3. 修改代码前必须用 CodeGraph 确认所有调用方都已知

**禁用**：用 grep 替代 CodeGraph 做调用链分析；用 read 替代 codegraph explore 做符号定位（CodeGraph 更快更准，且返回带行号的真实磁盘内容）。

**例外**：纯文本内容搜索（如日志关键字、文档内容）除外。

---

## 常见场景命令模板

| 场景 | 命令 | 说明 |
|------|------|------|
| **查符号定义+调用者** | `codegraph explore <pkg.Symbol>` | 返回定义、调用者、影响面、源文件 |
| **查影响面** | `codegraph impact <pkg.Symbol>` | 仅返回被哪些符号依赖 |
| **查调用链（被调方）** | `codegraph callees <pkg.Symbol>` | 返回该符号调用了哪些函数 |
| **索引状态** | `codegraph status` | 确认索引是否最新 |
| **全量重索引** | `codegraph sync --quiet` | 增量同步变更文件 |
| **初始化** | `codegraph init` | 首次使用（`.codegraph/` 已 gitignored） |

---

## 开发工作流集成

```
代码变更 → codegraph sync --quiet → codegraph explore <变更符号> 验证 → 提交
    │                                                      │
    └── 确认影响面（所有调用方已评估）←──────────────────────┘
```

**Agent 自动执行**（每次编码任务结束时）：
1. `codegraph sync --quiet` — 同步索引
2. 对本轮新增/修改的核心符号执行 `codegraph explore <symbol>` — 验证可达性
3. 检查 blast radius 中的调用方是否都已评估

**用户可选**（本地开发环境）：
```bash
# 安装 git post-commit hook（一次性设置）
cat > .git/hooks/post-commit <<'HOOK'
#!/bin/bash
codegraph sync --quiet 2>/dev/null
HOOK
chmod +x .git/hooks/post-commit
```

---

## 规则

- 新克隆仓库若未索引：先 `codegraph init`，再把 `.codegraph/` 加入 `.gitignore`（本地索引，禁止提交）。
- 每次变动后 `codegraph status` 确认新增/变更文件已入图；翻译类任务前用 `codegraph query <新符号>` 验证可达。
- `codegraph` 为本仓库**已验证存在的工具**（v1.1.6，`~/.local/bin/codegraph`），非假设。
