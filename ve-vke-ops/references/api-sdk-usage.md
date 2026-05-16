# VKE API & SDK Usage

## OpenAPI

- **Base URL:** `https://open.volcengineapi.com`
- **Service:** `vke`
- **Doc:** https://www.volcengine.com/docs/6460

## SDK Operations Map

| Goal | API Action | SDK Method | Required Fields |
|------|-----------|------------|----------------|
| Create Cluster | `CreateCluster` | `instance.Client.Request("vke", "CreateCluster", params)` | Name, ClusterConfig.SubnetIds, PodsConfig, ServicesConfig |
| Describe Cluster | `DescribeCluster` | `instance.Client.Request("vke", "DescribeCluster", params)` | ClusterId |
| List Clusters | `ListClusters` | `instance.Client.Request("vke", "ListClusters", params)` | Region (optional: PageNumber, PageSize, Name) |
| Update Cluster Config | `UpdateClusterConfig` | `instance.Client.Request("vke", "UpdateClusterConfig", params)` | ClusterId, config fields |
| Delete Cluster | `DeleteCluster` | `instance.Client.Request("vke", "DeleteCluster", params)` | ClusterId |
| Create NodePool | `CreateNodePool` | `instance.Client.Request("vke", "CreateNodePool", params)` | ClusterId, Name, NodeConfig |
| Describe NodePool | `DescribeNodePool` | `instance.Client.Request("vke", "DescribeNodePool", params)` | ClusterId, NodePoolId |
| Update NodePool | `UpdateNodePool` | `instance.Client.Request("vke", "UpdateNodePool", params)` | ClusterId, NodePoolId |
| Delete NodePool | `DeleteNodePool` | `instance.Client.Request("vke", "DeleteNodePool", params)` | ClusterId, NodePoolId |
| Add Nodes | `AddNodes` | `instance.Client.Request("vke", "AddNodes", params)` | ClusterId, NodePoolId, InstanceIds |
| Remove Nodes | `RemoveNodes` | `instance.Client.Request("vke", "RemoveNodes", params)` | ClusterId, NodePoolId, InstanceIds |
| Delete Nodes | `DeleteNodes` | `instance.Client.Request("vke", "DeleteNodes", params)` | ClusterId, NodePoolId, InstanceIds |
| List K8s Versions | `ListSupportedVersions` | `instance.Client.Request("vke", "ListSupportedVersions", params)` | Region |

## Response JSON Paths

| Operation | Success Path | Key Fields |
|-----------|-------------|------------|
| CreateCluster | `$.Result.ClusterId` | ClusterId |
| DescribeCluster | `$.Result.*` | Name, Status, KubernetesVersion, Endpoints |
| ListClusters | `$.Result.Items[]` | ClusterId, Name, Status, CreateTime |
| CreateNodePool | `$.Result.NodePoolId` | NodePoolId |
| DescribeNodePool | `$.Result.*` | Name, Status, AutoScaling, KubernetesConfig |

## Pagination

ListClusters supports pagination:
- `PageNumber`: Starting from 1
- `PageSize`: Default 10, max 100
- Response includes `$.Result.Page.TotalCount` for total items

## Go SDK Package

```go
import "github.com/volcengine/volc-sdk-golang/service/vke"

instance := vke.NewInstance()
instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
resp, err := instance.Client.Request("vke", "CreateCluster", params)
```
