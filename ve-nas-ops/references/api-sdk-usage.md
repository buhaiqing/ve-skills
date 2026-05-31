## OpenAPI

- Spec: https://www.volcengine.com/docs/6470
- Service ID: `nas`

## SDK Operations Map

| Goal | API operationId | SDK Method |
|------|-----------------|------------|
| Create file system | CreateFileSystem | `CreateFileSystem` |
| Describe file systems | DescribeFileSystems | `DescribeFileSystems` |
| Modify file system | ModifyFileSystem | `ModifyFileSystem` |
| Delete file system | DeleteFileSystem | `DeleteFileSystem` |
| Create mount target | CreateMountTarget | `CreateMountTarget` |
| Describe mount targets | DescribeMountTargets | `DescribeMountTargets` |
| Delete mount target | DeleteMountTarget | `DeleteMountTarget` |
| Create snapshot | CreateSnapshot | `CreateSnapshot` |
| Describe snapshots | DescribeSnapshots | `DescribeSnapshots` |
| Delete snapshot | DeleteSnapshot | `DeleteSnapshot` |
| Restore snapshot | RestoreSnapshot | `RestoreSnapshot` |
| Create permission group | CreatePermissionGroup | `CreatePermissionGroup` |
| Describe permission groups | DescribePermissionGroups | `DescribePermissionGroups` |

## Request / Response Notes

- Required fields: `FileSystemName`, `StorageType`, `Protocol`, `ChargeType`
- `StorageType` values: `Capacity`, `Performance`, `Extreme`
- `Protocol` values: `NFS`, `SMB`
- `ChargeType` values: `PostPaid`, `PrePaid`
- Pagination: `DescribeFileSystems` supports `PageSize` and `PageNumber`
