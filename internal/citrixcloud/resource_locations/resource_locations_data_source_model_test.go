// Copyright © 2026. Citrix Systems, Inc.

package resource_locations // whitebox testing in same package

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Read converts config into the model, which fails if the model declares fields the data source schema does not have.
func TestResourceLocationDataSourceConfigConvertsToModel(t *testing.T) {
	ctx := context.Background()
	dataSourceSchema := ResourceLocationDataSourceModel{}.GetSchema()
	objectType, ok := dataSourceSchema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)

	values := map[string]tftypes.Value{}
	for name, attributeType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}
	values["name"] = tftypes.NewValue(tftypes.String, "CME DEV GCP")
	config := tfsdk.Config{Schema: dataSourceSchema, Raw: tftypes.NewValue(objectType, values)}

	var data ResourceLocationDataSourceModel
	diagnostics := config.Get(ctx, &data)

	require.False(t, diagnostics.HasError(), "unexpected diagnostics: %v", diagnostics)
	assert.Equal(t, "CME DEV GCP", data.Name.ValueString())
}
