## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md) for CLI installation.

**CRITICAL Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`.

## Conventions

- Output is **JSON by default**
- Use `--help` for any action: `ve nas <Action> --help`
- Parameters use `--ParamName value` format

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateFileSystem | Yes | Create NFS/SMB file system |
| DescribeFileSystems | Yes | List with filters |
| ModifyFileSystem | Yes | Update attributes |
| DeleteFileSystem | Yes | Requires no active mount targets |
| CreateMountTarget | Yes | Create mount point |
| DescribeMountTargets | Yes | List with FS filter |
| DeleteMountTarget | Yes | Remove mount point |
| CreatePermissionGroup | Yes | Manage access rules |
| DescribePermissionGroups | Yes | List permission groups |
| CreateSnapshot | Yes | Create FS snapshot |
| DescribeSnapshots | Yes | List snapshots |
| DeleteSnapshot | Yes | Remove snapshot |
| RestoreSnapshot | Yes | Restore FS from snapshot |

## Command Map

| Goal | Example `ve` invocation |
|------|--------------------------|
| List file systems | `ve nas DescribeFileSystems --Region cn-beijing` |
| Create file system | `ve nas CreateFileSystem --Region cn-beijing --FileSystemName my-fs --StorageType Capacity --Protocol NFS --ChargeType PostPaid` |
| Create mount target | `ve nas CreateMountTarget --Region cn-beijing --FileSystemId enas-xxx --VpcId vpc-yyy --SubnetId subnet-zzz` |
| Create snapshot | `ve nas CreateSnapshot --Region cn-beijing --FileSystemId enas-xxx --SnapshotName pre-upgrade-backup` |
