package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"othello-backend/game"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WS message types
type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Client represents a single WS connection.
type Client struct {
	Conn    *websocket.Conn
	Session *game.Session
	Color   game.Player
	Send    chan []byte
	mu      sync.Mutex
}

// Hub manages all connected clients.
type Hub struct {
	Clients    map[string]*Client // session ID -> client
	mu         sync.RWMutex
	Manager    *game.Manager
	Register   chan *Client
	Unregister chan *Client
}

func NewHub(m *game.Manager) *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Manager:    m,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.Session.ID+"-"+client.Color.String()] = client
			h.mu.Unlock()

			// Send INIT message
			h.sendInit(client)

		case client := <-h.Unregister:
			key := client.Session.ID + "-" + client.Color.String()
			h.mu.Lock()
			delete(h.Clients, key)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) sendInit(client *Client) {
	s := client.Session
	data := map[string]any{
		"gameId":        s.ID,
		"board":         s.State.Board,
		"currentPlayer": int(s.State.CurrentPlayer),
		"size":          s.State.Size,
		"history":       s.State.History,
		"players": map[string]string{
			"BLACK": playerLabel(s, game.BLACK),
			"WHITE": playerLabel(s, game.WHITE),
		},
	}
	msg := WSMessage{Type: "INIT", Data: mustMarshal(data)}
	client.SendJSON(msg)
}

func playerLabel(s *game.Session, p game.Player) string {
	if label, ok := s.Players[p]; ok && label != "" {
		return label
	}
	if p == game.BLACK {
		return "黑棋"
	}
	return "白棋"
}

// Broadcast sends a message to all clients in a session except the sender.
func (h *Hub) Broadcast(session *game.GameState, sender *Client, msg WSMessage) {
	keyPrefix := session.Board[0][0] // dummy; use session ID
	_ = keyPrefix
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.Clients {
		if c.Session.ID == sender.Session.ID {
			c.SendJSON(msg)
		}
	}
}

func (c *Client) SendJSON(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	log.Printf("[WS SEND] remote=%s session=%s color=%s payload=%s",
		c.remoteAddr(), c.sessionID(), c.colorLabel(), string(data))
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Conn.WriteMessage(websocket.TextMessage, data)
}

func handleWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WS upgrade error:", err)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	// Read first message to determine game mode and join
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}
	log.Printf("[WS RECV] remote=%s session=%s color=%s payload=%s",
		conn.RemoteAddr().String(), "-", "-", string(msg))

	var wsMsg WSMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		conn.Close()
		return
	}

	if wsMsg.Type != "JOIN" {
		conn.WriteJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "First message must be JOIN"})})
		conn.Close()
		return
	}

	var joinData struct {
		Mode     string `json:"mode"`
		Color    string `json:"color"`
		Size     int    `json:"size"`
		GameID   string `json:"gameId,omitempty"`
	}
	if err := json.Unmarshal(wsMsg.Data, &joinData); err != nil {
		conn.Close()
		return
	}

	if joinData.Size != 6 && joinData.Size != 8 && joinData.Size != 10 {
		joinData.Size = 8
	}

	var color game.Player
	if joinData.Color == "WHITE" {
		color = game.WHITE
	} else {
		color = game.BLACK
	}

	mode := game.GameMode(joinData.Mode)
	if mode != game.ModePVE && mode != game.ModePVP {
		mode = game.ModePVE
	}

	var session *game.Session
	if mode == game.ModePVP {
		session = hub.Manager.JoinPvpSession(color, joinData.Size, joinData.GameID)
	} else {
		session = hub.Manager.CreateSession(mode, color, joinData.Size)
	}

	client.Session = session
	client.Color = color

	hub.Register <- client

	// Start reader and writer goroutines
	go client.writePump()
	go client.readPump(hub)

	// In PvE, AI may need to move first (e.g. player chose WHITE).
	go func() {
		client.Session.Mutex.Lock()
		defer client.Session.Mutex.Unlock()
		client.runAIMoveIfNeeded(hub)
	}()
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("[WS RECV] remote=%s session=%s color=%s payload=%s",
			c.remoteAddr(), c.sessionID(), c.colorLabel(), string(msg))

		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			continue
		}

		c.handleMessage(hub, wsMsg)
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()
	for data := range c.Send {
		c.Conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (c *Client) handleMessage(hub *Hub, msg WSMessage) {
	c.Session.Mutex.Lock()
	defer c.Session.Mutex.Unlock()

	if c.Session.State.GameOver {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Game is over"})})
		return
	}

	switch msg.Type {
	case "MOVE":
		c.handleMove(hub, msg.Data)
	case "UNDO":
		c.handleUndo(hub)
	case "PING":
		c.SendJSON(WSMessage{Type: "PONG"})
	}
}

func (c *Client) handleMove(hub *Hub, data json.RawMessage) {
	var moveData struct {
		R int `json:"r"`
		C int `json:"c"`
	}
	if err := json.Unmarshal(data, &moveData); err != nil {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Invalid move format"})})
		return
	}

	gs := c.Session.State
	if gs.CurrentPlayer != c.Color {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Not your turn"})})
		return
	}

	flips, ok := gs.DoMove(moveData.R, moveData.C, c.Color)
	if !ok {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Invalid move"})})
		return
	}

	// Send STATE back to all clients in session
	stateMsg := WSMessage{
		Type: "STATE",
		Data: mustMarshal(map[string]any{
			"board":         gs.Board,
			"currentPlayer": int(gs.CurrentPlayer),
			"lastMove":      map[string]int{"r": moveData.R, "c": moveData.C},
			"flipped":       flips,
			"history":       gs.History,
		}),
	}
	hub.Broadcast(gs, c, stateMsg)

	if c.resolvePassAndGameOver(hub) {
		return
	}

	c.runAIMoveIfNeeded(hub)
}

func (c *Client) handleUndo(hub *Hub) {
	if c.Session.Mode != game.ModePVE {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Undo only available in PvE"})})
		return
	}

	gs := c.Session.State
	if len(gs.History) == 0 {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Cannot undo"})})
		return
	}

	// Rebuild all prefix states and roll back to the previous "player turn" state.
	prefixStates := buildPrefixStates(gs.Size, gs.History)
	if len(prefixStates) == 0 {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Cannot undo"})})
		return
	}

	target := 0
	for i := len(prefixStates) - 2; i >= 0; i-- {
		if prefixStates[i].CurrentPlayer == c.Color {
			target = i
			break
		}
	}

	newGs := prefixStates[target]
	*gs = *newGs

	undoMsg := WSMessage{
		Type: "STATE",
		Data: mustMarshal(map[string]any{
			"board":         gs.Board,
			"currentPlayer": int(gs.CurrentPlayer),
			"history":       gs.History,
			"undone":        true,
		}),
	}
	hub.Broadcast(gs, c, undoMsg)

	// If undo leaves turn on AI, let AI move immediately.
	c.runAIMoveIfNeeded(hub)
}

func (c *Client) runAIMoveIfNeeded(hub *Hub) {
	gs := c.Session.State
	if c.Session.Mode != game.ModePVE || gs.GameOver {
		return
	}

	for c.Session.Mode == game.ModePVE && !gs.GameOver && gs.CurrentPlayer == c.Session.AI.Color {
		// Unlock briefly to avoid deadlock during AI computation.
		c.Session.Mutex.Unlock()
		time.Sleep(300 * time.Millisecond)
		c.Session.Mutex.Lock()

		if gs.GameOver || gs.CurrentPlayer != c.Session.AI.Color {
			return
		}

		bestPos := c.Session.AI.FindBestMove(gs)
		if bestPos == nil {
			// AI cannot move -> pass.
			gs.History = append(gs.History, game.Move{
				Player:   gs.CurrentPlayer,
				Position: nil,
			})
			gs.CurrentPlayer = gs.CurrentPlayer.Opponent()

			passMsg := WSMessage{
				Type: "STATE",
				Data: mustMarshal(map[string]any{
					"board":         gs.Board,
					"currentPlayer": int(gs.CurrentPlayer),
					"pass":          true,
					"history":       gs.History,
				}),
			}
			hub.Broadcast(gs, c, passMsg)
			if c.resolvePassAndGameOver(hub) {
				return
			}
			return
		}

		aiFlips, _ := gs.DoMove(bestPos.R, bestPos.C, gs.CurrentPlayer)
		aiMsg := WSMessage{
			Type: "AI_MOVE",
			Data: mustMarshal(map[string]any{
				"r":             bestPos.R,
				"c":             bestPos.C,
				"flipped":       aiFlips,
				"board":         gs.Board,
				"history":       gs.History,
				"currentPlayer": int(gs.CurrentPlayer),
			}),
		}
		hub.Broadcast(gs, c, aiMsg)
		if c.resolvePassAndGameOver(hub) {
			return
		}
	}
}

func (c *Client) resolvePassAndGameOver(hub *Hub) bool {
	gs := c.Session.State
	if gs.GameOver {
		c.handleGameOver(hub)
		return true
	}

	if len(gs.ValidMoves(gs.CurrentPlayer)) > 0 {
		return false
	}

	// Current player must pass.
	passColor := gs.CurrentPlayer
	gs.History = append(gs.History, game.Move{
		Player:   passColor,
		Position: nil,
	})
	gs.CurrentPlayer = passColor.Opponent()

	passMsg := WSMessage{
		Type: "STATE",
		Data: mustMarshal(map[string]any{
			"board":         gs.Board,
			"currentPlayer": int(gs.CurrentPlayer),
			"pass":          true,
			"history":       gs.History,
		}),
	}
	hub.Broadcast(gs, c, passMsg)

	if len(gs.ValidMoves(gs.CurrentPlayer)) == 0 {
		gs.GameOver = true
		c.handleGameOver(hub)
		return true
	}
	return false
}

func buildPrefixStates(size int, history []game.Move) []*game.GameState {
	states := make([]*game.GameState, 0, len(history)+1)
	gs := game.NewGameState(size)
	states = append(states, gs.Clone())

	for _, m := range history {
		if m.Position != nil {
			_, ok := gs.DoMove(m.Position.R, m.Position.C, m.Player)
			if !ok {
				break
			}
		} else {
			gs.History = append(gs.History, game.Move{
				Player:   m.Player,
				Position: nil,
			})
			gs.CurrentPlayer = m.Player.Opponent()
		}
		states = append(states, gs.Clone())
	}
	return states
}

func (c *Client) handleGameOver(hub *Hub) {
	gs := c.Session.State
	black, white := gs.Score()

	var winner string
	if black > white {
		winner = "BLACK"
	} else if white > black {
		winner = "WHITE"
	} else {
		winner = "DRAW"
	}

	overMsg := WSMessage{
		Type: "GAME_OVER",
		Data: mustMarshal(map[string]any{
			"winner":     winner,
			"blackScore": black,
			"whiteScore": white,
		}),
	}
	hub.Broadcast(gs, c, overMsg)
}

func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return data
}

func (c *Client) remoteAddr() string {
	if c == nil || c.Conn == nil || c.Conn.RemoteAddr() == nil {
		return "-"
	}
	return c.Conn.RemoteAddr().String()
}

func (c *Client) sessionID() string {
	if c == nil || c.Session == nil {
		return "-"
	}
	return c.Session.ID
}

func (c *Client) colorLabel() string {
	if c == nil {
		return "-"
	}
	switch c.Color {
	case game.BLACK:
		return "BLACK"
	case game.WHITE:
		return "WHITE"
	default:
		return "-"
	}
}

func main() {
	manager := game.NewManager()
	hub := NewHub(manager)
	go hub.Run()

	http.HandleFunc("/ws/game", func(w http.ResponseWriter, r *http.Request) {
		handleWS(hub, w, r)
	})

	log.Println("Othello backend starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
