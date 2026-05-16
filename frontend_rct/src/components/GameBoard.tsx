import { useEffect, useMemo, useRef, useState } from 'react'
import type { Player, Position } from '../types'

interface Props {
  board: number[][]
  currentPlayer: number
  showHint: boolean
  showHistory: boolean
  isPlayerTurn: boolean
  allowPreviewMoves?: boolean
  flippedCells: Position[]
  hintMove: Position | null
  historyEntries: Array<{ color: number; text: string; pass: boolean }>
  onPlace: (r: number, c: number) => void
}

/**
 * 棋盘组件：
 * 1) 渲染棋盘/棋子/坐标
 * 2) 计算合法落点
 * 3) 渲染提示红点与翻子动画
 * 4) 右侧对局记录滚动
 */
export default function GameBoard(props: Props) {
  const size = props.board.length || 8
  const boardReady = props.board.length > 0 && props.board.every((row) => row.length === props.board.length)

  const px = size <= 6 ? 64 : size === 8 ? 52 : 42
  const historyWidth = size <= 6 ? 198 : size === 8 ? 186 : 170
  const historyGap = size <= 6 ? 8 : size === 8 ? 6 : 5

  const [animatingFlips, setAnimatingFlips] = useState<Set<string>>(new Set())
  const historyListRef = useRef<HTMLDivElement | null>(null)

  useEffect(() => {
    if (!props.flippedCells || props.flippedCells.length === 0) return
    const next = new Set<string>()
    for (const f of props.flippedCells) next.add(`${f.r},${f.c}`)
    setAnimatingFlips(next)
    const timer = window.setTimeout(() => setAnimatingFlips(new Set()), 420)
    return () => window.clearTimeout(timer)
  }, [props.flippedCells])

  useEffect(() => {
    if (!historyListRef.current) return
    historyListRef.current.scrollTop = historyListRef.current.scrollHeight
  }, [props.historyEntries.length])

  const validMovesSet = useMemo(() => {
    const set = new Set<string>()
    if (!boardReady) return set
    if (!props.isPlayerTurn && !props.allowPreviewMoves) return set

    const player = props.currentPlayer as Player
    for (let r = 0; r < size; r++) {
      for (let c = 0; c < size; c++) {
        if (props.board[r][c] !== 0) continue
        if (getFlipsForHint(props.board, r, c, player, size).length > 0) set.add(`${r},${c}`)
      }
    }
    return set
  }, [boardReady, props.allowPreviewMoves, props.board, props.currentPlayer, props.isPlayerTurn, size])

  const cellAt = (r: number, c: number): number => {
    const row = props.board[r]
    if (!row || c < 0 || c >= row.length) return 0
    return row[c] ?? 0
  }

  const isFlipped = (r: number, c: number): boolean => animatingFlips.has(`${r},${c}`)
  const isValidMove = (r: number, c: number): boolean => validMovesSet.has(`${r},${c}`)
  const isBestMove = (r: number, c: number): boolean => !!props.hintMove && props.hintMove.r === r && props.hintMove.c === c

  const handleClick = (r: number, c: number) => {
    if (!props.isPlayerTurn) return
    if (!isValidMove(r, c)) return
    props.onPlace(r, c)
  }

  const colLabel = (c: number) => String.fromCharCode(65 + c)

  if (!boardReady) {
    return <div className="board-placeholder">Loading board...</div>
  }

  return (
    <div className="board-shell">
      <div
        className={`board ${props.showHistory ? 'history-open' : ''}`}
        style={{
          gridTemplateColumns: `24px repeat(${size}, var(--cell-size))`,
          ['--cell-size' as any]: `${px}px`,
          ['--history-width' as any]: `${historyWidth}px`,
          ['--history-gap' as any]: `${historyGap}px`,
        }}
      >
        <div className="coord-label" />
        {Array.from({ length: size }).map((_, c) => <div key={`col-${c}`} className="coord-label">{colLabel(c)}</div>)}

        {Array.from({ length: size }).map((_, r) => (
          <>
            <div key={`row-label-${r}`} className="coord-label">{r + 1}</div>
            {Array.from({ length: size }).map((__, c) => {
              const cell = cellAt(r, c)
              return (
                <div key={`${r}-${c}`} className={`cell ${!props.isPlayerTurn ? 'disabled' : ''}`} onClick={() => handleClick(r, c)}>
                  {cell !== 0 ? (
                    <div className={`disc ${cell === 1 ? 'black' : 'white'} ${isFlipped(r, c) ? 'flipping' : ''}`} />
                  ) : isValidMove(r, c) ? (
                    <div className={`hint ${props.showHint && isBestMove(r, c) ? 'on' : ''}`} />
                  ) : null}
                </div>
              )
            })}
          </>
        ))}

        {props.showHistory && (
          <aside className="board-history">
            <h4>对局记录</h4>
            <div ref={historyListRef} className="board-history-list">
              {props.historyEntries.map((m, i) => (
                <div key={i} className={`board-history-item ${m.pass ? 'pass' : (m.color === 1 ? 'black-move' : 'white-move')}`}>{m.text}</div>
              ))}
            </div>
          </aside>
        )}
      </div>
    </div>
  )
}

function getFlipsForHint(board: number[][], r: number, c: number, player: Player, size: number): Position[] {
  if (board[r][c] !== 0) return []
  const opp = player === 1 ? 2 : 1
  const dirs: [number, number][] = [[-1, -1], [-1, 0], [-1, 1], [0, -1], [0, 1], [1, -1], [1, 0], [1, 1]]
  const all: Position[] = []
  const inB = (rr: number, cc: number) => rr >= 0 && rr < size && cc >= 0 && cc < size

  for (const [dr, dc] of dirs) {
    let nr = r + dr
    let nc = c + dc
    const flips: Position[] = []
    while (inB(nr, nc) && board[nr][nc] === opp) {
      flips.push({ r: nr, c: nc })
      nr += dr
      nc += dc
    }
    if (flips.length > 0 && inB(nr, nc) && board[nr][nc] === player) all.push(...flips)
  }

  return all
}
