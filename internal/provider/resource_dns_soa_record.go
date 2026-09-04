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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-provider-dns/internal/validators/dnsvalidator"
)

var (
	_ resource.Resource              = (*dnsSOARecordResource)(nil)
	_ resource.ResourceWithConfigure = (*dnsSOARecordResource)(nil)
)

func NewDnsSOARecordResource() resource.Resource {
	return &dnsSOARecordResource{}
}

type dnsSOARecordResource struct {
	client *DNSClient
}

func (d *dnsSOARecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_soa_record"
}

func (d *dnsSOARecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SOA DNS record. This resource does not create or delete the SOA record as it " +
			"always exists for a zone. It only reads and updates the SOA record fields.",
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
			"mname": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The primary master name for the zone.",
			},
			"rname": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The email address of the administrator responsible for the zone.",
			},
			"refresh": schema.Int64Attribute{
				Required:    true,
				Description: "The refresh interval in seconds.",
			},
			"retry": schema.Int64Attribute{
				Required:    true,
				Description: "The retry interval in seconds.",
			},
			"expire": schema.Int64Attribute{
				Required:    true,
				Description: "The expire time in seconds.",
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3600),
				Description: "The TTL of the record. Defaults to `3600`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the zone.",
			},
		},
	}
}

func (d *dnsSOARecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create is not supported for SOA records as they always exist in a zone.
func (d *dnsSOARecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan soaRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(plan.Zone.ValueString())

	// Apply the update to the SOA record
	soaUpdate(ctx, d.client, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (d *dnsSOARecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state soaRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zone := state.Zone.ValueString()

	msg := new(dns.Msg)
	msg.SetQuestion(zone, dns.TypeSOA)

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error reading SOA record for %q: ", zone), err.Error())
		return
	}

	if r.Rcode != dns.RcodeSuccess || len(r.Answer) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	record, ok := r.Answer[0].(*dns.SOA)
	if !ok {
		resp.Diagnostics.AddError("error parsing SOA record:", "unexpected record type received")
		return
	}

	state.ID = types.StringValue(zone)
	state.MName = types.StringValue(record.Ns)
	state.RName = types.StringValue(record.Mbox)
	state.Refresh = types.Int64Value(int64(record.Refresh))
	state.Retry = types.Int64Value(int64(record.Retry))
	state.Expire = types.Int64Value(int64(record.Expire))
	state.TTL = types.Int64Value(int64(record.Hdr.Ttl))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dnsSOARecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state soaRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(plan.Zone.ValueString())

	soaUpdate(ctx, d.client, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete is not supported for SOA records as they always exist in a zone.
func (d *dnsSOARecordResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("SOA record cannot be deleted",
		"The SOA record is fundamental to a DNS zone and cannot be removed. "+
			"The resource will be removed from Terraform state but the record will continue to exist.")
}

type soaRecordResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Zone    types.String `tfsdk:"zone"`
	MName   types.String `tfsdk:"mname"`
	RName   types.String `tfsdk:"rname"`
	Refresh types.Int64  `tfsdk:"refresh"`
	Retry   types.Int64  `tfsdk:"retry"`
	Expire  types.Int64  `tfsdk:"expire"`
	TTL     types.Int64  `tfsdk:"ttl"`
}

func soaUpdate(_ context.Context, client *DNSClient, plan *soaRecordResourceModel, diags *diag.Diagnostics) {
	zone := plan.Zone.ValueString()

	// Read current SOA to get the serial number
	serial := uint32(0)
	msgQ := new(dns.Msg)
	msgQ.SetQuestion(zone, dns.TypeSOA)
	rQ, err := exchange(msgQ, true, client)
	if err == nil && rQ.Rcode == dns.RcodeSuccess && len(rQ.Answer) > 0 {
		if currentSOA, ok := rQ.Answer[0].(*dns.SOA); ok {
			serial = currentSOA.Serial
		}
	}

	msg := new(dns.Msg)
	msg.SetUpdate(zone)

	rrStr := fmt.Sprintf("%s %d SOA %s %s %d %d %d %d %d", zone, plan.TTL.ValueInt64(),
		plan.MName.ValueString(), plan.RName.ValueString(),
		serial, plan.Refresh.ValueInt64(), plan.Retry.ValueInt64(), plan.Expire.ValueInt64(),
		plan.TTL.ValueInt64())

	rr, err := dns.NewRR(rrStr)
	if err != nil {
		diags.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
		return
	}

	msg.Insert([]dns.RR{rr})

	r, err := exchange(msg, true, client)
	if err != nil {
		diags.AddError("Error updating SOA record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		diags.AddError(fmt.Sprintf("Error updating SOA record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return
	}
}
