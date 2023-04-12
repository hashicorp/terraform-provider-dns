package provider

import (
	"context"
	"fmt"
	"sort"

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
	_ resource.Resource                = (*dnsSRVRecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsSRVRecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsSRVRecordSetResource)(nil)
)

func NewDnsSRVRecordSetResource() resource.Resource {
	return &dnsSRVRecordSetResource{}
}

type dnsSRVRecordSetResource struct {
	client *DNSClient
}

func (d *dnsSRVRecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_srv_record_set"
}

func (d *dnsSRVRecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an SRV type DNS record set.",
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
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the fully qualified domain name of the record set.",
			},
		},
		Blocks: map[string]schema.Block{
			"srv": schema.SetNestedBlock{
				Description: "Can be specified multiple times for each SRV record.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"priority": schema.Int64Attribute{
							Required:    true,
							Description: "The priority for the record.",
						},
						"weight": schema.Int64Attribute{
							Required:    true,
							Description: "The weight for the record.",
						},
						"port": schema.Int64Attribute{
							Required:    true,
							Description: "The port for the service on the target.",
						},
						"target": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								dnsvalidator.IsZoneNameValid(),
							},
							Description: "The FQDN of the target, include the trailing dot.",
						},
					},
				},
			},
		},
	}
}

func (d *dnsSRVRecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsSRVRecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan srvRecordSetResourceModel

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

	var planSRV []srvBlockConfig

	resp.Diagnostics.Append(plan.SRV.ElementsAs(ctx, &planSRV, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Loop through all the new addresses and insert them
	for _, srv := range planSRV {
		rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, plan.TTL.ValueInt64(), srv.Priority.ValueInt64(),
			srv.Weight.ValueInt64(), srv.Port.ValueInt64(), srv.Target.ValueString())

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeSRV)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var srv []srvBlockConfig

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.SRV:
				s := srvBlockConfig{
					Priority: types.Int64Value(int64(r.Priority)),
					Weight:   types.Int64Value(int64(r.Weight)),
					Port:     types.Int64Value(int64(r.Port)),
					Target:   types.StringValue(r.Target),
				}
				srv = append(srv, s)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an SRV record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		plan.SRV, convertDiags = types.SetValueFrom(ctx, plan.SRV.ElementType(ctx), srv)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsSRVRecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state srvRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeSRV)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var srv []srvBlockConfig

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.SRV:
				s := srvBlockConfig{
					Priority: types.Int64Value(int64(r.Priority)),
					Weight:   types.Int64Value(int64(r.Weight)),
					Port:     types.Int64Value(int64(r.Port)),
					Target:   types.StringValue(r.Target),
				}
				srv = append(srv, s)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an SRV record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.SRV, convertDiags = types.SetValueFrom(ctx, state.SRV.ElementType(ctx), srv)
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

func (d *dnsSRVRecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state srvRecordSetResourceModel

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

	if !plan.SRV.Equal(state.SRV) {

		var planSRV, stateSRV []srvBlockConfig

		resp.Diagnostics.Append(plan.SRV.ElementsAs(ctx, &planSRV, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.SRV.ElementsAs(ctx, &stateSRV, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add []srvBlockConfig
		for _, newSRV := range planSRV {
			for _, oldSRV := range stateSRV {
				if oldSRV == newSRV {
					continue
				}
			}
			add = append(add, newSRV)
		}

		var remove []srvBlockConfig
		for _, oldSRV := range stateSRV {
			for _, newSRV := range planSRV {
				if oldSRV == newSRV {
					continue
				}
			}
			remove = append(remove, oldSRV)
		}

		// Loop through all the old addresses and remove them
		for _, srv := range remove {
			rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, plan.TTL.ValueInt64(), srv.Priority.ValueInt64(),
				srv.Weight.ValueInt64(), srv.Port.ValueInt64(), srv.Target.ValueString())

			rr_remove, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Remove([]dns.RR{rr_remove})
		}
		// Loop through all the new addresses and insert them
		for _, srv := range add {
			rrStr := fmt.Sprintf("%s %d SRV %d %d %d %s", fqdn, plan.TTL.ValueInt64(), srv.Priority.ValueInt64(),
				srv.Weight.ValueInt64(), srv.Port.ValueInt64(), srv.Target.ValueString())

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeSRV)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice
		var srv []srvBlockConfig

		for _, record := range answers {
			switch r := record.(type) {
			case *dns.SRV:
				s := srvBlockConfig{
					Priority: types.Int64Value(int64(r.Priority)),
					Weight:   types.Int64Value(int64(r.Weight)),
					Port:     types.Int64Value(int64(r.Port)),
					Target:   types.StringValue(r.Target),
				}
				srv = append(srv, s)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an SRV record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.SRV, convertDiags = types.SetValueFrom(ctx, state.SRV.ElementType(ctx), srv)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsSRVRecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state srvRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}
	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeSRV)...)
}

func (d *dnsSRVRecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	config, diags := resourceDnsImport_framework(req.ID, d.client)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), config.Name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), config.Zone)...)
}

type srvRecordSetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
	Name types.String `tfsdk:"name"`
	SRV  types.Set    `tfsdk:"srv"` //srvBlockConfig
	TTL  types.Int64  `tfsdk:"ttl"`
}
