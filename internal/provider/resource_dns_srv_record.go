// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                = (*dnsSRVRecordResource)(nil)
	_ resource.ResourceWithImportState = (*dnsSRVRecordResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsSRVRecordResource)(nil)
)

func NewDnsSRVRecordResource() resource.Resource {
	return &dnsSRVRecordResource{}
}

type dnsSRVRecordResource struct {
	client *DNSClient
}

func (d *dnsSRVRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_srv_record"
}

func (d *dnsSRVRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an individual SRV type DNS record.",
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
			"service": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				Description: "The service name in format _service._protocol.name.",
			},
			"priority": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The priority for the record.",
			},
			"weight": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The weight for the record.",
			},
			"port": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The port for the service on the target.",
			},
			"target": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
				},
				Description: "The FQDN of the target, include the trailing dot.",
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(3600),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Description: "The TTL of the record. Defaults to `3600`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record.",
			},
		},
	}
}

func (d *dnsSRVRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsSRVRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan srvRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := srvFQDN(plan.Service.ValueString(), plan.Name.ValueString(), plan.Zone.ValueString())
	plan.ID = types.StringValue(fqdn)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, plan.TTL.ValueInt64(),
		plan.Priority.ValueInt64(), plan.Weight.ValueInt64(), plan.Port.ValueInt64(), plan.Target.ValueString())

	rrInsert, err := dns.NewRR(rrStr)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
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

	answers, diags := srvRead(fqdn, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	for _, ans := range answers {
		srvRec, ok := ans.(*dns.SRV)
		if !ok {
			continue
		}
		if int(srvRec.Priority) == int(plan.Priority.ValueInt64()) &&
			int(srvRec.Weight) == int(plan.Weight.ValueInt64()) &&
			int(srvRec.Port) == int(plan.Port.ValueInt64()) &&
			srvRec.Target == plan.Target.ValueString() {
			plan.TTL = types.Int64Value(int64(srvRec.Hdr.Ttl))
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (d *dnsSRVRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state srvRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := srvFQDN(state.Service.ValueString(), state.Name.ValueString(), state.Zone.ValueString())

	answers, diags := srvRead(fqdn, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	found := false
	for _, ans := range answers {
		srvRec, ok := ans.(*dns.SRV)
		if !ok {
			continue
		}
		if int(srvRec.Priority) == int(state.Priority.ValueInt64()) &&
			int(srvRec.Weight) == int(state.Weight.ValueInt64()) &&
			int(srvRec.Port) == int(state.Port.ValueInt64()) &&
			srvRec.Target == state.Target.ValueString() {
			state.TTL = types.Int64Value(int64(srvRec.Hdr.Ttl))
			found = true
			break
		}
	}

	if found {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsSRVRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state srvRecordResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := srvFQDN(plan.Service.ValueString(), plan.Name.ValueString(), plan.Zone.ValueString())

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	rrStrRemove := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, state.TTL.ValueInt64(),
		state.Priority.ValueInt64(), state.Weight.ValueInt64(), state.Port.ValueInt64(), state.Target.ValueString())

	rrRemove, err := dns.NewRR(rrStrRemove)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrRemove), err.Error())
		return
	}
	msg.Remove([]dns.RR{rrRemove})

	rrStrInsert := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, plan.TTL.ValueInt64(),
		plan.Priority.ValueInt64(), plan.Weight.ValueInt64(), plan.Port.ValueInt64(), plan.Target.ValueString())

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

	state.Target = types.StringValue(plan.Target.ValueString())
	state.Priority = types.Int64Value(plan.Priority.ValueInt64())
	state.Weight = types.Int64Value(plan.Weight.ValueInt64())
	state.Port = types.Int64Value(plan.Port.ValueInt64())
	state.TTL = types.Int64Value(plan.TTL.ValueInt64())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dnsSRVRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state srvRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fqdn := srvFQDN(state.Service.ValueString(), state.Name.ValueString(), state.Zone.ValueString())

	msg := new(dns.Msg)
	msg.SetUpdate(state.Zone.ValueString())

	rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, state.TTL.ValueInt64(),
		state.Priority.ValueInt64(), state.Weight.ValueInt64(), state.Port.ValueInt64(), state.Target.ValueString())

	rr, err := dns.NewRR(rrStr)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
		return
	}
	msg.Remove([]dns.RR{rr})

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting DNS record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.Diagnostics.AddError(fmt.Sprintf("Error deleting DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return
	}
}

func (d *dnsSRVRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	config, diags := resourceDnsImport_framework(req.ID, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)...)
	if config.Name != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), config.Name)...)
	}
}

type srvRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Zone     types.String `tfsdk:"zone"`
	Name     types.String `tfsdk:"name"`
	Service  types.String `tfsdk:"service"`
	Priority types.Int64  `tfsdk:"priority"`
	Weight   types.Int64  `tfsdk:"weight"`
	Port     types.Int64  `tfsdk:"port"`
	Target   types.String `tfsdk:"target"`
	TTL      types.Int64  `tfsdk:"ttl"`
}

func srvFQDN(service, name, zone string) string {
	fqdn := service + "." + zone
	if name != "" {
		fqdn = name + "." + fqdn
	}
	return dns.Fqdn(fqdn)
}

func srvRead(fqdn string, client *DNSClient) ([]dns.RR, diag.Diagnostics) {
	var diags diag.Diagnostics

	msg := new(dns.Msg)
	msg.SetQuestion(fqdn, dns.TypeSRV)
	msg.RecursionDesired = client.recursive

	r, err := exchange(msg, true, client)
	if err != nil {
		diags.AddError("Error querying DNS record:", err.Error())
		return nil, diags
	}
	switch r.Rcode {
	case dns.RcodeSuccess:
		if len(r.Answer) == 0 {
			return nil, nil
		}
	case dns.RcodeNameError:
		return nil, nil
	default:
		diags.AddError("Error querying DNS record:", dns.RcodeToString[r.Rcode])
		return nil, diags
	}

	return r.Answer, nil
}
