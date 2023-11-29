// Copyright © 2023. Citrix Systems, Inc.

package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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
