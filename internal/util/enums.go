// Copyright © 2024. Citrix Systems, Inc.
package util

import (
	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	quickcreateservice "github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
)

func AwsEdcWorkspaceImageIngestionProcessEnumToString(ingestionProcess quickcreateservice.AwsEdcWorkspaceImageIngestionProcess) string {
	switch ingestionProcess {
	case quickcreateservice.AWSEDCWORKSPACEIMAGEINGESTIONPROCESS_REGULAR_BYOP:
		return "BYOL_REGULAR_BYOP"
	case quickcreateservice.AWSEDCWORKSPACEIMAGEINGESTIONPROCESS_GRAPHICS_G4_DN_BYOP:
		return "BYOL_GRAPHICS_G4DN_BYOP"
	default:
		return ""
	}
}

func AwsEdcWorkspaceImageTenancyEnumToString(imageTenancy quickcreateservice.AwsEdcWorkspaceImageTenancy) string {
	switch imageTenancy {
	case quickcreateservice.AWSEDCWORKSPACEIMAGETENANCY_DEDICATED:
		return "DEDICATED"
	case quickcreateservice.AWSEDCWORKSPACEIMAGETENANCY_DEFAULT:
		return "DEFAULT"
	default:
		return ""
	}
}

func AwsEdcWorkspaceImageStateEnumToString(imageState quickcreateservice.AwsEdcWorkspaceImageState) string {
	switch imageState {
	case quickcreateservice.AWSEDCWORKSPACEIMAGESTATE_AVAILABLE:
		return "AVAILABLE"
	case quickcreateservice.AWSEDCWORKSPACEIMAGESTATE_ERROR:
		return "ERROR"
	case quickcreateservice.AWSEDCWORKSPACEIMAGESTATE_PENDING:
		return "PENDING"
	case quickcreateservice.AWSEDCWORKSPACEIMAGESTATE_ERROR_INVALID_ACCOUNT:
		return "ERROR_INVALID_ACCOUNT"
	default:
		return ""
	}
}

func QcsSessionSupportEnumToString(imageState quickcreateservice.SessionSupport) string {
	switch imageState {
	case quickcreateservice.SESSIONSUPPORT_SINGLE_SESSION:
		return "SingleSession"
	case quickcreateservice.SESSIONSUPPORT_MULTI_SESSION:
		return "MultiSession"
	case quickcreateservice.SESSIONSUPPORT_UNKNOWN:
		return "Unknown"
	default:
		return ""
	}
}

func OperatingSystemTypeEnumToString(os quickcreateservice.OperatingSystemType) string {
	switch os {
	case quickcreateservice.OPERATINGSYSTEMTYPE_WINDOWS:
		return "WINDOWS"
	case quickcreateservice.OPERATINGSYSTEMTYPE_LINUX:
		return "LINUX"
	default:
		return ""
	}
}

func ComputeTypeEnumToString(os quickcreateservice.AwsEdcWorkspaceCompute) string {
	switch os {
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_VALUE:
		return "VALUE"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_STANDARD:
		return "STANDARD"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_PERFORMANCE:
		return "PERFORMANCE"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_POWER:
		return "POWER"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_POWERPRO:
		return "POWERPRO"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_GRAPHICS:
		return "GRAPHICS"
	case quickcreateservice.AWSEDCWORKSPACECOMPUTE_GRAPHICSPRO:
		return "GRAPHICSPRO"
	default:
		return ""
	}
}

func RunningModeEnumToString(os quickcreateservice.AwsEdcWorkspaceRunningMode) string {
	switch os {
	case quickcreateservice.AWSEDCWORKSPACERUNNINGMODE_ALWAYS_ON:
		return "ALWAYS_ON"
	case quickcreateservice.AWSEDCWORKSPACERUNNINGMODE_MANUAL:
		return "MANUAL"
	default:
		return ""
	}
}

func TaskStateEnumToString(os quickcreateservice.TaskState) string {
	switch os {
	case quickcreateservice.TASKSTATE_ACTIVE:
		return "ACTIVE"
	case quickcreateservice.TASKSTATE_COMPLETED:
		return "COMPLETED"
	case quickcreateservice.TASKSTATE_ERROR:
		return "ERROR"
	case quickcreateservice.TASKSTATE_PENDING:
		return "PENDING"
	case quickcreateservice.TASKSTATE_PROCESSING:
		return "PROCESSING"
	default:
		return ""
	}
}

func OrchestrationOSTypeEnumToString(os citrixorchestration.OsType) string {
	switch os {
	case citrixorchestration.OSTYPE_WINDOWS:
		return "Windows"
	case citrixorchestration.OSTYPE_LINUX:
		return "Linux"
	default:
		return ""
	}
}

func SessionSupportEnumToString(sessionSupport citrixorchestration.SessionSupport) string {
	switch sessionSupport {
	case citrixorchestration.SESSIONSUPPORT_SINGLE_SESSION:
		return "SingleSession"
	case citrixorchestration.SESSIONSUPPORT_MULTI_SESSION:
		return "MultiSession"
	default:
		return ""
	}
}
