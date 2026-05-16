# CLI — VKE (`ve vke`)

## Install and Config

- Install: `ve` CLI from https://github.com/volcengine/volcengine-cli/releases
- Credentials: `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION`
- Output: **JSON by default**

## Command Map

| Goal | CLI Command | Notes |
|------|-------------|-------|
| List K8s versions | `ve vke ListSupportedVersions --Region cn-beijing` | Check available versions |
| Create cluster | `ve vke CreateCluster --Region cn-beijing --Name my-cluster ...` | JSON output |
| Describe cluster | `ve vke DescribeCluster --ClusterId cls-xxx` | Full details |
| List clusters | `ve vke ListClusters --Region cn-beijing` | Paginated |
| Update cluster config | `ve vke UpdateClusterConfig --ClusterId cls-xxx --body '{"DeleteProtectionEnabled": true}'` | JSON body |
| Delete cluster | `ve vke DeleteCluster --ClusterId cls-xxx` | **Irreversible** |
| Create node pool | `ve vke CreateNodePool --ClusterId cls-xxx --Name pool-1 ...` | JSON output |
| Describe node pool | `ve vke DescribeNodePool --ClusterId cls-xxx --NodePoolId np-xxx` | Details |
| Update node pool | `ve vke UpdateNodePool --ClusterId cls-xxx --NodePoolId np-xxx --body '{}'` | JSON body |
| Delete node pool | `ve vke DeleteNodePool --ClusterId cls-xxx --NodePoolId np-xxx` | **Irreversible** |
| List nodes | `ve vke ListNodes --ClusterId cls-xxx --NodePoolId np-xxx` | Node details |
| Add nodes | `ve vke AddNodes --ClusterId cls-xxx --NodePoolId np-xxx --InstanceIds '["i-xxx"]'` | Add ECS nodes |
| Remove nodes | `ve vke RemoveNodes --ClusterId cls-xxx --NodePoolId np-xxx --InstanceIds '["i-xxx"]'` | Remove without delete |
| Delete nodes | `ve vke DeleteNodes --ClusterId cls-xxx --NodePoolId np-xxx --InstanceIds '["i-xxx"]'` | **Destroys ECS** |

## Parameter Discovery

```bash
ve vke --help                    # List all VKE actions
ve vke CreateCluster --help      # Show CreateCluster parameters
ve vke DescribeCluster --help    # Show DescribeCluster parameters
```

## JSON Output Examples

```bash
# Parse cluster ID from create response
ve vke CreateCluster --Name my-cluster ... | jq -r '.Result.ClusterId'

# Parse cluster status
ve vke DescribeCluster --ClusterId cls-xxx | jq -r '.Result.Status'

# List cluster names
ve vke ListClusters --Region cn-beijing | jq -r '.Result.Items[].Name'
```
