package pattern

import (
	"math"
	"sort"
	"time"
	"trading-system/backend/internal/models"

	"github.com/jinzhu/gorm"
)

type Service struct {
	db             *gorm.DB
	knownPatterns map[string]PatternDefinition
}

type PatternDefinition struct {
	Name         string
	Type         string // continuation, reversal, neutral
	MinCandles   int
	DetectFunc   func([]models.Candle) DetectedPattern
}

type DetectedPattern struct {
	Name       string
	Type       string
	Direction  string // bullish, bearish, neutral
	Confidence float64
	StartTime  int64
	EndTime    int64
	Metadata   string
}

func NewService() *Service {
	s := &Service{
		knownPatterns: make(map[string]PatternDefinition),
	}
	s.registerPatterns()
	return s
}

func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

// registerPatterns registers all known chart patterns
func (s *Service) registerPatterns() {
	// Candlestick Patterns
	s.knownPatterns["doji"] = PatternDefinition{
		Name:       "Doji",
		Type:       "reversal",
		MinCandles: 1,
		DetectFunc: s.detectDoji,
	}

	s.knownPatterns["hammer"] = PatternDefinition{
		Name:       "Hammer",
		Type:       "reversal",
		MinCandles: 1,
		DetectFunc: s.detectHammer,
	}

	s.knownPatterns["shooting_star"] = PatternDefinition{
		Name:       "Shooting Star",
		Type:       "reversal",
		MinCandles: 1,
		DetectFunc: s.detectShootingStar,
	}

	s.knownPatterns["engulfing_bullish"] = PatternDefinition{
		Name:       "Bullish Engulfing",
		Type:       "reversal",
		MinCandles: 2,
		DetectFunc: s.detectBullishEngulfing,
	}

	s.knownPatterns["engulfing_bearish"] = PatternDefinition{
		Name:       "Bearish Engulfing",
		Type:       "reversal",
		MinCandles: 2,
		DetectFunc: s.detectBearishEngulfing,
	}

	s.knownPatterns["morning_star"] = PatternDefinition{
		Name:       "Morning Star",
		Type:       "reversal",
		MinCandles: 3,
		DetectFunc: s.detectMorningStar,
	}

	s.knownPatterns["evening_star"] = PatternDefinition{
		Name:       "Evening Star",
		Type:       "reversal",
		MinCandles: 3,
		DetectFunc: s.detectEveningStar,
	}

	// Chart Patterns
	s.knownPatterns["head_shoulders"] = PatternDefinition{
		Name:       "Head and Shoulders",
		Type:       "reversal",
		MinCandles: 30,
		DetectFunc: s.detectHeadAndShoulders,
	}

	s.knownPatterns["double_top"] = PatternDefinition{
		Name:       "Double Top",
		Type:       "reversal",
		MinCandles: 20,
		DetectFunc: s.detectDoubleTop,
	}

	s.knownPatterns["double_bottom"] = PatternDefinition{
		Name:       "Double Bottom",
		Type:       "reversal",
		MinCandles: 20,
		DetectFunc: s.detectDoubleBottom,
	}

	s.knownPatterns["ascending_triangle"] = PatternDefinition{
		Name:       "Ascending Triangle",
		Type:       "continuation",
		MinCandles: 20,
		DetectFunc: s.detectAscendingTriangle,
	}

	s.knownPatterns["descending_triangle"] = PatternDefinition{
		Name:       "Descending Triangle",
		Type:       "continuation",
		MinCandles: 20,
		DetectFunc: s.detectDescendingTriangle,
	}

	s.knownPatterns["wedge_rising"] = PatternDefinition{
		Name:       "Rising Wedge",
		Type:       "reversal",
		MinCandles: 20,
		DetectFunc: s.detectRisingWedge,
	}

	s.knownPatterns["wedge_falling"] = PatternDefinition{
		Name:       "Falling Wedge",
		Type:       "reversal",
		MinCandles: 20,
		DetectFunc: s.detectFallingWedge,
	}
}

// DetectPattern scans for all known patterns
func (s *Service) DetectPattern(candles []models.Candle) DetectedPattern {
	var bestPattern DetectedPattern

	for _, p := range s.knownPatterns {
		if len(candles) >= p.MinCandles {
			detected := p.DetectFunc(candles)
			if detected.Confidence > bestPattern.Confidence {
				bestPattern = detected
			}
		}
	}

	return bestPattern
}

// GetPatterns returns detected patterns for a symbol
func (s *Service) GetPatterns(symbol string) ([]models.Pattern, error) {
	var patterns []models.Pattern
	err := s.db.Where("symbol = ?", symbol).Order("created_at DESC").Find(&patterns).Error
	return patterns, err
}

// LearnFromHistory analyzes historical data and learns pattern success rates
func (s *Service) LearnFromHistory(db *gorm.DB, symbol string, limit int) error {
	// This would analyze historical patterns and update confidence scores
	// For now, we'll create pattern records for detected patterns
	return nil
}

// Pattern Detection Functions

func (s *Service) detectDoji(candles []models.Candle) DetectedPattern {
	c := candles[len(candles)-1]
	body := math.Abs(c.Close - c.Open)
	range_ := c.High - c.Low

	if range_ > 0 && body/range_ < 0.1 {
		return DetectedPattern{
			Name:       "Doji",
			Type:       "reversal",
			Direction:  "neutral",
			Confidence: 0.7,
			StartTime:  c.OpenTime,
			EndTime:    c.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectHammer(candles []models.Candle) DetectedPattern {
	c := candles[len(candles)-1]
	body := math.Abs(c.Close - c.Open)
	lowerWick := math.Min(c.Open, c.Close) - c.Low
	upperWick := c.High - math.Max(c.Open, c.Close)

	// Hammer: small body, long lower wick, little or no upper wick
	if lowerWick > body*2 && upperWick < body {
		return DetectedPattern{
			Name:       "Hammer",
			Type:       "reversal",
			Direction:  "bullish",
			Confidence: 0.8,
			StartTime:  c.OpenTime,
			EndTime:    c.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectShootingStar(candles []models.Candle) DetectedPattern {
	c := candles[len(candles)-1]
	body := math.Abs(c.Close - c.Open)
	lowerWick := math.Min(c.Open, c.Close) - c.Low
	upperWick := c.High - math.Max(c.Open, c.Close)

	// Shooting star: small body, long upper wick, little or no lower wick
	if upperWick > body*2 && lowerWick < body {
		return DetectedPattern{
			Name:       "Shooting Star",
			Type:       "reversal",
			Direction:  "bearish",
			Confidence: 0.8,
			StartTime:  c.OpenTime,
			EndTime:    c.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectBullishEngulfing(candles []models.Candle) DetectedPattern {
	if len(candles) < 2 {
		return DetectedPattern{}
	}

	prev := candles[len(candles)-2]
	curr := candles[len(candles)-1]

	// Previous must be bearish (red)
	// Current must be bullish (green) and engulf previous
	prevBearish := prev.Close < prev.Open
	currBullish := curr.Close > curr.Open
	engulfs := curr.Open < prev.Close && curr.Close > prev.Open

	if prevBearish && currBullish && engulfs {
		return DetectedPattern{
			Name:       "Bullish Engulfing",
			Type:       "reversal",
			Direction:  "bullish",
			Confidence: 0.85,
			StartTime:  prev.OpenTime,
			EndTime:    curr.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectBearishEngulfing(candles []models.Candle) DetectedPattern {
	if len(candles) < 2 {
		return DetectedPattern{}
	}

	prev := candles[len(candles)-2]
	curr := candles[len(candles)-1]

	// Previous must be bullish (green)
	// Current must be bearish (red) and engulf previous
	prevBullish := prev.Close > prev.Open
	currBearish := curr.Close < curr.Open
	engulfs := curr.Open > prev.Open && curr.Close < prev.Close

	if prevBullish && currBearish && engulfs {
		return DetectedPattern{
			Name:       "Bearish Engulfing",
			Type:       "reversal",
			Direction:  "bearish",
			Confidence: 0.85,
			StartTime:  prev.OpenTime,
			EndTime:    curr.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectMorningStar(candles []models.Candle) DetectedPattern {
	if len(candles) < 3 {
		return DetectedPattern{}
	}

	first := candles[len(candles)-3]
	second := candles[len(candles)-2]
	third := candles[len(candles)-1]

	// First: bearish candle
	// Second: small body (doji-like)
	// Third: bullish candle closing above middle of first
	firstBearish := first.Close < first.Open
	secondSmall := math.Abs(second.Close-second.Open) < (first.High-first.Low)*0.3
	thirdBullish := third.Close > third.Open
	thirdEngulfs := third.Close > (first.Open+first.Close)/2

	if firstBearish && secondSmall && thirdBullish && thirdEngulfs {
		return DetectedPattern{
			Name:       "Morning Star",
			Type:       "reversal",
			Direction:  "bullish",
			Confidence: 0.9,
			StartTime:  first.OpenTime,
			EndTime:    third.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectEveningStar(candles []models.Candle) DetectedPattern {
	if len(candles) < 3 {
		return DetectedPattern{}
	}

	first := candles[len(candles)-3]
	second := candles[len(candles)-2]
	third := candles[len(candles)-1]

	firstBullish := first.Close > first.Open
	secondSmall := math.Abs(second.Close-second.Open) < (first.High-first.Low)*0.3
	thirdBearish := third.Close < third.Open
	thirdEngulfs := third.Close < (first.Open+first.Close)/2

	if firstBullish && secondSmall && thirdBearish && thirdEngulfs {
		return DetectedPattern{
			Name:       "Evening Star",
			Type:       "reversal",
			Direction:  "bearish",
			Confidence: 0.9,
			StartTime:  first.OpenTime,
			EndTime:    third.CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectHeadAndShoulders(candles []models.Candle) DetectedPattern {
	if len(candles) < 30 {
		return DetectedPattern{}
	}

	// Find local maxima and minima
	highs := findLocalMaxima(candles)
	if len(highs) < 3 {
		return DetectedPattern{}
	}

	// Check for head and shoulders pattern
	// Left shoulder, head (higher), right shoulder (lower or equal)
	last3 := highs[len(highs)-3:]
	if last3[1].Price > last3[0].Price && last3[1].Price > last3[2].Price {
		shoulderDiff := math.Abs(last3[0].Price - last3[2].Price)
		headHeight := last3[1].Price - math.Min(last3[0].Price, last3[2].Price)

		if shoulderDiff/headHeight < 0.3 { // Shoulders within 30% of each other
			return DetectedPattern{
				Name:       "Head and Shoulders",
				Type:       "reversal",
				Direction:  "bearish",
				Confidence: 0.8,
				StartTime:  last3[0].Time,
				EndTime:    last3[2].Time,
			}
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectDoubleTop(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	highs := findLocalMaxima(candles)
	if len(highs) < 2 {
		return DetectedPattern{}
	}

	last2 := highs[len(highs)-2:]
	priceDiff := math.Abs(last2[0].Price - last2[1].Price)
	avgPrice := (last2[0].Price + last2[1].Price) / 2

	if priceDiff/avgPrice < 0.02 { // Within 2%
		return DetectedPattern{
			Name:       "Double Top",
			Type:       "reversal",
			Direction:  "bearish",
			Confidence: 0.75,
			StartTime:  last2[0].Time,
			EndTime:    last2[1].Time,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectDoubleBottom(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	lows := findLocalMinima(candles)
	if len(lows) < 2 {
		return DetectedPattern{}
	}

	last2 := lows[len(lows)-2:]
	priceDiff := math.Abs(last2[0].Price - last2[1].Price)
	avgPrice := (last2[0].Price + last2[1].Price) / 2

	if priceDiff/avgPrice < 0.02 { // Within 2%
		return DetectedPattern{
			Name:       "Double Bottom",
			Type:       "reversal",
			Direction:  "bullish",
			Confidence: 0.75,
			StartTime:  last2[0].Time,
			EndTime:    last2[1].Time,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectAscendingTriangle(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	// Look for resistance (flat top) and higher lows
	highs := findLocalMaxima(candles)
	lows := findLocalMinima(candles)

	if len(highs) < 2 || len(lows) < 2 {
		return DetectedPattern{}
	}

	// Check if highs are relatively flat (resistance)
	lastHighs := highs[len(highs)-3:]
	highVariance := varianceOf(lastHighs)
	avgHigh := meanOfPrices(lastHighs)

	// Check if lows are rising
	lastLows := lows[len(lows)-3:]
	allRising := true
	for i := 1; i < len(lastLows); i++ {
		if lastLows[i].Price <= lastLows[i-1].Price {
			allRising = false
			break
		}
	}

	lowVariance := varianceOf(lastLows)
	avgLow := meanOfPrices(lastLows)

	// Resistance should be flat, lows should be rising
	if highVariance/avgHigh < 0.02 && allRising && lowVariance/avgLow < 0.05 {
		return DetectedPattern{
			Name:       "Ascending Triangle",
			Type:       "continuation",
			Direction:  "bullish",
			Confidence: 0.75,
			StartTime:  lastHighs[0].Time,
			EndTime:    candles[len(candles)-1].CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectDescendingTriangle(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	lows := findLocalMinima(candles)
	highs := findLocalMaxima(candles)

	if len(lows) < 2 || len(highs) < 2 {
		return DetectedPattern{}
	}

	// Check if lows are relatively flat (support)
	lastLows := lows[len(lows)-3:]
	lowVariance := varianceOf(lastLows)
	avgLow := meanOfPrices(lastLows)

	// Check if highs are falling
	lastHighs := highs[len(highs)-3:]
	allFalling := true
	for i := 1; i < len(lastHighs); i++ {
		if lastHighs[i].Price >= lastHighs[i-1].Price {
			allFalling = false
			break
		}
	}

	highVariance := varianceOf(lastHighs)
	avgHigh := meanOfPrices(lastHighs)

	if lowVariance/avgLow < 0.02 && allFalling && highVariance/avgHigh < 0.05 {
		return DetectedPattern{
			Name:       "Descending Triangle",
			Type:       "continuation",
			Direction:  "bearish",
			Confidence: 0.75,
			StartTime:  lastLows[0].Time,
			EndTime:    candles[len(candles)-1].CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectRisingWedge(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	// Rising wedge: both highs and lows rising, but highs rising faster
	highs := findLocalMaxima(candles)
	lows := findLocalMinima(candles)

	if len(highs) < 3 || len(lows) < 3 {
		return DetectedPattern{}
	}

	// Check if both are rising but converging
	last3Highs := highs[len(highs)-3:]
	last3Lows := lows[len(lows)-3:()

	highsRising := last3Highs[2].Price > last3Highs[1].Price && last3Highs[1].Price > last3Highs[0].Price
	lowsRising := last3Lows[2].Price > last3Lows[1].Price && last3Lows[1].Price > last3Lows[0].Price

	// Highs rising faster than lows = converging
	highSlope := (last3Highs[2].Price - last3Highs[0].Price) / 2
	lowSlope := (last3Lows[2].Price - last3Lows[0].Price) / 2

	if highsRising && lowsRising && highSlope > lowSlope {
		return DetectedPattern{
			Name:       "Rising Wedge",
			Type:       "reversal",
			Direction:  "bearish",
			Confidence: 0.7,
			StartTime:  last3Highs[0].Time,
			EndTime:    candles[len(candles)-1].CloseTime,
		}
	}

	return DetectedPattern{}
}

func (s *Service) detectFallingWedge(candles []models.Candle) DetectedPattern {
	if len(candles) < 20 {
		return DetectedPattern{}
	}

	highs := findLocalMaxima(candles)
	lows := findLocalMinima(candles)

	if len(highs) < 3 || len(lows) < 3 {
		return DetectedPattern{}
	}

	last3Highs := highs[len(highs)-3:]
	last3Lows := lows[len(lows)-3:()

	highsFalling := last3Highs[2].Price < last3Highs[1].Price && last3Highs[1].Price < last3Highs[0].Price
	lowsFalling := last3Lows[2].Price < last3Lows[1].Price && last3Lows[1].Price < last3Lows[0].Price

	highSlope := (last3Highs[2].Price - last3Highs[0].Price) / 2
	lowSlope := (last3Lows[2].Price - last3Lows[0].Price) / 2

	if highsFalling && lowsFalling && math.Abs(highSlope) > math.Abs(lowSlope) {
		return DetectedPattern{
			Name:       "Falling Wedge",
			Type:       "reversal",
			Direction:  "bullish",
			Confidence: 0.7,
			StartTime:  last3Highs[0].Time,
			EndTime:    candles[len(candles)-1].CloseTime,
		}
	}

	return DetectedPattern{}
}

// Helper functions

type PricePoint struct {
	Time  int64
	Price float64
}

func findLocalMaxima(candles []models.Candle) []PricePoint {
	var maxima []PricePoint
	window := 5

	for i := window; i < len(candles)-window; i++ {
		isMax := true
		for j := i - window; j <= i+window; j++ {
			if j != i && candles[j].High >= candles[i].High {
				isMax = false
				break
			}
		}
		if isMax {
			maxima = append(maxima, PricePoint{
				Time:  candles[i].OpenTime,
				Price: candles[i].High,
			})
		}
	}

	return maxima
}

func findLocalMinima(candles []models.Candle) []PricePoint {
	var minima []PricePoint
	window := 5

	for i := window; i < len(candles)-window; i++ {
		isMin := true
		for j := i - window; j <= i+window; j++ {
			if j != i && candles[j].Low <= candles[i].Low {
				isMin = false
				break
			}
		}
		if isMin {
			minima = append(minima, PricePoint{
				Time:  candles[i].OpenTime,
				Price: candles[i].Low,
			})
		}
	}

	return minima
}

func meanOfPrices(points []PricePoint) float64 {
	var sum float64
	for _, p := range points {
		sum += p.Price
	}
	return sum / float64(len(points))
}

func varianceOf(points []PricePoint) float64 {
	if len(points) < 2 {
		return 0
	}

	mean := meanOfPrices(points)
	var sumSq float64
	for _, p := range points {
		diff := p.Price - mean
		sumSq += diff * diff
	}

	return sumSq / float64(len(points))
}

// GetLearnedPatterns returns all pattern records from database
func (s *Service) GetLearnedPatterns() ([]models.Pattern, error) {
	var patterns []models.Pattern
	err := s.db.Order("created_at DESC").Limit(100).Find(&patterns).Error
	return patterns, err
}