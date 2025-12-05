// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

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
	_ resource.Resource                = (*dnsMXRecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsMXRecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsMXRecordSetResource)(nil)
)

func NewDnsMXRecordSetResource() resource.Resource {
	return &dnsMXRecordSetResource{}
}

type dnsMXRecordSetResource struct {
	client *DNSClient
}

func (d *dnsMXRecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mx_record_set"
}

func (d *dnsMXRecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an MX type DNS record set.",
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
				Description: "Always set to the fully qualified domain name of the record set",
			},
		},
		Blocks: map[string]schema.Block{
			"mx": schema.SetNestedBlock{
				Description: "Can be specified multiple times for each MX record.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"preference": schema.Int64Attribute{
							Required:    true,
							Description: "The preference for the record.",
						},
						"exchange": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								dnsvalidator.IsZoneNameValid(),
							},
							Description: "The FQDN of the mail exchange, include the trailing dot.",
						},
					},
				},
			},
		},
	}
}

func (d *dnsMXRecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsMXRecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mxRecordSetResourceModel

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

	var planMX []mxBlockConfig
	var diags diag.Diagnostics

	diags.Append(plan.MX.ElementsAs(ctx, &planMX, false)...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

	// Loop through all the new addresses and insert them
	for _, mx := range planMX {
		rrStr := fmt.Sprintf("%s %d MX %d %s", fqdn, plan.TTL.ValueInt64(), mx.Preference.ValueInt64(), mx.Exchange.ValueString())

		rr_insert, err := dns.NewRR(rrStr)
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
			return
		}

		msg.Insert([]dns.RR{rr_insert})
	}

	r, err := exchange(msg, true, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode), dns.RcodeToString[r.Rcode])
		return
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeMX)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var mx []mxBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.MX:
				m := mxBlockConfig{
					Preference: types.Int64Value(int64(r.Preference)),
					Exchange:   types.StringValue(r.Mx),
				}
				mx = append(mx, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:", "didn't get an MX record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		plan.MX, convertDiags = types.SetValueFrom(ctx, plan.MX.ElementType(ctx), mx)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsMXRecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mxRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeMX)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var mx []mxBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.MX:
				m := mxBlockConfig{
					Preference: types.Int64Value(int64(r.Preference)),
					Exchange:   types.StringValue(r.Mx),
				}
				mx = append(mx, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:", "didn't get an MX record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.MX, convertDiags = types.SetValueFrom(ctx, state.MX.ElementType(ctx), mx)
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

func (d *dnsMXRecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state mxRecordSetResourceModel

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

	if !plan.MX.Equal(state.MX) {

		var planMX, stateMX []mxBlockConfig

		resp.Diagnostics.Append(plan.MX.ElementsAs(ctx, &planMX, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.MX.ElementsAs(ctx, &stateMX, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add []mxBlockConfig
		for _, newMX := range planMX {
			for _, oldMX := range stateMX {
				if oldMX == newMX {
					continue
				}
			}
			add = append(add, newMX)
		}

		var remove []mxBlockConfig
		for _, oldMX := range stateMX {
			for _, newMX := range planMX {
				if oldMX == newMX {
					continue
				}
			}
			remove = append(remove, oldMX)
		}

		// Loop through all the old addresses and remove them
		for _, mx := range remove {
			rrStr := fmt.Sprintf("%s %d MX %d %s", fqdn, plan.TTL.ValueInt64(), mx.Preference.ValueInt64(), mx.Exchange.ValueString())

			rr_remove, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Remove([]dns.RR{rr_remove})
		}
		// Loop through all the new addresses and insert them
		for _, mx := range add {
			rrStr := fmt.Sprintf("%s %d MX %d %s", fqdn, plan.TTL.ValueInt64(), mx.Preference.ValueInt64(), mx.Exchange.ValueString())

			rr_insert, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Insert([]dns.RR{rr_insert})
		}

		r, err := exchange(msg, true, d.client)
		if err != nil {
			resp.Diagnostics.AddError("Error updating DNS record:", err.Error())
			return
		}
		if r.Rcode != dns.RcodeSuccess {
			resp.Diagnostics.AddError(fmt.Sprintf("Error updating DNS record: %v", r.Rcode),
				dns.RcodeToString[r.Rcode])
			return
		}
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeMX)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var mx []mxBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.MX:
				m := mxBlockConfig{
					Preference: types.Int64Value(int64(r.Preference)),
					Exchange:   types.StringValue(r.Mx),
				}
				mx = append(mx, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an MX record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.MX, convertDiags = types.SetValueFrom(ctx, state.MX.ElementType(ctx), mx)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsMXRecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mxRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeMX)...)
}

func (d *dnsMXRecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

type mxRecordSetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
	Name types.String `tfsdk:"name"`
	MX   types.Set    `tfsdk:"mx"` //mxBlockConfig
	TTL  types.Int64  `tfsdk:"ttl"`
}
