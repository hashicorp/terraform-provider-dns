// Copyright IBM Corp. 2016, 2025
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
	_ datasource.DataSource = (*dnsCNAMERecordSetDataSource)(nil)
)

func NewDnsCNAMERecordSetDataSource() datasource.DataSource {
	return &dnsCNAMERecordSetDataSource{}
}

type dnsCNAMERecordSetDataSource struct{}

func (d *dnsCNAMERecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_record_set"
}

func (d *dnsCNAMERecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS CNAME record set of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"cname": schema.StringAttribute{
				Computed:    true,
				Description: "A CNAME record associated with host.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host.",
			},
		},
	}
}

func (d *dnsCNAMERecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cnameRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	cname, err := net.LookupCNAME(host)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up CNAME records for %q: ", host), err.Error())
		return
	}

	config.CNAME = types.StringValue(cname)
	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type cnameRecordSetConfig struct {
	ID    types.String `tfsdk:"id"`
	Host  types.String `tfsdk:"host"`
	CNAME types.String `tfsdk:"cname"`
}
