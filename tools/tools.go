// Copyright © 2024. Citrix Systems, Inc.

//go:build tools

package tools

import (
	// Ensure documentation generator is not removed from go.mod.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
