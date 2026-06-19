# NAS Core Concepts

## Key Concepts

| Concept | Description |
|---------|-------------|
| **File System** | A managed NFS/SMB share with specified capacity and performance tier |
| **Mount Target** | A network endpoint (IP address) in a VPC subnet for accessing the file system |
| **Permission Group** | Access control rules that define which CIDRs can access a mount target |
| **Snapshot** | Point-in-time backup of a file system |

## Common Protocols

- **NFS** — Linux/Unix clients (vers=4.0 recommended)
- **SMB** — Windows clients (vers=3.0 recommended)
