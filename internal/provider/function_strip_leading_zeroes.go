// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-nettypes/iptypes"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewStripLeadingZeroesFunction() function.Function {
	return &stripLeadingZeroesFunction{}
}

type stripLeadingZeroesFunction struct {
}

func (f stripLeadingZeroesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Parameters: []function.Parameter{
			function.SetParameter{
				Name:        "addresses",
				ElementType: types.StringType,
			},
		},
		Return: function.SetReturn{
			ElementType: iptypes.IPAddressType{},
		},
	}
}

func (f stripLeadingZeroesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "strip_leading_zeroes"
}

func (f stripLeadingZeroesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var ipAddresses []string
	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &ipAddresses))
	if resp.Error != nil {
		return
	}

	parsedIPAddresses := make([]iptypes.IPAddress, len(ipAddresses))
	for i, ipAddress := range ipAddresses {
		// Note: this implementation isn't fully correct, I'm just being lazy :P
		parsedIPAddresses[i] = iptypes.NewIPAddressValue(stripLeadingZerosIPv6(stripLeadingZeros(ipAddress)))
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, parsedIPAddresses))
}
