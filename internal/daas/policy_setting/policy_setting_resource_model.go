// Copyright © 2024. Citrix Systems, Inc.

package policy_setting

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/citrix/terraform-provider-citrix/internal/daas/policies"
	"github.com/citrix/terraform-provider-citrix/internal/util"
	"github.com/citrix/terraform-provider-citrix/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicySettingModel struct {
	Id         types.String `tfsdk:"id"`
	PolicyId   types.String `tfsdk:"policy_id"`
	Name       types.String `tfsdk:"name"`
	UseDefault types.Bool   `tfsdk:"use_default"`
	Value      types.String `tfsdk:"value"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

func (PolicySettingModel) GetSchema() schema.Schema {
	return schema.Schema{
		Description: "CVAD --- Manages an instance of the Policy Setting." +
			"\n\n -> **Please Note** For detailed information about policy settings, please refer to [this document](https://github.com/citrix/terraform-provider-citrix/blob/main/internal/daas/policies/policy_set_resource.md).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id of the policy setting.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy_id": schema.StringAttribute{
				Description: "Id of the policy to which the setting belongs.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(util.GuidRegex), "must be specified with ID in GUID format"),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the policy setting.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use_default": schema.BoolAttribute{
				Description: "Indicate whether using default value for the policy setting.",
				Required:    true,
				Validators: []validator.Bool{
					validators.AlsoRequiresOneOfOnBoolValues([]bool{false}, path.MatchRelative().AtParent().AtName("value"), path.MatchRelative().AtParent().AtName("enabled")),
					validators.ConflictsWithOnBoolValues([]bool{true}, path.MatchRelative().AtParent().AtName("value"), path.MatchRelative().AtParent().AtName("enabled")),
				},
			},
			"value": schema.StringAttribute{
				Description: "Value of the policy setting.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether of the policy setting has enabled or allowed value.",
				Optional:    true,
			},
		},
	}
}

func (PolicySettingModel) GetAttributes() map[string]schema.Attribute {
	return PolicySettingModel{}.GetSchema().Attributes
}

func (PolicySettingModel) GetAttributesNamesToMask() map[string]bool {
	return map[string]bool{}
}

func buildSettingRequest(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, policySetting PolicySettingModel, action string) (citrixorchestration.SettingRequest, error) {
	settingRequest := citrixorchestration.SettingRequest{}
	defaultBoolSettingValueMap, err := policies.GetGpoBooleanSettingDefaultValueMap(ctx, diagnostics, client)
	if err != nil {
		return settingRequest, err
	}
	settingName := policySetting.Name.ValueString()
	settingRequest.SetSettingName(settingName)
	settingRequest.SetUseDefault(policySetting.UseDefault.ValueBool())
	if policySetting.UseDefault.ValueBool() {
		if defaultBoolSettingValueMap[settingName] != "" {
			settingRequest.SetSettingValue(defaultBoolSettingValueMap[settingName])
		}
	} else if !policySetting.Value.IsNull() {
		settingRequest.SetSettingValue(policySetting.Value.ValueString())
	} else if !policySetting.Enabled.IsNull() {
		if policySetting.Enabled.ValueBool() {
			settingRequest.SetSettingValue("1")
		} else {
			settingRequest.SetSettingValue("0")
		}
	} else {
		err := fmt.Errorf("Policy setting %s has `use_default` set to `false`, but does not have `value` or `enabled` field specified", policySetting.Name.ValueString())
		diagnostics.AddError(
			fmt.Sprintf("Error %s policy setting %s", action, settingName),
			err.Error(),
		)
		return settingRequest, err
	}
	return settingRequest, nil
}

func (r PolicySettingModel) RefreshPropertyValues(policySetting *citrixorchestration.SettingResponse) PolicySettingModel {
	r.Id = types.StringValue(policySetting.GetSettingGuid())
	r.PolicyId = types.StringValue(policySetting.GetPolicyGuid())
	r.Name = types.StringValue(policySetting.GetSettingName())
	r.UseDefault = types.BoolValue(policySetting.GetUseDefault())
	if !policySetting.GetUseDefault() {
		settingValue := types.StringValue(policySetting.GetSettingValue())
		if strings.EqualFold(policySetting.GetSettingValue(), "true") ||
			policySetting.GetSettingValue() == "1" {
			r.Enabled = types.BoolValue(true)
			r.Value = types.StringNull()
		} else if strings.EqualFold(policySetting.GetSettingValue(), "false") ||
			policySetting.GetSettingValue() == "0" {
			r.Enabled = types.BoolValue(false)
			r.Value = types.StringNull()
		} else {
			r.Enabled = types.BoolNull()
			r.Value = settingValue
		}
	}

	return r
}
