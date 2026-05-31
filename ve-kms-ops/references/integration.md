# KMS Integration

## Environment Setup

### Primary Path: ve CLI (Static Binary)

The `ve` CLI is a static Go binary with no runtime dependencies - the preferred execution path.

#### Installation

```bash
# Download from GitHub releases
# Latest: https://github.com/volcengine/volcengine-cli/releases

# macOS ARM64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-darwin-arm64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Linux x86_64
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify
ve version
```

#### Credential Configuration

**Environment Variables (Recommended for Agents):**

```bash
export VOLCENGINE_ACCESS_KEY="<your-access-key>"
export VOLCENGINE_SECRET_KEY="<masked>"  # Never log this value
export VOLCENGINE_REGION="cn-beijing"

# Verify (without exposing secret)
test -n "$VOLCENGINE_SECRET_KEY" && echo "✅ Credentials configured"
```

**Config File:**

```bash
mkdir -p ~/.volcengine
cat > ~/.volcengine/config.json << 'CONFIGEOF'
{
  "current": "default",
  "profiles": [
    {
      "name": "default",
      "mode": "AK",
      "access_key": "<your-access-key>",
      "secret_key": "<your-secret-key>",
      "region": "cn-beijing"
    }
  ]
}
CONFIGEOF
chmod 600 ~/.volcengine/config.json
```

### Fallback Path: JIT Go SDK

When `ve` CLI doesn't support a specific operation, use the JIT Go SDK approach.

#### Go Runtime Bootstrap

```bash
# Check if Go exists
if ! command -v go &> /dev/null; then
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    [ "$ARCH" = "x86_64" ] && ARCH="amd64"
    [ "$ARCH" = "aarch64" ] && ARCH="arm64"
    
    mkdir -p /tmp/go-runtime
    curl -fsSL "https://go.dev/dl/go1.21.0.${OS}-${ARCH}.tar.gz" | tar -xz -C /tmp/go-runtime
    
    export PATH="/tmp/go-runtime/go/bin:$PATH"
    export GOPATH="/tmp/go-workspace"
    export GOCACHE="/tmp/go-cache"
    export GOMODCACHE="/tmp/go-modcache"
    export GOPROXY="https://goproxy.cn,direct"
fi

go version
```

#### SDK Initialization

```bash
# Initialize workspace
mkdir -p /tmp/ve-kms-workspace
cd /tmp/ve-kms-workspace
go mod init ve-kms-script

# Get dependencies
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
```

## SDK Script Template

### Basic KMS Operations Template

```go
// main.go - KMS operations template
package main

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/base"
    "github.com/volcengine/volc-sdk-golang/service/kms"
)

func main() {
    // Initialize KMS client
    instance := kms.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Get operation from environment
    operation := os.Getenv("KMS_OPERATION")
    if operation == "" {
        operation = "DescribeKeys" // Default operation
    }
    
    // Prepare request parameters
    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }
    
    // Add operation-specific parameters
    switch operation {
    case "CreateKey":
        params["KeySpec"] = getEnvOrDefault("KEY_SPEC", "AES_256")
        params["KeyUsage"] = getEnvOrDefault("KEY_USAGE", "ENCRYPT_DECRYPT")
        if desc := os.Getenv("KEY_DESCRIPTION"); desc != "" {
            params["Description"] = desc
        }
        
    case "DescribeKey", "EnableKey", "DisableKey", "ScheduleKeyDeletion", "CancelKeyDeletion":
        params["KeyId"] = os.Getenv("KEY_ID")
        if operation == "ScheduleKeyDeletion" {
            params["PendingWindowInDays"] = getEnvOrDefaultInt("PENDING_WINDOW", 7)
        }
        
    case "Encrypt":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["Plaintext"] = os.Getenv("PLAINTEXT")
        if ctx := os.Getenv("ENCRYPTION_CONTEXT"); ctx != "" {
            var encContext map[string]string
            if err := json.Unmarshal([]byte(ctx), &encContext); err == nil {
                params["EncryptionContext"] = encContext
            }
        }
        
    case "Decrypt":
        params["CiphertextBlob"] = os.Getenv("CIPHERTEXT")
        if ctx := os.Getenv("ENCRYPTION_CONTEXT"); ctx != "" {
            var encContext map[string]string
            if err := json.Unmarshal([]byte(ctx), &encContext); err == nil {
                params["EncryptionContext"] = encContext
            }
        }
        
    case "GenerateDataKey", "GenerateDataKeyWithoutPlaintext":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["KeySpec"] = getEnvOrDefault("DATA_KEY_SPEC", "AES_256")
        if ctx := os.Getenv("ENCRYPTION_CONTEXT"); ctx != "" {
            var encContext map[string]string
            if err := json.Unmarshal([]byte(ctx), &encContext); err == nil {
                params["EncryptionContext"] = encContext
            }
        }
        
    case "DescribeKeyRotation", "UpdateKeyRotation":
        params["KeyId"] = os.Getenv("KEY_ID")
        if operation == "UpdateKeyRotation" {
            params["EnableAutomaticKeyRotation"] = getEnvOrDefault("ENABLE_ROTATION", "true") == "true"
        }
        
    case "PutKeyPolicy":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["PolicyName"] = getEnvOrDefault("POLICY_NAME", "default")
        params["Policy"] = os.Getenv("POLICY_DOCUMENT")
        
    case "GetKeyPolicy":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["PolicyName"] = getEnvOrDefault("POLICY_NAME", "default")
        
    case "CreateGrant":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["GranteePrincipal"] = os.Getenv("GRANTEE_PRINCIPAL")
        params["Operations"] = parseOperations(os.Getenv("GRANT_OPERATIONS"))
        if retiring := os.Getenv("RETIRING_PRINCIPAL"); retiring != "" {
            params["RetiringPrincipal"] = retiring
        }
        
    case "ListGrants":
        params["KeyId"] = os.Getenv("KEY_ID")
        
    case "RevokeGrant", "RetireGrant":
        params["KeyId"] = os.Getenv("KEY_ID")
        params["GrantId"] = os.Getenv("GRANT_ID")
    }
    
    // Make API call
    resp, err := instance.Client.Request("kms", operation, params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println(string(resp))
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
    // Simplified - in production use strconv.Atoi
    return defaultValue
}

func parseOperations(ops string) []string {
    if ops == "" {
        return []string{"Encrypt", "Decrypt"}
    }
    // Parse comma-separated operations
    return []string{"Encrypt", "Decrypt"} // Simplified
}
```

### Execute Script

```bash
# Export credentials
export VOLCENGINE_ACCESS_KEY="<your-access-key>"
export VOLCENGINE_SECRET_KEY="<your-secret-key>"
export VOLCENGINE_REGION="cn-beijing"

# Create key
export KMS_OPERATION="CreateKey"
export KEY_SPEC="AES_256"
export KEY_USAGE="ENCRYPT_DECRYPT"
export KEY_DESCRIPTION="Test key"
go run ./main.go

# Encrypt data
export KMS_OPERATION="Encrypt"
export KEY_ID="key-xxxxxxxxxxxx"
export PLAINTEXT="$(echo 'Hello, World!' | base64)"
go run ./main.go
```

## Integration Patterns

### Application Encryption Pattern

```go
// Application-level encryption using data keys
package main

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
    "os"
    
    "github.com/volcengine/volc-sdk-golang/service/kms"
)

type EncryptionService struct {
    kmsClient *kms.Instance
    keyId     string
}

func NewEncryptionService(keyId string) *EncryptionService {
    instance := kms.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    return &EncryptionService{
        kmsClient: instance,
        keyId:     keyId,
    }
}

// EncryptedData represents encrypted data with encrypted data key
type EncryptedData struct {
    Ciphertext     string `json:"ciphertext"`
    EncryptedKey   string `json:"encrypted_key"`
    IV             string `json:"iv"`
}

func (s *EncryptionService) Encrypt(plaintext []byte, context map[string]string) (*EncryptedData, error) {
    // Generate data key from KMS
    params := map[string]interface{}{
        "Region":    os.Getenv("VOLCENGINE_REGION"),
        "KeyId":     s.keyId,
        "KeySpec":   "AES_256",
        "EncryptionContext": context,
    }
    
    resp, err := s.kmsClient.Client.Request("kms", "GenerateDataKey", params)
    if err != nil {
        return nil, fmt.Errorf("failed to generate data key: %w", err)
    }
    
    // Parse response
    var result struct {
        Result struct {
            Plaintext      string `json:"Plaintext"`
            CiphertextBlob string `json:"CiphertextBlob"`
        } `json:"Result"`
    }
    if err := json.Unmarshal(resp, &result); err != nil {
        return nil, err
    }
    
    // Decode plaintext data key
    dataKey, err := base64.StdEncoding.DecodeString(result.Result.Plaintext)
    if err != nil {
        return nil, err
    }
    
    // Encrypt data locally with AES-GCM
    block, err := aes.NewCipher(dataKey)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    iv := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(iv, iv, plaintext, nil)
    
    // Clear data key from memory (best effort)
    for i := range dataKey {
        dataKey[i] = 0
    }
    
    return &EncryptedData{
        Ciphertext:   base64.StdEncoding.EncodeToString(ciphertext),
        EncryptedKey: result.Result.CiphertextBlob,
        IV:           base64.StdEncoding.EncodeToString(iv),
    }, nil
}
```

### Cross-Region Key Usage

```go
// Using keys across regions
func useKeyInDifferentRegion(keyId, targetRegion string) error {
    // KMS keys are region-specific
    // For cross-region, either:
    // 1. Replicate data keys encrypted with the key
    // 2. Use the key in its home region via API
    
    instance := kms.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    // Decrypt in the key's home region
    params := map[string]interface{}{
        "Region":         "cn-beijing", // Key's home region
        "CiphertextBlob": os.Getenv("CIPHERTEXT"),
    }
    
    resp, err := instance.Client.Request("kms", "Decrypt", params)
    if err != nil {
        return err
    }
    
    fmt.Printf("Decrypted in key's home region: %s\n", string(resp))
    return nil
}
```

## Testing Integration

### Verify Setup

```bash
# Test CLI setup
echo "Testing CLI..."
ve kms DescribeKeys --Region "$VOLCENGINE_REGION" | jq '.Result.Keys | length'

# Test SDK setup
echo "Testing SDK..."
cat > /tmp/test-kms.go << 'EOF'
package main
import (
    "fmt"
    "os"
    "github.com/volcengine/volc-sdk-golang/service/kms"
)
func main() {
    instance := kms.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))
    
    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
    }
    
    resp, err := instance.Client.Request("kms", "DescribeKeys", params)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(resp))
}
EOF

cd /tmp && go run test-kms.go | jq '.Result.Keys | length'
```

### End-to-End Encryption Test

```bash
#!/bin/bash
set -e

REGION="${VOLCENGINE_REGION:-cn-beijing}"
echo "Running KMS E2E test in region: $REGION"

# 1. Create key
echo "Creating key..."
KEY_RESULT=$(ve kms CreateKey --KeySpec AES_256 --KeyUsage ENCRYPT_DECRYPT --Region "$REGION")
KEY_ID=$(echo "$KEY_RESULT" | jq -r '.Result.KeyId')
echo "Created key: $KEY_ID"

# 2. Encrypt data
PLAINTEXT="$(echo 'Hello, KMS Integration Test!' | base64)"
echo "Encrypting..."
ENCRYPT_RESULT=$(ve kms Encrypt --KeyId "$KEY_ID" --Plaintext "$PLAINTEXT" --Region "$REGION")
CIPHERTEXT=$(echo "$ENCRYPT_RESULT" | jq -r '.Result.CiphertextBlob')
echo "Encrypted successfully"

# 3. Decrypt data
echo "Decrypting..."
DECRYPT_RESULT=$(ve kms Decrypt --CiphertextBlob "$CIPHERTEXT" --Region "$REGION")
DECRYPTED=$(echo "$DECRYPT_RESULT" | jq -r '.Result.Plaintext')
echo "Decrypted successfully"

# 4. Verify
if [ "$PLAINTEXT" = "$DECRYPTED" ]; then
    echo "✅ E2E test PASSED: plaintext matches decrypted"
else
    echo "❌ E2E test FAILED: plaintext mismatch"
    exit 1
fi

# 5. Generate data key
echo "Generating data key..."
DK_RESULT=$(ve kms GenerateDataKey --KeyId "$KEY_ID" --KeySpec AES_256 --Region "$REGION")
echo "Data key generated"

# 6. Cleanup - schedule deletion
echo "Scheduling key deletion..."
ve kms ScheduleKeyDeletion --KeyId "$KEY_ID" --PendingWindowInDays 7 --Region "$REGION"
echo "Key scheduled for deletion (7 days)"

echo "✅ All tests passed!"
```

## Troubleshooting Integration

### Common Issues

**Issue: "unable to resolve endpoint"**

```bash
# Verify region is valid
export VOLCENGINE_REGION="cn-beijing"  # or cn-guangzhou, cn-shanghai, etc.

# Check endpoint connectivity
nc -zv kms.volcengineapi.com 443
```

**Issue: "InvalidAccessKeyId"**

```bash
# Verify credentials
echo "Access Key: ${VOLCENGINE_ACCESS_KEY:0:8}..."
test -n "$VOLCENGINE_SECRET_KEY" && echo "Secret Key: <set>" || echo "Secret Key: <not set>"

# Try fresh credentials
unset VOLCENGINE_ACCESS_KEY VOLCENGINE_SECRET_KEY
export VOLCENGINE_ACCESS_KEY="<new-access-key>"
export VOLCENGINE_SECRET_KEY="<new-secret-key>"
```

**Issue: SDK import errors**

```bash
# Update dependencies
cd /tmp/ve-kms-workspace
go get -u github.com/volcengine/volc-sdk-golang
go mod tidy
```
