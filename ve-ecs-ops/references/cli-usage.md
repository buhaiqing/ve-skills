# CLI Usage — ECS (`ve`)

## Install and Config

See [Execution Environment Setup](../references/execution-environment.md) for CLI installation.

**CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`.

## Conventions

- Output is **JSON by default** — no `--output json` needed
- Use `--help` for any action: `ve ecs <Action> --help`
- Parameters use `--ParamName value` format
- JSON arrays: `--InstanceIds '["i-xxx","i-yyy"]'`

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| RunInstances | Yes | Primary create method |
| DescribeInstances | Yes | Full support |
| StartInstance / StartInstances | Yes | Single/batch |
| StopInstance / StopInstances | Yes | Single/batch, supports ForceStop |
| RebootInstance / RebootInstances | Yes | Single/batch |
| DeleteInstance / DeleteInstances | Yes | Single/batch |
| ModifyInstanceSpec | Yes | Requires stopped instance |
| ModifyInstanceAttribute | Yes | Name, description, password, hostname |
| CreateSnapshot | Yes | |
| DescribeSnapshots | Yes | |
| CreateImage | Yes | |
| DescribeImages | Yes | |
| DeleteImage | Yes | |
| CreateKeyPair | Yes | |
| DescribeKeyPairs | Yes | |
| ImportKeyPair | Yes | |
| DeleteKeyPair | Yes | |
| AttachVolume | Yes | |
| DetachVolume | Yes | |
| CreateVolume | Yes | |
| DescribeVolumes | Yes | |
| DeleteVolume | Yes | |
| AssignPrivateIpAddresses | Yes | |
| DescribeRegions | Yes | |
| DescribeZones | Yes | |
| DescribeInstanceTypes | Yes | |
| AuthorizeSecurityGroupIngress | Yes | |
| DescribeResourceQuota | Yes | |

## Command Map

| Goal | Example `ve` Invocation | Notes |
|------|------------------------|-------|
| List instances | `ve ecs DescribeInstances --Region cn-beijing` | JSON output |
| Filter by status | `ve ecs DescribeInstances --Region cn-beijing --Status RUNNING` | |
| Create instance | `ve ecs RunInstances --Region cn-beijing --InstanceType ecs.g3i.large --ImageId image-xxx --VpcId vpc-xxx --SubnetId subnet-xxx --InstanceName myserver --Password 'StrongP@ss123!' --ChargeType PostPaid` | All required params |
| Start instance | `ve ecs StartInstance --Region cn-beijing --InstanceId i-xxx` | |
| Stop instance (soft) | `ve ecs StopInstance --Region cn-beijing --InstanceId i-xxx` | |
| Stop instance (hard) | `ve ecs StopInstance --Region cn-beijing --InstanceId i-xxx --ForceStop true` | |
| Reboot instance | `ve ecs RebootInstance --Region cn-beijing --InstanceId i-xxx` | |
| Delete instance | `ve ecs DeleteInstance --Region cn-beijing --InstanceId i-xxx` | Must be stopped |
| Resize instance | `ve ecs ModifyInstanceSpec --Region cn-beijing --InstanceId i-xxx --InstanceType ecs.g3i.xlarge` | Must be stopped |
| Create snapshot | `ve ecs CreateSnapshot --Region cn-beijing --VolumeId vol-xxx --SnapshotName my-snapshot` | |
| List snapshots | `ve ecs DescribeSnapshots --Region cn-beijing` | |
| Create image | `ve ecs CreateImage --Region cn-beijing --InstanceId i-xxx --Name my-image` | |
| List images | `ve ecs DescribeImages --Region cn-beijing` | |
| Create key pair | `ve ecs CreateKeyPair --Region cn-beijing --KeyPairName my-key` | Save private key immediately |
| List key pairs | `ve ecs DescribeKeyPairs --Region cn-beijing` | |
| List regions | `ve ecs DescribeRegions` | |
| List zones | `ve ecs DescribeZones --Region cn-beijing` | |
| List instance types | `ve ecs DescribeInstanceTypes --InstanceTypeFamilyIds '["g3i"]'` | |
| Check quota | `ve ecs DescribeResourceQuota --Region cn-beijing` | |
