// Copyright © 2026. Citrix Systems, Inc.

package test

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/citrix/terraform-provider-citrix/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestSchemaDescriptionsPeriod validates that all schema descriptions end with a period
func TestSchemaDescriptionsPeriod(t *testing.T) {
	var failures []string

	iterateSchemas(func(resourceName, resourceType, attrName string, attr interface{}) {
		v := reflect.ValueOf(attr)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}

		description := v.FieldByName("Description").String()
		if description == "" {
			description = v.FieldByName("MarkdownDescription").String()
		}
		if description == "" || strings.HasSuffix(description, ".") {
			return
		}

		location := findSchemaLocation(resourceName)
		failures = append(failures, fmt.Sprintf("%s %s.%s[%s]: description must end with a period",
			location, resourceType, resourceName, attrName))
	})

	if len(failures) > 0 {
		t.Errorf("Schema descriptions missing period:\n%s", strings.Join(failures, "\n"))
	}
}

// TestSchemaDescriptionsEnum validates that enum values are documented in descriptions
func TestSchemaDescriptionsEnum(t *testing.T) {
	var failures []string

	iterateSchemas(func(resourceName, resourceType, attrName string, attr interface{}) {
		// Skip certain attributes that have large enum sets
		if shouldIgnoreEnumAttribute(attrName) {
			return
		}

		v := reflect.ValueOf(attr)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}

		description := v.FieldByName("Description").String()
		if description == "" {
			description = v.FieldByName("MarkdownDescription").String()
		}
		if description == "" {
			return
		}

		// Check for ignore directive
		if strings.Contains(description, "//schema:ignore-enum") {
			return
		}

		validatorsField := v.FieldByName("Validators")
		if !validatorsField.IsValid() || validatorsField.IsNil() {
			return
		}

		enumValues := extractEnumValues(validatorsField)
		if len(enumValues) == 0 {
			return
		}

		if !containsEnumDocumentation(description, enumValues) {
			location := findSchemaLocation(resourceName)
			failures = append(failures, fmt.Sprintf("%s %s.%s[%s]: description [%s] should list enum values %v",
				location, resourceType, resourceName, attrName, description, enumValues))
		}
	})

	if len(failures) > 0 {
		t.Errorf("Schema descriptions missing enum documentation:\n%s", strings.Join(failures, "\n"))
	}
}

// TestSchemaDescriptionsDefault validates that default values are documented in descriptions
func TestSchemaDescriptionsDefault(t *testing.T) {
	var failures []string

	iterateSchemas(func(resourceName, resourceType, attrName string, attr interface{}) {
		v := reflect.ValueOf(attr)
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}

		description := v.FieldByName("Description").String()
		if description == "" {
			description = v.FieldByName("MarkdownDescription").String()
		}
		if description == "" {
			return
		}

		defaultField := v.FieldByName("Default")
		if !defaultField.IsValid() || defaultField.IsNil() {
			return
		}

		defaultValue := extractDefaultValue(defaultField)
		if defaultValue == "" {
			return
		}

		if !containsDefaultDocumentation(description, defaultValue) {
			location := findSchemaLocation(resourceName)
			failures = append(failures, fmt.Sprintf("%s %s.%s[%s]: description should mention default value %q",
				location, resourceType, resourceName, attrName, defaultValue))
		}
	})

	if len(failures) > 0 {
		t.Errorf("Schema descriptions missing default documentation:\n%s", strings.Join(failures, "\n"))
	}
}

// iterateSchemas calls the given function for each attribute in all resources and datasources
func iterateSchemas(fn func(resourceName, resourceType, attrName string, attr interface{})) {
	ctx := context.Background()
	p := provider.New("test")()

	// Test all resources
	for _, factory := range p.Resources(ctx) {
		res := factory()
		var metadata resource.MetadataResponse
		var schema resource.SchemaResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "citrix"}, &metadata)
		res.Schema(ctx, resource.SchemaRequest{}, &schema)

		for attrName, attr := range schema.Schema.Attributes {
			fn(metadata.TypeName, "resource", attrName, attr)
		}
	}

	// Test all datasources
	for _, factory := range p.DataSources(ctx) {
		ds := factory()
		var metadata datasource.MetadataResponse
		var schema datasource.SchemaResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "citrix"}, &metadata)
		ds.Schema(ctx, datasource.SchemaRequest{}, &schema)

		for attrName, attr := range schema.Schema.Attributes {
			fn(metadata.TypeName, "datasource", attrName, attr)
		}
	}
}

func findSchemaLocation(resourceName string) string {
	// Extract the resource type name (e.g., "citrix_machine_catalog" -> "machine_catalog")
	parts := strings.Split(resourceName, "_")
	if len(parts) < 2 {
		return ""
	}
	resourceType := strings.Join(parts[1:], "_")

	// Handle special cases based on resource name patterns
	if strings.Contains(resourceType, "quickcreate") || strings.Contains(resourceType, "qcs") {
		return fmt.Sprintf("internal/quickcreate/%s/%s_resource_model.go", resourceType, resourceType)
	}
	if strings.HasPrefix(resourceType, "stf_") || strings.Contains(resourceName, "storefront") {
		cleanType := strings.TrimPrefix(resourceType, "stf_")
		return fmt.Sprintf("internal/storefront/%s/%s_resource_model.go", cleanType, cleanType)
	}

	// Default to daas path
	return fmt.Sprintf("internal/daas/%s/%s_resource_model.go", resourceType, resourceType)
}

func shouldIgnoreEnumAttribute(attrName string) bool {
	ignoredAttrs := []string{
		"minimum_functional_level",
		"timezone",
	}

	for _, ignored := range ignoredAttrs {
		if attrName == ignored {
			return true
		}
	}
	return false
}

// extractEnumValues extracts enum values from OneOf/OneOfCaseInsensitive validators
// Filters out common "null" values like "Unknown"
func extractEnumValues(validatorsField reflect.Value) []string {
	ignoredValues := map[string]bool{
		"Unknown": true,
	}

	for i := range validatorsField.Len() {
		validator := validatorsField.Index(i)
		if validator.Kind() == reflect.Interface {
			validator = validator.Elem()
		}

		// Skip NoneOf validators - we only want OneOf validators
		validatorType := validator.Type().String()
		if strings.Contains(strings.ToLower(validatorType), "noneof") {
			continue
		}

		// Look for "values" field (present in oneOfValidator and oneOfCaseInsensitiveValidator)
		valuesField := validator.FieldByName("values")
		if !valuesField.IsValid() || valuesField.Kind() != reflect.Slice {
			continue
		}

		var enumValues []string
		for j := range valuesField.Len() {
			val := valuesField.Index(j)
			// types.String has a ValueString() method, but we can also access via reflection
			if val.Kind() == reflect.Struct {
				if strField := val.FieldByName("value"); strField.IsValid() {
					value := strField.String()
					if !ignoredValues[value] {
						enumValues = append(enumValues, value)
					}
				}
			}
		}
		if len(enumValues) > 0 {
			return enumValues
		}
	}
	return nil
}

// extractDefaultValue extracts the default value from Default field
func extractDefaultValue(defaultField reflect.Value) string {
	if defaultField.Kind() == reflect.Ptr && !defaultField.IsNil() {
		defaultField = defaultField.Elem()
	}
	if defaultField.Kind() != reflect.Struct {
		return ""
	}

	// All default types (StaticString, StaticBool, StaticInt64, etc.) have a "value" field
	if valueField := defaultField.FieldByName("Value"); valueField.IsValid() {
		return fmt.Sprint(valueField.Interface())
	}
	return ""
}

// containsEnumDocumentation checks if description mentions enum values
func containsEnumDocumentation(desc string, enumValues []string) bool {
	descLower := strings.ToLower(desc)
	count := 0
	for _, val := range enumValues {
		valLower := strings.ToLower(val)
		// Check exact match or if the value is a substantial substring (handles plural/singular variations)
		if strings.Contains(descLower, valLower) || strings.Contains(desc, "`"+val+"`") {
			count++
		} else if len(val) > 5 && strings.Contains(descLower, valLower[:len(valLower)-1]) {
			// Check if description contains the value minus last character (e.g., "Identity" in "Identities")
			count++
		}
	}

	return count >= len(enumValues)
}

// containsDefaultDocumentation checks if description mentions the default value
func containsDefaultDocumentation(desc string, defaultValue string) bool {
	descLower := strings.ToLower(desc)
	if !strings.Contains(descLower, "default") {
		return false
	}
	return strings.Contains(descLower, strings.ToLower(defaultValue)) || strings.Contains(desc, "`"+defaultValue+"`")
}
