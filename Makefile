# Core Module Makefile (Go)
BINARY_NAME=kcloud-scheduler
DOCKER_IMAGE=kcloud-opt/core
VERSION?=0.1.0
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

.PHONY: help build test clean run docker-build

help: ## Show help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(GOBIN)/$(BINARY_NAME) ./cmd/scheduler

run: ## Run the application
	@go run ./cmd/scheduler

test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean: ## Clean build files
	@rm -rf $(GOBIN)
	@rm -f coverage.out

deps: ## Download dependencies
	@go mod download
	@go mod tidy

fmt: ## Format code
	@go fmt ./...
	@gofmt -w .

lint: ## Run linter
	@golangci-lint run

docker-build: ## Build Docker image
	@docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest

docker-push: ## Push Docker image
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

generate: ## Generate code (CRDs, clients, etc.)
	@controller-gen crd paths="./api/..." output:crd:artifacts:config=config/crd/bases
	@controller-gen object paths="./api/..."

install-tools: ## Install development tools
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Kubernetes deployment
deploy: ## Deploy to Kubernetes
	@kubectl apply -f config/crd/
	@kubectl apply -f config/rbac/
	@kubectl apply -f config/deployment/

undeploy: ## Remove from Kubernetes
	@kubectl delete -f config/deployment/
	@kubectl delete -f config/rbac/
	@kubectl delete -f config/crd/