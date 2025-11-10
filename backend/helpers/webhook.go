package helpers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"pocketvue/config"
	"strings"
	"time"

	svix "github.com/standard-webhooks/standard-webhooks/libraries/go"
)

// WebhookHeaders contains extracted webhook headers
type WebhookHeaders struct {
	ID        string
	Timestamp string
	Signature string
}

// ExtractWebhookHeaders extracts webhook headers from HTTP request headers
// Tries both canonical (Webhook-Id) and lowercase (webhook-id) forms
func ExtractWebhookHeaders(headers http.Header) WebhookHeaders {
	var wh WebhookHeaders

	// Try canonical form first, then lowercase
	wh.ID = headers.Get("Webhook-Id")
	if wh.ID == "" {
		wh.ID = headers.Get("webhook-id")
	}

	wh.Timestamp = headers.Get("Webhook-Timestamp")
	if wh.Timestamp == "" {
		wh.Timestamp = headers.Get("webhook-timestamp")
	}

	wh.Signature = headers.Get("Webhook-Signature")
	if wh.Signature == "" {
		wh.Signature = headers.Get("webhook-signature")
	}

	return wh
}

// VerifyWebhookSignature verifies the webhook signature using Standard Webhooks
func VerifyWebhookSignature(payload []byte, headers http.Header) error {
	secret := config.PolarWebhookSecret
	if secret == "" {
		return fmt.Errorf("POLAR_WEBHOOK_SECRET not configured")
	}

	// Standard Webhooks library expects secrets to be base64 encoded with whsec_ prefix
	// If the secret doesn't have the prefix, add it and encode
	if !strings.HasPrefix(secret, "whsec_") {
		// If it's not already base64, encode it
		if _, err := base64.StdEncoding.DecodeString(secret); err != nil {
			// Secret is plain text, encode it as base64
			secret = "whsec_" + base64.StdEncoding.EncodeToString([]byte(secret))
		} else {
			// Secret is already base64, just add prefix
			secret = "whsec_" + secret
		}
	}

	// Initialize webhook verifier
	wh, err := svix.NewWebhook(secret)
	if err != nil {
		return fmt.Errorf("failed to initialize webhook verifier: %w", err)
	}

	// Extract headers using helper function
	whHeaders := ExtractWebhookHeaders(headers)

	if whHeaders.ID == "" || whHeaders.Timestamp == "" || whHeaders.Signature == "" {
		return fmt.Errorf("missing required webhook headers")
	}

	// Verify the signature - pass headers directly as http.Header
	err = wh.Verify(payload, headers)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Additional timestamp validation (prevent replay attacks)
	// Parse Unix timestamp (seconds since epoch)
	var timestamp time.Time
	var timestampInt int64
	if _, err := fmt.Sscanf(whHeaders.Timestamp, "%d", &timestampInt); err == nil {
		timestamp = time.Unix(timestampInt, 0)
	} else {
		// Try parsing as RFC3339 as fallback
		timestamp, err = time.Parse(time.RFC3339, whHeaders.Timestamp)
		if err != nil {
			return fmt.Errorf("invalid timestamp format: %w", err)
		}
	}

	// Reject if timestamp is more than 5 minutes old
	if time.Since(timestamp) > 5*time.Minute {
		return fmt.Errorf("webhook timestamp too old")
	}

	return nil
}
