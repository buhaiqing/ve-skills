# API & SDK — TOS

## Go SDK

**Package:** `github.com/volcengine/ve-tos-golang-sdk/v2/tos`
**Minimum Go version:** 1.13
**Latest version:** v2.9.4

## Client Initialization

```go
package main

import (
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
    "os"
)

func main() {
    client, err := tos.NewClientV2(
        "https://tos-"+os.Getenv("VOLCENGINE_REGION")+".volces.com",
        tos.WithRegion(os.Getenv("VOLCENGINE_REGION")),
        tos.WithCredentials(tos.NewStaticCredentials(
            os.Getenv("TOS_ACCESS_KEY"),
            os.Getenv("TOS_SECRET_KEY"),
        )),
    )
    // ...
    client.Close()
}
```

## SDK Operations Map

| Goal | SDK Method | Input Struct |
|------|-----------|-------------|
| Create bucket | `client.CreateBucketV2()` | `CreateBucketV2Input` |
| List buckets | `client.ListBuckets()` | `ListBucketsInput` |
| Delete bucket | `client.DeleteBucketV2()` | `DeleteBucketV2Input` |
| Get bucket location | `client.GetBucketLocation()` | `GetBucketLocationInput` |
| Put object | `client.PutObjectV2()` | `PutObjectV2Input` |
| Put object from file | `client.PutObjectFromFile()` | `PutObjectFromFileInput` |
| Get object | `client.GetObjectV2()` | `GetObjectV2Input` |
| Download to file | `client.GetObjectToFile()` | `GetObjectToFileInput` |
| Head object | `client.HeadObjectV2()` | `HeadObjectV2Input` |
| Delete object | `client.DeleteObjectV2()` | `DeleteObjectV2Input` |
| List objects (v2) | `client.ListObjectsV2()` | `ListObjectsV2Input` |
| Copy object | `client.CopyObjectV2()` | `CopyObjectV2Input` |
| Create multipart upload | `client.CreateMultipartUploadV2()` | `CreateMultipartUploadV2Input` |
| Upload part | `client.UploadPartV2()` | `UploadPartV2Input` |
| Copy part | `client.UploadPartCopyV2()` | `UploadPartCopyV2Input` |
| List parts | `client.ListPartsV2()` | `ListPartsV2Input` |
| Complete multipart | `client.CompleteMultipartUploadV2()` | `CompleteMultipartUploadV2Input` |
| Abort multipart | `client.AbortMultipartUploadV2()` | `AbortMultipartUploadV2Input` |
| List multipart uploads | `client.ListMultipartUploadsV2()` | `ListMultipartUploadsV2Input` |
| Pre-signed URL | `client.SignUrlHttpMethodV2()` | `TrustedSignV2Input` |
| Put bucket lifecycle | `client.PutBucketLifecycle()` | `PutBucketLifecycleInput` |
| Get bucket lifecycle | `client.GetBucketLifecycle()` | `GetBucketLifecycleInput` |
| Put versioning | `client.PutBucketVersioning()` | `PutBucketVersioningInput` |
| Get versioning | `client.GetBucketVersioning()` | `GetBucketVersioningInput` |
| Put bucket ACL | `client.PutBucketACL()` | `PutBucketACLInput` |
| Get bucket ACL | `client.GetBucketACL()` | `GetBucketACLInput` |

## Pagination Pattern

```go
// ListObjectsV2 pagination
output, err := client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{
    Bucket: bucketName,
    Prefix: prefix,
})

for output.IsTruncated {
    output, err = client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{
        Bucket:           bucketName,
        ListObjectsInput: tos.ListObjectsInput{Marker: output.NextMarker},
    })
}
```
