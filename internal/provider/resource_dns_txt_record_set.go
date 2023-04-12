package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

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
	_ resource.Resource                = (*dnsTXTRecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsTXTRecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsTXTRecordSetResource)(nil)
)

func NewDnsTXTRecordSetResource() resource.Resource {
	return &dnsTXTRecordSetResource{}
}

type dnsTXTRecordSetResource struct {
	client *DNSClient
}

func (d *dnsTXTRecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_txt_record_set"
}

func (d *dnsTXTRecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates a TXT type DNS record set.",
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
				Description: "The name of the record set. The `zone` argument will be appended to this value to create " +
					"the full record path.",
			},
			"txt": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The text records this record set will be set to.",
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
				Description: "Always set to the fully qualified domain name of the record set.",
			},
		},
	}
}

func (d *dnsTXTRecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsTXTRecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan txtRecordSetResourceModel

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

	var planTXT []string

	resp.Diagnostics.Append(plan.TXT.ElementsAs(ctx, &planTXT, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Loop through all the new addresses and insert them
	for _, txt := range planTXT {
		rrStr := fmt.Sprintf("%s %d TXT \"%s\"", fqdn, plan.TTL.ValueInt64(), txt)

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeTXT)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var txt []string

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.TXT:
				txt = append(txt, strings.Join(r.Txt, ""))
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an TXT record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		plan.TXT, convertDiags = types.SetValueFrom(ctx, plan.TXT.ElementType(ctx), txt)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsTXTRecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state txtRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeTXT)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var txt []string

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.TXT:
				txt = append(txt, strings.Join(r.Txt, ""))
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an TXT record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.TXT, convertDiags = types.SetValueFrom(ctx, state.TXT.ElementType(ctx), txt)
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

func (d *dnsTXTRecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state txtRecordSetResourceModel

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

	if !plan.TXT.Equal(state.TXT) {

		var planTXT, stateTXT []string

		resp.Diagnostics.Append(plan.TXT.ElementsAs(ctx, &planTXT, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.TXT.ElementsAs(ctx, &stateTXT, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add []string
		for _, newTXT := range planTXT {
			for _, oldTXT := range stateTXT {
				if oldTXT == newTXT {
					continue
				}
			}
			add = append(add, newTXT)
		}

		var remove []string
		for _, oldTXT := range stateTXT {
			for _, newTXT := range planTXT {
				if oldTXT == newTXT {
					continue
				}
			}
			remove = append(remove, oldTXT)
		}

		// Loop through all the old addresses and remove them
		for _, txt := range remove {
			rrStr := fmt.Sprintf("%s %d TXT \"%s\"", fqdn, plan.TTL.ValueInt64(), txt)

			rr_remove, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Remove([]dns.RR{rr_remove})
		}
		// Loop through all the new addresses and insert them
		for _, txt := range add {
			rrStr := fmt.Sprintf("%s %d TXT \"%s\"", fqdn, plan.TTL.ValueInt64(), txt)

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeTXT)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var txt []string

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.TXT:
				txt = append(txt, strings.Join(r.Txt, ""))
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record: ",
					"didn't get an TXT record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.TXT, convertDiags = types.SetValueFrom(ctx, state.TXT.ElementType(ctx), txt)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsTXTRecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state txtRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeTXT)...)
}

func (d *dnsTXTRecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

type txtRecordSetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
	Name types.String `tfsdk:"name"`
	TXT  types.Set    `tfsdk:"txt"`
	TTL  types.Int64  `tfsdk:"ttl"`
}
