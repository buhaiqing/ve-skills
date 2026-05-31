# Security Group Knowledge Base

## Fault Pattern Library

### Pattern: Cannot Connect to ECS Instance

**Symptoms:** SSH/HTTP connection to ECS instance times out

**Root Causes:**
1. Security group inbound rule missing or incorrect
2. Rule with higher priority blocking traffic
3. Wrong port range or CIDR

**Resolution Steps:**
1. Check SG inbound rules: `ve vpc DescribeSecurityGroupAttributes`
2. Verify rule priority (enterprise SG)
3. Check if instance has multiple SGs (rules are additive)
4. Temporarily allow your IP to test connectivity

### Pattern: Unintended Port Exposure

**Symptoms:** Security scan reveals open ports to internet

**Root Causes:**
1. Overly permissive `0.0.0.0/0` CIDR
2. Wide port range (1/65535)
3. Missing or insufficient rules

**Resolution Steps:**
1. Audit all inbound rules with `0.0.0.0/0`
2. Replace with specific CIDR or SG references
3. Use enterprise SGs with strict priorities
