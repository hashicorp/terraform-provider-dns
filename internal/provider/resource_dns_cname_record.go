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
	_ resource.Resource                = (*dnsCNAMERecordResource)(nil)
	_ resource.ResourceWithImportState = (*dnsCNAMERecordResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsCNAMERecordResource)(nil)
)

func NewDnsCNAMERecordResource() resource.Resource {
	return &dnsCNAMERecordResource{}
}

type dnsCNAMERecordResource struct {
	client *DNSClient
}

func (d *dnsCNAMERecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cname_record"
}

func (d *dnsCNAMERecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a CNAME type DNS record.",
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsRecordNameValid(),
				},
				Description: "The name of the record. The `zone` argument will be appended to this value to create " +
					"the full record path.",
			},
			"cname": schema.StringAttribute{
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
				Description: "The TTL of the record set. Defaults to `3600`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record",
			},
		},
	}
}

func (d *dnsCNAMERecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsCNAMERecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cnameRecordResourceModel

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

	rrStrInsert := fmt.Sprintf("%s %d CNAME %s", rec_fqdn, plan.TTL.ValueInt64(), plan.CNAME.ValueString())

	rr_insert, err := dns.NewRR(rrStrInsert)
	if err != nil {
		resp.Diagnostics.AddError("DNS CNAME record create error",
			fmt.Sprintf("error reading DNS record (%s): %s", rrStrInsert, err.Error()))
		return
	}
	msg.Insert([]dns.RR{rr_insert})

	r, err := exchange_framework(msg, true, d.client)
	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("DNS CNAME record create error",
			fmt.Sprintf("Error updating DNS record: %s", err))
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("DNS CNAME record create error",
			fmt.Sprintf("Error updating DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode]))
		return
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeCNAME)
	if err != nil {
		resp.Diagnostics.AddError("DNS CNAME record create error", err.Error())
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("DNS CNAME record create error",
				"Error querying DNS record: multiple responses received")
			return
		}
		record := answers[0]
		cname, ttl, err := getCnameVal(record)
		if err != nil {
			resp.Diagnostics.AddError("DNS CNAME record create error",
				fmt.Sprintf("Error querying DNS record: %s", err))
			return
		}

		plan.CNAME = types.StringValue(cname)
		plan.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	} else {
		resp.State.RemoveResource(ctx)
	}

}

func (d *dnsCNAMERecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cnameRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeCNAME)
	if err != nil {
		resp.Diagnostics.AddError("DNS CNAME record read error", err.Error())
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("DNS CNAME record read error",
				"Error querying DNS record: multiple responses received")
			return
		}
		record := answers[0]
		cname, ttl, err := getCnameVal(record)
		if err != nil {
			resp.Diagnostics.AddError("DNS CNAME record read error",
				fmt.Sprintf("Error querying DNS record: %s", err))
			return
		}

		state.CNAME = types.StringValue(cname)
		state.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsCNAMERecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state cnameRecordResourceModel

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

	if !plan.CNAME.Equal(state.CNAME) {

		rrStrRemove := fmt.Sprintf("%s %d CNAME %s", rec_fqdn, plan.TTL.ValueInt64(), state.CNAME.ValueString())

		rr_remove, err := dns.NewRR(rrStrRemove)
		if err != nil {
			resp.Diagnostics.AddError("DNS CNAME record update error",
				fmt.Sprintf("error reading DNS record (%s): %s", rrStrRemove, err.Error()))
			return
		}

		rrStrInsert := fmt.Sprintf("%s %d CNAME %s", rec_fqdn, plan.TTL.ValueInt64(), plan.CNAME.ValueString())

		rr_insert, err := dns.NewRR(rrStrInsert)
		if err != nil {
			resp.Diagnostics.AddError("DNS CNAME record update error",
				fmt.Sprintf("error reading DNS record (%s): %s", rrStrInsert, err.Error()))
			return
		}

		msg.Remove([]dns.RR{rr_remove})
		msg.Insert([]dns.RR{rr_insert})

		r, err := exchange_framework(msg, true, d.client)
		if err != nil {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError("DNS CNAME record update error",
				fmt.Sprintf("Error updating DNS record: %s", err))
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError("DNS CNAME record update error",
				fmt.Sprintf("Error updating DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode]))
			return
		}
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeCNAME)
	if err != nil {
		resp.Diagnostics.AddError("DNS CNAME record update error", err.Error())
		return
	}

	if len(answers) > 0 {
		if len(answers) > 1 {
			resp.Diagnostics.AddError("DNS CNAME record update error",
				"Error querying DNS record: multiple responses received")
			return
		}
		record := answers[0]
		cname, ttl, err := getCnameVal(record)
		if err != nil {
			resp.Diagnostics.AddError("DNS CNAME record update error",
				fmt.Sprintf("Error querying DNS record: %s", err))
			return
		}

		state.CNAME = types.StringValue(cname)
		state.TTL = types.Int64Value(int64(ttl))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsCNAMERecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cnameRecordResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}
	err := resourceDnsDelete_framework(config, d.client, dns.TypeCNAME)
	if err != nil {
		resp.Diagnostics.AddError("Delete resource error", err.Error())
		return
	}
}

func (d *dnsCNAMERecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	config, err := resourceDnsImport_framework(req.ID, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Import resource error", err.Error())
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("name"), config.Name)
	resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)
}

type cnameRecordResourceModel struct {
	ID    types.String `tfsdk:"id"`
	Zone  types.String `tfsdk:"zone"`
	Name  types.String `tfsdk:"name"`
	CNAME types.String `tfsdk:"cname"`
	TTL   types.Int64  `tfsdk:"ttl"`
}
