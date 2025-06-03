.PHONY: all build-mcp-server build-collector build-all go-build push-mcp-server push-collector push-all scan

all: build-all

build-mcp-server:
	docker build -f Dockerfile.mcp-server -t ghcr.io/eli-nomasec/mcp-server:latest .

build-collector:
	docker build -f Dockerfile.collector -t ghcr.io/eli-nomasec/mcp-collector:latest .

build-all: build-mcp-server build-collector
	@echo "Built all images"

go-build:
	go build -o bin/mcp-server ./cmd/mcp-server
	go build -o bin/collector ./cmd/collector

push-mcp-server:
	docker push ghcr.io/eli-nomasec/mcp-server:latest

push-collector:
	docker push ghcr.io/eli-nomasec/mcp-collector:latest

push-all: push-mcp-server push-collector
	@echo "Pushed all images"

scan:
	trivy image ghcr.io/eli-nomasec/mcp-server:latest
	trivy image ghcr.io/eli-nomasec/mcp-collector:latest

k8s-deploy:
	kubectl config use-context colima-arm
	helm upgrade --install mcp-demo ./helm/mcp --namespace mcp-demo --create-namespace

k8s-clean:
	kubectl config use-context colima-arm
	helm uninstall mcp-demo --namespace mcp-demo || true
	kubectl delete namespace mcp-demo --ignore-not-found

test-integration:
	go test ./tests/integration/...

valkey-up:
	docker run -d --rm --name valkey-local -p 6379:6379 valkey/valkey:latest

valkey-down:
	docker stop valkey-local || true
