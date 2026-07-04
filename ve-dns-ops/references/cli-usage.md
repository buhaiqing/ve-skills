# CLI Usage — Volcengine DNS (`ve dns`)

> **jq Paths (`.Result.*`):** `.Domains[]` (domain list), `.Records` (record count), `.Records[]` (record details)

## Install and Config

- Install `ve` CLI: see [Volcengine CLI](https://github.com/volcengine/volcengine-cli)
- **Critical Credentials:** The `ve` CLI reads from env vars `VOLCENGINE_ACCESS_KEY` / `VOLCENGINE_SECRET_KEY` OR `~/.volcengine/config.json`

## Conventions (Agent Execution)

- Output is **JSON by default**
- CLI invocation: `ve dns <Action> --ParameterName value`
- All parameters use `--PascalCase` naming
- Region is a required parameter: `--Region "cn-beijing"`
- Boolean flags use `--FlagName` (true) or `--NoFlagName` (false)

## CLI Coverage Map

| Operation | Available via `ve`? | CLI Command |
|-----------|---------------------|-------------|
| ListDomains | ✅ Yes | `ve dns ListDomains --Region ...` |
| DescribeDomain | ✅ Yes | `ve dns DescribeDomain --Region ... --DomainName ...` |
| CreateDomain | ✅ Yes | `ve dns CreateDomain --Region ... --DomainName ...` |
| DeleteDomain | ✅ Yes | `ve dns DeleteDomain --Region ... --DomainName ...` |
| ModifyDomain | ✅ Yes | `ve dns ModifyDomain --Region ... --DomainName ...` |
| AddRecord | ✅ Yes | `ve dns AddRecord --Region ... --DomainName ... --RR ... --Type ... --Value ...` |
| UpdateRecord | ✅ Yes | `ve dns UpdateRecord --Region ... --DomainName ... --RecordId ...` |
| DeleteRecord | ✅ Yes | `ve dns DeleteRecord --Region ... --DomainName ... --RecordId ...` |
| ListRecords | ✅ Yes | `ve dns ListRecords --Region ... --DomainName ...` |
| DescribeDomainStatistics | ✅ Yes | `ve dns DescribeDomainStatistics --Region ... --DomainName ...` |
| DescribeDNSResolution | ✅ Yes | `ve dns DescribeDNSResolution --Region ... --DomainName ...` |
| BatchImportRecords | ✅ Yes | `ve dns BatchImportRecords --Region ... --DomainName ...` |

## Common Command Examples

### List Domains
```bash
ve dns ListDomains --Region "cn-beijing"
```

### Create Domain
```bash
ve dns CreateDomain --Region "cn-beijing" --DomainName "example.com"
```

### Delete Domain
```bash
ve dns DeleteDomain --Region "cn-beijing" --DomainName "example.com"
```

### Add A Record
```bash
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "www" --Type "A" --Value "192.168.1.1" --TTL 600
```

### Add MX Record
```bash
ve dns AddRecord --Region "cn-beijing" --DomainName "example.com" --RR "@" --Type "MX" --Value "mail.example.com" --Priority 10 --TTL 600
```

### Update Record
```bash
ve dns UpdateRecord --Region "cn-beijing" --DomainName "example.com" --RecordId "r-xxxxx" --Value "10.0.0.1" --TTL 300
```

### Delete Record
```bash
ve dns DeleteRecord --Region "cn-beijing" --DomainName "example.com" --RecordId "r-xxxxx"
```

### List Records
```bash
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com"

# Filter by type
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" --Type "A"
```

### Get Domain Statistics
```bash
ve dns DescribeDomainStatistics --Region "cn-beijing" --DomainName "example.com"
```

## Response Parsing Examples

Extract specific fields from JSON output:

```bash
# Get domain list as table
ve dns ListDomains --Region "cn-beijing" | jq -r '.Domains[] | [.DomainId, .DomainName, .Status] | @tsv'

# Get record count for a domain
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records | length'

# Get specific record by ID
ve dns ListRecords --Region "cn-beijing" --DomainName "example.com" | jq '.Records[] | select(.RecordId=="r-xxxxx")'

# Check domain exists
ve dns ListDomains --Region "cn-beijing" | jq -e '.Domains[] | select(.DomainName=="example.com")' > /dev/null
```
