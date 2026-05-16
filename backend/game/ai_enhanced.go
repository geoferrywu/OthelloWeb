package game

type enhancedABStrategy struct{ base *searchStrategy }

func (s *enhancedABStrategy) BestMove(gs *GameState, color Player, level AILevel) *Position {
	return abBestMove(gs, color, depthByLevel(gs.Size, level), s.base, false)
}
