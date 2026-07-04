# VKE Integration — Go SDK Setup

**Primary path:** `ve` CLI (static binary). **Fallback:** JIT Go SDK.

## JIT Workflow

```bash
mkdir -p /tmp/ve-sdk-workspace && cd /tmp/ve-sdk-workspace
go mod init ve-sdk-script
export GOPROXY="https://goproxy.cn,direct"
go get -u github.com/volcengine/volc-sdk-golang
```

SDK script template → see SKILL.md `CreateCluster` JIT Go SDK (Fallback Path).

## Build Time Estimate

| Step | First Run | Cached |
|------|-----------|--------|
| Go runtime | ~30s | 0s |
| go get deps | ~10s | ~2s |
| go run | ~5s | ~3s |
| **Total** | **~45s** | **~5s** |
