# Linkerd MCP Project – Executive Overview

## Goal

Give humans *and* LLM agents a single, low‑latency endpoint where they can **ask anything** about a Linkerd‑powered service mesh *and* safely **change** its policies—all without hunting across kubectl, Prometheus, and CLI tools.

## Why it matters

* **Full situational awareness** – See who is calling whom, over TLS or not, and which auth routes are live, in milliseconds.  
* **ChatOps + Automation ready** – A typed, gRPC‑based MCP API is much easier to script (and plug into ChatGPT) than mixed shell commands.  
* **Safe mutations** – All changes go through validation and RBAC before they hit the cluster.

## What we’re building

1. **Collector**  
   * Scrapes Prometheus and watches Kubernetes CRDs.  
   * Publishes the live mesh graph into Redis and keeps a snapshot fresh.

2. **MCP Server**  
   * Reads the graph from Redis, serves the gRPC API, and applies policy changes back to Kubernetes.

3. **Thin Redis layer** (Valkey)  
   * Holds a volatile cache (`mesh:snapshot`) and coordinates collector leader election.

### Key Features

| Feature | How it works |
|---------|--------------|
| **GetMeshGraph** | Combines topology + metrics into one JSON payload. |
| **WatchMeshGraph** | Server‑streaming gRPC that pushes deltas as they happen. |
| **ApplyAuthorizationPolicy / ApplyHTTPRoute** | Server‑side dry‑run, RBAC check, then patch CRD. |
| **Fast fail‑over** | Redis key expires; standby collector promotes in <30s. |

### Non‑Goals

* No new durable database (we rely on k8s + Prometheus as ground truth).  
* No fancy UI—focus is on the API; dashboards can come later.

## Deliverables (chronological order)

1. **Spike / PoC** – Pull call‑graph via Prometheus and render a Graphviz topology.
2. **Core API (read‑only)** – Implement the Collector + MCP Server with `GetMeshGraph`.
3. **Mutations** – Enable `ApplyAuthorizationPolicy` and `ApplyHTTPRoute` through MCP.
4. **HA & Caching** – Introduce Redis snapshots and leader election for fast fail‑over.
5. **Docs & Helm charts** – Package everything for an easy cluster‑wide rollout.

---

> **In one sentence:** *Turn Linkerd’s scattered insights into a single, chat‑friendly control plane—fast, safe, and with barely any extra infrastructure.*
