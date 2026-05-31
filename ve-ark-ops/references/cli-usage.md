## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md).

## Conventions

- Command prefix: `ve ark`
- Output is JSON by default

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| ListEndpoints | Yes | List inference endpoints |
| CreateEndpoint | Yes | Create inference endpoint |
| DeleteEndpoint | Yes | Remove endpoint |
| DescribeEndpoint | Yes | Get endpoint details |
| ListModels | Yes | List available models |
| DescribeModel | Yes | Get model details |
| CreateTrainingJob | Yes | Start fine-tuning job |
| DescribeTrainingJob | Yes | Check training status |
| ListDatasets | Yes | List datasets |
| CreateDataset | Yes | Upload dataset |

## Command Map

| Goal | Example |
|------|---------|
| List endpoints | `ve ark ListEndpoints --Region cn-beijing` |
| List models | `ve ark ListModels --Region cn-beijing` |
| Create endpoint | `ve ark CreateEndpoint --ModelName doubao-pro --EndpointName my-endpoint` |
