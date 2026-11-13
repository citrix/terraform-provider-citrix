// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/citrix/citrix-daas-rest-go/citrixorchestration"
	citrixdaasclient "github.com/citrix/citrix-daas-rest-go/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Inventory folders and custom properties of the Citrix-built Nutanix AHV Prism Central plugin.
const (
	NutanixPCClustersFolder  = "Clusters.folder"
	NutanixPCTemplatesFolder = "Templates.folder"
	NutanixPCClusterIdKey    = "ClusterId"
	NutanixPCCpuCoresKey     = "CPUCores"
)

// ParseNutanixPCTemplateVersionPath splits ...\<template>.template\<version>.templateversion into its template and version names.
func ParseNutanixPCTemplateVersionPath(xdPath string) (string, string, bool) {
	segments := strings.Split(xdPath, "\\")
	if len(segments) < 2 {
		return "", "", false
	}
	version, isVersion := strings.CutSuffix(segments[len(segments)-1], ".templateversion")
	template, isTemplate := strings.CutSuffix(segments[len(segments)-2], ".template")
	if !isVersion || !isTemplate {
		return "", "", false
	}
	return template, version, true
}

func GetNutanixPCCoresPerCpu(customProperties []citrixorchestration.NameValueStringPairModel) (int64, bool) {
	for _, property := range customProperties {
		if property.GetName() == NutanixPCCpuCoresKey {
			cores, err := strconv.ParseInt(property.GetValue(), 10, 64)
			return cores, err == nil
		}
	}
	return 0, false
}

func GetNutanixPCClusterId(customProperties []citrixorchestration.NameValueStringPairModel) string {
	for _, property := range customProperties {
		if property.GetName() == NutanixPCClusterIdKey {
			return property.GetValue()
		}
	}
	return ""
}

// NutanixPCCustomProperties returns the custom properties the Prism Central plugin reads for provisioning.
func NutanixPCCustomProperties(clusterId string, coresPerCpu int64) []citrixorchestration.NameValueStringPairModel {
	customProperties := []citrixorchestration.NameValueStringPairModel{}
	AppendNameValueStringPair(&customProperties, NutanixPCClusterIdKey, clusterId)
	AppendNameValueStringPair(&customProperties, NutanixPCCpuCoresKey, strconv.FormatInt(coresPerCpu, 10))
	return customProperties
}

func GetNutanixPCCluster(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, cluster string) (*citrixorchestration.HypervisorResourceResponseModel, *http.Response, error) {
	return GetSingleResourceFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorName, poolName, NutanixPCClustersFolder, cluster, ClusterResourceType, "")
}

// GetNutanixPCClusterName resolves a ClusterId custom property value to the cluster name, bypassing the inventory cache
// when the cached listing doesn't have the cluster yet.
func GetNutanixPCClusterName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, clusterId string) (string, *http.Response, error) {
	name, httpResp, err := findNutanixPCClusterName(ctx, client, diagnostics, hypervisorName, poolName, clusterId, false)
	if err != nil {
		return findNutanixPCClusterName(ctx, client, diagnostics, hypervisorName, poolName, clusterId, true)
	}
	return name, httpResp, nil
}

func findNutanixPCClusterName(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, clusterId string, noCache bool) (string, *http.Response, error) {
	clusters, httpResp, err := GetAllChildrenForResourcePath(ctx, client, diagnostics, hypervisorName, poolName, NutanixPCClustersFolder, ClusterResourceType, noCache, noCache)
	if err != nil {
		return "", httpResp, err
	}
	for _, cluster := range clusters {
		if strings.EqualFold(cluster.GetId(), clusterId) {
			return cluster.GetName(), httpResp, nil
		}
	}
	return "", httpResp, fmt.Errorf("no Nutanix Prism Central cluster with ID %s", clusterId)
}

func GetNutanixPCTemplateVersionPath(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, template, version string) (string, *http.Response, error) {
	folderPath := fmt.Sprintf("%s\\%s.template", NutanixPCTemplatesFolder, template)
	return GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorName, poolName, folderPath, version, "", "")
}

// GetNutanixPCNetworkMapping resolves networks under the selected cluster, because a Prism Central resource pool has no networks of its own.
func GetNutanixPCNetworkMapping(ctx context.Context, client *citrixdaasclient.CitrixDaasClient, diagnostics *diag.Diagnostics, hypervisorName, poolName, cluster string, networkMappings []NetworkMappingModel) ([]citrixorchestration.NetworkMapRequestModel, error) {
	folderPath := fmt.Sprintf("%s\\%s.cluster", NutanixPCClustersFolder, cluster)
	networkMapRequests := []citrixorchestration.NetworkMapRequestModel{}
	for _, networkMapping := range networkMappings {
		networkPath, httpResp, err := GetSingleResourcePathFromHypervisorWithNoCacheRetry(ctx, client, diagnostics, hypervisorName, poolName, folderPath, networkMapping.Network.ValueString(), NetworkResourceType, "")
		if err != nil {
			return nil, fmt.Errorf("network %s not found in cluster %s. TransactionId: %s: %w", networkMapping.Network.ValueString(), cluster, citrixdaasclient.GetTransactionIdFromHttpResponse(httpResp), err)
		}
		networkMapRequest := citrixorchestration.NetworkMapRequestModel{}
		networkMapRequest.SetNetworkDeviceNameOrId(networkMapping.NetworkDevice.ValueString())
		networkMapRequest.SetNetworkPath(networkPath)
		networkMapRequests = append(networkMapRequests, networkMapRequest)
	}
	return networkMapRequests, nil
}
