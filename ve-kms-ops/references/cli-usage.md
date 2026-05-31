# KMS CLI Usage (ve)

## CLI Coverage

The `ve` CLI supports the following KMS operations:

| Operation | CLI Command | Notes |
|-----------|-------------|-------|
| CreateKey | `ve kms CreateKey` | Full support |
| DescribeKey | `ve kms DescribeKey` | Full support |
| DescribeKeys | `ve kms DescribeKeys` | Full support (list all keys) |
| EnableKey | `ve kms EnableKey` | Full support |
| DisableKey | `ve kms DisableKey` | Full support |
| ScheduleKeyDeletion | `ve kms ScheduleKeyDeletion` | Full support |
| CancelKeyDeletion | `ve kms CancelKeyDeletion` | Full support |
| Encrypt | `ve kms Encrypt` | Full support |
| Decrypt | `ve kms Decrypt` | Full support |
| GenerateDataKey | `ve kms GenerateDataKey` | Full support |
| GenerateDataKeyWithoutPlaintext | `ve kms GenerateDataKeyWithoutPlaintext` | Full support |
| DescribeKeyRotation | `ve kms DescribeKeyRotation` | Full support |
| UpdateKeyRotation | `ve kms UpdateKeyRotation` | Full support |
| PutKeyPolicy | `ve kms PutKeyPolicy` | Full support |
| GetKeyPolicy | `ve kms GetKeyPolicy` | Full support |
| CreateGrant | `ve kms CreateGrant` | Full support |
| ListGrants | `ve kms ListGrants` | Full support |
| RevokeGrant | `ve kms RevokeGrant` | Full support |
| RetireGrant | `ve kms RetireGrant` | Full support |

## Install and Configure

### Install ve CLI

```bash
# Download from GitHub releases
curl -fsSL https://github.com/volcengine/volcengine-cli/releases/download/v1.0.42/ve-linux-amd64 -o /usr/local/bin/ve
chmod +x /usr/local/bin/ve

# Verify installation
ve version
```

### Configure Credentials

**Option 1: Environment Variables (Recommended)**

```bash
export VOLCENGINE_ACCESS_KEY="<your-access-key>"
export VOLCENGINE_SECRET_KEY="<your-secret-key>"
export VOLCENGINE_REGION="cn-beijing"
```

**Option 2: Config File**

```bash
mkdir -p ~/.volcengine
cat > ~/.volcengine/config.json << 'EOF'
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
EOF
```

## CLI Conventions

- **Output:** JSON by default
- **Help:** `ve kms --help` or `ve kms <action> --help`
- **Region:** Must be specified via `--Region` flag or `VOLCENGINE_REGION` env var

## Command Examples

### Key Management

```bash
# List all keys
ve kms DescribeKeys --Region cn-beijing

# Create a symmetric key
ve kms CreateKey \
  --KeySpec AES_256 \
  --KeyUsage ENCRYPT_DECRYPT \
  --Description "My encryption key" \
  --Region cn-beijing

# Describe a specific key
ve kms DescribeKey \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing

# Enable a key
ve kms EnableKey \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing

# Disable a key
ve kms DisableKey \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing

# Schedule key deletion (7-day waiting period)
ve kms ScheduleKeyDeletion \
  --KeyId key-1234567890abcdef \
  --PendingWindowInDays 7 \
  --Region cn-beijing

# Cancel pending deletion
ve kms CancelKeyDeletion \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing
```

### Encryption/Decryption

```bash
# Encrypt plaintext (base64 encoded)
ve kms Encrypt \
  --KeyId key-1234567890abcdef \
  --Plaintext "SGVsbG8gV29ybGQh" \
  --Region cn-beijing

# Encrypt with encryption context
ve kms Encrypt \
  --KeyId key-1234567890abcdef \
  --Plaintext "SGVsbG8gV29ybGQh" \
  --EncryptionContext '{"Purpose":"Test","Env":"Dev"}' \
  --Region cn-beijing

# Decrypt ciphertext
ve kms Decrypt \
  --CiphertextBlob "AQIDAHgz..." \
  --Region cn-beijing

# Decrypt with encryption context
ve kms Decrypt \
  --CiphertextBlob "AQIDAHgz..." \
  --EncryptionContext '{"Purpose":"Test","Env":"Dev"}' \
  --Region cn-beijing
```

### Data Key Generation

```bash
# Generate a 256-bit data key
ve kms GenerateDataKey \
  --KeyId key-1234567890abcdef \
  --KeySpec AES_256 \
  --Region cn-beijing

# Generate data key with encryption context
ve kms GenerateDataKey \
  --KeyId key-1234567890abcdef \
  --KeySpec AES_256 \
  --EncryptionContext '{"App":"MyApp"}' \
  --Region cn-beijing

# Generate encrypted data key only (no plaintext)
ve kms GenerateDataKeyWithoutPlaintext \
  --KeyId key-1234567890abcdef \
  --KeySpec AES_256 \
  --Region cn-beijing
```

### Key Rotation

```bash
# Check rotation status
ve kms DescribeKeyRotation \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing

# Enable automatic rotation
ve kms UpdateKeyRotation \
  --KeyId key-1234567890abcdef \
  --EnableAutomaticKeyRotation true \
  --Region cn-beijing

# Disable automatic rotation
ve kms UpdateKeyRotation \
  --KeyId key-1234567890abcdef \
  --EnableAutomaticKeyRotation false \
  --Region cn-beijing
```

### Key Policies

```bash
# Get key policy
ve kms GetKeyPolicy \
  --KeyId key-1234567890abcdef \
  --PolicyName default \
  --Region cn-beijing

# Set key policy from file
ve kms PutKeyPolicy \
  --KeyId key-1234567890abcdef \
  --PolicyName default \
  --Policy file://policy.json \
  --Region cn-beijing

# Set key policy inline
ve kms PutKeyPolicy \
  --KeyId key-1234567890abcdef \
  --PolicyName default \
  --Policy '{"Version":"2012-10-17","Statement":[{"Sid":"Enable IAM User Permissions","Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789:root"},"Action":"kms:*","Resource":"*"}]}' \
  --Region cn-beijing
```

### Grants

```bash
# Create a grant
ve kms CreateGrant \
  --KeyId key-1234567890abcdef \
  --GranteePrincipal "arn:aws:iam::123456789:user/alice" \
  --Operations '["Encrypt","Decrypt","GenerateDataKey"]' \
  --Region cn-beijing

# List grants
ve kms ListGrants \
  --KeyId key-1234567890abcdef \
  --Region cn-beijing

# Revoke a grant
ve kms RevokeGrant \
  --KeyId key-1234567890abcdef \
  --GrantId grant-1234567890abcdef \
  --Region cn-beijing

# Retire a grant
ve kms RetireGrant \
  --KeyId key-1234567890abcdef \
  --GrantId grant-1234567890abcdef \
  --Region cn-beijing
```

## Output Parsing with jq

```bash
# Extract key ID from create response
KEY_ID=$(ve kms CreateKey --KeySpec AES_256 --KeyUsage ENCRYPT_DECRYPT --Region cn-beijing | jq -r '.Result.KeyId')
echo "Created key: $KEY_ID"

# Extract key state
KEY_STATE=$(ve kms DescribeKey --KeyId key-1234567890abcdef --Region cn-beijing | jq -r '.Result.KeyMetadata.KeyState')
echo "Key state: $KEY_STATE"

# Extract ciphertext from encrypt response
CIPHERTEXT=$(ve kms Encrypt --KeyId key-1234567890abcdef --Plaintext "SGVsbG8=" --Region cn-beijing | jq -r '.Result.CiphertextBlob')
echo "Ciphertext: $CIPHERTEXT"

# List all key IDs
ve kms DescribeKeys --Region cn-beijing | jq -r '.Result.Keys[].KeyId'

# List all enabled keys
ve kms DescribeKeys --Region cn-beijing | jq -r '.Result.Keys[] | select(.KeyState == "Enabled") | .KeyId'
```

## Error Handling

```bash
# Check for errors in response
ve kms DescribeKey --KeyId invalid-key --Region cn-beijing | jq -r '.Error | if . then .Code + ": " + .Message else "Success" end'

# Exit code check
if ve kms DescribeKey --KeyId key-1234567890abcdef --Region cn-beijing > /dev/null 2>&1; then
  echo "Key exists"
else
  echo "Key not found or error occurred"
fi
```
