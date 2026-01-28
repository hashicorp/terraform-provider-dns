// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"
)

var (
	_ datasource.DataSource              = (*dnsAAAARecordSetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dnsAAAARecordSetDataSource)(nil)
)

func NewDnsAAAARecordSetDataSource() datasource.DataSource {
	return &dnsAAAARecordSetDataSource{}
}

type dnsAAAARecordSetDataSource struct {
	client *DNSClient
}

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
			"use_update_server": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to use the configured update DNS server",
			},
			"addrs": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "A list of IP addresses. IP addresses are always sorted to avoid constant changing plans.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the host.",
			},
		},
	}
}

func (d *dnsAAAARecordSetDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*DNSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *DNSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *dnsAAAARecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config aRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.UseUpdateServer.ValueBool() && d.client == nil {
		resp.Diagnostics.AddError("use_update_server enabled, but no update server configured", "If you set use_update_server to true, an update server needs to be configured for the provider")
		return
	}

	host := config.Host.ValueString()

	answers := []string{}
	if !config.UseUpdateServer.ValueBool() || d.client == nil {
		var err error
		_, answers, err = lookupIP(host)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("error looking up AAAA records for %q: ", host), err.Error())
			return
		}
	} else {
		records, diags := resourceDnsRead_framework_flags(dnsConfig{Name: host}, d.client, dns.TypeAAAA, dns.MsgHdr{RecursionDesired: true})
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		for _, record := range records {
			addr, _, err := getAAAAVal(record)
			if err != nil {
				resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
				return
			}

			answers = append(answers, addr)
		}
	}
	sort.Strings(answers)

	var convertDiags diag.Diagnostics
	config.Addrs, convertDiags = types.ListValueFrom(ctx, config.Addrs.ElementType(ctx), answers)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = config.Host
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}
