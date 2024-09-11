// Copyright © 2024. Citrix Systems, Inc.

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"time"

	ccadmins "github.com/citrix/citrix-daas-rest-go/ccadmins"
	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	citrixstorefrontclient "github.com/citrix/citrix-daas-rest-go/citrixstorefront/apis"
	citrixstorefront "github.com/citrix/citrix-daas-rest-go/citrixstorefront/models"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// AWS Role ARN Regex
const AwsRoleArnRegex string = `^arn:aws(-us-gov)?:iam::[0-9]{12}:role\/[a-zA-Z0-9+=,.@\-_]{1,64}$`

// Aws Access Key Id Regex
const AwsAccessKeyIdRegex string = `^[\w]+$`

// Aws Region Regex
const AwsRegionRegex string = `^[a-zA-Z0-9\-]+$`

// Aws Security Group ID Regex
const AwsSecurityGroupId = `^sg-[a-zA-Z0-9]+$`

// Aws Directory ID Regex
const AwsDirectoryId = `^d-[a-zA-Z0-9]+$`

// Aws Directory ID Regex
const AwsSubnetIdFormat = `^subnet-[a-zA-Z0-9]+$`

// Domain FQDN
const DomainFqdnRegex string = `^(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`

// SAM
const SamRegex string = `^[a-zA-Z][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]\\\w[\w\.\- ]+$`

// UPN
const UpnRegex string = `^[^@]+@\b(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`

const SamAndUpnRegex string = `^[a-zA-Z][a-zA-Z0-9\-_]{0,61}[a-zA-Z0-9]\\\w[\w\.\- ]+$|^[^@]+@\b(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`

// GUID
const GuidRegex string = `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$`

// GUID
const StoreFrontServerIdRegex string = `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[0-9]+[}]?$`

// IPv4
const IPv4Regex string = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`

// IPv4 with https
const IPv4RegexWithProtocol string = `^(http|https)://((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`

// AWS Network Name
const AwsNetworkNameRegex string = `^(\d{1,3}\.){3}\d{1,3}` + "`" + `/\d{1,3}\s\(vpc-.+\)\.network$`

// Date YYYY-MM-DD
const DateRegex string = `^\d{4}-\d{2}-\d{2}$`

// Time HH:MM
const TimeRegex string = `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`

// TimeSpan dd.HH:MM:SS
const TimeSpanRegex string = `^(\d+)\.((\d)|(1\d)|(2[0-3])):((\d)|[1-5][0-9]):((\d)|[1-5][0-9])$`

// ID of the Default Site Policy Set
const DefaultSitePolicySetId string = "00000000-0000-0000-0000-000000000000"

// SSL Thumbprint
const SslThumbprintRegex string = `^([0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`

// AWS EC2 Instance Type
const AwsEc2InstanceTypeRegex string = `^[a-z0-9]{1,15}\.[a-z0-9]{1,15}$`

// Active Directory Sid
const ActiveDirectorySidRegex string = `^S-1-[0-59]-\d{2}-\d{8,10}-\d{8,10}-\d{8,10}-[1-9]\d{3}$`

// AWS Machine Image ID REGEX
const AwsAmiRegex string = `^ami-[0-9a-f]{8,17}$`

// AWS Workspace Image ID REGEX
const AwsWsiRegex string = `^wsi-[0-9a-z]{9,63}$`

const AwsAmiAndWsiRegex string = `^ami-[0-9a-f]{8,17}$|^wsi-[0-9a-z]{9,63}$`

// OU Path
const OuPathFormat string = `^OU=.+,DC=.+$`

// Email REGEX
const EmailRegex string = `^[\w-\.]+@([\w-]+\.)+[\w-]+$`

// Okta Domain REGEX
const OktaDomainRegex string = `\.okta\.com$|\.okta-eu\.com$|\.oktapreview\.com$`

// Application Category Path
const AppCategoryPathRegex string = `^([a-zA-Z0-9 ]+\\)*[a-zA-Z0-9 ]+\\?$`

// SAML 2.0 Identity Provider Certificate REGEX
const SamlIdpCertRegex string = `\.[Pp][Ee][Mm]$|\.[Cc][Rr][Tt]$|\.[Cc][Ee][Rr]$`

// NOT_EXIST error code
const NOT_EXIST string = "NOT_EXIST"

// ID of the All Scope
const AllScopeId string = "00000000-0000-0000-0000-000000000000"

// ID of the Citrix Managed Users Scope
const CtxManagedScopeId string = "f71a1148-7030-467a-a6d3-4a6bcf6a6532"

// ID of the Citrix Managed Users Scope
const UsernameForDecoupledWorkspaces string = "[UNDEFINED]"

// Default QuickCreateService AWS Workspaces Scale Settings
const DefaultQcsAwsWorkspacesSessionIdleTimeoutMinutes int64 = 15
const DefaultQcsAwsWorkspacesOffPeakDisconnectTimeoutMinutes int64 = 15
const DefaultQcsAwsWorkspacesOffPeakLogOffTimeoutMinutes int64 = 5
const DefaultQcsAwsWorkspacesOffPeakBufferSizePercent int64 = 0

// Resource Types
const ImageVersionResourceType string = "ImageVersion"
const RegionResourceType string = "Region"
const ServiceOfferingResourceType string = "ServiceOffering"
const SnapshotResourceType string = "Snapshot"
const VhdResourceType string = "Vhd"
const VirtualPrivateCloudResourceType string = "VirtualPrivateCloud"
const VirtualMachineResourceType string = "Vm"
const TemplateResourceType string = "Template"
const StorageResourceType string = "Storage"
const NetworkResourceType string = "Network"
const SecurityGroupResourceType = "SecurityGroup"
const HostResourceType = "Host"

// Azure Storage Types
const StandardLRS = "Standard_LRS"
const StandardSSDLRS = "StandardSSD_LRS"
const Premium_LRS = "Premium_LRS"
const AzureEphemeralOSDisk = "Azure_Ephemeral_OS_Disk"

// Azure License Types
const WindowsClientLicenseType string = "Windows_Client"
const WindowsServerLicenseType string = "Windows_Server"

// GAC
const AssignmentPriority = 0
const GacAppName = "Workspace"

const SensitiveFieldMaskedValue = "*****"

const ProviderInitializationErrorMsg = "Provider initialization error"
const MissingProviderClientIdAndSecretErrorMsg = "client_id and client_secret fields must be set in the provider configuration to manage this resource via terraform."

const CitrixGatewayConnections = "Citrix Gateway connections"
const NonCitrixGatewayConnections = "Non-Citrix Gateway Connections"

const MetadataTerraformName = "ManagedBy"
const MetadataTerrafomValue = "Terraform"
const MetadataHypervisorSecretExpirationDateName = "Citrix_Orchestration_Hypervisor_Secret_Expiration_Date"
const MetadataCitrixPrefix = "citrix_"
const MetadataImageManagementPrepPrefix = "imagemanagementprep_"
const MetadataTaskDataPrefix = "taskdata_"
const MetadataTaskStatePrefix = "taskstate_"

var PlatformSettingsAssignedTo = []string{"AllUsersNoAuthentication"}

// <summary>
// Helper function to validate if a string is a valid UUID or null
// </summary>
// <param name="u">String to validate</param>
// <returns>True if string is a valid UUID, or is null. False if otherwise.</returns>
func IsValidUUIDorNull(u basetypes.StringValue) bool {
	if u.IsNull() {
		return true
	}
	return IsValidUUID(u.ValueString())
}

// <summary>
// Helper function to validate if a string is a valid UUID
// </summary>
// <param name="u">String to validate</param>
// <returns>True if string is a valid UUID. False if otherwise.</returns>
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// <summary>
// Helper function to read inner error message from a generic error returned from citrix-daas-rest-go
// </summary>
// <param name="err">Generic error returned from citrix-daas-rest-go</param>
// <returns>Inner error message</returns>
func ReadClientError(err error) string {
	genericOpenApiError, ok := err.(*citrixorchestration.GenericOpenAPIError)
	if !ok {
		return err.Error()
	}
	msg := genericOpenApiError.Body()
	if msg != nil {
		var msgObj citrixorchestration.ErrorData
		unmarshalError := json.Unmarshal(msg, &msgObj)
		if unmarshalError != nil {
			return err.Error()
		}
		return msgObj.GetErrorMessage()
	}

	return err.Error()
}

func ReadQcsClientError(err error) string {
	genericOpenApiError, ok := err.(*citrixquickcreate.GenericOpenAPIError)
	if !ok {
		return err.Error()
	}
	msg := genericOpenApiError.Body()
	if msg != nil {
		var msgObj citrixquickcreate.ErrorResponse
		unmarshalError := json.Unmarshal(msg, &msgObj)
		if unmarshalError != nil {
			return err.Error()
		}
		if msgObj.Detail.IsSet() {
			return msgObj.GetDetail()
		}
	}

	return err.Error()
}

// <summary>
// Helper function to serialize any struct value into a string
// </summary>
// <param name="model">Input struct value</param>
// <returns>Serialized string value of the struct</returns>
func ConvertToString(model any) (string, error) {
	body, err := json.Marshal(model)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// <summary>
// Extract Ids from a list of objects
// </summary>
// <param name="slice">Input list of objects</param>
// <returns>List of Ids extracted from input list</returns>
func GetIdsForOrchestrationObjects[objType any](slice []objType) []string {
	val := reflect.ValueOf(slice)
	ids := []string{}

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		idField := elem.FieldByName("Id")
		if !idField.IsValid() || idField.Kind() != reflect.String {
			continue
		}
		ids = append(ids, idField.String())
	}

	return ids
}

// <summary>
// Filter and Extract Ids from a list of scope responses
// </summary>
// <param name="scopeIdsInState">List of scope Ids in state or config</param>
// <param name="scopeResponses">List of scope objects from remote</param>
// <returns>List of Ids extracted from response</returns>
func GetIdsForFilteredScopeObjects(scopeIdsInState []string, scopeResponses []citrixorchestration.ScopeResponseModel) []string {
	if scopeIdsInState == nil {
		scopeIdsInState = []string{}
	}
	filteredScopes := []citrixorchestration.ScopeResponseModel{}
	for _, scope := range scopeResponses {
		if scope.GetIsTenantScope() && !slices.ContainsFunc(scopeIdsInState, func(scopeId string) bool {
			return strings.EqualFold(scopeId, scope.GetId())
		}) {
			continue
		}
		filteredScopes = append(filteredScopes, scope)
	}
	scopeIds := GetIdsForScopeObjects(filteredScopes)
	return scopeIds
}

// <summary>
// Extract Ids from a list of scope objects
// </summary>
// <param name="slice">Input list of objects</param>
// <returns>List of Ids extracted from input list</returns>
func GetIdsForScopeObjects[objType any](slice []objType) []string {
	ids := GetIdsForOrchestrationObjects(slice)
	filteredIds := []string{}

	for _, id := range ids {
		if id != AllScopeId && id != CtxManagedScopeId {
			filteredIds = append(filteredIds, id)
		}
	}
	return filteredIds
}

// <summary>
// Helper function for generating string validator for an enum value in Terraform schema.
// Only works when all eligible values for the enum type are supported by provider.
// When the eligible values are only partially supported, use custom string validator in schema.
// </summary>
// <param name="enum">Enum from citrix-daas-rest-go package</param>
// <returns>String validator for terraform schema</returns>
func GetValidatorFromEnum[V ~string, T []V](enum T) validator.String {
	var values []string
	for _, i := range enum {
		values = append(values, string(i))
	}
	return stringvalidator.OneOfCaseInsensitive(
		values...,
	)
}

type HttpErrorBody struct {
	ErrorMessage string `json:"errorMessage"`
	Detail       string `json:"detail"`
}

// <summary>
// Wrapper function for reading specific resource from remote with retries
// </summary>
// <param name="request">Request object for the GET call</param>
// <param name="ctx">Context from caller</param>
// <param name="client">Citrix DaaS client from provider context</param>
// <param name="resp">Response from the GET call</param>
// <param name="resourceType">Resource type that would be shown in error message if failed to read resource</param>
// <param name="resourceIdOrName">Resource ID or name that would be shown in error message if failed to read resource</param>
// <returns>Response of the Get call. Raw http response. Error if failed to read the resource.</returns>
func ReadResource[ResponseType any](request any, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, resourceType, resourceIdOrName string) (ResponseType, *http.Response, error) {
	response, httpResp, err := citrixdaasclient.ExecuteWithRetry[ResponseType](request, client)

	// Resource Location does not return an error if not found
	if resourceType == "Resource Location" {
		if httpResp.StatusCode == http.StatusNoContent {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("%s not found", resourceType),
				fmt.Sprintf("%s %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", resourceType, resourceIdOrName),
			)

			resp.State.RemoveResource(ctx)
			// Set err so that control does not go to refresh properties in the read method
			err = fmt.Errorf("could not find resource location %s", resourceIdOrName)
			return response, httpResp, err
		}
	}
	if err != nil && resp != nil {
		body, _ := io.ReadAll(httpResp.Body)
		httpErrorBody := HttpErrorBody{}
		json.Unmarshal(body, &httpErrorBody)
		if httpResp.StatusCode == http.StatusNotFound {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("%s not found", resourceType),
				fmt.Sprintf("%s %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", resourceType, resourceIdOrName),
			)

			resp.State.RemoveResource(ctx)
		} else if httpResp.StatusCode == http.StatusInternalServerError && httpErrorBody.Detail == "Object does not exist." {

			resp.Diagnostics.AddWarning(
				fmt.Sprintf("%s not found", resourceType),
				fmt.Sprintf("%s %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", resourceType, resourceIdOrName),
			)

			resp.State.RemoveResource(ctx)
		} else if httpResp.StatusCode == http.StatusBadRequest && httpErrorBody.ErrorMessage == "Cannot find this administrator "+resourceIdOrName {

			resp.Diagnostics.AddWarning(
				fmt.Sprintf("%s not found", resourceType),
				fmt.Sprintf("%s %s was not found and will be removed from the state file. An apply action will result in the creation of a new resource.", resourceType, resourceIdOrName),
			)

			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Error Reading %s %s", resourceType, resourceIdOrName),
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadClientError(err),
			)
		}
	}

	return response, httpResp, err
}

// <summary>
// Helper function to process async job response. Takes async job response and polls for result.
// </summary>
// <param name="ctx">Context from caller</param>
// <param name="client">Citrix DaaS client from provider context</param>
// <param name="jobResp">Job response from async API call</param>
// <param name="errContext">Context of the job to be use as Terraform diagnostic error message title</param>
// <param name="diagnostics">Terraform diagnostics from context</param>
// <param name="maxTimeout">Maximum timeout threashold for job status polling</param>
// <returns>Error if job polling failed or job itself ended in failed state</returns>
func ProcessAsyncJobResponse(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, jobResp *http.Response, errContext string, diagnostics *diag.Diagnostics, maxTimeout int, returnJobError bool) (err error) {
	txId := citrixdaasclient.GetTransactionIdFromHttpResponse(jobResp)

	jobId := citrixdaasclient.GetJobIdFromHttpResponse(*jobResp)
	jobResponseModel, err := client.WaitForJob(ctx, jobId, maxTimeout)

	if err != nil {
		diagnostics.AddError(
			errContext,
			"TransactionId: "+txId+
				"\nJobId: "+jobResponseModel.GetId()+
				"\nError message: "+jobResponseModel.GetErrorString(),
		)
		return err
	}

	if jobResponseModel.GetStatus() != citrixorchestration.JOBSTATUS_COMPLETE {
		errorMessage := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			detailedErrorFound := false
			for _, kvp := range jobResponseModel.GetErrorParameters() {
				if kvp.GetName() == "ErrorDetails" {
					errorDetails := kvp.GetValue()
					tflog.Error(ctx, errContext+"\n"+errorMessage+"\nError details: "+errorDetails)

					// Add additional specific error handling for Orchestration job failures here
					if strings.Contains(errorDetails, "No Citrix Workspace Cloud Connector was found") ||
						strings.Contains(errorDetails, "Hcl request is not allowed when connector is in outage mode") {
						detailedErrorFound = true
						errorMessage += "\nError Message: Ensure the Citrix Cloud Connectors in the zone are available and try again."
					}
					if strings.Contains(errorDetails, "Machine profile is not provided and master image has security type as trusted launch") {
						detailedErrorFound = true
						errorMessage += "\nError Message: Master image has security type as trusted launch, this requires machine_profile to be provided."
					}
					if strings.Contains(errorDetails, "The master image associated with this catalog is associated with a VM Generation that is not supported by the configured Service Offering") {
						detailedErrorFound = true
						errorMessage += "\nError Message: service_offering does not support the VM Generation of the master image associated with this catalog."
					}
					break
				}
			}

			if !detailedErrorFound {
				errorMessage += "\nError message: " + jobResponseModel.GetErrorString()
			}
		}

		diagnostics.AddError(
			errContext,
			errorMessage,
		)

		if returnJobError {
			return fmt.Errorf(errorMessage)
		}
	}

	return nil
}

func WaitForQcsDeploymentTaskWithDiags(ctx context.Context, diagnostics *diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, maxWaitTimeInSeconds int, taskId string, taskName, deploymentName string, errorContext string) error {
	task, httpResp, err := PollQcsTask(ctx, client, diagnostics, taskId, 10, maxWaitTimeInSeconds)
	if err != nil {
		diagnostics.AddError(
			fmt.Sprintf("Error %s AWS WorkSpaces Deployment: %s", errorContext, deploymentName),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
		return err
	}
	if task.DeploymentTask.GetTaskState() != citrixquickcreate.TASKSTATE_COMPLETED {
		diagnostics.AddError(
			fmt.Sprintf("Error %s AWS WorkSpaces Deployment: %s", errorContext, deploymentName),
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				fmt.Sprintf("\nError message: %s was not completed. It has state: %s", taskName, TaskStateEnumToString(task.DeploymentTask.GetTaskState())),
		)
		return err
	}
	return nil
}

// Represents a list item which supports being refreshed from a client model
type RefreshableListItemWithAttributes[clientType any] interface {
	// Gets the key to compare the item with the client model
	GetKey() string

	// Refreshes the item with the client model and returns the updated item
	RefreshListItem(context.Context, *diag.Diagnostics, clientType) ModelWithAttributes

	// Has to implement the ModelWithAttributes interface for conversion back to a Terraform model
	ModelWithAttributes
}

// These functions are used by RefreshListProperties
func GetOrchestrationRebootScheduleKey(r citrixorchestration.RebootScheduleResponseModel) string {
	return r.GetName()
}

func GetOrchestrationDesktopKey(r citrixorchestration.DesktopResponseModel) string {
	return r.GetPublishedName()
}

func GetOrchestrationHypervisorStorageKey(remote citrixorchestration.HypervisorStorageResourceResponseModel) string {
	return remote.GetName()
}

func GetOrchestrationNetworkMappingKey(remote citrixorchestration.NetworkMapResponseModel) string {
	return remote.GetDeviceId()
}

func GetOrchestrationRemotePcOuKey(remote citrixorchestration.RemotePCEnrollmentScopeResponseModel) string {
	return remote.GetOU()
}

func GetOrchestrationSmartAccessTagKey(remote citrixorchestration.SmartAccessTagResponseModel) string {
	return remote.GetFarm() + remote.GetFilter()
}

func GetOrchestrationAccessPolicyKey(remote citrixorchestration.AdvancedAccessPolicyResponseModel) string {
	if remote.GetIsBuiltIn() {
		if remote.GetAllowedConnection() == citrixorchestration.ALLOWEDCONNECTION_VIA_AG {
			return CitrixGatewayConnections
		}

		if remote.GetAllowedConnection() == citrixorchestration.ALLOWEDCONNECTION_NOT_VIA_AG {
			return NonCitrixGatewayConnections
		}
	}

	return remote.GetName()
}

func GetOrchestrationNameValueStringPairKey(remote citrixorchestration.NameValueStringPairModel) string {
	return remote.GetName()
}

func GetSTFGroupMemberKey(remote citrixstorefront.STFGroupMemberResponseModel) string {
	return *remote.GroupName.Get()
}

func GetSTFFarmSetKey(remote citrixstorefront.STFFarmSetResponseModel) string {
	return *remote.Name.Get()
}

func GetSTFRoamingGatewayKey(remote citrixstorefront.STFRoamingGatewayResponseModel) string {
	return *remote.Name.Get()
}

func GetSTFSTAUrlKey(remote citrixstorefront.STFSTAUrlModel) string {
	return *remote.StaUrl.Get()
}

func GetQcsAwsWorkspacesWithUsernameKey(remote citrixquickcreate.AwsEdcDeploymentMachine) string {
	return remote.GetUsername()
}

// <summary>
// Helper function for calculating the new state of a list of nested attribute, while
// keeping the order of the elements in the array intact, and adds missing elements
// from remote to state.
// Can be used for refreshing all list nested attributes.
// </summary>
// <param name="state">State values in Terraform model</param>
// <param name="remote">Remote values in client model</param>
// <param name="getClientKey">Function to get the Id from the client model</param>
// <returns>Terraform list for new state</returns>
func RefreshListValueProperties[tfType RefreshableListItemWithAttributes[clientType], clientType any](ctx context.Context, diagnostics *diag.Diagnostics, state types.List, remote []clientType, getClientKey func(clientType) string) types.List {
	unwrappedList := ObjectListToTypedArray[tfType](ctx, diagnostics, state)
	refreshedList := refreshListProperties[tfType, clientType](ctx, diagnostics, unwrappedList, remote, getClientKey)
	return TypedArrayToObjectList[tfType](ctx, diagnostics, refreshedList)
}

func refreshListProperties[tfType RefreshableListItemWithAttributes[clientType], clientType any](ctx context.Context, diagnostics *diag.Diagnostics, state []tfType, remote []clientType, getClientKey func(clientType) string) []tfType {
	if len(remote) == 0 {
		return nil
	}

	if state == nil {
		state = []tfType{}
	}

	stateItems := map[string]int{}
	for index, tfItem := range state {
		stateItems[tfItem.GetKey()] = index
	}

	newState := state
	for _, clientItem := range remote {
		clientKey := getClientKey(clientItem)
		index, exists := stateItems[clientKey]
		if exists {
			tfItem := state[index]
			newState[index] = tfItem.RefreshListItem(ctx, diagnostics, clientItem).(tfType)
		} else {
			var tfStructItem tfType
			if attributeMap, err := AttributeMapFromObject(tfStructItem); err == nil {
				// start with the null object to populate all nested lists/objects as null
				tfStructItem = defaultObjectFromObjectValue[tfType](ctx, types.ObjectNull(attributeMap))
				newStateItem := tfStructItem.RefreshListItem(ctx, diagnostics, clientItem).(tfType)
				newState = append(newState, newStateItem)
			} else {
				diagnostics.AddWarning("Error when creating empty "+reflect.TypeOf(tfStructItem).String(), err.Error())
			}
		}

		stateItems[clientKey] = -1 // Mark as visited. The ones not visited should be removed.
	}

	result := []tfType{}
	for _, tfItem := range newState {
		if stateItems[tfItem.GetKey()] == -1 {
			result = append(result, tfItem) // if visited, include. Not visited ones will not be included.
		}
	}

	return result
}

// <summary>
// Helper function for calculating the new state of a list of strings, while
// keeping the order of the elements in the array intact, and adds missing elements
// from remote to state.
// Can be used for refreshing all list of strings.
// </summary>
// <param name="state">State values in Terraform model</param>
// <param name="remote">Remote values in client model</param>
// <returns>Array in Terraform model for new state</returns>
func RefreshListValues(ctx context.Context, diagnostics *diag.Diagnostics, state types.List, remote []string) types.List {
	unwrappedList := StringListToStringArray(ctx, diagnostics, state)
	refreshedList := RefreshList(unwrappedList, remote)
	return StringArrayToStringList(ctx, diagnostics, refreshedList)
}

// <summary>
// Helper function for calculating the new state of a list of strings, while
// keeping the order of the elements in the array intact, and adds missing elements
// from remote to state.
// Can be used for refreshing list of strings.
// </summary>
// <param name="state">List of values in state</param>
// <param name="remote">List of values in remote</param>
func RefreshList(state []string, remote []string) []string {
	stateItems := map[string]bool{}
	for _, item := range state {
		stateItems[strings.ToLower(item)] = false // not visited
	}

	for _, item := range remote {
		itemInLowerCase := strings.ToLower(item)
		_, exists := stateItems[itemInLowerCase]
		if !exists {
			state = append(state, item)
		}
		stateItems[itemInLowerCase] = true
	}

	result := []string{}
	for _, item := range state {
		if stateItems[strings.ToLower(item)] {
			result = append(result, item)
		}
	}

	return result
}

// <summary>
// Global panic handler to catch all unexpected errors to prevent provider from crashing.
// Writes crash stack into local txt file for troubleshooting, and displays error message in Terrafor Diagnostics.
// </summary>
// <param name="diagnostics">Terraform Diagnostics from context</param>
func PanicHandler(diagnostics *diag.Diagnostics) {
	if r := recover(); r != nil {
		pc, _, _, ok := runtime.Caller(2) // 1=the panic, 2=who called the panic
		f := runtime.FuncForPC(pc)
		if !ok {
			panic(r)
		}
		msg := fmt.Sprintf("An unexpected error occurred in %s.\n\n%v", f.Name(), r)

		// write stack trace to disk so we don't dump on the console
		fileContents := fmt.Sprintf("%s\n\n%s", f.Name(), debug.Stack())
		file, err := os.CreateTemp("", "citrix_provider_crash_stack.*.txt")
		if err == nil {
			defer file.Close()
			_, err := file.WriteString(fileContents)
			if err == nil {
				msg += "\n\nPlease report this issue to the project maintainers and include this file if present: " + file.Name()
			}
		}

		diagnostics.AddError(
			"Unexpected error in the Citrix provider",
			msg,
		)
	}
}

// <summary>
// Helper function to get the allowed functional level values for setting the minimum functional level for machine catalog and deliver group.
// </summary>
func GetAllowedFunctionalLevelValues() []string {
	res := []string{}
	for _, v := range citrixorchestration.AllowedFunctionalLevelEnumValues {
		if v != citrixorchestration.FUNCTIONALLEVEL_UNKNOWN &&
			v != citrixorchestration.FUNCTIONALLEVEL_LMIN &&
			v != citrixorchestration.FUNCTIONALLEVEL_LMAX {
			res = append(res, string(v))
		}
	}

	return res
}

// <summary>
// Helper function to check the version requirement for DDC.
// </summary>
func CheckProductVersion(client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, requiredOrchestrationApiVersion int32, requiredProductMajorVersion int, requiredProductMinorVersion int, errorSummary, feature string) bool {
	// Validate DDC version
	if client.AuthConfig.OnPremises {
		productVersionSplit := strings.Split(client.ClientConfig.ProductVersion, ".")
		productMajorVersion, err := strconv.Atoi(productVersionSplit[0])
		if err != nil {
			diagnostics.AddError(
				errorSummary,
				"Error parsing product major version. Error: "+err.Error(),
			)
			return false
		}

		productMinorVersion, err := strconv.Atoi(productVersionSplit[1])
		if err != nil {
			diagnostics.AddError(
				errorSummary,
				"Error parsing product minor version. Error: "+err.Error(),
			)
			return false
		}

		if productMajorVersion < requiredProductMajorVersion ||
			(productMajorVersion == requiredProductMajorVersion && productMinorVersion < requiredProductMinorVersion) {
			diagnostics.AddError(
				errorSummary,
				fmt.Sprintf("%s is not supported for current DDC version %d.%d. Please upgrade your DDC product version to %d.%d or above.", feature, productMajorVersion, productMinorVersion, requiredProductMajorVersion, requiredProductMinorVersion),
			)
			return false
		}
	}

	// Validate Orchestration version
	if client.ClientConfig.OrchestrationApiVersion < requiredOrchestrationApiVersion {
		diagnostics.AddError(
			errorSummary,
			fmt.Sprintf("%s is not supported for current DDC version %d. Please upgrade your DDC product version to %d or above.", feature, client.ClientConfig.OrchestrationApiVersion, requiredOrchestrationApiVersion),
		)
		return false
	}

	return true
}

func GetProductMajorAndMinorVersion(client *citrixdaasclient.CitrixDaasClient) (int, int, error) {
	productVersionSplit := strings.Split(client.ClientConfig.ProductVersion, ".")
	productMajorVersion, err := strconv.Atoi(productVersionSplit[0])
	if err != nil {
		return 0, 0, err
	}

	productMinorVersion, err := strconv.Atoi(productVersionSplit[1])
	if err != nil {
		return 0, 0, err
	}

	return productMajorVersion, productMinorVersion, nil
}

// <summary>
// Helper function to check the version requirement for StoreFront.
// </summary>
func CheckStoreFrontVersion(client *citrixstorefrontclient.STFVersion, ctx context.Context, diagnostic *diag.Diagnostics, requiredMajorVersion int, requiredMinorVersion int) bool {
	// Validate StoreFront version
	versionRequest := client.STFVersionGetVersion(ctx)
	versionResponse, err := versionRequest.Execute()
	if err != nil {
		diagnostic.AddError(
			"Error fetching StoreFront version",
			"Error message: "+err.Error(),
		)
		return false
	}

	if versionResponse.Major.IsSet() && versionResponse.Minor.IsSet() {
		majorVersion := *versionResponse.Major.Get()
		minorVersion := *versionResponse.Minor.Get()

		if majorVersion < requiredMajorVersion ||
			(majorVersion == requiredMajorVersion && minorVersion < requiredMinorVersion) {
			return false
		}
	} else {
		diagnostic.AddError(
			"Error fetching StoreFront version",
			"Error message: StoreFront Major and Minor version not set",
		)
		return false
	}

	return true
}

// </summary>
// Helper function to refresh user list.
// </summary>
func RefreshUsersList(ctx context.Context, diags *diag.Diagnostics, usersSet types.Set, usersInRemote []citrixorchestration.IdentityUserResponseModel) types.Set {
	samNamesMap := map[string]int{}
	upnMap := map[string]int{}

	for index, userInRemote := range usersInRemote {
		userSamName := userInRemote.GetSamName()
		userPrincipalName := userInRemote.GetPrincipalName()
		if userSamName != "" {
			samNamesMap[strings.ToLower(userSamName)] = index
		}
		if userPrincipalName != "" {
			upnMap[strings.ToLower(userPrincipalName)] = index
		}
	}

	res := []string{}
	users := StringSetToStringArray(ctx, diags, usersSet)
	for _, user := range users {
		samRegex, _ := regexp.Compile(SamRegex)
		if samRegex.MatchString(user) {
			index, exists := samNamesMap[strings.ToLower(user)]
			if !exists {
				continue
			}
			res = append(res, user)
			samNamesMap[strings.ToLower(user)] = -1
			if index != -1 {
				userPrincipalName := usersInRemote[index].GetPrincipalName()
				_, exists = upnMap[strings.ToLower(userPrincipalName)]
				if exists {
					upnMap[strings.ToLower(userPrincipalName)] = -1
				}
			}

			continue
		}

		upnRegex, _ := regexp.Compile(UpnRegex)
		if upnRegex.MatchString(user) {
			index, exists := upnMap[strings.ToLower(user)]
			if !exists {
				continue
			}
			res = append(res, user)
			upnMap[strings.ToLower(user)] = -1
			if index != -1 {
				samName := usersInRemote[index].GetSamName()
				_, exists = samNamesMap[strings.ToLower(samName)]
				if exists {
					samNamesMap[strings.ToLower(samName)] = -1
				}
			}
		}
	}

	for samName, index := range samNamesMap {
		if index != -1 { // Users that are only in remote
			res = append(res, samName)
		}
	}

	return StringArrayToStringSet(ctx, diags, res)
}

// <summary>
// Helper function to fetch scope ids from scope names
// </summary>
func FetchScopeIdsByNames(ctx context.Context, diagnostics diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, scopeNames []string) ([]string, error) {
	getAdminScopesRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScopes(ctx)
	// get all the scopes
	getScopesResponse, httpResp, err := citrixdaasclient.AddRequestData(getAdminScopesRequest, client).Execute()
	if err != nil || getScopesResponse == nil {
		diagnostics.AddError(
			"Error fetch scope ids from names",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
		return nil, err
	}

	scopeNameIdMap := map[string]string{}
	for _, scope := range getScopesResponse.Items {
		scopeNameIdMap[scope.GetName()] = scope.GetId()
	}

	scopeIds := []string{}
	for _, scopeName := range scopeNames {
		scopeIds = append(scopeIds, scopeNameIdMap[scopeName])
	}

	return scopeIds, nil
}

// <summary>
// Helper function to fetch scope names from scope ids
// </summary>
func FetchScopeNamesByIds(ctx context.Context, diagnostics diag.Diagnostics, client *citrixdaasclient.CitrixDaasClient, scopeIds []string) ([]string, error) {
	getAdminScopesRequest := client.ApiClient.AdminAPIsDAAS.AdminGetAdminScopes(ctx)
	// Create new Policy Set
	getScopesResponse, httpResp, err := citrixdaasclient.AddRequestData(getAdminScopesRequest, client).Execute()
	if err != nil || getScopesResponse == nil {
		diagnostics.AddError(
			"Error fetch scope names from ids",
			"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
				"\nError message: "+ReadClientError(err),
		)
		return nil, err
	}

	scopeIdNameMap := map[string]types.String{}
	for _, scope := range getScopesResponse.Items {
		scopeIdNameMap[scope.GetId()] = types.StringValue(scope.GetName())
	}

	scopeNames := []string{}
	for _, scopeId := range scopeIds {
		scopeNames = append(scopeNames, scopeIdNameMap[scopeId].ValueString())
	}

	return scopeNames, nil
}

func GetUsersUsingIdentity(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, users []string) ([]citrixorchestration.IdentityUserResponseModel, *http.Response, error) {
	allUsersFromIdentity := []citrixorchestration.IdentityUserResponseModel{}

	getIncludedUsersRequest := client.ApiClient.IdentityAPIsDAAS.IdentityGetUsers(ctx)
	getIncludedUsersRequest = getIncludedUsersRequest.User(users).UserType(citrixorchestration.IDENTITYUSERTYPE_ALL)
	adUsers, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.IdentityUserResponseModelCollection](getIncludedUsersRequest, client)

	if err != nil {
		return allUsersFromIdentity, httpResp, err
	}

	allUsersFromIdentity = append(allUsersFromIdentity, adUsers.GetItems()...)

	if len(allUsersFromIdentity) < len(users) {
		getIncludedUsersRequest = getIncludedUsersRequest.User(users).UserType(citrixorchestration.IDENTITYUSERTYPE_ALL).Provider(citrixorchestration.IDENTITYPROVIDERTYPE_ALL)
		azureAdUsers, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixorchestration.IdentityUserResponseModelCollection](getIncludedUsersRequest, client)

		if err != nil {
			return allUsersFromIdentity, httpResp, err
		}

		allUsersFromIdentity = append(allUsersFromIdentity, azureAdUsers.GetItems()...)
	}

	err = VerifyIdentityUserListCompleteness(users, allUsersFromIdentity)

	if err != nil {
		return allUsersFromIdentity, httpResp, err
	}

	return allUsersFromIdentity, httpResp, nil
}

func GetUserIdsUsingIdentity(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, users []string) ([]string, *http.Response, error) {
	userIds := []string{}
	allUsersFromIdentity, httpResp, err := GetUsersUsingIdentity(ctx, client, users)
	if err != nil {
		return userIds, httpResp, err
	}

	for _, user := range allUsersFromIdentity {
		id := user.GetOid() // Azure AD users
		if id == "" {
			id = user.GetSid() // For AD users, OID is empty, use SID
		}
		userIds = append(userIds, id)
	}

	return userIds, httpResp, nil
}

func VerifyIdentityUserListCompleteness(inputUserNames []string, remoteUsers []citrixorchestration.IdentityUserResponseModel) error {
	missingUsers := []string{}
	for _, includedUser := range inputUserNames {
		userIndex := slices.IndexFunc(remoteUsers, func(i citrixorchestration.IdentityUserResponseModel) bool {
			return strings.EqualFold(includedUser, i.GetSamName()) || strings.EqualFold(includedUser, i.GetPrincipalName())
		})
		if userIndex == -1 {
			missingUsers = append(missingUsers, includedUser)
		}
	}

	if len(missingUsers) > 0 {
		return fmt.Errorf("The following users could not be found: " + strings.Join(missingUsers, ", "))
	}

	return nil
}

func GetConfigValuesForSchema(ctx context.Context, diags *diag.Diagnostics, m ModelWithAttributes) (string, map[string]interface{}) {
	sensitiveFields := GetSensitiveFieldsForAttribute(ctx, diags, m.GetAttributes())
	dataObj := TypedObjectToObjectValue(ctx, diags, m)
	return reflect.TypeOf(m).String(), GetConfigValuesForObject(ctx, diags, dataObj, sensitiveFields)
}

func GetConfigValuesForObject(ctx context.Context, diags *diag.Diagnostics, obj types.Object, sensitiveFields map[string]bool) map[string]interface{} {

	// Get the attributes for the object
	attributes := obj.Attributes()
	configValues := make(map[string]interface{})
	for name, attribute := range attributes {
		if _, ok := sensitiveFields[name]; ok {
			configValues[name] = SensitiveFieldMaskedValue
			continue
		}

		configValues[name] = GetAttributeValues(ctx, diags, attribute, sensitiveFields)
	}
	return configValues
}

func GetConfigValuesForMap(ctx context.Context, diags *diag.Diagnostics, configMap types.Map) map[string]interface{} {

	if configMap.IsNull() || configMap.IsUnknown() {
		return nil
	}
	configValues := make(map[string]interface{})
	for name, value := range configMap.Elements() {
		configValues[name] = GetAttributeValues(ctx, diags, value, map[string]bool{})
	}
	return configValues
}

func GetAttributeValues(ctx context.Context, diags *diag.Diagnostics, attribute attr.Value, sensitiveFields map[string]bool) interface{} {
	refVal := reflect.ValueOf(attribute)
	switch attribute.(type) {
	case types.String:
		return refVal.Interface().(types.String).ValueString()
	case types.Bool:
		return refVal.Interface().(types.Bool).ValueBool()
	case types.Int64:
		return refVal.Interface().(types.Int64).ValueInt64()
	case types.Float64:
		return refVal.Interface().(types.Float64).ValueFloat64()
	case types.List:
		reflectedList := refVal.Interface().(types.List)
		reflectedElementType := reflect.TypeOf(reflectedList.ElementType(ctx))
		if reflectedElementType.String() == "basetypes.StringType" {
			return StringListToStringArray(ctx, diags, reflectedList)
		} else if reflectedElementType.String() == "basetypes.ObjectType" {
			objectList := make([]map[string]interface{}, 0)
			for _, item := range reflectedList.Elements() {
				objectList = append(objectList, GetConfigValuesForObject(ctx, diags, item.(types.Object), sensitiveFields))
			}
			return objectList
		} else {
			diags.AddWarning("Invalid Element Type", "Following element type is not supported in lists: "+reflectedElementType.String())
			return nil
		}
	case types.Set:
		reflectedSet := refVal.Interface().(types.Set)
		reflectedElementType := reflect.TypeOf(reflectedSet.ElementType(ctx))
		if reflectedElementType.String() == "basetypes.StringType" {
			return StringSetToStringArray(ctx, diags, reflectedSet)
		} else if reflectedElementType.String() == "basetypes.ObjectType" {
			objectList := make([]map[string]interface{}, 0)
			for _, item := range reflectedSet.Elements() {
				objectList = append(objectList, GetConfigValuesForObject(ctx, diags, item.(types.Object), sensitiveFields))
			}
			return objectList
		} else {
			diags.AddWarning("Invalid Element Type", "Following element type is not supported in sets: "+reflectedElementType.String())
			return nil
		}
	case types.Map: // Revisit this once schema uses Maps
		reflectedMap := refVal.Interface().(types.Map)
		reflectedValueType := reflect.TypeOf(reflectedMap.ElementType(ctx))
		if reflectedValueType.String() == "basetypes.StringType" || reflectedValueType.String() == "basetypes.Int64Type" || reflectedValueType.String() == "basetypes.Float64Type" || reflectedValueType.String() == "basetypes.BoolType" {
			return GetConfigValuesForMap(ctx, diags, reflectedMap)
		} else if reflectedValueType.String() == "basetypes.ObjectType" {
			objectList := make(map[string]interface{})
			for key, item := range reflectedMap.Elements() {
				objectList[key] = GetConfigValuesForObject(ctx, diags, item.(types.Object), sensitiveFields)
			}
			return objectList
		} else {
			diags.AddWarning("Invalid Element Type", "Following element type is not supported in maps: "+reflectedValueType.String())
			return nil
		}

	case types.Object:
		return GetConfigValuesForObject(ctx, diags, refVal.Interface().(types.Object), sensitiveFields)
	default:
		// Unknown type
		diags.AddWarning("Invalid Attribute Type", "Attribute type not supported: "+reflect.TypeOf(attribute).String())
		return nil
	}
}

func GetSensitiveFieldsForAttribute(ctx context.Context, diags *diag.Diagnostics, attributes map[string]schema.Attribute) map[string]bool {
	sensitiveFields := map[string]bool{}
	for name, attribute := range attributes {

		fieldsMap, isSensitive := CheckIfFieldIsSensitive(ctx, diags, attribute)
		if isSensitive {
			if fieldsMap != nil {
				for key, val := range fieldsMap {
					sensitiveFields[key] = val
				}
			} else {
				sensitiveFields[name] = true
			}
		}
	}
	return sensitiveFields
}

func CheckIfFieldIsSensitive(ctx context.Context, diags *diag.Diagnostics, attribute schema.Attribute) (map[string]bool, bool) {

	// If root attribute is sensitive, return true.
	if attribute.IsSensitive() {
		return nil, true
	}

	switch attr := attribute.(type) {
	case schema.StringAttribute, schema.BoolAttribute, schema.Int64Attribute, schema.Float64Attribute, schema.ListAttribute, schema.SetAttribute, schema.MapAttribute:
		return nil, false
	case schema.SingleNestedAttribute:
		sensitiveFields := GetSensitiveFieldsForAttribute(ctx, diags, attr.Attributes)
		return sensitiveFields, len(sensitiveFields) > 0
	case schema.ListNestedAttribute:
		sensitiveFields := GetSensitiveFieldsForAttribute(ctx, diags, attr.NestedObject.Attributes)
		return sensitiveFields, len(sensitiveFields) > 0
	case schema.SetNestedAttribute:
		sensitiveFields := GetSensitiveFieldsForAttribute(ctx, diags, attr.NestedObject.Attributes)
		return sensitiveFields, len(sensitiveFields) > 0
	case schema.MapNestedAttribute: // Revisit this once schema uses MapNestedAttribute
		sensitiveFields := GetSensitiveFieldsForAttribute(ctx, diags, attr.NestedObject.Attributes)
		return sensitiveFields, len(sensitiveFields) > 0
	}
	diags.AddWarning("Invalid Attribute Type", "Attribute type not supported: "+reflect.TypeOf(attribute).String())
	return nil, false
}

// <summary>
// Helper function to poll the task until either the task completed or error out or timed out.
// </summary>
func PollQcsTask(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, taskId string, pollIntervalSeconds int, maxWaitTimeSeconds int) (*citrixquickcreate.GetTaskAsync200Response, *http.Response, error) {
	if pollIntervalSeconds == 0 {
		// Default to 10 seconds
		pollIntervalSeconds = 10
	}
	if maxWaitTimeSeconds == 0 {
		// Default to 5 minutes
		maxWaitTimeSeconds = 300
	}

	startTime := time.Now()
	getTaskRequest := client.QuickCreateClient.TasksQCS.GetTaskAsync(ctx, client.ClientConfig.CustomerId, taskId)

	var taskResponse *citrixquickcreate.GetTaskAsync200Response
	var httpResp *http.Response
	var err error

	for {
		if time.Since(startTime) > time.Second*time.Duration(maxWaitTimeSeconds) {
			break
		}

		taskResponse, httpResp, err := citrixdaasclient.ExecuteWithRetry[*citrixquickcreate.GetTaskAsync200Response](getTaskRequest, client)
		if err != nil {
			diagnostics.AddError(
				"Error polling task: "+taskId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadQcsClientError(err),
			)
			return nil, httpResp, err
		} else if taskResponse != nil &&
			((taskResponse.ResourceConnectionTask != nil && taskResponse.ResourceConnectionTask.GetTaskState() == citrixquickcreate.TASKSTATE_ERROR) ||
				(taskResponse.DeploymentTask != nil && taskResponse.DeploymentTask.GetTaskState() == citrixquickcreate.TASKSTATE_ERROR)) {
			diagnostics.AddError(
				"Task failed: "+taskId,
				"TransactionId: "+citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp)+
					"\nError message: "+ReadQcsClientError(err),
			)
			return nil, httpResp, err
		} else if taskResponse != nil &&
			((taskResponse.ResourceConnectionTask != nil && taskResponse.ResourceConnectionTask.GetTaskState() == citrixquickcreate.TASKSTATE_COMPLETED) ||
				(taskResponse.DeploymentTask != nil && taskResponse.DeploymentTask.GetTaskState() == citrixquickcreate.TASKSTATE_COMPLETED)) {
			return taskResponse, httpResp, nil
		}

		time.Sleep(time.Second * time.Duration(pollIntervalSeconds))
		continue
	}

	return taskResponse, httpResp, err
}

func GetCCAdminAccessPolicyNameKey(r ccadmins.AdministratorAccessPolicyModel) string {
	return r.GetDisplayName()
}
