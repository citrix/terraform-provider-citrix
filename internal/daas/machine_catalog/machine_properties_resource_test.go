// Copyright © 2026. Citrix Systems, Inc.

package machine_catalog // whitebox testing in same package

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newMachineTag creates a TagResponseModel as the Machines Tags endpoint returns it. numMachines is
// scoped to the queried machine: 1 when assigned directly, 0 when only inherited.
func newMachineTag(id, name string, numMachines int32) citrixorchestration.TagResponseModel {
	tag := citrixorchestration.TagResponseModel{}
	tag.SetId(id)
	tag.SetName(name)
	tag.SetNumMachines(numMachines)
	return tag
}

func TestGetMachineTagIds(t *testing.T) {
	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name           string
		tags           []citrixorchestration.TagResponseModel
		expectedTagIds []string
	}{
		{
			name: "Excludes tags inherited from a delivery group",
			tags: []citrixorchestration.TagResponseModel{
				newMachineTag("direct-tag-id", "direct", 1),
				newMachineTag("dg-tag-id", "inherited-from-delivery-group", 0),
			},
			expectedTagIds: []string{"direct-tag-id"},
		},
		{
			name: "Returns no tags when every tag is inherited",
			tags: []citrixorchestration.TagResponseModel{
				newMachineTag("dg-tag-id", "inherited-from-delivery-group", 0),
				newMachineTag("app-tag-id", "inherited-from-application", 0),
			},
			expectedTagIds: []string{},
		},
		{
			// Pins the predicate to the documented "equal to 1" rule rather than a "non-zero" reading.
			name: "Excludes tags whose machine count is not exactly one",
			tags: []citrixorchestration.TagResponseModel{
				newMachineTag("direct-tag-id", "direct", 1),
				newMachineTag("wider-scope-tag-id", "counted-across-many-machines", 50),
			},
			expectedTagIds: []string{"direct-tag-id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)

			mockMachinesAPI := citrixorchestration.GetMockMachinesAPIsDAAS(mockClient.APIClient)
			ctx := test.DaaSTestContext()

			tagCollection := &citrixorchestration.TagResponseModelCollection{}
			tagCollection.SetItems(tt.tags)
			mockMachinesAPI.On("MachinesGetMachineTagsExecute", mock.Anything).Return(
				tagCollection, testutil.MockSuccessResponse(), nil).Once()

			var diagnostics diag.Diagnostics
			tagIds, err := getMachineTagIds(ctx, client, &diagnostics, "domain|machine")

			require.NoError(t, err)
			assert.False(t, diagnostics.HasError())
			assert.ElementsMatch(t, tt.expectedTagIds, tagIds)
		})
	}
}
