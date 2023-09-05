package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceWithZoneModel interface {
	GetZone() types.String
}
