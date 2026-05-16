package game

type hybridStrategy struct{ base *searchStrategy }

func (s *hybridStrategy) BestMove(gs *GameState, color Player, level AILevel) *Position {
	depth := depthByLevel(gs.Size, level)
	if level == LevelHard {
		depth++
	}
	return abBestMove(gs, color, depth, s.base, false)
}
