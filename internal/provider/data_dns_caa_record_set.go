// Copyright (c) HashiCorp, Inc.
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
	_ datasource.DataSource = (*dnsCAARecordSetDataSource)(nil)
)

func NewDnsCAARecordSetDataSource() datasource.DataSource {
	return &dnsCAARecordSetDataSource{}
}

type dnsCAARecordSetDataSource struct{}

func (d *dnsCAARecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caa_record_set"
}

func (d *dnsCAARecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS CAA records for a domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "Domain to look up.",
			},
			"caa": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"flags": types.Int64Type,
						"tags":  types.StringType,
						"value": types.StringType,
					},
				},
				Description: "A list of records. They are sorted to stay consistent across runs.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the domain.",
			},
		},
	}
}

func (d *dnsCAARecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config caaRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := config.Domain.ValueString()
	records, err := net.LookupCAA(domain)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up CAA records for %q: ", domain), err.Error())
		return
	}

	// Sort by flags ascending, tags alphabetically, and value alphabetically
	sort.Slice(records, func(i, j int) bool {
		if records[i].Flags < records[j].Flags {
			return true
		}
		if records[i].Flags > records[j].Flags {
			return false
		}
    if records[i].Tag < records[j].Tag {
      return true
    }
    if records[i].Tag > records[j].Tag {
      return false
    }
		return records[i].Value < records[j].Value
	})

	caa := make([]caaBlockConfig, len(records))
	for i, record := range records {
		caa[i] = caaBlockConfig{
			Flags: types.Int64Value(int64(record.Flags)),
			Tag:   types.StringValue(record.Tag),
			Value: types.StringValue(record.Value),
		}
	}

	var convertDiags diag.Diagnostics
	config.CAA, convertDiags = types.ListValueFrom(ctx, config.CAA.ElementType(ctx), caa)
	if convertDiags.HasError() {
		resp.Diagnostics.Append(convertDiags...)
		return
	}

	config.ID = config.Domain
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type caaRecordSetConfig struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	CAA    types.List   `tfsdk:"caa"` //caaBlockConfig
}

type caaBlockConfig struct {
	Flags types.Int64  `tfsdk:"flags"`
	Tag   types.String `tfsdk:"tag"`
	Value types.String `tfsdk:"value"`
}
