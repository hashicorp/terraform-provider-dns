// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsCNAMERecursiveRecordSetDataSource)(nil)
)

func NewDnsCNAMERecursiveRecordSetDataSource() datasource.DataSource {
	return &dnsCNAMERecursiveRecordSetDataSource{}
}

type dnsCNAMERecursiveRecordSetDataSource struct{}

func (d *dnsCNAMERecursiveRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_recursive_cname_record_set"
}

func (d *dnsCNAMERecursiveRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS CNAME record set of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to recursively look up.",
			},
			"cnames": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Chained CNAME records associated with host.",
			},
			"last_cname": schema.StringAttribute{
				Computed:    true,
				Description: "The final CNAME at end of the chain.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host.",
			},
		},
	}
}

func (d *dnsCNAMERecursiveRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cnameRecursiveRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cnames := []string{}

	host := config.Host.ValueString()
	for cname, err := net.LookupCNAME(host); cname != host; cname, err = net.LookupCNAME(host) {
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("error looking up CNAME records for %q: ", host), err.Error())
			return
		}
		cnames = append(cnames, strings.Clone(cname))
		host = cname
	}

	config.CNAMES, _ = types.ListValueFrom(ctx, types.StringType, cnames)
	if len(cnames) > 0 {
		config.LastCNAME = types.StringValue(cnames[len(cnames)-1])
	}
	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type cnameRecursiveRecordSetConfig struct {
	ID        types.String `tfsdk:"id"`
	Host      types.String `tfsdk:"host"`
	CNAMES    types.List   `tfsdk:"cnames"`
	LastCNAME types.String `tfsdk:"last_cname"`
}
