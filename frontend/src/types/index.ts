export type Player = 0 | 1 | 2
export type GameMode = 'PVE' | 'PVP' | 'PVP_ONLINE'
export type Color = 'BLACK' | 'WHITE'
export type AILevel = 'easy' | 'normal' | 'hard'

export interface Position {
  r: number
  c: number
}

export interface MoveRecord {
  player: Player
  position: Position | null
  flipped: Position[]
  hintTag?: string
}

export interface GameInitData {
  gameId: string
  board: Player[][]
  currentPlayer: Player
  size: number
  history: MoveRecord[]
  players: { BLACK: string; WHITE: string }
  aiSettings?: { algorithm: string; level: AILevel }
  hintSettings?: { algorithm: string; level: AILevel }
  online?: {
    pairCode: string
    isHost: boolean
    ready: boolean
  }
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
  currentPlayer?: Player
}

export interface HintResultData {
  position: Position | null
  algorithm: { name: string; code: string }
  level: AILevel
}

export interface GameOverData {
  winner: 'BLACK' | 'WHITE' | 'DRAW'
  blackScore: number
  whiteScore: number
  reason?: 'NORMAL' | 'PLAYER_LEFT' | 'TIMEOUT'
  message?: string
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

