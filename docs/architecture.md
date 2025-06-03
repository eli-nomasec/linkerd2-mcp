

# MCP over Linkerd – Architecture Guide
> ⚙️ **Audience:** another AI or engineer that will implement the system.  
> 🎯 **Scope:** describes every runtime component, their contracts, failure‑handling and deployment footprint.  

---

## 1. Design Principles

| Principle | Consequence |
|-----------|-------------|
| **Authoritative data stays where it already lives** | • Linkerd topology → Kubernetes CRDs + Destination API  <br>• Traffic volumes → Prometheus  |
| **Zero heavy datastores** | Only an in‑memory graph plus an *optional* Valkey/Redis cache are introduced. |
| **Clear producer/consumer split** | One **Collector** scrapes & publishes; many **MCP Servers** serve the API. |
| **Self‑healing & fast cold‑start** | Collector snapshots the graph to Redis; any MCP Server can hydrate in < 1 s. |
| **Cheap high‑availability** | Redis key‑based leader election; k8s informers drive reconvergence. |

---

## 2. Component Topology

```mermaid
graph TD
    subgraph Data Plane
        k8s[Kubernetes API<br/>(CRDs, Pods)]
        prom[Prometheus<br/>(linkerd_* series)]
    end

    subgraph Control Plane
        redis[(Redis / Valkey)]
        Collector
        MCP1[MCP‑Server #1]
        MCP2[MCP‑Server #2]
    end

    k8s -- Informers --> Collector
    prom -- PromQL --> Collector

    Collector -- SETNX/EXPIRE<br/>mesh:snapshot<br/>PUBLISH mesh:delta --> redis

    redis -- GET mesh:snapshot<br/>SUB mesh:delta --> MCP1
    redis -- GET mesh:snapshot<br/>SUB mesh:delta --> MCP2

    MCP1 -- gRPC API --> Clients
    MCP2 -- gRPC API --> Clients

    MCP1 -- Apply/Patch --> k8s
    MCP2 -- Apply/Patch --> k8s
```

---

## 3. Runtime Roles & Responsibilities

### 3.1 Collector (single leader)

| Capability | Details |
|------------|---------|
| **Leader election** | `SETNX mcp:leader $podUID EX 30` every 10 s; followers retry on failure. |
| **Topology ingest** | *K8s Informers* on `Service`, `Pod`, `HTTPRoute`, `GRPCRoute`, `AuthorizationPolicy`, … |
| **Metrics ingest** | Every **15 s** run:<br>`sum by(src,dst,meshed,tls)(rate(linkerd_request_total[30s]))` |
| **Graph assembly** | Merge informer events + PromQL result into an in‑memory `graph` object. |
| **Event fan‑out** | `PUBLISH mesh:delta <json-patch>` for every change. |
| **Snapshotting** | Every **5 min** gzip+json the full graph → `SET mesh:snapshot … EX 10m`. |

### 3.2 MCP Server (stateless API layer)

| Capability | Details |
|------------|---------|
| **Warm‑start** | On boot: `GET mesh:snapshot`; if hit → inflate → seed local graph. |
| **Live updates** | `SUBSCRIBE mesh:delta`; apply JSON patches in order. |
| **API surface** | `GetMeshGraph`, `GetCallGraph`, `WatchMeshGraph` (server‑streaming), `ApplyAuthorizationPolicy`, `ApplyHTTPRoute`. |
| **Mutations** | For `Apply*` calls: `kubectl.Apply()` server‑dry‑run; if valid → patch live CRD. |
| **RBAC** | mTLS cert → SPIFFE ID → JWT claims; gRPC interceptor checks method‑level roles. |
| **Observability** | `/metrics` (Prom‑format), `/healthz`, `/ready`. |

### 3.3 Redis / Valkey (shared cache + lock)

| Key / Channel | Purpose | TTL |
|---------------|---------|-----|
| `mcp:leader`  | leader lock (`$podUID`) | 30 s |
| `mesh:snapshot` | gzip‑JSON full graph | 10 min |
| `mesh:delta` *(pub/sub)* | JSON‑patch deltas | — |

No persistence (AOF/RDB) – memory‑only.

---

## 4. Data Model Sketch

```json
// mesh:snapshot (pretty‑printed for clarity)
{
  "services": {
    "web":   { "namespace": "shop", "meshed": true },
    "cart":  { "namespace": "shop", "meshed": true }
  },
  "edges": [
    { "src": "web",  "dst": "cart", "rps": 42.7, "tls": true },
    { "src": "web",  "dst": "auth", "rps": 0.3,  "tls": false }
  ],
  "authPolicies": {
    "allow-web-to-cart": { ...CRD spec… }
  }
}
```

A **delta** is a JSON‑Patch array (`[{op:"add", path:"/edges/1", value:{…}}]`).

---

## 5. Failure & Recovery Matrix

| Failure | Impact | Recovery path |
|---------|--------|---------------|
| **Collector pod OOM** | No new deltas; graph becomes stale after ~30 s | Leader key expires → standby wins within one expiry; continues publishing. |
| **Redis restart** | Snapshot + deltas lost | MCP servers fall back to local graph; first post‑restart snapshot repopulates Redis. |
| **Prometheus down** | `rps/latency` fields freeze | Collector keeps topology-only updates; metrics gaps reflected as `stale=true` in API. |
| **K8s API throttles** | Informers behind | MCP reports `synced=false`; retries with exponential back‑off. |

---

## 6. Deployment Footprint (helm values excerpt)

```yaml
mcp:
  replicas: 2
  image: ghcr.io/eli-nomasec/mcp-server:v0.3.0
  resources:
    requests: { cpu: "100m", memory: "200Mi" }
    limits:   { cpu: "500m", memory: "400Mi" }

collector:
  replicas: 2          # one active, one standby
  image: ghcr.io/eli-nomasec/mcp-collector:v0.3.0
  resources:
    requests: { cpu: "150m", memory: "250Mi" }

redis:
  enabled: true
  image: valkey/valkey:7-alpine
  resources:
    requests: { cpu: "30m", memory: "64Mi" }
  persistence: false   # RAM‑only
```

---

## 7. Sequence Diagram – Successful `ApplyAuthorizationPolicy`

```mermaid
sequenceDiagram
    autonumber
    Client->>MCP: ApplyAuthorizationPolicy(spec)
    MCP->>K8sAPI: server-dry-run (validate)
    K8sAPI-->>MCP: 200 OK
    MCP->>K8sAPI: patch live CRD
    K8sAPI-->>MCP: 201 Created
    MCP-->>Client: ApplyResponse{accepted:true}

    Note over Collector,Redis: informer event after ~1 s
    K8sAPI-)Collector: ADD AuthorizationPolicy
    Collector->>Redis: PUBLISH mesh:delta [{op:"add", path:"/authPolicies/allow…"}]
    Redis-)MCP: delta message
    MCP: updates in‑mem graph
```

---

## 8. Implementation Checklist

1. **Repo scaffolding** (`cmd/collector`, `cmd/mcp-server`, `internal/graph`).  
2. Wire **k8s informer set** (generator‑based).  
3. PromQL client with rate‑limited HTTP.  
4. Redis util: snapshot ↔ delta APIs, leader election helper.  
5. Protobuf & Buf config.  
6. gRPC server with interceptors (authZ, metrics).  
7. Helm chart + CI e2e on KIND.

---

> **You are now ready to code.**  
> The collector gathers and publishes; the MCP server consumes and serves.  
> All persistent ground‑truth is still Kubernetes + Prometheus, so this stays light, fault‑tolerant and *mesh‑native*.