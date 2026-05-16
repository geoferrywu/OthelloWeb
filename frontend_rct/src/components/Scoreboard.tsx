interface Props {
  board: number[][]
  currentPlayer: number
}

/**
 * 记分板：按当前棋盘直接统计黑白子数。
 */
export default function Scoreboard({ board, currentPlayer }: Props) {
  let blackScore = 0
  let whiteScore = 0
  for (const row of board) {
    for (const cell of row) {
      if (cell === 1) blackScore++
      else if (cell === 2) whiteScore++
    }
  }

  return (
    <div className="scoreboard">
      <div className={`player-score ${currentPlayer === 1 ? 'active' : ''}`}>
        <span className="disc-icon black" />
        <span>黑方 <strong>{blackScore}</strong></span>
      </div>
      <div className={`player-score ${currentPlayer === 2 ? 'active' : ''}`}>
        <span className="disc-icon white" />
        <span>白方 <strong>{whiteScore}</strong></span>
      </div>
    </div>
  )
}
