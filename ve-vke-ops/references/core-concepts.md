# VKE Core Concepts

## Architecture

Volcengine VKE provides managed Kubernetes clusters with the following hierarchy:

```
VKE Cluster
├── Kubernetes Version (v1.24 - v1.30+)
├── Network Config (VPC, Subnets, VPC-CNI)
├── Service CIDR
├── Node Pool 1
│   ├── Node 1 (ECS instance)
│   ├── Node 2 (ECS instance)
│   └── ...
├── Node Pool 2 (different spec/zone)
│   └── ...
└── Add-ons (CoreDNS, kube-proxy, etc.)
```

## Cluster Types

| Type | Description | Control Plane | Use Case |
|------|-------------|---------------|----------|
| Managed Cluster | Volcengine manages control plane | Fully managed | Production workloads |
| Registered Cluster | Self-managed K8s registered to VKE | User managed | Hybrid/multi-cloud |

## Node Pool Management

Node pools group nodes with identical configurations:
- **Instance types**: ECS specifications (e.g., ecs.g1ie.xlarge)
- **Auto-scaling**: Min/max replicas, priority
- **Image**: veLinux or custom images
- **Storage**: System volume + data volumes (ESSD_PL0/PL1/PL2)
- **Security**: Security groups, login credentials, HIDS
- **Kubernetes**: Labels, taints, cordoning

## Network Modes

| Mode | Description | Performance | Isolation |
|------|-------------|-------------|-----------|
| VpcCniShared | Pods share VPC network | High | Pod-level |
| VpcCniExclusive | Each pod gets dedicated ENI | Highest | Full network isolation |
| Flannel | Overlay network | Moderate | Namespace-level |

## Kubernetes Versions

Supported versions evolve. Check with `ve vke ListSupportedVersions`. Version format: `v1.28`, `v1.29`, `v1.30`.

## Resource Relationships

- **Cluster** depends on → VPC, Subnets (network)
- **Node Pool** depends on → Cluster, ECS specs, Security Groups
- **Node** depends on → Node Pool, ECS instance
- **Pod** depends on → Node resources (CPU, memory)

## Delete Protection

Clusters and node pools support delete protection (`DeleteProtectionEnabled`). Must be disabled before deletion.

## Limits and Quotas

| Resource | Default Limit |
|----------|--------------|
| Clusters per account | 50 |
| Node pools per cluster | 100 |
| Nodes per node pool | 500 |
| Pods per node | Varies by network mode |
