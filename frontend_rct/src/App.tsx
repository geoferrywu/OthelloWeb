import { useEffect, useMemo, useRef, useState } from 'react'
import StartScreen from './components/StartScreen'
import Scoreboard from './components/Scoreboard'
import ControlPanel from './components/ControlPanel'
import GameBoard from './components/GameBoard'
import EndModal from './components/EndModal'
import { useWebSocket } from './hooks/useWebSocket'
import type { AILevel, Color, GameMode, Player, Position } from './types'

const BLACK: Player = 1
const WHITE: Player = 2

/**
 * React 版本主页面：
 * 与当前 Vue 版本功能保持一致，包含 PVE / PVP / 在线 PVP + 观战。
 */
export default function App() {
  const ws = useWebSocket()
  const lastAutoHintKeyRef = useRef('')

  const [gameStarted, setGameStarted] = useState(false)
  const [gameMode, setGameMode] = useState<GameMode>('PVE')
  const [playerColor, setPlayerColor] = useState<Player>(BLACK)
  const [aiColor, setAiColor] = useState<Player>(WHITE)
  const [boardSize, setBoardSize] = useState(8)

  const [showHistory, setShowHistory] = useState(false)
  const [showHint, setShowHint] = useState(false)
  const [isThinking, setIsThinking] = useState(false)
  const [passShown, setPassShown] = useState(false)
  const [lastMovePos, setLastMovePos] = useState<Position | null>(null)
  const [flippedCells, setFlippedCells] = useState<Position[]>([])

  const [aiAlgorithm, setAiAlgorithm] = useState('增强博弈')
  const [aiLevel, setAiLevel] = useState<AILevel>('normal')
  const [hintAlgorithm, setHintAlgorithm] = useState('增强博弈')
  const [hintLevel, setHintLevel] = useState<AILevel>('normal')
  const [pairCode, setPairCode] = useState('')

  const currentBoard = useMemo(() => ws.board || [], [ws.board])
  const boardReady = currentBoard.length > 0
  const isOnlineMode = gameMode === 'PVP_ONLINE'
  const onlineReady = !!ws.init?.online?.ready
  const isOnlineSpectator = isOnlineMode && !!ws.init?.online?.isSpectator

  const moveLog = useMemo(() => {
    return (ws.history || []).map((m) => {
      const side = m.player === BLACK ? '黑' : '白'
      const base = m.position ? `${side}:${coord(m.position.r, m.position.c)}` : `${side}跳过`
      const text = m.position && m.hintTag ? `${base}_${m.hintTag}` : base
      return { color: m.player, text, pass: m.position === null }
    })
  }, [ws.history])

  const isPlayerTurn = useMemo(() => {
    if (ws.gameOver) return false
    if (isOnlineSpectator) return false
    if (gameMode === 'PVP') return true
    return ws.currentPlayer === playerColor
  }, [ws.currentPlayer, ws.gameOver, gameMode, isOnlineSpectator, playerColor])

  const canUndo = useMemo(() => {
    if (gameMode === 'PVE') return (ws.history || []).length >= 1 && !ws.gameOver
    if (gameMode === 'PVP_ONLINE') return false
    return (ws.history || []).length >= 2 && !ws.gameOver && isPlayerTurn
  }, [gameMode, isPlayerTurn, ws.gameOver, ws.history])

  const currentPlayerName = ws.currentPlayer === BLACK ? '黑方' : '白方'
  const passColorName = ws.currentPlayer === BLACK ? '白方' : '黑方'

  const onlineWaitingText = useMemo(() => {
    if (!isOnlineMode) return ''
    if (isOnlineSpectator) return '观战模式'
    const role = ws.init?.online?.isHost ? '主玩家' : '客玩家'
    const code = ws.init?.online?.pairCode || pairCode
    return `${role}（配对码 ${code}），等待对手加入...`
  }, [isOnlineMode, isOnlineSpectator, pairCode, ws.init?.online?.isHost, ws.init?.online?.pairCode])

  const onlineCountdownText = useMemo(() => {
    if (!isOnlineMode || ws.countdown <= 0) return ''
    const myTurn = ws.currentPlayer === playerColor
    return myTurn ? `请在 ${ws.countdown} 秒内落子` : `对手剩余 ${ws.countdown} 秒`
  }, [isOnlineMode, playerColor, ws.countdown, ws.currentPlayer])

  const blackRole = useMemo(() => {
    if (gameMode === 'PVP_ONLINE') {
      if (isOnlineSpectator) return ''
      return playerColor === BLACK ? '你' : '对手'
    }
    return ws.init?.players?.BLACK || (gameMode === 'PVP' ? '玩家' : (playerColor === BLACK ? '玩家' : 'AI'))
  }, [gameMode, isOnlineSpectator, playerColor, ws.init?.players?.BLACK])

  const whiteRole = useMemo(() => {
    if (gameMode === 'PVP_ONLINE') {
      if (isOnlineSpectator) return ''
      return playerColor === WHITE ? '你' : '对手'
    }
    return ws.init?.players?.WHITE || (gameMode === 'PVP' ? '玩家' : (playerColor === WHITE ? '玩家' : 'AI'))
  }, [gameMode, isOnlineSpectator, playerColor, ws.init?.players?.WHITE])

  const handleStart = (mode: GameMode, color: Color, size: number, selectedAlgorithm: string, selectedLevel: AILevel, selectedPairCode?: string) => {
    setGameMode(mode)
    const nextPlayer = color === 'BLACK' ? BLACK : WHITE
    setPlayerColor(nextPlayer)
    setAiColor(nextPlayer === BLACK ? WHITE : BLACK)
    setBoardSize(size)
    setPairCode(selectedPairCode || '')
    setAiAlgorithm(selectedAlgorithm)
    setAiLevel(selectedLevel)
    setHintAlgorithm('增强博弈')
    setHintLevel('normal')
    setGameStarted(true)

    ws.joinGame(mode, color, size, selectedAlgorithm, selectedLevel, selectedPairCode)
  }

  const handlePlace = (r: number, c: number) => {
    if (!isPlayerTurn) return
    ws.sendMove(r, c)
    setLastMovePos({ r, c })
  }

  const handleUndo = () => {
    if (isOnlineMode) return
    ws.sendUndo()
    if (showHint) ws.requestHint(hintAlgorithm, hintLevel)
  }

  const toggleHint = () => {
    setShowHint((prev) => {
      const next = !prev
      if (next) ws.requestHint(hintAlgorithm, hintLevel)
      return next
    })
  }

  const handleBack = () => {
    if (gameStarted) ws.leaveGame()
    setGameStarted(false)
    setShowHistory(false)
    setShowHint(false)
    setLastMovePos(null)
    setFlippedCells([])
    setPairCode('')
    ws.connect()
  }

  const handleRestart = () => {
    setHintAlgorithm('增强博弈')
    setHintLevel('normal')
    ws.joinGame(gameMode, playerColor === BLACK ? 'BLACK' : 'WHITE', boardSize, aiAlgorithm, aiLevel, pairCode || undefined)
  }

  useEffect(() => {
    ws.connect()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (ws.passEvent) {
      setPassShown(true)
      const timer = window.setTimeout(() => setPassShown(false), 1500)
      return () => window.clearTimeout(timer)
    }
  }, [ws.passEvent])

  useEffect(() => {
    if (gameMode === 'PVE' && ws.currentPlayer === aiColor && !ws.gameOver) setIsThinking(true)
    else setIsThinking(false)

    if (!showHint || !isPlayerTurn || ws.gameOver) {
      lastAutoHintKeyRef.current = ''
      return
    }

    // 仅在“回合/算法/难度”上下文发生变化时发一次提示请求，避免渲染驱动的重复请求。
    const autoHintKey = `${ws.currentPlayer}|${hintAlgorithm}|${hintLevel}`
    if (lastAutoHintKeyRef.current === autoHintKey) return
    lastAutoHintKeyRef.current = autoHintKey
    ws.requestHint(hintAlgorithm, hintLevel)
  }, [
    aiColor,
    gameMode,
    hintAlgorithm,
    hintLevel,
    isPlayerTurn,
    showHint,
    ws.currentPlayer,
    ws.gameOver,
    ws.requestHint,
  ])

  useEffect(() => {
    if (!ws.errorMessage) return
    // 对局中不弹阻塞式 alert，避免阻断渲染。
    if (gameStarted) return
    window.alert(ws.errorMessage)
    ws.setErrorMessage('')
  }, [gameStarted, ws])

  useEffect(() => {
    if (!ws.gameOver || !isOnlineMode) return
    const reason = ws.overData?.reason
    if (reason === 'NORMAL') return
    const message = ws.overData?.message
    if (message) window.alert(message)
    handleBack()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOnlineMode, ws.gameOver])

  useEffect(() => {
    if (!ws.init || !isOnlineMode) return
    if (ws.init.selfColor === BLACK || ws.init.selfColor === WHITE) {
      setPlayerColor(ws.init.selfColor)
      setAiColor(ws.init.selfColor === BLACK ? WHITE : BLACK)
    }
  }, [isOnlineMode, ws.init])

  useEffect(() => {
    if (isThinking) setIsThinking(false)
  }, [ws.board])

  useEffect(() => {
    setFlippedCells(ws.flippedCells || [])
  }, [ws.flippedCells])

  return (
    <div className="container">
      <h1>Othello</h1>

      {ws.status === 'connecting' ? (
        <div className="connecting"><p>正在连接...</p></div>
      ) : !gameStarted ? (
        <StartScreen wsStatus={ws.status} onStart={handleStart} />
      ) : (
        <>
          <section className="game-shell">
            <Scoreboard board={currentBoard} currentPlayer={ws.currentPlayer} />

            <div className="status">
              {isOnlineMode && !onlineReady ? (
                <span>{onlineWaitingText}</span>
              ) : isOnlineMode && ws.countdown > 0 ? (
                <span>{onlineCountdownText}</span>
              ) : passShown ? (
                <span>{passColorName}跳过</span>
              ) : isThinking && ws.currentPlayer === aiColor ? (
                <span>{currentPlayerName}思考中...</span>
              ) : (
                <span>{currentPlayerName}落子</span>
              )}
            </div>

            {boardReady ? (
              <div className="board-layout">
                <div className="left-panel">
                  <aside className="side-info">
                    <h3>对弈双方</h3>
                    <p><strong>黑：</strong>{blackRole}</p>
                    <p><strong>白：</strong>{whiteRole}</p>
                  </aside>

                  <ControlPanel
                    canUndo={canUndo}
                    gameOver={ws.gameOver}
                    showHistory={showHistory}
                    showHint={showHint}
                    hintAlgorithm={hintAlgorithm}
                    hintLevel={hintLevel}
                    disableHint={isOnlineSpectator}
                    onUndo={handleUndo}
                    onToggleHistory={() => setShowHistory((p) => !p)}
                    onToggleHint={toggleHint}
                    onHintAlgorithmChange={(value) => {
                      setHintAlgorithm(value)
                      if (showHint) ws.requestHint(value, hintLevel)
                    }}
                    onHintLevelChange={(value) => {
                      const next = value as AILevel
                      setHintLevel(next)
                      if (showHint) ws.requestHint(hintAlgorithm, next)
                    }}
                    onBack={handleBack}
                  />
                </div>

                <GameBoard
                  board={currentBoard}
                  currentPlayer={ws.currentPlayer}
                  showHint={showHint || isOnlineSpectator}
                  showHistory={showHistory}
                  isPlayerTurn={isPlayerTurn}
                  allowPreviewMoves={isOnlineSpectator}
                  flippedCells={flippedCells}
                  hintMove={ws.hintMove}
                  historyEntries={moveLog}
                  onPlace={handlePlace}
                />
              </div>
            ) : (
              <div className="status">棋盘数据加载中...</div>
            )}
          </section>

          {ws.gameOver && ws.overData && !isOnlineMode && (
            <EndModal overData={ws.overData} actionText="再来一局" onRestart={handleRestart} />
          )}
          {ws.gameOver && ws.overData && isOnlineMode && ws.overData.reason === 'NORMAL' && (
            <EndModal overData={ws.overData} actionText={isOnlineSpectator ? '返回' : '再来一局'} onRestart={handleBack} />
          )}
        </>
      )}
    </div>
  )
}

function coord(r: number, c: number): string {
  return String.fromCharCode(65 + c) + (r + 1)
}

