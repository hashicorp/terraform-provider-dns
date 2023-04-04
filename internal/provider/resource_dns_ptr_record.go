package provider

import (
	"context"
	"fmt"

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
	_ resource.Resource                = (*dnsPTRRecordResource)(nil)
	_ resource.ResourceWithImportState = (*dnsPTRRecordResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsPTRRecordResource)(nil)
)

func NewDnsPTRRecordResource() resource.Resource {
	return &dnsPTRRecordResource{}
}

type dnsPTRRecordResource struct {
	client *DNSClient
}

func (d *dnsPTRRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ptr_record"
}

func (d *dnsPTRRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a PTR type DNS record.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
				},
				Description: "DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing dot.",
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
			"ptr": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
				},
				Description: "The canonical name this record will point to.",
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

func (d *dnsPTRRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsPTRRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ptrRecordSetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: plan.Name.ValueString(),
		Zone: plan.Zone.ValueString(),
	}
	rec_fqdn := resourceFQDN_framework(config)
	plan.ID = types.StringValue(rec_fqdn)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	//Insert new PTR record
	rrStrInsert := fmt.Sprintf("%s %d PTR %s", rec_fqdn, plan.TTL.ValueInt64(), plan.PTR.ValueString())

	rr_insert, err := dns.NewRR(rrStrInsert)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrInsert), err.Error())
		return
	}

	msg.Insert([]dns.RR{rr_insert})

	r, err := exchange_framework(msg, true, d.client)
	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode),
			dns.RcodeToString[r.Rcode])
		return
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypePTR)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		ptr, ttl, err := getPtrVal(record)
		if err != nil {
			resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
			return
		}

		plan.PTR = types.StringValue(ptr)
		plan.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsPTRRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ptrRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypePTR)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		ptr, ttl, err := getPtrVal(record)
		if err != nil {
			resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
			return
		}

		state.PTR = types.StringValue(ptr)
		state.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsPTRRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ptrRecordSetResourceModel

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
	rec_fqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	if !plan.PTR.Equal(state.PTR) {

		//Remove old PTR record
		rrStrRemove := fmt.Sprintf("%s %d PTR %s", rec_fqdn, plan.TTL.ValueInt64(), state.PTR.ValueString())

		rr_remove, err := dns.NewRR(rrStrRemove)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrRemove), err.Error())
			return
		}

		msg.Remove([]dns.RR{rr_remove})

		//Insert new PTR record
		rrStrInsert := fmt.Sprintf("%s %d PTR %s", rec_fqdn, plan.TTL.ValueInt64(), plan.PTR.ValueString())

		rr_insert, err := dns.NewRR(rrStrInsert)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStrInsert), err.Error())
			return
		}

		msg.Insert([]dns.RR{rr_insert})

		r, err := exchange_framework(msg, true, d.client)
		if err != nil {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
			return
		}
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypePTR)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("Error querying DNS record:", "multiple responses received")
			return
		}
		record := answers[0]
		ptr, ttl, err := getPtrVal(record)
		if err != nil {
			resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
			return
		}

		state.PTR = types.StringValue(ptr)
		state.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsPTRRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ptrRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}
	diags := resourceDnsDelete_framework(config, d.client, dns.TypePTR)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}
}

func (d *dnsPTRRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	config, diags := resourceDnsImport_framework(req.ID, d.client)
	if diags != nil {
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)
	if config.Name != "" {
		resp.State.SetAttribute(ctx, path.Root("name"), config.Name)
	}
}

type ptrRecordSetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
	Name types.String `tfsdk:"name"`
	PTR  types.String `tfsdk:"ptr"`
	TTL  types.Int64  `tfsdk:"ttl"`
}
