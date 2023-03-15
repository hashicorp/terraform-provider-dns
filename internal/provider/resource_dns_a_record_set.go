package provider

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"

	"github.com/hashicorp/terraform-provider-dns/internal/modifiers/int64modifier"
	"github.com/hashicorp/terraform-provider-dns/internal/validators/dnsvalidator"
)

var (
	_ resource.Resource                = (*dnsARecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsARecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsARecordSetResource)(nil)
)

func NewDnsARecordSetResource() resource.Resource {
	return &dnsARecordSetResource{}
}

func testProvider() dnsARecordSetResource {
	return dnsARecordSetResource{}
}

type dnsARecordSetResource struct {
	client *DNSClient
}

func (d *dnsARecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		resp.Diagnostics.AddError(
			"No Provider Configuration",
			"DNS update server is not set",
		)
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

func (d *dnsARecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_a_record_set"
}

func (d *dnsARecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an A type DNS record set.",
		Attributes: map[string]schema.Attribute{
			"zone": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsZoneNameValid(),
				},
				Description: "DNS zone the record set belongs to. It must be an FQDN, that is, include the trailing " +
					"dot.",
			},
			"name": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					dnsvalidator.IsRecordNameValid(),
				},
				Description: "The name of the record set. The `zone` argument will be appended to this value to " +
					"create the full record path.",
			},
			"addresses": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "The IPv4 addresses this record set will point to.",
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64modifier.Int64Default(3600),
					int64planmodifier.RequiresReplace(),
				},
				Description: "The TTL of the record set. Defaults to `3600`.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record set",
			},
		},
	}
}

func (d *dnsARecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan aRecordSetResourceModel

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

	var newAddrs []types.String
	plan.Addresses.ElementsAs(ctx, &newAddrs, true)

	// Loop through all the new addresses and insert them
	for _, addr := range newAddrs {
		rrStr := fmt.Sprintf("%s %d A %s", rec_fqdn, plan.TTL.ValueInt64(), stripLeadingZeros(addr.ValueString()))

		rr_insert, err := dns.NewRR(rrStr)
		if err != nil {
			resp.Diagnostics.AddError("DNS query error", fmt.Sprintf("Error reading DNS record (%s): %s", rrStr, err))
			return
		}

		msg.Insert([]dns.RR{rr_insert})
	}

	r, err := exchange_framework(msg, true, d.client)
	if err != nil {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("DNS update error", fmt.Sprintf("Error updating DNS record: %s", err))
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.State.RemoveResource(ctx)
		resp.Diagnostics.AddError("DNS update error", fmt.Sprintf("Error updating DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode]))
		return
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeA)
	if err != nil {
		resp.Diagnostics.AddError("DNS query error", err.Error())
		return
	}

	if len(answers) > 0 {
		ttl, addresses, diags := queryTTLAndAddresses(answers)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		var convertDiags diag.Diagnostics
		plan.Addresses, convertDiags = types.SetValueFrom(ctx, plan.Addresses.Type(ctx), addresses)
		if convertDiags.HasError() {
			return
		}
		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	} else {
		resp.State.RemoveResource(ctx)
	}

}

func (d *dnsARecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state aRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeA)
	if err != nil {
		resp.Diagnostics.AddError("DNS query error", err.Error())
		return
	}

	if len(answers) > 0 {
		ttl, addresses, diags := queryTTLAndAddresses(answers)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		var convertDiags diag.Diagnostics
		state.Addresses, convertDiags = types.SetValueFrom(ctx, state.Addresses.Type(ctx), addresses)
		if convertDiags.HasError() {
			return
		}
		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsARecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state aRecordSetResourceModel

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

	if !plan.Addresses.Equal(state.Addresses) {
		var oldAddrs []types.String
		state.Addresses.ElementsAs(ctx, &oldAddrs, true)

		var newAddrs []types.String
		plan.Addresses.ElementsAs(ctx, &newAddrs, true)

		resp.Diagnostics.Append(updateAddresses(oldAddrs, newAddrs, msg, rec_fqdn, plan.TTL.ValueInt64())...)
		if resp.Diagnostics.HasError() {
			return
		}

		r, err := exchange_framework(msg, true, d.client)
		if err != nil {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError("DNS update error", fmt.Sprintf("Error updating DNS record: %s", err))
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.State.RemoveResource(ctx)
			resp.Diagnostics.AddError("DNS update error", fmt.Sprintf("Error updating DNS record: %v (%s)", r.Rcode, dns.RcodeToString[r.Rcode]))
			return
		}
	}

	answers, err := resourceDnsRead_framework(config, d.client, dns.TypeA)
	if err != nil {
		resp.Diagnostics.AddError("DNS query error", err.Error())
		return
	}

	if len(answers) > 0 {
		ttl, addresses, diags := queryTTLAndAddresses(answers)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		var convertDiags diag.Diagnostics
		plan.Addresses, convertDiags = types.SetValueFrom(ctx, plan.Addresses.Type(ctx), addresses)
		if convertDiags.HasError() {
			return
		}
		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	} else {
		resp.State.RemoveResource(ctx)
	}
}

func (d *dnsARecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state aRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}
	err := resourceDnsDelete_framework(config, d.client, dns.TypeA)
	if err != nil {
		resp.Diagnostics.AddError("Delete resource error", err.Error())
		return
	}
}

func (d *dnsARecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	config, err := resourceDnsImport_framework(req.ID, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Import resource error", err.Error())
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("name"), config.Name)
	resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)
}

type aRecordSetResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Zone      types.String `tfsdk:"zone"`
	Name      types.String `tfsdk:"name"`
	Addresses types.Set    `tfsdk:"addresses"`
	TTL       types.Int64  `tfsdk:"ttl"`
}

func updateAddresses(oldAddrs []types.String, newAddrs []types.String, msg *dns.Msg, fqdn string, ttl int64) diag.Diagnostics {
	var diags diag.Diagnostics

	var remove []types.String
	for _, oldAddr := range oldAddrs {
		for _, newAddr := range newAddrs {
			if oldAddr.Equal(newAddr) {
				continue
			}
		}
		remove = append(remove, oldAddr)
	}

	var add []types.String
	for _, newAddr := range newAddrs {
		for _, oldAddr := range oldAddrs {
			if oldAddr.Equal(newAddr) {
				continue
			}
		}
		add = append(add, newAddr)
	}

	// Loop through all the old addresses and remove them
	for _, addr := range remove {
		rrStr := fmt.Sprintf("%s %d A %s", fqdn, ttl, stripLeadingZeros(addr.ValueString()))

		rr_remove, err := dns.NewRR(rrStr)
		if err != nil {
			diags.AddError("DNS query error", fmt.Sprintf("Error reading DNS record (%s): %s", rrStr, err))
			return diags
		}

		msg.Remove([]dns.RR{rr_remove})
	}
	// Loop through all the new addresses and insert them
	for _, addr := range add {
		rrStr := fmt.Sprintf("%s %d A %s", fqdn, ttl, stripLeadingZeros(addr.ValueString()))

		rr_insert, err := dns.NewRR(rrStr)
		if err != nil {
			diags.AddError("DNS query error", fmt.Sprintf("Error reading DNS record (%s): %s", rrStr, err))
			return diags
		}

		msg.Insert([]dns.RR{rr_insert})
	}

	return diags
}

func queryTTLAndAddresses(records []dns.RR) (sort.IntSlice, []string, diag.Diagnostics) {
	var ttl sort.IntSlice
	var addresses []string
	var diags diag.Diagnostics

	for _, record := range records {
		addr, t, err := getAVal(record)
		if err != nil {
			diags.AddError("Error querying DNS record", err.Error())
			return nil, nil, diags
		}
		addresses = append(addresses, addr)
		ttl = append(ttl, t)
	}
	sort.Sort(ttl)
	return ttl, addresses, diags
}

func (d dnsARecordSetResource) DnsClient() *DNSClient {
	return d.client
}
