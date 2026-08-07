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
	_ datasource.DataSource              = (*dnsDNAMERecordSetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dnsDNAMERecordSetDataSource)(nil)
)

func NewDnsDNAMERecordSetDataSource() datasource.DataSource {
	return &dnsDNAMERecordSetDataSource{}
}

type dnsDNAMERecordSetDataSource struct {
	client *DNSClient
}

func (d *dnsDNAMERecordSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dname_record_set"
}

func (d *dnsDNAMERecordSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNAME record set of the host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"target": schema.StringAttribute{
				Computed:    true,
				Description: "The DNAME target (alias) for the host.",
			},
			"ttl": schema.Int64Attribute{
				Computed:    true,
				Description: "The TTL of the record.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host.",
			},
		},
	}
}

func (d *dnsDNAMERecordSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsDNAMERecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config dnameRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()

	msg := new(dns.Msg)
	msg.SetQuestion(host, dns.TypeDNAME)

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up DNAME records for %q: ", host), err.Error())
		return
	}

	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up DNAME records for %q: ", host), "no DNAME record found")
		return
	}

	record, ok := r.Answer[0].(*dns.DNAME)
	if !ok {
		resp.Diagnostics.AddError("error parsing DNAME record:", "unexpected record type received")
		return
	}

	config.Target = types.StringValue(record.Target)
	config.TTL = types.Int64Value(int64(record.Hdr.Ttl))
	config.ID = types.StringValue(host)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type dnameRecordSetConfig struct {
	ID     types.String `tfsdk:"id"`
	Host   types.String `tfsdk:"host"`
	Target types.String `tfsdk:"target"`
	TTL    types.Int64  `tfsdk:"ttl"`
}
