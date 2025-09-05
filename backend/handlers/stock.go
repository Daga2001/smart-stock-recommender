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
// @Description Retrieves stock data from external API for a specific page and stores in database
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.PageRequest true "Page number to fetch"
// @Success 200 {object} models.ApiResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
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
// @Summary Fetch stocks in bulk for page range
// @Description Retrieves stock data from external API for a range of pages, clears existing data first
// @Tags stocks
// @Accept json
// @Produce json
// @Param request body models.BulkPageRequest true "Page range to fetch"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
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

	// Prevent excessive page ranges
	if req.EndPage-req.StartPage > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page range too large (max 100 pages)"})
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

// fetchStocksFromAPI fetches stock data from the external API for a given page.
func (h *StockHandler) fetchStocksFromAPI(page int) ([]models.StockRatings, error) {
	// fetch with retry logic to ensure we get some data
	// try up to 5 times with different page numbers
	return h.fetchStocksFromAPIWithRetry(page, 5)
}

// fetchStocksFromAPIWithRetry tries different page numbers until it gets 10 items
func (h *StockHandler) fetchStocksFromAPIWithRetry(originalPage, maxRetries int) ([]models.StockRatings, error) {
	// HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Try original page first, then use a better pattern
		tryPage := originalPage
		if attempt > 0 {
			// Prime number for better distribution
			tryPage = originalPage + attempt*13 
		}

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

		var apiResp models.ApiResponse
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		resp.Body.Close()

		if err != nil {
			continue
		}

		// Accept any data (not just 10 items) to reduce API calls
		if len(apiResp.Items) > 0 {
			println("✓ Page", originalPage, "-> found", len(apiResp.Items), "items at page", tryPage)
			return apiResp.Items, nil
		}
	}

	println("✗ Page", originalPage, "-> no data found after", maxRetries, "attempts")
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
	type result struct {
		stocks []models.StockRatings
		page   int
		err    error
	}

	// Calculate total pages to fetch
	pageCount := endPage - startPage + 1

	// Channel to collect results
	results := make(chan result, pageCount)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Rate limiter: max 20 concurrent requests for faster processing
	semaphore := make(chan struct{}, 20)

	// Fetch pages with rate limiting
	for page := startPage; page <= endPage; page++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			stocks, err := h.fetchStocksFromAPI(p)
			results <- result{stocks: stocks, page: p, err: err}
		}(page)
	}

	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allStocks []models.StockRatings
	totalFetched := 0
	pagesWithData := 0

	// for each result, check for errors and aggregate stocks
	for res := range results {
		if res.err != nil {
			return nil, 0, fmt.Errorf("failed to fetch page %d: %v", res.page, res.err)
		}
		if len(res.stocks) > 0 {
			allStocks = append(allStocks, res.stocks...)
			totalFetched += len(res.stocks)
			pagesWithData++
		}
	}

	println("Summary:", pagesWithData, "pages with data out of", pageCount, "total pages")

	if err := h.batchInsertStocks(allStocks); err != nil {
		return nil, 0, fmt.Errorf("failed to insert stocks: %v", err)
	}

	return allStocks, totalFetched, nil
}

// batchInsertStocks inserts multiple stock records into the database in a single transaction.
func (h *StockHandler) batchInsertStocks(stocks []models.StockRatings) error {
	// Validate if there are stocks to insert
	if len(stocks) == 0 {
		return nil
	}

	// Begin a transaction
	tx, err := h.DB.Begin()
	if err != nil {
		return err
	}

	// Ensure to rollback the transaction in case of an error
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO stock_ratings (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (ticker, time) DO NOTHING`)
	if err != nil {
		return err
	}

	// Close the statement when done
	defer stmt.Close()

	// Execute the statement for each stock
	for _, stock := range stocks {
		_, err := stmt.Exec(
			stock.Ticker, stock.TargetFrom, stock.TargetTo, stock.Company,
			stock.Action, stock.Brokerage, stock.RatingFrom, stock.RatingTo,
			stock.Time, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

/*
storeStock saves a stock entry into the database.
*/
func (h *StockHandler) storeStock(stock models.StockRatings) error {
	// Let's make the query
	// ✅ SAFE - Uses parameterized query
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
