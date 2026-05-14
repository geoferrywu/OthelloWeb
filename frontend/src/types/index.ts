export type Player = 0 | 1 | 2 // 0=EMPTY, 1=BLACK, 2=WHITE
export type GameMode = 'PVE' | 'PVP'
export type Color = 'BLACK' | 'WHITE'

export interface Position {
  r: number
  c: number
}

export interface MoveRecord {
  player: Player
  position: Position | null
  flipped: Position[]
}

export interface GameInitData {
  gameId: string
  board: Player[][]
  currentPlayer: Player
  size: number
  history: MoveRecord[]
  players: { BLACK: string; WHITE: string }
}

export interface GameStateData {
  board: Player[][]
  currentPlayer: Player
  lastMove?: Position
  flipped?: Position[]
  history: MoveRecord[]
  pass?: boolean
  undone?: boolean
}

export interface AIMoveData {
  r: number
  c: number
  flipped: Position[]
  board: Player[][]
  history: MoveRecord[]
}

export interface GameOverData {
  winner: 'BLACK' | 'WHITE' | 'DRAW'
  blackScore: number
  whiteScore: number
}

export interface WSMessage {
  type: string
  data?: Record<string, unknown>
}

export interface UIState {
  showHistory: boolean
  showHint: boolean
  isThinking: boolean
  isConnecting: boolean
}
