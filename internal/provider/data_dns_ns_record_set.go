package provider

import (
	"context"
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsNSRecordSetDataSource)(nil)
)

func NewDnsNSRecordSetDataSource() datasource.DataSource {
	return &dnsNSRecordSetDataSource{}
}

type dnsNSRecordSetDataSource struct{}

func (d *dnsNSRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ns_record_set"
}

func (d *dnsNSRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS NS records of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"nameservers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A list of nameservers. Nameservers are always sorted to avoid constant changing plans.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the domain",
			},
		},
	}
}

func (d *dnsNSRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config nsRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()
	nsRecords, err := net.LookupNS(host)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up NS records for %q: ", host), err.Error())
		return
	}

	nameservers := make([]string, len(nsRecords))
	for i, record := range nsRecords {
		nameservers[i] = record.Host
	}
	sort.Strings(nameservers)

	var convertDiags diag.Diagnostics
	config.Nameservers, convertDiags = types.ListValueFrom(ctx, config.Nameservers.ElementType(ctx), nameservers)
	if convertDiags.HasError() {
		resp.Diagnostics.Append(convertDiags...)
		return
	}

	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type nsRecordSetConfig struct {
	ID          types.String `tfsdk:"id"`
	Host        types.String `tfsdk:"host"`
	Nameservers types.List   `tfsdk:"nameservers"`
}
