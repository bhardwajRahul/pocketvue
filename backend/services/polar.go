package services

import (
	"context"
	"log"
	"pocketvue/config"

	"github.com/pocketbase/pocketbase"
	polargo "github.com/polarsource/polar-go"
	"github.com/polarsource/polar-go/models/components"
	"github.com/polarsource/polar-go/models/operations"
)

// PolarService handles Polar.sh payment gateway operations
type PolarService struct {
	client *polargo.Polar
}

// NewPolarService creates a new Polar service instance
func NewPolarService() *PolarService {
	client := polargo.New(
		polargo.WithServer(config.GetPolarServer()),
		polargo.WithSecurity(config.PolarAccessToken),
	)

	return &PolarService{
		client: client,
	}
}

// CreateCustomerAsync creates a Polar customer asynchronously for a new user
func (ps *PolarService) CreateCustomerAsync(app *pocketbase.PocketBase, userID, userEmail, userName string) {
	go ps.createCustomer(app, userID, userEmail, userName)
}

// createCustomer handles the actual Polar customer creation (private method)
func (ps *PolarService) createCustomer(app *pocketbase.PocketBase, userID, userEmail, userName string) {
	ctx := context.Background()

	if userEmail == "" {
		log.Printf("Warning: User %s has no email, skipping Polar customer creation", userID)
		return
	}

	// Create customer in Polar
	// Note: When using organization token, organization_id should not be set
	customerReq := components.CustomerCreate{
		ExternalID: polargo.Pointer(userID),
		Email:      userEmail,
		Name:       polargo.Pointer(userName),
	}

	res, err := ps.client.Customers.Create(ctx, customerReq)
	if err != nil {
		log.Printf("Error creating Polar customer for user %s: %v", userID, err)
		return
	}

	if res.Customer != nil {
		log.Printf("Successfully created Polar customer %s for user %s", res.Customer.ID, userID)

		// Update user record with Polar customer information
		userRecord, err := app.FindRecordById("users", userID)
		if err != nil {
			log.Printf("Error finding user %s to update Polar info: %v", userID, err)
			return
		}

		userRecord.Set("polar_customer_id", res.Customer.ID)
		userRecord.Set("polar_customer_created", res.Customer.CreatedAt)

		// Save the updated user record
		err = app.Save(userRecord)
		if err != nil {
			log.Printf("Error updating user %s with Polar customer info: %v", userID, err)
		} else {
			log.Printf("Updated user %s with Polar customer ID: %s", userID, res.Customer.ID)
		}
	}
}

// CreateCheckoutSession creates a Polar checkout session and returns the checkout URL
func (ps *PolarService) CreateCheckoutSession(productIDs []string, successURL, returnURL, userID, userEmail, userName string) (string, error) {
	ctx := context.Background()

	// Validate required parameters
	if len(productIDs) == 0 {
		return "", &CheckoutError{Message: "at least one product ID is required"}
	}
	if successURL == "" {
		return "", &CheckoutError{Message: "success_url is required"}
	}
	if userEmail == "" {
		return "", &CheckoutError{Message: "user email is required"}
	}

	// Build checkout request
	checkoutReq := components.CheckoutCreate{
		Products:           productIDs,
		ExternalCustomerID: polargo.Pointer(userID),
		CustomerEmail:      polargo.Pointer(userEmail),
		CustomerName:       polargo.Pointer(userName),
		SuccessURL:         polargo.Pointer(successURL),
	}

	// Add optional return URL if provided
	if returnURL != "" {
		checkoutReq.ReturnURL = polargo.Pointer(returnURL)
	}

	// Create checkout session
	res, err := ps.client.Checkouts.Create(ctx, checkoutReq)
	if err != nil {
		log.Printf("Error creating Polar checkout for user %s: %v", userID, err)
		return "", &CheckoutError{Message: "failed to create checkout session", Err: err}
	}

	if res.Checkout == nil {
		return "", &CheckoutError{Message: "checkout response is empty"}
	}

	log.Printf("Successfully created checkout session %s for user %s", res.Checkout.ID, userID)
	return res.Checkout.URL, nil
}

// CreateCustomerSession creates a Polar customer session for accessing the customer portal
func (ps *PolarService) CreateCustomerSession(userID, returnURL string) (string, error) {
	ctx := context.Background()

	// Validate required parameters
	if userID == "" {
		return "", &CheckoutError{Message: "user ID is required"}
	}

	// Create customer session request
	sessionReq := components.CustomerSessionCustomerExternalIDCreate{
		ExternalCustomerID: userID,
	}

	// Add optional return URL if provided
	if returnURL != "" {
		sessionReq.ReturnURL = polargo.Pointer(returnURL)
	}

	// Create customer session using the SDK's wrapper function
	res, err := ps.client.CustomerSessions.Create(
		ctx,
		operations.CreateCustomerSessionsCreateCustomerSessionCreateCustomerSessionCustomerExternalIDCreate(sessionReq),
	)
	if err != nil {
		log.Printf("Error creating customer session for user %s: %v", userID, err)
		return "", &CheckoutError{Message: "failed to create customer session", Err: err}
	}

	if res.CustomerSession == nil {
		return "", &CheckoutError{Message: "customer session response is empty"}
	}

	log.Printf("Successfully created customer session for user %s", userID)
	return res.CustomerSession.CustomerPortalURL, nil
}

// CheckoutError represents an error during checkout creation
type CheckoutError struct {
	Message string
	Err     error
}

func (e *CheckoutError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}
