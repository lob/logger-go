GOTOOLS := \
	github.com/alecthomas/gometalinter \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \

GOPACKAGES := $(go list ./...)

COVERAGE_PROFILE := build/coverage.out

.PHONY: dev-setup
dev-setup: ## Install all the build and lint dependencies
	@echo "---> Setup"
	go get -u $(GOTOOLS)
	gometalinter --install
	dep ensure

.PHONY: release
release: ## Build binary
	@echo "---> Building: (darwin, amd64)"
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-w -s" -race -o build/logger-go main.go

.PHONY: test
test: ## Run all the tests
	@echo "---> Testing"
	mkdir -p build
	go test -cover -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) -v -timeout=30s $(GOPACKAGES)

.PHONY: cover
cover: test ## Run all the tests and outputs the coverage report
	@echo "---> Coverage"
	go tool cover -html=$(COVERAGE_PROFILE) -o build/coverage.html

.PHONY: enforce ## Enforce code coverage
enforce: cover
	@echo "---> Enforcing coverage"
	./scripts/coverage.sh $(COVERAGE_PROFILE)

.PHONY: lint
lint: ## Run all the linters
	@echo "---> Linting..."
	gometalinter

.PHONY: clean
clean: ## Remove temporary files
	@echo "---> Cleaning"
	go clean
	rm -r build

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## default command
.DEFAULT_GOAL := release
