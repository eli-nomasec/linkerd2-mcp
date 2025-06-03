#!/bin/bash
set -euo pipefail

# Ensure we are using the colima-arm context
CTX=$(kubectl config current-context)
if [[ "$CTX" != "colima-arm" ]]; then
  echo "ERROR: Current context is '$CTX', but 'colima-arm' is required."
  exit 1
fi

# Assume Linkerd and Linkerd Viz are already installed

# Deploy MCP stack with Helm
helm install mcp ./helm/mcp --namespace mcp --create-namespace --set redis.enabled=true

# Wait for MCP server pod to be ready
kubectl -n mcp rollout status deployment/mcp-server

# Port-forward MCP server
kubectl -n mcp port-forward svc/mcp-server 10900:10900 &
PF_PID=$!
sleep 5

# Test GetMeshGraph
echo "Testing GetMeshGraph..."
grpcurl -plaintext localhost:10900 mcp.v1.MeshContext/GetMeshGraph

# Test ApplyAuthorizationPolicy
echo "Testing ApplyAuthorizationPolicy..."
grpcurl -plaintext -d '{"namespace":"default","name":"allow-foo","json_spec":"{\"foo\": \"bar\"}"}' localhost:10900 mcp.v1.MeshContext/ApplyAuthorizationPolicy

# Verify AuthorizationPolicy CR
echo "Verifying AuthorizationPolicy CRs in the cluster..."
kubectl get authorizationpolicies.policy.linkerd.io -A

# Cleanup
kill $PF_PID
echo "E2E test complete."
