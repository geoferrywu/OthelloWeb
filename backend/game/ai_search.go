package game

import "math"

// searchStrategy 聚合搜索所需的共享能力：评估、终局判断与搜索参数。
type searchStrategy struct {
	weights [][]int
	size    int
}

// depthByLevel 根据棋盘尺寸和等级返回搜索深度。
// easy 更快，hard 更强；用于增强博弈/PVS/混合博弈。
func depthByLevel(size int, level AILevel) int {
	switch level {
	case LevelEasy:
		if size <= 6 {
			return 3
		}
		if size <= 8 {
			return 2
		}
		return 2
	case LevelHard:
		if size <= 6 {
			return 7
		}
		if size <= 8 {
			return 5
		}
		return 4
	default:
		if size <= 6 {
			return 5
		}
		if size <= 8 {
			return 4
		}
		return 3
	}
}

// evaluate 评估当前局面对指定颜色的优劣。
// 当前实现 = 位置权重 + 行动力差（可继续扩展稳定子/前沿子等）。
func (s *searchStrategy) evaluate(board [][]Player, color Player) int {
	size := len(board)
	score := 0
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if board[r][c] == color {
				score += s.weights[r][c]
			} else if board[r][c] != EMPTY {
				score -= s.weights[r][c]
			}
		}
	}
	myMob := len(validMovesOnBoard(board, size, color))
	oppMob := len(validMovesOnBoard(board, size, color.Opponent()))
	score += (myMob - oppMob) * 5
	return score
}

// terminalScore 判断局面是否为终局并返回对应分数。
func (s *searchStrategy) terminalScore(board [][]Player, color Player) (int, bool) {
	size := len(board)
	if len(validMovesOnBoard(board, size, color)) > 0 || len(validMovesOnBoard(board, size, color.Opponent())) > 0 {
		return 0, false
	}
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
	if color == BLACK {
		return diff * 1000, true
	}
	return -diff * 1000, true
}

func abBestMove(gs *GameState, color Player, depth int, base *searchStrategy, usePVS bool) *Position {
	moves := gs.ValidMoves(color)
	if len(moves) == 0 {
		return nil
	}
	keys := make([]string, 0, len(moves))
	for k := range moves {
		keys = append(keys, k)
	}
	sortByFlips(keys, moves)
	bestScore := math.MinInt64
	bestKey := keys[0]
	for _, key := range keys {
		r, c := parseKey(key)
		nb := cloneBoard(gs)
		nb[r][c] = color
		for _, f := range moves[key] {
			nb[f.R][f.C] = color
		}
		var score int
		if usePVS {
			score = pvs(nb, depth-1, math.MinInt64, math.MaxInt64, false, color, base)
		} else {
			score = minimax(nb, depth-1, math.MinInt64, math.MaxInt64, false, color, base)
		}
		if score > bestScore {
			bestScore = score
			bestKey = key
		}
	}
	r, c := parseKey(bestKey)
	return &Position{R: r, C: c}
}

// minimax 为标准 Alpha-Beta 搜索主循环。
func minimax(board [][]Player, depth int, alpha, beta int, maximizing bool, color Player, base *searchStrategy) int {
	if term, ok := base.terminalScore(board, color); ok {
		return term
	}
	if depth == 0 {
		return base.evaluate(board, color)
	}

	current := color
	if !maximizing {
		current = color.Opponent()
	}
	moves := validMovesOnBoard(board, len(board), current)
	if len(moves) == 0 {
		return minimax(board, depth-1, alpha, beta, !maximizing, color, base)
	}
	keys := make([]string, 0, len(moves))
	for k := range moves {
		keys = append(keys, k)
	}
	sortByFlips(keys, moves)

	if maximizing {
		best := math.MinInt64
		for _, key := range keys {
			r, c := parseKey(key)
			nb := cloneBoardRaw(board, len(board))
			nb[r][c] = current
			for _, f := range moves[key] {
				nb[f.R][f.C] = current
			}
			best = maxInt(best, minimax(nb, depth-1, alpha, beta, false, color, base))
			alpha = maxInt(alpha, best)
			if alpha >= beta {
				break
			}
		}
		return best
	}

	best := math.MaxInt64
	for _, key := range keys {
		r, c := parseKey(key)
		nb := cloneBoardRaw(board, len(board))
		nb[r][c] = current
		for _, f := range moves[key] {
			nb[f.R][f.C] = current
		}
		best = minInt(best, minimax(nb, depth-1, alpha, beta, true, color, base))
		beta = minInt(beta, best)
		if alpha >= beta {
			break
		}
	}
	return best
}

// pvs 为 Principal Variation Search（主线剪枝）实现。
// 在走法排序较好时通常比标准 Alpha-Beta 更省节点。
func pvs(board [][]Player, depth int, alpha, beta int, maximizing bool, color Player, base *searchStrategy) int {
	if term, ok := base.terminalScore(board, color); ok {
		return term
	}
	if depth == 0 {
		return base.evaluate(board, color)
	}

	current := color
	if !maximizing {
		current = color.Opponent()
	}
	moves := validMovesOnBoard(board, len(board), current)
	if len(moves) == 0 {
		return pvs(board, depth-1, alpha, beta, !maximizing, color, base)
	}
	keys := make([]string, 0, len(moves))
	for k := range moves {
		keys = append(keys, k)
	}
	sortByFlips(keys, moves)

	first := true
	if maximizing {
		score := math.MinInt64
		for _, key := range keys {
			r, c := parseKey(key)
			nb := cloneBoardRaw(board, len(board))
			nb[r][c] = current
			for _, f := range moves[key] {
				nb[f.R][f.C] = current
			}
			var child int
			if first {
				child = pvs(nb, depth-1, alpha, beta, false, color, base)
				first = false
			} else {
				child = pvs(nb, depth-1, alpha, alpha+1, false, color, base)
				if child > alpha && child < beta {
					child = pvs(nb, depth-1, child, beta, false, color, base)
				}
			}
			score = maxInt(score, child)
			alpha = maxInt(alpha, score)
			if alpha >= beta {
				break
			}
		}
		return score
	}

	score := math.MaxInt64
	for _, key := range keys {
		r, c := parseKey(key)
		nb := cloneBoardRaw(board, len(board))
		nb[r][c] = current
		for _, f := range moves[key] {
			nb[f.R][f.C] = current
		}
		var child int
		if first {
			child = pvs(nb, depth-1, alpha, beta, true, color, base)
			first = false
		} else {
			child = pvs(nb, depth-1, beta-1, beta, true, color, base)
			if child > alpha && child < beta {
				child = pvs(nb, depth-1, alpha, child, true, color, base)
			}
		}
		score = minInt(score, child)
		beta = minInt(beta, score)
		if alpha >= beta {
			break
		}
	}
	return score
}
