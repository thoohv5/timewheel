include Makefile.common

VERSION=$(shell git describe --tags --always)

.PHONY: generate
# generate
generate:
	@echo ">> generating code"
	@$(GO) generate ./...
	@$(MAKE) format

.PHONY: build
# build
build:
	mkdir -p bin/ && $(GO) build -ldflags "-X main.Version=$(VERSION)" -o bin/ ./...


.PHONY: run
# run
run:
	$(GO) run $(PREFIX)/cmd/server --conf $(PREFIX)/configs

