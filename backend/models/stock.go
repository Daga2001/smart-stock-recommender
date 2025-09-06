package models

/*
	Models define the structure of the data used in the application,
	such as Stock and ApiResponse.
*/

import "time"

// StockRatings represents a stock rating entry.
type StockRatings struct {
	ID         int       `json:"id" db:"id" example:"1"`
	Ticker     string    `json:"ticker" db:"ticker" example:"AAPL"`
	TargetFrom string    `json:"target_from" db:"target_from" example:"$150.00"`
	TargetTo   string    `json:"target_to" db:"target_to" example:"$180.00"`
	Company    string    `json:"company" db:"company" example:"Apple Inc."`
	Action     string    `json:"action" db:"action" example:"target raised by"`
	Brokerage  string    `json:"brokerage" db:"brokerage" example:"Goldman Sachs"`
	RatingFrom string    `json:"rating_from" db:"rating_from" example:"Buy"`
	RatingTo   string    `json:"rating_to" db:"rating_to" example:"Strong Buy"`
	Time       time.Time `json:"time" db:"time" example:"2025-01-15T10:30:00Z"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" example:"2025-01-15T10:35:00Z"`
}

// ApiResponse represents the response from the external stock API.
type ApiResponse struct {
	Items    []StockRatings `json:"items"`
	NextPage string         `json:"next_page"`
}

// PageRequest represents the expected structure of the pagination request.
type PageRequest struct {
	Page int `json:"page" binding:"required" example:"1"`
}

type BulkPageRequest struct {
	StartPage int `json:"start_page" binding:"required" example:"1"`
	EndPage   int `json:"end_page" binding:"required" example:"100"`
}

type PaginationRequest struct {
	PageNumber int `json:"page_number" binding:"required" example:"1"`
	PageLength int `json:"page_length" binding:"required" example:"20"`
}
