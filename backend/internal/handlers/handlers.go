package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"zen_agentic_ai/internal/services/ai_brain"
	"zen_agentic_ai/internal/services/backtest"
	"zen_agentic_ai/internal/services/binance"
	"zen_agentic_ai/internal/services/pattern"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Handler struct {
	binanceService  *binance.Service
	aiBrainService  *ai_brain.Service
	backtestService *backtest.Service
	patternService  *pattern.Service
	db              *gorm.DB
	upgrader        websocket.Upgrader
}

func NewHandler(binance *binance.Service, aiBrain *ai_brain.Service, backtest *backtest.Service, pattern *pattern.Service, db *gorm.DB) *Handler {
	binance.SetDB(db)
	aiBrain.SetDB(db)
	backtest.SetDB(db)
	pattern.SetDB(db)

	return &Handler{
		binanceService:  binance,
		aiBrainService:  aiBrain,
		backtestService: backtest,
		patternService:  pattern,
		db:              db,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// GetTicker returns 24h ticker data
func (h *Handler) GetTicker(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	ticker, err := h.binanceService.Get24hTicker(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticker)
}

// GetKlines returns candlestick data
func (h *Handler) GetKlines(c *gin.Context) {
	symbol := c.DefaultQuery("symbol", "BTCUSDT")
	interval := c.DefaultQuery("interval", "1m")
	limit := 500
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	candles, err := h.binanceService.GetKlines(symbol, interval, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":    symbol,
		"interval":  interval,
		"candles":   candles,
		"count":     len(candles),
		"timestamp": time.Now().UnixMilli(),
	})
}

// GetOrderBook returns order book data
func (h *Handler) GetOrderBook(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	orderBook, err := h.binanceService.GetOrderBook(symbol, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":    symbol,
		"orderBook": orderBook,
		"timestamp": time.Now().UnixMilli(),
	})
}

// GetPrediction returns AI prediction for a symbol
func (h *Handler) GetPrediction(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	prediction, err := h.aiBrainService.GetPrediction(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No prediction found"})
		return
	}

	c.JSON(http.StatusOK, prediction)
}

// GetAnalysis returns full AI analysis
func (h *Handler) GetAnalysis(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	// Get latest candle data
	candles, err := h.binanceService.GetKlines(symbol, "1m", 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, predictions, err := h.aiBrainService.GetAnalysis(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":    result,
		"predictions": predictions,
		"candles":     candles,
	})
}

// TrainModel triggers model training
func (h *Handler) TrainModel(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol"`
		Epochs int    `json:"epochs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Symbol = "BTCUSDT"
		req.Epochs = 100
	}

	err := h.aiBrainService.TrainModel(req.Symbol, req.Epochs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "training_completed",
		"symbol": req.Symbol,
		"epochs": req.Epochs,
	})
}

// RunBacktest runs backtest simulation
func (h *Handler) RunBacktest(c *gin.Context) {
	var config backtest.BacktestConfig

	if err := c.ShouldBindJSON(&config); err != nil {
		// Set default configuration (20 years of data for BTC)
		now := time.Now()
		twentyYearsAgo := now.AddDate(-20, 0, 0)

		config = backtest.BacktestConfig{
			Symbol:         "BTCUSDT",
			Timeframe:       "1h",
			StartDate:       twentyYearsAgo.UnixMilli(),
			EndDate:         now.UnixMilli(),
			InitialCapital:  10000,
			Commission:      0.001,
			Slippage:        0.0005,
		}
	}

	result, err := h.backtestService.RunBacktest(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetBacktestResults returns all backtest results
func (h *Handler) GetBacktestResults(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	results, err := h.backtestService.GetBacktestResults(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": results})
}

// GetPatterns returns detected patterns for a symbol
func (h *Handler) GetPatterns(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	patterns, err := h.patternService.GetPatterns(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patterns": patterns})
}

// GetLearnedPatterns returns all learned patterns
func (h *Handler) GetLearnedPatterns(c *gin.Context) {
	patterns, err := h.patternService.GetLearnedPatterns()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"patterns": patterns})
}

// PlaceOrder places a trade order
func (h *Handler) PlaceOrder(c *gin.Context) {
	var order struct {
		Symbol   string  `json:"symbol"`
		Side     string  `json:"side"` // buy, sell
		Type     string  `json:"type"` // market, limit
		Quantity float64 `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order parameters"})
		return
	}

	result, err := h.binanceService.PlaceOrder(order.Symbol, order.Side, order.Type, order.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetBalance returns account balance
func (h *Handler) GetBalance(c *gin.Context) {
	balance, err := h.binanceService.GetAccountBalance()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

// HandleWebSocket handles WebSocket connections for real-time data
func (h *Handler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	// Send initial data
	symbols := []string{"BTCUSDT", "XAUUSDT"}

	for {
		// Get latest data for all symbols
		for _, symbol := range symbols {
			candles, err := h.binanceService.GetKlines(symbol, "1m", 1)
			if err == nil && len(candles) > 0 {
				ticker, _ := h.binanceService.Get24hTicker(symbol)

				data := map[string]interface{}{
					"type":      "ticker",
					"symbol":    symbol,
					"price":     candles[0].Close,
					"volume":    candles[0].Volume,
					"change":    ticker.PriceChangePercent,
					"timestamp": time.Now().UnixMilli(),
				}

				jsonData, _ := json.Marshal(data)
				conn.WriteMessage(websocket.TextMessage, jsonData)
			}
		}

		// Get latest analysis
		for _, symbol := range symbols {
			candles, err := h.binanceService.GetKlines(symbol, "5m", 200)
			if err == nil && len(candles) > 50 {
				analysis, err := h.aiBrainService.AnalyzeSymbol(candles, "5m")
				if err == nil && analysis != nil {
					data := map[string]interface{}{
						"type":        "analysis",
						"symbol":      symbol,
						"prediction": analysis.Prediction,
						"confidence":  analysis.Confidence,
						"entry":       analysis.EntryPrice,
						"target":      analysis.TargetPrice,
						"stopLoss":    analysis.StopLoss,
						"reason":      analysis.Reason,
						"timestamp":   time.Now().UnixMilli(),
					}

					jsonData, _ := json.Marshal(data)
					conn.WriteMessage(websocket.TextMessage, jsonData)
				}
			}
		}

		// Sleep for 1 second before next update
		time.Sleep(1 * time.Second)
	}
}