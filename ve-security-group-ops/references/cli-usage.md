# CLI Usage — Security Group (`ve vpc`)

> Security Group operations are part of the VPC service in the `ve` CLI.
> Verified at: `ve vpc --help`

# Common JSON Paths:
# CreateSecurityGroup:            $.Result.SecurityGroupId
# DescribeSecurityGroups:         $.Result.SecurityGroups[].{SecurityGroupId,SecurityGroupName,VpcId}
# DescribeSecurityGroupAttributes: $.Result.IngressRules, $.Result.EgressRules

## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md) for CLI installation.

**Credentials:** `VOLCENGINE_ACCESS_KEY`, `VOLCENGINE_SECRET_KEY`, `VOLCENGINE_REGION` env vars.

## Command Map

| Operation | CLI Command | Notes |
|-----------|------------|-------|
| List SGs | `ve vpc DescribeSecurityGroups --Region {{env.VOLCENGINE_REGION}}` | Paginate with `--MaxResults` + `--NextToken` |
| Create SG | `ve vpc CreateSecurityGroup --Region {{env.VOLCENGINE_REGION}} --VpcId {{user.vpc_id}} --SecurityGroupName {{user.sg_name}} --Description "{{user.description}}"` | Returns `$.Result.SecurityGroupId` |
| Describe SG | `ve vpc DescribeSecurityGroupAttributes --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}}` | Returns ingress/egress rules |
| Delete SG | `ve vpc DeleteSecurityGroup --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}}` | Security group must be empty |
| Add ingress rule | `ve vpc AuthorizeSecurityGroupIngress --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}} --CidrIp {{user.cidr}} --PortRange {{user.port}} --IpProtocol {{user.protocol}}` | Protocol: `tcp`, `udp`, `icmp` |
| Add egress rule | `ve vpc AuthorizeSecurityGroupEgress --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}} --CidrIp {{user.cidr}} --PortRange {{user.port}} --IpProtocol {{user.protocol}}` | Same params as ingress |
| Remove ingress | `ve vpc RevokeSecurityGroupIngress --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}} --CidrIp {{user.cidr}} --PortRange {{user.port}} --IpProtocol {{user.protocol}}` | Must match exact rule |
| Remove egress | `ve vpc RevokeSecurityGroupEgress --Region {{env.VOLCENGINE_REGION}} --SecurityGroupId {{user.sg_id}} --CidrIp {{user.cidr}} --PortRange {{user.port}} --IpProtocol {{user.protocol}}` | Must match exact rule |

## JSON Output Paths

| Operation | Key Path | Type | Description |
|-----------|----------|------|-------------|
| CreateSecurityGroup | `$.Result.SecurityGroupId` | string | Created SG ID |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].SecurityGroupId` | string | SG ID |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].SecurityGroupName` | string | SG name |
| DescribeSecurityGroups | `$.Result.SecurityGroups[].VpcId` | string | VPC ID |
| DescribeSecurityGroupAttributes | `$.Result.IngressRules` | array | Inbound rules |
| DescribeSecurityGroupAttributes | `$.Result.EgressRules` | array | Outbound rules |

## Coverage

All Security Group operations are CLI-covered. No SDK-only gaps.