# Zen Agentic AI - Ruflo Integration Guide

## Overview

This project integrates **Ruflo** (multi-agent AI orchestration) with your **Go trading backend** to create an intelligent, self-learning trading system.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Ruflo (Orchestrator)                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │   98+       │  │   Swarm     │  │   MCP Server            │  │
│  │   Agents    │  │   Coord.    │  │   (zen-trading)         │  │
│  └─────────────┘  └─────────────┘  └───────────┬─────────────┘  │
│                                                  │                │
│  ┌──────────────────────────────────────────────┴─────────────┐  │
│  │              Self-Learning Memory (HNSW)                  │  │
│  └────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                          │
                          │ MCP Protocol
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Go Backend (Port 8080)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐   │
│  │  Binance    │  │  AI Brain   │  │   Pattern Recognition  │   │
│  │  Service    │  │  (RSI,MACD) │  │   (Doji, Hammer...)     │   │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘   │
│  ┌─────────────┐  ┌─────────────┐                                 │
│  │  Backtest   │  │  Database   │                                 │
│  │  (20 years) │  │  (SQLite)   │                                 │
│  └─────────────┘  └─────────────┘                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Quick Start

### 1. Start the Go Backend

```bash
cd zen_agentic_ai/backend
go run cmd/server/main.go
```

The API runs at `http://localhost:8080`

### 2. Start Ruflo

```bash
cd zen_agentic_ai

# Initialize memory database
npx ruflo memory init

# Start the daemon (background workers)
npx ruflo daemon start

# Add the trading MCP server
claude mcp add zen-trading -- node mcp-server/server.js
```

### 3. Use in Claude Code

Once configured, you can use these MCP tools directly:

```
// Get real-time ticker
Call: get_ticker { symbol: "BTCUSDT" }

// Get AI analysis
Call: get_analysis { symbol: "BTCUSDT" }

// Run backtest
Call: run_backtest { symbol: "BTCUSDT", timeframe: "1h", initial_capital: 10000 }

// Get patterns
Call: get_patterns { symbol: "BTCUSDT" }
```

## Available MCP Tools

| Tool | Description |
|------|-------------|
| `get_ticker` | 24h ticker data (price, volume, change %) |
| `get_klines` | Candlestick data (OHLCV) |
| `get_orderbook` | Order book depth |
| `get_prediction` | AI prediction for symbol |
| `get_analysis` | Full AI analysis with indicators |
| `run_backtest` | Run 20-year backtest simulation |
| `get_patterns` | Detected candlestick/chart patterns |
| `get_balance` | Account balance from exchange |
| `place_order` | Place trade order (⚠️ real trading) |

## Ruflo Features Enabled

### Swarm Coordination
- Multiple agents can analyze different symbols simultaneously
- Hierarchical topology for task distribution

### Memory & Learning
- Trading decisions stored in HNSW vector database
- Pattern recognition improves over time

### Hooks
- Auto-trigger analysis on price movements
- Alert workers on significant patterns

## Manual MCP Registration

```bash
# Add to Claude Code
claude mcp add zen-trading -- node d:/Startup_media/Ai/zen_agentic_ai/mcp-server/server.js

# Or in VS Code with Claude Code extension
# Add to settings.json mcpServers section
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TRADING_API_URL` | `http://localhost:8080` | Go backend URL |
| `CLAUDE_FLOW_MODE` | `v3` | Ruflo mode |
| `CLAUDE_FLOW_TOPOLOGY` | `hierarchical-mesh` | Swarm topology |

## Security Notes

⚠️ **Trading Safety**
- `place_order` tool executes REAL trades
- Use paper trading mode for testing
- Keep API keys secure (in `.env`, never commit)

## Next Steps

1. Install `ruflo-neural-trader` plugin for advanced AI trading
2. Configure federation for multi-machine deployment
3. Set up Telegram/Discord alerts via hooks