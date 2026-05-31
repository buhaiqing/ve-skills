# Security Group Core Concepts

## Architecture

Security Groups act as virtual firewalls for ECS instances, controlling inbound and outbound traffic at the instance level.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Security Group** | A set of firewall rules for ECS instances |
| **Inbound Rule** | Controls traffic entering the instance |
| **Outbound Rule** | Controls traffic leaving the instance |
| **Enterprise SG** | Supports priority-based rules (1-100) |
| **Basic SG** | Simple allow/deny without priority |
| **SG-to-SG Reference** | Allow traffic from instances in another SG |

## Rule Components

| Component | Description | Example |
|-----------|-------------|---------|
| Protocol | `tcp`, `udp`, `icmp`, `all` | `tcp` |
| Port Range | Single port or range | `22/22`, `80/443` |
| Source/Dest | CIDR or security group ID | `10.0.0.0/8` |
| Policy | `accept` or `drop` | `accept` |
| Priority | 1-100 (enterprise SG only) | `1` |

## Security Best Practices

- Never use `0.0.0.0/0` for sensitive ports (22, 3389, 3306, 6379)
- Use SG-to-SG references instead of CIDR when possible
- Prefer enterprise SGs for granular control
- Default SG should have minimal rules
