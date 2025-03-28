// Copyright © 2024. Citrix Systems, Inc.

package util

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
)

const (
	attributeNameCreate = "create"
	attributeNameRead   = "read"
	attributeNameUpdate = "update"
	attributeNameDelete = "delete"
)

type TimeoutConfigs struct {
	Create        bool
	CreateDefault int32
	CreateMin     int32
	CreateMax     int32

	Read        bool
	ReadDefault int32
	ReadMin     int32
	ReadMax     int32

	Update        bool
	UpdateDefault int32
	UpdateMin     int32
	UpdateMax     int32

	Delete        bool
	DeleteDefault int32
	DeleteMin     int32
	DeleteMax     int32
}

func GetTimeoutSchema(resourceName string, configs TimeoutConfigs) schema.SingleNestedAttribute {
	timeoutOps := []string{}
	if configs.Create {
		timeoutOps = append(timeoutOps, attributeNameCreate)
	}
	if configs.Read {
		timeoutOps = append(timeoutOps, attributeNameRead)
	}
	if configs.Update {
		timeoutOps = append(timeoutOps, attributeNameUpdate)
	}
	if configs.Delete {
		timeoutOps = append(timeoutOps, attributeNameDelete)
	}
	return schema.SingleNestedAttribute{
		Attributes:  GetTimeoutAttributesMap(configs),
		Description: fmt.Sprintf("Timeout in minutes for the long-running jobs in %s resource's %s operation(s).", resourceName, strings.Join(timeoutOps, ", ")),
		Optional:    true,
	}
}

func GetTimeoutAttributesMap(configs TimeoutConfigs) map[string]schema.Attribute {
	descriptionFmt := `Timeout in minutes for the long-running jobs in %s operation. Defaults to %d.`
	minValueDescriptionFmt := ` Minimum value is %d.`
	maxValueDescriptionFmt := ` Maximum value is %d.`
	attributes := map[string]schema.Attribute{}

	if configs.Create {
		createAttribute := schema.Int32Attribute{
			Description: fmt.Sprintf(descriptionFmt, attributeNameCreate, configs.CreateDefault),
			Optional:    true,
			Computed:    true,
			Default:     int32default.StaticInt32(configs.CreateDefault),
		}

		if configs.CreateMin > 0 {
			createAttribute.Description += fmt.Sprintf(minValueDescriptionFmt, configs.CreateMin)
			createAttribute.Validators = append(createAttribute.Validators, int32validator.AtLeast(configs.CreateMin))
		}

		if configs.CreateMax > 0 {
			createAttribute.Description += fmt.Sprintf(maxValueDescriptionFmt, configs.CreateMax)
			createAttribute.Validators = append(createAttribute.Validators, int32validator.AtMost(configs.CreateMax))
		}

		attributes[attributeNameCreate] = createAttribute
	}

	if configs.Read {
		readAttribute := schema.Int32Attribute{
			Description: fmt.Sprintf(descriptionFmt, attributeNameRead, configs.ReadDefault),
			Optional:    true,
			Computed:    true,
			Default:     int32default.StaticInt32(configs.ReadDefault),
		}

		if configs.ReadMin > 0 {
			readAttribute.Description += fmt.Sprintf(minValueDescriptionFmt, configs.ReadMin)
			readAttribute.Validators = append(readAttribute.Validators, int32validator.AtLeast(configs.ReadMin))
		}

		if configs.ReadMax > 0 {
			readAttribute.Description += fmt.Sprintf(maxValueDescriptionFmt, configs.ReadMax)
			readAttribute.Validators = append(readAttribute.Validators, int32validator.AtMost(configs.ReadMax))
		}
		attributes[attributeNameRead] = readAttribute
	}

	if configs.Update {
		updateAttribute := schema.Int32Attribute{
			Description: fmt.Sprintf(descriptionFmt, attributeNameUpdate, configs.UpdateDefault),
			Optional:    true,
			Computed:    true,
			Default:     int32default.StaticInt32(configs.UpdateDefault),
		}

		if configs.UpdateMin > 0 {
			updateAttribute.Description += fmt.Sprintf(minValueDescriptionFmt, configs.UpdateMin)
			updateAttribute.Validators = append(updateAttribute.Validators, int32validator.AtLeast(configs.UpdateMin))
		}

		if configs.UpdateMax > 0 {
			updateAttribute.Description += fmt.Sprintf(maxValueDescriptionFmt, configs.UpdateMax)
			updateAttribute.Validators = append(updateAttribute.Validators, int32validator.AtMost(configs.UpdateMax))
		}
		attributes[attributeNameUpdate] = updateAttribute
	}

	if configs.Delete {
		deleteAttribute := schema.Int32Attribute{
			Description: fmt.Sprintf(descriptionFmt, attributeNameDelete, configs.DeleteDefault),
			Optional:    true,
			Computed:    true,
			Default:     int32default.StaticInt32(configs.DeleteDefault),
		}

		if configs.DeleteMin > 0 {
			deleteAttribute.Description += fmt.Sprintf(minValueDescriptionFmt, configs.DeleteMin)
			deleteAttribute.Validators = append(deleteAttribute.Validators, int32validator.AtLeast(configs.DeleteMin))
		}

		if configs.DeleteMax > 0 {
			deleteAttribute.Description += fmt.Sprintf(maxValueDescriptionFmt, configs.DeleteMax)
			deleteAttribute.Validators = append(deleteAttribute.Validators, int32validator.AtMost(configs.DeleteMax))
		}
		attributes[attributeNameDelete] = deleteAttribute
	}

	return attributes
}
