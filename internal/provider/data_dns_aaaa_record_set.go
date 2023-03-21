package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsAAAARecordSetDataSource)(nil)
)

func NewDnsAAAARecordSetDataSource() datasource.DataSource {
	return &dnsAAAARecordSetDataSource{}
}

type dnsAAAARecordSetDataSource struct{}

func (d *dnsAAAARecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_aaaa_record_set"
}

func (d *dnsAAAARecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS AAAA records of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"addrs": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP addresses. IP addresses are always sorted to avoid constant changing plans.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host",
			},
		},
	}
}

func (d *dnsAAAARecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config aRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	_, aaaa, err := lookupIP(host)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up AAAA records for %q: ", host), err.Error())
		return
	}
	sort.Strings(aaaa)

	var convertDiags diag.Diagnostics
	config.Addrs, convertDiags = types.ListValueFrom(ctx, config.Addrs.ElementType(ctx), aaaa)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
