package main

import (
	"database/sql"
	"log"
	"os"
	"smart-stock-recommender/database"
	"smart-stock-recommender/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	// API Routes from the Go Server
	api := r.Group("/api")
	{
		api.POST("/stocks", stockHandler.GetStocksByPage)
	}

	// define the port to run the server on
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Server starting on port %s", port)
	r.Run(":" + port)
}

// createTables creates the necessary tables in the database if they do not exist.
func createTables(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS stock_ratings (
		id SERIAL PRIMARY KEY,
		ticker VARCHAR(10) NOT NULL,
		target_from VARCHAR(20),
		target_to VARCHAR(20),
		company VARCHAR(255),
		action VARCHAR(100),
		brokerage VARCHAR(255),
		rating_from VARCHAR(50),
		rating_to VARCHAR(50),
		time TIMESTAMP,
		created_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(ticker, time)
	)`

	if _, err := db.Exec(query); err != nil {
		log.Fatal("Failed to create table:", err)
	}
}
