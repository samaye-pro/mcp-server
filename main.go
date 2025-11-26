package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// ----------------- MCP Protocol Structs -----------------

type MCPRequest struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type MCPResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type Ticket struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"` // pending | todo | done
}

// ---------------- Dummy Data ----------------

var tickets = []Ticket{
	{ID: 1, Title: "Fix login bug", Status: "pending"},
	{ID: 2, Title: "Add chat feature", Status: "todo"},
	{ID: 3, Title: "Improve UI", Status: "done"},
	{ID: 4, Title: "Email template update", Status: "pending"},
	{ID: 5, Title: "Refactor API", Status: "done"},
}

// ---------------- Ticket Tools ----------------

func getTicketsByStatus(status string) []Ticket {
	var result []Ticket
	for _, t := range tickets {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

// ---------------- MCP Handler ----------------

func mcpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, req.ID, "invalid_json", err.Error())
		return
	}

	switch req.Method {

	case "tool/get_pending_tickets":
		writeResponse(w, req.ID, getTicketsByStatus("pending"))

	case "tool/get_done_tickets":
		writeResponse(w, req.ID, getTicketsByStatus("done"))

	case "tool/get_todo_tickets":
		writeResponse(w, req.ID, getTicketsByStatus("todo"))

	default:
		writeError(w, req.ID, "unknown_method", "Method not supported")
	}
}

func writeResponse(w http.ResponseWriter, id string, result interface{}) {
	resp := MCPResponse{
		ID:     id,
		Result: result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, id string, code string, message string) {
	resp := MCPResponse{
		ID: id,
		Error: map[string]string{
			"code":    code,
			"message": message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ---------------- Main ----------------

func main() {
	http.HandleFunc("/mcp", mcpHandler)

	log.Println("MCP Ticket Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
