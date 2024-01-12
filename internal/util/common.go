// Copyright © 2023. Citrix Systems, Inc.

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"

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

const DomainFqdnRegex string = `^(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`
const UpnRegex string = `^[^@]+@\b(([a-zA-Z0-9-_]){1,63}\.)+[a-zA-Z]{2,63}$`
const GuidRegex string = `^[0-9a-fA-F]{8}-([0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}[}]?$`

type NameValueStringPairModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func ParseNameValueStringPairToClientModel(stringPairs []NameValueStringPairModel) *[]citrixorchestration.NameValueStringPairModel {
	var res = &[]citrixorchestration.NameValueStringPairModel{}
	for _, stringPair := range stringPairs {
		name := stringPair.Name.ValueString()
		value := stringPair.Value.ValueString()
		*res = append(*res, citrixorchestration.NameValueStringPairModel{
			Name:  *citrixorchestration.NewNullableString(&name),
			Value: *citrixorchestration.NewNullableString(&value),
		})
	}
	return res
}

func ParseNameValueStringPairToPluginModel(stringPairs []citrixorchestration.NameValueStringPairModel) *[]NameValueStringPairModel {
	var res = &[]NameValueStringPairModel{}
	for _, stringPair := range stringPairs {
		*res = append(*res, NameValueStringPairModel{
			Name:  types.StringValue(stringPair.GetName()),
			Value: types.StringValue(stringPair.GetValue()),
		})
	}
	return res
}

func AppendNameValueStringPair(stringPairs *[]citrixorchestration.NameValueStringPairModel, name string, appendValue string) {
	*stringPairs = append(*stringPairs, citrixorchestration.NameValueStringPairModel{
		Name:  *citrixorchestration.NewNullableString(&name),
		Value: *citrixorchestration.NewNullableString(&appendValue),
	})
}

func IsValidUUIDorNull(u basetypes.StringValue) bool {
	if u.IsNull() {
		return true
	}
	return IsValidUUID(u.ValueString())
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

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

func ConvertBaseStringArrayToPrimitiveStringArray(v []types.String) []string {
	res := []string{}
	for _, stringVal := range v {
		res = append(res, stringVal.ValueString())
	}

	return res
}

func ConvertPrimitiveStringArrayToBaseStringArray(v []string) []types.String {
	res := []types.String{}
	for _, stringVal := range v {
		res = append(res, types.StringValue(stringVal))
	}

	return res
}

func TypeBoolToString(from types.Bool) string {
	return strconv.FormatBool(from.ValueBool())
}

func StringToTypeBool(from string) types.Bool {
	result, _ := strconv.ParseBool(from)
	return types.BoolValue(result)
}

func ConvertToString(model any) (string, error) {
	body, err := json.Marshal(model)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func GetValidatorFromEnum[V ~string, T []V](enum T) validator.String {
	var values []string
	for _, i := range enum {
		values = append(values, string(i))
	}
	return stringvalidator.OneOfCaseInsensitive(
		values...,
	)
}

func ReadResource[ResponseType any](request any, ctx context.Context, client *citrixdaasclient.CitrixDaasClient, resp *resource.ReadResponse, resourceType, resourceIdOrName string) (ResponseType, *http.Response, error) {
	response, httpResp, err := citrixdaasclient.ExecuteWithRetry[ResponseType](request, client)
	if err != nil && resp != nil {
		if httpResp.StatusCode == http.StatusNotFound {
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

func ProcessAsyncJobResponse(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, jobResp *http.Response, errContext string, diagnostics *diag.Diagnostics, maxTimeout int) (err error) {
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
	}

	return nil
}

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
