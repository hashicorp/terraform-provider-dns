// Copyright IBM Corp. 2017, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type CodeMigrationServer struct{}

// The actual code migrations
func (c CodeMigrationServer) GetCodeMigrations(ctx context.Context, req *tfprotov6.GetCodeMigrationsRequest) (*tfprotov6.GetCodeMigrationsResponse, error) {
	return &tfprotov6.GetCodeMigrationsResponse{
		CodeMigrations: []tfprotov6.CodeMigration{
			{
				TypeName: "dns_a_record_set",
				Name:     "strip_leading_zeroes",
				Migration: tfprotov6.Migration_TransformAttr{
					// addresses is a set of strings that are all IPv4 addresses
					TargetAttrPath: tftypes.NewAttributePath().WithAttributeName("addresses"),
					// The only input to this provider-defined function is a set(string), which should be "addresses" from above
					FunctionName: "provider::dns::strip_leading_zeroes",
				},
			},
			{
				TypeName: "dns_aaaa_record_set",
				Name:     "strip_leading_zeroes",
				Migration: tfprotov6.Migration_TransformAttr{
					// addresses is a set of strings that are all IPv6 addresses
					TargetAttrPath: tftypes.NewAttributePath().WithAttributeName("addresses"),
					// The only input to this provider-defined function is a set(string), which should be "addresses" from above
					FunctionName: "provider::dns::strip_leading_zeroes",
				},
			},
		},
	}, nil
}

func (c CodeMigrationServer) GetProviderSchema(ctx context.Context, req *tfprotov6.GetProviderSchemaRequest) (*tfprotov6.GetProviderSchemaResponse, error) {
	// Just a little cheeky trick to ensure they all have the same provider schema implementation w/o duplicating :P
	fwProvider := providerserver.NewProtocol6(NewFrameworkProvider())()
	resp, err := fwProvider.GetProviderSchema(ctx, req)
	if err != nil {
		return nil, err
	}

	return &tfprotov6.GetProviderSchemaResponse{
		Provider: resp.Provider,
	}, nil
}

// All of the RPC implementations that need to return an empty response
func (c CodeMigrationServer) ConfigureProvider(context.Context, *tfprotov6.ConfigureProviderRequest) (*tfprotov6.ConfigureProviderResponse, error) {
	return &tfprotov6.ConfigureProviderResponse{}, nil
}

func (c CodeMigrationServer) GetFunctions(context.Context, *tfprotov6.GetFunctionsRequest) (*tfprotov6.GetFunctionsResponse, error) {
	return &tfprotov6.GetFunctionsResponse{}, nil
}

func (c CodeMigrationServer) GetMetadata(context.Context, *tfprotov6.GetMetadataRequest) (*tfprotov6.GetMetadataResponse, error) {
	return &tfprotov6.GetMetadataResponse{}, nil
}

func (c CodeMigrationServer) GetResourceIdentitySchemas(context.Context, *tfprotov6.GetResourceIdentitySchemasRequest) (*tfprotov6.GetResourceIdentitySchemasResponse, error) {
	return &tfprotov6.GetResourceIdentitySchemasResponse{}, nil
}

func (c CodeMigrationServer) ValidateProviderConfig(context.Context, *tfprotov6.ValidateProviderConfigRequest) (*tfprotov6.ValidateProviderConfigResponse, error) {
	return &tfprotov6.ValidateProviderConfigResponse{}, nil
}

func (c CodeMigrationServer) StopProvider(context.Context, *tfprotov6.StopProviderRequest) (*tfprotov6.StopProviderResponse, error) {
	return &tfprotov6.StopProviderResponse{}, nil
}

// All of the following RPCs aren't called unless this provider server implements a resource/ephemeral/data source/etc.
func (c CodeMigrationServer) ApplyResourceChange(context.Context, *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) CallFunction(context.Context, *tfprotov6.CallFunctionRequest) (*tfprotov6.CallFunctionResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) CloseEphemeralResource(context.Context, *tfprotov6.CloseEphemeralResourceRequest) (*tfprotov6.CloseEphemeralResourceResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ImportResourceState(context.Context, *tfprotov6.ImportResourceStateRequest) (*tfprotov6.ImportResourceStateResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) MoveResourceState(context.Context, *tfprotov6.MoveResourceStateRequest) (*tfprotov6.MoveResourceStateResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) OpenEphemeralResource(context.Context, *tfprotov6.OpenEphemeralResourceRequest) (*tfprotov6.OpenEphemeralResourceResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) PlanResourceChange(context.Context, *tfprotov6.PlanResourceChangeRequest) (*tfprotov6.PlanResourceChangeResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ReadDataSource(context.Context, *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ReadResource(context.Context, *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) RenewEphemeralResource(context.Context, *tfprotov6.RenewEphemeralResourceRequest) (*tfprotov6.RenewEphemeralResourceResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) UpgradeResourceIdentity(context.Context, *tfprotov6.UpgradeResourceIdentityRequest) (*tfprotov6.UpgradeResourceIdentityResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) UpgradeResourceState(context.Context, *tfprotov6.UpgradeResourceStateRequest) (*tfprotov6.UpgradeResourceStateResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ValidateDataResourceConfig(context.Context, *tfprotov6.ValidateDataResourceConfigRequest) (*tfprotov6.ValidateDataResourceConfigResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ValidateEphemeralResourceConfig(context.Context, *tfprotov6.ValidateEphemeralResourceConfigRequest) (*tfprotov6.ValidateEphemeralResourceConfigResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) ValidateResourceConfig(context.Context, *tfprotov6.ValidateResourceConfigRequest) (*tfprotov6.ValidateResourceConfigResponse, error) {
	panic("unimplemented")
}

func (c CodeMigrationServer) GenerateResourceConfig(context.Context, *tfprotov6.GenerateResourceConfigRequest) (*tfprotov6.GenerateResourceConfigResponse, error) {
	panic("unimplemented")
}
