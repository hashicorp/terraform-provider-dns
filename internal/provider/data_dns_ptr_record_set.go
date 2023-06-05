// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsPTRRecordSetDataSource)(nil)
)

func NewDnsPTRRecordSetDataSource() datasource.DataSource {
	return &dnsPTRRecordSetDataSource{}
}

type dnsPTRRecordSetDataSource struct{}

func (d *dnsPTRRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ptr_record_set"
}

func (d *dnsPTRRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS PTR record set of the ip address.",
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				Required:    true,
				Description: "IP address to look up.",
			},
			"ptr": schema.StringAttribute{
				Computed:    true,
				Description: "A PTR record associated with `ip_address`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the IP address.",
			},
		},
	}
}

func (d *dnsPTRRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ptrRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipAddress := config.IPAddress.ValueString()
	names, err := net.LookupAddr(ipAddress)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up PTR records for %q: ", names), err.Error())
		return
	}
	if len(names) == 0 {
		resp.Diagnostics.AddError("DNS PTR record read error",
			fmt.Sprintf("error looking up PTR records for %q: no records found", ipAddress))

	}

	config.PTR = types.StringValue(names[0])
	config.ID = types.StringValue(ipAddress)
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type ptrRecordSetConfig struct {
	ID        types.String `tfsdk:"id"`
	IPAddress types.String `tfsdk:"ip_address"`
	PTR       types.String `tfsdk:"ptr"`
}
