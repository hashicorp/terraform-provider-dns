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
	_ datasource.DataSource = (*dnsMXRecordSetDataSource)(nil)
)

func NewDnsMXRecordSetDataSource() datasource.DataSource {
	return &dnsMXRecordSetDataSource{}
}

type dnsMXRecordSetDataSource struct{}

func (d *dnsMXRecordSetDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mx_record_set"
}

func (d *dnsMXRecordSetDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get DNS MX records for a domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "Domain to look up.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always set to the domain",
			},
			"mx": schema.ListAttribute{
				Description: "A list of records. They are sorted by ascending preference then alphabetically by " +
					"exchange to stay consistent across runs.",
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"preference": types.Int64Type,
						"exchange":   types.StringType,
					},
				},
			},
		},
	}
}

func (d *dnsMXRecordSetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config mxRecordSetConfig

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := config.Domain.ValueString()
	records, err := net.LookupMX(domain)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("error looking up MX records for %q: ", domain), err.Error())
		return
	}

	// Sort by preference ascending, and host alphabetically
	sort.Slice(records, func(i, j int) bool {
		if records[i].Pref < records[j].Pref {
			return true
		}
		if records[i].Pref > records[j].Pref {
			return false
		}
		return records[i].Host < records[j].Host
	})

	mx := make([]mxBlockConfig, len(records))
	for i, record := range records {
		mx[i] = mxBlockConfig{
			Preference: types.Int64Value(int64(record.Pref)),
			Exchange:   types.StringValue(record.Host),
		}
	}

	var convertDiags diag.Diagnostics
	config.MX, convertDiags = types.ListValueFrom(ctx, config.MX.ElementType(ctx), mx)
	if convertDiags.HasError() {
		resp.Diagnostics.Append(convertDiags...)
		return
	}

	config.ID = config.Domain
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

type mxRecordSetConfig struct {
	ID     types.String `tfsdk:"id"`
	Domain types.String `tfsdk:"domain"`
	MX     types.List   `tfsdk:"mx"` //mxBlockConfig
}

type mxBlockConfig struct {
	Preference types.Int64  `tfsdk:"preference"`
	Exchange   types.String `tfsdk:"exchange"`
}
