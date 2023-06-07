// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsTXTRecordSetDataSource)(nil)
)

func NewDnsTXTRecordSetDataSource() datasource.DataSource {
	return &dnsTXTRecordSetDataSource{}
}

type dnsTXTRecordSetDataSource struct{}

func (d *dnsTXTRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_txt_record_set"
}

func (d *dnsTXTRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS TXT record set of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"record": schema.StringAttribute{
				Computed:    true,
				Description: "The first TXT record.",
			},
			"records": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of TXT records.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host.",
			},
		},
	}
}

func (d *dnsTXTRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config txtRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	records, err := net.LookupTXT(host)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up TXT records for %q: ", host), err.Error())
		return
	}

	if len(records) > 0 {
		config.Record = types.StringValue(records[0])
	} else {
		config.Record = types.StringNull()
	}

	var convertDiags diag.Diagnostics
	config.Records, convertDiags = types.ListValueFrom(ctx, config.Records.ElementType(ctx), records)
	if convertDiags.HasError() {
		resp.Diagnostics.Append(convertDiags...)
		return
	}

	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type txtRecordSetConfig struct {
	ID      types.String `tfsdk:"id"`
	Host    types.String `tfsdk:"host"`
	Record  types.String `tfsdk:"record"`
	Records types.List   `tfsdk:"records"`
}
