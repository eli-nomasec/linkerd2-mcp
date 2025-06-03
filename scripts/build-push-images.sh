#!/bin/bash
set -euo pipefail

ORG=eli-nomasec
TAG=latest

# Build MCP server image
docker build -t ghcr.io/$ORG/mcp-server:$TAG -f Dockerfile.mcp-server .

# Build collector image
docker build -t ghcr.io/$ORG/mcp-collector:$TAG -f Dockerfile.collector .

# Push images to GitHub Container Registry
docker push ghcr.io/$ORG/mcp-server:$TAG
docker push ghcr.io/$ORG/mcp-collector:$TAG

echo "Images pushed to ghcr.io/$ORG"
