GO_BIN?=$(shell pwd)/.bin
BINARY_DIR=bin
BINARY_NAME=terraform-provider-hetznerdns

.PHONY: build testacc test lint generate docs fmt tools

build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME)

testacc:
	TF_ACC=1 go test -v ./internal/provider -timeout 30m

test:
	go test -v ./... -timeout 30m

lint fmt:
	${GO_BIN}/golangci-lint run --fix ./...
	go tool tfproviderlintx -fix ./...

generate docs:
	go generate ./...

download:
	@echo Download go.mod dependencies
	@go mod download

tools:
	mkdir -p ${GO_BIN}
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b ${GO_BIN} latest
	GOBIN=${GO_BIN} go install tool
