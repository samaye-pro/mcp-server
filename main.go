package main

import (
        "encoding/json"
        "fmt"
        "log"
        "net/http"

        "github.com/gorilla/websocket"
)

type MCPRequest struct {
        ID     string          `json:"id"`
        Method string          `json:"method"`
        Params json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
        ID     string      `json:"id"`
        Result interface{} `json:"result,omitempty"`
        Error  *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
        Code    int    `json:"code"`
        Message string `json:"message"`
}

type InitializeParams struct {
        ClientInfo map[string]interface{} `json:"clientInfo,omitempty"`
}

type ToolCallParams struct {
        Name      string                 `json:"name"`
        Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type Ticket struct {
        ID     string `json:"id"`
        Title  string `json:"title"`
        Status string `json:"status"`
}

type TicketsResponse struct {
        Tickets []Ticket `json:"tickets"`
}

var upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
                return true
        },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
                log.Printf("WebSocket upgrade error: %v", err)
                return
        }
        defer conn.Close()

        log.Println("Client connected")

        for {
                _, message, err := conn.ReadMessage()
                if err != nil {
                        log.Printf("Read error: %v", err)
                        break
                }

                var req MCPRequest
                if err := json.Unmarshal(message, &req); err != nil {
                        log.Printf("JSON unmarshal error: %v", err)
                        sendError(conn, "", -32700, "Parse error")
                        continue
                }

                log.Printf("Received request: method=%s, id=%s", req.Method, req.ID)

                response := handleRequest(req)
                
                responseBytes, err := json.Marshal(response)
                if err != nil {
                        log.Printf("JSON marshal error: %v", err)
                        continue
                }

                if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
                        log.Printf("Write error: %v", err)
                        break
                }

                log.Printf("Sent response for id=%s", req.ID)
        }

        log.Println("Client disconnected")
}

func handleRequest(req MCPRequest) MCPResponse {
        switch req.Method {
        case "initialize":
                return handleInitialize(req)
        case "tools/list":
                return handleToolsList(req)
        case "tools/call":
                return handleToolCall(req)
        default:
                return MCPResponse{
                        ID: req.ID,
                        Error: &MCPError{
                                Code:    -32601,
                                Message: fmt.Sprintf("Method not found: %s", req.Method),
                        },
                }
        }
}

func handleInitialize(req MCPRequest) MCPResponse {
        return MCPResponse{
                ID: req.ID,
                Result: map[string]interface{}{
                        "protocolVersion": "1.0",
                        "serverInfo": map[string]interface{}{
                                "name":    "go-mcp-demo",
                                "version": "1.0.0",
                        },
                        "capabilities": map[string]interface{}{
                                "tools": map[string]interface{}{
                                        "call": map[string]interface{}{
                                                "enabled": true,
                                        },
                                        "list": map[string]interface{}{
                                                "enabled":     true,
                                                "listChanged": false,
                                        },
                                },
                        },
                },
        }
}

func handleToolsList(req MCPRequest) MCPResponse {
        tools := []map[string]interface{}{
                {
                        "name":        "get_pending_tickets",
                        "description": "Returns a list of pending tickets",
                        "inputSchema": map[string]interface{}{
                                "type":       "object",
                                "properties": map[string]interface{}{},
                        },
                },
                {
                        "name":        "get_done_tickets",
                        "description": "Returns a list of completed tickets",
                        "inputSchema": map[string]interface{}{
                                "type":       "object",
                                "properties": map[string]interface{}{},
                        },
                },
                {
                        "name":        "get_todo_tickets",
                        "description": "Returns a list of todo tickets",
                        "inputSchema": map[string]interface{}{
                                "type":       "object",
                                "properties": map[string]interface{}{},
                        },
                },
        }

        return MCPResponse{
                ID: req.ID,
                Result: map[string]interface{}{
                        "tools": tools,
                },
        }
}

func handleToolCall(req MCPRequest) MCPResponse {
        var params ToolCallParams
        if err := json.Unmarshal(req.Params, &params); err != nil {
                return MCPResponse{
                        ID: req.ID,
                        Error: &MCPError{
                                Code:    -32602,
                                Message: "Invalid params",
                        },
                }
        }

        switch params.Name {
        case "get_pending_tickets":
                return MCPResponse{
                        ID: req.ID,
                        Result: TicketsResponse{
                                Tickets: []Ticket{
                                        {ID: "T1", Title: "Fix login bug", Status: "pending"},
                                        {ID: "T2", Title: "Database indexing", Status: "pending"},
                                },
                        },
                }
        case "get_done_tickets":
                return MCPResponse{
                        ID: req.ID,
                        Result: TicketsResponse{
                                Tickets: []Ticket{
                                        {ID: "T10", Title: "Payment integration", Status: "done"},
                                        {ID: "T11", Title: "Email system", Status: "done"},
                                },
                        },
                }
        case "get_todo_tickets":
                return MCPResponse{
                        ID: req.ID,
                        Result: TicketsResponse{
                                Tickets: []Ticket{
                                        {ID: "T20", Title: "Create dashboard UI", Status: "todo"},
                                        {ID: "T21", Title: "Add search filter", Status: "todo"},
                                },
                        },
                }
        default:
                return MCPResponse{
                        ID: req.ID,
                        Error: &MCPError{
                                Code:    -32602,
                                Message: fmt.Sprintf("Unknown tool: %s", params.Name),
                        },
                }
        }
}

func sendError(conn *websocket.Conn, id string, code int, message string) {
        response := MCPResponse{
                ID: id,
                Error: &MCPError{
                        Code:    code,
                        Message: message,
                },
        }
        responseBytes, _ := json.Marshal(response)
        conn.WriteMessage(websocket.TextMessage, responseBytes)
}

func main() {
        http.HandleFunc("/ws", handleWebSocket)
        
        fmt.Println("MCP Server running on ws://localhost:8080/ws")
        log.Fatal(http.ListenAndServe(":8080", nil))
}
