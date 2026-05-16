package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	IsHost  bool
	Send    chan []byte
	mu      sync.Mutex
}

// Hub manages all connected clients.
type Hub struct {
	Clients    map[string]*Client // session ID -> client
	turnTimers map[string]*time.Timer
	warnTimers map[string]*time.Timer
	mu         sync.RWMutex
	Manager    *game.Manager
	Register   chan *Client
	Unregister chan *Client
}

func NewHub(m *game.Manager) *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		turnTimers: make(map[string]*time.Timer),
		warnTimers: make(map[string]*time.Timer),
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
			h.handleDisconnect(client)
		}
	}
}

func (h *Hub) sendInit(client *Client) {
	s := client.Session
	blackLabel := playerLabel(s, game.BLACK)
	whiteLabel := playerLabel(s, game.WHITE)
	if s.Mode == game.ModePVE {
		aiName := string(s.AISettings.Algorithm)
		aiLevel := string(s.AISettings.Level)
		if s.AI != nil && s.AI.Color == game.BLACK {
			blackLabel = "AI(" + aiName + ", " + aiLevel + ")"
		}
		if s.AI != nil && s.AI.Color == game.WHITE {
			whiteLabel = "AI(" + aiName + ", " + aiLevel + ")"
		}
	}
	data := map[string]any{
		"gameId":        s.ID,
		"board":         s.State.Board,
		"currentPlayer": int(s.State.CurrentPlayer),
		"selfColor":     int(client.Color),
		"size":          s.State.Size,
		"history":       s.State.History,
		"players": map[string]string{
			"BLACK": blackLabel,
			"WHITE": whiteLabel,
		},
		"aiSettings": map[string]string{
			"algorithm": string(s.AISettings.Algorithm),
			"level":     string(s.AISettings.Level),
		},
		"hintSettings": map[string]string{
			"algorithm": string(s.HintSettings[client.Color].Algorithm),
			"level":     string(s.HintSettings[client.Color].Level),
		},
	}
	if s.Mode == game.ModePVPOnline {
		data["online"] = map[string]any{
			"pairCode": s.PairCode,
			"isHost":   client.IsHost,
			"ready":    s.Ready,
		}
	}
	msg := WSMessage{Type: "INIT", Data: mustMarshal(data)}
	client.SendJSON(msg)
}

func (h *Hub) refreshOnlineInit(sessionID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.Clients {
		if c.Session != nil && c.Session.ID == sessionID {
			h.sendInit(c)
		}
	}
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

func (h *Hub) stopOnlineTurnTimers(sessionID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if t, ok := h.turnTimers[sessionID]; ok {
		t.Stop()
		delete(h.turnTimers, sessionID)
	}
	if t, ok := h.warnTimers[sessionID]; ok {
		t.Stop()
		delete(h.warnTimers, sessionID)
	}
}

func (h *Hub) startOnlineTurnTimers(sessionID string) {
	s, ok := h.Manager.GetSession(sessionID)
	if !ok || s.Mode != game.ModePVPOnline || !s.Ready || s.State.GameOver {
		return
	}
	h.stopOnlineTurnTimers(sessionID)

	warnTimer := time.AfterFunc(50*time.Second, func() {
		s2, ok := h.Manager.GetSession(sessionID)
		if !ok {
			return
		}
		s2.Mutex.Lock()
		defer s2.Mutex.Unlock()
		if s2.State.GameOver || !s2.Ready {
			return
		}
		h.Broadcast(s2.State, &Client{Session: s2}, WSMessage{
			Type: "COUNTDOWN",
			Data: mustMarshal(map[string]any{"seconds": 10}),
		})
	})

	turnTimer := time.AfterFunc(60*time.Second, func() {
		s2, ok := h.Manager.GetSession(sessionID)
		if !ok {
			return
		}
		s2.Mutex.Lock()
		defer s2.Mutex.Unlock()
		if s2.State.GameOver || !s2.Ready {
			return
		}
		s2.State.GameOver = true
		black, white := s2.State.Score()
		timeoutMsg := WSMessage{
			Type: "GAME_OVER",
			Data: mustMarshal(map[string]any{
				"winner":     "DRAW",
				"blackScore": black,
				"whiteScore": white,
				"reason":     "TIMEOUT",
				"message":    "超过60秒未落子，对局结束",
			}),
		}
		h.Broadcast(s2.State, &Client{Session: s2}, timeoutMsg)
		h.stopOnlineTurnTimers(sessionID)
		h.Manager.InvalidateOnlineCodeBySessionID(sessionID)
	})

	h.mu.Lock()
	h.warnTimers[sessionID] = warnTimer
	h.turnTimers[sessionID] = turnTimer
	h.mu.Unlock()
}

func (h *Hub) handleDisconnect(client *Client) {
	if client == nil || client.Session == nil {
		return
	}
	if client.Session.Mode != game.ModePVPOnline {
		return
	}
	client.Session.Mutex.Lock()
	defer client.Session.Mutex.Unlock()
	if client.Session.State.GameOver {
		return
	}
	client.Session.State.GameOver = true
	black, white := client.Session.State.Score()
	h.Broadcast(client.Session.State, &Client{Session: client.Session}, WSMessage{
		Type: "GAME_OVER",
		Data: mustMarshal(map[string]any{
			"winner":     "DRAW",
			"blackScore": black,
			"whiteScore": white,
			"reason":     "PLAYER_LEFT",
			"message":    "有玩家中途退出，对局结束",
		}),
	})
	h.stopOnlineTurnTimers(client.Session.ID)
	h.Manager.InvalidateOnlineCodeBySessionID(client.Session.ID)
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
		Mode        string `json:"mode"`
		Color       string `json:"color"`
		Size        int    `json:"size"`
		GameID      string `json:"gameId,omitempty"`
		PairCode    string `json:"pairCode,omitempty"`
		AIAlgorithm string `json:"aiAlgorithm,omitempty"`
		AILevel     string `json:"aiLevel,omitempty"`
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
	if mode != game.ModePVE && mode != game.ModePVP && mode != game.ModePVPOnline {
		mode = game.ModePVE
	}
	aiAlgorithm := game.ParseAlgorithmName(joinData.AIAlgorithm)
	aiLevel := game.ParseLevel(joinData.AILevel)

	var session *game.Session
	if mode == game.ModePVP {
		session = hub.Manager.JoinPvpSession(color, joinData.Size, joinData.GameID, aiAlgorithm, aiLevel)
	} else if mode == game.ModePVPOnline {
		if len(joinData.PairCode) != 4 {
			conn.WriteJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "配对码必须为4位数字"})})
			conn.Close()
			return
		}
		for _, ch := range joinData.PairCode {
			if ch < '0' || ch > '9' {
				conn.WriteJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "配对码必须为4位数字"})})
				conn.Close()
				return
			}
		}
		result := hub.Manager.JoinPvpOnlineSession(joinData.PairCode, color, joinData.Size, aiAlgorithm, aiLevel)
		if result.Reject != "" {
			conn.WriteJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": result.Reject})})
			conn.Close()
			return
		}
		session = result.Session
		client.IsHost = result.IsHost
		color = result.AssignedColor
	} else {
		session = hub.Manager.CreateSession(mode, color, joinData.Size, aiAlgorithm, aiLevel)
	}

	client.Session = session
	client.Color = color

	hub.Register <- client
	if mode == game.ModePVPOnline && session.Ready {
		hub.refreshOnlineInit(session.ID)
		hub.startOnlineTurnTimers(session.ID)
	}

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

	if c.Session.Mode == game.ModePVPOnline && !c.Session.Ready && msg.Type != "PING" {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "等待对手加入后开始游戏"})})
		return
	}

	if c.Session.State.GameOver {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Game is over"})})
		return
	}

	switch msg.Type {
	case "MOVE":
		c.handleMove(hub, msg.Data)
	case "HINT":
		c.handleHint(msg.Data)
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
	movePlayer := c.Color
	if c.Session.Mode == game.ModePVP {
		// Local PvP allows one client to alternately play both sides.
		movePlayer = gs.CurrentPlayer
	}
	if gs.CurrentPlayer != movePlayer {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Not your turn"})})
		return
	}

	flips, ok := gs.DoMove(moveData.R, moveData.C, movePlayer)
	if !ok {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Invalid move"})})
		return
	}
	if hintPos := c.Session.LastHint[movePlayer]; hintPos != nil && hintPos.R == moveData.R && hintPos.C == moveData.C {
		last := len(gs.History) - 1
		if last >= 0 {
			gs.History[last].HintTag = string(c.Session.HintSettings[movePlayer].Algorithm) + "_" + string(c.Session.HintSettings[movePlayer].Level)
		}
	}
	c.Session.LastHint[movePlayer] = nil

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
	if c.Session.Mode == game.ModePVPOnline {
		hub.startOnlineTurnTimers(c.Session.ID)
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
	c.Session.LastHint[c.Color] = nil

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

func (c *Client) handleHint(data json.RawMessage) {
	var hintData struct {
		Algorithm string `json:"algorithm"`
		Level     string `json:"level"`
	}
	_ = json.Unmarshal(data, &hintData)
	if c.Session.State.GameOver {
		c.SendJSON(WSMessage{Type: "ERROR", Data: mustMarshal(map[string]string{"message": "Game is over"})})
		return
	}
	hintPlayer := c.Color
	if c.Session.Mode == game.ModePVP {
		hintPlayer = c.Session.State.CurrentPlayer
	}
	alg := c.Session.HintSettings[hintPlayer].Algorithm
	lv := c.Session.HintSettings[hintPlayer].Level
	if hintData.Algorithm != "" {
		alg = game.ParseAlgorithmName(hintData.Algorithm)
	}
	if hintData.Level != "" {
		lv = game.ParseLevel(hintData.Level)
	}
	c.Session.HintSettings[hintPlayer] = game.HintSettings{Algorithm: alg, Level: lv}
	engine := game.NewHintEngine(c.Session.State.Size, alg)
	best := engine.BestMove(c.Session.State, hintPlayer, lv)
	c.Session.LastHint[hintPlayer] = best
	c.SendJSON(WSMessage{
		Type: "HINT_RESULT",
		Data: mustMarshal(map[string]any{
			"position": best,
			"algorithm": map[string]string{
				"name": string(alg),
				"code": game.AlgorithmProfile(alg).Code,
			},
			"level": string(lv),
		}),
	})
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
		c.Session.LastHint[c.Color] = nil
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
			before := len(gs.History)
			_, ok := gs.DoMove(m.Position.R, m.Position.C, m.Player)
			if !ok {
				break
			}
			if len(gs.History) > before {
				gs.History[len(gs.History)-1].HintTag = m.HintTag
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
			"reason":     "NORMAL",
		}),
	}
	hub.Broadcast(gs, c, overMsg)
	if c.Session.Mode == game.ModePVPOnline {
		hub.stopOnlineTurnTimers(c.Session.ID)
		hub.Manager.InvalidateOnlineCodeBySessionID(c.Session.ID)
	}
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

	port := os.Getenv("OTHELLO_BACKEND_PORT")
	if port == "" {
		port = "8088"
	}

	addr := ":" + port
	log.Println("Othello backend starting on " + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
