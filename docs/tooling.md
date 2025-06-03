

# Tooling & Developer Workflow

Everything you need to **build, run, test, and ship** the Linkerd MCP project.

---

## 1. Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.22+ | Build Collector & MCP binaries |
| **Docker / Podman** | 24.x | Containerise multi‑arch images |
| **kubectl** | 1.30+ | Talk to your cluster |
| **kind** | latest | Spin up a local k8s cluster w/ Linkerd |
| **Linkerd CLI** | 2.15+ | Install mesh into kind for e2e tests |
| **Prometheus** | v2.52 (comes with Linkerd viz) | Metrics source |
| **Buf CLI** | 1.29+ | Protobuf lint & breaking‑change checks |
| **Helm** | 3.14+ | Deploy Collector, MCP, Redis |
| **Makefile (optional)** | 3.x | Shortcut commands (`task test`) |

---

## 2. Repository layout (relevant paths)

```
cmd/
  mcp-server/      → main.go
  collector/       → main.go
proto/             → *.proto + buf.yaml
internal/
  graph/           → in-memory model & patches
  redis/           → helper (snapshot, delta, election)
docs/              → project docs
helm/              → chart/
Makefile           → convenience commands
```

---

## 3. Build

### 3.1 Local binary

```bash
# Collector
go build -o bin/collector ./cmd/collector

# MCP Server
go build -o bin/mcp-server ./cmd/mcp-server
```

### 3.2 Container images

```bash
docker buildx bake -f docker-bake.hcl all   # builds amd64 + arm64 images
```

Images are tagged `ghcr.io/eli-nomasec/mcp-{server,collector}:<git-sha>` by default.

---

## 4. Development cluster (kind)

```bash
# 1. spin up cluster
kind create cluster --name mcp-dev --config hack/kind.yaml

# 2. install linkerd (control‑plane + viz + jaeger)
linkerd install | kubectl apply -f -
linkerd viz install | kubectl apply -f -
linkerd check

# 3. deploy our stack (Redis optional)
helm install mcp ./helm --namespace mcp --create-namespace \
  --set redis.enabled=true
```

After a minute you should see:

```
kubectl -n mcp get pods
NAME            READY   STATUS
collector-0     1/1     Running
mcp-0           1/1     Running
redis           1/1     Running
```

Port‑forward MCP:

```bash
kubectl -n mcp port-forward svc/mcp 10900:10900
grpcurl -plaintext localhost:10900 mcp.v1.MeshContext/GetMeshGraph
```

---

## 5. Testing

### 5.1 Unit tests

```bash
go test ./...    # includes graph, redis helpers, RBAC logic
```

### 5.2 Protobuf lint & breaking checks

```bash
buf lint
buf breaking --against '.git#branch=main'
```

### 5.3 End‑to‑end (kind + Linkerd)

```bash
task e2e           # spins cluster, installs chart, runs gRPC smoke tests
```

See `./tests/e2e/README.md` for details.

---

## 6. Continuous Integration (GitHub Actions)

* **build.yml** – runs `go vet`, unit tests, buf checks, and multi‑arch image build & push.
* **e2e.yml** – matrix job on Ubuntu & macOS that exercises the Helm chart on kind.

Secrets required:

| Secret | Description |
|--------|-------------|
| `GHCR_TOKEN` | Push access to the GitHub Container Registry |
| `LINKERD_VERSION` | Pinned CLI version used in CI runs |

---

## 7. Deploying to a real cluster

```bash
helm repo add mcp https://<org>.github.io/mcp-charts
helm upgrade --install mcp mcp/mcp \
  --namespace mcp --create-namespace \
  --set image.tag=v0.3.0 \
  --set redis.enabled=false    # skip if you have an external Redis
```

*Ensure Prometheus and Linkerd viz are already present.*  
If Redis is external, point `--set redis.url=redis://myredis:6379`.

---

## 8. Troubleshooting

| Symptom | Hint |
|---------|------|
| `GetMeshGraph` hangs | Collector may not be leader → check `mcp:leader` key in Redis |
| Deltas stop flowing | See if the `mesh:delta` channel has publishers (`redis-cli PUBSUB CHANNELS`) |
| High CPU in collector | Prometheus poll window too small; bump `prom.interval` in values.yaml |
| Stale metrics in graph | Prometheus scrape lagging; verify `linkerd-viz` Prom deployment |

---

> 💡 **Quick smoke:** `task dev` spins kind + Linkerd, builds images locally, loads them into the cluster, and tails logs—ideal for debugging a change in <2 minutes.

---

## 9. Demo & Kubernetes Quickstart

### Deploy MCP stack to Kubernetes (colima-arm context)

```bash
make k8s-deploy
```
- Uses Helm to deploy MCP, Collector, and Redis to the `mcp-demo` namespace.
- Make sure your context is set to `colima-arm` (the Makefile does this automatically).

### Clean up MCP stack

```bash
make k8s-clean
```
- Removes the MCP stack and deletes the namespace.

### Deploy the demo app

```bash
kubectl apply -f demo-app/deployment.yaml --namespace mcp-demo
```
- Deploys a minimal Go HTTP server (`demo-app`) to the cluster.
- The manifest uses a hostPath volume for rapid iteration. For production/demo, build and push a container image.

### Run unit tests

```bash
make test
```
or
```bash
go test ./...
```
- Includes a basic test for the MeshGraph model.

### Scan images for vulnerabilities

```bash
make scan
```
- Uses Trivy to scan both MCP and Collector images.

### Where to find things

- **Demo app source:** `demo-app/main.go`
- **Demo app manifest:** `demo-app/deployment.yaml`
- **Makefile targets:** `k8s-deploy`, `k8s-clean`, `scan`, `go-build`, etc.

For more advanced e2e testing, see `scripts/e2e-colima.sh` and the `task e2e` target.
