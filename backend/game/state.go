package game

// Player represents a disc color on the board.
type Player int

const (
	EMPTY Player = iota
	BLACK
	WHITE
)

func (p Player) Opponent() Player {
	if p == BLACK {
		return WHITE
	}
	return BLACK
}

func (p Player) String() string {
	switch p {
	case BLACK:
		return "BLACK"
	case WHITE:
		return "WHITE"
	default:
		return "-"
	}
}

// Position is a board cell coordinate.
type Position struct {
	R int `json:"r"`
	C int `json:"c"`
}

// Move records a single move in the game.
type Move struct {
	Player   Player     `json:"player"`
	Position *Position  `json:"position"` // nil means pass
	Flipped  []Position `json:"flipped"`
	HintTag  string     `json:"hintTag,omitempty"`
}

// GameState holds the complete state of an Othello game.
type GameState struct {
	Board         [][]Player `json:"board"`
	CurrentPlayer Player     `json:"currentPlayer"`
	Size          int        `json:"size"`
	History       []Move     `json:"history"`
	GameOver      bool       `json:"gameOver"`
}

// NewGameState creates an initialized board.
func NewGameState(size int) *GameState {
	board := make([][]Player, size)
	for i := range board {
		board[i] = make([]Player, size)
	}

	mid := size / 2
	half := (size+1)/2 - 1
	c1, c2 := half, mid

	board[c1][c1] = WHITE
	board[c1][c2] = BLACK
	board[c2][c1] = BLACK
	board[c2][c2] = WHITE

	return &GameState{
		Board:         board,
		CurrentPlayer: BLACK,
		Size:          size,
		History:       make([]Move, 0),
		GameOver:      false,
	}
}

// Clone returns a deep copy of the game state.
func (gs *GameState) Clone() *GameState {
	clone := &GameState{
		Board:         make([][]Player, gs.Size),
		CurrentPlayer: gs.CurrentPlayer,
		Size:          gs.Size,
		History:       make([]Move, len(gs.History)),
		GameOver:      gs.GameOver,
	}
	for r := 0; r < gs.Size; r++ {
		clone.Board[r] = make([]Player, gs.Size)
		copy(clone.Board[r], gs.Board[r])
	}
	copy(clone.History, gs.History)
	return clone
}

// Score returns the disc count for each player.
func (gs *GameState) Score() (int, int) {
	var black, white int
	for r := 0; r < gs.Size; r++ {
		for c := 0; c < gs.Size; c++ {
			if gs.Board[r][c] == BLACK {
				black++
			} else if gs.Board[r][c] == WHITE {
				white++
			}
		}
	}
	return black, white
}
