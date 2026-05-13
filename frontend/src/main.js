// Trading System Frontend - Main Application
import { createChart, CrosshairMode } from 'lightweight-charts';

// Configuration
const API_BASE = 'http://localhost:8080/api/v1';
const WS_URL = 'ws://localhost:8080/ws';

// State
let currentSymbol = 'BTCUSDT';
let currentTimeframe = '1m';
let chart = null;
let candleSeries = null;
let volumeSeries = null;
let ws = null;

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  initChart();
  connectWebSocket();
  loadInitialData();
  setupEventListeners();
});

// Initialize TradingView-style chart
function initChart() {
  const chartContainer = document.getElementById('chart');

  chart = createChart(chartContainer, {
    layout: {
      background: { color: '#1a1a25' },
      textColor: '#a0a0b0',
    },
    grid: {
      vertLines: { color: '#2a2a3a' },
      horzLines: { color: '#2a2a3a' },
    },
    crosshair: {
      mode: CrosshairMode.Normal,
    },
    rightPriceScale: {
      borderColor: '#2a2a3a',
    },
    timeScale: {
      borderColor: '#2a2a3a',
      timeVisible: true,
      secondsVisible: false,
    },
  });

  // Candlestick series
  candleSeries = chart.addCandlestickSeries({
    upColor: '#00d4aa',
    downColor: '#ff4757',
    borderUpColor: '#00d4aa',
    borderDownColor: '#ff4757',
    wickUpColor: '#00d4aa',
    wickDownColor: '#ff4757',
  });

  // Volume series
  volumeSeries = chart.addHistogramSeries({
    color: '#3498ff',
    priceFormat: {
      type: 'volume',
    },
    priceScaleId: 'volume',
  });

  chart.priceScale('volume').applyOptions({
    scaleMargins: {
      top: 0.8,
      bottom: 0,
    },
  });

  // Handle resize
  window.addEventListener('resize', () => {
    chart.applyOptions({
      width: chartContainer.clientWidth,
      height: chartContainer.clientHeight,
    });
  });
}

// Connect to WebSocket for real-time updates
function connectWebSocket() {
  ws = new WebSocket(WS_URL);

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);

      if (data.type === 'ticker') {
        updateTickerData(data);
      } else if (data.type === 'analysis') {
        updateAnalysis(data);
      }
    } catch (e) {
      console.error('WebSocket message error:', e);
    }
  };

  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };

  ws.onclose = () => {
    console.log('WebSocket closed, reconnecting...');
    setTimeout(connectWebSocket, 5000);
  };
}

// Load initial data
async function loadInitialData() {
  await loadKlines();
  await loadAnalysis();
  await loadOrderBook();
}

// Load candlestick data from API
async function loadKlines() {
  try {
    const response = await fetch(
      `${API_BASE}/klines?symbol=${currentSymbol}&interval=${currentTimeframe}&limit=500`
    );
    const data = await response.json();

    if (data.candles) {
      const formattedCandles = data.candles.map(c => ({
        time: c.OpenTime / 1000,
        open: c.Open,
        high: c.High,
        low: c.Low,
        close: c.Close,
      }));

      candleSeries.setData(formattedCandles);

      // Format volume data
      const volumeData = data.candles.map(c => ({
        time: c.OpenTime / 1000,
        value: c.Volume,
        color: c.Close >= c.Open ? 'rgba(0, 212, 170, 0.5)' : 'rgba(255, 71, 87, 0.5)',
      }));

      volumeSeries.setData(volumeData);

      // Fit content
      chart.timeScale().fitContent();
    }
  } catch (error) {
    console.error('Failed to load klines:', error);
  }
}

// Load AI analysis
async function loadAnalysis() {
  try {
    const response = await fetch(`${API_BASE}/ai/analysis/${currentSymbol}`);
    const data = await response.json();

    if (data.analysis) {
      updateAnalysisPanel(data.analysis);
    }
  } catch (error) {
    console.error('Failed to load analysis:', error);
  }
}

// Update analysis panel
function updateAnalysisPanel(analysis) {
  document.getElementById('analysis-symbol').textContent = analysis.symbol.replace('USDT', '/USDT');

  const predictionEl = document.getElementById('ai-prediction');
  predictionEl.textContent = analysis.prediction?.toUpperCase() || '--';
  predictionEl.className = `prediction-value ${analysis.prediction || 'hold'}`;

  const confidence = (analysis.confidence * 100).toFixed(1);
  document.getElementById('ai-confidence').textContent = `${confidence}%`;
  document.getElementById('confidence-bar').style.width = `${confidence}%`;

  document.getElementById('entry-price').textContent = formatPrice(analysis.entry_price);
  document.getElementById('target-price').textContent = formatPrice(analysis.target_price);
  document.getElementById('stop-loss').textContent = formatPrice(analysis.stop_loss);
  document.getElementById('ai-reason').textContent = analysis.reason || 'No reason provided';

  // Update indicators
  if (analysis.indicators) {
    document.getElementById('rsi-value').textContent = analysis.indicators.rsi?.toFixed(2) || '--';
    document.getElementById('macd-value').textContent = analysis.indicators.macd_histogram?.toFixed(4) || '--';
    document.getElementById('ema20-value').textContent = formatPrice(analysis.indicators.ema_20);
    document.getElementById('ema50-value').textContent = formatPrice(analysis.indicators.ema_50);
  }

  // Update sidebar prediction
  if (currentSymbol === 'BTCUSDT') {
    document.getElementById('btc-signal').textContent = analysis.prediction?.toUpperCase() || '--';
    document.getElementById('btc-signal').className = `prediction-value ${analysis.prediction || 'hold'}`;
    document.getElementById('btc-confidence').textContent = `${confidence}%`;
  } else {
    document.getElementById('gold-signal').textContent = analysis.prediction?.toUpperCase() || '--';
    document.getElementById('gold-signal').className = `prediction-value ${analysis.prediction || 'hold'}`;
    document.getElementById('gold-confidence').textContent = `${confidence}%`;
  }
}

// Load order book
async function loadOrderBook() {
  try {
    const response = await fetch(`${API_BASE}/orderbook/${currentSymbol}?limit=10`);
    const data = await response.json();

    if (data.orderBook) {
      updateOrderBook(data.orderBook);
    }
  } catch (error) {
    console.error('Failed to load orderbook:', error);
  }
}

// Update order book display
function updateOrderBook(orderbook) {
  const asks = orderbook.Asks.slice(0, 5).reverse();
  const bids = orderbook.Bids.slice(0, 5);

  let html = '';

  asks.forEach(ask => {
    html += `<div class="orderbook-row">
      <span class="ask">${parseFloat(ask[0]).toFixed(2)}</span>
      <span>${parseFloat(ask[1]).toFixed(4)}</span>
    </div>`;
  });

  const midPrice = (parseFloat(asks[0]?.[0]) + parseFloat(bids[0]?.[0])) / 2;
  html += `<div style="text-align: center; padding: 8px 0; color: var(--text-primary); font-size: 16px; font-weight: 600;">
    ${midPrice.toFixed(2)}
  </div>`;

  bids.forEach(bid => {
    html += `<div class="orderbook-row">
      <span>${parseFloat(bid[1]).toFixed(4)}</span>
      <span class="bid">${parseFloat(bid[0]).toFixed(2)}</span>
    </div>`;
  });

  document.getElementById('orderbook').innerHTML = html;
}

// Update ticker data
function updateTickerData(data) {
  const isBTC = data.symbol === 'BTCUSDT';
  const priceEl = isBTC ? 'btc-price' : 'gold-price';
  const changeEl = isBTC ? 'btc-change' : 'gold-change';
  const priceSidebarEl = isBTC ? 'btc-price-sidebar' : 'gold-price-sidebar';
  const changeSidebarEl = isBTC ? 'btc-change-sidebar' : 'gold-change-sidebar';

  // Header
  document.getElementById(priceEl).textContent = formatPrice(data.price);

  const changeValue = parseFloat(data.change);
  const changeClass = changeValue >= 0 ? 'positive' : 'negative';
  const changePrefix = changeValue >= 0 ? '+' : '';

  document.getElementById(changeEl).textContent = `${changePrefix}${changeValue.toFixed(2)}%`;
  document.getElementById(changeEl).className = `stat-value ${changeClass}`;

  // Sidebar
  document.getElementById(priceSidebarEl).textContent = formatPrice(data.price);
  document.getElementById(changeSidebarEl).textContent = `${changePrefix}${changeValue.toFixed(2)}%`;
  document.getElementById(changeSidebarEl).className = `symbol-change ${changeClass}`;
}

// Update analysis from WebSocket
function updateAnalysis(data) {
  if (data.symbol === currentSymbol) {
    updateAnalysisPanel({
      symbol: data.symbol,
      prediction: data.prediction,
      confidence: data.confidence,
      entry_price: data.entry,
      target_price: data.target,
      stop_loss: data.stopLoss,
      reason: data.reason,
    });
  }

  // Update sidebar for both symbols
  const isBTC = data.symbol === 'BTCUSDT';
  if (isBTC) {
    document.getElementById('btc-signal').textContent = data.prediction?.toUpperCase() || '--';
    document.getElementById('btc-signal').className = `prediction-value ${data.prediction || 'hold'}`;
    document.getElementById('btc-confidence').textContent = `${(data.confidence * 100).toFixed(1)}%`;
  } else {
    document.getElementById('gold-signal').textContent = data.prediction?.toUpperCase() || '--';
    document.getElementById('gold-signal').className = `prediction-value ${data.prediction || 'hold'}`;
    document.getElementById('gold-confidence').textContent = `${(data.confidence * 100).toFixed(1)}%`;
  }
}

// Setup event listeners
function setupEventListeners() {
  // Symbol selection
  document.querySelectorAll('.symbol-item').forEach(item => {
    item.addEventListener('click', () => {
      document.querySelectorAll('.symbol-item').forEach(i => i.classList.remove('active'));
      item.classList.add('active');

      currentSymbol = item.dataset.symbol;
      loadKlines();
      loadAnalysis();
      loadOrderBook();
    });
  });

  // Timeframe selection
  document.querySelectorAll('.tf-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.tf-btn').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');

      currentTimeframe = btn.dataset.tf;
      loadKlines();
    });
  });

  // Run backtest
  document.getElementById('run-backtest').addEventListener('click', runBacktest);
}

// Run backtest
async function runBacktest() {
  const symbol = document.getElementById('backtest-symbol').value;
  const capital = document.getElementById('initial-capital').value;

  const now = new Date();
  const twentyYearsAgo = new Date(now.getFullYear() - 20, now.getMonth(), now.getDate());

  const config = {
    symbol: symbol,
    timeframe: '1h',
    start_date: twentyYearsAgo.getTime(),
    end_date: now.getTime(),
    initial_capital: parseFloat(capital),
    commission: 0.001,
    slippage: 0.0005,
  };

  document.getElementById('run-backtest').textContent = 'Running...';
  document.getElementById('run-backtest').disabled = true;

  try {
    const response = await fetch(`${API_BASE}/backtest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config),
    });

    const result = await response.json();

    // Display results
    document.getElementById('backtest-results').style.display = 'block';

    const returnEl = document.getElementById('bt-return');
    returnEl.textContent = `${result.total_return?.toFixed(2)}%`;
    returnEl.className = `result-value ${result.total_return >= 0 ? 'positive' : 'negative'}`;

    document.getElementById('bt-trades').textContent = result.total_trades;
    document.getElementById('bt-winrate').textContent = `${result.win_rate?.toFixed(1)}%`;
    document.getElementById('bt-drawdown').textContent = `${result.max_drawdown?.toFixed(2)}%`;
    document.getElementById('bt-sharpe').textContent = result.sharpe_ratio?.toFixed(2) || '--';

  } catch (error) {
    console.error('Backtest error:', error);
    alert('Failed to run backtest: ' + error.message);
  } finally {
    document.getElementById('run-backtest').textContent = 'Run Backtest';
    document.getElementById('run-backtest').disabled = false;
  }
}

// Helper functions
function formatPrice(price) {
  if (!price) return '--';
  if (price > 1000) return price.toFixed(2);
  if (price > 1) return price.toFixed(4);
  return price.toFixed(6);
}

// Add candle to chart (real-time update)
function addCandleToChart(candle) {
  candleSeries.update({
    time: candle.time,
    open: candle.open,
    high: candle.high,
    low: candle.low,
    close: candle.close,
  });
}