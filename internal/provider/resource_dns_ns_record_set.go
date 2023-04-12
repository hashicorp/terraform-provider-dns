package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
	_ resource.Resource                = (*dnsNSRecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsNSRecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsNSRecordSetResource)(nil)
)

func NewDnsNSRecordSetResource() resource.Resource {
	return &dnsNSRecordSetResource{}
}

type dnsNSRecordSetResource struct {
	client *DNSClient
}

func (d *dnsNSRecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ns_record_set"
}

func (d *dnsNSRecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an NS type DNS record set.",
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsRecordNameValid(),
				},
				Description: "The name of the record set. The `zone` argument will be appended to this value to create " +
					"the full record path.",
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
			"nameservers": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(dnsvalidator.IsZoneNameValid()),
				},
				Description: "The nameservers this record set will point to.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record set.",
			},
		},
	}
}

func (d *dnsNSRecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsNSRecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan nsRecordSetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: plan.Name.ValueString(),
		Zone: plan.Zone.ValueString(),
	}
	fqdn := resourceFQDN_framework(config)
	plan.ID = types.StringValue(fqdn)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	var planNS []string

	resp.Diagnostics.Append(plan.Nameservers.ElementsAs(ctx, &planNS, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Loop through all the new nameservers and insert them
	for _, nameserver := range planNS {
		rrStr := fmt.Sprintf("%s %d NS %s", fqdn, plan.TTL.ValueInt64(), nameserver)

		rr_insert, err := dns.NewRR(rrStr)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
			return
		}

		msg.Insert([]dns.RR{rr_insert})
	}

	r, err := exchange_framework(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeNS)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var nameservers []string

		for _, record := range answers {
			nameserver, t, err := getNSVal(record)
			if err != nil {
				resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
				return
			}
			nameservers = append(nameservers, nameserver)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		plan.Nameservers, convertDiags = types.SetValueFrom(ctx, plan.Nameservers.ElementType(ctx), nameservers)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsNSRecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state nsRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeNS)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var nameservers []string

		for _, record := range answers {
			nameserver, t, err := getNSVal(record)
			if err != nil {
				resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
				return
			}
			nameservers = append(nameservers, nameserver)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.Nameservers, convertDiags = types.SetValueFrom(ctx, state.Nameservers.ElementType(ctx), nameservers)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsNSRecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state nsRecordSetResourceModel

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
	fqdn := resourceFQDN_framework(config)

	msg := new(dns.Msg)
	msg.SetUpdate(plan.Zone.ValueString())

	if !plan.Nameservers.Equal(state.Nameservers) {

		var planNS, stateNS []string

		resp.Diagnostics.Append(plan.Nameservers.ElementsAs(ctx, &planNS, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.Nameservers.ElementsAs(ctx, &stateNS, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add []string
		for _, newNS := range planNS {
			for _, oldNS := range stateNS {
				if oldNS == newNS {
					continue
				}
			}
			add = append(add, newNS)
		}

		var remove []string
		for _, oldNS := range stateNS {
			for _, newNS := range planNS {
				if oldNS == newNS {
					continue
				}
			}
			remove = append(remove, oldNS)
		}

		// Loop through all the old nameservers and remove them
		for _, nameserver := range remove {
			rrStr := fmt.Sprintf("%s %d NS %s", fqdn, plan.TTL.ValueInt64(), nameserver)

			rr_remove, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Remove([]dns.RR{rr_remove})
		}
		// Loop through all the new nameservers and insert them
		for _, nameserver := range add {
			rrStr := fmt.Sprintf("%s %d NS %s", fqdn, plan.TTL.ValueInt64(), nameserver)

			rr_insert, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Insert([]dns.RR{rr_insert})
		}

		r, err := exchange_framework(msg, true, d.client)
		if err != nil {
			resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
			return
		}
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeNS)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var nameservers []string

		for _, record := range answers {
			nameserver, t, err := getNSVal(record)
			if err != nil {
				resp.Diagnostics.AddError("Error querying DNS record:", err.Error())
				return
			}
			nameservers = append(nameservers, nameserver)
			ttl = append(ttl, t)
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.Nameservers, convertDiags = types.SetValueFrom(ctx, state.Nameservers.ElementType(ctx), nameservers)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsNSRecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state nsRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeNS)...)
}

func (d *dnsNSRecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	config, diags := resourceDnsImport_framework(req.ID, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), config.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)...)
}

type nsRecordSetResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Zone        types.String `tfsdk:"zone"`
	Name        types.String `tfsdk:"name"`
	Nameservers types.Set    `tfsdk:"nameservers"`
	TTL         types.Int64  `tfsdk:"ttl"`
}
