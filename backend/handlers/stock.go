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
