## Execution Flows (Agent-Readable)

Every operation: **Pre-flight → Execute → Validate → Recover**. Do not skip phases.

**Preference hint:** CLI is preferred for coverage and simplicity; Go SDK is used for operations CLI does not expose.

---

### Operation: CreateLoadBalancer — Create ALB Instance

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| VPC exists | `ve vpc DescribeVpcs` | VPC found | HALT; create VPC first via `ve-vpc-ops` |
| Subnet exists | `ve vpc DescribeSubnets` | Subnet found | HALT; create subnet first |
| Credentials | `test -n "$VOLCENGINE_ACCESS_KEY" && test -n "$VOLCENGINE_SECRET_KEY"` | Both set | HALT; configure credentials |
| Region | Verify `{{user.region}}` is valid ALB region | Region supported | Suggest valid region |
| Quota | `ve alb DescribeLoadBalancers` | Under quota | HALT; request quota increase |
| Name unique | ALB name not in use | No conflict | Use different name |

#### Execution — CLI (`ve`) (Primary Path)

```bash
# Create a private ALB (internal)
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "private"

# Create a public ALB (internet-facing)
ve alb CreateLoadBalancer \
  --Region "{{user.region}}" \
  --VpcId "{{user.vpc_id}}" \
  --SubnetId "{{user.subnet_id}}" \
  --LoadBalancerName "{{user.alb_name}}" \
  --Type "public" \
  --EipBillingConfig '{"Bandwidth":10,"EipBillingType":"PayByTraffic","ISP":"BGP"}'
```

> **Type Values:** `public` (internet-facing with EIP) or `private` (internal VPC only)

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/alb"
)

func main() {
    instance := alb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    resp, err := instance.CreateLoadBalancer(&alb.CreateLoadBalancerInput{
        VpcId:            "{{user.vpc_id}}",
        SubnetId:         "{{user.subnet_id}}",
        LoadBalancerName: "{{user.alb_name}}",
        Type:             "private",
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "CreateLoadBalancer failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(resp.Result.LoadBalancerId)
}
```

#### Post-execution Validation

```bash
# Parse ALB ID
ALB_ID=$(echo "$RESPONSE" | jq -r '.Result.LoadBalancerId')

# Poll until active
for i in $(seq 1 60); do
  STATUS=$(ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"$ALB_ID\"]" | jq -r '.Result.LoadBalancers[0].Status')
  echo "Status: $STATUS (attempt $i/60)"
  [ "$STATUS" = "active" ] && break
  sleep 5
done

# Report details
echo "ALB created: $ALB_ID"
ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds "[\"$ALB_ID\"]" | jq '.Result.LoadBalancers[0]'
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidVpc.NotFound` | 0 | HALT; VPC does not exist — `[ERROR] VPC not found. Create VPC via ve-vpc-ops first.` |
| `InvalidSubnet.NotFound` | 0 | HALT; subnet does not exist — `[ERROR] Subnet not found. Create subnet first.` |
| `InvalidParameter` | 1 | Fix args from error message — `[ERROR] InvalidParameter: Request parameter is invalid.` |
| `QuotaExceeded.LoadBalancer` | 0 | HALT; request quota increase — `[ERROR] ALB quota exceeded.` |
| `InsufficientBalance` | 0 | HALT; recharge account — `[ERROR] Insufficient balance.` |
| `Throttling` / 429 | 3 | 1s, 2s, 4s backoff; retry — `WARNING Rate limit reached.` |
| `InternalError` | 3 | 2s, 4s, 8s backoff; retry, then HALT — `[ERROR] InternalError.` |

---

### Operation: DescribeLoadBalancers — List/Describe ALB Instances

#### Execution — CLI (`ve`)

```bash
# List all ALBs
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}"

# Filter by ID
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerIds '["{{user.alb_id}}"]'

# Filter by name
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerName "{{user.alb_name}}"

# Filter by VPC
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --VpcId "{{user.vpc_id}}"

# Filter by type
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" --Type "public"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| ID | `.Result.LoadBalancers[].LoadBalancerId` | Format `alb-xxxxxxxxx` |
| Name | `.Result.LoadBalancers[].LoadBalancerName` | User-defined |
| Type | `.Result.LoadBalancers[].Type` | `public` or `private` |
| Status | `.Result.LoadBalancers[].Status` | Human-readable state |
| Address | `.Result.LoadBalancers[].Address` | Internal VPC IP |
| EIP | `.Result.LoadBalancers[].EipAddress` | Public IP (public type) |
| VPC | `.Result.LoadBalancers[].VpcId` | Parent VPC |
| Created | `.Result.LoadBalancers[].CreateTime` | ISO 8601 |

---

### Operation: DeleteLoadBalancer — Delete ALB Instance

#### Pre-flight (Safety Gate)

1. **MUST** obtain explicit user confirmation: irreversible delete of `{{user.alb_name}}` (`{{user.alb_id}}`)
2. **MUST NOT** proceed without clear user assent
3. **WARN** about downstream impact: all listeners, rules, and server groups will be affected
4. **Confirmation pattern:** present a block and require the user to type the ALB name to confirm:

   ```
   ⚠️ DESTRUCTIVE ACTION — IRREVERSIBLE
   You are about to DELETE ALB:
   - Name: {{user.alb_name}}
   - ID:   {{user.alb_id}}
   - Region: {{user.region}}
   All listeners, rules, and server groups will be permanently removed.
   Type the ALB name "{{user.alb_name}}" to confirm: _
   ```
4. Verify no active listeners:

```bash
# Check for listeners before deletion
ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}"

# If listeners exist, warn user and delete them
for listener_id in $(ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[].ListenerId'); do
  echo "WARNING: Listener $listener_id will be deleted"
  ve alb DeleteListener --Region "{{user.region}}" --ListenerId "$listener_id"
done

# For public ALB, check EIP association
ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds '["{{user.alb_id}}"]' | jq '.Result.LoadBalancers[0].EipAddress'
```

#### Execution — CLI (`ve`)

```bash
ve alb DeleteLoadBalancer --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}"
```

#### Post-execution Validation

```bash
# Poll until ALB is deleted (NotFound)
for i in $(seq 1 60); do
  if ve alb DescribeLoadBalancers --Region "{{user.region}}" --LoadBalancerIds '["{{user.alb_id}}"]' 2>&1 | grep -q "NotFound"; then
    echo "ALB deleted successfully"
    break
  fi
  echo "Waiting for deletion... (attempt $i/60)"
  sleep 5
done
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidLoadBalancer.NotFound` | 0 | Already deleted; skip — `[INFO] ALB already deleted.` |
| `DependencyViolation.Listener` | 0 | HALT; delete listeners first — `[ERROR] Listeners still attached. Delete listeners first.` |
| `IncorrectStatus.LoadBalancer` | 3 | 10s wait for stable status — `[WARNING] ALB not in deletable state. Waiting...` |
| `Forbidden.RAM` | 0 | HALT; check IAM permissions — `[ERROR] IAM permission denied. Check Volcengine IAM policies.` |

---

### Operation: CreateListener — Create ALB Listener

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| ALB exists | `ve alb DescribeLoadBalancers` | Status `active` | HALT |
| Port available | `ve alb DescribeListeners` | Port not in use | Use different port |
| Certificate (HTTPS) | Verify cert exists | Certificate valid | HALT; create certificate first |

#### Execution — CLI (`ve`)

```bash
# Create HTTP listener
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTP" \
  --Port "{{user.listener_port}}" \
  --ListenerName "{{user.listener_name}}"

# Create HTTPS listener (with certificate)
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "{{user.listener_name}}" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2"

# Create HTTPS listener with server group association
ve alb CreateListener \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Protocol "HTTPS" \
  --Port "443" \
  --ListenerName "https-listener" \
  --CertificateId "{{user.certificate_id}}" \
  --TLSPolicy "tls-1-2" \
  --ServerGroupId "{{user.server_group_id}}"
```

**Protocol Values:** `HTTP`, `HTTPS`, `gRPC`

#### Execution — JIT Go SDK (Fallback)

```go
package main

import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/alb"
)

func main() {
    instance := alb.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    resp, err := instance.CreateListener(&alb.CreateListenerInput{
        LoadBalancerId: "{{user.alb_id}}",
        Protocol:       "HTTP",
        Port:           80,
        ListenerName:   "{{user.listener_name}}",
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "CreateListener failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(resp.Result.ListenerId)
}
```

#### Post-execution Validation

```bash
# Verify listener creation
echo "Listener created:"
ve alb DescribeListeners --Region "{{user.region}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidLoadBalancer.NotFound` | 0 | HALT; ALB does not exist — `[ERROR] ALB not found. Verify the ALB ID.` |
| `PortConflict.Listener` | 0 | HALT; port already in use — `[ERROR] Port conflict. Use a different port.` |
| `InvalidCertificate.NotFound` | 0 | HALT; cert does not exist — `[ERROR] Certificate not found. Create cert first.` |
| `QuotaExceeded.Listener` | 0 | HALT; listener quota exceeded — `[ERROR] Listener quota exceeded.` |

---

### Operation: CreateServerGroup — Create Server Group

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| ALB exists | `ve alb DescribeLoadBalancers` | Status `active` | HALT |
| Name unique | `ve alb DescribeServerGroups` | Name not in use | Use different name |

#### Execution — CLI (`ve`)

```bash
# Create server group with health check
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --Description "Production server group" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "HTTP",
    "Method": "GET",
    "Uri": "/health",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3,
    "HealthyHttpCode": "200"
  }'

# Create server group (TCP health check)
ve alb CreateServerGroup \
  --Region "{{user.region}}" \
  --ServerGroupName "{{user.server_group_name}}" \
  --ServerGroupType "instance" \
  --HealthCheckConfig '{
    "Enabled": true,
    "Protocol": "TCP",
    "Timeout": 3,
    "Interval": 5,
    "HealthyThreshold": 3,
    "UnhealthyThreshold": 3
  }'
```

**ServerGroupType Values:** `instance` (ECS instance), `ip` (IP address)

#### Post-execution Validation

```bash
SG_ID=$(echo "$RESPONSE" | jq -r '.Result.ServerGroupId')
echo "Server Group created: $SG_ID"

# Verify server group details
ve alb DescribeServerGroups --Region "{{user.region}}" --ServerGroupIds "[\"$SG_ID\"]" | jq '.Result.ServerGroups[0]'
```

---

### Operation: CreateRule — Create Routing Rule

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Listener exists | `ve alb DescribeListeners` | Listener found | HALT |
| Server group exists | `ve alb DescribeServerGroups` | Server group found | HALT; create server group first |

#### Execution — CLI (`ve`)

```bash
# Path-based routing rule
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "{{user.domain}}" \
  --Url "{{user.path_pattern}}" \
  --RuleName "api-route"

# Host-based routing rule only
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Domain "api.example.com" \
  --RuleName "host-rule"

# Path-based rule with multiple server groups (weighted)
ve alb CreateRule \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Url "/api/*" \
  --RuleName "api-weighted"
```

**Pattern Examples:** `/api/*`, `/images/*.jpg`, `/*`

#### Post-execution Validation

```bash
# Verify rule creation
ve alb DescribeRules --Region "{{user.region}}" --ListenerId "{{user.listener_id}}" | jq -r '.Result.Rules[] | "\(.RuleId): \(.Domain) \(.Url)"'
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidListener.NotFound` | 0 | HALT; listener does not exist — `[ERROR] Listener not found. Verify listener ID.` |
| `InvalidServerGroup.NotFound` | 0 | HALT; server group does not exist — `[ERROR] Server group not found. Create server group first.` |
| `InvalidRule.Pattern` | 0 | HALT; URL pattern invalid — `[ERROR] Invalid URL pattern. Use wildcard patterns like /api/*.` |
| `QuotaExceeded.Rule` | 0 | HALT; rule quota exceeded — `[ERROR] Rule quota exceeded.` |

---

### Operation: AddServersToGroup — Add Backend Servers to Server Group

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Server group exists | `ve alb DescribeServerGroups` | Server group found | HALT |
| ECS instances exist | `ve ecs DescribeInstances` | Instances found | HALT |

#### Execution — CLI (`ve`)

```bash
# Add ECS instances as backend servers
ve alb AddServersToGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --Servers '[{"ServerId":"i-xxx","Port":8080,"Weight":100,"ServerType":"ecs"},{"ServerId":"i-yyy","Port":8080,"Weight":100,"ServerType":"ecs"}]'
```

#### Post-execution Validation

```bash
ve alb DescribeServerGroups --Region "{{user.region}}" --ServerGroupIds "[\"{{user.server_group_id}}\"]" | jq -r '.Result.ServerGroups[0].Servers[] | "\(.ServerId):\(.Port) weight=\(.Weight) status=\(.Status)"'
```

---

### Operation: DescribeServerGroups — List Server Groups

#### Execution — CLI (`ve`)

```bash
# List all server groups
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}"

# Filter by ID
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds "[\"{{user.server_group_id}}\"]"

# Filter by name
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupName "{{user.server_group_name}}"
```

---

### Operation: DescribeRules — List Routing Rules

#### Execution — CLI (`ve`)

```bash
# List all rules for a listener
ve alb DescribeRules --Region "{{env.VOLCENGINE_REGION}}" --ListenerId "{{user.listener_id}}"
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Rule ID | `.Result.Rules[].RuleId` | Format `rule-xxxxxxxxx` |
| Domain | `.Result.Rules[].Domain` | Host header match |
| URL | `.Result.Rules[].Url` | Path pattern match |
| ServerGroupId | `.Result.Rules[].ServerGroupId` | Target server group |
| RuleName | `.Result.Rules[].RuleName` | User-defined name |

---

### Operation: ModifyLoadBalancerAttributes — Modify ALB Attributes

#### Execution — CLI (`ve`)

```bash
# Rename ALB
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --LoadBalancerName "{{user.alb_name}}"

# Modify description
ve alb ModifyLoadBalancerAttributes \
  --Region "{{user.region}}" \
  --LoadBalancerId "{{user.alb_id}}" \
  --Description "Updated description"
```

---

### Operation: DeleteListener — Delete Listener

#### Pre-flight (Safety Gate)

- **MUST** obtain confirmation before deleting a listener
- **MUST** check for attached rules and warn user
- **Confirmation pattern:** require the user to type the listener ID `{{user.listener_id}}` to confirm deletion

```bash
# Check for attached rules
ve alb DescribeRules --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

#### Execution

```bash
ve alb DeleteListener --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

---

### Operation: DeleteRule — Delete Routing Rule

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation: delete rule `{{user.rule_id}}`; require the user to type the rule ID to confirm

#### Execution

```bash
ve alb DeleteRule --Region "{{user.region}}" --RuleId "{{user.rule_id}}"
```

---

### Operation: ModifyServerGroupAttributes — Modify Server Group

#### Execution — CLI (`ve`)

```bash
# Rename server group
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerGroupName "new-name"

# Modify health check config
ve alb ModifyServerGroupAttributes \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --HealthCheckConfig '{"Enabled":true,"Uri":"/healthz","Interval":10}'
```

---

### Operation: RemoveServersFromGroup — Remove Backend Servers

#### Pre-flight (Safety Gate)

- **MUST** obtain confirmation before removing servers
- **MUST** warn about traffic interruption if this is the only server group member

#### Execution

```bash
ve alb RemoveServersFromGroup \
  --Region "{{user.region}}" \
  --ServerGroupId "{{user.server_group_id}}" \
  --ServerIds '[{"ServerId":"i-xxx","ServerType":"ecs"}]'
```

---

### Operation: DescribeListeners — List ALB Listeners

#### Execution — CLI (`ve`)

```bash
# List all listeners for an ALB
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}"

# Filter by listener ID
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" --ListenerIds '["{{user.listener_id}}"]'
```

#### Present to User

| Field | JSON Path | Notes |
|-------|-----------|-------|
| Listener ID | `.Result.Listeners[].ListenerId` | Format `lsnr-xxxxxxxxx` |
| Protocol | `.Result.Listeners[].Protocol` | HTTP, HTTPS, gRPC |
| Port | `.Result.Listeners[].Port` | Listening port |
| Status | `.Result.Listeners[].Status` | `active` or `pending` |
| CertificateId | `.Result.Listeners[].CertificateId` | HTTPS only |
| ServerGroupId | `.Result.Listeners[].ServerGroupId` | Associated server group |

---

### Operation: ModifyListenerAttributes — Modify Listener Configuration

#### Pre-flight Checks

| Check | Method | Expected | On Failure |
|-------|--------|----------|------------|
| Listener exists | `ve alb DescribeListeners` | Listener found | HALT |

#### Execution

```bash
# Rename listener
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --ListenerName "updated-listener-name"

# Update certificate for HTTPS listener
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --CertificateId "{{user.certificate_id}}"

# Set request timeout
ve alb ModifyListenerAttributes \
  --Region "{{user.region}}" \
  --ListenerId "{{user.listener_id}}" \
  --RequestTimeout 120
```

#### Validation

```bash
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" --ListenerIds '["{{user.listener_id}}"]' | jq '.Result.Listeners[0]'
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidListener.NotFound` | 0 | HALT; listener does not exist — `[ERROR] Listener not found. Verify listener ID.` |
| `InvalidCertificate.NotFound` | 0 | HALT; cert does not exist — `[ERROR] Certificate not found.` |

---

### Operation: DeleteServerGroup — Delete Server Group

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before deleting a server group
- **MUST** warn about traffic interruption if the server group is associated with active listeners
- **MUST** check if any listeners reference this server group
- **Confirmation pattern:** require the user to type the server group ID `{{user.server_group_id}}` to confirm deletion

```bash
# Check which listeners reference this server group
ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "{{user.alb_id}}" | jq -r '.Result.Listeners[] | select(.ServerGroupId == "{{user.server_group_id}}") | "Listener: \(.ListenerId) on \(.Protocol):\(.Port)"'
```

#### Execution

```bash
# ⚠️ Ensure no listeners reference this server group before deletion
ve alb DeleteServerGroup --Region "{{user.region}}" --ServerGroupId "{{user.server_group_id}}"
```

#### Post-execution Validation

```bash
# Verify server group is deleted
ve alb DescribeServerGroups --Region "{{env.VOLCENGINE_REGION}}" --ServerGroupIds '["{{user.server_group_id}}"]' 2>&1 | grep -q "InvalidServerGroup.NotFound" && echo "Server group deleted successfully" || echo "Server group still exists"
```

#### Failure Recovery

| Error pattern | Max retries | Recovery |
|--------------|-------------|---------|
| `InvalidServerGroup.NotFound` | 0 | Already deleted; skip — `[INFO] Server group already deleted.` |
| `DependencyViolation` | 0 | HALT; listeners still reference this group — `[ERROR] Listeners still reference this server group. Remove associations first.` |

---

### FinOps Operation: DescribeIdleLoadBalancers — Find Idle ALBs

> **UX caveat:** This heuristic flags ALBs with **no listeners** as idle. It does NOT account for server groups still attached, reserved capacity, or billing commitments. Before deletion, confirm the ALB truly has no traffic and no attached server groups — never delete solely on this signal.

#### Execution

```bash
# List all ALBs and check for idle ones (no listeners, no traffic)
ALL_ALBS=$(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | select(.Status == "active") | .LoadBalancerId')

for ALB_ID in $ALL_ALBS; do
  LISTENER_COUNT=$(ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq '.Result.Listeners | length')
  if [ "$LISTENER_COUNT" -eq 0 ]; then
    echo "IDLE ALB: $ALB_ID - no listeners configured"
  fi
done
```

---

### FinOps Operation: DescribeCostSummary — ALB Cost Estimation

#### Execution

```bash
# List ALBs with creation time and type for cost analysis
echo "=== ALB Cost Summary ==="
echo ""
echo "| ALB ID | Name | Type | Status | Created |"
echo "|--------|------|------|--------|---------|"
ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | "|\(.LoadBalancerId)|\(.LoadBalancerName)|\(.Type)|\(.Status)|\(.CreateTime)|"'
```

---

### FinOps Operation: CleanupOrphanedListeners — Remove Unused Listeners

#### Pre-flight (Safety Gate)

- **MUST** obtain explicit confirmation before deleting any listeners
- **MUST** list all affected listeners with their IDs and ports
- **MUST** allow user to exclude specific listeners

```bash
# Find listeners with no rules or server groups attached
echo "=== Orphaned Listeners ==="
ALL_ALBS=$(ve alb DescribeLoadBalancers --Region "{{env.VOLCENGINE_REGION}}" | jq -r '.Result.LoadBalancers[] | .LoadBalancerId')

for ALB_ID in $ALL_ALBS; do
  ve alb DescribeListeners --Region "{{env.VOLCENGINE_REGION}}" --LoadBalancerId "$ALB_ID" | jq -r '.Result.Listeners[] | "\(.ListenerId) \(.Protocol):\(.Port)"'
done
```

#### Execution

```bash
# Delete orphaned listener
ve alb DeleteListener --Region "{{user.region}}" --ListenerId "{{user.listener_id}}"
```

---

