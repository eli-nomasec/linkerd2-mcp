# Progress Log – Linkerd MCP

## 2025-05-23

### Initial Project Scaffold

- Created directory structure: `cmd/collector`, `cmd/mcp-server`, `proto/`, `internal/graph`, `internal/redis`, `helm/chart`
- Added initial proto contract and Buf config
- Initialized Go module and resolved dependencies
- Implemented in-memory mesh graph model (`internal/graph/graph.go`)
- Scaffolded Redis utility package (`internal/redis/redis.go`)
- Scaffolded main.go for Collector and MCP Server
- Added project README
- Installed Kubernetes and Prometheus client dependencies
- Integrated Kubernetes client-go and informer factory in Collector
- Ran `go mod tidy` to resolve all dependencies

### 2025-05-23

- Added and started Kubernetes Service informer in the collector (cmd/collector/main.go)
- Ready to add event handlers for mesh graph updates and additional informers

**Next Steps:**  
- Service informer event handlers now update the in-memory mesh graph in the collector (add/update/delete)
- Pod informer and event handlers implemented; collector now tracks pod add/update/delete events
- Pod event handlers now extract "app" label to associate pods with services (mesh enrichment)
- Collector ready for HTTPRoute, GRPCRoute, AuthorizationPolicy CRD informer integration and Prometheus polling
- Prometheus integration scaffolded in collector; Prometheus metric polling now updates mesh.Edges with live traffic topology
- Collector is ready for HTTPRoute, GRPCRoute, AuthorizationPolicy CRD informer implementation (requires codegen or dynamic client)
- All core mesh graph and metric logic is complete
- Redis snapshot publishing implemented in the collector; mesh graph is now periodically published to Redis
- MCP server now builds and serves the GetMeshGraph gRPC endpoint
- MCP server hydrates mesh graph from Redis snapshot at startup and serves the gRPC API
- MCP server subscribes to mesh:delta for live updates and keeps mesh graph in sync
- ApplyAuthorizationPolicy mutation endpoint is now fully implemented in the MCP server: updates mesh graph and publishes deltas to Redis
- Collector now subscribes to mesh:delta and updates AuthPolicies in memory for reconciliation
- Collector now scaffolds periodic reconciliation of AuthPolicies to Kubernetes
- Collector now implements full dynamic client logic for reconciling AuthorizationPolicy CRs in Kubernetes, based on the mesh graph's AuthPolicies
- The system now supports end-to-end policy mutation, propagation, and reconciliation from gRPC API to live cluster state
- All major control plane flows are implemented and scaffolded for further testing and refinement
- Main.go files for both collector and MCP server have been cleaned up for clarity and maintainability
- Next: integration testing, error handling, and further enhancements
