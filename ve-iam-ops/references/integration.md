# Integration

## Environment Setup

**Primary path:** `ve` CLI (static Go binary, no runtime dependencies)

**Fallback path:** JIT Go SDK (dynamic script generation + `go run`)

### Prerequisites

1. **Volcengine CLI installed:**
   ```bash
   ve version
   # Expected: ve version v1.0.x
   ```

2. **Credentials configured:**
   ```bash
   export VOLCENGINE_ACCESS_KEY="<your-access-key>"
   export VOLCENGINE_SECRET_KEY="<masked>"
   export VOLCENGINE_REGION="cn-beijing"
   ```

3. **Verify IAM access:**
   ```bash
   ve iam ListUsers --Region $VOLCENGINE_REGION
   ```

### Go Runtime Bootstrap

If Agent Runtime lacks Go, JIT download from official source:

```bash
# Check Go runtime
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/{{env.GO_VERSION}}.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
fi

go version
```

> **Go version strategy:**
> - **JIT download:** Go 1.21+ (stable)
> - **Script compatibility:** Go 1.14+ (minimum)

### JIT Go SDK Workflow

1. **Initialize workspace:**
   ```bash
   mkdir -p /tmp/ve-sdk-workspace
   cd /tmp/ve-sdk-workspace
   go mod init ve-sdk-script
   ```

2. **Get dependencies:**
   ```bash
   # Set proxy for China CDN mirror (faster download)
   export GOPROXY="https://goproxy.cn,direct"
   
   # Volcengine SDK
   go get -u github.com/volcengine/volc-sdk-golang
   ```

3. **Generate script** (Agent dynamically creates operation-specific .go file)

4. **Execute:**
   ```bash
   go run ./main.go
   ```

### SDK Package Structure

| Service | Go SDK Package |
|---------|---------------|
| IAM | `github.com/volcengine/volc-sdk-golang/service/iam` |
| STS | `github.com/volcengine/volc-sdk-golang/service/sts` |

> Find package names at: https://github.com/volcengine/volc-sdk-golang/tree/main/service

## IAM SDK Script Template

```go
// main.go (single-file script template)
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/iam"
)

func main() {
    // Initialize IAM service instance
    instance := iam.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    instance.Client.SetRegion(os.Getenv("VOLCENGINE_REGION"))
    
    // Prepare request parameters
    params := make(map[string]interface{})
    params["Region"] = os.Getenv("VOLCENGINE_REGION")
    params["UserName"] = os.Getenv("USER_NAME")
    
    // Make API call
    resp, err := instance.Client.Request("iam", "GetUser", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "API call failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}
```

> Use `os.Getenv("KEY")` for all credentials. Never hardcode secrets in scripts.

## Cross-Account Role Assumption

### Setup Trust Relationship

```bash
# Create role in target account for source account to assume
ve iam CreateRole \
  --RoleName CrossAccountRole \
  --AssumeRolePolicyDocument '{
    "Statement": [{
      "Effect": "Allow",
      "Principal": {"STS": ["trn:sts::123456789012:root"]},
      "Action": "sts:AssumeRole"
    }]
  }' \
  --Region cn-beijing

# Attach policy to role
ve iam AttachRolePolicy \
  --RoleName CrossAccountRole \
  --PolicyName ReadOnlyPolicy \
  --Region cn-beijing
```

### Assume Role from Source Account

```bash
# Assume the role
ASSUME_OUTPUT=$(ve sts AssumeRole \
  --RoleTrn "trn:iam::999999999999:role/CrossAccountRole" \
  --RoleSessionName "cross-account-session" \
  --Region cn-beijing)

# Extract temporary credentials
export TEMP_ACCESS_KEY=$(echo $ASSUME_OUTPUT | jq -r '.Result.Credentials.AccessKeyId')
export TEMP_SECRET_KEY=$(echo $ASSUME_OUTPUT | jq -r '.Result.Credentials.SecretKey')
export TEMP_SESSION_TOKEN=$(echo $ASSUME_OUTPUT | jq -r '.Result.Credentials.SessionToken')

# Use temporary credentials (masked in logs)
echo "Using temporary credentials: $TEMP_ACCESS_KEY"
```

## Service-Linked Roles

Some Volcengine services automatically create service-linked roles:

```bash
# List service-linked roles
ve iam ListRoles --Region cn-beijing | jq '.Result.Roles[] | select(.Path == "/service-role/")'

# Common service-linked roles:
# - /service-role/ECSAutoScalingRole
# - /service-role/RDSBackupRole
# - /service-role/TOSServiceRole
```

## Terraform Integration

```hcl
# IAM user
resource "volcengine_iam_user" "example" {
  user_name = "example-user"
}

# IAM policy
resource "volcengine_iam_policy" "example" {
  policy_name = "example-policy"
  policy_document = jsonencode({
    Statement = [{
      Effect   = "Allow"
      Action   = ["ecs:DescribeInstances"]
      Resource = "*"
    }]
  })
}

# Attach policy to user
resource "volcengine_iam_user_policy_attachment" "example" {
  user_name  = volcengine_iam_user.example.user_name
  policy_name = volcengine_iam_policy.example.policy_name
}
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy IAM Resources

on:
  push:
    paths:
      - 'iam-policies/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install Volcengine CLI
        run: |
          curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/{{env.ve_version}}/ve-linux-amd64 -o /usr/local/bin/ve
          chmod +x /usr/local/bin/ve
      
      - name: Configure credentials
        run: |
          export VOLCENGINE_ACCESS_KEY="${{ secrets.VOLCENGINE_ACCESS_KEY }}"
          export VOLCENGINE_SECRET_KEY="${{ secrets.VOLCENGINE_SECRET_KEY }}"
          export VOLCENGINE_REGION="cn-beijing"
      
      - name: Deploy policies
        run: |
          for policy in iam-policies/*.json; do
            name=$(basename $policy .json)
            ve iam CreatePolicy \
              --PolicyName $name \
              --PolicyDocument "$(cat $policy)" \
              --Region $VOLCENGINE_REGION || \
            ve iam UpdatePolicy \
              --PolicyName $name \
              --PolicyDocument "$(cat $policy)" \
              --Region $VOLCENGINE_REGION
          done
```

## Policy Validation

```bash
# Validate policy syntax using jq
validate_policy() {
  local file=$1
  
  # Check valid JSON
  if ! jq '.' "$file" > /dev/null 2>&1; then
    echo "Error: Invalid JSON"
    return 1
  fi
  
  # Check required fields
  if ! jq -e '.Statement' "$file" > /dev/null 2>&1; then
    echo "Error: Missing Statement"
    return 1
  fi
  
  # Check each statement has required fields
  local statements=$(jq '.Statement | length' "$file")
  for ((i=0; i<$statements; i++)); do
    if ! jq -e ".Statement[$i].Effect" "$file" > /dev/null 2>&1; then
      echo "Error: Statement $i missing Effect"
      return 1
    fi
    if ! jq -e ".Statement[$i].Action" "$file" > /dev/null 2>&1; then
      echo "Error: Statement $i missing Action"
      return 1
    fi
    if ! jq -e ".Statement[$i].Resource" "$file" > /dev/null 2>&1; then
      echo "Error: Statement $i missing Resource"
      return 1
    fi
  done
  
  echo "Policy is valid"
  return 0
}

# Usage
validate_policy /tmp/policy.json
```

## Credential Rotation

```bash
#!/bin/bash
# rotate-credentials.sh

USER=$1
REGION=${VOLCENGINE_REGION:-cn-beijing}

if [ -z "$USER" ]; then
  echo "Usage: $0 <username>"
  exit 1
fi

echo "Rotating credentials for user: $USER"

# 1. Create new access key
echo "Creating new access key..."
NEW_KEY=$(ve iam CreateAccessKey --UserName $USER --Region $REGION)
NEW_AK=$(echo $NEW_KEY | jq -r '.Result.AccessKey.AccessKeyId')
NEW_SK=$(echo $NEW_KEY | jq -r '.Result.AccessKey.SecretKey')

echo "New access key created: $NEW_AK"
echo "Secret key (SAVE THIS): <masked>"

# 2. List old keys
echo "Listing old access keys..."
OLD_KEYS=$(ve iam ListAccessKeys --UserName $USER --Region $REGION | jq -r '.Result.AccessKeyMetadata[].AccessKeyId')

# 3. Wait for confirmation to deactivate old keys
read -p "Deactivate old keys? (y/n) " confirm
if [ "$confirm" = "y" ]; then
  for key in $OLD_KEYS; do
    if [ "$key" != "$NEW_AK" ]; then
      echo "Deactivating: $key"
      ve iam UpdateAccessKey --UserName $USER --AccessKeyId $key --Status Inactive --Region $REGION
    fi
  done
fi

echo "Rotation complete. Update your applications with the new credentials."
```
