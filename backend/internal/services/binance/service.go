package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"trading-system/backend/internal/models"

	"github.com/jinzhu/gorm"
)

type Service struct {
	apiKey    string
	secretKey string
	client    *http.Client
	baseURL   string
	db        *gorm.DB
}

type TickerResponse struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	LastPrice          string `json:"lastPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
}

type KlineResponse struct {
	OpenTime        int64
	Open            string
	High            string
	Low             string
	Close           string
	Volume          string
	CloseTime       int64
	QuoteVolume     string
	NumTrades       int
}

type OrderBook struct {
	LastUpdateId int64     `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

func NewService(apiKey, secretKey string) *Service {
	return &Service{
		apiKey:    apiKey,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
		baseURL:   "https://api.binance.com",
	}
}

func (s *Service) SetDB(db *gorm.DB) {
	s.db = db
}

// Generate signature for authenticated requests
func (s *Service) generateSignature(queryString string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// Make authenticated request
func (s *Service) signedRequest(method, endpoint string, params map[string]string) ([]byte, error) {
	// Sort params
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var queryString string
	for i, k := range keys {
		if i > 0 {
			queryString += "&"
		}
		queryString += k + "=" + params[k]
	}

	queryString += "&timestamp=" + strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	signature := s.generateSignature(queryString)
	queryString += "&signature=" + signature

	url := s.baseURL + endpoint + "?" + queryString

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-MBX-APIKEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Get24hTicker returns 24h ticker data
func (s *Service) Get24hTicker(symbol string) (*TickerResponse, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/24hr?symbol=%s", s.baseURL, symbol)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ticker TickerResponse
	if err := json.NewDecoder(resp.Body).Decode(&ticker); err != nil {
		return nil, err
	}

	return &ticker, nil
}

// GetKlines returns candlestick data - optimized for milliseconds
func (s *Service) GetKlines(symbol, interval string, limit int) ([]models.Candle, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d", s.baseURL, symbol, interval, limit)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawData [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
		return nil, err
	}

	var candles []models.Candle
	for _, d := range rawData {
		openTime := int64(d[0].(float64))
		candle := models.Candle{
			Symbol:    symbol,
			OpenTime:  openTime,
			Open:      parseFloat(d[1]),
			High:      parseFloat(d[2]),
			Low:       parseFloat(d[3]),
			Close:     parseFloat(d[4]),
			Volume:    parseFloat(d[5]),
			CloseTime: int64(d[6].(float64)),
		}
		candles = append(candles, candle)
	}

	return candles, nil
}

// GetHistoricalData fetches historical data for backtesting (up to 20 years)
func (s *Service) GetHistoricalData(symbol, interval string, startTime, endTime int64) ([]models.Candle, error) {
	var allCandles []models.Candle
	start := startTime

	for start < endTime {
		url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=1000",
			s.baseURL, symbol, interval, start, endTime)

		resp, err := s.client.Get(url)
		if err != nil {
			return nil, err
		}

		var rawData [][]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&rawData); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(rawData) == 0 {
			break
		}

		for _, d := range rawData {
			openTime := int64(d[0].(float64))
			candle := models.Candle{
				Symbol:    symbol,
				OpenTime:  openTime,
				Open:      parseFloat(d[1]),
				High:      parseFloat(d[2]),
				Low:       parseFloat(d[3]),
				Close:     parseFloat(d[4]),
				Volume:    parseFloat(d[5]),
				CloseTime: int64(d[6].(float64)),
			}
			allCandles = append(allCandles, candle)
		}

		// Move to next batch
		if len(rawData) > 0 {
			lastTime := int64(rawData[len(rawData)-1][0].(float64))
			start = lastTime + 1
		} else {
			break
		}

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return allCandles, nil
}

// GetOrderBook returns current order book
func (s *Service) GetOrderBook(symbol string, limit int) (*OrderBook, error) {
	url := fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", s.baseURL, symbol, limit)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orderBook OrderBook
	if err := json.NewDecoder(resp.Body).Decode(&orderBook); err != nil {
		return nil, err
	}

	return &orderBook, nil
}

// GetAccountBalance returns account balance
func (s *Service) GetAccountBalance() (map[string]float64, error) {
	resp, err := s.signedRequest("GET", "/api/v3/account", map[string]string{})
	if err != nil {
		return nil, err
	}

	var account map[string]interface{}
	if err := json.Unmarshal(resp, &account); err != nil {
		return nil, err
	}

	balances := make(map[string]float64)
	if balancesRaw, ok := account["balances"].([]interface{}); ok {
		for _, b := range balancesRaw {
			bal := b.(map[string]interface{})
			free := parseFloatString(bal["free"].(string))
			locked := parseFloatString(bal["locked"].(string))
			if free+locked > 0 {
				balances[bal["asset"].(string)] = free + locked
			}
		}
	}

	return balances, nil
}

// PlaceOrder places a trade order
func (s *Service) PlaceOrder(symbol, side, orderType string, quantity float64) (map[string]interface{}, error) {
	params := map[string]string{
		"symbol":      symbol,
		"side":        strings.ToUpper(side),
		"type":        strings.ToUpper(orderType),
		"quantity":    fmt.Sprintf("%.8f", quantity),
		"timestamp":   strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
	}

	resp, err := s.signedRequest("POST", "/api/v3/order", params)
	if err != nil {
		return nil, err
	}

	var order map[string]interface{}
	if err := json.Unmarshal(resp, &order); err != nil {
		return nil, err
	}

	return order, nil
}

func parseFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func parseFloatString(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}