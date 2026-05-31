## OpenAPI

- Spec: https://www.volcengine.com/docs/6409
- Service ID: `vpc` (Security Group APIs are part of VPC service)

## SDK Operations Map

| Goal | API operationId | SDK Method |
|------|-----------------|------------|
| Create security group | CreateSecurityGroup | `CreateSecurityGroup` |
| Describe security groups | DescribeSecurityGroups | `DescribeSecurityGroups` |
| Delete security group | DeleteSecurityGroup | `DeleteSecurityGroup` |
| Describe SG attributes | DescribeSecurityGroupAttributes | `DescribeSecurityGroupAttributes` |
| Authorize ingress | AuthorizeSecurityGroupIngress | `AuthorizeSecurityGroupIngress` |
| Authorize egress | AuthorizeSecurityGroupEgress | `AuthorizeSecurityGroupEgress` |
| Revoke ingress | RevokeSecurityGroupIngress | `RevokeSecurityGroupIngress` |
| Revoke egress | RevokeSecurityGroupEgress | `RevokeSecurityGroupEgress` |

## Request / Response Notes

- Required fields: `SecurityGroupName`, `VpcId` for CreateSecurityGroup
- Ingress/egress: `IpProtocol` (tcp/udp/icmp/all), `PortRange` (e.g. `22/22`), `CidrIp` or `SourceSecurityGroupId`
- Enterprise SG: `SecurityGroupType` = `enterprise` supports priority-based rules (1-100)
- Default SG cannot be deleted
