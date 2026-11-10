// Copyright © 2026. Citrix Systems, Inc.

package util // whitebox testing in same package

import (
	"net/http"
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	"github.com/citrix/citrix-daas-rest-go/test"
	testutil "github.com/citrix/terraform-provider-citrix/internal/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// newZoneForResourceLocation creates a ZoneResponseModel synced from the given resource location.
// Every zone shares one name so tests cover a stale zone with the same name as the current one.
func newZoneForResourceLocation(id, resourceLocationId string) citrixorchestration.ZoneResponseModel {
	zone := citrixorchestration.ZoneResponseModel{}
	zone.SetId(id)
	zone.SetName("West Europe")
	resourceLocation := citrixorchestration.RefResponseModel{}
	resourceLocation.SetId(resourceLocationId)
	zone.SetResourceLocation(resourceLocation)
	return zone
}

// newZoneCollection creates a single page ZoneResponseModelCollection.
func newZoneCollection(zones ...citrixorchestration.ZoneResponseModel) *citrixorchestration.ZoneResponseModelCollection {
	collection := &citrixorchestration.ZoneResponseModelCollection{}
	collection.SetItems(zones)
	return collection
}

// ========== Tests ==========

func TestGetZoneIdForResourceLocation(t *testing.T) {
	staleZone := newZoneForResourceLocation("zone-old", "rl-deleted")
	currentZone := newZoneForResourceLocation("zone-new", "RL-Current")

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name               string
		zones              []citrixorchestration.ZoneResponseModel
		resourceLocationId string
		expectedZoneId     string
	}{
		{
			name:               "Matches by resource location id when a stale zone has the same name",
			zones:              []citrixorchestration.ZoneResponseModel{staleZone, currentZone},
			resourceLocationId: "RL-Current",
			expectedZoneId:     "zone-new",
		},
		{
			name:               "Matches resource location id case insensitively",
			zones:              []citrixorchestration.ZoneResponseModel{currentZone},
			resourceLocationId: "rl-current",
			expectedZoneId:     "zone-new",
		},
		{
			name:               "Returns empty when the zone has not synced yet",
			zones:              []citrixorchestration.ZoneResponseModel{staleZone},
			resourceLocationId: "rl-current",
			expectedZoneId:     "",
		},
		{
			name:               "Returns empty when there are no zones",
			zones:              []citrixorchestration.ZoneResponseModel{},
			resourceLocationId: "rl-current",
			expectedZoneId:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zoneId := GetZoneIdForResourceLocation(tt.zones, tt.resourceLocationId)

			assert.Equal(t, tt.expectedZoneId, zoneId)
		})
	}
}

func TestPollZoneForResourceLocation(t *testing.T) {
	zoneDetail := &citrixorchestration.ZoneDetailResponseModel{}
	zoneDetail.SetId("zone-new")
	zoneDetail.SetName("West Europe")

	tests := []struct { // uses Table-driven-tests pattern https://go.dev/wiki/TableDrivenTests
		name           string
		setupMocks     func(*citrixorchestration.MockZonesAPIsDAAS)
		expectedZoneId string
		expectError    bool
	}{
		{
			name: "Returns the zone synced from the resource location",
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS) {
				mockZonesAPI.On("ZonesGetZonesExecute", mock.Anything).Return(
					newZoneCollection(
						newZoneForResourceLocation("zone-old", "rl-deleted"),
						newZoneForResourceLocation("zone-new", "rl-current"),
					),
					testutil.MockSuccessResponse(), nil).Once()
				mockZonesAPI.On("ZonesGetZoneExecute", mock.MatchedBy(func(r citrixorchestration.ApiZonesGetZoneRequest) bool {
					return citrixorchestration.MockApiZonesGetZoneRequest{ApiZonesGetZoneRequest: r}.GetNameOrId() == "zone-new"
				})).Return(zoneDetail, testutil.MockSuccessResponse(), nil).Once()
			},
			expectedZoneId: "zone-new",
		},
		{
			name: "Returns the error when listing zones fails",
			setupMocks: func(mockZonesAPI *citrixorchestration.MockZonesAPIsDAAS) {
				mockZonesAPI.On("ZonesGetZonesExecute", mock.Anything).Return(
					(*citrixorchestration.ZoneResponseModelCollection)(nil),
					testutil.MockErrorResponse(403, "txn-error"),
					assert.AnError,
				).Once()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mockClient := test.NewTestDaaSClient()
			defer mockClient.AssertExpectations(t)

			mockZonesAPI := citrixorchestration.GetMockZonesAPIsDAAS(mockClient.APIClient)
			tt.setupMocks(mockZonesAPI)

			zone, httpResp, err := PollZoneForResourceLocation(test.DaaSTestContext(), client, "rl-current")

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, zone)
				require.NotNil(t, httpResp)
				assert.Equal(t, http.StatusForbidden, httpResp.StatusCode)
			} else {
				require.NoError(t, err)
				require.NotNil(t, zone)
				assert.Equal(t, tt.expectedZoneId, zone.GetId())
			}
		})
	}
}
