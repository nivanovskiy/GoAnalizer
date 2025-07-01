package main

import (
        "log"
        "net/http"
        "os"

        "github.com/gin-gonic/gin"
        "github.com/performance-analyzer/config"
        "github.com/performance-analyzer/database"
        "github.com/performance-analyzer/handlers"
        "github.com/performance-analyzer/middleware"
        "github.com/performance-analyzer/services"
)

func main() {
        // Initialize configuration
        cfg := config.New()

        // Initialize database connection
        db, err := database.Connect(cfg.GetDatabaseURL())
        if err != nil {
                log.Fatalf("Failed to connect to database: %v", err)
        }
        defer db.Close()

        // Run database migrations
        if err := database.Migrate(db); err != nil {
                log.Fatalf("Failed to run migrations: %v", err)
        }

        // Initialize services
        aiClient := services.NewAIClient(cfg.GetAIModelURL())
        analyzer := services.NewAnalyzer(db, aiClient)

        // Start background analyzer
        go analyzer.StartBackgroundProcessor()

        // Initialize handlers
        handler := handlers.New(db, analyzer)

        // Setup Gin router with detailed logging
        router := gin.New()
        
        // Add detailed HTTP request/response logging
        router.Use(middleware.DetailedHTTPLogger())
        router.Use(gin.Recovery())

        // API routes
        api := router.Group("/")
        {
                api.POST("/initAnalize/:tenant/:repo/:uuid", handler.InitAnalyze)
                api.POST("/sendFile/:uuid", handler.SendFile)
                api.POST("/sendResults/:uuid", handler.SendResults)
                api.GET("/getAnalizeResults/:uuid", handler.GetAnalyzeResults)
        }

        // Root endpoint with API documentation
        router.GET("/", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{
                        "service": "Performance Analyzer API",
                        "version": "1.0.0",
                        "status": "running",
                        "endpoints": gin.H{
                                "POST /initAnalize/{tenant}/{repo}/{uuid}": "Initialize analysis pipeline",
                                "POST /sendFile/{uuid}":                    "Upload project file for analysis",
                                "POST /sendResults/{uuid}":                 "Submit performance test results",
                                "GET /getAnalizeResults/{uuid}":            "Get analysis results",
                                "GET /health":                              "Health check",
                        },
                        "description": "REST API for performance testing analysis with AI-powered insights",
                })
        })

        // Health check endpoint
        router.GET("/health", func(c *gin.Context) {
                c.JSON(http.StatusOK, gin.H{"status": "healthy"})
        })

        // Start server
        port := os.Getenv("PORT")
        if port == "" {
                port = "5000"
        }

        log.Printf("Starting server on port %s with detailed HTTP logging enabled", port)
        if err := router.Run("0.0.0.0:" + port); err != nil {
                log.Fatalf("Failed to start server: %v", err)
        }
}
