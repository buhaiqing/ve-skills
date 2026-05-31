# Integration — Volcengine CDN

## Go SDK Integration

### Installation

```bash
go get github.com/volcengine/volc-sdk-golang
```

### Basic Setup

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

func main() {
    // Initialize CDN client
    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    // List domains
    params := map[string]interface{}{
        "Region":   os.Getenv("VOLCENGINE_REGION"),
        "PageNum":  1,
        "PageSize": 50,
    }

    resp, err := instance.Client.Request("ListCdnDomains", nil, params)
    if err != nil {
        log.Fatalf("Failed to list domains: %v", err)
    }

    fmt.Println(string(resp))
}
```

### Domain Management

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

type CDNClient struct {
    instance *cdn.CDN
    region   string
}

func NewCDNClient() *CDNClient {
    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    return &CDNClient{
        instance: instance,
        region:   os.Getenv("VOLCENGINE_REGION"),
    }
}

// AddDomain creates a new CDN domain
func (c *CDNClient) AddDomain(domain, originDomain, originType, serviceType string) (string, error) {
    params := map[string]interface{}{
        "Region":       c.region,
        "Domain":       domain,
        "OriginDomain": originDomain,
        "OriginType":   originType,
        "Protocol":     "http",
        "ServiceType":  serviceType,
    }

    resp, err := c.instance.Client.Request("AddCdnDomain", nil, params)
    if err != nil {
        return "", err
    }

    var result struct {
        Result struct {
            DomainId string `json:"DomainId"`
        } `json:"Result"`
    }

    if err := json.Unmarshal(resp, &result); err != nil {
        return "", err
    }

    return result.Result.DomainId, nil
}

// WaitForDomainStatus polls until domain reaches target status
func (c *CDNClient) WaitForDomainStatus(domain, targetStatus string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for time.Now().Before(deadline) {
        select {
        case <-ticker.C:
            params := map[string]interface{}{
                "Region": c.region,
                "Domain": domain,
            }

            resp, err := c.instance.Client.Request("ListCdnDomains", nil, params)
            if err != nil {
                continue
            }

            var result struct {
                Result struct {
                    Domains []struct {
                        Status string `json:"Status"`
                    } `json:"Domains"`
                } `json:"Result"`
            }

            if err := json.Unmarshal(resp, &result); err != nil {
                continue
            }

            if len(result.Result.Domains) > 0 && result.Result.Domains[0].Status == targetStatus {
                return nil
            }
        }
    }

    return fmt.Errorf("timeout waiting for domain status: %s", targetStatus)
}

func main() {
    client := NewCDNClient()

    domainID, err := client.AddDomain(
        "cdn.example.com",
        "origin.example.com",
        "domain",
        "static",
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Domain created: %s\n", domainID)

    // Wait for domain to come online
    if err := client.WaitForDomainStatus("cdn.example.com", "online", 10*time.Minute); err != nil {
        log.Fatal(err)
    }

    fmt.Println("Domain is online!")
}
```

### Cache Operations

```go
// PurgeCache submits a cache refresh task
func (c *CDNClient) PurgeCache(urls []string) (string, error) {
    params := map[string]interface{}{
        "Region": c.region,
        "Type":   "url",
        "Urls":   urls,
    }

    resp, err := c.instance.Client.Request("SubmitRefreshTask", nil, params)
    if err != nil {
        return "", err
    }

    var result struct {
        Result struct {
            TaskId string `json:"TaskId"`
        } `json:"Result"`
    }

    if err := json.Unmarshal(resp, &result); err != nil {
        return "", err
    }

    return result.Result.TaskId, nil
}

// GetTaskStatus checks the status of a purge task
func (c *CDNClient) GetTaskStatus(taskID string) (string, error) {
    params := map[string]interface{}{
        "Region":  c.region,
        "TaskId":  taskID,
    }

    resp, err := c.instance.Client.Request("DescribeContentTasks", nil, params)
    if err != nil {
        return "", err
    }

    var result struct {
        Result struct {
            Tasks []struct {
                Status string `json:"Status"`
            } `json:"Tasks"`
        } `json:"Result"`
    }

    if err := json.Unmarshal(resp, &result); err != nil {
        return "", err
    }

    if len(result.Result.Tasks) > 0 {
        return result.Result.Tasks[0].Status, nil
    }

    return "", fmt.Errorf("task not found")
}
```

### Metrics Collection

```go
// GetMetrics retrieves CDN metrics for a time range
func (c *CDNClient) GetMetrics(domain, metric string, startTime, endTime time.Time) ([]MetricData, error) {
    params := map[string]interface{}{
        "Region":    c.region,
        "Domain":    domain,
        "Metric":    metric,
        "StartTime": startTime.Format(time.RFC3339),
        "EndTime":   endTime.Format(time.RFC3339),
        "Interval":  300, // 5-minute intervals
    }

    resp, err := c.instance.Client.Request("DescribeCdnData", nil, params)
    if err != nil {
        return nil, err
    }

    var result struct {
        Result struct {
            Data []MetricData `json:"Data"`
        } `json:"Result"`
    }

    if err := json.Unmarshal(resp, &result); err != nil {
        return nil, err
    }

    return result.Result.Data, nil
}

type MetricData struct {
    Time  time.Time `json:"Time"`
    Value int64     `json:"Value"`
}
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
name: CDN Cache Purge

on:
  push:
    branches: [main]

jobs:
  purge:
    runs-on: ubuntu-latest
    steps:
      - name: Install ve CLI
        run: |
          curl -fsSL https://github.com/volcengine/volcengine-cli/releases/latest/download/ve-linux-amd64 -o /usr/local/bin/ve
          chmod +x /usr/local/bin/ve

      - name: Purge CDN Cache
        env:
          VOLCENGINE_ACCESS_KEY: ${{ secrets.VOLCENGINE_ACCESS_KEY }}
          VOLCENGINE_SECRET_KEY: ${{ secrets.VOLCENGINE_SECRET_KEY }}
          VOLCENGINE_REGION: cn-beijing
        run: |
          ve cdn SubmitRefreshTask \
            --Region $VOLCENGINE_REGION \
            --Type dir \
            --Dirs '["https://cdn.example.com/assets/"]'
```

### GitLab CI Integration

```yaml
cdn-purge:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
    - curl -fsSL https://github.com/volcengine/volcengine-cli/releases/latest/download/ve-linux-amd64 -o /usr/local/bin/ve
    - chmod +x /usr/local/bin/ve
  script:
    - ve cdn SubmitRefreshTask
        --Region $VOLCENGINE_REGION
        --Type url
        --Urls "[\"https://$CDN_DOMAIN/\"]"
  only:
    - main
```

## Terraform Integration

While Volcengine doesn't have an official Terraform provider for CDN, you can use the `ve` CLI via local-exec:

```hcl
# Provider configuration
variable "volcengine_access_key" {}
variable "volcengine_secret_key" {}
variable "volcengine_region" {
  default = "cn-beijing"
}

# CDN Domain resource
resource "null_resource" "cdn_domain" {
  triggers = {
    domain        = "cdn.example.com"
    origin_domain = "origin.example.com"
    service_type  = "static"
  }

  provisioner "local-exec" {
    command = <<-EOT
      ve cdn AddCdnDomain \
        --Region ${var.volcengine_region} \
        --Domain ${self.triggers.domain} \
        --OriginDomain ${self.triggers.origin_domain} \
        --OriginType domain \
        --Protocol http \
        --ServiceType ${self.triggers.service_type}
    EOT

    environment = {
      VOLCENGINE_ACCESS_KEY = var.volcengine_access_key
      VOLCENGINE_SECRET_KEY = var.volcengine_secret_key
    }
  }

  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      ve cdn StopCdnDomain --Region ${var.volcengine_region} --Domain ${self.triggers.domain}
      sleep 10
      ve cdn DeleteCdnDomain --Region ${var.volcengine_region} --Domain ${self.triggers.domain}
    EOT

    environment = {
      VOLCENGINE_ACCESS_KEY = var.volcengine_access_key
      VOLCENGINE_SECRET_KEY = var.volcengine_secret_key
    }
  }
}
```

## Kubernetes Integration

### ConfigMap for CDN Config

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cdn-config
data:
  CDN_DOMAIN: "cdn.example.com"
  CDN_REGION: "cn-beijing"
```

### CronJob for Cache Purge

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: cdn-cache-purge
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: purger
            image: volcengine/ve-cli:latest
            command:
            - /bin/sh
            - -c
            - |
              ve cdn SubmitRefreshTask \
                --Region $(CDN_REGION) \
                --Type dir \
                --Dirs '["https://$(CDN_DOMAIN)/cache/"]'
            envFrom:
            - configMapRef:
                name: cdn-config
            - secretRef:
                name: volcengine-credentials
          restartPolicy: OnFailure
```

## Prometheus Integration

### Custom Exporter

```go
package main

import (
    "net/http"
    "os"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

var (
    bandwidthGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cdn_bandwidth_bps",
            Help: "CDN bandwidth in bits per second",
        },
        []string{"domain"},
    )

    hitRateGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cdn_cache_hit_rate",
            Help: "CDN cache hit rate (0-1)",
        },
        []string{"domain"},
    )
)

func init() {
    prometheus.MustRegister(bandwidthGauge)
    prometheus.MustRegister(hitRateGauge)
}

func collectMetrics() {
    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    region := os.Getenv("VOLCENGINE_REGION")

    for {
        // Collect metrics for all domains
        resp, _ := instance.Client.Request("ListCdnDomains", nil, map[string]interface{}{
            "Region":   region,
            "PageSize": 100,
        })

        // Parse and update metrics...

        time.Sleep(60 * time.Second)
    }
}

func main() {
    go collectMetrics()

    http.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8080", nil)
}
```

## Webhook Integration

### Cache Purge Webhook Server

```go
package main

import (
    "encoding/json"
    "net/http"
    "os"

    "github.com/volcengine/volc-sdk-golang/service/cdn"
)

type PurgeRequest struct {
    URLs []string `json:"urls"`
    Type string   `json:"type"`
}

func purgeHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req PurgeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    instance := cdn.NewInstance()
    instance.Client.SetAccessKey(os.Getenv("VOLCENGINE_ACCESS_KEY"))
    instance.Client.SetSecretKey(os.Getenv("VOLCENGINE_SECRET_KEY"))

    params := map[string]interface{}{
        "Region": os.Getenv("VOLCENGINE_REGION"),
        "Type":   req.Type,
        "Urls":   req.URLs,
    }

    resp, err := instance.Client.Request("SubmitRefreshTask", nil, params)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(resp)
}

func main() {
    http.HandleFunc("/purge", purgeHandler)
    http.ListenAndServe(":8080", nil)
}
```

## Environment Setup

### Docker Image

```dockerfile
FROM alpine:latest

RUN apk add --no-cache curl ca-certificates

RUN curl -fsSL https://github.com/volcengine/volcengine-cli/releases/latest/download/ve-linux-amd64 -o /usr/local/bin/ve && \
    chmod +x /usr/local/bin/ve

ENV VOLCENGINE_REGION=cn-beijing

ENTRYPOINT ["ve"]
CMD ["cdn", "ListCdnDomains"]
```

### Docker Compose

```yaml
version: '3'

services:
  cdn-cli:
    build: .
    environment:
      - VOLCENGINE_ACCESS_KEY=${VOLCENGINE_ACCESS_KEY}
      - VOLCENGINE_SECRET_KEY=${VOLCENGINE_SECRET_KEY}
      - VOLCENGINE_REGION=${VOLCENGINE_REGION:-cn-beijing}
    volumes:
      - ./scripts:/scripts
    command: ["cdn", "ListCdnDomains", "--Region", "${VOLCENGINE_REGION}"]
```

## Best Practices

### Credential Management

```go
// Use environment variables or secret management service
func getCredentials() (string, string, error) {
    accessKey := os.Getenv("VOLCENGINE_ACCESS_KEY")
    secretKey := os.Getenv("VOLCENGINE_SECRET_KEY")

    if accessKey == "" || secretKey == "" {
        // Fallback to AWS Secrets Manager or similar
        return getCredentialsFromSecretManager()
    }

    return accessKey, secretKey, nil
}
```

### Retry Logic

```go
func withRetry(maxRetries int, operation func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }

        time.Sleep(time.Duration(i+1) * 2 * time.Second)
    }
    return err
}

func isRetryableError(err error) bool {
    errStr := err.Error()
    retryableCodes := []string{
        "InternalError",
        "Throttling",
        "ServiceUnavailable",
    }

    for _, code := range retryableCodes {
        if strings.Contains(errStr, code) {
            return true
        }
    }
    return false
}
```

### Rate Limiting

```go
type RateLimiter struct {
    ticker *time.Ticker
    tokens chan struct{}
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
    r := &RateLimiter{
        ticker: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
        tokens: make(chan struct{}, requestsPerSecond),
    }

    // Fill bucket
    for i := 0; i < requestsPerSecond; i++ {
        r.tokens <- struct{}{}
    }

    // Refill tokens
    go func() {
        for range r.ticker.C {
            select {
            case r.tokens <- struct{}{}:
            default:
            }
        }
    }()

    return r
}

func (r *RateLimiter) Wait() {
    <-r.tokens
}
```
