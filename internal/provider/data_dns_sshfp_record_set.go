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
	_ datasource.DataSource              = (*dnsSSHFPRecordSetDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dnsSSHFPRecordSetDataSource)(nil)
)

func NewDnsSSHFPRecordSetDataSource() datasource.DataSource {
	return &dnsSSHFPRecordSetDataSource{}
}

type dnsSSHFPRecordSetDataSource struct {
	client *DNSClient
}

func (d *dnsSSHFPRecordSetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshfp_record_set"
}

func (d *dnsSSHFPRecordSetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get SSHFP records of a host.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Required:    true,
				Description: "Host to look up.",
			},
			"algorithm": schema.Int64Attribute{
				Computed:    true,
				Description: "The SSHFP algorithm number.",
			},
			"fingerprint_type": schema.Int64Attribute{
				Computed:    true,
				Description: "The SSHFP fingerprint type.",
			},
			"fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The SSHFP fingerprint.",
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

func (d *dnsSSHFPRecordSetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dnsSSHFPRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config sshfpRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := config.Host.ValueString()

	msg := new(dns.Msg)
	msg.SetQuestion(host, dns.TypeSSHFP)

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up SSHFP records for %q: ", host), err.Error())
		return
	}

	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up SSHFP records for %q: ", host), "no SSHFP record found")
		return
	}

	record, ok := r.Answer[0].(*dns.SSHFP)
	if !ok {
		resp.Diagnostics.AddError("error parsing SSHFP record:", "unexpected record type received")
		return
	}

	config.Algorithm = types.Int64Value(int64(record.Algorithm))
	config.FingerprintType = types.Int64Value(int64(record.Type))
	config.Fingerprint = types.StringValue(record.FingerPrint)
	config.TTL = types.Int64Value(int64(record.Hdr.Ttl))
	config.ID = types.StringValue(host)

	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type sshfpRecordSetConfig struct {
	ID              types.String `tfsdk:"id"`
	Host            types.String `tfsdk:"host"`
	Algorithm       types.Int64  `tfsdk:"algorithm"`
	FingerprintType types.Int64  `tfsdk:"fingerprint_type"`
	Fingerprint     types.String `tfsdk:"fingerprint"`
	TTL             types.Int64  `tfsdk:"ttl"`
}
