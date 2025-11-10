package helpers

import (
	"pocketvue/config"
)

// BuildFrontendURL constructs a frontend URL with the given path
func BuildFrontendURL(path string) string {
	baseURL := config.FrontendURL
	// Ensure baseURL doesn't end with /
	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	// Ensure path starts with /
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return baseURL + path
}

// BuildWorkspaceURL constructs a workspace-specific frontend URL
func BuildWorkspaceURL(workspaceSlug, path string) string {
	// Ensure path starts with /
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return BuildFrontendURL("/" + workspaceSlug + path)
}

// BuildCheckoutSuccessURL constructs a checkout success URL
func BuildCheckoutSuccessURL(workspaceSlug, returnPath string) string {
	if workspaceSlug != "" {
		if returnPath == "" {
			returnPath = "/dashboard"
		}
		return BuildWorkspaceURL(workspaceSlug, returnPath+"?checkout=success")
	}
	return BuildFrontendURL("/checkout/success")
}

// BuildCheckoutReturnURL constructs a checkout return URL
func BuildCheckoutReturnURL(workspaceSlug, returnPath string) string {
	if workspaceSlug != "" {
		if returnPath == "" {
			returnPath = "/dashboard"
		}
		return BuildWorkspaceURL(workspaceSlug, returnPath)
	}
	return BuildFrontendURL("/dashboard")
}

// BuildCustomerPortalReturnURL constructs a customer portal return URL
func BuildCustomerPortalReturnURL(workspaceSlug, returnPath string) string {
	if workspaceSlug != "" {
		if returnPath == "" {
			returnPath = "/dashboard/settings/billing"
		}
		return BuildWorkspaceURL(workspaceSlug, returnPath)
	}
	return BuildFrontendURL("/dashboard/settings/billing")
}


