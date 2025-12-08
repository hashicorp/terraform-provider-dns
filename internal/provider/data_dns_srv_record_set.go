// Copyright IBM Corp. 2016, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = (*dnsSRVRecordSetDataSource)(nil)
)

func NewDnsSRVRecordSetDataSource() datasource.DataSource {
	return &dnsSRVRecordSetDataSource{}
}

type dnsSRVRecordSetDataSource struct{}

func (d *dnsSRVRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_srv_record_set"
}

func (d *dnsSRVRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS SRV records for a service.",
		Attributes: map[string]schema.Attribute{
			"service": schema.StringAttribute{
				Required:    true,
				Description: "Service to look up.",
			},
			"srv": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"priority": types.Int64Type,
						"weight":   types.Int64Type,
						"port":     types.Int64Type,
						"target":   types.StringType,
					},
				},
				Description: "A list of records. They are sorted to stay consistent across runs.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the service.",
			},
		},
	}
}

func (d *dnsSRVRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config srvRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	service := config.Service.ValueString()
	_, records, err := net.LookupSRV("", "", service)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up SRV records for %q: ", service), err.Error())
		return
	}

	// Sort by priority ascending, weight descending, target
	// alphabetically, and port ascending
	sort.Slice(records, func(i, j int) bool {
		if records[i].Priority < records[j].Priority {
			return true
		}
		if records[i].Priority > records[j].Priority {
			return false
		}
		if records[i].Weight > records[j].Weight {
			return true
		}
		if records[i].Weight < records[j].Weight {
			return false
		}
		if records[i].Target < records[j].Target {
			return true
		}
		if records[i].Target > records[j].Target {
			return false
		}
		return records[i].Port < records[j].Port
	})

	srv := make([]srvBlockConfig, len(records))
	for i, record := range records {
		srv[i] = srvBlockConfig{
			Priority: types.Int64Value(int64(record.Priority)),
			Weight:   types.Int64Value(int64(record.Weight)),
			Port:     types.Int64Value(int64(record.Port)),
			Target:   types.StringValue(record.Target),
		}
	}

	var convertDiags diag.Diagnostics
	config.SRV, convertDiags = types.ListValueFrom(ctx, config.SRV.ElementType(ctx), srv)
	if convertDiags.HasError() {
		resp.Diagnostics.Append(convertDiags...)
		return
	}

	config.ID = config.Service
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type srvRecordSetConfig struct {
	ID      types.String `tfsdk:"id"`
	Service types.String `tfsdk:"service"`
	SRV     types.List   `tfsdk:"srv"` //srvBlockConfig
}

type srvBlockConfig struct {
	Priority types.Int64  `tfsdk:"priority"`
	Weight   types.Int64  `tfsdk:"weight"`
	Port     types.Int64  `tfsdk:"port"`
	Target   types.String `tfsdk:"target"`
}
