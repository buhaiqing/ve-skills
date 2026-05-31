# Monitoring IAM

## Key Metrics and Indicators

### Security Metrics

| Metric | Source | Alert Threshold | Description |
|--------|--------|-----------------|-------------|
| `iam:AccessKeyAge` | IAM API | > 90 days | Access key age in days |
| `iam:UnusedAccessKeys` | IAM API | > 30 days unused | Inactive access keys |
| `iam:ConsoleLoginFailures` | Audit logs | > 5 per hour | Failed console logins |
| `iam:MFAUsage` | Audit logs | < 100% for privileged | MFA adoption rate |
| `iam:PolicyAttachmentCount` | IAM API | > 8 per identity | Approaching limit |
| `iam:RoleAssumptionRate` | STS logs | Spike detection | Unusual assumption patterns |
| `iam:RootAccountUsage` | Audit logs | Any usage | Root account activity |

### Operational Metrics

| Metric | Source | Purpose |
|--------|--------|---------|
| `iam:UserCount` | IAM API | Track user growth |
| `iam:RoleCount` | IAM API | Track role growth |
| `iam:PolicyCount` | IAM API | Track policy growth |
| `iam:GroupCount` | IAM API | Track group growth |
| `iam:CredentialReportAge` | IAM API | Report freshness |

## Audit Log Analysis

### Access Key Usage Detection

```bash
# Generate and analyze credential report
ve iam GenerateCredentialReport --Region $VOLCENGINE_REGION
REPORT=$(ve iam GetCredentialReport --Region $VOLCENGINE_REGION | jq -r '.Result.Content' | base64 -d)

# Find unused access keys (> 90 days)
echo "$REPORT" | awk -F',' 'NR>1 {
  if ($10 == "false" && $11 > 90) {
    print "Unused key: " $3 " (age: " $11 " days)"
  }
}'

# Find keys never used
echo "$REPORT" | awk -F',' 'NR>1 {
  if ($10 == "false" && $11 == "not_supported") {
    print "Never used: " $3
  }
}'
```

### Console Login Monitoring

```bash
# Check console login profile status
ve iam GetLoginProfile --UserName $USER --Region $VOLCENGINE_REGION

# List users with console access
ve iam ListUsers --Region $VOLCENGINE_REGION | jq -r '.Result.Users[].UserName' | while read user; do
  ve iam GetLoginProfile --UserName $user --Region $VOLCENGINE_REGION 2>/dev/null && echo "$user: console enabled"
done
```

### Policy Usage Audit

```bash
# List all custom policies and their attachments
ve iam ListPolicies --Scope Local --Region $VOLCENGINE_REGION | jq -r '.Result.Policies[].PolicyName' | while read policy; do
  ATTACHED=$(ve iam ListEntitiesForPolicy --PolicyName $policy --Region $VOLCENGINE_REGION | jq '.Result.PolicyGroups | length + .Result.PolicyUsers | length + .Result.PolicyRoles | length')
  echo "$policy: $ATTACHED attachments"
done

# Find unattached policies (candidates for cleanup)
ve iam ListPolicies --Scope Local --Region $VOLCENGINE_REGION | jq -r '.Result.Policies[] | select(.AttachmentCount == 0) | .PolicyName'
```

## Health Checks

### Daily Security Checklist

```bash
#!/bin/bash
# iam-daily-check.sh

REGION="${VOLCENGINE_REGION:-cn-beijing}"

echo "=== IAM Daily Security Check ==="
echo "Date: $(date)"
echo ""

# 1. Check for users without MFA
echo "--- Users without MFA ---"
ve iam ListUsers --Region $REGION | jq -r '.Result.Users[].UserName' | while read user; do
  # Check if user has MFA (via credential report simulation)
  ve iam GetUser --UserName $user --Region $REGION | jq -r '[.Result.User.UserName, "Check MFA in console"] | @tsv'
done
echo ""

# 2. Check for old access keys (> 90 days)
echo "--- Access keys older than 90 days ---"
ve iam GenerateCredentialReport --Region $REGION 2>/dev/null
REPORT=$(ve iam GetCredentialReport --Region $REGION 2>/dev/null | jq -r '.Result.Content' | base64 -d)
echo "$REPORT" | awk -F',' 'NR>1 && $9 > 90 {print $3 ": " $9 " days old"}'
echo ""

# 3. Check for inactive keys (> 30 days unused)
echo "--- Inactive access keys ---"
echo "$REPORT" | awk -F',' 'NR>1 && $10 == "false" && $11 > 30 {print $3 ": unused for " $11 " days"}'
echo ""

# 4. Check policy limits (approaching 10)
echo "--- Identities with many policies ---"
ve iam ListUsers --Region $REGION | jq -r '.Result.Users[].UserName' | while read user; do
  COUNT=$(ve iam ListAttachedUserPolicies --UserName $user --Region $REGION | jq '.Result.AttachedPolicies | length')
  if [ "$COUNT" -gt 5 ]; then
    echo "$user: $COUNT policies attached"
  fi
done
echo ""

# 5. Check for unattached policies
echo "--- Unattached policies (cleanup candidates) ---"
ve iam ListPolicies --Scope Local --Region $REGION | jq -r '.Result.Policies[] | select(.AttachmentCount == 0) | .PolicyName'
echo ""

echo "=== Check Complete ==="
```

## Alerting Rules

### Critical Alerts

| Condition | Severity | Response |
|-----------|----------|----------|
| Root account login | CRITICAL | Immediate investigation |
| New access key created | HIGH | Verify authorized creation |
| Policy with wildcard (*) permissions | HIGH | Review policy scope |
| MFA device deactivated | HIGH | Verify authorized change |
| Role trust policy modified | HIGH | Review trust changes |

### Warning Alerts

| Condition | Severity | Response |
|-----------|----------|----------|
| Access key age > 90 days | WARNING | Schedule rotation |
| User inactive > 90 days | WARNING | Review account necessity |
| Policy attachments > 8 | WARNING | Plan policy consolidation |
| Unused policy > 180 days | WARNING | Consider deletion |

## Compliance Checks

### CIS Benchmark Checks

```bash
# Check 1.1: Avoid the use of the root account
# Check 1.2: Ensure MFA is enabled for the root account
# Check 1.3: Ensure credentials unused for 90 days or greater are disabled
# Check 1.4: Ensure access keys are rotated every 90 days
# Check 1.5: Ensure IAM password policy requires at least one uppercase letter
# Check 1.6: Ensure IAM password policy requires at least one lowercase letter
# Check 1.7: Ensure IAM password policy requires at least one symbol
# Check 1.8: Ensure IAM password policy requires at least one number
# Check 1.9: Ensure IAM password policy requires minimum length of 14
# Check 1.10: Ensure IAM password policy prevents password reuse

# Run credential report for most checks
ve iam GenerateCredentialReport --Region $VOLCENGINE_REGION
```

## Reporting

### Monthly IAM Report

```bash
#!/bin/bash
# iam-monthly-report.sh

REGION="${VOLCENGINE_REGION:-cn-beijing}"

echo "# IAM Monthly Report"
echo "Generated: $(date)"
echo "Region: $REGION"
echo ""

# User summary
echo "## User Summary"
USER_COUNT=$(ve iam ListUsers --Region $REGION | jq '.Result.Users | length')
echo "- Total users: $USER_COUNT"
echo ""

# Role summary
echo "## Role Summary"
ROLE_COUNT=$(ve iam ListRoles --Region $REGION | jq '.Result.Roles | length')
echo "- Total roles: $ROLE_COUNT"
echo ""

# Policy summary
echo "## Policy Summary"
POLICY_COUNT=$(ve iam ListPolicies --Scope Local --Region $REGION | jq '.Result.Policies | length')
echo "- Custom policies: $POLICY_COUNT"
echo ""

# Group summary
echo "## Group Summary"
GROUP_COUNT=$(ve iam ListGroups --Region $REGION | jq '.Result.Groups | length')
echo "- Total groups: $GROUP_COUNT"
echo ""

# Security findings
echo "## Security Findings"
echo "- Access keys requiring rotation: [Check credential report]"
echo "- Users without MFA: [Check console]"
echo "- Unattached policies: $(ve iam ListPolicies --Scope Local --Region $REGION | jq '[.Result.Policies[] | select(.AttachmentCount == 0)] | length')"
echo ""
```
