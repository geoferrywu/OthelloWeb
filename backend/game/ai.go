package game

// AIAlgorithmName 使用中文短名作为算法标识，前后端统一用该值切换算法。
type AIAlgorithmName string

const (
	AlgorithmEnhanced AIAlgorithmName = "增强博弈"
	AlgorithmPVS      AIAlgorithmName = "主线剪枝"
	AlgorithmMCTS     AIAlgorithmName = "蒙特树搜"
	AlgorithmHybrid   AIAlgorithmName = "混合博弈"
)

// AILevel 统一三档难度。
type AILevel string

const (
	LevelEasy   AILevel = "easy"
	LevelNormal AILevel = "normal"
	LevelHard   AILevel = "hard"
)

// AIProfile 提供算法代码（用于历史记录后缀和面板显示）。
type AIProfile struct {
	Name AIAlgorithmName `json:"name"`
	Code string          `json:"code"`
}

var algorithmProfiles = map[AIAlgorithmName]AIProfile{
	AlgorithmEnhanced: {Name: AlgorithmEnhanced, Code: "abx"},
	AlgorithmPVS:      {Name: AlgorithmPVS, Code: "pvs"},
	AlgorithmMCTS:     {Name: AlgorithmMCTS, Code: "mcts"},
	AlgorithmHybrid:   {Name: AlgorithmHybrid, Code: "mix"},
}

func ParseAlgorithmName(name string) AIAlgorithmName {
	a := AIAlgorithmName(name)
	if _, ok := algorithmProfiles[a]; ok {
		return a
	}
	return AlgorithmEnhanced
}

func ParseLevel(level string) AILevel {
	l := AILevel(level)
	switch l {
	case LevelEasy, LevelNormal, LevelHard:
		return l
	default:
		return LevelNormal
	}
}

func AlgorithmProfile(name AIAlgorithmName) AIProfile {
	if p, ok := algorithmProfiles[name]; ok {
		return p
	}
	return algorithmProfiles[AlgorithmEnhanced]
}

// AIStrategy 定义统一算法接口，所有AI策略都通过该接口提供最佳落子。
type AIStrategy interface {
	BestMove(gs *GameState, color Player, level AILevel) *Position
}

// AI 保存对局内锁定的算法与等级配置（开局后不再改变）。
type AI struct {
	Color     Player
	Algorithm AIAlgorithmName
	Level     AILevel
	strategy  AIStrategy
}

func NewAI(size int, color Player, algorithm AIAlgorithmName, level AILevel) *AI {
	return &AI{
		Color:     color,
		Algorithm: algorithm,
		Level:     level,
		strategy:  NewAIStrategyFactory(size).Create(algorithm),
	}
}

func (ai *AI) FindBestMove(gs *GameState) *Position {
	if ai == nil || ai.strategy == nil {
		return nil
	}
	return ai.strategy.BestMove(gs, ai.Color, ai.Level)
}

// NewHintEngine 为提示功能创建独立策略实例。HINT 走这里，不影响对局AI配置。
func NewHintEngine(size int, algorithm AIAlgorithmName) AIStrategy {
	return NewAIStrategyFactory(size).Create(algorithm)
}

type AIStrategyFactory struct {
	size int
}

func NewAIStrategyFactory(size int) *AIStrategyFactory {
	return &AIStrategyFactory{size: size}
}

func (f *AIStrategyFactory) Create(name AIAlgorithmName) AIStrategy {
	base := &searchStrategy{weights: weightMatrix(f.size), size: f.size}
	switch name {
	case AlgorithmPVS:
		return &pvsStrategy{base: base}
	case AlgorithmMCTS:
		return newMCTSStrategy(base)
	case AlgorithmHybrid:
		return &hybridStrategy{base: base}
	case AlgorithmEnhanced:
		fallthrough
	default:
		return &enhancedABStrategy{base: base}
	}
}
