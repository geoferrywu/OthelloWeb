package game

import "math"

// AI holds the AI evaluation context.
type AI struct {
	weights [][]int
	depth   int
	Color   Player
}

// NewAI creates an AI instance for a given board size and player color.
func NewAI(size int, color Player, depth int) *AI {
	return &AI{
		weights: weightMatrix(size),
		depth:   depth,
		Color:   color,
	}
}

// AIDepth returns the recommended search depth for a board size.
func AIDepth(size int) int {
	switch {
	case size <= 6:
		return 6
	case size <= 8:
		return 4
	default:
		return 3
	}
}

// FindBestMove returns the best move position for the AI's color.
func (ai *AI) FindBestMove(gs *GameState) *Position {
	moves := gs.ValidMoves(ai.Color)
	if len(moves) == 0 {
		return nil
	}

	bestScore := math.MinInt64
	var bestPos *Position

	for key, flips := range moves {
		r, c := parseKey(key)
		nb := cloneBoard(gs)
		nb[r][c] = ai.Color
		for _, f := range flips {
			nb[f.R][f.C] = ai.Color
		}

		score := ai.minimax(nb, ai.depth-1, math.MinInt64, math.MaxInt64, false)
		if score > bestScore {
			bestScore = score
			bestPos = &Position{R: r, C: c}
		}
	}

	return bestPos
}

// evaluate scores a board from the AI's perspective.
func (ai *AI) evaluate(board [][]Player, size int) int {
	score := 0
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if board[r][c] == ai.Color {
				score += ai.weights[r][c]
			} else if board[r][c] != EMPTY {
				score -= ai.weights[r][c]
			}
		}
	}
	return score
}

func (ai *AI) minimax(board [][]Player, depth int, alpha, beta int, maximizing bool) int {
	size := len(board)

	// Determine current player
	var currentPlayer Player
	if maximizing {
		currentPlayer = ai.Color
	} else {
		currentPlayer = ai.Color.Opponent()
	}

	moves := validMovesOnBoard(board, size, currentPlayer)
	keys := make([]string, 0, len(moves))
	for k := range moves {
		keys = append(keys, k)
	}

	if depth == 0 || len(keys) == 0 {
		// Check if game is truly over (neither player can move)
		nextPlayer := currentPlayer.Opponent()
		if len(keys) == 0 && len(validMovesOnBoard(board, size, nextPlayer)) == 0 {
			black, white := 0, 0
			for r := 0; r < size; r++ {
				for c := 0; c < size; c++ {
					if board[r][c] == BLACK {
						black++
					} else if board[r][c] == WHITE {
						white++
					}
				}
			}
			diff := black - white
			if ai.Color == BLACK {
				return diff * 100
			}
			return -diff * 100
		}
		return ai.evaluate(board, size)
	}

	// Sort by flip count for better pruning
	sortByFlips(keys, moves)

	if maximizing {
		best := math.MinInt64
		for _, key := range keys {
			r, c := parseKey(key)
			nb := cloneBoardRaw(board, size)
			for _, f := range moves[key] {
				nb[f.R][f.C] = ai.Color
			}
			nb[r][c] = ai.Color
			best = maxInt(best, ai.minimax(nb, depth-1, alpha, beta, false))
			alpha = maxInt(alpha, best)
			if alpha >= beta {
				break
			}
		}
		return best
	} else {
		best := math.MaxInt64
		opp := ai.Color.Opponent()
		for _, key := range keys {
			r, c := parseKey(key)
			nb := cloneBoardRaw(board, size)
			for _, f := range moves[key] {
				nb[f.R][f.C] = opp
			}
			nb[r][c] = opp
			best = minInt(best, ai.minimax(nb, depth-1, alpha, beta, true))
			beta = minInt(beta, best)
			if alpha >= beta {
				break
			}
		}
		return best
	}
}

// weightMatrix creates a positional weight matrix.
func weightMatrix(size int) [][]int {
	w := make([][]int, size)
	for i := range w {
		w[i] = make([]int, size)
	}

	edge := size - 1

	// Corner bonus
	w[0][0] = 120
	w[0][edge] = 120
	w[edge][0] = 120
	w[edge][edge] = 120

	// Near-corner penalty
	nearCorners := [][2]int{
		{0, 1}, {1, 0}, {1, 1},
		{0, edge - 1}, {1, edge - 1}, {1, edge},
		{edge - 1, 0}, {edge, 1}, {edge - 1, 1},
		{edge - 1, edge}, {edge, edge - 1}, {edge - 1, edge - 1},
	}
	for _, nc := range nearCorners {
		if nc[0] >= 0 && nc[0] < size && nc[1] >= 0 && nc[1] < size {
			w[nc[0]][nc[1]] = -40
		}
	}

	// Edge bonus
	for i := 1; i < size-1; i++ {
		w[0][i] = 20
		w[edge][i] = 20
		w[i][0] = 20
		w[i][edge] = 20
	}

	return w
}

func validMovesOnBoard(board [][]Player, size int, player Player) map[string][]Position {
	dirs := [][2]int{
		{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1},
	}
	inBounds := func(r, c int) bool { return r >= 0 && r < size && c >= 0 && c < size }

	moves := make(map[string][]Position)
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if board[r][c] != EMPTY {
				continue
			}
			var allFlips []Position
			for _, d := range dirs {
				nr, nc := r+d[0], c+d[1]
				var flips []Position
				opp := player.Opponent()
				for inBounds(nr, nc) && board[nr][nc] == opp {
					flips = append(flips, Position{R: nr, C: nc})
					nr += d[0]
					nc += d[1]
				}
				if len(flips) > 0 && inBounds(nr, nc) && board[nr][nc] == player {
					allFlips = append(allFlips, flips...)
				}
			}
			if len(allFlips) > 0 {
				moves[posKey(r, c)] = allFlips
			}
		}
	}
	return moves
}

func cloneBoard(gs *GameState) [][]Player {
	nb := make([][]Player, gs.Size)
	for r := 0; r < gs.Size; r++ {
		nb[r] = make([]Player, gs.Size)
		copy(nb[r], gs.Board[r])
	}
	return nb
}

func cloneBoardRaw(board [][]Player, size int) [][]Player {
	nb := make([][]Player, size)
	for r := 0; r < size; r++ {
		nb[r] = make([]Player, size)
		copy(nb[r], board[r])
	}
	return nb
}

func parseKey(key string) (int, int) {
	var r, c int
	comma := -1
	for i := 0; i < len(key); i++ {
		if key[i] == ',' {
			comma = i
			break
		}
		r = r*10 + int(key[i]-'0')
	}
	for i := comma + 1; i < len(key); i++ {
		c = c*10 + int(key[i]-'0')
	}
	return r, c
}

func sortByFlips(keys []string, moves map[string][]Position) {
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if len(moves[keys[j]]) > len(moves[keys[i]]) {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
