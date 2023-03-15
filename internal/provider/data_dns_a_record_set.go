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
	_ datasource.DataSource = (*dnsARecordSetDataSource)(nil)
)

func NewDnsARecordSetDataSource() datasource.DataSource {
	return &dnsARecordSetDataSource{}
}

type dnsARecordSetDataSource struct{}

func (d dnsARecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_a_record_set"
}

func (d dnsARecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "",
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

func (d dnsARecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config aRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	a, _, err := lookupIP(host)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up A records for %q: ", host), err.Error())
		return
	}
	sort.Strings(a)

	var convertDiags diag.Diagnostics
	config.Addrs, convertDiags = types.ListValueFrom(ctx, config.Addrs.ElementType(ctx), a)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type aRecordSetConfig struct {
	ID    types.String `tfsdk:"id"`
	Host  types.String `tfsdk:"host"`
	Addrs types.List   `tfsdk:"addrs"`
}
