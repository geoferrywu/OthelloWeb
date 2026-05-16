import { useCallback, useMemo, useRef, useState } from 'react'
import type {
  AIMoveData,
  AILevel,
  Color,
  ConnectionStatus,
  GameInitData,
  GameMode,
  GameOverData,
  GameStateData,
  HintResultData,
  MoveRecord,
  Position,
  WSMessage,
} from '../types'

/**
 * 将 WebSocket 通讯与状态管理封装在 Hook 中，
 * 让页面组件只关注“展示与交互”，不直接处理底层网络细节。
 */
export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const countdownTimerRef = useRef<number | null>(null)

  const [status, setStatus] = useState<ConnectionStatus>('disconnected')
  const [init, setInit] = useState<GameInitData | null>(null)
  const [board, setBoard] = useState<number[][] | null>(null)
  const [currentPlayer, setCurrentPlayer] = useState<number>(1)
  const [history, setHistory] = useState<MoveRecord[]>([])
  const [gameOver, setGameOver] = useState(false)
  const [overData, setOverData] = useState<GameOverData | null>(null)
  const [passEvent, setPassEvent] = useState(false)
  const [flippedCells, setFlippedCells] = useState<Position[]>([])
  const [hintMove, setHintMove] = useState<Position | null>(null)
  const [errorMessage, setErrorMessage] = useState('')
  const [countdown, setCountdown] = useState(0)

  const stopCountdown = useCallback(() => {
    if (countdownTimerRef.current !== null) {
      window.clearInterval(countdownTimerRef.current)
      countdownTimerRef.current = null
    }
  }, [])

  const startCountdown = useCallback((from: number) => {
    stopCountdown()
    setCountdown(Math.max(0, from))
    if (from <= 0) return

    countdownTimerRef.current = window.setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          stopCountdown()
          return 0
        }
        return prev - 1
      })
    }, 1000)
  }, [stopCountdown])

  const resolveWsUrl = useCallback(() => {
    const envUrl = (import.meta as any).env?.VITE_WS_URL as string | undefined
    if (envUrl && envUrl.trim()) return envUrl

    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const host = location.hostname
    const frontendPort = (import.meta as any).env?.VITE_FRONTEND_PORT
    const backendPort = (import.meta as any).env?.VITE_BACKEND_PORT

    if (String(location.port) === String(frontendPort)) {
      return `${protocol}://${location.host}/ws/game`
    }
    return `${protocol}://${host}:${backendPort}/ws/game`
  }, [])

  const handleServerMessage = useCallback((msg: WSMessage) => {
    switch (msg.type) {
      case 'INIT': {
        const data = msg.data as unknown as GameInitData
        setInit(data)
        setBoard((data.board || []) as number[][])
        setCurrentPlayer(data.currentPlayer)
        setHistory(data.history || [])
        setGameOver(false)
        setOverData(null)
        setPassEvent(false)
        setFlippedCells([])
        setHintMove(data.online?.activeHint || null)
        stopCountdown()
        setCountdown(0)
        break
      }
      case 'STATE': {
        const data = msg.data as unknown as GameStateData
        setBoard((data.board || []) as number[][])
        setCurrentPlayer(data.currentPlayer)
        if (data.history) setHistory(data.history)
        setPassEvent(!!data.pass)
        setFlippedCells(data.flipped || [])
        setHintMove(null)
        stopCountdown()
        setCountdown(0)
        break
      }
      case 'COUNTDOWN': {
        const data = msg.data as unknown as { seconds?: number }
        startCountdown(Math.max(0, data.seconds || 0))
        break
      }
      case 'AI_MOVE': {
        const data = msg.data as unknown as AIMoveData
        setBoard((data.board || []) as number[][])
        if (typeof data.currentPlayer === 'number') setCurrentPlayer(data.currentPlayer)
        else setCurrentPlayer((prev) => (prev === 1 ? 2 : 1))
        if (data.history) setHistory(data.history)
        setPassEvent(false)
        setFlippedCells(data.flipped || [])
        setHintMove(null)
        stopCountdown()
        setCountdown(0)
        break
      }
      case 'HINT_RESULT': {
        const data = msg.data as unknown as HintResultData
        setHintMove(data.position || null)
        break
      }
      case 'GAME_OVER': {
        setOverData(msg.data as unknown as GameOverData)
        setGameOver(true)
        stopCountdown()
        setCountdown(0)
        break
      }
      case 'ERROR': {
        const data = msg.data as unknown as { message?: string }
        setErrorMessage(data?.message || '请求失败')
        break
      }
      default:
        break
    }
  }, [startCountdown, stopCountdown])

  const connect = useCallback(() => {
    const ws = wsRef.current
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    setStatus('connecting')
    const next = new WebSocket(resolveWsUrl())
    wsRef.current = next

    next.onopen = () => setStatus('connected')
    next.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        handleServerMessage(msg)
      } catch {
        // 忽略非法报文，避免影响正常渲染流程。
      }
    }
    next.onclose = () => setStatus('disconnected')
    next.onerror = () => setStatus('disconnected')
  }, [handleServerMessage, resolveWsUrl])

  const send = useCallback((type: string, data?: Record<string, unknown>) => {
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type, data }))
  }, [])

  const joinGame = useCallback((mode: GameMode, color: Color, size: number, aiAlgorithm: string, aiLevel: AILevel, pairCode?: string) => {
    const ws = wsRef.current
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.close()
      wsRef.current = null
    }

    setStatus('disconnected')
    setErrorMessage('')
    stopCountdown()
    setCountdown(0)

    connect()

    const checkOpen = () => {
      const cur = wsRef.current
      if (cur && cur.readyState === WebSocket.OPEN) {
        send('JOIN', { mode, color, size, aiAlgorithm, aiLevel, pairCode })
        return
      }
      if (cur && cur.readyState === WebSocket.CONNECTING) {
        setTimeout(checkOpen, 50)
        return
      }
      connect()
      setTimeout(checkOpen, 100)
    }

    checkOpen()
  }, [connect, send, stopCountdown])

  const sendMove = useCallback((r: number, c: number) => send('MOVE', { r, c }), [send])
  const sendUndo = useCallback(() => send('UNDO'), [send])
  const requestHint = useCallback((algorithm: string, level: AILevel) => send('HINT', { algorithm, level }), [send])

  const leaveGame = useCallback(() => {
    const ws = wsRef.current
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.close()
      wsRef.current = null
    }
    setStatus('disconnected')
    stopCountdown()
    setCountdown(0)
  }, [stopCountdown])

  const reconnect = useCallback(() => {
    const ws = wsRef.current
    if (ws) ws.close()
    wsRef.current = null
    connect()
  }, [connect])

  return useMemo(() => ({
    status,
    init,
    board,
    currentPlayer,
    history,
    gameOver,
    overData,
    passEvent,
    flippedCells,
    hintMove,
    errorMessage,
    countdown,
    connect,
    joinGame,
    sendMove,
    sendUndo,
    requestHint,
    reconnect,
    leaveGame,
    setErrorMessage,
  }), [
    status,
    init,
    board,
    currentPlayer,
    history,
    gameOver,
    overData,
    passEvent,
    flippedCells,
    hintMove,
    errorMessage,
    countdown,
    connect,
    joinGame,
    sendMove,
    sendUndo,
    requestHint,
    reconnect,
    leaveGame,
  ])
}
