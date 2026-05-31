# NAS Core Concepts

## Architecture

NAS (Network Attached Storage) provides scalable, shared file storage accessible by multiple compute instances via NFS or SMB protocols.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **File System** | A managed NFS/SMB share with specified capacity and performance tier |
| **Mount Target** | A network endpoint (IP address) in a VPC subnet for accessing the file system |
| **Permission Group** | Access control rules that define which CIDRs can access a mount target |
| **Snapshot** | Point-in-time backup of a file system |

## Storage Types

| Type | IOPS (per TB) | Throughput (per TB) | Latency | Use Case |
|------|---------------|---------------------|---------|----------|
| Capacity | 50 | 5 MB/s | ~10ms | Archival, backup, infrequent access |
| Performance | 300 | 15 MB/s | ~1ms | General purpose, dev/test |
| Extreme | 1000 | 50 MB/s | <0.5ms | HPC, databases |

## Common Protocols

- **NFS** — Linux/Unix clients (vers=4.0 recommended)
- **SMB** — Windows clients (vers=3.0 recommended)
