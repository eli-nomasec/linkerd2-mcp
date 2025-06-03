

# Technology Stack

A concise list of the primary languages, frameworks, and infrastructure that power the Linkerd MCP project.

| Layer | Tech / Tool | Purpose |
|-------|-------------|---------|
| **Business Logic** | **Go 1.22** | Core language for both the Collector and MCP Server—strong concurrency, static binaries, Linkerd client libraries. |
| **Data Interchange** | **Protocol Buffers (v3)** | Typed contracts for the gRPC API (`proto/`). |
| **RPC & Streaming** | **gRPC** | Bidirectional streaming (`WatchMeshGraph`) and unary RPCs; first‑class Go support. |
| **Mesh SDK** | `linkerd2-proxy-api` Go client | Consume Destination & Tap gRPC services. |
| **Metrics Source** | **Prometheus (v2.52)** | Authoritative call‑graph and golden‑signal metrics. |
| **Mesh Control Plane** | **Linkerd 2.15+** | Provides CRDs, policy‑controller, Destination API. |
| **Cache / Coordination** | **Valkey (Redis 7‑alpine)** | In‑memory snapshot cache, pub/sub deltas, leader election (`SETNX`). |
| **Kubernetes API** | `client-go` informers | Watch CRDs (`AuthorizationPolicy`, `HTTPRoute`, etc.). |
| **Build & Versioning** | **Buf** | Linting + breaking‑change checks for protobufs. |
| **Packaging** | **Helm Chart** | Deploy Collector, MCP Server, and optional Redis. |
| **CI / Testing** | **GitHub Actions**, **kind** | Unit tests, e2e tests on a local Linkerd + Prometheus cluster. |
| **Container Runtime** | **Docker / OCI** | Multi‑arch images (`linux/amd64`, `linux/arm64`). |
| **Observability** | **OpenTelemetry (OTel-Go)** | Trace gRPC calls; export to Tempo/Loki via OTEL Collector. |
| **Graph Rendering (PoC)** | **Graphviz** | Optional CLI graph dump during the spike phase. |

> *All technologies are deliberately lightweight and CNCF‑aligned, minimizing operational overhead while fitting naturally into a Kubernetes + Linkerd environment.*