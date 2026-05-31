## OpenAPI

- Spec: https://www.volcengine.com/docs/6668
- Base path: `https://open.volcengineapi.com`
- Service ID: `functiongraph`

## SDK Operations Map

| Goal | API operationId | SDK Method |
|------|-----------------|------------|
| Create function | CreateFunction | `CreateFunction` |
| Get function | GetFunction | `GetFunction` |
| List functions | ListFunctions | `ListFunctions` |
| Update function | UpdateFunction | `UpdateFunction` |
| Delete function | DeleteFunction | `DeleteFunction` |
| Invoke function | InvokeFunction | `InvokeFunction` |
| Create trigger | CreateTrigger | `CreateTrigger` |
| Delete trigger | DeleteTrigger | `DeleteTrigger` |
| List triggers | ListTriggers | `ListTriggers` |
| Publish version | PublishVersion | `PublishVersion` |
| List versions | ListVersions | `ListVersions` |
| Create alias | CreateAlias | `CreateAlias` |
| Delete alias | DeleteAlias | `DeleteAlias` |
| List aliases | ListAliases | `ListAliases` |

## Request / Response Notes

- Required fields: `FunctionName`, `Runtime`, `Handler`, `CodeType`
- `CodeType` values: `URL`, `Zip`, `Image`
- Pagination: `ListFunctions` and `ListTriggers` support `PageSize` and `PageNumber` parameters
