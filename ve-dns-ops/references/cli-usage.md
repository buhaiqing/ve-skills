## Install and Config

See [Execution Environment Setup](../../ve-skill-generator/references/execution-environment.md).

## Conventions

- Command prefix: `ve dns`
- Output is JSON by default

## CLI vs API Coverage

| Operation | Available via `ve`? | Notes |
|-----------|---------------------|-------|
| CreateDomain | Yes | Add a domain |
| ListDomains | Yes | List all domains |
| DeleteDomain | Yes | Remove a domain |
| CreateRecordSet | Yes | Add DNS record |
| DescribeRecordSets | Yes | List records |
| UpdateRecordSet | Yes | Modify record |
| DeleteRecordSet | Yes | Remove record |
| DescribeDNSStatistics | Yes | Query DNS metrics |

## Command Map

| Goal | Example |
|------|---------|
| List domains | `ve dns ListDomains --Region cn-beijing` |
| Add A record | `ve dns CreateRecordSet --ZoneName example.com --RecordType A --Value 1.2.3.4` |
| Add CNAME | `ve dns CreateRecordSet --ZoneName example.com --RecordType CNAME --Value target.example.com` |
