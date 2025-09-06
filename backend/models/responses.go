package models

import "time"

/*
	Models for API responses, only used for documentation purposes.
	These structs are not used in the actual code logic.
*/

// StockResponse represents a single stock rating response
type StockResponse struct {
	Items    []StockRatings `json:"items" example:"[{\"id\":1,\"ticker\":\"AAPL\",\"target_from\":\"$150.00\",\"target_to\":\"$180.00\",\"company\":\"Apple Inc.\",\"action\":\"target raised by\",\"brokerage\":\"Goldman Sachs\",\"rating_from\":\"Buy\",\"rating_to\":\"Strong Buy\",\"time\":\"2025-01-15T10:30:00Z\",\"created_at\":\"2025-01-15T10:35:00Z\"}]"`
	NextPage string         `json:"next_page" example:"AAPL"`
}

// BulkResponse represents bulk operation response
type BulkResponse struct {
	Message      string         `json:"message" example:"Successfully fetched and stored stock data"`
	PagesFetched string         `json:"pages_fetched" example:"1-1000"`
	Stocks       []StockRatings `json:"stocks"`
	TotalStocks  int            `json:"total_stocks" example:"7860"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	PageNumber   int  `json:"page_number" example:"1"`
	PageLength   int  `json:"page_length" example:"20"`
	TotalRecords int  `json:"total_records" example:"2520"`
	TotalPages   int  `json:"total_pages" example:"126"`
	HasNext      bool `json:"has_next" example:"true"`
	HasPrevious  bool `json:"has_previous" example:"false"`
}

// PaginatedResponse represents paginated stock ratings response
type PaginatedResponse struct {
	Data       []StockRatings `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

// TargetChanges represents target price change metrics
type TargetChanges struct {
	Raised     int `json:"raised" example:"1200"`
	Lowered    int `json:"lowered" example:"800"`
	Maintained int `json:"maintained" example:"520"`
}

// MarketSentiment represents market sentiment analysis
type MarketSentiment struct {
	BullishCount      int     `json:"bullish_count" example:"1400"`
	BearishCount      int     `json:"bearish_count" example:"600"`
	NeutralCount      int     `json:"neutral_count" example:"520"`
	BullishPercentage float64 `json:"bullish_percentage" example:"55.6"`
	BearishPercentage float64 `json:"bearish_percentage" example:"23.8"`
	NeutralPercentage float64 `json:"neutral_percentage" example:"20.6"`
}

// BrokerageActivity represents brokerage activity data
type BrokerageActivity struct {
	Name     string `json:"name" example:"Goldman Sachs"`
	Activity int    `json:"activity" example:"150"`
}

// ActiveStock represents most active stock data
type ActiveStock struct {
	Ticker      string `json:"ticker" example:"AAPL"`
	Company     string `json:"company" example:"Apple Inc."`
	RatingCount int    `json:"rating_count" example:"25"`
}

// MetricsData represents all metrics data
type MetricsData struct {
	TotalRecords        int                          `json:"total_records" example:"2520"`
	TargetChanges       TargetChanges                `json:"target_changes"`
	MarketSentiment     MarketSentiment              `json:"market_sentiment"`
	RatingDistribution  map[string]int               `json:"rating_distribution"`
	TopBrokerages       []BrokerageActivity          `json:"top_brokerages"`
	MostActiveStocks    []ActiveStock                `json:"most_active_stocks"`
	RecentActivity      int                          `json:"recent_activity" example:"125"`
	GeneratedAt         time.Time                    `json:"generated_at" example:"2025-01-15T10:30:00Z"`
	Description         string                       `json:"description" example:"Comprehensive stock market analytics based on analyst ratings and target price changes"`
}

// MetricsResponse represents metrics endpoint response
type MetricsResponse struct {
	Success bool        `json:"success" example:"true"`
	Metrics MetricsData `json:"metrics"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid JSON format in request body"`
}

// GenericErrorResponse represents generic server error response
type GenericErrorResponse struct {
	Error string `json:"error" example:"Internal server error occurred"`
}