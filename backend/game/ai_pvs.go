package game

type pvsStrategy struct{ base *searchStrategy }

func (s *pvsStrategy) BestMove(gs *GameState, color Player, level AILevel) *Position {
	return abBestMove(gs, color, depthByLevel(gs.Size, level), s.base, true)
}
