# AI Trading System - Complete Project

A comprehensive real-time trading system with AI-powered analysis, pattern recognition, and 20-year backtesting for BTC and Gold.

## Project Structure

```
trading_system/
├── backend/                    # Go backend (MVC structure)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go         # Entry point
│   ├── internal/
│   │   ├── config/             # Configuration
│   │   ├── models/             # Database models
│   │   ├── services/           # Business logic
│   │   │   ├── binance/        # Binance API integration
│   │   │   ├── ai_brain/       # AI decision engine
│   │   │   ├── pattern/        # Pattern recognition
│   │   │   └── backtest/       # 20-year backtesting
│   │   ├── handlers/           # HTTP handlers
│   │   └── middleware/         # CORS, rate limiting
│   └── Dockerfile
│
├── frontend/                   # Lit/Vanilla JS frontend
│   ├── index.html              # Main HTML with embedded CSS
│   ├── src/
│   │   └── main.js             # Frontend logic
│   ├── package.json
│   ├── vite.config.js
│   └── Dockerfile
│
└── docker/
    ├── docker-compose.yml      # Docker orchestration
    ├── .env                    # Environment variables
    └── data/                   # Database storage
```

## Features

### 1. Real-time Market Data (millisecond updates)
- Live BTC/USDT and XAU/USDT prices
- Candlestick charts (TradingView-style)
- Order book visualization
- WebSocket for real-time updates

### 2. AI Brain - Decision Engine
- Technical indicators: RSI, MACD, EMA (20/50/200), Bollinger Bands, ATR
- Multi-factor decision making
- Trade level calculation (entry, target, stop-loss)
- Confidence scoring

### 3. Pattern Recognition
- **Candlestick Patterns**: Doji, Hammer, Shooting Star, Engulfing, Morning/Evening Star
- **Chart Patterns**: Head & Shoulders, Double Top/Bottom, Triangles, Wedges
- Pattern confidence scoring

### 4. 20-Year Backtesting
- Historical data from Binance (up to 20 years)
- Support for BTC/USDT and XAU/USDT
- Multiple timeframes (1m, 5m, 15m, 1H, 4H, 1D)
- Performance metrics: Return, Win Rate, Drawdown, Sharpe Ratio

### 5. Docker Deployment
- Containerized backend and frontend
- Automatic API key injection
- Production-ready configuration

## Quick Start

### Option 1: Docker (Recommended)

```bash
cd docker
docker-compose up -d
```

Access:
- Frontend: http://localhost:3000
- API: http://localhost:8080

### Option 2: Manual Running

**Backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/ticker/:symbol` | 24h ticker data |
| `GET /api/v1/klines` | Candlestick data |
| `GET /api/v1/orderbook/:symbol` | Order book |
| `GET /api/v1/ai/prediction/:symbol` | AI prediction |
| `GET /api/v1/ai/analysis/:symbol` | Full AI analysis |
| `POST /api/v1/backtest` | Run backtest |
| `GET /api/v1/patterns/:symbol` | Detected patterns |

## WebSocket

Connect to `ws://localhost:8080/ws` for real-time:
- Ticker updates
- AI analysis updates

## Technical Stack

- **Backend**: Go 1.21, Gin framework, GORM (SQLite)
- **Frontend**: Vanilla JS, Lightweight Charts, Vite
- **Database**: SQLite (file-based)
- **Docker**: Docker Compose

## Configuration

Edit `.env` file in `docker/` folder:
- `BINANCE_API_KEY`: Your Binance API key
- `BINANCE_SECRET_KEY`: Your Binance secret

## Security Notes

⚠️ Your API keys have:
- Unrestricted IP access
- Withdrawal permissions enabled

**Recommendation**: Create new keys with:
- IP restrictions (your server IP only)
- Only enable Reading + Spot Trading (disable Withdrawal)

## Performance

- Millisecond-level data updates via WebSocket
- 20-year historical data for backtesting
- Real-time pattern detection
- AI analysis refreshes every minute

## Future Enhancements

1. Machine Learning model integration (TensorFlow/PyTorch)
2. More indicators (Stochastic, Ichimoku, VWAP)
3. Paper trading mode
4. Telegram/Discord notifications
5. Multiple strategy support
6. Portfolio management

## License

MIT License