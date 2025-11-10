package routes

import (
	"encoding/json"
	"log"
	"pocketvue/helpers"
	"pocketvue/services"

	"github.com/pocketbase/pocketbase/core"
)

// CreateCheckoutRequest represents the request body for creating a checkout session
type CreateCheckoutRequest struct {
	Products      []string `json:"products"`
	WorkspaceSlug string   `json:"workspace_slug"` // Optional: for workspace-specific success URL
	ReturnPath    string   `json:"return_path"`    // Optional: custom return path (defaults to /dashboard)
}

// CreateCheckoutResponse represents the response for a successful checkout creation
type CreateCheckoutResponse struct {
	URL string `json:"url"`
}

// CreateCustomerPortalRequest represents the request body for creating a customer portal session
type CreateCustomerPortalRequest struct {
	WorkspaceSlug string `json:"workspace_slug"` // Optional: for workspace-specific return URL
	ReturnPath    string `json:"return_path"`    // Optional: custom return path (defaults to /dashboard/settings/billing)
}

// CreateCustomerPortalResponse represents the response for a successful customer portal creation
type CreateCustomerPortalResponse struct {
	URL string `json:"url"`
}

// CreateCheckoutSession creates a Polar checkout session for the authenticated user
func CreateCheckoutSession(e *core.RequestEvent) error {
	// Get authenticated user
	user, err := helpers.GetAuthenticatedUser(e)
	if err != nil {
		return err
	}

	// Parse request body
	var req CreateCheckoutRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		log.Printf("Error parsing checkout request: %v", err)
		return helpers.JSONBadRequest(e, "invalid request body")
	}

	// Validate required fields
	if len(req.Products) == 0 {
		return helpers.JSONBadRequest(e, "products field is required and must contain at least one product ID")
	}

	// Build URLs using helper functions
	successURL := helpers.BuildCheckoutSuccessURL(req.WorkspaceSlug, req.ReturnPath)
	returnURL := helpers.BuildCheckoutReturnURL(req.WorkspaceSlug, req.ReturnPath)

	// Extract user information
	userID := user.GetString("id")
	userEmail := user.GetString("email")
	userName := user.GetString("name")

	log.Printf("CreateCheckoutSession called by user: ID=%s, Email=%s, Products=%v",
		userID, userEmail, req.Products)

	// Create Polar service and checkout session
	polarService := services.NewPolarService()
	checkoutURL, err := polarService.CreateCheckoutSession(
		req.Products,
		successURL,
		returnURL,
		userID,
		userEmail,
		userName,
	)

	if err != nil {
		log.Printf("Error creating checkout session for user %s: %v", userID, err)
		return helpers.JSONInternalServerError(e, "failed to create checkout session")
	}

	// Return checkout URL
	return helpers.JSONSuccess(e, CreateCheckoutResponse{
		URL: checkoutURL,
	})
}

// CreateCustomerPortalSession creates a Polar customer portal session for the authenticated user
func CreateCustomerPortalSession(e *core.RequestEvent) error {
	// Get authenticated user
	user, err := helpers.GetAuthenticatedUser(e)
	if err != nil {
		return err
	}

	// Parse request body
	var req CreateCustomerPortalRequest
	if err := json.NewDecoder(e.Request.Body).Decode(&req); err != nil {
		log.Printf("Error parsing customer portal request: %v", err)
		return helpers.JSONBadRequest(e, "invalid request body")
	}

	// Build return URL using helper function
	returnURL := helpers.BuildCustomerPortalReturnURL(req.WorkspaceSlug, req.ReturnPath)

	// Extract user ID
	userID := user.GetString("id")

	log.Printf("CreateCustomerPortalSession called by user: ID=%s", userID)

	// Create Polar service and customer portal session
	polarService := services.NewPolarService()
	portalURL, err := polarService.CreateCustomerSession(userID, returnURL)

	if err != nil {
		log.Printf("Error creating customer portal session for user %s: %v", userID, err)
		return helpers.JSONInternalServerError(e, "failed to create customer portal session")
	}

	// Return portal URL
	return helpers.JSONSuccess(e, CreateCustomerPortalResponse{
		URL: portalURL,
	})
}
