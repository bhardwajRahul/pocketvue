package routes

import (
	"encoding/json"
	"io"
	"log"
	"pocketvue/helpers"
	"pocketvue/services"
	"pocketvue/types"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// HandlePolarWebhook handles incoming Polar webhook events
func HandlePolarWebhook(e *core.RequestEvent) error {
	// Read the raw request body (needed for signature verification)
	body, err := io.ReadAll(e.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook body: %v", err)
		return helpers.JSONBadRequest(e, "failed to read request body")
	}

	// Verify webhook signature - pass the entire request headers
	// Header extraction is handled internally by VerifyWebhookSignature
	if err := helpers.VerifyWebhookSignature(body, e.Request.Header); err != nil {
		log.Printf("Webhook signature verification failed: %v", err)
		return helpers.JSONUnauthorized(e, "invalid signature")
	}

	// Parse the webhook event
	var event types.WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Error parsing webhook event: %v", err)
		return helpers.JSONBadRequest(e, "invalid JSON payload")
	}

	log.Printf("Received webhook event: type=%s, timestamp=%s", event.Type, event.Timestamp)

	// Re-marshal the data for individual handlers
	eventData, err := json.Marshal(event.Data)
	if err != nil {
		log.Printf("Error marshaling event data: %v", err)
		return helpers.JSONInternalServerError(e, "failed to process event data")
	}

	// Create webhook service
	app, ok := e.App.(*pocketbase.PocketBase)
	if !ok {
		log.Printf("Failed to cast app to PocketBase")
		return helpers.JSONInternalServerError(e, "internal server error")
	}
	webhookService := services.NewWebhookService(app)

	// Route to appropriate handler based on event type
	var handlerErr error
	switch event.Type {
	case "subscription.created":
		handlerErr = webhookService.HandleSubscriptionCreated(eventData)

	case "subscription.updated":
		handlerErr = webhookService.HandleSubscriptionUpdated(eventData)

	case "subscription.active":
		handlerErr = webhookService.HandleSubscriptionActive(eventData)

	case "subscription.canceled":
		handlerErr = webhookService.HandleSubscriptionCanceled(eventData)

	case "subscription.revoked":
		handlerErr = webhookService.HandleSubscriptionRevoked(eventData)

	case "order.created":
		handlerErr = webhookService.HandleOrderCreated(eventData)

	case "order.paid":
		handlerErr = webhookService.HandleOrderPaid(eventData)

	case "product.created":
		handlerErr = webhookService.HandleProductCreated(eventData)

	case "product.updated":
		handlerErr = webhookService.HandleProductUpdated(eventData)

	default:
		log.Printf("Unhandled webhook event type: %s", event.Type)
		// Return 200 OK for unhandled events to prevent retries
		return helpers.JSONSuccess(e, map[string]string{
			"message": "event type not handled",
		})
	}

	// Check if handler returned an error
	if handlerErr != nil {
		log.Printf("Error handling webhook event %s: %v", event.Type, handlerErr)
		// Return 500 to trigger Polar's retry mechanism
		return helpers.JSONInternalServerError(e, "failed to process webhook")
	}

	// Return success response
	return helpers.JSONSuccess(e, map[string]string{
		"message": "webhook processed successfully",
	})
}
