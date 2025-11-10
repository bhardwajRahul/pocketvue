package routes

import (
	"pocketvue/constants"
	"pocketvue/helpers"
	"pocketvue/types"

	"github.com/pocketbase/pocketbase/core"
)

// GetProducts returns all non-archived products
func GetProducts(e *core.RequestEvent) error {
	// Fetch all non-archived products, ordered by created date
	records, err := helpers.FindAllRecords(e.App, constants.CollectionPolarProducts)
	if err != nil {
		return helpers.JSONInternalServerError(e, "failed to fetch products")
	}

	// Filter out archived products and convert to response structs
	var activeProducts []types.ProductResponse
	for _, record := range records {
		if !record.GetBool("is_archived") {
			product := types.ProductResponse{
				ID:                     record.GetString("id"),
				Name:                   record.GetString("name"),
				PriceAmount:            record.GetInt("price_amount"),
				PriceCurrency:          record.GetString("price_currency"),
				RecurringInterval:      record.GetString("recurring_interval"),
				RecurringIntervalCount: record.GetInt("recurring_interval_count"),
				IsRecurring:            record.GetBool("is_recurring"),
				PolarPriceID:           record.GetString("polar_price_id"),
			}

			// Add optional fields if they exist
			if desc := record.GetString("description"); desc != "" {
				product.Description = desc
			}
			if trialInterval := record.GetString("trial_interval"); trialInterval != "" {
				product.TrialInterval = trialInterval
			}
			if trialIntervalCount := record.GetInt("trial_interval_count"); trialIntervalCount > 0 {
				product.TrialIntervalCount = trialIntervalCount
			}

			activeProducts = append(activeProducts, product)
		}
	}

	return helpers.JSONSuccess(e, activeProducts)
}
