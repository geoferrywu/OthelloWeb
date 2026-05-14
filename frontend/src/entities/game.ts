import type { Position, Player } from '../types'

const DIRS: [number, number][] = [
  [-1, -1], [-1, 0], [-1, 1],
  [0, -1],           [0, 1],
  [1, -1],  [1, 0],  [1, 1],
]

const EMPTY: Player = 0
const BLACK: Player = 1
const WHITE: Player = 2

function inBounds(r: number, c: number, size: number): boolean {
  return r >= 0 && r < size && c >= 0 && c < size
}

function getFlips(board: Player[][], r: number, c: number, player: Player, size: number): Position[] {
  if (board[r][c] !== EMPTY) return []
  const opp = player === BLACK ? WHITE : BLACK
  const all: Position[] = []
  for (const [dr, dc] of DIRS) {
    let nr = r + dr, nc = c + dc
    const flips: Position[] = []
    while (inBounds(nr, nc, size) && board[nr][nc] === opp) {
      flips.push({ r: nr, c: nc })
      nr += dr
      nc += dc
    }
    if (flips.length > 0 && inBounds(nr, nc, size) && board[nr][nc] === player) {
      all.push(...flips)
    }
  }
  return all
}

function validMoves(board: Player[][], player: Player, size: number): Map<string, Position[]> {
  const moves = new Map<string, Position[]>()
  for (let r = 0; r < size; r++) {
    for (let c = 0; c < size; c++) {
      const flips = getFlips(board, r, c, player, size)
      if (flips.length > 0) {
        moves.set(`${r},${c}`, flips)
      }
    }
  }
  return moves
}

function weightMatrix(size: number): number[][] {
  const w = Array.from({ length: size }, () => Array(size).fill(0))
  const edge = size - 1
  w[0][0] = 120; w[0][edge] = 120; w[edge][0] = 120; w[edge][edge] = 120
  const near = [
    [0,1],[1,0],[1,1],[0,edge-1],[1,edge-1],[1,edge],
    [edge-1,0],[edge,1],[edge-1,1],[edge-1,edge],[edge,edge-1],[edge-1,edge-1]
  ]
  for (const [r, c] of near) {
    if (r >= 0 && r < size && c >= 0 && c < size) w[r][c] = -40
  }
  for (let i = 1; i < size - 1; i++) {
    w[0][i] = 20; w[edge][i] = 20; w[i][0] = 20; w[i][edge] = 20
  }
  return w
}

function cloneBoard(board: Player[][], size: number): Player[][] {
  return board.map(row => [...row])
}

function evaluate(board: Player[][], weights: number[][], aiColor: Player, size: number): number {
  let score = 0
  for (let r = 0; r < size; r++) {
    for (let c = 0; c < size; c++) {
      if (board[r][c] === aiColor) score += weights[r][c]
      else if (board[r][c] !== EMPTY) score -= weights[r][c]
    }
  }
  return score
}

function minimax(
  board: Player[][],
  depth: number,
  alpha: number,
  beta: number,
  maximizing: boolean,
  aiColor: Player,
  size: number,
  weights: number[][]
): number {
  const player = maximizing ? aiColor : (aiColor === BLACK ? WHITE : BLACK)
  const moves = validMoves(board, player, size)
  const keys = Array.from(moves.keys())

  if (depth === 0 || keys.length === 0) {
    const next = player === BLACK ? WHITE : BLACK
    const nextMoves = validMoves(board, next, size)
    if (keys.length === 0 && nextMoves.size === 0) {
      let bl = 0, wh = 0
      for (let r = 0; r < size; r++)
        for (let c = 0; c < size; c++) {
          if (board[r][c] === BLACK) bl++
          else if (board[r][c] === WHITE) wh++
        }
      const diff = bl - wh
      return aiColor === BLACK ? diff * 100 : -diff * 100
    }
    return evaluate(board, weights, aiColor, size)
  }

  keys.sort((a, b) => moves.get(b)!.length - moves.get(a)!.length)

  if (maximizing) {
    let best = -Infinity
    for (const key of keys) {
      const [r, c] = key.split(',').map(Number)
      const nb = cloneBoard(board, size)
      nb[r][c] = aiColor
      for (const f of moves.get(key)!) nb[f.r][f.c] = aiColor
      best = Math.max(best, minimax(nb, depth - 1, alpha, beta, false, aiColor, size, weights))
      alpha = Math.max(alpha, best)
      if (alpha >= beta) break
    }
    return best
  } else {
    const opp = aiColor === BLACK ? WHITE : BLACK
    let best = Infinity
    for (const key of keys) {
      const [r, c] = key.split(',').map(Number)
      const nb = cloneBoard(board, size)
      nb[r][c] = opp
      for (const f of moves.get(key)!) nb[f.r][f.c] = opp
      best = Math.min(best, minimax(nb, depth - 1, alpha, beta, true, aiColor, size, weights))
      beta = Math.min(beta, best)
      if (alpha >= beta) break
    }
    return best
  }
}

export function getBestMove(
  board: Player[][],
  player: Player,
  size: number
): Position | null {
  const moves = validMoves(board, player, size)
  const keys = Array.from(moves.keys())
  if (keys.length === 0) return null

  const depth = size <= 6 ? 5 : size <= 8 ? 3 : 2
  const weights = weightMatrix(size)

  let bestScore = -Infinity
  let bestKey = keys[0]
  for (const key of keys) {
    const [r, c] = key.split(',').map(Number)
    const nb = cloneBoard(board, size)
    nb[r][c] = player
    for (const f of moves.get(key)!) nb[f.r][f.c] = player
    const sc = minimax(nb, depth - 1, -Infinity, Infinity, false, player, size, weights)
    if (sc > bestScore) {
      bestScore = sc
      bestKey = key
    }
  }

  const [r, c] = bestKey.split(',').map(Number)
  return { r, c }
}
