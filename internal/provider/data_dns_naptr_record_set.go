// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"
)

var (
	_ datasource.DataSource              = (*dnsNAPTRRecordSetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dnsNAPTRRecordSetDataSource)(nil)
)

func NewDnsNAPTRRecordSetDataSource() datasource.DataSource {
	return &dnsNAPTRRecordSetDataSource{}
}

type dnsNAPTRRecordSetDataSource struct {
	client *DNSClient
}

func (d *dnsNAPTRRecordSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_naptr_record_set"
}

func (d *dnsNAPTRRecordSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get NAPTR records for a zone.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "Zone to look up.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "The name to look up. If not set, the zone is used.",
			},
			"order": schema.Int64Attribute{
				Computed:    true,
				Description: "The NAPTR order value.",
			},
			"preference": schema.Int64Attribute{
				Computed:    true,
				Description: "The NAPTR preference value.",
			},
			"flags": schema.StringAttribute{
				Computed:    true,
				Description: "The NAPTR flags string.",
			},
			"service": schema.StringAttribute{
				Computed:    true,
				Description: "The NAPTR service string.",
			},
			"regexp": schema.StringAttribute{
				Computed:    true,
				Description: "The NAPTR regular expression.",
			},
			"replacement": schema.StringAttribute{
				Computed:    true,
				Description: "The NAPTR replacement domain.",
			},
			"ttl": schema.Int64Attribute{
				Computed:    true,
				Description: "The TTL of the record.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the zone.",
			},
		},
	}
}

func (d *dnsNAPTRRecordSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*DNSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *DNSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *dnsNAPTRRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config naptrRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := config.Zone.ValueString()
	name := ""
	if !config.Name.IsNull() {
		name = config.Name.ValueString()
	}

	fqdn := zone
	if name != "" {
		fqdn = name + "." + zone
	}

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeNAPTR)

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up NAPTR records for %q: ", fqdn), err.Error())
		return
	}

	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up NAPTR records for %q: ", fqdn), "no NAPTR record found")
		return
	}

	record, ok := r.Answer[0].(*dns.NAPTR)
	if !ok {
		resp.Diagnostics.AddError("error parsing NAPTR record:", "unexpected record type received")
		return
	}

	config.Order = types.Int64Value(int64(record.Order))
	config.Preference = types.Int64Value(int64(record.Preference))
	config.Flags = types.StringValue(record.Flags)
	config.Service = types.StringValue(record.Service)
	config.Regexp = types.StringValue(record.Regexp)
	config.Replacement = types.StringValue(record.Replacement)
	config.TTL = types.Int64Value(int64(record.Hdr.Ttl))
	config.ID = types.StringValue(zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type naptrRecordSetConfig struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	Name        types.String `tfsdk:"name"`
	Order       types.Int64  `tfsdk:"order"`
	Preference  types.Int64  `tfsdk:"preference"`
	Flags       types.String `tfsdk:"flags"`
	Service     types.String `tfsdk:"service"`
	Regexp      types.String `tfsdk:"regexp"`
	Replacement types.String `tfsdk:"replacement"`
	TTL         types.Int64  `tfsdk:"ttl"`
}
