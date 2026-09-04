// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
  "strconv"
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
	_ resource.Resource                = (*dnsCAARecordSetResource)(nil)
	_ resource.ResourceWithImportState = (*dnsCAARecordSetResource)(nil)
	_ resource.ResourceWithConfigure   = (*dnsCAARecordSetResource)(nil)
)

func NewDnsCAARecordSetResource() resource.Resource {
	return &dnsCAARecordSetResource{}
}

type dnsCAARecordSetResource struct {
	client *DNSClient
}

func (d *dnsCAARecordSetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caa_record_set"
}

func (d *dnsCAARecordSetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an CAA type DNS record set.",
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
				Description: "Always set to the fully qualified domain name of the record set.",
			},
		},
		Blocks: map[string]schema.Block{
			"caa": schema.SetNestedBlock{
				Description: "Can be specified multiple times for each CAA record.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"flags": schema.Int64Attribute{
							Required:    true,
							Description: "The flags for the record.",
						},
						"tag": schema.StringAttribute{
							Required: true,
              // Validators: []validator.String{
							//	dnsvalidator.IsCAATagValid(),
							//},
							Description: "The tag of the CAA record, must be one of 'issue', 'issuewild', 'iodef'.",
						},
            "value": schema.StringAttribute{
              Required: true,
              Description: "The value for the record.  Do not include outer quotes, escape inner quotes.",
            },
					},
				},
			},
		},
	}
}

func (d *dnsCAARecordSetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (d *dnsCAARecordSetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan caaRecordSetResourceModel

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

	var planCAA []caaBlockConfig
	var diags diag.Diagnostics

	diags.Append(plan.CAA.ElementsAs(ctx, &planCAA, false)...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

	// Loop through all the new addresses and insert them
	for _, caa := range planCAA {
		rrStr := fmt.Sprintf("%s %d CAA %d %s %s", fqdn, plan.TTL.ValueInt64(), caa.Flags.ValueInt64(), caa.Tag.ValueString(), strconv.Quote(caa.Value.ValueString()))

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeCAA)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var caa []caaBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.CAA:
				m := caaBlockConfig{
					Flags: types.Int64Value(int64(r.Flags)),
					Tag:   types.StringValue(r.Tag),
          Value: types.StringValue(r.Value),
				}
				caa = append(caa, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:", "didn't get a CAA record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		plan.CAA, convertDiags = types.SetValueFrom(ctx, plan.CAA.ElementType(ctx), caa)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		plan.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (d *dnsCAARecordSetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state caaRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeCAA)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var caa []caaBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.CAA:
				m := caaBlockConfig{
					Flags: types.Int64Value(int64(r.Flags)),
					Tag:   types.StringValue(r.Tag),
          Value: types.StringValue(r.Value),
				}
				caa = append(caa, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:", "didn't get a CAA record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.CAA, convertDiags = types.SetValueFrom(ctx, state.CAA.ElementType(ctx), caa)
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

func (d *dnsCAARecordSetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state caaRecordSetResourceModel

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

	if !plan.CAA.Equal(state.CAA) {

		var planCAA, stateCAA []caaBlockConfig

		resp.Diagnostics.Append(plan.CAA.ElementsAs(ctx, &planCAA, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		resp.Diagnostics.Append(state.CAA.ElementsAs(ctx, &stateCAA, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		var add []caaBlockConfig
		for _, newCAA := range planCAA {
			for _, oldCAA := range stateCAA {
				if oldCAA == newCAA {
					continue
				}
			}
			add = append(add, newCAA)
		}

		var remove []caaBlockConfig
		for _, oldCAA := range stateCAA {
			for _, newCAA := range planCAA {
				if oldCAA == newCAA {
					continue
				}
			}
			remove = append(remove, oldCAA)
		}

		// Loop through all the old addresses and remove them
		for _, caa := range remove {
		  rrStr := fmt.Sprintf("%s %d CAA %d %s %s", fqdn, plan.TTL.ValueInt64(), caa.Flags.ValueInt64(), caa.Tag.ValueString(), strconv.Quote(caa.Value.ValueString()))

			rr_remove, err := dns.NewRR(rrStr)
			if err != nil {
				resp.Diagnostics.AddError(fmt.Sprintf("Error reading DNS record (%s):", rrStr), err.Error())
				return
			}

			msg.Remove([]dns.RR{rr_remove})
		}
		// Loop through all the new addresses and insert them
		for _, caa := range add {
		  rrStr := fmt.Sprintf("%s %d CAA %d %s %s", fqdn, plan.TTL.ValueInt64(), caa.Flags.ValueInt64(), caa.Tag.ValueString(), strconv.Quote(caa.Value.ValueString()))

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

	answers, diags := resourceDnsRead_framework(config, d.client, dns.TypeCAA)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if len(answers) > 0 {
		var ttl sort.IntSlice

		var caa []caaBlockConfig
		for _, record := range answers {
			switch r := record.(type) {
			case *dns.CAA:
				m := caaBlockConfig{
					Flags: types.Int64Value(int64(r.Flags)),
					Tag:   types.StringValue(r.Tag),
          Value: types.StringValue(r.Value),
				}
				caa = append(caa, m)
				ttl = append(ttl, int(r.Hdr.Ttl))
			default:
				resp.Diagnostics.AddError("Error querying DNS record:",
					"didn't get an CAA record")
				return
			}
		}
		sort.Sort(ttl)

		var convertDiags diag.Diagnostics
		state.CAA, convertDiags = types.SetValueFrom(ctx, state.CAA.ElementType(ctx), caa)
		if convertDiags.HasError() {
			resp.Diagnostics.Append(convertDiags...)
			return
		}

		state.TTL = types.Int64Value(int64(ttl[0]))

		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (d *dnsCAARecordSetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state caaRecordSetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	config := dnsConfig{
		Name: state.Name.ValueString(),
		Zone: state.Zone.ValueString(),
	}

	resp.Diagnostics.Append(resourceDnsDelete_framework(config, d.client, dns.TypeCAA)...)
}

func (d *dnsCAARecordSetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

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

type caaRecordSetResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Zone types.String `tfsdk:"zone"`
	Name types.String `tfsdk:"name"`
	CAA  types.Set    `tfsdk:"caa"` //caaBlockConfig
	TTL  types.Int64  `tfsdk:"ttl"`
}
