// Copyright © 2023. Citrix Systems, Inc.

package policies

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/util"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &policySetResource{}
	_ resource.ResourceWithConfigure   = &policySetResource{}
	_ resource.ResourceWithImportState = &policySetResource{}
)

// NewPolicySetResource is a helper function to simplify the provider implementation.
func NewPolicySetResource() resource.Resource {
	return &policySetResource{}
}

// policySetResource is the resource implementation.
type policySetResource struct {
	client *citrixdaasclient.CitrixDaasClient
}

// Metadata returns the data source type name.
func (r *policySetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_set"
}

// Configure implements resource.ResourceWithConfigure.
func (r *policySetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*citrixdaasclient.CitrixDaasClient)
}

// Schema implements resource.Resource.
func (*policySetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a policy set and the policies within it. The order of the policies specified in this resource reflect the policy priority. This feature will be officially supported for On-Premises with DDC version 2402 and above and will be made available for Cloud soon.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "GUID identifier of the policy set.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy set.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Type of the policy set. Type can be one of `SitePolicies`, `DeliveryGroupPolicies`, `SiteTemplates`, or `CustomTemplates`.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"SitePolicies",
						"DeliveryGroupPolicies",
						"SiteTemplates",
						"CustomTemplates"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the policy set.",
				Optional:    true,
			},
			"scopes": schema.SetAttribute{
				ElementType: types.StringType,
				Description: "The names of the scopes for the policy set to apply on.",
				Required:    true,
			},
			"policies": schema.ListNestedAttribute{
				Description: "Ordered list of policies. The order of policies in the list determines the priority of the policies.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the policy.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the policy.",
							Optional:    true,
						},
						"is_enabled": schema.BoolAttribute{
							Description: "Indicate whether the policy is being enabled.",
							Required:    true,
						},
						"policy_settings": schema.SetNestedAttribute{
							Description: "Set of policy settings.",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "Name of the policy setting name.",
										Required:    true,
									},
									"use_default": schema.BoolAttribute{
										Description: "Indicate whether using default value for the policy setting.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "Value of the policy setting.",
										Required:    true,
									},
								},
							},
						},
						"policy_filters": schema.SetNestedAttribute{
							Description: "Set of policy filters.",
							Required:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Type of the policy filter. Type can be one of `AccessControl`, `BranchRepeater`, `ClientIP`, `ClientName`, `DesktopGroup`, `DesktopKind`, `OU`, `User`, and `DesktopTag`",
										Required:    true,
										Validators: []validator.String{
											stringvalidator.OneOf([]string{
												"AccessControl",
												"BranchRepeater",
												"ClientIP",
												"ClientName",
												"DesktopGroup",
												"DesktopKind",
												"OU",
												"User",
												"DesktopTag"}...),
										},
									},
									"data": schema.StringAttribute{
										Description: "Data of the policy filter.",
										Optional:    true,
									},
									"is_enabled": schema.BoolAttribute{
										Description: "Indicate whether the policy is being enabled.",
										Required:    true,
									},
									"is_allowed": schema.BoolAttribute{
										Description: "Indicate the filtered policy is allowed or denied if the filter condition is met.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},
			"is_assigned": schema.BoolAttribute{
				Description: "Indicate whether the policy set is being assigned to delivery groups.",
				Computed:    true,
			},
		},
	}
}

// Create implements resource.Resource.
func (r *policySetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySets, err := getPolicySets(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	for _, policySet := range policySets {
		if strings.EqualFold(policySet.GetName(), plan.Name.ValueString()) {
			resp.Diagnostics.AddError(
				"Error Creating Policy Set",
				"Policy Set with name "+plan.Name.ValueString()+" already exists",
			)
			return
		}
	}

	var createPolicySetRequestBody = &citrixorchestration.PolicySetRequest{}
	createPolicySetRequestBody.SetName(plan.Name.ValueString())
	createPolicySetRequestBody.SetDescription(plan.Description.ValueString())
	createPolicySetRequestBody.SetPolicySetType(plan.Type.ValueString())

	createPolicySetRequestBody.SetScopes(util.ConvertBaseStringArrayToPrimitiveStringArray(plan.Scopes))

	createPolicySetRequest := r.client.ApiClient.GpoDAAS.GpoCreateGpoPolicySet(ctx)
	createPolicySetRequest = createPolicySetRequest.PolicySetRequest(*createPolicySetRequestBody)

	// Create new Policy Set
	policySetResponse, httpResp, err := citrixdaasclient.AddRequestData(createPolicySetRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Policy Set",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Create new policies
	batchRequestModel, err := constructCreatePolicyBatchRequestModel(plan.Policies, policySetResponse.GetPolicySetGuid(), policySetResponse.GetName(), r.client, resp.Diagnostics)
	if err != nil {
		return
	}

	successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Policies to Policy Set "+policySetResponse.GetName(),
			"TransactionId: "+txId+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	if successfulJobs < len(plan.Policies) {
		errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(plan.Policies))
		resp.Diagnostics.AddError(
			"Error adding Policies to Policy Set "+policySetResponse.GetName(),
			"TransactionId: "+txId+
				"\n"+errMsg,
		)
	}

	// Try getting the new policy set with policy set GUID
	policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	if len(policySet.Policies) > 0 {
		// Update Policy Priority
		policyPriorityRequest := constructPolicyPriorityRequest(ctx, r.client, policySet, plan.Policies)
		// Update policy priorities in the Policy Set
		policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, r.client).Execute()
		if err != nil || !policyPriorityResponse {
			resp.Diagnostics.AddError(
				"Error Changing Policy Priorities in Policy Set "+policySet.GetPolicySetGuid(),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+util.ReadClientError(err),
			)
		}
	}

	// Try getting the new policy set with policy set GUID
	policySet, err = getPolicySet(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetResponse.GetPolicySetGuid())
	if err != nil {
		return
	}

	util.RefreshList(plan.Scopes, policySet.Scopes)

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(policySet, policies)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read implements resource.Resource.
func (r *policySetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	var state PolicySetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySet, err := readPolicySet(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	policies, err := readPolicies(ctx, r.client, resp, state.Id.ValueString())
	if err != nil {
		return
	}

	util.RefreshList(state.Scopes, policySet.Scopes)

	state = state.RefreshPropertyValues(policySet, policies)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update implements resource.Resource.
func (r *policySetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from plan
	var plan PolicySetResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed policy set properties from Orchestration
	policySetId := plan.Id.ValueString()
	policySetName := plan.Name.ValueString()

	_, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	policySets, err := getPolicySets(ctx, r.client, &resp.Diagnostics)
	if err != nil {
		return
	}

	for _, policySet := range policySets {
		if strings.EqualFold(policySet.GetName(), policySetName) && !strings.EqualFold(policySet.GetPolicySetGuid(), policySetId) {
			resp.Diagnostics.AddError(
				"Error Updating Policy Set "+policySetId,
				"Policy Set with name "+policySetName+" already exists",
			)
			return
		}
	}

	stateAndPlanDiff, _ := req.State.Raw.Diff(req.Plan.Raw)
	var policiesModified bool
	for _, diff := range stateAndPlanDiff {
		if diff.Path.Steps()[0].Equal(tftypes.AttributeName("policies")) {
			policiesModified = true
			break
		}
	}

	if policiesModified {
		// Get Remote Policies
		policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetId)
		if err != nil {
			return
		}

		// Setup batch requests
		deletePolicyBatchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating policies in policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not update policies within the policy set, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		// Clean up all the policies, settings, and filters in policy set
		for index, policy := range policies.Items {
			relativeUrl := fmt.Sprintf("/gpo/policies/%s", policy.GetPolicyGuid())

			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(fmt.Sprintf("deletePolicy%s", strconv.Itoa(index)))
			batchRequestItem.SetMethod(http.MethodDelete)
			batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetHeaders(batchApiHeaders)
			deletePolicyBatchRequestItems = append(deletePolicyBatchRequestItems, batchRequestItem)
		}

		var deletePolicyBatchRequestModel citrixorchestration.BatchRequestModel
		deletePolicyBatchRequestModel.SetItems(deletePolicyBatchRequestItems)

		successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, deletePolicyBatchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error cleanup Policies in Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(deletePolicyBatchRequestItems) {
			errMsg := fmt.Sprintf("An error occurred while deleting policies in the Policy Set. %d of %d policies were deleted from the Policy Set.", successfulJobs, len(deletePolicyBatchRequestItems))
			resp.Diagnostics.AddError(
				"Error deleting Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return
		}

		// Create all the policies, settings, and filters in the plan
		createPoliciesBatchRequestModel, err := constructCreatePolicyBatchRequestModel(plan.Policies, plan.Id.ValueString(), plan.Name.ValueString(), r.client, resp.Diagnostics)
		if err != nil {
			return
		}

		successfulJobs, txId, err = citrixdaasclient.PerformBatchOperation(ctx, r.client, createPoliciesBatchRequestModel)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\nError message: "+util.ReadClientError(err),
			)
			return
		}

		if successfulJobs < len(createPoliciesBatchRequestModel.Items) {
			errMsg := fmt.Sprintf("An error occurred while adding policies to the Policy Set. %d of %d policies were added to the Policy Set.", successfulJobs, len(createPoliciesBatchRequestModel.Items))
			resp.Diagnostics.AddError(
				"Error adding Policies to Policy Set "+policySetName,
				"TransactionId: "+txId+
					"\n"+errMsg,
			)

			return
		}

		// Update policy priority
		policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
		if err != nil {
			return
		}

		if len(policySet.Policies) > 0 {
			policyPriorityRequest := constructPolicyPriorityRequest(ctx, r.client, policySet, plan.Policies)
			// Update policy priorities in the Policy Set
			policyPriorityResponse, httpResp, err := citrixdaasclient.AddRequestData(policyPriorityRequest, r.client).Execute()
			if err != nil || !policyPriorityResponse {
				resp.Diagnostics.AddError(
					"Error updating Policy Priorities in Policy Set "+policySet.GetPolicySetGuid(),
					"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
						"\nError message: "+util.ReadClientError(err),
				)
				return
			}
		}
	}

	// Construct the update model
	var editPolicySetRequestBody = &citrixorchestration.PolicySetRequest{}
	editPolicySetRequestBody.SetName(policySetName)
	editPolicySetRequestBody.SetDescription(plan.Description.ValueString())
	scopeIds, err := fetchScopeIdsByNames(ctx, r.client, resp.Diagnostics, plan.Scopes)
	if err != nil {
		return
	}
	editPolicySetRequestBody.SetScopes(util.ConvertBaseStringArrayToPrimitiveStringArray(scopeIds))

	editPolicySetRequest := r.client.ApiClient.GpoDAAS.GpoUpdateGpoPolicySet(ctx, policySetId)
	editPolicySetRequest = editPolicySetRequest.PolicySetRequest(*editPolicySetRequestBody)

	// Update Policy Set
	httpResp, err := citrixdaasclient.AddRequestData(editPolicySetRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Policy Set",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}

	// Try getting the new policy set with policy set GUID
	policySet, err := getPolicySet(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	policies, err := getPolicies(ctx, r.client, &resp.Diagnostics, policySetId)
	if err != nil {
		return
	}

	util.RefreshList(plan.Scopes, policySet.Scopes)

	// Map response body to schema and populate Computed attribute values
	plan = plan.RefreshPropertyValues(policySet, policies)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete implements resource.Resource.
func (r *policySetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	defer util.PanicHandler(&resp.Diagnostics)

	// Retrieve values from state
	var state PolicySetResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policySetId := state.Id.ValueString()
	policySetName := state.Name.ValueString()
	// Get delivery groups and check if the current policy set is assigned to one of them
	getDeliveryGroupsRequest := r.client.ApiClient.DeliveryGroupsAPIsDAAS.DeliveryGroupsGetDeliveryGroups(ctx)
	deliveryGroups, httpResp, err := citrixdaasclient.AddRequestData(getDeliveryGroupsRequest, r.client).Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nCould not get delivery group associated with the policy set, unexpected error: "+util.ReadClientError(err),
		)
		return
	}
	associatedDeliveryGroupIds := []string{}
	for _, deliveryGroup := range deliveryGroups.Items {
		if deliveryGroup.GetPolicySetGuid() == policySetId {
			associatedDeliveryGroupIds = append(associatedDeliveryGroupIds, deliveryGroup.GetId())
		}
	}

	if len(associatedDeliveryGroupIds) > 0 {
		// Unassign policy set from delivery groups to unblock delete operation
		batchApiHeaders, httpResp, err := generateBatchApiHeaders(r.client)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not remove policy set from delivery groups, unexpected error: "+util.ReadClientError(err),
			)
			return
		}
		batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
		var editDeliveryGroupRequestBody citrixorchestration.EditDeliveryGroupRequestModel
		editDeliveryGroupRequestBody.SetPolicySetGuid(util.DefaultSitePolicySetId)
		editDeliveryGroupStringBody, err := util.ConvertToString(editDeliveryGroupRequestBody)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error policy set "+policySetName+" from delivery groups",
				"An unexpected error occurred: "+err.Error(),
			)
			return
		}

		for index, deliveryGroupId := range associatedDeliveryGroupIds {
			relativeUrl := fmt.Sprintf("/DeliveryGroups/%s", deliveryGroupId)
			var batchRequestItem citrixorchestration.BatchRequestItemModel
			batchRequestItem.SetReference(strconv.Itoa(index))
			batchRequestItem.SetMethod(http.MethodPatch)
			batchRequestItem.SetRelativeUrl(r.client.GetBatchRequestItemRelativeUrl(relativeUrl))
			batchRequestItem.SetBody(editDeliveryGroupStringBody)
			batchRequestItem.SetHeaders(batchApiHeaders)
			batchRequestItems = append(batchRequestItems, batchRequestItem)
		}

		if len(batchRequestItems) > 0 {
			// If there are any machines that need to be put in maintenance mode
			var batchRequestModel citrixorchestration.BatchRequestModel
			batchRequestModel.SetItems(batchRequestItems)
			successfulJobs, txId, err := citrixdaasclient.PerformBatchOperation(ctx, r.client, batchRequestModel)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error unassign policy set "+policySetName+" from delivery groups "+policySetName,
					"TransactionId: "+txId+
						"\nError Message: "+util.ReadClientError(err),
				)
				return
			}

			if successfulJobs < len(batchRequestItems) {
				errMsg := fmt.Sprintf("An error occurred removing policy set %s from delivery groups. Unassigned from %d of %d delivery groups.", policySetName, successfulJobs, len(batchRequestItems))
				resp.Diagnostics.AddError(
					"Error deleting Policy Set "+policySetName,
					"TransactionId: "+txId+
						"\n"+errMsg,
				)

				return
			}
		}
	}

	// Delete existing Policy Set
	deletePolicySetRequest := r.client.ApiClient.GpoDAAS.GpoDeleteGpoPolicySet(ctx, policySetId)
	httpResp, err = citrixdaasclient.AddRequestData(deletePolicySetRequest, r.client).Execute()
	if err != nil && httpResp.StatusCode != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Error Deleting Policy Set "+policySetName,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *policySetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Gets the policy set and logs any errors
func getPolicySets(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics) ([]citrixorchestration.PolicySetResponse, error) {
	getPolicySetsRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySets(ctx)
	policySets, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfPolicySetResponse](getPolicySetsRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Sets",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return policySets.Items, err
}

func getPolicySet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.PolicySetResponse, error) {
	getPolicySetRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySet(ctx, policySetId)
	getPolicySetRequest = getPolicySetRequest.WithPolicies(true)
	policySet, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.PolicySetResponse](getPolicySetRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policy Set "+policySetId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return policySet, err
}

func readPolicySet(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, policySetId string) (*citrixorchestration.PolicySetResponse, error) {
	getPolicySetRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicySet(ctx, policySetId)
	getPolicySetRequest = getPolicySetRequest.WithPolicies(true)
	policySet, _, err := util.ReadResource[*citrixorchestration.PolicySetResponse](getPolicySetRequest, ctx, client, resp, "PolicySet", policySetId)
	return policySet, err
}

// Gets the policy set and logs any errors
func getPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetId string) (*citrixorchestration.CollectionEnvelopeOfPolicyResponse, error) {
	getPoliciesRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicies(ctx)
	getPoliciesRequest = getPoliciesRequest.PolicySetGuid(policySetId)
	getPoliciesRequest = getPoliciesRequest.WithFilters(true)
	getPoliciesRequest = getPoliciesRequest.WithSettings(true)
	policies, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.CollectionEnvelopeOfPolicyResponse](getPoliciesRequest, client)
	if err != nil {
		diagnostics.AddError(
			"Error Reading Policies in Policy Set "+policySetId,
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
	}

	return policies, err
}

func readPolicies(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, policySetId string) (*citrixorchestration.CollectionEnvelopeOfPolicyResponse, error) {
	getPoliciesRequest := client.ApiClient.GpoDAAS.GpoReadGpoPolicies(ctx)
	getPoliciesRequest = getPoliciesRequest.PolicySetGuid(policySetId)
	getPoliciesRequest = getPoliciesRequest.WithFilters(true)
	getPoliciesRequest = getPoliciesRequest.WithSettings(true)
	policies, _, err := util.ReadResource[*citrixorchestration.CollectionEnvelopeOfPolicyResponse](getPoliciesRequest, ctx, client, resp, "Policies", policySetId)
	return policies, err
}

func generateBatchApiHeaders(client *citrixdaasclient.CitrixDaasClient) ([]citrixorchestration.NameValueStringPairModel, *http.Response, error) {
	headers := []citrixorchestration.NameValueStringPairModel{}

	cwsAuthToken, httpResp, err := client.SignIn()
	var token string
	if err != nil {
		return headers, httpResp, err
	}

	if cwsAuthToken != "" {
		token = strings.Split(cwsAuthToken, "=")[1]
		var header citrixorchestration.NameValueStringPairModel
		header.SetName("Authorization")
		header.SetValue("Bearer " + token)
		headers = append(headers, header)
	}

	return headers, httpResp, err
}

func constructCreatePolicyBatchRequestModel(policiesToCreate []PolicyModel, policySetGuid string, policySetName string, client *citrixdaasclient.CitrixDaasClient, diagnostic diag.Diagnostics) (citrixorchestration.BatchRequestModel, error) {
	batchRequestItems := []citrixorchestration.BatchRequestItemModel{}
	var batchRequestModel citrixorchestration.BatchRequestModel

	for policyIndex, policyToCreate := range policiesToCreate {
		var createPolicyRequest = citrixorchestration.PolicyRequest{}
		createPolicyRequest.SetName(policyToCreate.Name.ValueString())
		createPolicyRequest.SetDescription(policyToCreate.Description.ValueString())
		createPolicyRequest.SetIsEnabled(policyToCreate.IsEnabled.ValueBool())
		// Add Policy Settings
		policySettings := []citrixorchestration.SettingRequest{}
		for _, policySetting := range policyToCreate.PolicySettings {
			settingRequest := citrixorchestration.SettingRequest{}
			settingRequest.SetSettingName(policySetting.Name.ValueString())
			settingRequest.SetUseDefault(policySetting.UseDefault.ValueBool())
			settingRequest.SetSettingValue(policySetting.Value.ValueString())
			policySettings = append(policySettings, settingRequest)
		}
		createPolicyRequest.SetSettings(policySettings)

		// Add Policy Filters
		policyFilters := []citrixorchestration.FilterRequest{}
		for _, policyFilter := range policyToCreate.PolicyFilters {
			filterRequest := citrixorchestration.FilterRequest{}
			filterRequest.SetFilterType(policyFilter.Type.ValueString())
			filterRequest.SetFilterData(policyFilter.Data.ValueString())
			filterRequest.SetIsAllowed(policyFilter.IsAllowed.ValueBool())
			filterRequest.SetIsEnabled(policyFilter.IsEnabled.ValueBool())
			policyFilters = append(policyFilters, filterRequest)
		}
		createPolicyRequest.SetFilters(policyFilters)

		createPolicyRequestBodyString, err := util.ConvertToString(createPolicyRequest)
		if err != nil {
			diagnostic.AddError(
				"Error adding Policy "+policyToCreate.Name.ValueString()+" to Policy Set "+policySetName,
				"An unexpected error occurred: "+err.Error(),
			)
			return batchRequestModel, err
		}

		batchApiHeaders, httpResp, err := generateBatchApiHeaders(client)
		if err != nil {
			diagnostic.AddError(
				"Error deleting policy from policy set "+policySetName,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nCould not delete policies within the policy set to be updated, unexpected error: "+util.ReadClientError(err),
			)
			return batchRequestModel, err
		}

		relativeUrl := fmt.Sprintf("/gpo/policies?policySetGuid=%s", policySetGuid)

		var batchRequestItem citrixorchestration.BatchRequestItemModel
		batchRequestItem.SetReference(fmt.Sprintf("createPolicy%d", policyIndex))
		batchRequestItem.SetMethod(http.MethodPost)
		batchRequestItem.SetRelativeUrl(client.GetBatchRequestItemRelativeUrl(relativeUrl))
		batchRequestItem.SetHeaders(batchApiHeaders)
		batchRequestItem.SetBody(createPolicyRequestBodyString)
		batchRequestItems = append(batchRequestItems, batchRequestItem)
	}

	batchRequestModel.SetItems(batchRequestItems)
	return batchRequestModel, nil
}

func constructPolicyPriorityRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, policySet *citrixorchestration.PolicySetResponse, planedPolicies []PolicyModel) citrixorchestration.ApiGpoRankGpoPoliciesRequest {
	// 1. Construct map of policy name: policy id
	// 2. Construct array of policy id based on the policy name order
	// 3. post policy priority
	policyNameIdMap := map[types.String]types.String{}
	if policySet.GetPolicies() != nil {
		for _, policy := range policySet.GetPolicies() {
			policyNameIdMap[types.StringValue(policy.GetPolicyName())] = types.StringValue(policy.GetPolicyGuid())
		}
	}
	policyPriority := []types.String{}
	for _, policyToCreate := range planedPolicies {
		policyPriority = append(policyPriority, policyNameIdMap[policyToCreate.Name])
	}

	policySetId := policySet.GetPolicySetGuid()
	createPolicyPriorityRequest := client.ApiClient.GpoDAAS.GpoRankGpoPolicies(ctx)
	createPolicyPriorityRequest = createPolicyPriorityRequest.PolicySetGuid(policySetId)
	createPolicyPriorityRequest = createPolicyPriorityRequest.RequestBody(util.ConvertBaseStringArrayToPrimitiveStringArray(policyPriority))
	return createPolicyPriorityRequest
}

func fetchScopeIdsByNames(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics diag.Diagnostics, scopeNames []types.String) ([]types.String, error) {
	getAdminScopesRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScopes(ctx)
	// Create new Policy Set
	getScopesResponse, httpResp, err := citrixdaasclient.AddRequestData(getAdminScopesRequest, client).Execute()
	if err != nil || getScopesResponse == nil {
		diagnostics.AddError(
			"Error fetch scope ids from names",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+util.ReadClientError(err),
		)
		return nil, err
	}

	scopeNameIdMap := map[string]types.String{}
	for _, scope := range getScopesResponse.Items {
		scopeNameIdMap[scope.GetName()] = types.StringValue(scope.GetId())
	}

	scopeIds := []types.String{}
	for _, scopeName := range scopeNames {
		scopeIds = append(scopeIds, scopeNameIdMap[scopeName.ValueString()])
	}

	return scopeIds, nil
}
