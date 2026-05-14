package game

var dirs = [][2]int{
	{-1, -1}, {-1, 0}, {-1, 1},
	{0, -1}, {0, 1},
	{1, -1}, {1, 0}, {1, 1},
}

func inBounds(r, c, size int) bool {
	return r >= 0 && r < size && c >= 0 && c < size
}

// GetFlips returns all cells that would be flipped by placing a disc at (r, c).
func (gs *GameState) GetFlips(r, c int, player Player) []Position {
	if gs.Board[r][c] != EMPTY {
		return nil
	}
	opp := player.Opponent()
	var all []Position
	for _, d := range dirs {
		nr, nc := r+d[0], c+d[1]
		var flips []Position
		for inBounds(nr, nc, gs.Size) && gs.Board[nr][nc] == opp {
			flips = append(flips, Position{R: nr, C: nc})
			nr += d[0]
			nc += d[1]
		}
		if len(flips) > 0 && inBounds(nr, nc, gs.Size) && gs.Board[nr][nc] == player {
			all = append(all, flips...)
		}
	}
	return all
}

// ValidMoves returns a map of valid move keys ("r,c") to the cells that would be flipped.
func (gs *GameState) ValidMoves(player Player) map[string][]Position {
	moves := make(map[string][]Position)
	for r := 0; r < gs.Size; r++ {
		for c := 0; c < gs.Size; c++ {
			if flips := gs.GetFlips(r, c, player); len(flips) > 0 {
				key := posKey(r, c)
				moves[key] = flips
			}
		}
	}
	return moves
}

// DoMove executes a move on the board. Returns false if the move is invalid.
func (gs *GameState) DoMove(r, c int, player Player) ([]Position, bool) {
	moves := gs.ValidMoves(player)
	key := posKey(r, c)
	flips, ok := moves[key]
	if !ok {
		return nil, false
	}

	gs.Board[r][c] = player
	for _, f := range flips {
		gs.Board[f.R][f.C] = player
	}

	gs.History = append(gs.History, Move{
		Player:   player,
		Position: &Position{R: r, C: c},
		Flipped:  flips,
	})

	gs.CurrentPlayer = player.Opponent()
	return flips, true
}

// TryMove attempts a move and handles pass logic.
// Returns true if the game should continue, false if game over.
func (gs *GameState) TryMove(r, c int, player Player) (moved bool, gameOver bool) {
	moves := gs.ValidMoves(player)
	key := posKey(r, c)
	flips, ok := moves[key]
	if !ok {
		// Check if current player has any moves
		if len(moves) > 0 {
			return false, false
		}
		// Current player must pass
		gs.History = append(gs.History, Move{
			Player:   player,
			Position: nil,
		})
		gs.CurrentPlayer = player.Opponent()
		nextMoves := gs.ValidMoves(gs.CurrentPlayer)
		if len(nextMoves) == 0 {
			gs.GameOver = true
			return false, true
		}
		return false, false
	}

	gs.Board[r][c] = player
	for _, f := range flips {
		gs.Board[f.R][f.C] = player
	}
	gs.History = append(gs.History, Move{
		Player:   player,
		Position: &Position{R: r, C: c},
		Flipped:  flips,
	})
	gs.CurrentPlayer = player.Opponent()
	return true, false
}

// PosKey creates a unique key for a position.
func posKey(r, c int) string {
	return itoa(r) + "," + itoa(c)
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	if n < 100 {
		return string(rune('0'+n/10)) + string(rune('0'+n%10))
	}
	return string(rune('0'+n/100)) + string(rune('0'+(n/10)%10)) + string(rune('0'+n%10))
}
