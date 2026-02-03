// Copyright © 2026. Citrix Systems, Inc.

package linters

import (
	"github.com/citrix/terraform-provider-citrix/custom-linters/executewithretry"
	"github.com/citrix/terraform-provider-citrix/custom-linters/panichandler"
	"github.com/citrix/terraform-provider-citrix/custom-linters/unknowncheck"
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("panichandler", NewPanicHandler)
	register.Plugin("unknowncheck", NewUnknownCheck)
	register.Plugin("executewithretry", NewExecuteWithRetry)
}

type PanicHandlerPlugin struct{}

func NewPanicHandler(settings any) (register.LinterPlugin, error) {
	return &PanicHandlerPlugin{}, nil
}

func (p *PanicHandlerPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		panichandler.Analyzer,
	}, nil
}

func (p *PanicHandlerPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

type UnknownCheckPlugin struct{}

func NewUnknownCheck(settings any) (register.LinterPlugin, error) {
	return &UnknownCheckPlugin{}, nil
}

func (p *UnknownCheckPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		unknowncheck.Analyzer,
	}, nil
}

func (p *UnknownCheckPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

type ExecuteWithRetryPlugin struct{}

func NewExecuteWithRetry(settings any) (register.LinterPlugin, error) {
	return &ExecuteWithRetryPlugin{}, nil
}

func (p *ExecuteWithRetryPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		executewithretry.Analyzer,
	}, nil
}

func (p *ExecuteWithRetryPlugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
