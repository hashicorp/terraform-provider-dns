// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/miekg/dns"
)

// testProviderSchemaConfig returns a tfsdk.Config for the given schema.
func testProviderSchemaConfig(t *testing.T, ctx context.Context, schema schema.Schema, values map[string]attr.Value) tfsdk.Config {
	t.Helper()

	schemaType := schema.Type()
	objectType, ok := schemaType.(types.ObjectType)

	if !ok {
		t.Fatalf("expected schema type of types.Object, got: %T", schemaType)
	}

	objectValue := types.ObjectValueMust(objectType.AttributeTypes(), values)
	tfTypesValue, err := objectValue.ToTerraformValue(ctx)

	if err != nil {
		t.Fatalf("unexpected error converting to tftypes: %s", err)
	}

	return tfsdk.Config{
		Schema: schema,
		Raw:    tfTypesValue,
	}
}

// nolint:paralleltest // includes environment variable testing
func TestDnsProviderConfigure(t *testing.T) {
	ctx := context.Background()
	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	testProvider := NewFrameworkProvider()

	testProvider.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}

	schema := schemaResp.Schema

	// Prevent external environment variable values from affecting this test
	t.Setenv("DNS_UPDATE_KEYALGORITHM", "")
	t.Setenv("DNS_UPDATE_KEYNAME", "")
	t.Setenv("DNS_UPDATE_KEYSECRET", "")
	t.Setenv("DNS_UPDATE_KEYTAB", "")
	t.Setenv("DNS_UPDATE_PASSWORD", "")
	t.Setenv("DNS_UPDATE_PORT", "")
	t.Setenv("DNS_UPDATE_REALM", "")
	t.Setenv("DNS_UPDATE_RETRIES", "")
	t.Setenv("DNS_UPDATE_SERVER", "")
	t.Setenv("DNS_UPDATE_TRANSPORT", "")
	t.Setenv("DNS_UPDATE_TIMEOUT", "")
	t.Setenv("DNS_UPDATE_USERNAME", "")
	t.Setenv("DNS_UPDATE_EDNS_MSG_SIZE", "")

	testCases := map[string]struct {
		env      map[string]string
		request  provider.ConfigureRequest
		expected *provider.ConfigureResponse
	}{
		"no-config-or-env": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-port-config": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Value(1053),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    ":1053",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-port-config-and-env": {
			env: map[string]string{
				"DNS_UPDATE_PORT": "2053",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Value(1053),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    ":1053",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-port-env": {
			env: map[string]string{
				"DNS_UPDATE_PORT": "1053",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    ":1053",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-port-env-invalid": {
			env: map[string]string{
				"DNS_UPDATE_PORT": "not-an-int",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid DNS_UPDATE_PORT environment variable:",
						"strconv.Atoi: parsing \"not-an-int\": invalid syntax",
					),
				},
				ResourceData: nil,
			},
		},
		"update-server-config": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringValue("example.com"),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    "example.com:53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-server-config-and-env": {
			env: map[string]string{
				"DNS_UPDATE_SERVER": "example.org",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringValue("example.com"),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    "example.com:53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-server-env": {
			env: map[string]string{
				"DNS_UPDATE_SERVER": "example.com",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "udp",
					},
					retries:     3,
					srv_addr:    "example.com:53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-config-duration": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringValue("5s"),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						Timeout: 5 * time.Second,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-config-number": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringValue("5"),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						Timeout: 5 * time.Second,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-config-and-env": {
			env: map[string]string{
				"DNS_UPDATE_TIMEOUT": "10",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringValue("5"),
									"transport":     types.StringNull(),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						Timeout: 5 * time.Second,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-env-duration": {
			env: map[string]string{
				"DNS_UPDATE_TIMEOUT": "5s",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						Timeout: 5 * time.Second,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-env-number": {
			env: map[string]string{
				"DNS_UPDATE_TIMEOUT": "5",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						Timeout: 5 * time.Second,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-timeout-env-invalid": {
			env: map[string]string{
				"DNS_UPDATE_TIMEOUT": "not-an-int",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Invalid DNS Provider Timeout Value",
						"Timeout cannot be parsed as an integer: strconv.Atoi: parsing \"not-an-int\": invalid syntax",
					),
				},
				ResourceData: nil,
			},
		},
		"update-transport-config": {
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringValue("tcp"),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "tcp",
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "tcp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-transport-config-and-env": {
			env: map[string]string{
				"DNS_UPDATE_TRANSPORT": "tcp6",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListValueMust(
						providerUpdateModel{}.objectType(),
						[]attr.Value{
							types.ObjectValueMust(
								providerUpdateModel{}.objectAttributeTypes(),
								map[string]attr.Value{
									"gssapi":        types.ListNull(providerGssapiModel{}.objectType()),
									"key_name":      types.StringNull(),
									"key_algorithm": types.StringNull(),
									"key_secret":    types.StringNull(),
									"port":          types.Int64Null(),
									"server":        types.StringNull(),
									"retries":       types.Int64Null(),
									"timeout":       types.StringNull(),
									"transport":     types.StringValue("tcp"),
								},
							),
						},
					),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "tcp",
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "tcp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-transport-env": {
			env: map[string]string{
				"DNS_UPDATE_TRANSPORT": "tcp",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net: "tcp",
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "tcp",
					ednsMsgSize: 4096,
				},
			},
		},
		"update-edns-msg-size-env": {
			env: map[string]string{
				"DNS_UPDATE_EDNS_MSG_SIZE": "65535",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						UDPSize: 0,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 65535,
				},
			},
		},
		"update-edns-msg-size-env-too-large": {
			env: map[string]string{
				"DNS_UPDATE_EDNS_MSG_SIZE": "70000",
			},
			request: provider.ConfigureRequest{
				Config: testProviderSchemaConfig(t, ctx, schema, map[string]attr.Value{
					"update": types.ListNull(providerUpdateModel{}.objectType()),
				}),
			},
			expected: &provider.ConfigureResponse{
				ResourceData: &DNSClient{
					c: &dns.Client{
						Net:     "udp",
						UDPSize: 0,
					},
					retries:     3,
					srv_addr:    ":53",
					transport:   "udp",
					ednsMsgSize: 65535,
				},
			},
		},
	}

	for name, testCase := range testCases {

		// nolint:paralleltest // includes environment variable testing
		t.Run(name, func(t *testing.T) {
			for envKey, envValue := range testCase.env {
				t.Setenv(envKey, envValue)
			}

			got := &provider.ConfigureResponse{}

			testProvider.Configure(ctx, testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected, cmp.AllowUnexported(DNSClient{}), cmpopts.IgnoreUnexported(dns.Client{})); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
