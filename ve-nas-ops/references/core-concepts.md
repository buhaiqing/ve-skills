# NAS Core Concepts

## Key Concepts

| Concept | Description |
|---------|-------------|
| **File System** | ℹ️ A managed NFS/SMB share with specified capacity and performance tier |
| **Mount Target** | ℹ️ A network endpoint (IP address) in a VPC subnet → accessible file system |
| **Permission Group** | ℹ️ Access control rules defining which CIDRs can access a mount target |
| **Snapshot** | ℹ️ Point-in-time backup of a file system |

## Common Protocols

- **NFS** — ℹ️ Linux/Unix clients (vers=4.0 recommended)
- **SMB** — ℹ️ Windows clients (vers=3.0 recommended)
