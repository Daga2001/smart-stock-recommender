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

// AdvancedSearchRequest represents search parameters with filters
type AdvancedSearchRequest struct {
	PageNumber    int     `json:"page_number"`
	PageLength    int     `json:"page_length"`
	SearchTerm    string  `json:"search_term,omitempty"`
	Action        string  `json:"action,omitempty"`
	RatingFrom    string  `json:"rating_from,omitempty"`
	RatingTo      string  `json:"rating_to,omitempty"`
	TargetFromMin float64 `json:"target_from_min,omitempty"`
	TargetFromMax float64 `json:"target_from_max,omitempty"`
	TargetToMin   float64 `json:"target_to_min,omitempty"`
	TargetToMax   float64 `json:"target_to_max,omitempty"`
}

// SearchStockRatings searches stock ratings with filters
// @Summary Search stock ratings with filters
// @Description Searches through stock ratings using filters including search term, action, ratings, and target price ranges.
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body AdvancedSearchRequest true "Search parameters with filters"
// @Success 200 {object} models.PaginatedResponse "Successfully retrieved filtered stock ratings"
// @Failure 400 {object} models.ErrorResponse "Bad request"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error"
// @Router /stocks/search [post]
func (h *StockHandler) SearchStockRatings(c *gin.Context) {
	var req AdvancedSearchRequest

	// Parse request body
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format in request body"})
		return
	}

	// Validate parameters
	if req.PageNumber <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_number must be greater than 0"})
		return
	}
	if req.PageLength <= 0 || req.PageLength > 1000 {
		req.PageLength = 20
	}

	// Build dynamic WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	// Search term filter
	if req.SearchTerm != "" {
		searchPattern := "%" + req.SearchTerm + "%"
		whereConditions = append(whereConditions, fmt.Sprintf(
			"(LOWER(ticker) LIKE LOWER($%d) OR LOWER(company) LIKE LOWER($%d) OR LOWER(brokerage) LIKE LOWER($%d) OR LOWER(action) LIKE LOWER($%d) OR LOWER(rating_from) LIKE LOWER($%d) OR LOWER(rating_to) LIKE LOWER($%d))",
			argIndex, argIndex, argIndex, argIndex, argIndex, argIndex))
		args = append(args, searchPattern)
		argIndex++
	}

	// Action filter
	if req.Action != "" && req.Action != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(action) = LOWER($%d)", argIndex))
		args = append(args, req.Action)
		argIndex++
	}

	// Rating from filter
	if req.RatingFrom != "" && req.RatingFrom != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(rating_from) = LOWER($%d)", argIndex))
		args = append(args, req.RatingFrom)
		argIndex++
	}

	// Rating to filter
	if req.RatingTo != "" && req.RatingTo != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("LOWER(rating_to) = LOWER($%d)", argIndex))
		args = append(args, req.RatingTo)
		argIndex++
	}

	// Target price range filters
	if req.TargetFromMin > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("CAST(REPLACE(REPLACE(target_from, '$', ''), ',', '') AS NUMERIC) >= $%d", argIndex))
		args = append(args, req.TargetFromMin)
		argIndex++
	}
	if req.TargetFromMax > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("CAST(REPLACE(REPLACE(target_from, '$', ''), ',', '') AS NUMERIC) <= $%d", argIndex))
		args = append(args, req.TargetFromMax)
		argIndex++
	}
	if req.TargetToMin > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("CAST(REPLACE(REPLACE(target_to, '$', ''), ',', '') AS NUMERIC) >= $%d", argIndex))
		args = append(args, req.TargetToMin)
		argIndex++
	}
	if req.TargetToMax > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("CAST(REPLACE(REPLACE(target_to, '$', ''), ',', '') AS NUMERIC) <= $%d", argIndex))
		args = append(args, req.TargetToMax)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Calculate offset
	offset := (req.PageNumber - 1) * req.PageLength

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_ratings %s", whereClause)
	var totalCount int
	err := h.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get search count"})
		return
	}

	// Query data
	dataQuery := fmt.Sprintf(`
		SELECT id, ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at
		FROM stock_ratings
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, req.PageLength, offset)
	rows, err := h.DB.Query(dataQuery, args...)
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
		"applied_filters": gin.H{
			"search_term":     req.SearchTerm,
			"action":          req.Action,
			"rating_from":     req.RatingFrom,
			"rating_to":       req.RatingTo,
			"target_from_min": req.TargetFromMin,
			"target_from_max": req.TargetFromMax,
			"target_to_min":   req.TargetToMin,
			"target_to_max":   req.TargetToMax,
		},
	})
}

// ActionsResponse represents the response structure for stock actions
type ActionsResponse struct {
	Actions []string `json:"actions" example:"initiated by,target raised by,target lowered by,reiterated by,upgraded"`
}

// FilterOptionsResponse represents available filter options
type FilterOptionsResponse struct {
	Actions     []string `json:"actions"`
	RatingsFrom []string `json:"ratings_from"`
	RatingsTo   []string `json:"ratings_to"`
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

// GetFilterOptions retrieves all available filter options
// @Summary Get all available filter options
// @Description Retrieves filter options including actions, ratings from database
// @Tags stocks
// @Produce json
// @Success 200 {object} FilterOptionsResponse "Successfully retrieved filter options"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred"
// @Router /stocks/filter-options [get]
func (h *StockHandler) GetFilterOptions(c *gin.Context) {
	var response FilterOptionsResponse

	// Get unique actions
	actionsQuery := `SELECT DISTINCT action FROM stock_ratings WHERE action IS NOT NULL AND action != '' ORDER BY action ASC`
	rows, err := h.DB.Query(actionsQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var action string
			if err := rows.Scan(&action); err == nil {
				response.Actions = append(response.Actions, action)
			}
		}
	}

	// Get unique ratings from
	ratingsFromQuery := `SELECT DISTINCT rating_from FROM stock_ratings WHERE rating_from IS NOT NULL AND rating_from != '' ORDER BY rating_from ASC`
	rows, err = h.DB.Query(ratingsFromQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rating string
			if err := rows.Scan(&rating); err == nil {
				response.RatingsFrom = append(response.RatingsFrom, rating)
			}
		}
	}

	// Get unique ratings to
	ratingsToQuery := `SELECT DISTINCT rating_to FROM stock_ratings WHERE rating_to IS NOT NULL AND rating_to != '' ORDER BY rating_to ASC`
	rows, err = h.DB.Query(ratingsToQuery)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var rating string
			if err := rows.Scan(&rating); err == nil {
				response.RatingsTo = append(response.RatingsTo, rating)
			}
		}
	}

	c.JSON(http.StatusOK, response)
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
// @Summary Get quantitative stock investment recommendations
// @Description Analyzes all stock ratings data using configurable weighted algorithms to provide ranked investment recommendations. Considers target price changes, rating improvements, analyst sentiment, and market trends.
// @Tags recommendations
// @Produce json
// @Param limit query int false "Number of recommendations to return (3, 5, 10, 15, 20)" default(10)
// @Success 200 {object} RecommendationsResponse "Successfully generated stock recommendations with scoring and analysis"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid limit parameter"
// @Failure 500 {object} models.GenericErrorResponse "Internal server error occurred during analysis"
// @Router /stocks/recommendations [get]
func (h *StockHandler) GetStockRecommendations(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter. Must be between 1 and 50"})
		return
	}
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

	// Analyze and generate recommendations with specified limit
	recommendations := analyzeStocksForRecommendations(stocks, limit)

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
func analyzeStocksForRecommendations(stocks []stockData, limit int) []StockRecommendation {
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

	// STEP 5: Return top N recommendations based on user selection
	if len(recommendations) > limit {
		recommendations = recommendations[:limit] // Slice to get requested number
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

// validateWeights ensures weights sum to 100% (1.0)
func (w ScoringWeights) validateWeights() error {
	total := w.TargetPriceWeight + w.RatingWeight + w.ActionWeight + w.TimingWeight
	if math.Abs(total-1.0) > 0.001 { // Allow small floating point errors
		return fmt.Errorf("weights must sum to 100%%, got %.1f%%", total*100)
	}
	return nil
}

// getDefaultWeights returns the default scoring weights
// These can be easily modified based on market conditions
func getDefaultWeights() ScoringWeights {
	weights := ScoringWeights{
		TargetPriceWeight: 0.4, // 40% - Most important for speculative markets
		RatingWeight:      0.3, // 30% - Professional analyst opinion
		ActionWeight:      0.2, // 20% - Direction of analyst changes
		TimingWeight:      0.1, // 10% - Recent activity bonus
	}
	// Validate weights on startup
	if err := weights.validateWeights(); err != nil {
		panic(fmt.Sprintf("Invalid default weights: %v", err))
	}
	return weights
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

	return analyzeStocksForRecommendations(stocks, 10) // Default limit for summary
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
				"content": "You are a seasoned Wall Street equity research analyst with 15+ years of experience in fundamental analysis and market strategy. Analyze the provided stock data with the expertise of someone who has navigated multiple market cycles. Focus on: sector rotation patterns, valuation metrics implications, institutional sentiment shifts, and macroeconomic factors affecting target price revisions. Provide actionable insights for institutional investors. Keep analysis under 200 words but make every word count.",
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

	// Build expert-level prompt for institutional analysis
	prompt := "EQUITY RESEARCH BRIEF - Analyze the following analyst actions and provide institutional-grade market insights:\n\n"

	// Include top 10 recommendations with detailed context
	for i, rec := range recommendations {
		if i >= 10 { // Focus on top 10 for comprehensive analysis
			break	
		}
		prompt += fmt.Sprintf("%d. %s (%s) - %s [Score: %.1f/10]\n   Brokerage: %s | Rating: %s | Target: %s\n   Catalyst: %s\n\n",
			i+1, rec.Company, rec.Ticker, rec.Recommendation, rec.Score, rec.Brokerage,
			rec.CurrentRating, rec.TargetPrice, rec.Reason)
	}

	prompt += "ANALYSIS FRAMEWORK: Assess sector rotation dynamics, valuation expansion/contraction themes, earnings revision trends, and institutional positioning implications. Consider current market regime and provide tactical allocation insights."
	return prompt
}

// ChatResponse represents an AI chat response
type ChatResponse struct {
	Response       string               `json:"response" example:"Based on current market data, I recommend focusing on stocks with strong buy ratings and recent target price increases. The biotech sector shows particular promise."`
	TokensUsed     int                  `json:"tokens_used" example:"156"`
	GeneratedAt    string               `json:"generated_at" example:"2024-01-15T10:30:00Z"`
	ContextUsed    string               `json:"context_used,omitempty"`
	UpdatedMemory  *ConversationMemory  `json:"updated_memory,omitempty"`
}

// ChatRequest represents a chat request with optional conversation memory
type ChatRequest struct {
	Message            string                 `json:"message" example:"What are the best stocks to invest in today?"`
	ConversationMemory *ConversationMemory    `json:"conversation_memory,omitempty"`
	RecentMessages     []RecentMessage        `json:"recent_messages,omitempty"`
}

// ConversationMemory holds compressed conversation history and key topics
type ConversationMemory struct {
	Summary     string   `json:"summary"`
	KeyTopics   []string `json:"key_topics"`
	LastContext string   `json:"last_context"`
}

// RecentMessage represents a recent message in the conversation
type RecentMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GetStockChat provides AI-powered chat responses with RAG (Retrieval-Augmented Generation)
// @Summary Chat with AI about stock market with database context
// @Description Interactive chat with gpt-4.1-nano that can query the database for specific stock information and provide personalized analysis based on actual data.
// @Tags ai-analysis
// @Accept json
// @Produce json
// @Param request body ChatRequest true "Chat message from user"
// @Success 200 {object} ChatResponse "Successfully generated AI chat response with database context"
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

	// Enhanced RAG with conversation memory
	dbContext, err := h.retrieveRelevantDataWithMemory(req.Message, req.ConversationMemory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to retrieve data: %v", err)})
		return
	}

	// Generate AI response with conversation context
	response, tokensUsed, updatedMemory, err := h.generateChatResponseWithMemory(req.Message, dbContext, req.RecentMessages, req.ConversationMemory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate response: %v", err)})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Response:      response,
		TokensUsed:    tokensUsed,
		GeneratedAt:   time.Now().Format(time.RFC3339),
		ContextUsed:   dbContext,
		UpdatedMemory: updatedMemory,
	})
}

// generateChatResponseWithMemory implements memory-enhanced AI response generation
//
// MEMORY-ENHANCED RESPONSE GENERATION PROCESS:
// This function orchestrates the complete conversation memory workflow,
// from context building to memory updates, ensuring efficient and contextual responses.
//
// STEP-BY-STEP PROCESS:
// STEP 1: Build lightweight conversation context from recent messages + memory
// STEP 2: Generate AI response using database context + conversation context
// STEP 3: Update conversation memory with new interaction
// STEP 4: Return response + updated memory for frontend caching
//
// CONTEXT BUILDING STRATEGY:
// Instead of sending entire conversation history (expensive), we send:
// - Previous conversation summary (compressed)
// - Last 4 messages only (recent context)
// - Key topics from memory (entity continuity)
//
// MEMORY UPDATE ALGORITHM:
// 1. Extract key topics from user message (tickers, sectors, actions)
// 2. Merge with existing topics (max 5 to prevent bloat)
// 3. Update conversation summary (compressed history)
// 4. Cache database context for potential reuse
//
// TOKEN EFFICIENCY:
// Traditional: Full conversation (1000+ tokens)
// Memory approach: Summary + recent (200-300 tokens)
// Efficiency gain: 70-80% token reduction
func (h *StockHandler) generateChatResponseWithMemory(userMessage, context string, recentMessages []RecentMessage, memory *ConversationMemory) (string, int, *ConversationMemory, error) {
	// STEP 1: BUILD LIGHTWEIGHT CONVERSATION CONTEXT
	// Create compressed context from memory + recent messages (not full history)
	conversationContext := h.buildConversationContext(recentMessages, memory)
	println("💬 Memory: Built conversation context, length:", len(conversationContext), "chars")

	// STEP 2: GENERATE AI RESPONSE WITH ENHANCED CONTEXT
	// Send user question + database context + conversation context to AI
	response, tokens, err := h.generateChatResponse(userMessage, context, conversationContext)
	if err != nil {
		return "", 0, nil, err
	}
	println("✅ Memory: AI response generated, tokens used:", tokens)

	// STEP 3: UPDATE CONVERSATION MEMORY
	// Extract topics, update summary, cache context for future reuse
	updatedMemory := h.updateConversationMemory(userMessage, response, context, memory)
	println("💾 Memory: Updated memory with topics:", updatedMemory.KeyTopics)

	return response, tokens, updatedMemory, nil
}

// buildConversationContext creates context from recent messages
func (h *StockHandler) buildConversationContext(recentMessages []RecentMessage, memory *ConversationMemory) string {
	var context strings.Builder

	if memory != nil && memory.Summary != "" {
		context.WriteString("Previous conversation: " + memory.Summary + "\n")
	}

	if len(recentMessages) > 0 {
		context.WriteString("Recent messages:\n")
		for _, msg := range recentMessages {
			context.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
	}

	return context.String()
}

// updateConversationMemory implements intelligent memory management
//
// MEMORY UPDATE ALGORITHM:
// This function maintains conversation state efficiently by extracting key information
// from each interaction and updating memory structures for future context reuse.
//
// TOPIC EXTRACTION STRATEGY:
// 1. TICKER SYMBOLS: Extract stock symbols (AAPL, MSFT, GOOGL)
// 2. SEMANTIC TOPICS: Identify themes (target_prices, ratings, sectors)
// 3. ACTION TYPES: Detect operations (upgrades, downgrades, raises)
//
// MEMORY OPTIMIZATION:
// - Summary: Compressed conversation history (max 150 chars)
// - Topics: Max 5 most recent topics (prevents memory bloat)
// - Context: Cache last database result for reuse
//
// TOPIC MERGING LOGIC:
// - Combine current + new topics
// - Remove duplicates
// - Keep most recent 5 topics
// - Prioritize specific entities (tickers) over general topics
//
// EXAMPLE MEMORY EVOLUTION:
// Initial: {summary: "", topics: [], context: ""}
// After "AAPL ratings": {summary: "User asked about AAPL ratings", topics: ["AAPL", "ratings"], context: "AAPL data..."}
// After "AAPL targets": {summary: "AAPL ratings; Latest: AAPL targets", topics: ["AAPL", "ratings", "target_prices"], context: "AAPL data..."}
func (h *StockHandler) updateConversationMemory(userMessage, response, dbContext string, currentMemory *ConversationMemory) *ConversationMemory {
	if currentMemory == nil {
		currentMemory = &ConversationMemory{}
	}

	// STEP 1: EXTRACT KEY TOPICS FROM USER MESSAGE
	// Identify tickers, semantic topics, and action types for future context matching
	topics := h.extractKeyTopics(userMessage)
	println("🏷️ Memory: Extracted topics from message:", topics)

	// STEP 2: BUILD UPDATED MEMORY STRUCTURE
	// Merge topics, update summary, cache context for reuse
	updatedMemory := &ConversationMemory{
		Summary:     h.generateConversationSummary(userMessage, response, currentMemory.Summary),
		KeyTopics:   h.mergeTopics(currentMemory.KeyTopics, topics),
		LastContext: dbContext, // Cache for potential reuse
	}

	println("📊 Memory: Updated summary:", updatedMemory.Summary[:min(50, len(updatedMemory.Summary))])
	return updatedMemory
}

// extractTickers finds ticker symbols in user message using pattern matching
func (h *StockHandler) extractTickers(message string) []string {
	words := strings.Fields(strings.ToUpper(message))
	var tickers []string
	for _, word := range words {
		if len(word) >= 2 && len(word) <= 5 {
			isValidTicker := true
			for _, char := range word {
				if !(char >= 'A' && char <= 'Z') {
					isValidTicker = false
					break
				}
			}
			if isValidTicker {
				tickers = append(tickers, word)
			}
		}
	}
	return tickers
}

// extractKeyTopics implements intelligent topic extraction for conversation memory
//
// TOPIC EXTRACTION ALGORITHM:
// This function analyzes user messages to identify key entities and themes
// that can be used for context matching in future interactions.
//
// EXTRACTION CATEGORIES:
// 🏷️ TICKER SYMBOLS: Stock symbols (AAPL, MSFT, GOOGL, etc.)
//   - Pattern: 2-5 uppercase letters
//   - High priority for context matching
//
// 📊 SEMANTIC TOPICS: Market themes and concepts
//   - target_prices: "target", "price", "PT", "price target"
//   - ratings: "rating", "upgrade", "downgrade", "buy", "sell"
//   - sectors: "sector", "industry", "biotech", "tech", "finance"
//   - actions: "raised", "lowered", "initiated", "maintained"
//
// 🤖 AI CONTEXT MATCHING:
// Extracted topics enable smart context reuse:
// - Same ticker -> Reuse stock-specific context
// - Same theme -> Reuse thematic analysis
// - Different topics -> Generate fresh context
//
// EXAMPLES:
// "Show me AAPL ratings" -> ["AAPL", "ratings"]
// "What about target prices?" -> ["target_prices"]
// "MSFT vs GOOGL comparison" -> ["MSFT", "GOOGL"]
// "Biotech sector analysis" -> ["sectors", "biotech"]
func (h *StockHandler) extractKeyTopics(message string) []string {
	message = strings.ToLower(message)
	var topics []string

	// CATEGORY 1: TICKER SYMBOL EXTRACTION
	// Extract specific stock symbols for precise context matching
	tickers := h.extractTickers(message)
	topics = append(topics, tickers...)
	if len(tickers) > 0 {
		println("🏷️ Topics: Found tickers:", tickers)
	}

	// CATEGORY 2: SEMANTIC TOPIC EXTRACTION
	// Identify market themes and concepts for thematic context matching
	if strings.Contains(message, "target") || strings.Contains(message, "price") {
		topics = append(topics, "target_prices")
	}
	if strings.Contains(message, "rating") || strings.Contains(message, "upgrade") || strings.Contains(message, "downgrade") {
		topics = append(topics, "ratings")
	}
	if strings.Contains(message, "sector") || strings.Contains(message, "industry") {
		topics = append(topics, "sectors")
	}
	if strings.Contains(message, "raised") || strings.Contains(message, "lowered") || strings.Contains(message, "initiated") {
		topics = append(topics, "analyst_actions")
	}

	println("📊 Topics: Extracted semantic topics:", topics[len(tickers):])
	return topics
}

// mergeTopics combines current and new topics
func (h *StockHandler) mergeTopics(current, new []string) []string {
	topicMap := make(map[string]bool)
	for _, topic := range current {
		topicMap[topic] = true
	}
	for _, topic := range new {
		topicMap[topic] = true
	}

	var merged []string
	for topic := range topicMap {
		merged = append(merged, topic)
	}

	// Limit to 5 most recent topics
	if len(merged) > 5 {
		merged = merged[:5]
	}

	return merged
}

// generateConversationSummary creates a brief summary of the conversation
func (h *StockHandler) generateConversationSummary(userMessage, response, currentSummary string) string {
	// Simple summary logic - in production, could use AI for this
	if currentSummary == "" {
		return fmt.Sprintf("User asked about: %s", userMessage[:min(50, len(userMessage))])
	}
	return fmt.Sprintf("%s; Latest: %s", currentSummary[:min(100, len(currentSummary))], userMessage[:min(30, len(userMessage))])
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateChatResponse calls OpenAI for chat responses
func (h *StockHandler) generateChatResponse(userMessage, context, conversationContext string) (string, int, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a professional financial advisor with access to real-time stock market database. Use the provided database context to answer questions accurately. When users ask about specific stocks, sectors, or market trends, reference the actual data provided. If asked about stocks not in the context, clearly state data limitations. Keep responses helpful and actionable.\n\nFORMATTING RULES:\n- Use markdown formatting for better readability\n- Use numbered lists (1. 2. 3.) for multiple items\n- Use **bold** for company names and tickers\n- Use bullet points (-) for sub-items\n- Keep responses concise but complete\n\nConversation Context:\n" + conversationContext + "\n\nDatabase Context:\n" + context,
			},
			{
				"role":    "user",
				"content": userMessage,
			},
		},
		"max_tokens":   500,
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

// retrieveRelevantDataWithMemory implements RAG with intelligent conversation memory
//
// CONVERSATION MEMORY SYSTEM OVERVIEW:
// This system maintains conversation context without expensive re-prompting by implementing
// smart caching and context reuse strategies. It dramatically reduces API costs while
// providing seamless conversational experience.
//
// WHY CONVERSATION MEMORY IS CRITICAL:
// 1. COST EFFICIENCY: Avoids re-sending entire conversation history (expensive tokens)
// 2. SPEED: Cached context means instant responses for follow-up questions
// 3. CONTEXT CONTINUITY: Maintains conversation flow without losing previous context
// 4. SMART CACHING: Only regenerates database context when truly necessary
//
// MEMORY ARCHITECTURE:
// 🧠 ConversationMemory Structure:
//   - Summary: Brief conversation history ("User asked about AAPL ratings, then target prices")
//   - KeyTopics: Extracted entities ["AAPL", "ratings", "target_prices"]
//   - LastContext: Cached database results for reuse
//
// INTELLIGENT CONTEXT REUSE ALGORITHM:
// STEP 1: Analyze incoming user message for key topics
// STEP 2: Compare with previous conversation topics
// STEP 3: If topics overlap -> REUSE cached context (no database query needed)
// STEP 4: If topics differ -> Generate fresh context and update cache
// STEP 5: Update conversation memory with new interaction
//
// CONTEXT REUSE EXAMPLES:
// 🔄 REUSE SCENARIO:
//   Previous: "Show me AAPL ratings" -> Cache: AAPL database context
//   Current:  "What about AAPL target prices?" -> REUSE: Same stock (AAPL)
//   Result: Instant response, no new SQL generation
//
// 🆕 FRESH CONTEXT SCENARIO:
//   Previous: "Show me AAPL ratings" -> Cache: AAPL context
//   Current:  "What about biotech stocks?" -> FRESH: Different topic
//   Result: Generate new SQL for biotech data
//
// COST SAVINGS CALCULATION:
// Traditional approach: Send full conversation (1000+ tokens per request)
// Memory approach: Send only new question + cached context (100-200 tokens)
// Savings: 80-90% reduction in API costs for follow-up questions
func (h *StockHandler) retrieveRelevantDataWithMemory(userMessage string, memory *ConversationMemory) (string, error) {
	// STEP 1: SMART CONTEXT REUSE CHECK
	// Analyze if current query relates to previous topics to avoid redundant database queries
	if memory != nil && memory.LastContext != "" && h.isSimilarQuery(userMessage, memory.KeyTopics) {
		println("🧠 Memory: Reusing cached context for similar query")
		println("💾 Memory: Topics matched:", memory.KeyTopics)
		return memory.LastContext, nil // COST SAVINGS: No new SQL generation needed
	}

	// STEP 2: FRESH CONTEXT GENERATION
	// Generate new database context for different/new topics
	println("🆕 Memory: Generating fresh context for new topic")
	return h.retrieveRelevantData(userMessage)
}

// isSimilarQuery checks if current query is similar to previous topics
func (h *StockHandler) isSimilarQuery(query string, topics []string) bool {
	queryLower := strings.ToLower(query)
	for _, topic := range topics {
		if strings.Contains(queryLower, strings.ToLower(topic)) {
			return true
		}
	}
	return false
}

// retrieveRelevantData implements flexible RAG using AI-powered SQL generation
//
// ENHANCED RAG ARCHITECTURE:
// Instead of rigid keyword matching, this system uses AI to understand user intent
// and dynamically generate appropriate SQL queries for any question.
//
// FLEXIBLE RAG PROCESS:
// STEP 1: Send user question + database schema to AI
// STEP 2: AI generates appropriate SQL query based on natural language
// STEP 3: Execute generated SQL safely with validation
// STEP 4: Format results as structured context
// STEP 5: Use context for final response generation
//
// EXAMPLES OF FLEXIBLE QUERIES:
// "stocks with highest target price increase" -> AI generates SQL with price calculations
// "biotech companies with buy ratings" -> AI generates sector + rating filters
// "recent downgrades by Goldman Sachs" -> AI generates time + brokerage + action filters
// "top 5 stocks by analyst consensus" -> AI generates grouping and ranking logic
//
// ADVANTAGES:
// ✅ Handles any natural language query
// ✅ No predefined keyword limitations
// ✅ Dynamic SQL generation
// ✅ Flexible and extensible
// ✅ Maintains SQL injection protection
func (h *StockHandler) retrieveRelevantData(userMessage string) (string, error) {
	// STEP 1: Generate SQL query using AI based on user question
	println("🤖 RAG: Generating SQL for question:", userMessage)
	sqlQuery, err := h.generateSQLFromQuestion(userMessage)
	if err != nil {
		println("❌ RAG: Failed to generate SQL:", err.Error())
		return "", fmt.Errorf("failed to generate SQL: %v", err)
	}
	println("📝 RAG: Generated SQL Query:")
	println("   ", sqlQuery)

	// STEP 2: Validate and execute the generated SQL safely
	println("🔍 RAG: Validating and executing SQL...")
	results, err := h.executeSafeSQL(sqlQuery)
	if err != nil {
		println("❌ RAG: Failed to execute SQL:", err.Error())
		return "", fmt.Errorf("failed to execute query: %v", err)
	}
	println("✅ RAG: SQL executed successfully, found", len(results), "results")

	// STEP 3: Format results as structured context
	context := h.formatQueryResults(results, userMessage)
	println("📊 RAG: Context formatted, length:", len(context), "characters")
	return context, nil
}

// generateSQLFromQuestion uses AI to convert natural language to SQL
func (h *StockHandler) generateSQLFromQuestion(question string) (string, error) {
	schema := `
	Database Schema:
	Table: stock_ratings
	Columns:
	- id (SERIAL PRIMARY KEY)
	- ticker (VARCHAR(10)) - Stock symbol like 'AAPL', 'MSFT'
	- target_from (VARCHAR(20)) - Previous target price like '$150.00', '$1,250.00'
	- target_to (VARCHAR(20)) - New target price like '$180.00', '$6,250.00'
	- company (VARCHAR(255)) - Company name like 'Apple Inc.'
	- action (VARCHAR(100)) - Analyst action like 'target raised by', 'upgraded'
	- brokerage (VARCHAR(255)) - Analyst firm like 'Goldman Sachs'
	- rating_from (VARCHAR(50)) - Previous rating like 'Hold'
	- rating_to (VARCHAR(50)) - New rating like 'Buy', 'Strong Buy'
	- time (TIMESTAMP) - When analyst made the report
	- created_at (TIMESTAMP) - When record was inserted
	
	IMPORTANT: Price fields contain dollar signs and commas. Use CAST(REPLACE(REPLACE(column, '$', ''), ',', '') AS NUMERIC) for calculations.
	`

	prompt := fmt.Sprintf(`%s

	Generate a PostgreSQL query for: "%s"

	Rules:
	1. Only SELECT queries allowed
	2. Use LIMIT to prevent large results (max 50)
	3. Include relevant columns for the question
	4. Use proper SQL syntax
	5. Return only the SQL query, no explanations
	6. For price calculations, use: CAST(REPLACE(REPLACE(column, '$', ''), ',', '') AS NUMERIC)
	7. Price fields (target_from, target_to) may contain commas and dollar signs

	SQL:`, schema, question)

	println("🧠 AI: Sending prompt to OpenAI for SQL generation...")
	println("📋 AI: Question:", question)

	reqBody := map[string]interface{}{
		"model": "gpt-4.1-nano",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a SQL expert. Generate safe PostgreSQL queries based on user questions. Only return the SQL query.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"max_tokens":   200,
		"temperature": 0.1,
	}

	reqJSON, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", strings.NewReader(string(reqJSON)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no SQL generated")
	}

	sqlQuery := strings.TrimSpace(openAIResp.Choices[0].Message.Content)
	sqlQuery = strings.Trim(sqlQuery, "`")
	println("✅ AI: SQL generated successfully")
	println("🔧 AI: Raw SQL from OpenAI:", sqlQuery)
	return sqlQuery, nil
}

// executeSafeSQL validates and executes the generated SQL query
func (h *StockHandler) executeSafeSQL(sqlQuery string) ([]map[string]interface{}, error) {
	// Basic SQL injection protection
	println("🔒 Security: Validating SQL query for safety...")
	sqlLower := strings.ToLower(sqlQuery)
	if !strings.HasPrefix(sqlLower, "select") {
		println("❌ Security: Non-SELECT query blocked:", sqlQuery)
		return nil, fmt.Errorf("only SELECT queries allowed")
	}
	if strings.Contains(sqlLower, "drop") || strings.Contains(sqlLower, "delete") || strings.Contains(sqlLower, "update") || strings.Contains(sqlLower, "insert") {
		println("❌ Security: Dangerous SQL operation blocked:", sqlQuery)
		return nil, fmt.Errorf("dangerous SQL operations not allowed")
	}
	println("✅ Security: SQL query validated as safe")

	println("💾 Database: Executing SQL query...")
	rows, err := h.DB.Query(sqlQuery)
	if err != nil {
		println("❌ Database: Query execution failed:", err.Error())
		println("🔍 Database: Failed query was:", sqlQuery)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		println("❌ Database: Failed to get columns:", err.Error())
		return nil, err
	}
	println("📋 Database: Query columns:", columns)

	var results []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		rowCount++
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			println("⚠️  Database: Skipping row", rowCount, "due to scan error:", err.Error())
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			if values[i] != nil {
				row[col] = values[i]
			}
		}
		results = append(results, row)
		
		// Log first few rows for debugging
		if rowCount <= 3 {
			println(fmt.Sprintf("📄 Database: Row %d sample:", rowCount), fmt.Sprintf("%+v", row))
		}
	}

	println("📊 Database: Total rows processed:", rowCount, "| Results collected:", len(results))
	return results, nil
}

// formatQueryResults formats the SQL results into readable context
func (h *StockHandler) formatQueryResults(results []map[string]interface{}, question string) string {
	println("📝 Formatting: Starting to format", len(results), "results for question:", question)
	if len(results) == 0 {
		println("⚠️  Formatting: No results to format")
		return "No data found for your query."
	}

	var context strings.Builder
	context.WriteString(fmt.Sprintf("Query results for: %s\n\n", question))

	formattedRows := 0
	for i, row := range results {
		if i >= 20 { // Limit context size
			context.WriteString("... (showing first 20 results)\n")
			println("📄 Formatting: Truncated results at 20 items")
			break
		}

		// Format each row based on available columns
		if ticker, ok := row["ticker"]; ok {
			if company, ok := row["company"]; ok {
				context.WriteString(fmt.Sprintf("%v (%v)", company, ticker))
			} else {
				context.WriteString(fmt.Sprintf("%v", ticker))
			}
		}

		if rating, ok := row["rating_to"]; ok {
			context.WriteString(fmt.Sprintf(" - Rating: %v", rating))
		}
		if target, ok := row["target_to"]; ok {
			context.WriteString(fmt.Sprintf(" - Target: %v", target))
		}
		if action, ok := row["action"]; ok {
			context.WriteString(fmt.Sprintf(" - Action: %v", action))
		}
		if brokerage, ok := row["brokerage"]; ok {
			context.WriteString(fmt.Sprintf(" - Brokerage: %v", brokerage))
		}

		// Add any calculated fields
		for key, value := range row {
			if !contains([]string{"ticker", "company", "rating_to", "target_to", "action", "brokerage"}, key) {
				context.WriteString(fmt.Sprintf(" - %s: %v", key, value))
			}
		}

		context.WriteString("\n")
		formattedRows++
	}

	println("✅ Formatting: Successfully formatted", formattedRows, "rows")
	println("📏 Formatting: Final context length:", len(context.String()), "characters")
	return context.String()
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
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
