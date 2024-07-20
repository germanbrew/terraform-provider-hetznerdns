// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build tools

package tools

import (
	// Documentation generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.4"
	// Linting
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1"
)
