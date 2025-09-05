package handlers

/*
	Handlers are responsible for processing incoming HTTP requests,
	interacting with the database, and returning appropriate responses.
*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"smart-stock-recommender/models"
	"time"

	"github.com/gin-gonic/gin"
)

// StockHandler handles stock-related requests.
type StockHandler struct {
	DB *sql.DB
}

// NewStockHandler creates a new instance of StockHandler with the given database connection.
// It returns a pointer to the StockHandler.
func NewStockHandler(db *sql.DB) *StockHandler {
	return &StockHandler{DB: db}
}

/*
GetStocksByPage handles fetching stocks from the external API:
  - https://api.karenai.click/swechallenge/list?next_page=AVBP

with pagination.

Expected Body format:

	{
		"page": 1
	}
*/
func (h *StockHandler) GetStocksByPage(c *gin.Context) {
	// Parse JSON from request body
	var req models.PageRequest

	// Decode the JSON request body
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
		return
	}

	// Check if 'page' field is provided
	if req.Page == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required field 'page' in request body"})
		return
	}

	// Validate page number is positive
	if req.Page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page number must be positive"})
		return
	}

	// Fetch from external API
	apiURL := fmt.Sprintf("https://api.karenai.click/swechallenge/list?next_page=%d", req.Page)
	httpReq, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set Authorization Header with the API token from environment variable
	httpReq.Header.Set("Authorization", "Token "+os.Getenv("API_TOKEN"))

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}

	// Get the response
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	// Close the response body
	defer resp.Body.Close()

	// Decode response
	var apiResp models.ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
		return
	}
	println("Fetched", len(apiResp.Items), "items from API page:", req.Page)

	// Store in database
	for _, stock := range apiResp.Items {
		println("Storing stock:", stock.Ticker, "at time:", stock.Time.String())
		h.storeStock(stock)
	}

	// Return the fetched data
	c.JSON(http.StatusOK, apiResp)
}

/*
storeStock saves a stock entry into the database.
*/
func (h *StockHandler) storeStock(stock models.stock_ratings) error {
	// Let's make the query
	query := `
		INSERT INTO stock_ratings (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, time) DO NOTHING`

	// Execute the query with the stock data
	_, err := h.DB.Exec(query,
		stock.Ticker, stock.TargetFrom, stock.TargetTo, stock.Company,
		stock.Action, stock.Brokerage, stock.RatingFrom, stock.RatingTo,
		stock.Time, time.Now())

	if err != nil {
		println("error storing stock:", err.Error())
	} else {
		println("successfully stored stock:", stock.Ticker)
	}
	return err
}
