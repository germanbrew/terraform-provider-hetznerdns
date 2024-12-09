// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"golang.org/x/net/idna"
)

var _ function.Function = idnaFunction{}

func NewIdnaFunction() function.Function {
	return idnaFunction{}
}

type idnaFunction struct{}

func (r idnaFunction) Metadata(_ context.Context, _ function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "idna"
}

func (r idnaFunction) Definition(_ context.Context, _ function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "idna function",
		Description: "idna converts a IDN domain or domain label to its ASCII form (Punnycode).",
		MarkdownDescription: "idna converts a [IDN](https://en.wikipedia.org/wiki/Internationalized_domain_name) domain or domain label to its ASCII form (Punnycode). " +
			"For example, `provider::hetznerdns::idna(\"bücher.example.com\")` is " +
			"\"xn--bcher-kva.example.com\", and `provider::hetznerdns::idna(\"golang\")` is \"golang\". " +
			"If an error is encountered it will return an error and a (partially) processed result.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "domain",
				MarkdownDescription: "domain to convert",
			},
		},
		Return: function.StringReturn{},
	}
}

func (r idnaFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var (
		domain string
		err    error
	)

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &domain))

	if resp.Error != nil {
		return
	}

	domain, err = idna.New().ToASCII(domain)
	if err != nil {
		resp.Error = function.NewFuncError("failed to convert domain to ASCII: " + err.Error())

		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, domain))
}
