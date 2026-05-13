package backtest

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
	"zen_agentic_ai/internal/models"
	"zen_agentic_ai/internal/services/ai_brain"
	"zen_agentic_ai/internal/services/binance"

	"gorm.io/gorm"
)

type Service struct {
	db              *gorm.DB
	binanceService  *binance.Service
	aiBrainService  *ai_brain.Service
	maxHistoricalYears int
}

type BacktestConfig struct {
	Symbol        string  `json:"symbol"`
	Timeframe     string  `json:"timeframe"`
	StartDate     int64   `json:"start_date"`
	EndDate       int64   `json:"end_date"`
	InitialCapital float64 `json:"initial_capital"`
	Commission    float64 `json:"commission"` // 0.001 for 0.1%
	Slippage      float64 `json:"slippage"`   // 0.0005 for 0.05%
}

type BacktestResult struct {
	Symbol            string    `json:"symbol"`
	Timeframe         string    `json:"timeframe"`
	StartDate         int64     `json:"start_date"`
	EndDate           int64     `json:"end_date"`
	InitialCapital    float64   `json:"initial_capital"`
	FinalCapital      float64   `json:"final_capital"`
	TotalReturn       float64   `json:"total_return"`
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	MaxDrawdown       float64   `json:"max_drawdown"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	AvgTradeDuration  float64   `json:"avg_trade_duration"`
	Trades            []Trade   `json:"trades"`
}

type Trade struct {
	EntryTime   int64   `json:"entry_time"`
	EntryCandle int     `json:"entry_candle"`
	Direction   string  `json:"direction"`
	EntryPrice  float64 `json:"entry_price"`
	ExitTime    int64   `json:"exit_time"`
	ExitCandle  int     `json:"exit_candle"`
	ExitPrice   float64 `json:"exit_price"`
	PnL         float64 `json:"pnl"`
	PnLPercent  float64 `json:"pnl_percent"`
	Duration    int     `json:"duration"`
	Reason      string  `json:"reason"`
}

type BacktestStats struct {
	TotalReturn    float64   `json:"total_return"`
	AnnualReturn   float64   `json:"annual_return"`
	Volatility     float64   `json:"volatility"`
	SharpeRatio    float64   `json:"sharpe_ratio"`
	MaxDrawdown    float64   `json:"max_drawdown"`
	WinRate        float64   `json:"win_rate"`
	ProfitFactor   float64   `json:"profit_factor"`
	AvgWin         float64   `json:"avg_win"`
	AvgLoss        float64   `json:"avg_loss"`
	BestTrade      float64   `json:"best_trade"`
	WorstTrade     float64   `json:"worst_trade"`
}

func NewService(binance *binance.Service, aiBrain *ai_brain.Service) *Service {
	return &Service{
		binanceService:   binance,
		aiBrainService:   aiBrain,
		maxHistoricalYears: 20,
	}
}

func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

// RunBacktest executes backtesting with 20 years of data for BTC and Gold
func (s *Service) RunBacktest(config BacktestConfig) (*BacktestResult, error) {
	fmt.Printf("Starting backtest for %s from %s to %s\n",
		config.Symbol,
		time.Unix(config.StartDate/1000, 0).Format("2006-01-02"),
		time.Unix(config.EndDate/1000, 0).Format("2006-01-02"))

	// Fetch historical data
	candles, err := s.binanceService.GetHistoricalData(config.Symbol, config.Timeframe, config.StartDate, config.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %v", err)
	}

	fmt.Printf("Loaded %d candles for backtesting\n", len(candles))

	if len(candles) < 100 {
		return nil, fmt.Errorf("insufficient data for backtesting: need at least 100 candles, got %d", len(candles))
	}

	// Run backtest simulation
	result := s.simulateBacktest(candles, config)

	// Save result to database
	s.saveBacktestResult(result)

	return result, nil
}

// simulateBacktest runs the actual backtest simulation
func (s *Service) simulateBacktest(candles []models.Candle, config BacktestConfig) *BacktestResult {
	capital := config.InitialCapital
	var trades []Trade
	var equityCurve []float64

	position := 0 // 0: no position, 1: long, -1: short
	var entryPrice float64
	var entryTime int64
	var entryCandle int

	maxCapital := capital
	peakCapital := capital
	var maxDrawdown float64

	totalWinPnl := 0.0
	totalLossPnl := 0.0

	// Warm up period for indicators (200 candles)
	warmup := 200

	// Process each candle
	for i := warmup; i < len(candles)-1; i++ {
		// Get lookback data for analysis
		lookback := candles[:i+1]

		// Analyze current market
		analysis, err := s.aiBrainService.AnalyzeSymbol(lookback, config.Timeframe)
		if err != nil {
			continue
		}

		// Apply slippage
		currentPrice := candles[i].Close
		slippage := config.Slippage

		// Entry logic
		if position == 0 {
			if analysis.Prediction == "buy" && analysis.Confidence > 0.65 {
				// Enter long
				position = 1
				entryPrice = currentPrice * (1 + slippage)
				entryTime = candles[i].OpenTime
				entryCandle = i
			} else if analysis.Prediction == "sell" && analysis.Confidence > 0.65 {
				// Enter short
				position = -1
				entryPrice = currentPrice * (1 - slippage)
				entryTime = candles[i].OpenTime
				entryCandle = i
			}
		} else {
			// Exit logic
			shouldExit := false

			// Take profit at target
			if position == 1 && currentPrice >= analysis.TargetPrice {
				shouldExit = true
			} else if position == -1 && currentPrice <= analysis.TargetPrice {
				shouldExit = true
			}

			// Stop loss
			if position == 1 && currentPrice <= analysis.StopLoss {
				shouldExit = true
			} else if position == -1 && currentPrice >= analysis.StopLoss {
				shouldExit = true
			}

			// Time-based exit (max hold 50 candles)
			if i-entryCandle > 50 {
				shouldExit = true
			}

			// Opposite signal
			if (position == 1 && analysis.Prediction == "sell" && analysis.Confidence > 0.7) ||
				(position == -1 && analysis.Prediction == "buy" && analysis.Confidence > 0.7) {
				shouldExit = true
			}

			if shouldExit {
				var exitPrice float64
				if position == 1 {
					exitPrice = currentPrice * (1 - slippage)
				} else {
					exitPrice = currentPrice * (1 + slippage)
				}

				var pnl, pnlPercent float64
				if position == 1 {
					pnl = (exitPrice - entryPrice) * (capital / entryPrice)
					pnlPercent = (exitPrice - entryPrice) / entryPrice * 100
				} else {
					pnl = (entryPrice - exitPrice) * (capital / entryPrice)
					pnlPercent = (entryPrice - exitPrice) / entryPrice * 100
				}

				// Apply commission
				commission := capital * config.Commission * 2 // Entry + Exit
				pnl -= commission

				capital += pnl
				if pnl > 0 {
					totalWinPnl += pnl
				} else {
					totalLossPnl += -pnl
				}

				trade := Trade{
					EntryTime:   entryTime,
					EntryCandle: entryCandle,
					Direction:   map[int]string{1: "long", -1: "short"}[position],
					EntryPrice:  entryPrice,
					ExitTime:    candles[i].OpenTime,
					ExitCandle:  i,
					ExitPrice:   exitPrice,
					PnL:         pnl,
					PnLPercent:  pnlPercent,
					Duration:    i - entryCandle,
					Reason:      analysis.Reason,
				}
				trades = append(trades, trade)

				// Reset position
				position = 0
			}
		}

		// Track equity and drawdown
		equityCurve = append(equityCurve, capital)

		if capital > peakCapital {
			peakCapital = capital
		}

		currentDrawdown := (peakCapital - capital) / peakCapital * 100
		if currentDrawdown > maxDrawdown {
			maxDrawdown = currentDrawdown
		}

		if capital > maxCapital {
			maxCapital = capital
		}
	}

	// Close any open position at the end
	if position != 0 {
		lastPrice := candles[len(candles)-1].Close
		var pnl float64

		if position == 1 {
			pnl = (lastPrice - entryPrice) * (capital / entryPrice)
		} else {
			pnl = (entryPrice - lastPrice) * (capital / entryPrice)
		}

		capital += pnl

		trade := Trade{
			EntryTime:   entryTime,
			EntryCandle: entryCandle,
			Direction:   map[int]string{1: "long", -1: "short"}[position],
			EntryPrice:  entryPrice,
			ExitTime:    candles[len(candles)-1].OpenTime,
			ExitCandle:  len(candles) - 1,
			ExitPrice:   lastPrice,
			PnL:         pnl,
			Duration:    len(candles) - 1 - entryCandle,
		}
		trades = append(trades, trade)
	}

	// Calculate statistics
	totalTrades := len(trades)
	winningTrades := 0
	losingTrades := 0
	var totalDuration int
	var bestTrade, worstTrade float64

	for _, t := range trades {
		if t.PnL > 0 {
			winningTrades++
		} else {
			losingTrades++
		}

		totalDuration += t.Duration

		if t.PnL > bestTrade {
			bestTrade = t.PnL
		}
		if t.PnL < worstTrade {
			worstTrade = t.PnL
		}
	}

	totalReturn := ((capital - config.InitialCapital) / config.InitialCapital) * 100
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	avgTradeDuration := 0.0
	if totalTrades > 0 {
		avgTradeDuration = float64(totalDuration) / float64(totalTrades)
	}

	// Calculate Sharpe Ratio
	sharpe := s.calculateSharpeRatio(equityCurve)

	result := &BacktestResult{
		Symbol:           config.Symbol,
		Timeframe:        config.Timeframe,
		StartDate:        config.StartDate,
		EndDate:          config.EndDate,
		InitialCapital:   config.InitialCapital,
		FinalCapital:     capital,
		TotalReturn:      totalReturn,
		TotalTrades:      totalTrades,
		WinningTrades:    winningTrades,
		LosingTrades:     losingTrades,
		WinRate:          winRate,
		MaxDrawdown:      maxDrawdown,
		SharpeRatio:      sharpe,
		AvgTradeDuration: avgTradeDuration,
		Trades:           trades,
	}

	return result
}

// calculateSharpeRatio computes Sharpe ratio from equity curve
func (s *Service) calculateSharpeRatio(equityCurve []float64) float64 {
	if len(equityCurve) < 2 {
		return 0
	}

	// Calculate returns
	returns := make([]float64, len(equityCurve)-1)
	for i := 1; i < len(equityCurve); i++ {
		returns[i-1] = (equityCurve[i] - equityCurve[i-1]) / equityCurve[i-1]
	}

	// Calculate mean return
	var meanReturn float64
	for _, r := range returns {
		meanReturn += r
	}
	meanReturn /= float64(len(returns))

	// Calculate standard deviation
	var variance float64
	for _, r := range returns {
		diff := r - meanReturn
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(len(returns)))

	if stdDev == 0 {
		return 0
	}

	// Annualize (assuming 252 trading days)
	sharpe := (meanReturn * 252) / (stdDev * math.Sqrt(252))

	return sharpe
}

// saveBacktestResult stores result in database
func (s *Service) saveBacktestResult(result *BacktestResult) {
	tradeLog, _ := json.Marshal(result.Trades)

	dbResult := models.BacktestResult{
		Symbol:           result.Symbol,
		Timeframe:        result.Timeframe,
		StartDate:        result.StartDate,
		EndDate:          result.EndDate,
		InitialCapital:   result.InitialCapital,
		FinalCapital:     result.FinalCapital,
		TotalReturn:      result.TotalReturn,
		TotalTrades:      result.TotalTrades,
		WinningTrades:    result.WinningTrades,
		LosingTrades:     result.LosingTrades,
		WinRate:          result.WinRate,
		MaxDrawdown:      result.MaxDrawdown,
		SharpeRatio:      result.SharpeRatio,
		AvgTradeDuration: result.AvgTradeDuration,
		TradeLog:         string(tradeLog),
		CreatedAt:        time.Now(),
	}

	s.db.Create(&dbResult)
}

// GetBacktestResults returns all backtest results
func (s *Service) GetBacktestResults(limit int) ([]models.BacktestResult, error) {
	var results []models.BacktestResult
	err := s.db.Order("created_at DESC").Limit(limit).Find(&results).Error
	return results, err
}

// GetAvailableDataRange returns available historical data for a symbol
func (s *Service) GetAvailableDataRange(symbol string) (startTime, endTime int64, err error) {
	// Get current time
	endTime = time.Now().UnixMilli()

	// Calculate start time (20 years ago)
	startTime = time.Now().AddDate(-s.maxHistoricalYears, 0, 0).UnixMilli()

	return startTime, endTime, nil
}

// GetAvailableSymbols returns supported symbols
func (s *Service) GetAvailableSymbols() []string {
	return []string{
		"BTCUSDT", // Bitcoin
		"XAUUSDT", // Gold
	}
}

// GetBacktestSummary returns summary statistics
func (s *Service) GetBacktestSummary(symbol string) (*BacktestStats, error) {
	var results []models.BacktestResult
	err := s.db.Where("symbol = ?", symbol).Find(&results).Error
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no backtest results found for %s", symbol)
	}

	// Aggregate statistics from all results
	var totalReturn float64
	var totalTrades, winningTrades int
	var totalPnl, totalWinPnl, totalLossPnl float64

	for _, r := range results {
		totalReturn += r.TotalReturn
		totalTrades += r.TotalTrades
		winningTrades += r.WinningTrades
		totalPnl += (r.FinalCapital - r.InitialCapital)
	}

	avgReturn := totalReturn / float64(len(results))
	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	profitFactor := 0.0
	if totalLossPnl > 0 {
		profitFactor = totalWinPnl / totalLossPnl
	}

	return &BacktestStats{
		TotalReturn:   totalReturn,
		AnnualReturn: avgReturn / 20, // Assuming 20 years
		WinRate:       winRate,
		ProfitFactor: profitFactor,
	}, nil
}