package models

import (
	"time"

	"gorm.io/gorm"
)

type Candle struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"index" json:"symbol"`
	OpenTime  int64     `gorm:"index" json:"open_time"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	CloseTime int64     `json:"close_time"`
	CreatedAt time.Time `json:"created_at"`
}

type Pattern struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Symbol       string    `gorm:"index" json:"symbol"`
	PatternType  string    `json:"pattern_type"`
	Confidence   float64   `json:"confidence"`
	Direction    string    `json:"direction"` // bullish, bearish, neutral
	StartTime    int64     `json:"start_time"`
	EndTime      int64     `json:"end_time"`
	TargetPrice  float64   `json:"target_price"`
	StopLoss     float64   `json:"stop_loss"`
	ActualResult string    `json:"actual_result"` // success, failed, pending
	Metadata     string    `json:"metadata"` // JSON stored pattern data
	CreatedAt    time.Time `json:"created_at"`
}

type Prediction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Symbol      string    `gorm:"index" json:"symbol"`
	Prediction  string    `json:"prediction"` // buy, sell, hold
	Confidence  float64   `json:"confidence"`
	EntryPrice  float64   `json:"entry_price"`
	TargetPrice float64   `json:"target_price"`
	StopLoss    float64   `json:"stop_loss"`
	Timeframe   string    `json:"timeframe"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

type BacktestResult struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
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
	TradeLog          string    `json:"trade_log"` // JSON array of trades
	CreatedAt         time.Time `json:"created_at"`
}

type Trade struct {
	EntryTime   int64   `json:"entry_time"`
	ExitTime    int64   `json:"exit_time"`
	Direction   string  `json:"direction"` // long, short
	EntryPrice  float64 `json:"entry_price"`
	ExitPrice   float64 `json:"exit_price"`
	PnL         float64 `json:"pnl"`
	PnLPercent  float64 `json:"pnl_percent"`
}

func (Candle) TableName() string {
	return "candles"
}

func (Pattern) TableName() string {
	return "patterns"
}

func (Prediction) TableName() string {
	return "predictions"
}

func (BacktestResult) TableName() string {
	return "backtest_results"
}

// AutoMigrate creates all tables
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Candle{}, &Pattern{}, &Prediction{}, &BacktestResult{})
}