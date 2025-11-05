.PHONY: help test lint fmt vet security check demo-before demo-after demo-verify

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "}; /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run all tests
	go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...

fmt: ## Run code formatting
	go fmt ./...

vet: ## Run static code analysis
	go vet ./...

lint: vet fmt ## Run linting/formatting tools
	golangci-lint run

security: ## Run security scans
	gosec ./...
	govulncheck ./...

check: fmt lint test security ## Run all checks: formatting, linting, tests, and security

demo-before: ## Run demo showing the bug before the fix
	@echo "=== Running demo: BEFORE FIX ==="
	@echo "This demonstrates the bug where credentials are ignored when HTTPClient is provided"
	@cd example/demo_before && go run main.go

demo-after: ## Run demo showing the fix working
	@echo "=== Running demo: AFTER FIX ==="
	@echo "This demonstrates that both credentials and custom HTTPClient are now honored"
	@cd example/demo_after && go run main.go

demo-verify: demo-after ## Verify the fix is working (alias for demo-after)
	@echo ""
	@echo "✓ Verification complete: Credentials + Custom HTTPClient fix is working!"
