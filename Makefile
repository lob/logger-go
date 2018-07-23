GOTOOLS := \
	github.com/alecthomas/gometalinter \
	github.com/golang/dep/cmd/dep \
	golang.org/x/tools/cmd/cover \

GOPACKAGES := $(go list ./...)

COVERAGE_PROFILE := build/coverage.out

.PHONY: setup
setup: ## Installs all development dependencies
	@echo "---> Setting up..."
	@go get -u $(GOTOOLS) && gometalinter --install && echo "Successful!"

.PHONY: deps
deps: ## Ensures all Go dependencies are in sync
	@echo "---> Ensuring deps are in sync..."
	@dep ensure && echo "Successful!"

.PHONY: test
test: ## Runs all the tests and outputs the coverage report
	@echo "---> Testing..."
	@mkdir -p build && \
	 RELEASE=test12345 go test -cover -covermode=atomic -coverprofile=$(COVERAGE_PROFILE) -v -timeout=30s $(GOPACKAGES) && \
	 go tool cover -html=build/coverage.out -o build/coverage.html

.PHONY: enforce ## Enforces code coverage
enforce: test
	@echo "---> Enforcing coverage..."
	@./scripts/coverage.sh $(COVERAGE_PROFILE)

.PHONY: lint
lint: ## Runs all linters
	@echo "---> Linting..."
	@gometalinter && echo "Successful!"

.PHONY: clean
clean: ## Removes Go temporary build files build directory
	@echo "---> Cleaning..."
	@go clean && rm -rf build && echo "Successful!"

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## default command
.DEFAULT_GOAL := help
