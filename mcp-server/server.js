import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolSchema,
  ListResourcesSchema,
  ListToolsSchema,
  ReadResourceSchema,
} from '@modelcontextprotocol/sdk/types.js';

const API_BASE = process.env.TRADING_API_URL || 'http://localhost:8080';

async function callTradingAPI(endpoint, method = 'GET', body = null) {
  const url = `${API_BASE}${endpoint}`;
  const options = {
    method,
    headers: { 'Content-Type': 'application/json' },
  };
  if (body) options.body = JSON.stringify(body);

  try {
    const response = await fetch(url, options);
    return await response.json();
  } catch (error) {
    return { error: error.message };
  }
}

const tools = [
  {
    name: 'get_ticker',
    description: 'Get 24h ticker data for a trading symbol (BTCUSDT, XAUUSDT, etc.)',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
      },
    },
  },
  {
    name: 'get_klines',
    description: 'Get candlestick/kline data for a symbol',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
        interval: { type: 'string', default: '1m', description: 'Time interval (1m, 5m, 15m, 1h, 4h, 1d)' },
        limit: { type: 'number', default: 500, description: 'Number of candles' },
      },
    },
  },
  {
    name: 'get_orderbook',
    description: 'Get order book data for a symbol',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
        limit: { type: 'number', default: 20, description: 'Order book depth' },
      },
    },
  },
  {
    name: 'get_prediction',
    description: 'Get AI prediction for a symbol',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
      },
    },
  },
  {
    name: 'get_analysis',
    description: 'Get full AI analysis including predictions, indicators, and candles',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
      },
    },
  },
  {
    name: 'run_backtest',
    description: 'Run backtest simulation with historical data',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
        timeframe: { type: 'string', default: '1h', description: 'Time interval' },
        start_date: { type: 'string', description: 'Start date (ISO)' },
        end_date: { type: 'string', description: 'End date (ISO)' },
        initial_capital: { type: 'number', default: 10000 },
        commission: { type: 'number', default: 0.001 },
        slippage: { type: 'number', default: 0.0005 },
      },
    },
  },
  {
    name: 'get_patterns',
    description: 'Get detected candlestick and chart patterns for a symbol',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
      },
    },
  },
  {
    name: 'get_balance',
    description: 'Get account balance from exchange',
    inputSchema: { type: 'object', properties: {} },
  },
  {
    name: 'place_order',
    description: 'Place a trade order (WARNING: real trading)',
    inputSchema: {
      type: 'object',
      properties: {
        symbol: { type: 'string', default: 'BTCUSDT', description: 'Trading symbol' },
        side: { type: 'string', enum: ['buy', 'sell'], description: 'Order side' },
        type: { type: 'string', enum: ['market', 'limit'], default: 'market' },
        quantity: { type: 'number', description: 'Order quantity' },
      },
      required: ['side', 'quantity'],
    },
  },
];

const server = new Server(
  { name: 'zen-trading-mcp', version: '1.0.0' },
  { capabilities: { tools: {} } }
);

server.setRequestHandler(ListToolsSchema, async () => ({ tools }));

server.setRequestHandler(CallToolSchema, async ({ name, arguments: args }) => {
  try {
    let result;

    switch (name) {
      case 'get_ticker':
        result = await callTradingAPI(`/api/v1/ticker/${args.symbol || 'BTCUSDT'}`);
        break;

      case 'get_klines':
        const klineParams = new URLSearchParams({
          symbol: args.symbol || 'BTCUSDT',
          interval: args.interval || '1m',
          limit: args.limit || 500,
        });
        result = await callTradingAPI(`/api/v1/klines?${klineParams}`);
        break;

      case 'get_orderbook':
        result = await callTradingAPI(`/api/v1/orderbook/${args.symbol || 'BTCUSDT'}?limit=${args.limit || 20}`);
        break;

      case 'get_prediction':
        result = await callTradingAPI(`/api/v1/ai/prediction/${args.symbol || 'BTCUSDT'}`);
        break;

      case 'get_analysis':
        result = await callTradingAPI(`/api/v1/ai/analysis/${args.symbol || 'BTCUSDT'}`);
        break;

      case 'run_backtest':
        const backtestConfig = {
          symbol: args.symbol || 'BTCUSDT',
          timeframe: args.timeframe || '1h',
          start_date: args.start_date ? new Date(args.start_date).getTime() : new Date(Date.now() - 20 * 365 * 24 * 60 * 60 * 1000).getTime(),
          end_date: args.end_date ? new Date(args.end_date).getTime() : Date.now(),
          initial_capital: args.initial_capital || 10000,
          commission: args.commission || 0.001,
          slippage: args.slippage || 0.0005,
        };
        result = await callTradingAPI('/api/v1/backtest', 'POST', backtestConfig);
        break;

      case 'get_patterns':
        result = await callTradingAPI(`/api/v1/patterns/${args.symbol || 'BTCUSDT'}`);
        break;

      case 'get_balance':
        result = await callTradingAPI('/api/v1/balance');
        break;

      case 'place_order':
        result = await callTradingAPI('/api/v1/order', 'POST', args);
        break;

      default:
        result = { error: `Unknown tool: ${name}` };
    }

    return { content: [{ type: 'text', text: JSON.stringify(result, null, 2) }] };
  } catch (error) {
    return { content: [{ type: 'text', text: JSON.stringify({ error: error.message }) }] };
  }
});

server.setRequestHandler(ListResourcesSchema, async () => ({ resources: [] }));
server.setRequestHandler(ReadResourceSchema, async () => ({ contents: [] }));

const transport = new StdioServerTransport();
await server.connect(transport);
console.error('Zen Trading MCP Server running...');