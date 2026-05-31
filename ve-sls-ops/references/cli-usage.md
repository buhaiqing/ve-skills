## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md) for CLI installation.

**CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`.

## Conventions

- Command prefix: `ve tls` (SLS/TLS service)
- Output is **JSON by default**
- Use `--help` for any action: `ve tls <Action> --help`
- Parameters use `--ParamName value` format

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateProject | Yes | Create log project |
| DescribeProjects | Yes | List with filters |
| DeleteProject | Yes | Remove project |
| CreateTopic | Yes | Create log topic |
| DescribeTopics | Yes | List with TTL/shard info |
| ModifyTopic | Yes | Update TTL, shard count |
| DeleteTopic | Yes | Remove topic |
| CreateIndex | Yes | Enable full-text or key-value |
| DescribeIndex | Yes | Query index config |
| ModifyIndex | Yes | Update index settings |
| CreateShipper | Yes | Create log delivery task |
| DescribeShippers | Yes | List shippers |
| ModifyShipper | Yes | Update shipper config |
| DeleteShipper | Yes | Remove shipper |
| SearchLogs | Yes | Query log content |

## Command Map

| Goal | Example `ve` invocation |
|------|--------------------------|
| List projects | `ve tls DescribeProjects --Region cn-beijing` |
| Create topic | `ve tls CreateTopic --Region cn-beijing --ProjectId xxx --TopicName app-error --Ttl 30` |
| Create index | `ve tls CreateIndex --Region cn-beijing --ProjectId xxx --TopicId yyy --FullText '{"CaseSensitive": false}'` |
| Search logs | `ve tls SearchLogs --Region cn-beijing --ProjectId xxx --TopicId yyy --Query "error"` |
