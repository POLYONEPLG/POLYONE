package main
 
import (
    "encoding/json"
    "log"
    "time"

    "github.com/gorilla/websocket"
)

// ClientMessageType defines the type of message received from clients.
type ClientMessageType string

const (
    SubscribeRequest    ClientMessageType = "subscribe"
    UnsubscribeRequest  ClientMessageType = "unsubscribe"
    AgentControlRequest ClientMessageType = "agent_control"
    TransactionQuery    ClientMessageType = "transaction_query"
    HeartbeatPong       ClientMessageType = "pong"
)

// ClientMessage represents the structure of a message received from a client.
type ClientMessage struct {
    Type    ClientMessageType `json:"type"`
    Payload interface{}       `json:"payload"`
}

// SubscribePayload defines the payload for subscription requests.
type SubscribePayload struct {
    Topic string `json:"topic"` // e.g., agent_id or tx_id
}

// AgentControlPayload defines the payload for agent control commands.
type AgentControlPayload struct {
    AgentID string `json:"agent_id"`
    Command string `json:"command"` // e.g., "start", "stop", "update_config"
    Params  map[string]interface{} `json:"params,omitempty"`
}

// TransactionQueryPayload defines the payload for transaction queries.
type TransactionQueryPayload struct {
    TxID       string `json:"tx_id,omitempty"`
    AgentID    string `json:"agent_id,omitempty"`
    Blockchain string `json:"blockchain,omitempty"` // e.g., "Solana"
    Limit      int    `json:"limit,omitempty"`      // Number of transactions to return
}

// ErrorResponse defines the structure for error messages sent to clients.
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// ResponseMessage defines the structure for server responses to clients.
type ResponseMessage struct {
    Type    string      `json:"type"`
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorResponse `json:"error,omitempty"`
}

// HandleClientMessage processes incoming messages from a client and dispatches to appropriate handlers.
func (s *WebSocketServer) HandleClientMessage(client *Client, message []byte) {
    var msg ClientMessage
    if err := json.Unmarshal(message, &msg); err != nil {
        log.Printf("Failed to unmarshal client message: %v", err)
        s.sendErrorToClient(client, 400, "Invalid message format")
        return
    }

    client.LastActive = time.Now()

    switch msg.Type {
    case SubscribeRequest:
        s.handleSubscribe(client, msg.Payload)
    case UnsubscribeRequest:
        s.handleUnsubscribe(client, msg.Payload)
    case AgentControlRequest:
        s.handleAgentControl(client, msg.Payload)
    case TransactionQuery:
        s.handleTransactionQuery(client, msg.Payload)
    case HeartbeatPong:
        // Heartbeat pong is handled in the readPump; no additional action needed here
        log.Printf("Received pong from client")
    default:
        log.Printf("Unknown message type received: %s", msg.Type)
        s.sendErrorToClient(client, 400, "Unknown message type")
    }
}

// handleSubscribe processes a subscription request from a client.
func (s *WebSocketServer) handleSubscribe(client *Client, payload interface{}) {
    data, ok := payload.(map[string]interface{})
    if !ok {
        s.sendErrorToClient(client, 400, "Invalid subscribe payload")
        return
    }

    topic, ok := data["topic"].(string)
    if !ok || topic == "" {
        s.sendErrorToClient(client, 400, "Missing or invalid topic in subscribe request")
        return
    }

    s.Mutex.Lock()
    client.Topics[topic] = true
    s.Mutex.Unlock()

    log.Printf("Client subscribed to topic: %s", topic)
    response := ResponseMessage{
        Type:    "subscribe_response",
        Success: true,
        Data:    map[string]string{"topic": topic},
    }
    s.sendResponseToClient(client, response)
}

// handleUnsubscribe processes an unsubscription request from a client.
func (s *WebSocketServer) handleUnsubscribe(client *Client, payload interface{}) {
    data, ok := payload.(map[string]interface{})
    if !ok {
        s.sendErrorToClient(client, 400, "Invalid unsubscribe payload")
        return
    }

    topic, ok := data["topic"].(string)
    if !ok || topic == "" {
        s.sendErrorToClient(client, 400, "Missing or invalid topic in unsubscribe request")
        return
    }

    s.Mutex.Lock()
    delete(client.Topics, topic)
    s.Mutex.Unlock()

    log.Printf("Client unsubscribed from topic: %s", topic)
    response := ResponseMessage{
        Type:    "unsubscribe_response",
        Success: true,
        Data:    map[string]string{"topic": topic},
    }
    s.sendResponseToClient(client, response)
}

// handleAgentControl processes agent control commands from a client.
func (s *WebSocketServer) handleAgentControl(client *Client, payload interface{}) {
    data, ok := payload.(map[string]interface{})
    if !ok {
        s.sendErrorToClient(client, 400, "Invalid agent control payload")
        return
    }

    agentID, ok := data["agent_id"].(string)
    if !ok || agentID == "" {
        s.sendErrorToClient(client, 400, "Missing or invalid agent_id in control request")
        return
    }

    command, ok := data["command"].(string)
    if !ok || command == "" {
        s.sendErrorToClient(client, 400, "Missing or invalid command in control request")
        return
    }

    // Simulate processing the command (replace with actual logic for agent control)
    log.Printf("Processing agent control command: %s for agent: %s", command, agentID)
    success := true
    errorMsg := ""
    responseData := map[string]interface{}{
        "agent_id": agentID,
        "command":  command,
    }

    switch command {
    case "start":
        // Placeholder: Start agent logic
        log.Printf("Starting agent %s", agentID)
        responseData["status"] = "started"
    case "stop":
        // Placeholder: Stop agent logic
        log.Printf("Stopping agent %s", agentID)
        responseData["status"] = "stopped"
    case "update_config":
        // Placeholder: Update agent configuration
        params, _ := data["params"].(map[string]interface{})
        log.Printf("Updating config for agent %s with params: %v", agentID, params)
        responseData["status"] = "config_updated"
    default:
        success = false
        errorMsg = "Unsupported command"
        log.Printf("Unsupported command: %s for agent: %s", command, agentID)
    }

    if success {
        // Broadcast an agent status update (optional, based on your use case)
        s.SendAgentStatusUpdate(agentID, responseData["status"].(string), "Command processed")
        response := ResponseMessage{
            Type:    "agent_control_response",
            Success: true,
            Data:    responseData,
        }
        s.sendResponseToClient(client, response)
    } else {
        s.sendErrorToClient(client, 400, errorMsg)
    }
}

// handleTransactionQuery processes transaction query requests from a client.
func (s *WebSocketServer) handleTransactionQuery(client *Client, payload interface{}) {
    data, ok := payload.(map[string]interface{})
    if !ok {
        s.sendErrorToClient(client, 400, "Invalid transaction query payload")
        return
    }

    txID, _ := data["tx_id"].(string)
    agentID, _ := data["agent_id"].(string)
    blockchain, _ := data["blockchain"].(string)
    limitFloat, _ := data["limit"].(float64)
    limit := int(limitFloat)
    if limit <= 0 {
        limit = 10 // Default limit if not specified or invalid
    }

    // Validate input
    if txID == "" && agentID == "" {
        s.sendErrorToClient(client, 400, "Must provide tx_id or agent_id for transaction query")
        return
    }

    // Simulate fetching transaction data (replace with actual blockchain query logic)
    log.Printf("Querying transactions for tx_id: %s, agent_id: %s, blockchain: %s, limit: %d", txID, agentID, blockchain, limit)
    mockTransactions := []TransactionPayload{}
    if txID != "" {
        mockTransactions = append(mockTransactions, TransactionPayload{
            TxID:        txID,
            Status:      "confirmed",
            Timestamp:   time.Now().Add(-10 * time.Minute),
            Amount:      "0.5 SOL",
            Blockchain:  "Solana",
            FromAddress: "addr1",
            ToAddress:   "addr2",
        })
    } else if agentID != "" {
        for i := 0; i < limit && i < 3; i++ {
            mockTransactions = append(mockTransactions, TransactionPayload{
                TxID:        "tx-" + agentID + "-" + string(rune(i)),
                Status:      "confirmed",
                Timestamp:   time.Now().Add(time.Duration(-i-1) * time.Hour),
                Amount:      "0.1 SOL",
                Blockchain:  "Solana",
                FromAddress: "addr1",
                ToAddress:   "addr2",
            })
        }
    }

    response := ResponseMessage{
        Type:    "transaction_query_response",
        Success: true,
        Data:    map[string]interface{}{
            "transactions": mockTransactions,
            "count":        len(mockTransactions),
        },
    }
    s.sendResponseToClient(client, response)
    log.Printf("Sent transaction query response with %d transactions", len(mockTransactions))
}

// sendResponseToClient sends a success response to the client.
func (s *WebSocketServer) sendResponseToClient(client *Client, response ResponseMessage) {
    jsonData, err := json.Marshal(response)
    if err != nil {
        log.Printf("Failed to marshal response: %v", err)
        return
    }

    if err := client.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
        log.Printf("Failed to send response to client: %v", err)
    }
}

// sendErrorToClient sends an error response to the client.
func (s *WebSocketServer) sendErrorToClient(client *Client, code int, message string) {
    response := ResponseMessage{
        Type:    "error",
        Success: false,
        Error: &ErrorResponse{
            Code:    code,
            Message: message,
        },
    }
    jsonData, err := json.Marshal(response)
    if err != nil {
        log.Printf("Failed to marshal error response: %v", err)
        return
    }

    if err := client.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
        log.Printf("Failed to send error response to client: %v", err)
    }
}
