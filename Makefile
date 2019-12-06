#### Environment ####

# Enforce usage of the Go modules system.
export GO111MODULE := on

# Determine where `go get` will install binaries to.
GOBIN := $(HOME)/go/bin
ifdef GOPATH
	GOBIN := $(GOPATH)/bin
endif

# Current operating system name.
PLATFORM := $(shell uname -s)

# All target for when make is run on its own.
.PHONY: all
all: test lint

#### Binary Dependencies ####

# Install binary for goimports.
goimports := $(GOBIN)/goimports
$(goimports):
	@cd /tmp && go get -u golang.org/x/tools/cmd/goimports

# Install binary for golangci-lint.
golangci-lint := $(GOBIN)/golangci-lint
$(golangci-lint):
	@./scripts/install-golangci-lint $(golangci-lint)

# Install binary for goreleaser.
goreleaser := $(GOBIN)/goreleaser
$(goreleaser):
	@./scripts/install-goreleaser $(goreleaser)

# Install binary for upx.
ifeq "$(PLATFORM)" "Linux"
upx := $(GOBIN)/upx
else
upx := /usr/local/bin/upx
endif
$(upx):
	@./scripts/install-upx $(upx)

#### Linting ####

# Run code linters.
.PHONY: lint
lint: $(golangci-lint) style
	$(golangci-lint) run

# Run code formatters. Unformatted code will fail in CircleCI.
.PHONY: style
style: $(goimports)
ifdef GITHUB_ACTIONS
	$(goimports) -l .
else
	$(goimports) -l -w .
endif

#### Testing ####

# Run Go tests and generate a JUnit XML style test report for ingestion by CircleCI.
.PHONY: test
test: $(go-junit-report)
	@go test -v -race -cover ./...

#### Release ####

.PHONY: compress
compress: $(upx)
	$(upx) --best --ultra-brute dist/retry_*/retry*

.PHONY: release
release: $(goreleaser)
ifdef GITHUB_ACTIONS
	$(goreleaser) release
else
	$(goreleaser) --rm-dist --skip-publish --skip-validate --snapshot
endif
