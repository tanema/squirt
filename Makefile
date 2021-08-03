.PHONY: build
install: ## Build and install the squirt binary onto the $PATH
	@CGO_ENABLED=0 go install github.com/tanema/squirt/cmd/squirt;
lint: # Lint all packages
ifeq ("$(shell which golint 2>/dev/null)","")
	@go get -u golang.org/x/lint/golint
endif
	@golint -set_exit_status ./...
test: ## lint and test the code
	go test ./...
	go vet ./...
clean: ## Remove all build artifacts
	@rm -rf build && echo "project cleaned";
all: windows apple linux freebsd ## Build a release binary for all platforms
windows: ## Build release windows binaries
	@export GOOS=windows GOARCH=386 EXT=.exe; $(MAKE) build;
	@export GOOS=windows GOARCH=amd64 EXT=.exe; $(MAKE) build;
apple: ## Build release apple binaries into the build folder
	@export GOOS=darwin GOARCH=386; $(MAKE) build;
	@export GOOS=darwin GOARCH=amd64; $(MAKE) build;
linux: ## Build release linux binaries into the build folder
	@export GOOS=linux GOARCH=386; $(MAKE) build;
	@export GOOS=linux GOARCH=amd64; $(MAKE) build;
freebsd: ## Build release freebasd binaries into the build folder
	@export GOOS=freebsd GOARCH=386; $(MAKE) build;
	@export GOOS=freebsd GOARCH=amd64; $(MAKE) build;
build:
	@mkdir -p build/${GOOS}-${GOARCH} && \
    echo "[${GOOS}-${GOARCH}] build started" && \
		CGO_ENABLED=0 go build \
			-ldflags="-s -w" \
			-o build/${GOOS}-${GOARCH}/squirt${EXT} \
			github.com/tanema/squirt/cmd/squirt && \
		echo "[${GOOS}-${GOARCH}] build complete";
help:
	@grep -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

