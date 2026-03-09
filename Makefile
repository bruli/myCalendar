SHELL := /bin/bash

# ⚙️ Configuration
APP             ?= mycalendar

DOCKERFILE ?= Dockerfile

IMAGE_REG  ?= ghcr.io/bruli
IMAGE_NAME := $(IMAGE_REG)/$(APP)
VERSION    ?= 0.1.4
CURRENT_IMAGE := $(IMAGE_NAME):$(VERSION)

GOLANGCI_LINT_VERSION ?= v2.10.0

# Default goal
.DEFAULT_GOAL := help


# ────────────────────────────────────────────────────────────────
# 🧹 Code quality: format, lint, tests
# ────────────────────────────────────────────────────────────────
.PHONY: fmt
fmt:
	@set -euo pipefail; \
	echo "👉 Formatting code with gofumpt..."; \
	go tool gofumpt -w .

.PHONY: security
security:
	@set -euo pipefail; \
	echo "👉 Check security"; \
	go tool govulncheck ./...

.PHONY: install-lint
install-lint:
	@set -euo pipefail; \
    echo "🔧 Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
    	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: lint
lint: install-lint
	@set -euo pipefail; \
	echo "🚀 Executing golangci-lint..."; \
    golangci-lint run ./...

.PHONY: test
test:
	@set -euo pipefail; \
	echo "🧪 Running unit tests (race, JSON → tparse)..."; \
	go test -race ./... -json -cover -coverprofile=coverage.out| go tool tparse -all

.PHONY: clean
clean:
	@set -euo pipefail; \
	echo "🧹 Cleaning local artifacts..."; \
	rm -rf bin dist coverage .*cache || true; \
	go clean -testcache

.PHONY: check
check: fmt lint security test
	echo "✅ Format, linter and tests success."

.PHONY: docker-login
docker-login:
	echo "🔐 Logging into Docker registry...";
	echo "$$CR_PAT" | docker login ghcr.io -u bruli --password-stdin

.PHONY: docker-push-image
docker-push-image: docker-login
	echo "🐳 Building and pushing Docker image $(CURRENT_IMAGE) ...";
	docker buildx build \
		--platform linux/arm64 \
		-t $(CURRENT_IMAGE) \
		-f $(DOCKERFILE) \
		--push \
		.
	 echo "✅ Image $(CURRENT_IMAGE) pushed successfully."

# ────────────────────────────────────────────────────────────────
# ℹ️ Help
# ────────────────────────────────────────────────────────────────
help:
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:' Makefile | awk -F':' '{print "  - " $$1}'
