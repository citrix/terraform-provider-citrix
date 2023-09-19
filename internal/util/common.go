package util

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	citrixorchestration "github.com/citrix/citrix-daas-rest-go/citrixorchestration"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	msg := err.(*citrixorchestration.GenericOpenAPIError).Body()
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

func GetJobIdFromHttpResponse(httpResponse http.Response) string {
	locationHeader := httpResponse.Header.Get("Location")
	locationHeaderParts := strings.Split(locationHeader, "/")
	jobId := locationHeaderParts[len(locationHeaderParts)-1]

	return jobId
}

func GetTransactionIdFromHttpResponse(httpResponse *http.Response) string {
	if httpResponse == nil {
		return "failed before request was sent"
	}
	return httpResponse.Header.Get("Citrix-TransactionId")
}

func TypeBoolToString(from types.Bool) string {
	return strconv.FormatBool(from.ValueBool())
}

func StringToTypeBool(from string) types.Bool {
	result, _ := strconv.ParseBool(from)
	return types.BoolValue(result)
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
