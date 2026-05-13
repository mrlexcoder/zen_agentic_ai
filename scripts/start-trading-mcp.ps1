# Zen Trading MCP Server Startup Script

# Start the Go backend in background
Write-Host "Starting Go Backend..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd 'd:\Startup_media\Ai\zen_agentic_ai\backend'; go run cmd/server/main.go"

# Wait a moment for backend to start
Start-Sleep -Seconds 3

# Start MCP server
Write-Host "Starting Trading MCP Server..." -ForegroundColor Cyan
node "d:\Startup_media\Ai\zen_agentic_ai\mcp-server\server.js"