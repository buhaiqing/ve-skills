## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md) for CLI installation.

**CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`.

## Conventions

- Output is **JSON by default**
- Use `--help` for any action: `ve functiongraph <Action> --help`
- Parameters use `--ParamName value` format
- Command prefix: `ve functiongraph`

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateFunction | Yes | Primary create method |
| GetFunction | Yes | Full support |
| ListFunctions | Yes | Full support |
| UpdateFunction | Yes | Code/config update |
| DeleteFunction | Yes | Requires user confirmation |
| InvokeFunction | Yes | Supports RequestResponse and Event |
| CreateTrigger | Yes | Timer, APIG, HTTP, CTS, RocketMQ |
| DeleteTrigger | Yes | By trigger ID |
| ListTriggers | Yes | List all triggers for a function |
| PublishVersion | Yes | Create function version |
| ListVersions | Yes | List all versions |
| CreateAlias | Yes | Create version alias |
| DeleteAlias | Yes | Delete version alias |
| ListAliases | Yes | List all aliases |
| GetFunctionLogs | Yes | Query function logs |
| GetFunctionMetrics | Yes | Query monitoring metrics |

## Command Map

| Goal | Example `ve` invocation |
|------|--------------------------|
| Create function | `ve functiongraph CreateFunction --FunctionName my-func --Runtime Python3.9 --Handler index.handler --CodeType URL --CodeUrl https://...` |
| List functions | `ve functiongraph ListFunctions --Region cn-beijing` |
| Invoke function | `ve functiongraph InvokeFunction --FunctionName my-func --InvocationType RequestResponse --Payload '{"key":"value"}'` |
| Delete function | `ve functiongraph DeleteFunction --FunctionName my-func` |
| Create trigger | `ve functiongraph CreateTrigger --FunctionName my-func --TriggerType Timer --TriggerName hourly-trigger --TimerConfig '{"Schedule":"0 * * * *","Enable":true}'` |
