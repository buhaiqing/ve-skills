# SecurityOps — Security Group Security Operations Framework

> SecurityOps deep content per TE-7. SKILL.md `references/*.md` link here.
> **Purpose**: Network security baseline, rule audit, incident response, compliance checks.

## Security Baseline Checklist

```markdown
## Security Group Baseline — [Date]

### Rule Governance
- [ ] No 0.0.0.0/0 (IPv4) or ::/0 (IPv6) on management ports (22/SSH, 3389/RDP, 22/SFTP)
- [ ] Inbound rules scoped to specific source CIDR blocks, not wide ranges
- [ ] Outbound rules restricted by default (no 0.0.0.0/0 on all ports)
- [ ] Default deny-all inbound rule present as lowest priority
- [ ] Rules documented with description field (service name, purpose, owner)

### Least Privilege Network
- [ ] Security groups follow application tier model (web → app → db)
- [ ] Cross-tier communication limited to required ports only
- [ ] No security group assigned to unrelated instances (clean mapping)
- [ ] Ephemeral port ranges scoped where possible (not 1024-65535)

### Hygiene & Lifecycle
- [ ] Unused security groups identified and removed (no attached resources)
- [ ] Orphaned rules (references to deleted SGs/instances) cleaned up
- [ ] Rule change history tracked with descriptions and dates

### Audit & Monitoring
- [ ] Security group rule changes trigger alarm rules via ve-cms-ops
- [ ] New public ingress rules require documented approval
- [ ] Quarterly full rule audit performed and documented
- [ ] Drift detection enabled (compare actual vs baseline rules)
```

## Security Incident Response

### Triage Workflow

```
[Security Group Alert]
    │
    ├── Is it a rule change?
    │   ├── NewIngressRuleAdded (0.0.0.0/0) → Check authorization
    │   │   ├── Authorized change (ticket documented) → Verify expiry, approve
    │   │   └── Unauthorized → Revoke rule immediately, audit IAM access
    │   ├── RuleRemoved → Verify service impact
    │   │   └── If impact → Restore rule, check change management process
    │   └── CrossAccountRuleAdded → Verify trust relationship
    │       └── Delegate to ve-iam-ops for cross-account audit
    │
    ├── Is it a traffic anomaly?
    │   ├── PortScanDetected (unusual connection attempts) → Review denied logs
    │   │   ├── From external IP → Update SG rules, block CIDR
    │   │   └── From internal IP → Security incident, escalate
    │   ├── UnexpectedOutboundTraffic → C2 beacon check
    │   │   └── Review outbound rules, restrict to known endpoints
    │   └── VolumeSpikeOnCertainPort → Verify service behavior
    │       ├── Legitimate traffic → Tune alarm threshold
    │       └── DDoS attack → Enable VPC ACL, delegate to ve-vpc-ops
    │
    ├── Is it a hygiene issue?
    │   ├── OrphanedSGDetected → Check for attached instances
    │   │   └── No attachments → Delete SG (requires confirmation)
    │   ├── RedundantRulesFound → Merge overlapping CIDR blocks
    │   └── OverlyBroadPortRange → Narrow to specific ports
    │
    └── Unknown → Escalate to ve-cms-ops for correlation + security team
```

### Containment Steps (destructive ops — require confirmation)

| Tier | Action | Confirmation Required |
|------|--------|-----------------------|
| 🟢 Low | Revoke single inbound rule, update rule description | No |
| 🟡 Medium | Revoke multiple rules, replace with stricter rules | Yes — via {{user.*}} |
| 🔴 High | Delete security group (detach from all instances first) | Yes + secondary verification |
| 🚨 Critical | Replace all rules with deny-all (instance isolation) | Yes + documented approval |

## Vulnerability Scanning Patterns

### Common Vulnerability Classes

| Class | SG Specific Risk | Detection Method |
|-------|------------------|-----------------|
| Overly Permissive Ingress | 0.0.0.0/0 on SSH/RDP/MySQL/Redis | `ve vpc DescribeSecurityGroupAttributes` — rule audit |
| Unrestricted Egress | 0.0.0.0/0 on all outbound traffic | Review outbound rules per SG |
| Port Scanning Target | Wide port range (1-65535) open to internet | Check port range in ingress rules |
| Unused Security Group | SG exists with no instance attachment | `ve vpc DescribeSecurityGroups` — check via {{user.sg_id}} cross-ref with ve-ecs-ops |
| Cross-Tier Violation | App SG allows direct DB port access | Compare SG-to-SG reference rules |

### Automated Scanning

```bash
# Security group audit checklist — run quarterly
echo "=== Security Group Audit $(date +%Y-%m-%d) ==="

# 1. Find all SGs with 0.0.0.0/0 on management ports
# Uses DescribeSecurityGroupAttributes to inspect per-SG ingress rules
echo "→ Checking for overly permissive SGs..."
for sg_id in $(ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.SecurityGroups[].SecurityGroupId'); do
  echo "→ SG: $sg_id"
  ve vpc DescribeSecurityGroupAttributes --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId "$sg_id" |
    jq '.Result.IngressRules[] | select(.CidrIp == "0.0.0.0/0") | {PortRange, Policy, Priority}'
done
echo "→ Review and restrict these rules"

# 2. Check for SGs with no instances attached (orphaned)
# DescribeSecurityGroups does not return instance count; cross-ref via ve-ecs-ops
echo "→ Finding orphaned SGs..."
ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}} |
  jq '.Result.SecurityGroups[] | {SecurityGroupId, SecurityGroupName, VpcId}'
echo "→ Cross-reference with ve-ecs-ops to verify each SG has attached instances"

# 3. Audit all SGs for wide / unrestricted port ranges
echo "→ Checking for wide port ranges..."
for sg_id in $(ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.SecurityGroups[].SecurityGroupId'); do
  echo "→ SG: $sg_id"
  ve vpc DescribeSecurityGroupAttributes --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId "$sg_id" |
    jq '.Result.IngressRules[] | {CidrIp, PortRange, Action, Priority}'
done
echo "→ Review: flag any rule with PortRange covering unusual ranges (e.g. 1-65535)"

# 4. Verify outbound rules
echo "→ Checking outbound restrictions..."
for sg_id in $(ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}} | jq -r '.Result.SecurityGroups[].SecurityGroupId'); do
  ve vpc DescribeSecurityGroupAttributes --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId "$sg_id" |
    jq '.Result.EgressRules[] | select(.CidrIp == "0.0.0.0/0") | {PortRange, Action, Priority}'
done
echo "→ Restrict outbound where possible"

# 5. Cross-check SG references for tier isolation
echo "→ Delegate to ve-vpc-ops for network tier audit"
```

## Compliance Mapping

| Control Framework | SG Mapping | Verification |
|-------------------|------------|--------------|
| **SOC2** — Network Security | SG rules restrict access | ve-security-group-ops rule audit |
| **PCI-DSS** — Network Segmentation | Tier-based SG model | Cross-tier rule verification |
| **ISO 27001** — A.13 Communications Security | Network access control | Quarterly rule audit |
| **NIST 800-53** — SC-7 Boundary Protection | SG as network boundary | Ingress/egress rule review |
| **CIS** — Cloud Foundation | No 0.0.0.0/0 on management ports | Automated SG scan |

## Cross-Skill Security Routing

| Security Symptom | Delegate To | Action |
|-----------------|-------------|--------|
| IAM permission denied on SG modification | ve-iam-ops | Policy audit + attach/detach |
| Instance public IP exposure | ve-eip-ops | EIP release or associate to bastion |
| VPC ACL conflicting with SG rules | ve-vpc-ops | ACL rule audit + alignment |
| Abnormal traffic pattern across SGs | ve-cms-ops | Alarm correlation + anomaly detection |
| Instance compromised behind SG | ve-ecs-ops | Instance isolation + forensic snapshot |
| KMS VPC endpoint SG misconfigured | ve-kms-ops | Key access endpoint verification |