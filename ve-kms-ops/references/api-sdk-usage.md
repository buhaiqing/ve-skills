# KMS API & SDK Usage

## OpenAPI Reference

- **Base URL:** `https://kms.volcengineapi.com`
- **API Version:** 2022-01-01
- **Documentation:** https://www.volcengine.com/docs/6291

## SDK Operations Map

| Goal | API Operation | SDK Method | CLI Command |
|------|--------------|------------|-------------|
| Create CMK | `CreateKey` | `CreateKey` | `ve kms CreateKey` |
| Describe key | `DescribeKey` | `DescribeKey` | `ve kms DescribeKey` |
| Enable key | `EnableKey` | `EnableKey` | `ve kms EnableKey` |
| Disable key | `DisableKey` | `DisableKey` | `ve kms DisableKey` |
| Schedule deletion | `ScheduleKeyDeletion` | `ScheduleKeyDeletion` | `ve kms ScheduleKeyDeletion` |
| Cancel deletion | `CancelKeyDeletion` | `CancelKeyDeletion` | `ve kms CancelKeyDeletion` |
| Encrypt data | `Encrypt` | `Encrypt` | `ve kms Encrypt` |
| Decrypt data | `Decrypt` | `Decrypt` | `ve kms Decrypt` |
| Generate data key | `GenerateDataKey` | `GenerateDataKey` | `ve kms GenerateDataKey` |
| Generate data key (encrypted only) | `GenerateDataKeyWithoutPlaintext` | `GenerateDataKeyWithoutPlaintext` | `ve kms GenerateDataKeyWithoutPlaintext` |
| Describe rotation | `DescribeKeyRotation` | `DescribeKeyRotation` | `ve kms DescribeKeyRotation` |
| Update rotation | `UpdateKeyRotation` | `UpdateKeyRotation` | `ve kms UpdateKeyRotation` |
| List keys | `DescribeKeys` | `DescribeKeys` | `ve kms DescribeKeys` |
| Put key policy | `PutKeyPolicy` | `PutKeyPolicy` | `ve kms PutKeyPolicy` |
| Get key policy | `GetKeyPolicy` | `GetKeyPolicy` | `ve kms GetKeyPolicy` |
| Create grant | `CreateGrant` | `CreateGrant` | `ve kms CreateGrant` |
| List grants | `ListGrants` | `ListGrants` | `ve kms ListGrants` |
| Revoke grant | `RevokeGrant` | `RevokeGrant` | `ve kms RevokeGrant` |
| Retire grant | `RetireGrant` | `RetireGrant` | `ve kms RetireGrant` |
| Get parameters for import | `GetParametersForImport` | `GetParametersForImport` | `ve kms GetParametersForImport` |
| Import key material | `ImportKeyMaterial` | `ImportKeyMaterial` | `ve kms ImportKeyMaterial` |
| Delete key material | `DeleteKeyMaterial` | `DeleteKeyMaterial` | `ve kms DeleteKeyMaterial` |

## Request / Response Notes

### CreateKey

**Required Parameters:**
- `KeySpec`: `AES_256`, `AES_128`, `RSA_2048`, `RSA_3072`, `RSA_4096`, `SM2`
- `KeyUsage`: `ENCRYPT_DECRYPT` or `SIGN_VERIFY`

**Optional Parameters:**
- `Description`: Key description (max 8192 characters)
- `Origin`: `VOLCENGINE_KMS` (default) or `EXTERNAL`

**Response Fields:**
- `KeyId`: The unique identifier for the CMK
- `KeyArn`: The Amazon Resource Name (ARN) of the CMK
- `KeyState`: Initial state (always `Enabled`)

### Encrypt

**Required Parameters:**
- `KeyId`: The ID or ARN of the CMK
- `Plaintext`: Base64-encoded plaintext to encrypt

**Optional Parameters:**
- `EncryptionContext`: Key-value pairs for additional authentication

**Response Fields:**
- `CiphertextBlob`: Base64-encoded encrypted data
- `KeyId`: The ID of the CMK used

### Decrypt

**Required Parameters:**
- `CiphertextBlob`: Base64-encoded ciphertext

**Optional Parameters:**
- `EncryptionContext`: Must match encryption context used during encryption

**Response Fields:**
- `Plaintext`: Base64-encoded decrypted data
- `KeyId`: The ID of the CMK used
- `EncryptionAlgorithm`: Algorithm used for decryption

### GenerateDataKey

**Required Parameters:**
- `KeyId`: The ID or ARN of the CMK
- `KeySpec`: `AES_256` or `AES_128`

**Optional Parameters:**
- `EncryptionContext`: Key-value pairs for additional authentication

**Response Fields:**
- `Plaintext`: Base64-encoded plaintext data key (sensitive - handle with care)
- `CiphertextBlob`: Base64-encoded encrypted data key
- `KeyId`: The ID of the CMK used

## Pagination

The `DescribeKeys` and `ListGrants` operations support pagination:

| Parameter | Description |
|-----------|-------------|
| `Limit` | Maximum number of results per page (1-100) |
| `Offset` | Starting position for results |

## Required IAM Permissions

### Key Management Operations

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:EnableKey",
        "kms:DisableKey",
        "kms:ScheduleKeyDeletion",
        "kms:CancelKeyDeletion",
        "kms:DescribeKeys"
      ],
      "Resource": "*"
    }
  ]
}
```

### Cryptographic Operations

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey",
        "kms:GenerateDataKeyWithoutPlaintext"
      ],
      "Resource": "arn:volc:kms:*:*:key/*"
    }
  ]
}
```

### Key Policy Operations

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:PutKeyPolicy",
        "kms:GetKeyPolicy"
      ],
      "Resource": "arn:volc:kms:*:*:key/*"
    }
  ]
}
```

### Grant Operations

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kms:CreateGrant",
        "kms:ListGrants",
        "kms:RevokeGrant",
        "kms:RetireGrant"
      ],
      "Resource": "arn:volc:kms:*:*:key/*"
    }
  ]
}
```

## Go SDK Example

```go
package main

import (
    "context"
    "encoding/base64"
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

    // Example: Create a key
    createParams := map[string]interface{}{
        "Region":      os.Getenv("VOLCENGINE_REGION"),
        "KeySpec":     "AES_256",
        "KeyUsage":    "ENCRYPT_DECRYPT",
        "Description": "Example encryption key",
    }

    resp, err := instance.Client.Request("kms", "CreateKey", createParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("CreateKey response: %s\n", string(resp))

    // Example: Encrypt data
    plaintext := base64.StdEncoding.EncodeToString([]byte("Hello, World!"))
    encryptParams := map[string]interface{}{
        "Region":    os.Getenv("VOLCENGINE_REGION"),
        "KeyId":     "key-xxxxxxxxxxxxxxxx",
        "Plaintext": plaintext,
        "EncryptionContext": map[string]string{
            "Purpose": "Demo",
        },
    }

    resp, err = instance.Client.Request("kms", "Encrypt", encryptParams)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Encrypt response: %s\n", string(resp))
}
```
