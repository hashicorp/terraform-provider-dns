// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ action.Action               = (*dnsAction)(nil)
	_ action.ActionWithModifyPlan = (*dnsAction)(nil)
)

func NewDnsAction() action.Action {
	return &dnsAction{}
}

type dnsAction struct{}

func (d *dnsAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_do_thing"
}

func (d *dnsAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	sdkV2Schema := resourceDnsARecordSet()

	resp.Schema = schema.LifecycleSchema{
		ExecutionOrder: schema.ExecutionOrderAfter,
		LinkedResource: schema.RawV5LinkedResource{
			TypeName:    "dns_a_record_set",
			Description: "Record set to perform an action against",
			Schema:      sdkV2Schema.ProtoSchema(ctx),
		},
		Attributes: map[string]schema.Attribute{
			"new_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// The plan modification here just takes the new_id field from the action config and plans to update the "id" field on the linked resource.
//
// This eventually won't be valid in TF core because "id" is optional/computed, but one can dream :).
func (d *dnsAction) ModifyPlan(ctx context.Context, req action.ModifyPlanRequest, resp *action.ModifyPlanResponse) {
	var newID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("new_id"), &newID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.LinkedResources[0].Plan.SetAttribute(ctx, path.Root("id"), newID)...)
}

func (d *dnsAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var newID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("new_id"), &newID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, newID.String())

	// TODO: eventually set on response linked resource
	// resp.Diagnostics.Append(resp.LinkedResources[0].State.SetAttribute(ctx, path.Root("id"), newID)...)
}
