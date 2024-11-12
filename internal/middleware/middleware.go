package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	citrixclient "github.com/citrix/citrix-daas-rest-go/client"
)

func MiddlewareAuthWithCustomerIdHeaderFunc(authClient *citrixclient.CitrixDaasClient, r *http.Request) {
	// Auth
	if authClient != nil && r.Header.Get("Authorization") == "" {
		token, _, err := authClient.SignIn()
		if err != nil {
			tflog.Error(r.Context(), "Could not sign into Citrix DaaS, error: "+err.Error())
		}
		r.Header["Authorization"] = []string{token}
	}

	r.Header["Citrix-CustomerId"] = []string{authClient.ClientConfig.CustomerId}
	r.Header["Accept"] = []string{"application/json"}
	r.URL.Path = strings.Replace(r.URL.Path, "/"+authClient.ClientConfig.CustomerId+"/administrators", "", 1)

	// TransactionId
	transactionId := r.Header.Get("Citrix-TransactionId")
	if transactionId == "" {
		transactionId = uuid.NewString()
		r.Header.Add("Citrix-TransactionId", transactionId)
	}

	// Log the request
	tflog.Info(r.Context(), "Orchestration API request", map[string]interface{}{
		"url":           r.URL.String(),
		"method":        r.Method,
		"transactionId": transactionId,
	})
}

func MiddlewareAuthFunc(authClient *citrixclient.CitrixDaasClient, r *http.Request) {
	// Auth
	if authClient != nil && r.Header.Get("Authorization") == "" {
		token, _, err := authClient.SignIn()
		if err != nil {
			tflog.Error(r.Context(), "Could not sign into Citrix DaaS, error: "+err.Error())
		}
		r.Header["Authorization"] = []string{token}
	}

	// TransactionId
	transactionId := r.Header.Get("Citrix-TransactionId")
	if transactionId == "" {
		transactionId = uuid.NewString()
		r.Header.Add("Citrix-TransactionId", transactionId)
	}

	// Log the request
	tflog.Info(r.Context(), "Orchestration API request", map[string]interface{}{
		"url":           r.URL.String(),
		"method":        r.Method,
		"transactionId": transactionId,
	})
}

// Middleware Auth for Wem OnPrem Client with SessionId
func MiddlewareAuthWithSessionIdFunc(authClient *citrixclient.CitrixDaasClient, r *http.Request) {
	if authClient != nil && r.Header.Get("Authorization") == "" {
		token, _, err := authClient.SignInWemOnPrem()
		if err != nil {
			tflog.Error(r.Context(), "Could not sign into Citrix Wem On-Premise Host, error: "+err.Error())
		}
		r.Header["Authorization"] = []string{token}
	}

	// TransactionId
	transactionId := r.Header.Get("Citrix-TransactionId")
	if transactionId == "" {
		transactionId = uuid.NewString()
		r.Header.Add("Citrix-TransactionId", transactionId)
	}

	// Log the request
	tflog.Info(r.Context(), "WEM On-Prem Host API request", map[string]interface{}{
		"url":           r.URL.String(),
		"method":        r.Method,
		"transactionId": transactionId,
	})
}
