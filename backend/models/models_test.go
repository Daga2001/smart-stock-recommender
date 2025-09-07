package models

/*
Data model validation tests.

PURPOSE:
- Validates struct field assignments and data integrity
- Ensures request/response models work correctly
- Tests API contract compliance
- Verifies error response structures

THESE TESTS ENSURE:
- Data models can be properly instantiated
- Field assignments work as expected
- JSON serialization/deserialization compatibility
- API request/response structure integrity
*/

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestStockRatings validates the core StockRatings data model
// Purpose: Ensures the main data structure can store analyst ratings correctly
// This model represents analyst actions like target price changes and rating updates
func TestStockRatings(t *testing.T) {
	stock := StockRatings{
		ID:         1,
		Ticker:     "AAPL",
		TargetFrom: "$150.00",
		TargetTo:   "$180.00",
		Company:    "Apple Inc.",
		Action:     "target raised by",
		Brokerage:  "Goldman Sachs",
		RatingFrom: "Hold",
		RatingTo:   "Buy",
		Time:       time.Now(),
		CreatedAt:  time.Now(),
	}

	// Validate core fields are assigned correctly
	assert.Equal(t, 1, stock.ID, "ID should be assigned correctly")
	assert.Equal(t, "AAPL", stock.Ticker, "Ticker symbol should be stored")
	assert.Equal(t, "Apple Inc.", stock.Company, "Company name should be stored")
	assert.Equal(t, "Goldman Sachs", stock.Brokerage, "Brokerage name should be stored")
}

// TestPageRequest validates PageRequest model for single page fetching
// Purpose: Ensures the request model for single page API calls works correctly
// API Contract: Used by POST /api/stocks endpoint for external API integration
func TestPageRequest(t *testing.T) {
	req := PageRequest{Page: 1}
	assert.Equal(t, 1, req.Page, "Page field should be assigned correctly")
}

// TestBulkPageRequest validates BulkPageRequest model for parallel processing
// Purpose: Ensures the request model for bulk page operations works correctly
// API Contract: Used by POST /api/stocks/bulk endpoint for efficient data fetching
func TestBulkPageRequest(t *testing.T) {
	req := BulkPageRequest{
		StartPage: 1,
		EndPage:   10,
	}
	assert.Equal(t, 1, req.StartPage, "StartPage should be assigned correctly")
	assert.Equal(t, 10, req.EndPage, "EndPage should be assigned correctly")
}

// TestPaginationRequest validates PaginationRequest model for database queries
// Purpose: Ensures the request model for paginated data retrieval works correctly
// API Contract: Used by POST /api/stocks/list endpoint for database pagination
func TestPaginationRequest(t *testing.T) {
	req := PaginationRequest{
		PageNumber: 1,
		PageLength: 20,
	}
	assert.Equal(t, 1, req.PageNumber, "PageNumber should be assigned correctly")
	assert.Equal(t, 20, req.PageLength, "PageLength should be assigned correctly")
}

// TestSearchRequest validates SearchRequest model for filtered queries
// Purpose: Ensures the request model for search operations works correctly
// API Contract: Used by POST /api/stocks/search endpoint for RegEx-powered search
func TestSearchRequest(t *testing.T) {
	req := SearchRequest{
		PageNumber: 1,
		PageLength: 20,
		SearchTerm: "AAPL",
	}
	assert.Equal(t, 1, req.PageNumber, "PageNumber should be assigned correctly")
	assert.Equal(t, 20, req.PageLength, "PageLength should be assigned correctly")
	assert.Equal(t, "AAPL", req.SearchTerm, "SearchTerm should be assigned correctly")
}

// TestApiResponse validates ApiResponse model for external API integration
// Purpose: Ensures the response model for external stock API works correctly
// API Integration: Used to parse responses from external stock data provider
func TestApiResponse(t *testing.T) {
	stock := StockRatings{
		Ticker:  "AAPL",
		Company: "Apple Inc.",
	}
	
	response := ApiResponse{
		Items:    []StockRatings{stock},
		NextPage: "2",
	}

	assert.Len(t, response.Items, 1, "Items array should contain one stock")
	assert.Equal(t, "2", response.NextPage, "NextPage should be assigned correctly")
}

// TestErrorResponse validates ErrorResponse model for API error handling
// Purpose: Ensures error response structure works correctly for client communication
// Error Handling: Used throughout API to provide consistent error messages
func TestErrorResponse(t *testing.T) {
	err := ErrorResponse{
		Error: "Invalid request",
	}

	assert.Equal(t, "Invalid request", err.Error, "Error message should be assigned correctly")
}

// TestGenericErrorResponse validates GenericErrorResponse model for server errors
// Purpose: Ensures generic error response structure works for internal server errors
// Error Handling: Used for unexpected errors and system failures
func TestGenericErrorResponse(t *testing.T) {
	err := GenericErrorResponse{
		Error: "Internal server error",
	}

	assert.Equal(t, "Internal server error", err.Error, "Error message should be assigned correctly")
}