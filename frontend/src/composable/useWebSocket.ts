import { ref, type Ref } from 'vue'
import type {
  WSMessage,
  GameMode,
  Color,
  GameInitData,
  GameStateData,
  AIMoveData,
  GameOverData,
  MoveRecord,
  Position,
  AILevel,
  HintResultData,
} from '../types'

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected'

interface UseWebSocketReturn {
  status: Ref<ConnectionStatus>
  init: Ref<GameInitData | null>
  board: Ref<number[][] | null>
  currentPlayer: Ref<number>
  history: Ref<MoveRecord[]>
  gameOver: Ref<boolean>
  overData: Ref<GameOverData | null>
  passEvent: Ref<boolean>
  flippedCells: Ref<Position[]>
  hintMove: Ref<Position | null>
  errorMessage: Ref<string>
  countdown: Ref<number>
  connect: () => void
  joinGame: (mode: GameMode, color: Color, size: number, aiAlgorithm: string, aiLevel: AILevel, pairCode?: string) => void
  sendMove: (r: number, c: number) => void
  sendUndo: () => void
  requestHint: (algorithm: string, level: AILevel) => void
  reconnect: () => void
  leaveGame: () => void
}

let ws: WebSocket | null = null
let countdownTimer: number | null = null

export function useWebSocket(): UseWebSocketReturn {
  const status = ref<ConnectionStatus>('disconnected') as Ref<ConnectionStatus>
  const init = ref<GameInitData | null>(null) as Ref<GameInitData | null>
  const board = ref<number[][] | null>(null)
  const currentPlayer = ref(1)
  const history = ref<MoveRecord[]>([])
  const gameOver = ref(false)
  const overData = ref<GameOverData | null>(null)
  const passEvent = ref(false)
  const flippedCells = ref<Position[]>([])
  const hintMove = ref<Position | null>(null)
  const errorMessage = ref('')
  const countdown = ref(0)

  function stopCountdown() {
    if (countdownTimer !== null) {
      window.clearInterval(countdownTimer)
      countdownTimer = null
    }
  }

  function startCountdown(from: number) {
    stopCountdown()
    countdown.value = Math.max(0, from)
    if (countdown.value <= 0) return
    countdownTimer = window.setInterval(() => {
      if (countdown.value <= 0) {
        stopCountdown()
        return
      }
      countdown.value -= 1
      if (countdown.value <= 0) stopCountdown()
    }, 1000)
  }

  function resolveWsUrl(): string {
    const envUrl = import.meta.env.VITE_WS_URL as string | undefined
    if (envUrl && envUrl.trim().length > 0) return envUrl
    const protocol = location.protocol === 'https:' ? 'wss' : 'ws'
    const host = location.hostname
    const frontendPort = import.meta.env.VITE_FRONTEND_PORT
    const backendPort = import.meta.env.VITE_BACKEND_PORT
    if (String(location.port) === String(frontendPort)) {
      return `${protocol}://${location.host}/ws/game`
    }
    return `${protocol}://${host}:${backendPort}/ws/game`
  }

  function connect(onMessage?: (msg: WSMessage) => void) {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return
    status.value = 'connecting'
    ws = new WebSocket(resolveWsUrl())
    ws.onopen = () => { status.value = 'connected' }
    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        handleServerMessage(msg)
        onMessage?.(msg)
      } catch {
      }
    }
    ws.onclose = () => { status.value = 'disconnected' }
    ws.onerror = () => { status.value = 'disconnected' }
  }

  function handleServerMessage(msg: WSMessage) {
    switch (msg.type) {
      case 'INIT': {
        const data = msg.data as unknown as GameInitData
        init.value = data
        board.value = data.board as unknown as number[][]
        currentPlayer.value = data.currentPlayer
        history.value = data.history || []
        gameOver.value = false
        overData.value = null
        passEvent.value = false
        flippedCells.value = []
        hintMove.value = data.online?.activeHint || null
        stopCountdown()
        countdown.value = 0
        break
      }
      case 'STATE': {
        const data = msg.data as unknown as GameStateData
        board.value = data.board as unknown as number[][]
        currentPlayer.value = data.currentPlayer
        if (data.history) history.value = data.history
        passEvent.value = !!data.pass
        flippedCells.value = data.flipped || []
        hintMove.value = null
        stopCountdown()
        countdown.value = 0
        break
      }
      case 'COUNTDOWN': {
        const data = msg.data as unknown as { seconds?: number }
        startCountdown(Math.max(0, data.seconds || 0))
        break
      }
      case 'AI_MOVE': {
        const data = msg.data as unknown as AIMoveData
        board.value = data.board as unknown as number[][]
        if (typeof data.currentPlayer === 'number') currentPlayer.value = data.currentPlayer
        else currentPlayer.value = currentPlayer.value === 1 ? 2 : 1
        if (data.history) history.value = data.history
        passEvent.value = false
        flippedCells.value = data.flipped || []
        hintMove.value = null
        stopCountdown()
        countdown.value = 0
        break
      }
      case 'HINT_RESULT': {
        const data = msg.data as unknown as HintResultData
        hintMove.value = data.position || null
        break
      }
      case 'GAME_OVER': {
        overData.value = msg.data as unknown as GameOverData
        gameOver.value = true
        stopCountdown()
        countdown.value = 0
        break
      }
      case 'ERROR': {
        const data = msg.data as unknown as { message?: string }
        errorMessage.value = data?.message || '请求失败'
        break
      }
    }
  }

  function send(type: string, data?: Record<string, unknown>) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type, data }))
  }

  function joinGame(mode: GameMode, color: Color, size: number, aiAlgorithm: string, aiLevel: AILevel, pairCode?: string) {
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.close()
      ws = null
    }
    status.value = 'disconnected'

    errorMessage.value = ''
    stopCountdown()
    countdown.value = 0
    connect()
    const checkOpen = () => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        send('JOIN', { mode, color, size, aiAlgorithm, aiLevel, pairCode })
        return
      }
      if (ws && ws.readyState === WebSocket.CONNECTING) {
        setTimeout(checkOpen, 50)
        return
      }
      connect()
      setTimeout(checkOpen, 100)
    }
    checkOpen()
  }

  function sendMove(r: number, c: number) { send('MOVE', { r, c }) }
  function sendUndo() { send('UNDO') }
  function requestHint(algorithm: string, level: AILevel) { send('HINT', { algorithm, level }) }

  function reconnect() {
    if (ws) ws.close()
    ws = null
    connect()
  }

  function leaveGame() {
    if (ws) {
      ws.onclose = null
      ws.onerror = null
      ws.close()
      ws = null
    }
    status.value = 'disconnected'
    stopCountdown()
    countdown.value = 0
  }

  return {
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
    connect: () => connect(),
    joinGame,
    sendMove,
    sendUndo,
    requestHint,
    reconnect,
    leaveGame,
  }
}

