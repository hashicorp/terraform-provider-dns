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
	_ datasource.DataSource              = (*dnsSOARecordSetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dnsSOARecordSetDataSource)(nil)
)

func NewDnsSOARecordSetDataSource() datasource.DataSource {
	return &dnsSOARecordSetDataSource{}
}

type dnsSOARecordSetDataSource struct {
	client *DNSClient
}

func (d *dnsSOARecordSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_soa_record_set"
}

func (d *dnsSOARecordSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get the SOA record of a DNS zone.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required:    true,
				Description: "Zone to look up, must be an FQDN ending with a dot.",
			},
			"mname": schema.StringAttribute{
				Computed:    true,
				Description: "The primary master name for the zone.",
			},
			"rname": schema.StringAttribute{
				Computed:    true,
				Description: "The email address of the administrator responsible for the zone.",
			},
			"serial": schema.Int64Attribute{
				Computed:    true,
				Description: "The serial number of the zone.",
			},
			"refresh": schema.Int64Attribute{
				Computed:    true,
				Description: "The refresh interval in seconds.",
			},
			"retry": schema.Int64Attribute{
				Computed:    true,
				Description: "The retry interval in seconds.",
			},
			"expire": schema.Int64Attribute{
				Computed:    true,
				Description: "The expire time in seconds.",
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

func (d *dnsSOARecordSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsSOARecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config soaRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := config.Zone.ValueString()

	msg := new(dns.Msg)
	msg.SetQuestion(zone, dns.TypeSOA)

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up SOA record for %q: ", zone), err.Error())
		return
	}

	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up SOA record for %q: ", zone), "no SOA record found")
		return
	}

	record, ok := r.Answer[0].(*dns.SOA)
	if !ok {
		resp.Diagnostics.AddError("error parsing SOA record:", "unexpected record type received")
		return
	}

	config.MName = types.StringValue(record.Ns)
	config.RName = types.StringValue(record.Mbox)
	config.Serial = types.Int64Value(int64(record.Serial))
	config.Refresh = types.Int64Value(int64(record.Refresh))
	config.Retry = types.Int64Value(int64(record.Retry))
	config.Expire = types.Int64Value(int64(record.Expire))
	config.TTL = types.Int64Value(int64(record.Hdr.Ttl))
	config.ID = types.StringValue(zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type soaRecordSetConfig struct {
	ID      types.String `tfsdk:"id"`
	Zone    types.String `tfsdk:"zone"`
	MName   types.String `tfsdk:"mname"`
	RName   types.String `tfsdk:"rname"`
	Serial  types.Int64  `tfsdk:"serial"`
	Refresh types.Int64  `tfsdk:"refresh"`
	Retry   types.Int64  `tfsdk:"retry"`
	Expire  types.Int64  `tfsdk:"expire"`
	TTL     types.Int64  `tfsdk:"ttl"`
}
