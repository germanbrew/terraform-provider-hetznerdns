BINARY_DIR=bin
BINARY_NAME=terraform-provider-hetznerdns

.PHONY: build testacc test lint generate fmt

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

.PHONY: update
update:  ## Run dependency updates
	@go get -u ./...
	@go mod tidy
	@go -C tools get -u
	@go -C tools mod tidy

generate:
	go generate ./...

fmt: install-tools
	@-go fmt ./...
	@-tools/bin/gci write .
	@-tools/bin/gofumpt -l -w .
	@-tools/bin/goimports -l -w .
	@-tools/bin/wsl -strict-append -test=true -fix ./...
	@-tools/bin/perfsprint -fix ./...
	@-tools/bin/tfproviderlintx -fix ./...
	@tools/bin/golangci-lint run ./... --fix


# In order to help reduce toil related to managing tooling for the open telemetry collector
# this section of the makefile looks at only requiring command definitions to be defined
# as part of $(TOOLS_MOD_DIR)/tools.go, following the existing practice.
# Modifying the tools' `go.mod` file will trigger a rebuild of the tools to help
# ensure that all contributors are using the most recent version to make builds repeatable everywhere.
TOOLS_MOD_DIR    := tools
TOOLS_MOD_REGEX  := "\s+_\s+\".*\""
TOOLS_PKG_NAMES  := $(shell grep -E $(TOOLS_MOD_REGEX) < $(TOOLS_MOD_DIR)/tools.go | tr -d " _\"")
TOOLS_BIN_DIR    := bin
TOOLS_BIN_NAMES  := $(addprefix $(TOOLS_BIN_DIR)/, $(notdir $(TOOLS_PKG_NAMES)))

.PHONY: install-tools
install-tools: $(TOOLS_BIN_NAMES)

$(TOOLS_BIN_DIR):
	@mkdir -p $@

$(TOOLS_BIN_NAMES): $(TOOLS_BIN_DIR) $(TOOLS_MOD_DIR)/go.mod
	go build -C $(TOOLS_MOD_DIR) -o $@ -trimpath $(filter %/$(notdir $@),$(TOOLS_PKG_NAMES))
