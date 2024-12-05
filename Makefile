BINARY_DIR=bin
BINARY_NAME=terraform-provider-hetznerdns

.PHONY: build testacc test lint generate docs fmt

build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME)

testacc:
	TF_ACC=1 go test -v ./internal/provider -timeout 30m

test:
	go test -v ./... -timeout 30m

lint:
	golangci-lint run ./...
	go run github.com/bflad/tfproviderlint/cmd/tfproviderlintx@latest ./...

generate docs:
	go generate ./...

fmt:
	go fmt ./...
	-go run mvdan.cc/gofumpt@latest -l -w .
	-go run golang.org/x/tools/cmd/goimports@latest -l -w .
	-go run github.com/bombsimon/wsl/v4/cmd...@latest -strict-append -test=true -fix ./...
	-go run github.com/catenacyber/perfsprint@latest -fix ./...
	-go run github.com/bflad/tfproviderlint/cmd/tfproviderlintx@latest -fix ./...

download:
	@echo Download go.mod dependencies
	@go mod download

install-devtools: download
	@echo Installing tools from tools.go
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
