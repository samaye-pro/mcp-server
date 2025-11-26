// main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ---------- MCP message structs ----------
type MCPRequest struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method,omitempty"`
	Result map[string]interface{} `json:"result,omitempty"`
	Error  map[string]interface{} `json:"error,omitempty"`
}

type Ticket struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"` // pending | todo | done
}

// ---------- Dummy ticket store ----------
var tickets = []Ticket{
	{ID: "T1", Title: "Fix login bug", Status: "pending"},
	{ID: "T2", Title: "Database indexing", Status: "pending"},
	{ID: "T10", Title: "Payment integration", Status: "done"},
	{ID: "T11", Title: "Email system", Status: "done"},
	{ID: "T20", Title: "Create dashboard UI", Status: "todo"},
	{ID: "T21", Title: "Add search filter", Status: "todo"},
}

// ---------- helpers ----------
func ticketsByStatus(status string) []Ticket {
	out := make([]Ticket, 0)
	for _, t := range tickets {
		if t.Status == status {
			out = append(out, t)
		}
	}
	return out
}

func writeJSONConn(conn *websocket.Conn, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.TextMessage, b)
}

// ---------- WebSocket handler ----------
var upgrader = websocket.Upgrader{
	// Allow all origins for testing / demo purposes.
	CheckOrigin: func(r *http.Request) bool { return true },
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	log.Printf("client connected: %s\n", clientAddr)

	// loop read messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("client disconnected: %s\n", clientAddr)
			} else {
				log.Printf("read error: %v\n", err)
			}
			return
		}

		var req MCPRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("invalid json from %s: %v\n", clientAddr, err)
			// reply with error
			_ = writeJSONConn(conn, MCPResponse{
				ID: req.ID,
				Error: map[string]interface{}{
					"code":    "INVALID_JSON",
					"message": err.Error(),
				},
			})
			continue
		}

		log.Printf("recv from %s: method=%s id=%s params=%v\n", clientAddr, req.Method, req.ID, req.Params)

		switch req.Method {
		case "initialize":
			// reply with initialized and tools
			result := map[string]interface{}{
				"status": "initialized",
				"server": "go-mcp-tickets",
				"tools": []map[string]interface{}{
					{"name": "get_pending_tickets", "description": "Get pending tickets"},
					{"name": "get_done_tickets", "description": "Get done tickets"},
					{"name": "get_todo_tickets", "description": "Get todo tickets"},
				},
			}
			resp := MCPResponse{
				ID:     req.ID,
				Method: "initialized",
				Result: result,
			}
			if err := writeJSONConn(conn, resp); err != nil {
				log.Printf("write error: %v\n", err)
				return
			}

		case "tools/call":
			// params should include { "name": "<toolName>", "arguments": {...} }
			nameAny, ok := req.Params["name"]
			if !ok {
				_ = writeJSONConn(conn, MCPResponse{
					ID: req.ID,
					Error: map[string]interface{}{
						"code":    "MISSING_TOOL_NAME",
						"message": "params.name is required",
					},
				})
				continue
			}
			toolName, _ := nameAny.(string)
			// optional arguments
			var args map[string]interface{}
			if a, ok := req.Params["arguments"]; ok {
				if m, ok := a.(map[string]interface{}); ok {
					args = m
				}
			}

			// Execute tool
			switch toolName {
			case "get_pending_tickets":
				resp := MCPResponse{
					ID: req.ID,
					Result: map[string]interface{}{
						"tickets": ticketsByStatus("pending"),
						"meta": map[string]interface{}{
							"count": len(ticketsByStatus("pending")),
							"args":  args,
						},
					},
				}
				if err := writeJSONConn(conn, resp); err != nil {
					log.Printf("write error: %v\n", err)
					return
				}

			case "get_done_tickets":
				resp := MCPResponse{
					ID: req.ID,
					Result: map[string]interface{}{
						"tickets": ticketsByStatus("done"),
						"meta": map[string]interface{}{
							"count": len(ticketsByStatus("done")),
							"args":  args,
						},
					},
				}
				if err := writeJSONConn(conn, resp); err != nil {
					log.Printf("write error: %v\n", err)
					return
				}

			case "get_todo_tickets":
				resp := MCPResponse{
					ID: req.ID,
					Result: map[string]interface{}{
						"tickets": ticketsByStatus("todo"),
						"meta": map[string]interface{}{
							"count": len(ticketsByStatus("todo")),
							"args":  args,
						},
					},
				}
				if err := writeJSONConn(conn, resp); err != nil {
					log.Printf("write error: %v\n", err)
					return
				}

			default:
				_ = writeJSONConn(conn, MCPResponse{
					ID: req.ID,
					Error: map[string]interface{}{
						"code":    "TOOL_NOT_FOUND",
						"message": fmt.Sprintf("Unknown tool: %s", toolName),
					},
				})
			}

		case "ping":
			_ = writeJSONConn(conn, MCPResponse{
				ID: req.ID,
				Result: map[string]interface{}{
					"pong": time.Now().UTC().Format(time.RFC3339),
				},
			})

		default:
			_ = writeJSONConn(conn, MCPResponse{
				ID: req.ID,
				Error: map[string]interface{}{
					"code":    "UNKNOWN_METHOD",
					"message": fmt.Sprintf("Method not supported: %s", req.Method),
				},
			})
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)

	log.Println("MCP WebSocket Ticket Server running on :8080/ws (bind 0.0.0.0:8080)")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatalf("server failed: %v\n", err)
	}
}
