package services

import (
	"encoding/json"
	"fmt"
	"log"
	"pocketvue/constants"
	"pocketvue/types"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

// WebhookService handles Polar webhook events
type WebhookService struct {
	app *pocketbase.PocketBase
}

// NewWebhookService creates a new webhook service instance
func NewWebhookService(app *pocketbase.PocketBase) *WebhookService {
	return &WebhookService{
		app: app,
	}
}

// findUserByExternalID finds a user by their external customer ID
func (ws *WebhookService) findUserByExternalID(externalID string) (*core.Record, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external_id is empty")
	}

	records, err := ws.app.FindRecordsByFilter(
		constants.CollectionUsers,
		fmt.Sprintf("id = '%s'", externalID),
		"-created",
		1,
		0,
	)

	if err != nil || len(records) == 0 {
		return nil, fmt.Errorf("user not found with external_id: %s", externalID)
	}

	return records[0], nil
}

// updateUserSubscription updates user subscription fields based on subscription data
// statusOverride allows overriding the status (e.g., "active" for subscription.active events)
func (ws *WebhookService) updateUserSubscription(subData types.SubscriptionWebhookData, statusOverride string) (*core.Record, error) {
	if subData.Customer.ExternalID == nil {
		return nil, fmt.Errorf("subscription event has no external_id, subscription_id=%s", subData.ID)
	}

	user, err := ws.findUserByExternalID(*subData.Customer.ExternalID)
	if err != nil {
		return nil, err
	}

	// Determine status - use override if provided, otherwise use data status
	status := subData.Status
	if statusOverride != "" {
		status = statusOverride
	}

	// Update subscription fields
	user.Set("subscription_id", subData.ID)
	user.Set("subscription_status", status)
	user.Set("subscription_product_id", subData.ProductID)
	user.Set("subscription_current_period_end", subData.CurrentPeriodEnd)
	user.Set("subscription_cancel_at_period_end", subData.CancelAtPeriodEnd)

	if err := ws.app.Save(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// HandleSubscriptionCreated handles subscription.created events
func (ws *WebhookService) HandleSubscriptionCreated(data []byte) error {
	var subData types.SubscriptionWebhookData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	user, err := ws.updateUserSubscription(subData, "")
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil // Return nil to prevent retries
	}

	log.Printf("Subscription created for user %s: subscription_id=%s, status=%s",
		user.Id, subData.ID, subData.Status)

	return nil
}

// HandleSubscriptionUpdated handles subscription.updated events
func (ws *WebhookService) HandleSubscriptionUpdated(data []byte) error {
	var subData types.SubscriptionWebhookData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	user, err := ws.updateUserSubscription(subData, "")
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	log.Printf("Subscription updated for user %s: subscription_id=%s, status=%s",
		user.Id, subData.ID, subData.Status)

	return nil
}

// HandleSubscriptionActive handles subscription.active events
func (ws *WebhookService) HandleSubscriptionActive(data []byte) error {
	var subData types.SubscriptionWebhookData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	user, err := ws.updateUserSubscription(subData, "active")
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	log.Printf("Subscription activated for user %s: subscription_id=%s", user.Id, subData.ID)

	return nil
}

// HandleSubscriptionCanceled handles subscription.canceled events
func (ws *WebhookService) HandleSubscriptionCanceled(data []byte) error {
	var subData types.SubscriptionWebhookData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	if subData.Customer.ExternalID == nil {
		log.Printf("Warning: subscription.canceled event has no external_id, subscription_id=%s", subData.ID)
		return nil
	}

	user, err := ws.findUserByExternalID(*subData.Customer.ExternalID)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	// Only update cancellation-specific fields
	user.Set("subscription_status", "canceled")
	user.Set("subscription_cancel_at_period_end", subData.CancelAtPeriodEnd)

	if err := ws.app.Save(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("Subscription canceled for user %s: subscription_id=%s, cancel_at_period_end=%v",
		user.Id, subData.ID, subData.CancelAtPeriodEnd)

	return nil
}

// HandleSubscriptionRevoked handles subscription.revoked events
func (ws *WebhookService) HandleSubscriptionRevoked(data []byte) error {
	var subData types.SubscriptionWebhookData
	if err := json.Unmarshal(data, &subData); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	if subData.Customer.ExternalID == nil {
		log.Printf("Warning: subscription.revoked event has no external_id, subscription_id=%s", subData.ID)
		return nil
	}

	user, err := ws.findUserByExternalID(*subData.Customer.ExternalID)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	// Revocation is immediate - set status and clear cancel flag
	user.Set("subscription_status", "revoked")
	user.Set("subscription_cancel_at_period_end", false)

	if err := ws.app.Save(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("Subscription revoked for user %s: subscription_id=%s (immediate access loss)",
		user.Id, subData.ID)

	return nil
}

// HandleOrderCreated handles order.created events
func (ws *WebhookService) HandleOrderCreated(data []byte) error {
	var orderData types.OrderWebhookData
	if err := json.Unmarshal(data, &orderData); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	if orderData.Customer.ExternalID == nil {
		log.Printf("Warning: order.created event has no external_id, order_id=%s", orderData.ID)
		return nil
	}

	log.Printf("Order created: order_id=%s, user_id=%s, status=%s, billing_reason=%s",
		orderData.ID, *orderData.Customer.ExternalID, orderData.Status, orderData.BillingReason)

	return nil
}

// HandleOrderPaid handles order.paid events
func (ws *WebhookService) HandleOrderPaid(data []byte) error {
	var orderData types.OrderWebhookData
	if err := json.Unmarshal(data, &orderData); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	if orderData.Customer.ExternalID == nil {
		log.Printf("Warning: order.paid event has no external_id, order_id=%s", orderData.ID)
		return nil
	}

	user, err := ws.findUserByExternalID(*orderData.Customer.ExternalID)
	if err != nil {
		log.Printf("Warning: %v", err)
		return nil
	}

	user.Set("last_payment_status", "paid")

	// If this is the first payment for a subscription, ensure subscription is marked as active
	if orderData.BillingReason == "subscription_create" && orderData.SubscriptionID != nil {
		user.Set("subscription_status", "active")
		user.Set("subscription_id", *orderData.SubscriptionID)
		user.Set("subscription_product_id", orderData.ProductID)
	}

	if err := ws.app.Save(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("Order paid for user %s: order_id=%s, amount=%d %s, billing_reason=%s",
		user.Id, orderData.ID, orderData.TotalAmount, orderData.Currency, orderData.BillingReason)

	return nil
}

// setProductRecordFields sets product record fields from product webhook data
func (ws *WebhookService) setProductRecordFields(record *core.Record, productData types.ProductWebhookData) {
	// Get the first price (assuming one price per product)
	var priceAmount int
	var priceCurrency string
	var priceID string

	if len(productData.Prices) > 0 {
		price := productData.Prices[0]
		priceAmount = price.PriceAmount
		priceCurrency = price.PriceCurrency
		priceID = price.ID
	}

	// Set product fields
	record.Set("id", productData.ID)
	record.Set("name", productData.Name)
	if productData.Description != nil {
		record.Set("description", *productData.Description)
	}
	record.Set("price_amount", priceAmount)
	record.Set("price_currency", priceCurrency)
	record.Set("recurring_interval", productData.RecurringInterval)
	record.Set("recurring_interval_count", productData.RecurringIntervalCount)
	record.Set("is_recurring", productData.IsRecurring)
	record.Set("is_archived", productData.IsArchived)

	if productData.TrialInterval != nil {
		record.Set("trial_interval", *productData.TrialInterval)
	}
	if productData.TrialIntervalCount != nil {
		record.Set("trial_interval_count", *productData.TrialIntervalCount)
	}
	record.Set("polar_price_id", priceID)
}

// HandleProductCreated handles product.created events
func (ws *WebhookService) HandleProductCreated(data []byte) error {
	var productData types.ProductWebhookData
	if err := json.Unmarshal(data, &productData); err != nil {
		return fmt.Errorf("failed to parse product data: %w", err)
	}

	// Create product record
	collection, err := ws.app.FindCollectionByNameOrId(constants.CollectionPolarProducts)
	if err != nil {
		return fmt.Errorf("failed to find polar_products collection: %w", err)
	}

	record := core.NewRecord(collection)
	ws.setProductRecordFields(record, productData)

	if err := ws.app.Save(record); err != nil {
		return fmt.Errorf("failed to create product record: %w", err)
	}

	priceAmount := 0
	priceCurrency := ""
	if len(productData.Prices) > 0 {
		priceAmount = productData.Prices[0].PriceAmount
		priceCurrency = productData.Prices[0].PriceCurrency
	}

	log.Printf("Product created: product_id=%s, name=%s, price=%d %s",
		productData.ID, productData.Name, priceAmount, priceCurrency)

	return nil
}

// HandleProductUpdated handles product.updated events
func (ws *WebhookService) HandleProductUpdated(data []byte) error {
	var productData types.ProductWebhookData
	if err := json.Unmarshal(data, &productData); err != nil {
		return fmt.Errorf("failed to parse product data: %w", err)
	}

	// Find existing product record
	record, err := ws.app.FindRecordById(constants.CollectionPolarProducts, productData.ID)
	if err != nil {
		log.Printf("Product not found in database, creating: product_id=%s", productData.ID)
		// If product doesn't exist, create it
		return ws.HandleProductCreated(data)
	}

	// Update product record using shared helper
	ws.setProductRecordFields(record, productData)

	if err := ws.app.Save(record); err != nil {
		return fmt.Errorf("failed to update product record: %w", err)
	}

	priceAmount := 0
	priceCurrency := ""
	if len(productData.Prices) > 0 {
		priceAmount = productData.Prices[0].PriceAmount
		priceCurrency = productData.Prices[0].PriceCurrency
	}

	log.Printf("Product updated: product_id=%s, name=%s, price=%d %s, archived=%v",
		productData.ID, productData.Name, priceAmount, priceCurrency, productData.IsArchived)

	return nil
}
