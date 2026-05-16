package game

func weightMatrix(size int) [][]int {
	w := make([][]int, size)
	for i := range w {
		w[i] = make([]int, size)
	}

	edge := size - 1

	w[0][0] = 120
	w[0][edge] = 120
	w[edge][0] = 120
	w[edge][edge] = 120

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
