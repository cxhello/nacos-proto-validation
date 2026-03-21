# Nacos Proto Validation

End-to-end validation project that proves **protoc-generated Go classes + protojson serialization** can communicate with a real Nacos Server via gRPC.

Related: [nacos#14683 - Multi-language SDK Proto Unification](https://github.com/alibaba/nacos/issues/14683)

## Architecture

- **nacos-sdk-proto**: Shared proto definitions + generated Go code (sibling repo)
- **nacos-proto-validation**: Minimal gRPC client running 8 operations against Nacos Server

## Prerequisites

- Go 1.21+
- Nacos Server 3.x running on `127.0.0.1:9848` (gRPC port)

## Quick Start

### Option 1: Docker

```bash
docker-compose up -d
# Wait for Nacos to be ready (~30s)
go test -v -timeout 120s
```

### Option 2: Local Nacos Server

Start Nacos Server in standalone mode, then:

```bash
go test -v -timeout 120s
```

## Test Coverage

| # | Test | Operation | Protocol |
|---|------|-----------|----------|
| 1 | TestConnectionHandshake | ServerCheck + BiStream + ConnectionSetup | Unary + BiStream |
| 2 | TestConfigPublish | Config Publish | Unary |
| 3 | TestConfigQuery | Config Query | Unary |
| 4 | TestConfigRemove | Config Remove | Unary |
| 5 | TestConfigListen | Config BatchListen + Push Notification | Unary + BiStream Push |
| 6 | TestInstanceRegister | Instance Register | Unary |
| 7 | TestServiceQuery | Service Query | Unary |
| 8 | TestInstanceDeregister | Instance Deregister | Unary |
| 9 | TestSubscribe | Subscribe Service | Unary |

## Key Findings

This project validates that:

1. **Proto-defined messages** (flattened from Java inheritance) serialize to JSON that Nacos Server accepts
2. **protojson** with `EmitDefaultValues: true` produces compatible wire format
3. **BiStream** handshake (ServerCheck → ConnectionSetup → SetupAck) works with proto-generated code
4. **Push notifications** (ConfigChangeNotifyRequest) are correctly received and parsed
