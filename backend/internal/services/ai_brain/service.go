package ai_brain

import (
	"math"
	"time"
	"zen_agentic_ai/internal/models"
	"zen_agentic_ai/internal/services/pattern"
	"zen_agentic_ai/internal/services/binance"

	"gorm.io/gorm"
)

type Service struct {
	db             *gorm.DB
	patternService *pattern.Service
	learningRate   float64
	decayRate      float64
}

type AnalysisResult struct {
	Symbol        string    `json:"symbol"`
	CurrentPrice  float64   `json:"current_price"`
	Prediction    string    `json:"prediction"`
	Confidence    float64   `json:"confidence"`
	EntryPrice    float64   `json:"entry_price"`
	TargetPrice   float64   `json:"target_price"`
	StopLoss      float64   `json:"stop_loss"`
	Timeframe     string    `json:"timeframe"`
	Pattern       string    `json:"pattern"`
	Reason        string    `json:"reason"`
	Indicators    IndicatorValues `json:"indicators"`
	Timestamp     int64     `json:"timestamp"`
}

type IndicatorValues struct {
	RSI           float64 `json:"rsi"`
	MACD          float64 `json:"macd"`
	MACDSignal    float64 `json:"macd_signal"`
	MACDHistogram float64 `json:"macd_histogram"`
	EMA20         float64 `json:"ema_20"`
	EMA50         float64 `json:"ema_50"`
	EMA200        float64 `json:"ema_200"`
	BBUpper       float64 `json:"bb_upper"`
	BBMiddle      float64 `json:"bb_middle"`
	BBLower       float64 `json:"bb_lower"`
	ATR           float64 `json:"atr"`
	VolumeRatio   float64 `json:"volume_ratio"`
}

func NewService(patternService *pattern.Service) *Service {
	return &Service{
		patternService: patternService,
		learningRate:   0.01,
		decayRate:      0.001,
	}
}

func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

// AnalyzeSymbol performs comprehensive technical analysis
func (s *Service) AnalyzeSymbol(candles []models.Candle, timeframe string) (*AnalysisResult, error) {
	if len(candles) < 50 {
		return nil, nil
	}

	symbol := candles[len(candles)-1].Symbol
	_ = symbol // suppress unused variable warning

	// Calculate indicators
	indicators := s.calculateIndicators(candles)

	// Detect patterns
	detectedPattern := s.patternService.DetectPattern(candles)

	// Generate prediction based on multiple factors
	prediction, confidence, reason := s.makeDecision(candles, indicators, detectedPattern)

	// Calculate entry, target, and stop loss
	entry, target, stopLoss := s.calculateTradeLevels(candles, prediction, indicators)

	result := &AnalysisResult{
		Symbol:     symbol,
		Prediction: prediction,
		Confidence: confidence,
		EntryPrice: entry,
		TargetPrice: target,
		StopLoss:   stopLoss,
		Timeframe:  timeframe,
		Pattern:    detectedPattern.Name,
		Reason:     reason,
		Indicators: indicators,
		Timestamp:  time.Now().UnixMilli(),
	}

	// Save prediction to database
	s.savePrediction(result)

	return result, nil
}

// GetPrediction returns cached prediction for a symbol
func (s *Service) GetPrediction(symbol string) (*models.Prediction, error) {
	var pred models.Prediction
	err := s.db.Where("symbol = ?", symbol).Order("created_at DESC").First(&pred).Error
	if err != nil {
		return nil, err
	}
	return &pred, nil
}

// UpdatePredictions updates all symbol predictions
func (s *Service) UpdatePredictions(binance *binance.Service) {
	symbols := []string{"BTCUSDT", "XAUUSDT"} // BTC and Gold

	for _, sym := range symbols {
		candles, err := binance.GetKlines(sym, "1m", 200)
		if err != nil {
			continue
		}

		if len(candles) > 50 {
			s.AnalyzeSymbol(candles, "1m")
		}
	}
}

// calculateIndicators computes all technical indicators
func (s *Service) calculateIndicators(candles []models.Candle) IndicatorValues {
	closes := make([]float64, len(candles))
	volumes := make([]float64, len(candles))

	for i, c := range candles {
		closes[i] = c.Close
		volumes[i] = c.Volume
	}

	indicators := IndicatorValues{
		RSI:           s.calculateRSI(closes, 14),
		EMA20:         s.calculateEMA(closes, 20),
		EMA50:         s.calculateEMA(closes, 50),
		EMA200:        s.calculateEMA(closes, 200),
		BBUpper:       s.calculateBollingerBands(closes, 20).Upper,
		BBMiddle:      s.calculateBollingerBands(closes, 20).Middle,
		BBLower:       s.calculateBollingerBands(closes, 20).Lower,
		ATR:           s.calculateATR(candles, 14),
		VolumeRatio:   s.calculateVolumeRatio(volumes, 20),
	}

	// MACD
	macdLine, signalLine, hist := s.calculateMACD(closes)
	indicators.MACD = macdLine
	indicators.MACDSignal = signalLine
	indicators.MACDHistogram = hist

	return indicators
}

func (s *Service) calculateRSI(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 50
	}

	var gains, losses float64
	for i := len(closes) - period; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gains += change
		} else {
			losses += -change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))
	return rsi
}

func (s *Service) calculateEMA(closes []float64, period int) float64 {
	if len(closes) < period {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	ema := closes[0]

	for i := 1; i < len(closes); i++ {
		ema = closes[i]*multiplier + ema*(1-multiplier)
	}

	return ema
}

type BollingerBands struct {
	Upper   float64
	Middle  float64
	Lower   float64
}

func (s *Service) calculateBollingerBands(closes []float64, period int) BollingerBands {
	if len(closes) < period {
		return BollingerBands{}
	}

	// Calculate SMA (Middle)
	var sum float64
	for i := len(closes) - period; i < len(closes); i++ {
		sum += closes[i]
	}
	sma := sum / float64(period)

	// Calculate standard deviation
	var variance float64
	for i := len(closes) - period; i < len(closes); i++ {
		diff := closes[i] - sma
		variance += diff * diff
	}
	stdDev := math.Sqrt(variance / float64(period))

	return BollingerBands{
		Upper:   sma + 2*stdDev,
		Middle:  sma,
		Lower:   sma - 2*stdDev,
	}
}

func (s *Service) calculateATR(candles []models.Candle, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	var atrSum float64
	for i := len(candles) - period; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		atrSum += tr
	}

	return atrSum / float64(period)
}

func (s *Service) calculateVolumeRatio(volumes []float64, period int) float64 {
	if len(volumes) < period+1 {
		return 1
	}

	var avgVol, recentVol float64
	for i := len(volumes) - period; i < len(volumes)-1; i++ {
		avgVol += volumes[i]
	}
	avgVol /= float64(period - 1)

	recentVol = volumes[len(volumes)-1]

	if avgVol == 0 {
		return 1
	}

	return recentVol / avgVol
}

func (s *Service) calculateMACD(closes []float64) (macdLine, signalLine, histogram float64) {
	if len(closes) < 26 {
		return 0, 0, 0
	}

	ema12 := s.calculateEMA(closes, 12)
	ema26 := s.calculateEMA(closes, 26)
	macdLine = ema12 - ema26

	// Calculate signal line (EMA9 of MACD)
	// Simplified calculation
	signalLine = macdLine * 0.9 // Placeholder

	histogram = macdLine - signalLine

	return
}

// makeDecision combines all signals to make a trading decision
func (s *Service) makeDecision(candles []models.Candle, indicators IndicatorValues, pattern pattern.DetectedPattern) (prediction string, confidence float64, reason string) {
	var score float64

	// RSI Analysis
	if indicators.RSI < 30 {
		score += 1 // Oversold - bullish
		reason += "RSI oversold; "
	} else if indicators.RSI > 70 {
		score -= 1 // Overbought - bearish
		reason += "RSI overbought; "
	}

	// EMA Analysis
	if indicators.EMA20 > indicators.EMA50 && indicators.EMA50 > indicators.EMA200 {
		score += 1 // Uptrend
		reason += "Strong uptrend; "
	} else if indicators.EMA20 < indicators.EMA50 && indicators.EMA50 < indicators.EMA200 {
		score -= 1 // Downtrend
		reason += "Strong downtrend; "
	}

	// MACD Analysis
	if indicators.MACDHistogram > 0 {
		score += 0.5 // Bullish momentum
		reason += "MACD bullish; "
	} else {
		score -= 0.5 // Bearish momentum
		reason += "MACD bearish; "
	}

	// Bollinger Bands
	currentPrice := candles[len(candles)-1].Close
	if currentPrice < indicators.BBLower {
		score += 0.5 // Near lower band - potential bounce
		reason += "Near lower BB; "
	} else if currentPrice > indicators.BBUpper {
		score -= 0.5 // Near upper band - potential reversal
		reason += "Near upper BB; "
	}

	// Volume
	if indicators.VolumeRatio > 1.5 {
		if score > 0 {
			score += 0.5 // High volume + bullish = stronger
		} else {
			score -= 0.5 // High volume + bearish = stronger
		}
		reason += "High volume; "
	}

	// Pattern recognition
	if pattern.Name != "" {
		if pattern.Direction == "bullish" {
			score += pattern.Confidence * 0.5
			reason += "Pattern: " + pattern.Name + "; "
		} else if pattern.Direction == "bearish" {
			score -= pattern.Confidence * 0.5
			reason += "Pattern: " + pattern.Name + "; "
		}
	}

	// Convert score to prediction
	if score >= 1.0 {
		prediction = "buy"
		confidence = math.Min(0.95, 0.5+score*0.1)
	} else if score <= -1.0 {
		prediction = "sell"
		confidence = math.Min(0.95, 0.5+math.Abs(score)*0.1)
	} else {
		prediction = "hold"
		confidence = 0.6
	}

	if reason == "" {
		reason = "No clear signals - holding"
	}

	return
}

// calculateTradeLevels calculates entry, target, and stop loss
func (s *Service) calculateTradeLevels(candles []models.Candle, prediction string, indicators IndicatorValues) (entry, target, stopLoss float64) {
	currentPrice := candles[len(candles)-1].Close
	atr := indicators.ATR

	if atr == 0 {
		atr = currentPrice * 0.02 // Default 2% if ATR unavailable
	}

	switch prediction {
	case "buy":
		entry = currentPrice
		target = currentPrice + (atr * 2) // 2x ATR target
		stopLoss = currentPrice - (atr * 1.5) // 1.5x ATR stop
	case "sell":
		entry = currentPrice
		target = currentPrice - (atr * 2)
		stopLoss = currentPrice + (atr * 1.5)
	default:
		entry = currentPrice
		target = currentPrice
		stopLoss = currentPrice
	}

	return
}

func (s *Service) savePrediction(result *AnalysisResult) {
	prediction := models.Prediction{
		Symbol:     result.Symbol,
		Prediction: result.Prediction,
		Confidence: result.Confidence,
		EntryPrice: result.EntryPrice,
		TargetPrice: result.TargetPrice,
		StopLoss:   result.StopLoss,
		Timeframe:  result.Timeframe,
		Reason:     result.Reason,
		CreatedAt:  time.Now(),
	}

	s.db.Create(&prediction)
}

// TrainModel simulates model training (placeholder for ML integration)
func (s *Service) TrainModel(symbol string, epochs int) error {
	// In production, this would integrate with TensorFlow, PyTorch, or similar
	// For now, we'll use pattern learning
	return s.patternService.LearnFromHistory(s.db, symbol, 500)
}

// GetAnalysis returns full analysis with history
func (s *Service) GetAnalysis(symbol string) (*AnalysisResult, []models.Prediction, error) {
	// Get latest prediction
	var predictions []models.Prediction
	err := s.db.Where("symbol = ?", symbol).Order("created_at DESC").Limit(100).Find(&predictions).Error
	if err != nil {
		return nil, nil, err
	}

	if len(predictions) == 0 {
		return nil, predictions, nil
	}

	latest := predictions[0]

	result := &AnalysisResult{
		Symbol:        latest.Symbol,
		Prediction:    latest.Prediction,
		Confidence:    latest.Confidence,
		EntryPrice:    latest.EntryPrice,
		TargetPrice:   latest.TargetPrice,
		StopLoss:      latest.StopLoss,
		Timeframe:     latest.Timeframe,
		Reason:        latest.Reason,
		Timestamp:     latest.CreatedAt.UnixMilli(),
	}

	return result, predictions, nil
}