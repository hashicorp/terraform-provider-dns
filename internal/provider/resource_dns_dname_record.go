// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"

	"github.com/hashicorp/terraform-provider-dns/internal/validators/dnsvalidator"
)

var (
	_ resource.Resource                = (*dnsDNAMERecordResource)(nil)
	_ resource.ResourceWithImportState = (*dnsDNAMERecordResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsDNAMERecordResource)(nil)
)

func NewDnsDNAMERecordResource() resource.Resource {
	return &dnsDNAMERecordResource{}
}

type dnsDNAMERecordResource struct {
	client *DNSClient
}

func (d *dnsDNAMERecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dname_record"
}

func (d *dnsDNAMERecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a DNAME type DNS record.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
				},
				Description: "DNS zone the record belongs to. It must be an FQDN, that is, include the trailing dot.",
			},
			"name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsRecordNameValid(),
				},
				Description: "The name of the record. The `zone` argument will be appended to this value to create " +
					"the full record path.",
			},
			"target": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The DNAME target (alias) for the record.",
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3600),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The TTL of the record set. Defaults to `3600`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record.",
			},
		},
	}
}

func (d *dnsDNAMERecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsDNAMERecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnameRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: plan.Name.ValueString(),
		Zone: plan.Zone.ValueString(),
	}
	recFqdn := resourceFQDN_framework(config)
	plan.ID = types.StringValue(recFqdn)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	rrStrInsert := fmt.Sprintf("%s %d DNAME %s", recFqdn, plan.TTL.ValueInt64(), plan.Target.ValueString())

	rrInsert, err := dns.NewRR(rrStrInsert)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrInsert), err.Error())
		return
	}
	msg.Insert([]dns.RR{rrInsert})

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeDNAME)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		dnameRecord, ok := record.(*dns.DNAME)
		if !ok {
			resp.Diagnostics.AddError("Error querying DNS record:", "didn't get a DNAME record")
			return
		}

		plan.Target = types.StringValue(dnameRecord.Target)
		plan.TTL = types.Int64Value(int64(dnameRecord.Hdr.Ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsDNAMERecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnameRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeDNAME)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		dnameRecord, ok := record.(*dns.DNAME)
		if !ok {
			resp.Diagnostics.AddError("Error querying DNS record:", "didn't get a DNAME record")
			return
		}

		state.Target = types.StringValue(dnameRecord.Target)
		state.TTL = types.Int64Value(int64(dnameRecord.Hdr.Ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsDNAMERecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state dnameRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: plan.Name.ValueString(),
		Zone: plan.Zone.ValueString(),
	}
	recFqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	if !plan.Target.Equal(state.Target) {

		rrStrRemove := fmt.Sprintf("%s %d DNAME %s", recFqdn, plan.TTL.ValueInt64(), state.Target.ValueString())

		rrRemove, err := dns.NewRR(rrStrRemove)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrRemove), err.Error())
			return
		}

		rrStrInsert := fmt.Sprintf("%s %d DNAME %s", recFqdn, plan.TTL.ValueInt64(), plan.Target.ValueString())

		rrInsert, err := dns.NewRR(rrStrInsert)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrInsert), err.Error())
			return
		}

		msg.Remove([]dns.RR{rrRemove})
		msg.Insert([]dns.RR{rrInsert})

		r, err := exchange(msg, true, d.client)
		if err != nil {
			resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
			return
		}
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeDNAME)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		dnameRecord, ok := record.(*dns.DNAME)
		if !ok {
			resp.Diagnostics.AddError("Error querying DNS record:", "didn't get a DNAME record")
			return
		}

		state.Target = types.StringValue(dnameRecord.Target)
		state.TTL = types.Int64Value(int64(dnameRecord.Hdr.Ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsDNAMERecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnameRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeDNAME)...)
}

func (d *dnsDNAMERecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var state dnameRecordResourceModel

	config, diags := resourceDnsImport_framework(req.ID, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	state.ID = types.StringValue(req.ID)
	state.Name = types.StringValue(config.Name)
	state.Zone = types.StringValue(config.Zone)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type dnameRecordResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Zone   types.String `tfsdk:"zone"`
	Name   types.String `tfsdk:"name"`
	Target types.String `tfsdk:"target"`
	TTL    types.Int64  `tfsdk:"ttl"`
}
