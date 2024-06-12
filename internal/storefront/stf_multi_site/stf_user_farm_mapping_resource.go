// Copyright © 2024. Citrix Systems, Inc.
package stf_multi_site

import (
	"context"
	"fmt"
	"strings"

	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &stfUserFarmMappingResource{}
	_ resource.ResourceWithConfigure   = &stfUserFarmMappingResource{}
	_ resource.ResourceWithImportState = &stfUserFarmMappingResource{}
)

// stfUserFarmMappingResource is a helper function to simplify the provider implementation.
func NewSTFUserFarmMappingResource() resource.Resource {
	return &stfUserFarmMappingResource{}
}

// stfUserFarmMappingResource is the resource implementation.
type stfUserFarmMappingResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the resource type name.
func (r *stfUserFarmMappingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_stf_user_farm_mapping"
}

// Configure adds the provider configured client to the resource.
func (r *stfUserFarmMappingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Create implements resource.Resource.
func (r *stfUserFarmMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan STFUserFarmMappingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create and Get new STF UserFarmMapping
	storeVirtualPath := plan.VirtualPath.ValueString()
	storeVirtualPathNullableString := citrixstorefront.NewNullableString(&storeVirtualPath)

	userFarmMappingName := plan.Name.ValueString()
	userFarmMappingNameNullableString := citrixstorefront.NewNullableString(&userFarmMappingName)

	groupMembers := BuildSTFUserFarmMappingGroupList(ctx, &resp.Diagnostics, plan.GroupMembers)
	equivalentFarmSets := BuildSTFEquivalentFarmSetRequestModelList(ctx, &resp.Diagnostics, plan.EquivalentFarmSets)

	getRequest := r.client.StorefrontClient.MultiSiteSF.STFMultiSiteGetUserFarmMapping(ctx, *storeVirtualPathNullableString, *userFarmMappingNameNullableString)
	_, err := getRequest.Execute()
	if err == nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Failed to create StoreFront UserFarmMapping `%s`", userFarmMappingName),
			fmt.Sprintf("StoreFront UserFarmMapping with name `%s` already exists", userFarmMappingName),
		)
		return
	} else if !strings.EqualFold(err.Error(), util.NOT_EXIST) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error verify the existence of StoreFront UserFarmMapping `%s`", userFarmMappingName),
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return
	}

	createdNewSTFUserFarmMapping, err := CreateAndGetNewSTFUserFarmMapping(ctx, &resp.Diagnostics, r.client, storeVirtualPath, userFarmMappingName, equivalentFarmSets, groupMembers)
	if err != nil {
		return
	}

	plan.RefreshPropertyValues(ctx, &resp.Diagnostics, createdNewSTFUserFarmMapping)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.ResourceWithConfigure.
func (r *stfUserFarmMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Get current state
	var state STFUserFarmMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storeVirtualPath := state.VirtualPath.ValueString()
	storeVirtualPathNullableString := citrixstorefront.NewNullableString(&storeVirtualPath)

	userFarmMappingName := state.Name.ValueString()
	userFarmMappingNameNullableString := citrixstorefront.NewNullableString(&userFarmMappingName)

	getRequest := r.client.StorefrontClient.MultiSiteSF.STFMultiSiteGetUserFarmMapping(ctx, *storeVirtualPathNullableString, *userFarmMappingNameNullableString)
	getResult, err := getRequest.Execute()
	if err != nil {
		if strings.EqualFold(err.Error(), util.NOT_EXIST) {
			resp.Diagnostics.AddWarning(
				"UserFarmMapping not found",
				"UserFarmMapping Service was not found and will be removed from the state file. An apply action will result in the creation of a new resource.",
			)
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error fetch details of StoreFront UserFarmMapping",
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return
	}

	state.RefreshPropertyValues(ctx, &resp.Diagnostics, getResult)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *stfUserFarmMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete implements resource.ResourceWithConfigure.
func (r *stfUserFarmMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state STFUserFarmMappingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storeVirtualPath := state.VirtualPath.ValueString()
	storeVirtualPathNullableString := citrixstorefront.NewNullableString(&storeVirtualPath)

	userFarmMappingName := state.Name.ValueString()
	userFarmMappingNameNullableString := citrixstorefront.NewNullableString(&userFarmMappingName)

	// Delete existing STF UserFarmMapping
	deleteRequest := r.client.StorefrontClient.MultiSiteSF.STFMultiSiteRemoveUserFarmMapping(ctx, *storeVirtualPathNullableString, *userFarmMappingNameNullableString)
	_, err := deleteRequest.Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Error deleting StoreFront UserFarmMapping `%s` with virtual path `%s`", userFarmMappingName, storeVirtualPath),
			"\nError message: "+err.Error(),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *stfUserFarmMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	idSegments := strings.SplitN(req.ID, ",", 2)

	if (len(idSegments) != 2) || (idSegments[0] == "" || idSegments[1] == "") {
		resp.Diagnostics.AddError(
			"Invalid Import Identifier",
			fmt.Sprintf("Expected format: `store_virtual_path,name`, got: `%q`", req.ID),
		)
		return
	}

	// Retrieve import ID and save to id attribute
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("store_virtual_path"), idSegments[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idSegments[1])...)
}

func BuildSTFUserFarmMappingGroupList(ctx context.Context, diagnostics *diag.Diagnostics, plannedGroupMembers types.List) []citrixstorefront.STFUserFarmMappingGroup {
	groupMembersInput := util.ObjectListToTypedArray[UserFarmMappingGroup](ctx, diagnostics, plannedGroupMembers)
	groupMembers := []citrixstorefront.STFUserFarmMappingGroup{}

	for _, groupMemberInput := range groupMembersInput {
		groupMember := citrixstorefront.STFUserFarmMappingGroup{}
		groupMember.SetGroupName(groupMemberInput.GroupName.ValueString())
		groupMember.SetAccountSid(groupMemberInput.AccountSid.ValueString())

		groupMembers = append(groupMembers, groupMember)
	}

	return groupMembers
}

func BuildSTFEquivalentFarmSetRequestModelList(ctx context.Context, diagnostics *diag.Diagnostics, plannedEquivalentFarmSets types.List) []citrixstorefront.STFEquivalentFarmSetRequestModel {
	equivalentFarmSetsInput := util.ObjectListToTypedArray[EquivalentFarmSet](ctx, diagnostics, plannedEquivalentFarmSets)
	equivalentFarmSets := []citrixstorefront.STFEquivalentFarmSetRequestModel{}

	for _, equivalentFarmSetInput := range equivalentFarmSetsInput {
		equivalentFarmSet := citrixstorefront.STFEquivalentFarmSetRequestModel{}
		equivalentFarmSet.SetName(equivalentFarmSetInput.Name.ValueString())
		equivalentFarmSet.SetAggregationGroupName(equivalentFarmSetInput.AggregationGroupName.ValueString())
		equivalentFarmSet.SetFarmsAreIdentical(equivalentFarmSetInput.FarmsAreIdentical.ValueBool())
		equivalentFarmSet.SetLoadBalanceMode(equivalentFarmSetInput.LoadBalanceMode.ValueString())
		equivalentFarmSet.SetPrimaryFarms(util.StringListToStringArray(ctx, diagnostics, equivalentFarmSetInput.PrimaryFarms))
		equivalentFarmSet.SetBackupFarms(util.StringListToStringArray(ctx, diagnostics, equivalentFarmSetInput.BackupFarms))

		equivalentFarmSets = append(equivalentFarmSets, equivalentFarmSet)
	}

	return equivalentFarmSets
}

func CreateAndGetNewSTFUserFarmMapping(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, virtualPath string, name string, equivalentFarmSets []citrixstorefront.STFEquivalentFarmSetRequestModel, groupMembers []citrixstorefront.STFUserFarmMappingGroup) (citrixstorefront.STFUserFarmMappingResponseModel, error) {
	virtualPathNullableString := citrixstorefront.NewNullableString(&virtualPath)
	nameNullableString := citrixstorefront.NewNullableString(&name)
	addRequest := client.StorefrontClient.MultiSiteSF.STFMultiSiteAddUserFarmMapping(ctx, *virtualPathNullableString, *nameNullableString, equivalentFarmSets, groupMembers)
	_, err := addRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error adding StoreFront UserFarmMapping",
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return citrixstorefront.STFUserFarmMappingResponseModel{}, err
	}

	getRequest := client.StorefrontClient.MultiSiteSF.STFMultiSiteGetUserFarmMapping(ctx, *virtualPathNullableString, *nameNullableString)
	getResult, err := getRequest.Execute()
	if err != nil {
		diagnostics.AddError(
			"Error fetch details of StoreFront UserFarmMapping",
			fmt.Sprintf("Error Message: %s", err.Error()),
		)
		return citrixstorefront.STFUserFarmMappingResponseModel{}, err
	}
	return getResult, nil
}
