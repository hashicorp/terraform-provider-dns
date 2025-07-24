// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
)

// Note: At the moment, this is just testing that we can decode a linked resource defined in SDKv2, in action
// logic that is defined in Framework. All setting of data in Framework is possible due to the internal conversion
// that linked resource schemas undergo during the PlanAction/InvokeAction RPCs.
func TestAccDnsARecordSet_Action(t *testing.T) {
	ctx := context.Background()

	actionTypeName := "dns_do_thing"
	dnsProvider := GetDNSProvider(t)
	actionConfigType := GetDNSActionSchemaType(t, ctx)
	InitializeProvider(t, ctx, dnsProvider)

	testCases := map[string]struct {
		config               map[string]tftypes.Value
		linkedResourcePlan   map[string]tftypes.Value
		linkedResourceConfig map[string]tftypes.Value
		expectedValidateResp tfprotov5.ValidateActionConfigResponse
		expectedPlanResp     tfprotov5.PlanActionResponse
		expectedInvokeEvents []tfprotov5.InvokeActionEvent
	}{
		"test": {
			// terraform-plugin-go equivalent of:
			//
			//	resource "dns_a_record_set" "foo" {
			//    zone = "example.com."
			//    name = "foo"
			//    addresses = ["192.168.000.001", "192.168.000.002"]
			//    ttl = 300
			//
			//    lifecycle {
			//      action_trigger {
			//        events = ["after_create"]
			//        actions = [action.dns_do_thing.test]
			//      }
			//    }
			//	}
			//
			//	action "dns_do_thing" "test" {
			//    linked_resource = resource.dns_a_record_set.foo
			//	  config {
			//	    new_id = "hello-new-id-123"
			//	  }
			//	}
			config: map[string]tftypes.Value{
				"new_id": tftypes.NewValue(tftypes.String, "hello-new-id-123"),
			},
			linkedResourceConfig: map[string]tftypes.Value{
				"zone": tftypes.NewValue(tftypes.String, "example.com."),
				"name": tftypes.NewValue(tftypes.String, "foo"),
				"addresses": tftypes.NewValue(
					tftypes.Set{ElementType: tftypes.String},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.String, "192.168.000.001"),
						tftypes.NewValue(tftypes.String, "192.168.000.002"),
					},
				),
				"ttl": tftypes.NewValue(tftypes.Number, 300),
				"id":  tftypes.NewValue(tftypes.String, nil),
			},
			// Same as config, but the computed attribute is unknown
			linkedResourcePlan: map[string]tftypes.Value{
				"zone": tftypes.NewValue(tftypes.String, "example.com."),
				"name": tftypes.NewValue(tftypes.String, "foo"),
				"addresses": tftypes.NewValue(
					tftypes.Set{ElementType: tftypes.String},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.String, "192.168.000.001"),
						tftypes.NewValue(tftypes.String, "192.168.000.002"),
					},
				),
				"ttl": tftypes.NewValue(tftypes.Number, 300),
				"id":  tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // The linked resource originally plans unknown for the computed attribute
			},
			expectedPlanResp: tfprotov5.PlanActionResponse{
				LinkedResources: []*tfprotov5.PlannedLinkedResource{
					{
						PlannedState: dynamicValueMust(
							tftypes.NewValue(
								resourceDnsARecordSet().ProtoSchema(ctx).ValueType(),
								map[string]tftypes.Value{
									"zone": tftypes.NewValue(tftypes.String, "example.com."),
									"name": tftypes.NewValue(tftypes.String, "foo"),
									"addresses": tftypes.NewValue(
										tftypes.Set{ElementType: tftypes.String},
										[]tftypes.Value{
											tftypes.NewValue(tftypes.String, "192.168.000.001"),
											tftypes.NewValue(tftypes.String, "192.168.000.002"),
										},
									),
									"ttl": tftypes.NewValue(tftypes.Number, 300),
									"id":  tftypes.NewValue(tftypes.String, "hello-new-id-123"), // Modification of linked resource from the action plan
								},
							),
						),
					},
				},
			},
			expectedInvokeEvents: []tfprotov5.InvokeActionEvent{
				{
					Type: tfprotov5.CompletedInvokeActionEventType{},
				},
			},
			expectedValidateResp: tfprotov5.ValidateActionConfigResponse{},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			testActionConfig, err := tfprotov5.NewDynamicValue(actionConfigType, tftypes.NewValue(actionConfigType, tc.config))
			if err != nil {
				t.Fatal(err)
			}

			validateResp, err := dnsProvider.ValidateActionConfig(ctx, &tfprotov5.ValidateActionConfigRequest{
				ActionType: actionTypeName,
				Config:     &testActionConfig,
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(*validateResp, tc.expectedValidateResp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			// Don't plan/invoke if we had diagnostics during validate
			if len(validateResp.Diagnostics) > 0 {
				return
			}

			planResp, err := dnsProvider.PlanAction(ctx, &tfprotov5.PlanActionRequest{
				ActionType: actionTypeName,
				Config:     &testActionConfig,
				LinkedResources: []*tfprotov5.ProposedLinkedResource{
					{
						PriorState: dynamicValueMust(
							tftypes.NewValue(
								resourceDnsARecordSet().ProtoSchema(ctx).ValueType(),
								nil, // Just assuming it's a create for now
							),
						),
						PlannedState: dynamicValueMust(
							tftypes.NewValue(
								resourceDnsARecordSet().ProtoSchema(ctx).ValueType(),
								tc.linkedResourcePlan,
							),
						),
						Config: dynamicValueMust(
							tftypes.NewValue(
								resourceDnsARecordSet().ProtoSchema(ctx).ValueType(),
								tc.linkedResourceConfig,
							),
						),
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(*planResp, tc.expectedPlanResp); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			// Don't invoke if we had diagnostics during plan
			if len(planResp.Diagnostics) > 0 {
				return
			}

			invokeResp, err := dnsProvider.InvokeAction(ctx, &tfprotov5.InvokeActionRequest{
				ActionType: actionTypeName,
				Config:     &testActionConfig,
			})
			if err != nil {
				t.Fatal(err)
			}

			// Grab all the events
			events := slices.Collect(invokeResp.Events)
			if diff := cmp.Diff(events, tc.expectedInvokeEvents); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func dynamicValueMust(value tftypes.Value) *tfprotov5.DynamicValue {
	dynamicValue, err := tfprotov5.NewDynamicValue(value.Type(), value)

	if err != nil {
		panic(err)
	}

	return &dynamicValue
}

func GetDNSProvider(t *testing.T) tfprotov5.ProviderServerWithActions {
	t.Helper()

	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(NewFrameworkProvider()),
		New().GRPCProvider,
	}

	provider, err := tf5muxserver.NewMuxServer(context.Background(), providers...)
	if err != nil {
		t.Fatal(err)
	}

	return provider.ProviderServer().(tfprotov5.ProviderServerWithActions)
}

func GetDNSActionSchemaType(t *testing.T, ctx context.Context) tftypes.Type {
	t.Helper()

	actionSchemaResp := action.SchemaResponse{}
	NewDnsAction().Schema(ctx, action.SchemaRequest{}, &actionSchemaResp)

	return actionSchemaResp.Schema.Type().TerraformType(ctx)
}

func InitializeProvider(t *testing.T, ctx context.Context, DNSProvider tfprotov5.ProviderServer) {
	t.Helper()

	// Just for show in our case here, since the DNS provider isn't muxed, so technically you could skip this.
	// Mimicking what core will eventually do.
	schemaResp, err := DNSProvider.GetProviderSchema(ctx, &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: Since we already have the schema/type in the same codebase, we don't need to read the schema
	// so this is just asserting that GetProviderSchema is actually returning the action schema as expected.
	if len(schemaResp.ActionSchemas) == 0 {
		t.Fatalf("expected to find action schemas and didn't find any!")
	}

	nullObject, err := tfprotov5.NewDynamicValue(tftypes.Object{}, tftypes.NewValue(tftypes.Object{}, nil))
	if err != nil {
		t.Fatal(err)
	}

	// Just for show in our case here, since the DNS provider has no configuration.
	// Mimicking what core will eventually do.
	_, err = DNSProvider.ConfigureProvider(ctx, &tfprotov5.ConfigureProviderRequest{
		Config: &nullObject,
	},
	)
	if err != nil {
		t.Fatal(err)
	}
}
