// Copyright © 2024. Citrix Systems, Inc.

package util

import (
	"context"
	"fmt"
	"slices"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	datasourceSchema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ RefreshableListItemWithAttributes[citrixorchestration.NameValueStringPairModel] = NameValueStringPairModel{}

// Terraform model for name value string pair
type NameValueStringPairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// RefreshListItem implements RefreshableListItemWithAttributes.
func (r NameValueStringPairModel) RefreshListItem(ctx context.Context, diag *diag.Diagnostics, retmote citrixorchestration.NameValueStringPairModel) ResourceModelWithAttributes {
	r.Name = types.StringValue(retmote.GetName())
	r.Value = types.StringValue(retmote.GetValue())

	return r
}

func (r NameValueStringPairModel) GetKey() string {
	return r.Name.ValueString()
}

func (r NameValueStringPairModel) GetSchema() schema.NestedAttributeObject {
	return schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Metadata name.",
				Required:    true,
			},
			"value": schema.StringAttribute{
				Description: "Metadata value.",
				Required:    true,
			},
		},
	}
}

func (r NameValueStringPairModel) GetAttributes() map[string]schema.Attribute {
	return NameValueStringPairModel{}.GetSchema().Attributes
}

func (r NameValueStringPairModel) GetDataSourceSchema() datasourceSchema.NestedAttributeObject {
	return datasourceSchema.NestedAttributeObject{
		Attributes: map[string]datasourceSchema.Attribute{
			"name": datasourceSchema.StringAttribute{
				Description: "Metadata name.",
				Computed:    true,
			},
			"value": datasourceSchema.StringAttribute{
				Description: "Metadata value.",
				Computed:    true,
			},
		},
	}
}

func (r NameValueStringPairModel) GetDataSourceAttributes() map[string]datasourceSchema.Attribute {
	return NameValueStringPairModel{}.GetDataSourceSchema().Attributes
}

func (r NameValueStringPairModel) ValidateConfig(ctx context.Context, diagnostics *diag.Diagnostics, index int) bool {
	metadataName := r.Name.ValueString()
	if strings.EqualFold(metadataName, MetadataTerraformName) {
		diagnostics.AddAttributeError(
			path.Root("metadata").AtListIndex(index),
			"Incorrect Attribute Configuration",
			fmt.Sprintf("%s is a reserved metadata name and cannot be used. Please use a different name.", MetadataTerraformName),
		)

		return false
	}

	metadataNameLower := strings.ToLower(metadataName)
	if strings.HasPrefix(metadataNameLower, MetadataCitrixPrefix) ||
		strings.HasPrefix(metadataNameLower, MetadataImageManagementPrepPrefix) ||
		strings.HasPrefix(metadataNameLower, MetadataTaskDataPrefix) ||
		strings.HasPrefix(metadataNameLower, MetadataTaskStatePrefix) {
		diagnostics.AddAttributeError(
			path.Root("metadata").AtListIndex(index),
			"Incorrect Attribute Configuration",
			fmt.Sprintf("%s has a reserved metadata prefix name and cannot be used. Please use a different name.", metadataName),
		)
	}

	return true
}

func ValidateMetadataConfig(ctx context.Context, diagnostics *diag.Diagnostics, metadata []NameValueStringPairModel) bool {
	metadataMap := map[string]string{}
	for index, md := range metadata {
		if !md.ValidateConfig(ctx, diagnostics, index) {
			return false
		}

		metadataName := strings.ToLower(md.Name.ValueString())
		if metadataValue, exists := metadataMap[metadataName]; exists {
			diagnostics.AddAttributeError(
				path.Root("metadata").AtListIndex(index),
				"Incorrect Attribute Configuration",
				fmt.Sprintf("Metadata name %s already exists with value %s", md.Name.ValueString(), metadataValue),
			)

			return false
		}
		metadataMap[metadataName] = md.Value.ValueString()
	}
	return true
}

func GetMetadataListSchema(resource string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Description:  fmt.Sprintf("Metadata for the %s.", resource),
		Optional:     true,
		NestedObject: NameValueStringPairModel{}.GetSchema(),
		Validators: []validator.List{
			listvalidator.SizeAtLeast(1),
		},
	}
}

// <summary>
// Helper function to parse an array of name value pairs in terraform model to an array of name value pairs in client model
// </summary>
// <param name="stringPairs">Original string pair array in terraform model</param>
// <returns>String pair array in client model</returns>
func ParseNameValueStringPairToClientModel(stringPairs []NameValueStringPairModel) []citrixorchestration.NameValueStringPairModel {
	var res = []citrixorchestration.NameValueStringPairModel{}
	for _, stringPair := range stringPairs {
		name := stringPair.Name.ValueString()
		value := stringPair.Value.ValueString()
		res = append(res, citrixorchestration.NameValueStringPairModel{
			Name:  *citrixorchestration.NewNullableString(&name),
			Value: *citrixorchestration.NewNullableString(&value),
		})
	}
	return res
}

// <summary>
// Helper function to parse an array of name value pairs in client model to an array of name value pairs in terraform model
// </summary>
// <param name="stringPairs">Original string pair array in client model</param>
// <returns>String pair array in terraform model</returns>
func ParseNameValueStringPairToPluginModel(stringPairs []citrixorchestration.NameValueStringPairModel) []NameValueStringPairModel {
	var res = []NameValueStringPairModel{}
	for _, stringPair := range stringPairs {
		res = append(res, NameValueStringPairModel{
			Name:  types.StringValue(stringPair.GetName()),
			Value: types.StringValue(stringPair.GetValue()),
		})
	}
	return res
}

// <summary>
// Helper function to append new name value pairs to an array of NameValueStringPairModel in place
// </summary>
// <param name="stringPairs">Original string pair array to append to</param>
// <param name="name">Name of the new string pair to be added</param>
// <param name="appendValue">Value of the new string pair to be added</param>
func AppendNameValueStringPair(stringPairs *[]citrixorchestration.NameValueStringPairModel, name string, appendValue string) {
	*stringPairs = append(*stringPairs, citrixorchestration.NameValueStringPairModel{
		Name:  *citrixorchestration.NewNullableString(&name),
		Value: *citrixorchestration.NewNullableString(&appendValue),
	})
}

func AppendTerraformMetadataInfo(stringPairs *[]citrixorchestration.NameValueStringPairModel) {
	AppendNameValueStringPair(stringPairs, MetadataTerraformName, MetadataTerrafomValue)
}

func GetMetadataRequestModel(ctx context.Context, diagnostics *diag.Diagnostics, planMetadata []NameValueStringPairModel) []citrixorchestration.NameValueStringPairModel {
	metadata := []citrixorchestration.NameValueStringPairModel{}

	AppendTerraformMetadataInfo(&metadata)
	if planMetadata != nil {
		additionalMetadata := ParseNameValueStringPairToClientModel(planMetadata)
		metadata = append(metadata, additionalMetadata...)
	}

	return metadata
}

// <summary>
// Helper function to delete metadata that is not a part of the plan. The value of the metadata is set to empty string to delete the KVP from the remote. (STUD-31858)
// </summary>
func GetUpdatedMetadataRequestModel(ctx context.Context, diagnostics *diag.Diagnostics, stateMetadata []NameValueStringPairModel, planMetadata []NameValueStringPairModel) []citrixorchestration.NameValueStringPairModel {
	metadata := GetMetadataRequestModel(ctx, diagnostics, planMetadata)
	metadataMap := make(map[string]bool)
	for _, item := range metadata {
		metadataMap[item.GetName()] = true
	}

	// Add metadata Name from state with empty value to delete them from the remote
	for _, pair := range stateMetadata {
		if _, exists := metadataMap[pair.Name.ValueString()]; !exists {
			AppendNameValueStringPair(&metadata, pair.Name.ValueString(), "")
		}
	}
	return metadata
}

// / <summary>
// / Helper function to include only metadata that are a part of the state
// / </summary>
func GetEffectiveMetadata(stateMetadata []NameValueStringPairModel, remoteMetadata []citrixorchestration.NameValueStringPairModel) []citrixorchestration.NameValueStringPairModel {
	if len(stateMetadata) == 0 {
		return []citrixorchestration.NameValueStringPairModel{}
	}

	stateMetadaMap := map[string]bool{}
	for _, stateMetadataItem := range stateMetadata {
		metadataName := strings.ToLower(stateMetadataItem.Name.ValueString())
		stateMetadaMap[metadataName] = true
	}

	return slices.DeleteFunc(remoteMetadata, func(data citrixorchestration.NameValueStringPairModel) bool {
		metadataName := strings.ToLower(data.GetName())
		_, exists := stateMetadaMap[metadataName]
		return !exists
	})
}

// MergeMetadata merges metadata from remote and plan, adds Terraform metadata, and ensures no duplicates.
func MergeMetadata(remoteMetadata []NameValueStringPairModel, planMetadata []NameValueStringPairModel) []citrixorchestration.NameValueStringPairModel {
	metadata := []citrixorchestration.NameValueStringPairModel{}
	metadataMap := make(map[string]bool)

	// Add plan metadata to metadata and track existing keys
	for _, item := range planMetadata {
		key := strings.ToLower(item.Name.ValueString())
		AppendNameValueStringPair(&metadata, item.Name.ValueString(), item.Value.ValueString())
		metadataMap[key] = true
	}

	// Add remote metadata if not already present in plan metadata
	for _, item := range remoteMetadata {
		key := strings.ToLower(item.Name.ValueString())
		if !metadataMap[key] {
			AppendNameValueStringPair(&metadata, item.Name.ValueString(), item.Value.ValueString())
			metadataMap[key] = true
		}
	}

	// Only append Terraform metadata if not already present
	if !metadataMap[MetadataTerraformName] {
		AppendTerraformMetadataInfo(&metadata)
	}

	return metadata
}
