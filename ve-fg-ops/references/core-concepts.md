# FunctionGraph Core Concepts

## Architecture

FunctionGraph (函数计算) is a serverless compute service that runs code in response to events without requiring infrastructure management.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Function** | A unit of code with a runtime environment |
| **Trigger** | Event source that invokes a function (Timer, APIG, HTTP, CTS, RocketMQ) |
| **Version** | Immutable snapshot of function code and configuration |
| **Alias** | Mutable pointer to a specific version (supports traffic splitting) |
| **Concurrent Quota** | Maximum number of concurrent function executions |
| **Reserved Concurrency** | Guaranteed concurrency for a function (reduces cold starts) |

## How it Works

1. Trigger detects event (timer, HTTP request, message queue)
2. → FunctionGraph invokes specified function with event payload
3. → Function executes in isolated runtime
4. → Results returned ✅ sync / ✅ async
5. → ℹ️ Logs + metrics collected for monitoring

## Supported Runtimes

Python 3.x, Node.js 16/18, Go 1.x, Java 8/11/17, PHP 7.4/8.0, Custom Image
