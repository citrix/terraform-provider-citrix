// Copyright © 2023. Citrix Systems, Inc.

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Domain FQDN
const DomainFqdnRegex string = `^(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`

// SAM
const SamRegex string = `^[a-zA-Z][a-zA-Z0-9\- ]{0,61}[a-zA-Z0-9]\\\w[\w\.\- ]+$`

// GUID
const GuidRegex string = `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$`

// IPv4
const IPv4Regex string = `^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`

// IPv4 with https
const IPv4RegexWithProtocol string = `^(http|https)://((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`

// Date YYYY-MM-DD
const DateRegex string = `^\d{4}-\d{2}-\d{2}$`

// Time HH:MM
const TimeRegex string = `^([0-1][0-9]|2[0-3]):[0-5][0-9]$`

// ID of the Default Site Policy Set
const DefaultSitePolicySetId string = "00000000-0000-0000-0000-000000000000"

// SSL Thumbprint
const SslThumbprintRegex string = `^([0-9a-fA-F]{40}|[0-9a-fA-F]{64})$`

// AWS EC2 Instance Type
const AwsEc2InstanceTypeRegex string = `^[a-z0-9]{1,15}\.[a-z0-9]{1,15}$`

// Resource Types
const ImageVersionResourceType string = "ImageVersion"
const RegionResourceType string = "Region"
const ServiceOfferingResourceType string = "ServiceOffering"
const SnapshotResourceType string = "Snapshot"
const VhdResourceType string = "Vhd"
const VirtualPrivateCloudResourceType string = "VirtualPrivateCloud"
const VirtualMachineResourceType string = "Vm"
const TemplateSpecVersionResourceType string = "TemplateSpecVersion"

// Terraform model for name value string pair
type NameValueStringPairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
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

// <summary>
// Helper function to convert array of terraform strings to array of golang primitive strings
// </summary>
// <param name="v">Array of terraform stringsArray of golang primitive strings</param>
// <returns>Array of golang primitive strings</returns>
func ConvertBaseStringArrayToPrimitiveStringArray(v []types.String) []string {
	res := []string{}
	for _, stringVal := range v {
		res = append(res, stringVal.ValueString())
	}

	return res
}

// <summary>
// Helper function to convert array of golang primitive strings to array of terraform strings
// </summary>
// <param name="v">Array of golang primitive strings</param>
// <returns>Array of terraform strings</returns>
func ConvertPrimitiveStringArrayToBaseStringArray(v []string) []types.String {
	res := []types.String{}
	for _, stringVal := range v {
		res = append(res, types.StringValue(stringVal))
	}

	return res
}

// <summary>
// Helper function to convert terraform bool value to string
// </summary>
// <param name="from">Boolean value in terraform bool</param>
// <returns>Boolean value in string</returns>
func TypeBoolToString(from types.Bool) string {
	return strconv.FormatBool(from.ValueBool())
}

// <summary>
// Helper function to convert string to terraform boolean value
// </summary>
// <param name="from">Boolean value in string</param>
// <returns>Boolean value in terraform types.Bool</returns>
func StringToTypeBool(from string) types.Bool {
	result, _ := strconv.ParseBool(from)
	return types.BoolValue(result)
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
		errorDetail := "TransactionId: " + txId +
			"\nJobId: " + jobResponseModel.GetId()

		if jobResponseModel.GetStatus() == citrixorchestration.JOBSTATUS_FAILED {
			errorDetail = errorDetail + "\nError message: " + jobResponseModel.GetErrorString()
		}

		diagnostics.AddError(
			errContext,
			errorDetail,
		)

		if returnJobError {
			return fmt.Errorf(errorDetail)
		}
	}

	return nil
}

// <summary>
// Helper function for calculating the new state of a list of nested attribute, while
// keeping the order of the elements in the array intact, and adds missing elements
// from remote to state.
// Can be used for refreshing all list nested attributes.
// </summary>
// <param name="state">State values in Terraform model</param>
// <param name="tfId">Name of the identifier field in Terraform model</param>
// <param name="remote">Remote values in client model</param>
// <param name="clientId">Name of the identifier field in client model</param>
// <param name="refreshFunc">Name of the refresh properties function defined in the terraform model</param>
// <returns>Array in Terraform model for new state</returns>
func RefreshListProperties[tfType any, clientType any](state []tfType, tfId string, remote []clientType, clientId string, refreshFunc string) []tfType {
	if len(remote) == 0 {
		return nil
	}

	if state == nil {
		state = []tfType{}
	}

	stateItems := map[string]int{}
	for index, item := range state {
		value := reflect.ValueOf(&item).Elem()
		id := value.FieldByName(tfId).Interface().(basetypes.StringValue)
		stateItems[id.ValueString()] = index
	}

	var tfItem tfType
	tfStruct := reflect.TypeOf(tfItem)

	method, _ := tfStruct.MethodByName(refreshFunc)
	newState := state
	var id string
	for _, item := range remote {
		value := reflect.ValueOf(&item).Elem()
		valueType := value.FieldByName(clientId).Type()
		if valueType == reflect.TypeOf(citrixorchestration.NullableString{}) {
			idNullable := value.FieldByName(clientId).Interface().(citrixorchestration.NullableString)
			if idNullable.IsSet() {
				id = *idNullable.Get()
			}
		} else {
			id = value.FieldByName(clientId).Interface().(string)
		}
		index, exists := stateItems[id]
		requestValue := reflect.ValueOf(item)
		if exists {
			newStateItemReflectValue := method.Func.Call([]reflect.Value{reflect.ValueOf(state[index]), requestValue})[0]
			newState[index] = newStateItemReflectValue.Interface().(tfType)
		} else {
			tfStructItem := reflect.New(tfStruct).Elem()
			newStateItemReflectValue := method.Func.Call([]reflect.Value{tfStructItem, requestValue})[0]
			newState = append(newState, newStateItemReflectValue.Interface().(tfType))
		}

		stateItems[id] = -1 // Mark as visited. The ones not visited should be removed.
	}

	result := []tfType{}
	for _, item := range newState {
		value := reflect.ValueOf(&item).Elem()
		id := value.FieldByName(tfId).Interface().(basetypes.StringValue)

		if stateItems[id.ValueString()] == -1 {
			result = append(result, item) // if visited, include. Not visited ones will not be included.
		}
	}

	return result
}

// <summary>
// Helper function for calculating the new state of a list of strings, while
// keeping the order of the elements in the array intact, and adds missing elements
// from remote to state.
// Can be used for refreshing list of strings.
// </summary>
// <param name="state">List of values in state</param>
// <param name="remote">List of values in remote</param>
func RefreshList(state []types.String, remote []string) []types.String {
	stateItems := map[string]bool{}
	for _, item := range state {
		stateItems[strings.ToLower(item.ValueString())] = false // not visited
	}

	for _, item := range remote {
		itemInLowerCase := strings.ToLower(item)
		_, exists := stateItems[itemInLowerCase]
		if !exists {
			state = append(state, types.StringValue(item))
		}
		stateItems[itemInLowerCase] = true
	}

	result := []types.String{}
	for _, item := range state {
		if stateItems[strings.ToLower(item.ValueString())] {
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
		file, err := ioutil.TempFile("", "citrix_provider_crash_stack.*.txt")
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
