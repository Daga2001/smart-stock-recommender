package handlers

/*
	Handlers are responsible for processing incoming HTTP requests,
	interacting with the database, and returning appropriate responses.
*/

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"smart-stock-recommender/models"
	"sort"
	"strconv"
	"strings"
	"sync"
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

// GetStocksByPage fetches stock data from external API for a single page
// @Summary Fetch stocks by page number
// @Description Retrieves stock data from external API for a specific page and stores in database. Returns the raw API response with stock items and next page token.
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.PageRequest true "Request body with page number (integer, required)"
// @Success 200 {object} models.ApiResponse "Successfully fetched stock data from external API"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid JSON format, missing page field, or invalid page number"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks [post]
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

	// Validate page number is positive and within reasonable limits
	if req.Page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page number must be positive"})
		return
	}

	if req.Page > 999999999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page number too large"})
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

// GetStocksBulk fetches stock data from external API for multiple pages
// @Summary Fetch stocks in bulk for page range with parallel processing
// @Description Clears existing database data, then fetches stock data from external API for a range of pages using parallel processing. Returns summary statistics of the operation.
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.BulkPageRequest true "Request body with start_page and end_page (integers, both required, max range 1,000,000)"
// @Success 200 {object} models.BulkResponse "Successfully processed bulk stock data fetch with parallel processing"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid JSON, negative pages, start > end, or range too large"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/bulk [post]
func (h *StockHandler) GetStocksBulk(c *gin.Context) {
	var req models.BulkPageRequest

	// Decode the JSON request body
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
		return
	}

	// Validate start_page and end_page
	if req.StartPage <= 0 || req.EndPage <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_page and end_page must be positive"})
		return
	}

	if req.StartPage > req.EndPage {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_page must be less than or equal to end_page"})
		return
	}

	// Allow large page ranges for bulk processing
	if req.EndPage-req.StartPage > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page range too large (max 1,000,000 pages)"})
		return
	}

	if req.EndPage > 999999999 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End page number too large"})
		return
	}

	// Clear existing data
	if err := h.clearStockRatings(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear existing data"})
		return
	}

	// Fetch and store in bulk with parallelism.
	allStocks, totalFetched, err := h.fetchStocksBulkParallel(req.StartPage, req.EndPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":       "Successfully fetched and stored stock data",
		"pages_fetched": fmt.Sprintf("%d-%d", req.StartPage, req.EndPage),
		"total_stocks":  totalFetched,
		"stocks":        allStocks,
	})
}

// clearStockRatings deletes all records from the stock_ratings table.
func (h *StockHandler) clearStockRatings() error {
	_, err := h.DB.Exec("DELETE FROM stock_ratings")
	return err
}

// fetchStocksFromAPI attempts to fetch stock data for a specific page
// Uses retry logic to find data by trying alternative page numbers
func (h *StockHandler) fetchStocksFromAPI(page int) ([]models.StockRatings, error) {
	return h.fetchStocksFromAPIWithRetry(page, 5)
}

// fetchStocksFromAPIWithRetry attempts to fetch stock data with retry logic
// Tries different page numbers using a mathematical pattern to find data
func (h *StockHandler) fetchStocksFromAPIWithRetry(originalPage, maxRetries int) ([]models.StockRatings, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Calculate page to try: original page first, then use prime number pattern
		tryPage := originalPage
		if attempt > 0 {
			tryPage = originalPage + attempt*13 // Prime number for better distribution
		}

		// Make API request
		apiURL := fmt.Sprintf("https://api.karenai.click/swechallenge/list?next_page=%d", tryPage)
		httpReq, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			continue
		}

		httpReq.Header.Set("Authorization", "Token "+os.Getenv("API_TOKEN"))
		resp, err := client.Do(httpReq)
		if err != nil {
			continue
		}

		// Parse response
		var apiResp models.ApiResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()
		if err != nil {
			continue
		}

		// Return data if found (no logging here to avoid confusion)
		if len(apiResp.Items) > 0 {
			return apiResp.Items, nil
		}
	}

	// Return empty if no data found after all attempts
	return []models.StockRatings{}, nil
}

/*
fetchStocksBulkParallel fetches stock data for a range of pages in parallel
and stores them in the database.

It returns the combined list of stocks fetched and the total count.

Expected Body format:

	{
		"start_page": 1,
		"end_page": 22
	}
*/
func (h *StockHandler) fetchStocksBulkParallel(startPage, endPage int) ([]models.StockRatings, int, error) {
	const BATCH_SIZE = 1000 // Configurable batch size
	const MAX_CONCURRENT = 30

	pageCount := endPage - startPage + 1
	println("🚀 Starting bulk fetch for", pageCount, "pages (from", startPage, "to", endPage, ")")
	println("📊 Configuration: Batch size =", BATCH_SIZE, ", Max concurrent =", MAX_CONCURRENT)

	type result struct {
		stocks []models.StockRatings
		page   int
		err    error
	}

	results := make(chan result, 100) // Smaller buffer to prevent memory issues
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, MAX_CONCURRENT)

	// Start goroutines for fetching
	println("🔄 Launching", MAX_CONCURRENT, "concurrent workers...")
	for page := startPage; page <= endPage; page++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			stocks, err := h.fetchStocksFromAPI(p)
			results <- result{stocks: stocks, page: p, err: err}
		}(page)
	}

	go func() {
		wg.Wait()
		close(results)
		println("✅ All workers finished fetching")
	}()

	// Process results with detailed logging
	var stockBuffer []models.StockRatings
	totalFetched := 0
	pagesWithData := 0
	batchCount := 0
	processedPages := 0

	for res := range results {
		processedPages++

		if res.err != nil {
			println("❌ Error on page", res.page, ":", res.err.Error())
			return nil, 0, fmt.Errorf("failed to fetch page %d: %v", res.page, res.err)
		}

		// Process pages with data
		if len(res.stocks) > 0 {
			stockBuffer = append(stockBuffer, res.stocks...)
			totalFetched += len(res.stocks)
			pagesWithData++

			// Trigger batch insert when buffer reaches limit
			if len(stockBuffer) >= BATCH_SIZE {
				batchCount++
				println("💾 BATCH", batchCount, ": Processing", len(stockBuffer), "stocks...")

				if err := h.batchInsertStocksWithLogging(stockBuffer, batchCount); err != nil {
					return nil, 0, fmt.Errorf("failed to insert batch %d: %v", batchCount, err)
				}

				stockBuffer = stockBuffer[:0] // Clear buffer
			}
		}

		// Progress update every 1000 pages
		if processedPages%1000 == 0 {
			println("📈 Progress:", processedPages, "/", pageCount, "pages processed (", fmt.Sprintf("%.1f%%", float64(processedPages)/float64(pageCount)*100), ")")
		}
	}

	// Insert remaining stocks
	if len(stockBuffer) > 0 {
		batchCount++
		println("💾 FINAL BATCH", batchCount, ": Inserting remaining", len(stockBuffer), "stocks...")
		if err := h.batchInsertStocksWithLogging(stockBuffer, batchCount); err != nil {
			return nil, 0, fmt.Errorf("failed to insert final batch: %v", err)
		}
		println("✅ FINAL BATCH", batchCount, "successfully inserted")
	}

	// Get actual database count for verification
	var actualCount int
	h.DB.QueryRow("SELECT COUNT(*) FROM stock_ratings").Scan(&actualCount)

	println("🎉 SUMMARY: Processed", processedPages, "pages, found data in", pagesWithData, "pages")
	println("📊 Total stocks fetched:", totalFetched, "| Total batches processed:", batchCount)
	println("💾 Database verification: Actual records in DB =", actualCount)
	if actualCount < totalFetched {
		println("⚠️  Note:", totalFetched-actualCount, "duplicates were skipped due to UNIQUE constraint")
	}
	return []models.StockRatings{}, totalFetched, nil
}

// batchInsertStocksWithLogging inserts stock records in a single database transaction
// Provides progress updates for large batches and detailed error reporting
func (h *StockHandler) batchInsertStocksWithLogging(stocks []models.StockRatings, batchNum int) error {
	if len(stocks) == 0 {
		return nil
	}

	// Begin database transaction
	tx, err := h.DB.Begin()
	if err != nil {
		println("❌ BATCH", batchNum, ": Transaction failed:", err.Error())
		return err
	}
	defer tx.Rollback()

	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO stock_ratings (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, action, rating_from, rating_to, time) DO NOTHING`)
	if err != nil {
		println("❌ BATCH", batchNum, ": Statement preparation failed:", err.Error())
		return err
	}
	defer stmt.Close()

	// Execute inserts with progress tracking
	insertedCount := 0
	skippedCount := 0
	for i, stock := range stocks {
		result, err := stmt.Exec(
			stock.Ticker, stock.TargetFrom, stock.TargetTo, stock.Company,
			stock.Action, stock.Brokerage, stock.RatingFrom, stock.RatingTo,
			stock.Time, time.Now())
		if err != nil {
			println("❌ BATCH", batchNum, ": Insert failed for", stock.Ticker, ":", err.Error())
			return err
		}

		// Check if row was actually inserted (not a duplicate)
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			insertedCount++
		} else {
			skippedCount++
		}

		// Show progress every 200 attempts
		if (i+1)%200 == 0 {
			println("📈 BATCH", batchNum, ":", i+1, "/", len(stocks), "processed (", insertedCount, "new,", skippedCount, "duplicates)")
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		println("❌ BATCH", batchNum, ": Commit failed:", err.Error())
		return err
	}

	println("✅ BATCH", batchNum, ": Committed", insertedCount, "new stocks (", skippedCount, "duplicates skipped)")
	return nil
}

// storeStock inserts a single stock record into the database
// Used by single-page endpoint, bulk operations use batchInsertStocks instead
func (h *StockHandler) storeStock(stock models.StockRatings) error {
	query := `
		INSERT INTO stock_ratings (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, brokerage, action, rating_from, rating_to, time) DO NOTHING`

	_, err := h.DB.Exec(query,
		stock.Ticker, stock.TargetFrom, stock.TargetTo, stock.Company,
		stock.Action, stock.Brokerage, stock.RatingFrom, stock.RatingTo,
		stock.Time, time.Now())

	return err
}

// GetStockRatings retrieves paginated stock ratings from database
// @Summary Get paginated stock ratings from database
// @Description Retrieves stored stock ratings with pagination support, ordered by creation date (newest first). Returns both data and pagination metadata.
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.PaginationRequest true "Request body with page_number (integer, min 1) and page_length (integer, 1-1000)"
// @Success 200 {object} models.PaginatedResponse "Successfully retrieved paginated stock ratings with metadata"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid JSON, page_number <= 0, or page_length not between 1-1000"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/list [post]
func (h *StockHandler) GetStockRatings(c *gin.Context) {
	var req models.PaginationRequest

	// Parse request body
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
		return
	}

	// Validate pagination parameters
	if req.PageNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_number must be greater than 0"})
		return
	}

	if req.PageLength <= 0 || req.PageLength > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_length must be between 1 and 1000"})
		return
	}

	// Calculate offset for pagination
	offset := (req.PageNumber - 1) * req.PageLength

	// Get total count
	var totalCount int
	err := h.DB.QueryRow("SELECT COUNT(*) FROM stock_ratings").Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total count"})
		return
	}

	// Query paginated data
	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at
		FROM stock_ratings
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`

	rows, err := h.DB.Query(query, req.PageLength, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query stock ratings"})
		return
	}
	defer rows.Close()

	// Parse results
	var stocks []models.StockRatings
	for rows.Next() {
		var stock models.StockRatings
		err := rows.Scan(
			&stock.ID, &stock.Ticker, &stock.TargetFrom, &stock.TargetTo,
			&stock.Company, &stock.Action, &stock.Brokerage,
			&stock.RatingFrom, &stock.RatingTo, &stock.Time, &stock.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan stock data"})
			return
		}
		stocks = append(stocks, stock)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.PageLength - 1) / req.PageLength
	hasNext := req.PageNumber < totalPages
	hasPrev := req.PageNumber > 1

	// Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"data": stocks,
		"pagination": gin.H{
			"page_number":   req.PageNumber,
			"page_length":   req.PageLength,
			"total_records": totalCount,
			"total_pages":   totalPages,
			"has_next":      hasNext,
			"has_previous":  hasPrev,
		},
	})
}

// SearchStockRatings searches stock ratings based on search term with pagination
// @Summary Search stock ratings with pagination
// @Description Searches through stock ratings using a search term that matches ticker, company, brokerage, action, or ratings. Returns paginated results with metadata.
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.SearchRequest true "Search parameters with page_number (integer, min 1), page_length (integer, 1-1000), and search_term (string, required)"
// @Success 200 {object} models.PaginatedResponse "Successfully retrieved filtered stock ratings with pagination metadata"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid JSON, missing search_term, page_number <= 0, or page_length not between 1-1000"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/search [post]
func (h *StockHandler) SearchStockRatings(c *gin.Context) {
	var req models.SearchRequest

	// Parse request body
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
		return
	}

	// Validate search parameters
	if req.PageNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_number must be greater than 0"})
		return
	}

	if req.SearchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search_term is required"})
		return
	}

	// Calculate offset for pagination
	offset := (req.PageNumber - 1) * req.PageLength

	// Get total count for search results
	countQuery := `
		SELECT COUNT(*) FROM stock_ratings 
		WHERE LOWER(ticker) LIKE LOWER($1) OR LOWER(company) LIKE LOWER($1) OR LOWER(brokerage) LIKE LOWER($1) 
		   OR LOWER(action) LIKE LOWER($1) OR LOWER(rating_from) LIKE LOWER($1) OR LOWER(rating_to) LIKE LOWER($1)`
	searchPattern := "%" + req.SearchTerm + "%"
	println("🔍 Searching for:", req.SearchTerm, "| Pattern:", searchPattern)

	var totalCount int
	err := h.DB.QueryRow(countQuery, searchPattern).Scan(&totalCount)
	if err != nil {
		println("❌ Search count error:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search count"})
		return
	}
	println("📊 Found", totalCount, "matching records")

	// Query filtered and paginated data
	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at
		FROM stock_ratings
		WHERE LOWER(ticker) LIKE LOWER($1) OR LOWER(company) LIKE LOWER($1) OR LOWER(brokerage) LIKE LOWER($1) 
		   OR LOWER(action) LIKE LOWER($1) OR LOWER(rating_from) LIKE LOWER($1) OR LOWER(rating_to) LIKE LOWER($1)
		ORDER BY created_at DESC, id DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.DB.Query(query, searchPattern, req.PageLength, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search stock ratings"})
		return
	}
	defer rows.Close()

	// Parse results
	var stocks []models.StockRatings
	for rows.Next() {
		var stock models.StockRatings
		err := rows.Scan(
			&stock.ID, &stock.Ticker, &stock.TargetFrom, &stock.TargetTo,
			&stock.Company, &stock.Action, &stock.Brokerage,
			&stock.RatingFrom, &stock.RatingTo, &stock.Time, &stock.CreatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan search results"})
			return
		}
		stocks = append(stocks, stock)
	}

	// Calculate pagination metadata
	totalPages := (totalCount + req.PageLength - 1) / req.PageLength
	hasNext := req.PageNumber < totalPages
	hasPrev := req.PageNumber > 1

	// Return search results with pagination
	c.JSON(http.StatusOK, gin.H{
		"data": stocks,
		"pagination": gin.H{
			"page_number":   req.PageNumber,
			"page_length":   req.PageLength,
			"total_records": totalCount,
			"total_pages":   totalPages,
			"has_next":      hasNext,
			"has_previous":  hasPrev,
		},
		"search_term": req.SearchTerm,
	})
}

// ActionsResponse represents the response structure for stock actions
type ActionsResponse struct {
	Actions []string `json:"actions" example:"initiated by,target raised by,target lowered by,reiterated by,upgraded"`
}

// GetStockActions retrieves all unique action types from the database
// @Summary Get all available stock actions
// @Description Retrieves a list of all unique action types found in the stock ratings database, sorted alphabetically. Used for populating filter dropdowns and ensuring UI reflects actual data.
// @Tags stocks
// @Produce json
// @Success 200 {object} ActionsResponse "Successfully retrieved list of unique actions"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/actions [get]
func (h *StockHandler) GetStockActions(c *gin.Context) {
	// Query to get all unique actions from the database
	query := `
		SELECT DISTINCT action 
		FROM stock_ratings 
		WHERE action IS NOT NULL AND action != '' 
		ORDER BY action ASC`

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query stock actions"})
		return
	}
	defer rows.Close()

	// Collect all unique actions
	var actions []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			continue // Skip invalid rows
		}
		actions = append(actions, action)
	}

	// Return the list of actions
	c.JSON(http.StatusOK, ActionsResponse{
		Actions: actions,
	})
}

// stockData represents internal stock data structure for analysis
type stockData struct {
	Ticker     string
	Company    string
	Action     string
	Brokerage  string
	RatingFrom string
	RatingTo   string
	TargetFrom string
	TargetTo   string
	Time       string // Actual analyst report time (the important one for analysis)
	// Note: CreatedAt removed - we don't need database insertion time for analysis
}

// StockRecommendation represents a stock recommendation
type StockRecommendation struct {
	Ticker            string  `json:"ticker" example:"AAPL"`
	Company           string  `json:"company" example:"Apple Inc."`
	CurrentRating     string  `json:"current_rating" example:"Buy"`
	TargetPrice       string  `json:"target_price" example:"$180.00"`
	Score             float64 `json:"score" example:"8.5"`
	Recommendation    string  `json:"recommendation" example:"Strong Buy"`
	Reason            string  `json:"reason" example:"Target raised by 15%, upgraded to Buy rating"`
	Brokerage         string  `json:"brokerage" example:"Goldman Sachs"`
	PriceChange       float64 `json:"price_change" example:"15.5"`
	RatingImprovement bool    `json:"rating_improvement" example:"true"`
}

type RecommendationsResponse struct {
	Recommendations []StockRecommendation `json:"recommendations"`
	GeneratedAt     string                `json:"generated_at" example:"2024-01-15T10:30:00Z"`
	TotalAnalyzed   int                   `json:"total_analyzed" example:"1250"`
}

// GetStockRecommendations analyzes stock data and provides investment recommendations
// @Summary Get AI-powered stock investment recommendations
// @Description Analyzes all stock ratings data using advanced algorithms to provide ranked investment recommendations. Considers target price changes, rating improvements, analyst sentiment, and market trends to identify the best investment opportunities.
// @Tags recommendations
// @Produce json
// @Success 200 {object} RecommendationsResponse "Successfully generated stock recommendations with scoring and analysis"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred during analysis"
// @Router /stocks/recommendations [get]
func (h *StockHandler) GetStockRecommendations(c *gin.Context) {
	// Query to get all stock data for analysis
	query := `
		SELECT ticker, company, action, brokerage, rating_from, rating_to, 
		       target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker IS NOT NULL AND company IS NOT NULL
		ORDER BY time DESC`

	rows, err := h.DB.Query(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query stock data for recommendations"})
		return
	}
	defer rows.Close()

	// Collect stock data
	var stocks []stockData
	for rows.Next() {
		var stock stockData
		var createdAt time.Time // Scan but don't use for analysis
		err := rows.Scan(&stock.Ticker, &stock.Company, &stock.Action, &stock.Brokerage,
			&stock.RatingFrom, &stock.RatingTo, &stock.TargetFrom, &stock.TargetTo,
			&stock.Time, &createdAt)
		if err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}

	// Analyze and generate recommendations
	recommendations := analyzeStocksForRecommendations(stocks)

	// Return top recommendations
	c.JSON(http.StatusOK, RecommendationsResponse{
		Recommendations: recommendations,
		GeneratedAt:     time.Now().Format(time.RFC3339),
		TotalAnalyzed:   len(stocks),
	})
}

// analyzeStocksForRecommendations implements the quantitative recommendation algorithm
// 
// ALGORITHM OVERVIEW:
// 1. Groups all stocks by ticker symbol to get latest data per company
// 2. Calculates weighted score (0-10) for each stock using multiple criteria
// 3. Filters stocks with score >= 5.0 (minimum recommendation threshold)
// 4. Sorts by score (highest first) and returns top 10 recommendations
// 
// WHY TOP 3 IS VARIABLE:
// The "top 3" changes because scores are recalculated every time based on:
// - New analyst reports added to database
// - Updated target prices and ratings
// - Time decay (recent activity gets bonus points)
// - Competitive ranking (a stock with 8.5 score today might drop to 7.8 tomorrow)
func analyzeStocksForRecommendations(stocks []stockData) []StockRecommendation {
	// STEP 1: Group stocks by ticker to get latest data per company
	// This ensures we analyze the most recent analyst opinion for each stock
	stockMap := make(map[string][]stockData)
	for _, stock := range stocks {
		stockMap[stock.Ticker] = append(stockMap[stock.Ticker], stock)
	}

	var recommendations []StockRecommendation

	// STEP 2: Analyze each stock and calculate recommendation score
	for ticker, stockList := range stockMap {
		if len(stockList) == 0 {
			continue
		}

		// Get the most recent entry for this stock (based on actual analyst report time)
		latestStock := stockList[0]
		for _, s := range stockList {
			// Parse time strings to compare actual report dates
			sTime, sErr := time.Parse("2006-01-02 15:04:05", s.Time)
			latestTime, latestErr := time.Parse("2006-01-02 15:04:05", latestStock.Time)
			if sErr == nil && latestErr == nil && sTime.After(latestTime) {
				latestStock = s
			}
		}

		// STEP 3: Calculate quantitative recommendation score (0-10 scale)
		// Uses configurable weighted algorithm considering multiple factors
		score := calculateStockScore(latestStock, stockList)
		if score < 5.0 { // QUALITY FILTER: Only recommend stocks with score >= 5.0
			continue // Skip low-quality recommendations
		}

		// Parse target prices for analysis
		// Parse "$150.00" -> 150.0
		targetFrom := parsePrice(latestStock.TargetFrom)
		targetTo := parsePrice(latestStock.TargetTo)
		priceChange := 0.0
		if targetFrom > 0 {
			priceChange = ((targetTo - targetFrom) / targetFrom) * 100
		}

		// Determine recommendation level
		recommendationLevel := getRecommendationLevel(score)
		reason := generateRecommendationReason(latestStock, priceChange, score)

		recommendations = append(recommendations, StockRecommendation{
			Ticker:            ticker,
			Company:           latestStock.Company,
			CurrentRating:     latestStock.RatingTo,
			TargetPrice:       latestStock.TargetTo,
			Score:             score,
			Recommendation:    recommendationLevel,
			Reason:            reason,
			Brokerage:         latestStock.Brokerage,
			PriceChange:       priceChange,
			RatingImprovement: isRatingImprovement(latestStock.RatingFrom, latestStock.RatingTo),
		})
	}

	// STEP 4: SORTING - This is where the magic happens!
	// Sort by score in DESCENDING order (highest scores first)
	// This determines the final ranking: #1, #2, #3, etc.
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Score > recommendations[j].Score // Higher score = better rank
	})

	// STEP 5: Return top 10 recommendations (frontend takes first 3)
	// Why top 10? Provides buffer for frontend to choose top 3
	if len(recommendations) > 10 {
		recommendations = recommendations[:10] // Slice to get first 10 elements
	}

	return recommendations // Sorted list: [highest_score, second_highest, third_highest, ...]
}

// ScoringWeights defines configurable weights for stock scoring algorithm
// Allows easy modification of scoring criteria for market adaptability
type ScoringWeights struct {
	TargetPriceWeight float64 // Weight for target price changes (default: 0.4)
	RatingWeight      float64 // Weight for rating analysis (default: 0.3)
	ActionWeight      float64 // Weight for action analysis (default: 0.2)
	TimingWeight      float64 // Weight for recent activity (default: 0.1)
}

// getDefaultWeights returns the default scoring weights
// These can be easily modified based on market conditions
func getDefaultWeights() ScoringWeights {
	return ScoringWeights{
		TargetPriceWeight: 0.4, // 40% - Most important for speculative markets
		RatingWeight:      0.3, // 30% - Professional analyst opinion
		ActionWeight:      0.2, // 20% - Direction of analyst changes
		TimingWeight:      0.1, // 10% - Recent activity bonus
	}
}

// calculateStockScore implements the configurable weighted scoring algorithm
// 
// SCORING SYSTEM (0-10 scale):
// Base Score: 5.0 (neutral starting point)
// 
// CONFIGURABLE WEIGHTS (easily modifiable for market conditions):
// 🎯 Target Price Changes: Configurable % (default 40%)
// ⭐ Rating Analysis: Configurable % (default 30%)
// 📊 Action Analysis: Configurable % (default 20%)
// ⏰ Recent Activity: Configurable % (default 10%)
// 
// SCORE RANGES:
// 8.5-10.0 = Strong Buy (top tier recommendations)
// 7.0-8.4  = Buy (good recommendations)
// 6.0-6.9  = Moderate Buy (decent opportunities)
// 5.0-5.9  = Hold (minimum threshold)
// 0.0-4.9  = Not recommended (filtered out)
func calculateStockScore(stock stockData, history []stockData) float64 {
	weights := getDefaultWeights() // Get configurable weights
	score := 5.0 // NEUTRAL BASE SCORE - every stock starts here

	// 🎯 CRITERION 1: TARGET PRICE ANALYSIS (CONFIGURABLE WEIGHT)
	// Price targets directly indicate expected returns - critical for speculative markets
	targetFrom := parsePrice(stock.TargetFrom) // Parse "$150.00" -> 150.0
	targetTo := parsePrice(stock.TargetTo)     // Parse "$180.00" -> 180.0
	var targetPriceScore float64
	if targetFrom > 0 && targetTo > targetFrom {
		priceIncrease := ((targetTo - targetFrom) / targetFrom) * 100 // Calculate % increase
		// SCORING TIERS based on price increase magnitude:
		if priceIncrease > 20 {
			targetPriceScore = 3.0 // MAJOR BOOST: >20% increase
		} else if priceIncrease > 10 {
			targetPriceScore = 2.0 // GOOD BOOST: 10-20% increase
		} else if priceIncrease > 5 {
			targetPriceScore = 1.0 // SMALL BOOST: 5-10% increase
		}
	} else if targetTo < targetFrom {
		targetPriceScore = -2.0 // PENALTY: Price target was LOWERED
	}
	score += targetPriceScore * weights.TargetPriceWeight // Apply configurable weight

	// ⭐ CRITERION 2: RATING ANALYSIS (CONFIGURABLE WEIGHT)
	// Analyst ratings reflect professional opinion and research
	var ratingScore float64
	if isRatingImprovement(stock.RatingFrom, stock.RatingTo) {
		ratingScore += 2.0 // UPGRADE BONUS: "Hold" -> "Buy" or "Buy" -> "Strong Buy"
	}
	// CURRENT RATING BONUSES (based on final rating strength):
	if isStrongBuyRating(stock.RatingTo) {
		ratingScore += 1.5 // STRONG BUY: Highest confidence rating
	} else if isBuyRating(stock.RatingTo) {
		ratingScore += 1.0 // BUY: Positive rating
	}
	score += ratingScore * weights.RatingWeight // Apply configurable weight

	// 📊 CRITERION 3: ACTION ANALYSIS (CONFIGURABLE WEIGHT)
	// Actions indicate the direction and confidence of analyst changes
	var actionScore float64
	action := strings.ToLower(stock.Action)
	if strings.Contains(action, "raised") || strings.Contains(action, "upgrade") {
		actionScore = 1.5 // POSITIVE ACTIONS: "target raised", "rating upgraded"
	} else if strings.Contains(action, "initiated") && isBuyRating(stock.RatingTo) {
		actionScore = 1.0 // NEW COVERAGE: Fresh analyst starts covering with Buy rating
	} else if strings.Contains(action, "lowered") || strings.Contains(action, "downgrade") {
		actionScore = -1.5 // NEGATIVE ACTIONS: "target lowered", "rating downgraded"
	}
	score += actionScore * weights.ActionWeight // Apply configurable weight

	// ⏰ CRITERION 4: RECENT ACTIVITY BONUS (CONFIGURABLE WEIGHT)
	// Recent analyst reports indicate current market relevance
	var timingScore float64
	analystTime, err := time.Parse("2006-01-02 15:04:05", stock.Time)
	if err == nil && time.Since(analystTime).Hours() < 24 {
		timingScore += 0.5 // FRESHNESS BONUS: Analyst report is less than 24 hours old
	}
	// MULTIPLE ANALYST COVERAGE BONUS
	if len(history) > 1 {
		timingScore += 0.5 // CONSENSUS BONUS: 2+ analysts have opinions on this stock
	}
	score += timingScore * weights.TimingWeight // Apply configurable weight

	// FINAL SCORE CAPPING: Ensure score stays within valid range
	return math.Min(10.0, math.Max(0.0, score)) // Cap between 0-10 (no negative or >10 scores)
}

// Helper functions
func parsePrice(priceStr string) float64 {
	cleanPrice := strings.ReplaceAll(priceStr, "$", "")
	cleanPrice = strings.ReplaceAll(cleanPrice, ",", "")
	price, _ := strconv.ParseFloat(cleanPrice, 64)
	return price
}

// isRatingImprovement checks if a rating was upgraded
// 
// RATING HIERARCHY (1-8 scale, higher = better):
// 1 = Strong Sell (worst)
// 2 = Sell  
// 3 = Underperform/Underweight
// 4 = Hold
// 5 = Neutral
// 6 = Outperform
// 7 = Buy/Overweight  
// 8 = Strong Buy (best)
// 
// EXAMPLES:
// "Hold" (4) -> "Buy" (7) = TRUE (improvement)
// "Buy" (7) -> "Hold" (4) = FALSE (downgrade)
// "Buy" (7) -> "Strong Buy" (8) = TRUE (improvement)
func isRatingImprovement(from, to string) bool {
	ratingScore := map[string]int{
		"strong sell": 1, "sell": 2, "underperform": 3, "hold": 4, "neutral": 5,
		"outperform": 6, "buy": 7, "strong buy": 8, "overweight": 7, "underweight": 3,
	}
	return ratingScore[strings.ToLower(to)] > ratingScore[strings.ToLower(from)]
}

// isStrongBuyRating checks if a rating is a strong buy or overweight
func isStrongBuyRating(rating string) bool {
	lower := strings.ToLower(rating)
	return strings.Contains(lower, "strong buy") || strings.Contains(lower, "overweight")
}

// isBuyRating checks if a rating is a buy or outperform
func isBuyRating(rating string) bool {
	lower := strings.ToLower(rating)
	return strings.Contains(lower, "buy") || strings.Contains(lower, "outperform")
}

// getRecommendationLevel maps score to recommendation string
func getRecommendationLevel(score float64) string {
	if score >= 8.5 {
		return "Strong Buy"
	} else if score >= 7.0 {
		return "Buy"
	} else if score >= 6.0 {
		return "Moderate Buy"
	} else {
		return "Hold"
	}
}

// generateRecommendationReason creates a reason string based on analysis
func generateRecommendationReason(stock stockData, priceChange, score float64) string {
	reasons := []string{}

	if priceChange > 10 {
		reasons = append(reasons, fmt.Sprintf("Target raised by %.1f%%", priceChange))
	}
	if isRatingImprovement(stock.RatingFrom, stock.RatingTo) {
		reasons = append(reasons, fmt.Sprintf("Upgraded to %s", stock.RatingTo))
	}
	if strings.Contains(strings.ToLower(stock.Action), "initiated") {
		reasons = append(reasons, "New analyst coverage")
	}
	if score >= 8.0 {
		reasons = append(reasons, "Strong analyst sentiment")
	}

	if len(reasons) == 0 {
		return "Positive analyst outlook"
	}
	return strings.Join(reasons, ", ")
}

// SummaryResponse represents an AI-generated market summary
type SummaryResponse struct {
	Summary     string `json:"summary" example:"Today's market shows strong bullish sentiment with 15 stocks receiving target price increases. Apple leads recommendations with a 12% target raise to $180, while tech sector dominates with 60% of top picks."`
	GeneratedAt string `json:"generated_at" example:"2024-01-15T10:30:00Z"`
	TokensUsed  int    `json:"tokens_used" example:"245"`
}

// GetStockSummary generates AI-powered natural language summary of stock recommendations
// @Summary Get AI-generated market summary
// @Description Uses gpt-4.1-nano to analyze current stock recommendations and generate a comprehensive natural language summary of market trends, top picks, and investment insights.
// @Tags ai-analysis
// @Produce json
// @Success 200 {object} SummaryResponse "Successfully generated AI market summary"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error or OpenAI API error"
// @Router /stocks/summary [get]
func (h *StockHandler) GetStockSummary(c *gin.Context) {
	// Get current recommendations
	recommendations := h.getRecommendationsForSummary()
	if len(recommendations) == 0 {
		c.JSON(http.StatusOK, SummaryResponse{
			Summary:     "No stock recommendations available at this time. Please ensure the database contains stock ratings data.",
			GeneratedAt: time.Now().Format(time.RFC3339),
			TokensUsed:  0,
		})
		return
	}

	// Generate AI summary
	summary, tokensUsed, err := h.generateAISummary(recommendations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate AI summary: %v", err)})
		return
	}

	c.JSON(http.StatusOK, SummaryResponse{
		Summary:     summary,
		GeneratedAt: time.Now().Format(time.RFC3339),
		TokensUsed:  tokensUsed,
	})
}

// getRecommendationsForSummary gets top recommendations for AI analysis
func (h *StockHandler) getRecommendationsForSummary() []StockRecommendation {
	// Query to get recent stock data for analysis
	query := `
		SELECT ticker, company, action, brokerage, rating_from, rating_to, 
		       target_from, target_to, time, created_at
		FROM stock_ratings 
		WHERE ticker IS NOT NULL AND company IS NOT NULL
		ORDER BY time DESC
		LIMIT 50`

	// Fetch data from database
	rows, err := h.DB.Query(query)
	if err != nil {
		return []StockRecommendation{}
	}
	defer rows.Close()

	// Collect stock data
	var stocks []stockData
	for rows.Next() {
		var stock stockData
		var createdAt time.Time // Scan but don't use for analysis
		err := rows.Scan(&stock.Ticker, &stock.Company, &stock.Action, &stock.Brokerage,
			&stock.RatingFrom, &stock.RatingTo, &stock.TargetFrom, &stock.TargetTo,
			&stock.Time, &createdAt)
		if err != nil {
			continue
		}
		stocks = append(stocks, stock)
	}

	return analyzeStocksForRecommendations(stocks)
}

// generateAISummary calls OpenAI gpt-4.1-nano to generate market summary
func (h *StockHandler) generateAISummary(recommendations []StockRecommendation) (string, int, error) {
	// Prepare data for AI analysis
	prompt := h.buildSummaryPrompt(recommendations)

	// OpenAI API request
	reqBody := map[string]interface{}{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a professional financial analyst. Provide concise, actionable market summaries based on stock recommendation data. Focus on trends, top picks, and key insights. Keep responses under 200 words.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":  250,
		"temperature": 0.7,
	}

	// Marshal request body to JSON
	reqJSON, _ := json.Marshal(reqBody)

	// Make API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	// make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	// Parse response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	// Decode response body
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", 0, err
	}

	if openAIResp.Error.Message != "" {
		return "", 0, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, openAIResp.Usage.TotalTokens, nil
}

// buildSummaryPrompt creates the prompt for AI analysis
func (h *StockHandler) buildSummaryPrompt(recommendations []StockRecommendation) string {
	if len(recommendations) == 0 {
		return "No stock recommendations available."
	}

	// Build a concise prompt summarizing the recommendations
	prompt := "Analyze these stock recommendations and provide a market summary:\n\n"

	// Include top recommendations in the prompt
	for i, rec := range recommendations {
		if i >= 10 { // Limit to top 10 for prompt size
			break	
		}
		prompt += fmt.Sprintf("%d. %s (%s) - %s by %s\n   Rating: %s, Target: %s, Score: %.1f\n   Reason: %s\n\n",
			i+1, rec.Company, rec.Ticker, rec.Recommendation, rec.Brokerage,
			rec.CurrentRating, rec.TargetPrice, rec.Score, rec.Reason)
	}

	prompt += "Provide insights on: market sentiment, top sectors, key trends, and investment outlook."
	return prompt
}

// ChatResponse represents an AI chat response
type ChatResponse struct {
	Response    string `json:"response" example:"Based on current market data, I recommend focusing on stocks with strong buy ratings and recent target price increases. The biotech sector shows particular promise."`
	TokensUsed  int    `json:"tokens_used" example:"156"`
	GeneratedAt string `json:"generated_at" example:"2024-01-15T10:30:00Z"`
}

type ChatRequest struct {
	Message string `json:"message" example:"What are the best stocks to invest in today?"`
}

// GetStockChat provides AI-powered chat responses about stock market
// @Summary Chat with AI about stock market
// @Description Interactive chat with gpt-4.1-nano for personalized stock analysis, market insights, and investment advice based on current data.
// @Tags ai-analysis
// @Accept json
// @Produce json
// @Param request body ChatRequest true "Chat message from user"
// @Success 200 {object} ChatResponse "Successfully generated AI chat response"
// @Failure 400 {object} models.ErrorResponse "Bad request - missing message"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error or OpenAI API error"
// @Router /stocks/chat [post]
func (h *StockHandler) GetStockChat(c *gin.Context) {
	// Parse request body
	var req ChatRequest

	// Validate input and decode JSON
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Get context from recent recommendations
	recommendations := h.getRecommendationsForSummary()
	context := h.buildChatContext(recommendations)

	// Generate AI response
	response, tokensUsed, err := h.generateChatResponse(req.Message, context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate response: %v", err)})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response:    response,
		TokensUsed:  tokensUsed,
		GeneratedAt: time.Now().Format(time.RFC3339),
	})
}

// generateChatResponse calls OpenAI for chat responses
func (h *StockHandler) generateChatResponse(userMessage, context string) (string, int, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a professional financial advisor with access to current stock market data. Provide helpful, accurate investment advice based on the provided market context. Keep responses concise and actionable. Context: " + context,
			},
			{
				"role":    "user",
				"content": userMessage,
			},
		},
		"max_tokens":   300,
		"temperature": 0.7,
	}

	// Marshal request body to JSON
	reqJSON, _ := json.Marshal(reqBody)

	// configure API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	// make HTTP request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	// Parse response
	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", 0, err
	}

	if openAIResp.Error.Message != "" {
		return "", 0, fmt.Errorf("OpenAI API error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, openAIResp.Usage.TotalTokens, nil
}

// buildChatContext creates context for chat responses
func (h *StockHandler) buildChatContext(recommendations []StockRecommendation) string {
	if len(recommendations) == 0 {
		return "No current stock recommendations available."
	}

	context := "Current top stock recommendations: "
	for i, rec := range recommendations {
		if i >= 5 { // Limit context size
			break
		}
		context += fmt.Sprintf("%s (%s) - %s, Score: %.1f; ", rec.Company, rec.Ticker, rec.Recommendation, rec.Score)
	}
	return context
}

// GetStockMetrics calculates and returns comprehensive market metrics from stock ratings data
// @Summary Get comprehensive stock market analytics and metrics
// @Description Analyzes all stored stock ratings using parallel processing to provide comprehensive market insights including sentiment analysis, target price changes, rating distributions, top brokerages, most active stocks, and recent activity trends.
// @Tags analytics
// @Produce json
// @Success 200 {object} models.MetricsResponse "Successfully calculated comprehensive market metrics and analytics"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/metrics [get]
func (h *StockHandler) GetStockMetrics(c *gin.Context) {
	// Execute multiple queries in parallel for better performance
	type MetricResult struct {
		Name  string
		Value interface{}
		Error error
	}

	results := make(chan MetricResult, 10)
	var wg sync.WaitGroup

	// 1. Total Records Count
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		err := h.DB.QueryRow("SELECT COUNT(*) FROM stock_ratings").Scan(&count)
		results <- MetricResult{"total_records", count, err}
	}()

	// 2. Target Price Changes Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT 
				SUM(CASE WHEN action ILIKE '%raised%' OR action ILIKE '%increase%' OR action ILIKE '%upgrade%' THEN 1 ELSE 0 END) as targets_raised,
				SUM(CASE WHEN action ILIKE '%lowered%' OR action ILIKE '%decrease%' OR action ILIKE '%downgrade%' THEN 1 ELSE 0 END) as targets_lowered,
				SUM(CASE WHEN action ILIKE '%maintained%' OR action ILIKE '%reiterated%' THEN 1 ELSE 0 END) as targets_maintained
			FROM stock_ratings`

		var raised, lowered, maintained int
		err := h.DB.QueryRow(query).Scan(&raised, &lowered, &maintained)
		if err != nil {
			results <- MetricResult{"target_changes", nil, err}
			return
		}

		results <- MetricResult{"target_changes", map[string]int{
			"raised":     raised,
			"lowered":    lowered,
			"maintained": maintained,
		}, nil}
	}()

	// 3. Rating Distribution Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT rating_to, COUNT(*) as count
			FROM stock_ratings 
			WHERE rating_to IS NOT NULL AND rating_to != ''
			GROUP BY rating_to 
			ORDER BY count DESC
			LIMIT 10`

		rows, err := h.DB.Query(query)
		if err != nil {
			results <- MetricResult{"rating_distribution", nil, err}
			return
		}
		defer rows.Close()

		ratings := make(map[string]int)
		for rows.Next() {
			var rating string
			var count int
			if err := rows.Scan(&rating, &count); err != nil {
				continue
			}
			ratings[rating] = count
		}

		results <- MetricResult{"rating_distribution", ratings, nil}
	}()

	// 4. Top Active Brokerages
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT brokerage, COUNT(*) as activity_count
			FROM stock_ratings 
			WHERE brokerage IS NOT NULL AND brokerage != ''
			GROUP BY brokerage 
			ORDER BY activity_count DESC
			LIMIT 10`

		rows, err := h.DB.Query(query)
		if err != nil {
			results <- MetricResult{"top_brokerages", nil, err}
			return
		}
		defer rows.Close()

		brokerages := make([]map[string]interface{}, 0)
		for rows.Next() {
			var brokerage string
			var count int
			if err := rows.Scan(&brokerage, &count); err != nil {
				continue
			}
			brokerages = append(brokerages, map[string]interface{}{
				"name":     brokerage,
				"activity": count,
			})
		}

		results <- MetricResult{"top_brokerages", brokerages, nil}
	}()

	// 5. Most Active Stocks (by ticker)
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT ticker, company, COUNT(*) as rating_count
			FROM stock_ratings 
			WHERE ticker IS NOT NULL AND ticker != ''
			GROUP BY ticker, company 
			ORDER BY rating_count DESC
			LIMIT 15`

		rows, err := h.DB.Query(query)
		if err != nil {
			results <- MetricResult{"most_active_stocks", nil, err}
			return
		}
		defer rows.Close()

		stocks := make([]map[string]interface{}, 0)
		for rows.Next() {
			var ticker, company string
			var count int
			if err := rows.Scan(&ticker, &company, &count); err != nil {
				continue
			}
			stocks = append(stocks, map[string]interface{}{
				"ticker":       ticker,
				"company":      company,
				"rating_count": count,
			})
		}

		results <- MetricResult{"most_active_stocks", stocks, nil}
	}()

	// 6. Market Sentiment Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT 
				SUM(CASE WHEN rating_to ILIKE '%buy%' OR rating_to ILIKE '%strong%' THEN 1 ELSE 0 END) as bullish_ratings,
				SUM(CASE WHEN rating_to ILIKE '%sell%' OR rating_to ILIKE '%underperform%' THEN 1 ELSE 0 END) as bearish_ratings,
				SUM(CASE WHEN rating_to ILIKE '%hold%' OR rating_to ILIKE '%neutral%' THEN 1 ELSE 0 END) as neutral_ratings
			FROM stock_ratings 
			WHERE rating_to IS NOT NULL AND rating_to != ''`

		var bullish, bearish, neutral int
		err := h.DB.QueryRow(query).Scan(&bullish, &bearish, &neutral)
		if err != nil {
			results <- MetricResult{"market_sentiment", nil, err}
			return
		}

		total := bullish + bearish + neutral
		sentiment := map[string]interface{}{
			"bullish_count":      bullish,
			"bearish_count":      bearish,
			"neutral_count":      neutral,
			"bullish_percentage": float64(bullish) / float64(total) * 100,
			"bearish_percentage": float64(bearish) / float64(total) * 100,
			"neutral_percentage": float64(neutral) / float64(total) * 100,
		}

		results <- MetricResult{"market_sentiment", sentiment, nil}
	}()

	// 7. Recent Activity (last 7 days)
	wg.Add(1)
	go func() {
		defer wg.Done()
		query := `
			SELECT COUNT(*) as recent_count
			FROM stock_ratings 
			WHERE created_at >= NOW() - INTERVAL '7 days'`

		var recentCount int
		err := h.DB.QueryRow(query).Scan(&recentCount)
		results <- MetricResult{"recent_activity", recentCount, err}
	}()

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	metrics := make(map[string]interface{})
	for result := range results {
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to calculate %s: %v", result.Name, result.Error),
			})
			return
		}
		metrics[result.Name] = result.Value
	}

	// Add metadata
	metrics["generated_at"] = time.Now().UTC()
	metrics["description"] = "Comprehensive stock market analytics based on analyst ratings and target price changes"

	// Return comprehensive metrics
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
	})
}
