// Copyright © 2024. Citrix Systems, Inc.

package util

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type ModelWithAttributes interface {
	GetAttributes() map[string]schema.Attribute // workaround because NestedAttributeObject and SingleNestedAttribute do not share a base type
}

// Store the attribute map for each model type so we don't have to regenerate it every time
var attributeMapCache sync.Map

// Store the default object for each object type so we don't have to regenerate it every time
var defaultObjectCache sync.Map

// <summary>
// Helper function to convert a model to a map of attribute types. Used when converting back to a types.Object
// </summary>
// <param name="m">Model to convert, must implement the ModelWithSchema interface</param>
// <returns>Map of attribute types</returns>
func AttributeMapFromObject(m ModelWithAttributes) (map[string]attr.Type, error) {
	keyName := reflect.TypeOf(m).String()
	if attributes, ok := attributeMapCache.Load(keyName); ok {
		return attributes.(map[string]attr.Type), nil
	}

	// not doing an extra sync/double checked lock because generating the attribute map is pretty quick
	attributeMap, err := attributeMapFromSchema(m.GetAttributes())
	if err != nil {
		return nil, err
	}
	attributeMapCache.Store(keyName, attributeMap)
	return attributeMap, nil
}

// <summary>
// Helper function to convert a schema map to a map of attribute types. Used when converting back to a types.Object
// </summary>
// <param name="s">Schema map of the object</param>
// <returns>Map of attribute types</returns>
func attributeMapFromSchema(s map[string]schema.Attribute) (map[string]attr.Type, error) {
	var attributeTypes = map[string]attr.Type{}
	for attributeName, attribute := range s {
		attrib, err := attributeToTerraformType(attribute)
		if err != nil {
			return nil, err
		}
		attributeTypes[attributeName] = attrib
	}
	return attributeTypes, nil
}

// Converts a schema.Attribute to a terraform attr.Type. Will recurse if the attribute contains a nested object or list of nested objects.
func attributeToTerraformType(attribute schema.Attribute) (attr.Type, error) {
	switch attrib := attribute.(type) {
	case schema.StringAttribute:
		return types.StringType, nil
	case schema.BoolAttribute:
		return types.BoolType, nil
	case schema.NumberAttribute:
		return types.NumberType, nil
	case schema.Int64Attribute:
		return types.Int64Type, nil
	case schema.Float64Attribute:
		return types.Float64Type, nil
	case schema.ListAttribute:
		return types.ListType{ElemType: attrib.ElementType}, nil
	case schema.ListNestedAttribute:
		// list of object, recurse
		nestedAttributes, err := attributeMapFromSchema(attrib.NestedObject.Attributes)
		if err != nil {
			return nil, err
		}
		return types.ListType{ElemType: types.ObjectType{AttrTypes: nestedAttributes}}, nil
	case schema.ObjectAttribute:
		return attrib.GetType(), nil
	case schema.SingleNestedAttribute:
		// object, recurse
		nestedAttributes, err := attributeMapFromSchema(attrib.Attributes)
		if err != nil {
			return nil, err
		}
		return types.ObjectType{AttrTypes: nestedAttributes}, nil
	case schema.SetAttribute:
		return types.SetType{ElemType: attrib.ElementType}, nil
	case schema.SetNestedAttribute:
		// set of object, recurse
		nestedAttributes, err := attributeMapFromSchema(attrib.NestedObject.Attributes)
		if err != nil {
			return nil, err
		}
		return types.SetType{ElemType: types.ObjectType{AttrTypes: nestedAttributes}}, nil
	case schema.MapAttribute:
		return types.MapType{ElemType: attrib.ElementType}, nil
	case schema.MapNestedAttribute:
		// map of object, recurse
		nestedAttributes, err := attributeMapFromSchema(attrib.NestedObject.Attributes)
		if err != nil {
			return nil, err
		}
		return types.MapType{ElemType: types.ObjectType{AttrTypes: nestedAttributes}}, nil
	}
	return nil, fmt.Errorf("unsupported attribute type: %s", attribute)
}

// Helper function to get and cache the default object including populating nested types.List and types.Object so they aren't nil
func defaultObjectFromObjectValue[objTyp any](ctx context.Context, v types.Object) objTyp {
	var temp objTyp
	keyName := reflect.TypeOf(temp).String()
	if defaultObject, ok := defaultObjectCache.Load(keyName); ok {
		return defaultObject.(objTyp)
	}

	// not doing an extra sync/double checked lock because generating the default object is pretty quick
	// Use reflect to build a top level map from tfsdk:field_name to the reflect field value
	attributeByTag := map[string]reflect.Value{}
	val := reflect.ValueOf(&temp).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if tag, ok := field.Tag.Lookup("tfsdk"); ok {
			attributeByTag[tag] = val.Field(i)
		}
	}

	m := v.AttributeTypes(ctx)
	for attributeName, attributeVal := range m {
		if reflectAttribute, ok := attributeByTag[attributeName]; ok {
			// If this attribute is a nested attribute, use the reflect field to create a new null/unknown with the proper attributeMap
			// If this isn't done the framework will return errors like "Value Conversion Error, Expected framework type from provider logic ... Received framework type from provider logic: types._____[]"
			if attributeVal, ok := attributeVal.(types.ObjectType); ok {
				attributeMap := attributeVal.AttributeTypes()
				if v.IsNull() {
					reflectAttribute.Set(reflect.ValueOf(types.ObjectNull(attributeMap)))
				} else {
					reflectAttribute.Set(reflect.ValueOf(types.ObjectUnknown(attributeMap)))
				}
			}
			if attributeVal, ok := attributeVal.(types.ListType); ok {
				elemType := attributeVal.ElementType()
				if v.IsNull() {
					reflectAttribute.Set(reflect.ValueOf(types.ListNull(elemType)))
				} else {
					reflectAttribute.Set(reflect.ValueOf(types.ListUnknown(elemType)))
				}
			}
			if attributeVal, ok := attributeVal.(types.SetType); ok {
				elemType := attributeVal.ElementType()
				if v.IsNull() {
					reflectAttribute.Set(reflect.ValueOf(types.SetNull(elemType)))
				} else {
					reflectAttribute.Set(reflect.ValueOf(types.SetUnknown(elemType)))
				}
			}
			if attributeVal, ok := attributeVal.(types.MapType); ok {
				elemType := attributeVal.ElementType()
				if v.IsNull() {
					reflectAttribute.Set(reflect.ValueOf(types.MapNull(elemType)))
				} else {
					reflectAttribute.Set(reflect.ValueOf(types.MapUnknown(elemType)))
				}
			}
		}
	}
	defaultObjectCache.Store(keyName, temp)
	return temp
}

// <summary>
// Helper function to convert a native terraform object to a golang object of the specified type.
// Use TypedObjectToObjectValue to go the other way.
// </summary>
// <param name="ctx">context</param>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Object in the native terraform types.Object wrapper</param>
// <returns>Object of the specified type</returns>
func ObjectValueToTypedObject[objTyp any](ctx context.Context, diagnostics *diag.Diagnostics, v types.Object) objTyp {
	temp := defaultObjectFromObjectValue[objTyp](ctx, v)
	if v.IsNull() || v.IsUnknown() {
		return temp
	}

	diags := v.As(ctx, &temp, basetypes.ObjectAsOptions{})
	if diags != nil {
		diagnostics.Append(diags...)
	}
	return temp
}

// <summary>
// Helper function to convert a golang object to a native terraform object.
// Use ObjectValueToTypedObject to go the other way.
// </summary>
// <param name="ctx">"context</param>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Object of the specified type</param>
// <param name="s">Schema map of the object</param>
// <returns>Object in the native terraform types.Object wrapper</returns>
func TypedObjectToObjectValue(ctx context.Context, diagnostics *diag.Diagnostics, v ModelWithAttributes) types.Object {
	attributesMap, err := AttributeMapFromObject(v)
	if err != nil {
		diagnostics.AddError("Error converting schema to attribute map", err.Error())
	}
	if v == nil {
		return types.ObjectNull(attributesMap)
	}

	obj, diags := types.ObjectValueFrom(ctx, attributesMap, v)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.ObjectUnknown(attributesMap)
	}
	return obj
}

// <summary>
// Helper function to convert a native terraform list of objects to a golang slice of the specified type
// Use TypedArrayToObjectList to go the other way.
// </summary>
// <param name="ctx">context</param>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">List of object in the native terraform types.List wrapper</param>
// <returns>Array of the specified type</returns>
func ObjectListToTypedArray[objTyp any](ctx context.Context, diagnostics *diag.Diagnostics, v types.List) []objTyp {
	res := make([]types.Object, 0, len(v.Elements()))
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	// convert to slice of TF type
	diags := v.ElementsAs(ctx, &res, false)
	if diags != nil {
		diagnostics.Append(diags...)
		return nil
	}

	// convert to slice of real objects
	arr := make([]objTyp, 0, len(res))
	for _, val := range res {
		arr = append(arr, ObjectValueToTypedObject[objTyp](ctx, diagnostics, val))
	}
	return arr
}

// <summary>
// Helper function to convert a golang slice to a native terraform list of objects.
// Use ObjectListToTypedArray to go the other way.
// </summary>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Slice of objects</param>
// <returns>types.List</returns>
func TypedArrayToObjectList[objTyp ModelWithAttributes](ctx context.Context, diagnostics *diag.Diagnostics, v []objTyp) types.List {
	var t objTyp
	attributesMap, err := AttributeMapFromObject(t)
	if err != nil {
		diagnostics.AddError("Error converting schema to attribute map", err.Error())
	}

	if v == nil {
		return types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	}

	res := make([]types.Object, 0, len(v))
	for _, val := range v {
		res = append(res, TypedObjectToObjectValue(ctx, diagnostics, val))
	}
	list, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: attributesMap}, res)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.ListNull(types.ObjectType{AttrTypes: attributesMap})
	}
	return list
}

// <summary>
// Helper function to convert a native terraform list of objects to a golang slice of the specified type
// Use TypedArrayToObjectSet to go the other way.
// </summary>
// <param name="ctx">context</param>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">List of object in the native terraform types.List wrapper</param>
// <returns>Array of the specified type</returns>
func ObjectSetToTypedArray[objTyp any](ctx context.Context, diagnostics *diag.Diagnostics, v types.Set) []objTyp {
	res := make([]types.Object, 0, len(v.Elements()))
	if v.IsNull() || v.IsUnknown() {
		return nil
	}

	// convert to slice of TF type
	diags := v.ElementsAs(ctx, &res, false)
	if diags != nil {
		diagnostics.Append(diags...)
		return nil
	}

	// convert to slice of real objects
	arr := make([]objTyp, 0, len(res))
	for _, val := range res {
		arr = append(arr, ObjectValueToTypedObject[objTyp](ctx, diagnostics, val))
	}
	return arr
}

// <summary>
// Helper function to convert a golang slice to a native terraform list of objects.
// Use ObjectSetToTypedArray to go the other way.
// </summary>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Slice of objects</param>
// <returns>types.List</returns>
func TypedArrayToObjectSet[objTyp ModelWithAttributes](ctx context.Context, diagnostics *diag.Diagnostics, v []objTyp) types.Set {
	var t objTyp
	attributesMap, err := AttributeMapFromObject(t)
	if err != nil {
		diagnostics.AddError("Error converting schema to attribute map", err.Error())
	}

	if v == nil {
		return types.SetNull(types.ObjectType{AttrTypes: attributesMap})
	}

	res := make([]types.Object, 0, len(v))
	for _, val := range v {
		res = append(res, TypedObjectToObjectValue(ctx, diagnostics, val))
	}
	set, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: attributesMap}, res)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.SetNull(types.ObjectType{AttrTypes: attributesMap})
	}
	return set
}

// <summary>
// Helper function to convert a terraform list of terraform strings to array of golang primitive strings.
// Use StringArrayToStringList to go the other way.
// </summary>
// <param name="v">List of terraform strings</param>
// <returns>Array of golang primitive strings</returns>
func StringListToStringArray(ctx context.Context, diagnostics *diag.Diagnostics, v types.List) []string {
	res := make([]types.String, 0, len(v.Elements()))

	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	// convert to slice of TF type
	diags := v.ElementsAs(ctx, &res, false)
	if diags != nil {
		diagnostics.Append(diags...)
		return nil
	}

	arr := []string{}
	for _, stringVal := range res {
		arr = append(arr, stringVal.ValueString())
	}

	return arr
}

// <summary>
// Helper function to convert a golang slice of string to a native terraform list of strings.
// Use StringListToStringArray to go the other way.
// </summary>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Slice of strings</param>
// <returns>types.List</returns>
func StringArrayToStringList(ctx context.Context, diagnostics *diag.Diagnostics, v []string) types.List {
	if v == nil {
		return types.ListNull(types.StringType)
	}

	res := make([]types.String, 0, len(v))
	for _, val := range v {
		res = append(res, basetypes.NewStringValue(val))
	}
	list, diags := types.ListValueFrom(ctx, types.StringType, res)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.ListNull(types.StringType)
	}
	return list
}

// <summary>
// Helper function to convert a terraform set of terraform strings to array of golang primitive strings.
// Use StringArrayToStringSet to go the other way.
// </summary>
// <param name="v">Set of terraform strings</param>
// <returns>Array of golang primitive strings</returns>
func StringSetToStringArray(ctx context.Context, diagnostics *diag.Diagnostics, v types.Set) []string {
	res := make([]types.String, 0, len(v.Elements()))

	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	// convert to slice of TF type
	diags := v.ElementsAs(ctx, &res, false)
	if diags != nil {
		diagnostics.Append(diags...)
		return nil
	}

	arr := []string{}
	for _, stringVal := range res {
		arr = append(arr, stringVal.ValueString())
	}

	return arr
}

// <summary>
// Helper function to convert a golang slice of string to a native terraform set of strings.
// Use StringSetToStringArray to go the other way.
// </summary>
// <param name="diagnostics">Any issues will be appended to these diagnostics</param>
// <param name="v">Slice of strings</param>
// <returns>types.Set</returns>
func StringArrayToStringSet(ctx context.Context, diagnostics *diag.Diagnostics, v []string) types.Set {
	if v == nil {
		return types.SetNull(types.StringType)
	}

	res := make([]types.String, 0, len(v))
	for _, val := range v {
		res = append(res, basetypes.NewStringValue(val))
	}
	set, diags := types.SetValueFrom(ctx, types.StringType, res)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.SetNull(types.StringType)
	}
	return set
}

// <summary>
// Helper function to convert array of terraform strings to array of golang primitive strings
// Deprecated: Remove after we fully move to types.List
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
// Helper function to convert array of golang primitive interface to native terraform list of strings
// </summary>
// <param name="v">Array of golang primitive interface</param>
// <returns>Terraform list of strings</returns>
func ConvertPrimitiveInterfaceArrayToStringList(ctx context.Context, diagnostics *diag.Diagnostics, v []interface{}) (types.List, string) {
	if v == nil {
		return types.ListNull(types.StringType), ""
	}

	res := make([]types.String, 0, len(v))
	for _, val := range v {
		switch stringVal := val.(type) {
		case string:
			res = append(res, basetypes.NewStringValue(stringVal))
		default:
			return types.ListNull(types.StringType), "At this time, only string values are supported in arrays."
		}
	}

	resList, diags := types.ListValueFrom(ctx, types.StringType, res)
	if diags != nil {
		diagnostics.Append(diags...)
		return types.ListNull(types.StringType), "An error occurred when converting base string array to list of strings"
	}

	return resList, ""
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
