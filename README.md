# ve-skills

火山引擎（Volcengine）相关的 Agent Skills 集合。

## 概述

本项目是火山引擎（Volcengine）运维 Agent Skills 集合，提供云产品的自动化运维、监控和管理能力。

> **Skills Farm 是一套 Meta Skill（元技能）体系**——将运维知识转化为结构化的、AI Agent 可解析、可执行、可验证的声明式规范。

### 关键特性

| 特性 | 说明 |
|------|------|
| **占位符机制** | `{{env.*}}`（环境变量）、`{{user.*}}`（用户输入）、`{{output.*}}`（输出捕获），实现人机双通道 |
| **职责委托** | `SHOULD/SHOULD NOT Use` 定义边界，跨产品操作自动委派 |
| **生成器** | 基于 OpenAPI 规范自动生成 Skill 框架模板，支持人工审核和完善 |
| **CLI-first 执行** | 优先使用 `ve` CLI（静态 Go 二进制），CLI 不支持时 JIT 构建 Go SDK 脚本 |
| **安全机制** | 凭证隔离（`{{env.*}}` 不暴露）、操作安全门（删除/恢复需确认） |
| **跨平台设计** | 基于标准 Markdown + OpenSpec，支持多种 Agent 框架接入 |

## 项目结构

```
ve-skills/
├── README.md                                    # 本文件
├── .gitignore                                   # Git 排除规则
├── ve-skill-generator/                          # Skill 生成器（Meta Skill）
│   ├── SKILL.md
│   ├── assets/
│   │   ├── example-config.yaml
│   │   └── eval_queries.json
│   └── references/
│       ├── ve-skill-template.md                 # Skill 模板
│       ├── execution-environment.md             # CLI + Go SDK 环境配置
│       ├── cli-behavior.md                      # ve CLI 行为规范
│       ├── governance-and-adversarial-review.md # 治理与对抗审查
│       ├── enhanced-self-healing-framework.md   # 自愈框架
│       ├── user-experience-spec.md              # UX 规范
│       ├── optimization-analysis.md             # 三维度优化分析
│       ├── prompt-library.md                    # 结构化 prompt 库
│       └── aiops-best-practices.md              # AIOps 最佳实践
├── ve-ecs-ops/                                  # 云服务器 ECS（待生成）
├── ve-rds-ops/                                  # 云数据库 RDS（待生成）
├── ve-redis-ops/                                # 云数据库 Redis（待生成）
├── ve-vpc-ops/                                  # 私有网络 VPC（待生成）
└── ve-slb-ops/                                  # 负载均衡 SLB（待生成）
```

## 快速开始

### 1. 安装 ve CLI

```bash
# 从 GitHub Releases 下载（以 Linux x86_64 为例）
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# 验证安装
ve version
```

### 2. 配置凭证

```bash
# 方式一：环境变量（推荐）
export VOLCENGINE_ACCESS_KEY="{{env.VOLCENGINE_ACCESS_KEY}}"
export VOLCENGINE_SECRET_KEY="{{env.VOLCENGINE_SECRET_KEY}}"
export VOLCENGINE_REGION="cn-beijing"

# 方式二：交互式配置
ve configure set --profile default --region cn-beijing --access-key ak --secret-key sk
```

### 3. 生成新 Skill

在 Agent Runtime 中引用生成器，然后提供提示词：

> "生成火山引擎 ECS 的 Skill，名称 ve-ecs-ops，核心功能：实例生命周期管理、磁盘、快照"

**生成结构**：
```
ve-ecs-ops/
├── SKILL.md
├── references/
│   ├── cli-usage.md
│   ├── api-sdk-usage.md
│   ├── core-concepts.md
│   ├── integration.md
│   ├── monitoring.md
│   └── troubleshooting.md
└── assets/
    └── example-config.yaml
```

## 火山引擎 CLI 行为特征

### 正确 CLI 调用模式

```bash
# 基本 API 调用
ve <service> <action> --<parameter> value

# 示例
ve ecs DescribeInstances --Region cn-beijing
ve rds_mysql ListDBInstanceIPLists --InstanceId "xxxxxx"

# JSON 参数传递
ve rds_mysql ModifyDBInstanceIPList --body '{"InstanceId":"xxx", "GroupName": "xxx"}'

# 查看帮助
ve ecs DescribeInstances --help
```

### ve CLI 安装

**从 GitHub Releases 下载：**

| 平台 | 二进制名称 |
|------|-----------|
| macOS (Apple Silicon) | `ve-darwin-arm64` |
| macOS (Intel) | `ve-darwin-amd64` |
| Linux x86_64 | `ve-linux-amd64` |
| Linux ARM64 | `ve-linux-arm64` |

下载地址：https://github.com/volcengine/volcengine-cli/releases

> 注意：从 v1.0.20 开始，命令前缀由 `volcengine-cli` 更新为 `ve`。

### 凭证配置

**环境变量（Agent 执行推荐）：**
```bash
export VOLCENGINE_ACCESS_KEY="..."
export VOLCENGINE_SECRET_KEY="..."
export VOLCENGINE_REGION="cn-beijing"
```

**交互式配置：**
```bash
ve configure set --profile default --region cn-beijing --access-key ak --secret-key sk
```

**配置文件：**
```json
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "YOUR_AK",
      "secret_key": "YOUR_SK",
      "region": "cn-beijing"
    }
  ]
}
```

配置文件路径：`~/.volcengine/config.json`

## Skill 编写要点

- CLI 示例：用 `bash`，JSON 用 `json`，YAML 用 `yaml`
- 表格展示：产品列表、监控指标、告警阈值
- 凭证配置见上方环境变量章节

## 参考资源

- [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- [Volcengine SDK for Go](https://github.com/volcengine/volc-sdk-golang)
- [Agent Skills OpenSpec](https://agentskills.io/specification)
- [火山引擎帮助文档](https://www.volcengine.com/docs)
- [ve-skill-generator/SKILL.md](ve-skill-generator/SKILL.md) — 生成器的完整使用说明
