// @title Smart Stock Recommender API
// @version 1.0
// @description API for fetching and managing stock ratings data
// @host localhost:8081
// @BasePath /api
package main

import (
	"database/sql"
	"log"
	"os"
	"smart-stock-recommender/database"
	_ "smart-stock-recommender/docs"
	"smart-stock-recommender/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// main is the entry point of the application.
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to database
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create tables
	createTables(db)

	// Initialize handlers
	stockHandler := handlers.NewStockHandler(db)

	// Setup router
	// gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API Routes from the Go Server
	api := r.Group("/api")
	{
		api.POST("/stocks", stockHandler.GetStocksByPage)
		api.POST("/stocks/bulk", stockHandler.GetStocksBulk)
		api.POST("/stocks/list", stockHandler.GetStockRatings)
		api.POST("/stocks/search", stockHandler.SearchStockRatings)
		api.GET("/stocks/actions", stockHandler.GetStockActions)
		api.GET("/stocks/recommendations", stockHandler.GetStockRecommendations)
		api.GET("/stocks/summary", stockHandler.GetStockSummary)
		api.POST("/stocks/chat", stockHandler.GetStockChat)
		api.GET("/stocks/metrics", stockHandler.GetStockMetrics)
	}

	// define the port to run the server on
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Start server
	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}

// createTables creates the necessary tables in the database if they do not exist.
func createTables(db *sql.DB) {
	// Query to create stock_ratings table
	query := `
	CREATE TABLE IF NOT EXISTS stock_ratings (
		id SERIAL PRIMARY KEY,
		ticker VARCHAR(10) NOT NULL,
		target_from VARCHAR(20) NOT NULL,
		target_to VARCHAR(20) NOT NULL,
		company VARCHAR(255) NOT NULL,
		action VARCHAR(100) NOT NULL,
		brokerage VARCHAR(255) NOT NULL,
		rating_from VARCHAR(50),
		rating_to VARCHAR(50),
		time TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(ticker, brokerage, action, rating_from, rating_to, time)
	)`

	// Execute the query
	if _, err := db.Exec(query); err != nil {
		log.Fatal("Failed to create table:", err)
	}
}
