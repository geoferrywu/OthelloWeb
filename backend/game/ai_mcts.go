package game

import (
	"math/rand"
	"time"
)

// mctsStrategy 通过随机模拟评估候选走法胜率。
// 当前版本为轻量实现，便于在 1 秒预算内稳定返回结果。
type mctsStrategy struct {
	base *searchStrategy
	rng  *rand.Rand
}

func newMCTSStrategy(base *searchStrategy) *mctsStrategy {
	return &mctsStrategy{base: base, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// BestMove 对每个候选走法执行多次随机对局，选择胜局数最高者。
func (s *mctsStrategy) BestMove(gs *GameState, color Player, level AILevel) *Position {
	moves := gs.ValidMoves(color)
	if len(moves) == 0 {
		return nil
	}
	iter := 200
	switch level {
	case LevelEasy:
		iter = 80
	case LevelHard:
		iter = 400
	}
	keys := make([]string, 0, len(moves))
	for k := range moves {
		keys = append(keys, k)
	}
	bestKey := keys[0]
	bestWins := -1
	for _, key := range keys {
		wins := 0
		for i := 0; i < iter; i++ {
			r, c := parseKey(key)
			nb := cloneBoard(gs)
			nb[r][c] = color
			for _, f := range moves[key] {
				nb[f.R][f.C] = color
			}
			if s.rollout(nb, color.Opponent(), color) > 0 {
				wins++
			}
		}
		if wins > bestWins {
			bestWins = wins
			bestKey = key
		}
	}
	r, c := parseKey(bestKey)
	return &Position{R: r, C: c}
}

// rollout 从给定局面随机模拟到终局，返回相对 perspective 的胜负结果。
func (s *mctsStrategy) rollout(board [][]Player, current Player, perspective Player) int {
	size := len(board)
	pass := 0
	for pass < 2 {
		moves := validMovesOnBoard(board, size, current)
		if len(moves) == 0 {
			pass++
			current = current.Opponent()
			continue
		}
		pass = 0
		keys := make([]string, 0, len(moves))
		for k := range moves {
			keys = append(keys, k)
		}
		pick := keys[s.rng.Intn(len(keys))]
		r, c := parseKey(pick)
		board[r][c] = current
		for _, f := range moves[pick] {
			board[f.R][f.C] = current
		}
		current = current.Opponent()
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
	if perspective == WHITE {
		diff = -diff
	}
	if diff > 0 {
		return 1
	}
	if diff == 0 {
		return 0
	}
	return -1
}
