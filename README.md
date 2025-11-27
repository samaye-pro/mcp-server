# Overview

This project is a fully functional **Model Context Protocol (MCP) WebSocket server** written in Go. The server implements the MCP specification, providing bidirectional JSON messaging over WebSocket on port 8080.

The server exposes three ticket management tools that return hardcoded ticket data for demonstration purposes:
- **get_pending_tickets**: Returns pending tickets
- **get_done_tickets**: Returns completed tickets
- **get_todo_tickets**: Returns todo tickets

## Key Features

- MCP-compliant WebSocket server running on `ws://localhost:8080/ws`
- Proper MCP handshake with `initialize` method
- Tool discovery via `tools/list` method
- Tool execution via `tools/call` method
- Bidirectional JSON messaging with proper request/response ID tracking
- Error handling for invalid methods and unknown tools
- Three ticket management tools with predefined datasets

# User Preferences

Preferred communication style: Simple, everyday language.

# System Architecture

## Backend Architecture

**Language**: Go 1.24.4  
**Framework**: Native Go with gorilla/websocket library  
**Server Type**: WebSocket server (not HTTP REST)

### Core Components

1. **WebSocket Handler** (`handleWebSocket`):
    - Upgrades HTTP connections to WebSocket
    - Manages client connections and message routing
    - Handles JSON marshaling/unmarshaling
    - Maintains bidirectional communication loop

2. **Request Router** (`handleRequest`):
    - Routes incoming MCP requests by method name
    - Supports: `initialize`, `tools/list`, `tools/call`
    - Returns proper error responses for unknown methods

3. **Method Handlers**:
    - `handleInitialize`: Returns MCP-compliant server capabilities
    - `handleToolsList`: Returns available tool definitions
    - `handleToolCall`: Executes requested tools by name

### MCP Message Structures

- **MCPRequest**: Incoming request with id, method, and params
- **MCPResponse**: Outgoing response with id, result, and error
- **MCPError**: Error structure with code and message
- **ToolCallParams**: Parameters for tool execution (tool name and arguments)
- **InitializeParams**: Parameters for initialization handshake

### Tool Implementation

Each tool returns a fixed dataset of tickets:
- Tools are argument-free (no input parameters required)
- Response format: `{"tickets": [...]}`
- Each ticket has: id, title, status

## File Structure

```
.
├── main.go       # Complete MCP server implementation
├── go.mod        # Go module definition
├── go.sum        # Go dependency checksums
├── .gitignore    # Excludes build artifacts
└── replit.md     # This documentation file
```

## MCP Compliance

The server implements the following MCP protocol features:

### Initialize Response
```json
{
  "id": "request-id",
  "result": {
    "protocolVersion": "1.0",
    "serverInfo": {
      "name": "go-mcp-demo",
      "version": "1.0.0"
    },
    "capabilities": {
      "tools": {
        "call": {"enabled": true},
        "list": {"enabled": true, "listChanged": false}
      }
    }
  }
}
```

### Tools List Response
Returns three tool definitions with JSON Schema for inputs (all tools require no arguments).

### Tools Call Response
Returns ticket datasets based on the requested tool name.

# External Dependencies

## Go Modules
- **gorilla/websocket** (v1.5.3): WebSocket protocol implementation
    - Purpose: Handles WebSocket connection upgrade and message framing
    - Used for bidirectional communication with MCP clients

## Development Tools
- Go 1.24.4: Compiler and runtime
- Go modules: Dependency management

# Running the Server

The server is configured to run automatically via the "MCP Server" workflow:
- Command: `go run main.go`
- Port: 8080
- Endpoint: `/ws`
- Output: Console logs

When started, the server displays: `MCP Server running on ws://localhost:8080/ws`

# Future Enhancements

Potential improvements suggested by architectural review:
1. Add automated handshake and tool-call tests
2. Document tool payloads and expected datasets for consumers
3. Tighten WebSocket origin checks for production deployment
4. Add dynamic ticket management with CRUD operations
5. Implement persistent storage (database or file system)
6. Add authentication and authorization
