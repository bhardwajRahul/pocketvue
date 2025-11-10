package types

// ProductResponse represents a product in the API response
type ProductResponse struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description,omitempty"`
	PriceAmount              int    `json:"price_amount"`
	PriceCurrency            string `json:"price_currency"`
	RecurringInterval        string `json:"recurring_interval"`
	RecurringIntervalCount   int    `json:"recurring_interval_count"`
	IsRecurring              bool   `json:"is_recurring"`
	TrialInterval            string `json:"trial_interval,omitempty"`
	TrialIntervalCount       int    `json:"trial_interval_count,omitempty"`
	PolarPriceID             string `json:"polar_price_id"`
}


