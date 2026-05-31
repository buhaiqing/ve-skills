## OpenAPI

- Spec: https://www.volcengine.com/docs/6411
- Service ID: `tls`

## SDK Operations Map

| Goal | API operationId | SDK Method |
|------|-----------------|------------|
| Create project | CreateProject | `CreateProject` |
| Describe projects | DescribeProjects | `DescribeProjects` |
| Delete project | DeleteProject | `DeleteProject` |
| Create topic | CreateTopic | `CreateTopic` |
| Describe topics | DescribeTopics | `DescribeTopics` |
| Modify topic | ModifyTopic | `ModifyTopic` |
| Delete topic | DeleteTopic | `DeleteTopic` |
| Create index | CreateIndex | `CreateIndex` |
| Describe index | DescribeIndex | `DescribeIndex` |
| Modify index | ModifyIndex | `ModifyIndex` |
| Create shipper | CreateShipper | `CreateShipper` |
| Describe shippers | DescribeShippers | `DescribeShippers` |
| Modify shipper | ModifyShipper | `ModifyShipper` |
| Delete shipper | DeleteShipper | `DeleteShipper` |
| Search logs | SearchLogs | `SearchLogs` |

## Request / Response Notes

- Required fields: `ProjectName` for CreateProject, `TopicName` + `ProjectId` + `Ttl` for CreateTopic
- `Ttl` values: integer in days (1-365)
- Pagination: Most list operations support `PageNumber` and `PageSize`
- Log search: Use `Query` parameter for log filtering (supports Lucene syntax)
