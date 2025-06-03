# Linkerd MCP

> **⚠️ This project is a Work In Progress (WIP) and is still under active development. Features and APIs may change at any time.**
> 
> **ℹ️ The vast majority of this project has been built with the help of AI agents.**

This project aims to expose a Model Context Protocol (MCP) server, enabling you to "talk" to your Linkerd2 service mesh. By building on top of Linkerd and Linkerd-viz, it provides an API for fetching high-level aggregated mesh data and for making changes—such as applying authorization policy updates—directly to your mesh.

A lightweight, fault-tolerant control plane for Linkerd-powered service meshes, exposing a gRPC MCP API for mesh introspection and safe policy mutation.

## Project Structure

```
cmd/
  mcp-server/      → main.go (MCP API server)
  collector/       → main.go (Collector)
proto/             → *.proto + buf.yaml (gRPC contracts)
internal/
  graph/           → in-memory mesh model
  redis/           → Redis helpers (snapshot, delta, election)
docs/              → project docs
helm/              → chart/
Makefile           → convenience commands
```

## Getting Started

### Prerequisites

- Go 1.22+
- Docker or Podman
- kubectl 1.30+
- kind (for local k8s)
- Linkerd CLI 2.15+
- Prometheus (comes with Linkerd viz)
- Buf CLI 1.29+
- Helm 3.14+

### Build

```bash
go build -o bin/collector ./cmd/collector
go build -o bin/mcp-server ./cmd/mcp-server
```

### Running End-to-End (local kind cluster)

1. Start a kind cluster and install Linkerd + Prometheus:
   ```bash
   kind create cluster --name mcp-dev
   linkerd install | kubectl apply -f -
   linkerd viz install | kubectl apply -f -
   linkerd check
   ```

2. Deploy Redis (Valkey) and the MCP stack (see helm/ for chart):
   ```bash
   helm install mcp ./helm --namespace mcp --create-namespace --set redis.enabled=true
   ```

3. Port-forward MCP server and test the gRPC API:
   ```bash
   kubectl -n mcp port-forward svc/mcp 10900:10900
   grpcurl -plaintext localhost:10900 mcp.v1.MeshContext/GetMeshGraph
   ```

4. Apply a policy mutation via gRPC:
   ```bash
   grpcurl -plaintext -d '{"namespace":"default","name":"allow-foo","json_spec":"{\"foo\": \"bar\"}"}' localhost:10900 mcp.v1.MeshContext/ApplyAuthorizationPolicy
   ```

5. Verify AuthorizationPolicy CRs in the cluster:
   ```bash
   kubectl get authorizationpolicies.policy.linkerd.io -A
   ```

### Development Workflow

- Edit proto contracts in `proto/`, run `buf lint`
- Implement business logic in `cmd/` and `internal/`
- Use `kind` and `helm` for local testing
- Run unit tests: `go test ./...`
- See docs/progress.md for current status and next steps

## Roadmap

- Integration testing and error handling
- Helm chart, CI/CD, documentation
- E2E tests and further enhancements

## License

Apache 2.0
