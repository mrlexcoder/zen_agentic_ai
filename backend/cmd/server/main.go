package main

import (
	"log"
	"trading-system/backend/internal/config"
	"trading-system/backend/internal/handlers"
	"trading_system/backend/internal/middleware"
	"trading-system/backend/internal/models"
	"trading-system/backend/internal/services/ai_brain"
	"trading_system/backend/internal/services/backtest"
	"trading_system/backend/internal/services/binance"
	"trading-system/backend/internal/services/pattern"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := models.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	binanceService := binance.NewService(cfg.BinanceAPIKey, cfg.BinanceSecretKey)
	patternService := pattern.NewService()
	aiBrainService := ai_brain.NewService(patternService)
	backtestService := backtest.NewService(binanceService, aiBrainService)

	// Initialize handlers
	h := handlers.NewHandler(binanceService, aiBrainService, backtestService, patternService, db)

	// Start cron jobs
	c := cron.New()
	c.AddFunc("@every 1m", func() {
		aiBrainService.UpdatePredictions(binanceService)
	})
	c.Start()

	// Setup Gin router
	r := gin.Default()
	r.Use(middleware.CORS())

	// WebSocket endpoint for real-time data
	r.GET("/ws", h.HandleWebSocket)

	// API routes
	api := r.Group("/api/v1")
	{
		// Market data
		api.GET("/ticker/:symbol", h.GetTicker)
		api.GET("/klines", h.GetKlines)
		api.GET("/orderbook/:symbol", h.GetOrderBook)

		// AI Brain
		api.GET("/ai/prediction/:symbol", h.GetPrediction)
		api.GET("/ai/analysis/:symbol", h.GetAnalysis)
		api.POST("/ai/train", h.TrainModel)

		// Backtesting
		api.POST("/backtest", h.RunBacktest)
		api.GET("/backtest/results", h.GetBacktestResults)

		// Patterns
		api.GET("/patterns/:symbol", h.GetPatterns)
		api.GET("/patterns/learned", h.GetLearnedPatterns)

		// Trading (if enabled)
		api.POST("/trade/order", h.PlaceOrder)
		api.GET("/trade/balance", h.GetBalance)
	}

	// Serve frontend
	r.Static("/static", "./frontend/dist")

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}