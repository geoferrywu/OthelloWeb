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
  connect: () => void
  joinGame: (mode: GameMode, color: Color, size: number) => void
  sendMove: (r: number, c: number) => void
  sendUndo: () => void
  reconnect: () => void
}

let ws: WebSocket | null = null

export function useWebSocket(): UseWebSocketReturn {
  const status = ref<ConnectionStatus>('disconnected') as Ref<ConnectionStatus>
  const init = ref<GameInitData | null>(null) as Ref<GameInitData | null>
  const board = ref<number[][] | null>(null)
  const currentPlayer = ref(1)
  const history = ref<MoveRecord[]>([])
  const gameOver = ref(false)
  const overData = ref<GameOverData | null>(null)
  const passEvent = ref(false)

  function connect(onMessage?: (msg: WSMessage) => void) {
    if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    status.value = 'connecting'
    const wsUrl = `ws://${location.host}/ws/game`
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      status.value = 'connected'
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        handleServerMessage(msg)
        onMessage?.(msg)
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      status.value = 'disconnected'
    }

    ws.onerror = () => {
      status.value = 'disconnected'
    }
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
        break
      }
      case 'STATE': {
        const data = msg.data as unknown as GameStateData
        board.value = data.board as unknown as number[][]
        currentPlayer.value = data.currentPlayer
        if (data.history) history.value = data.history
        passEvent.value = !!data.pass
        break
      }
      case 'AI_MOVE': {
        const data = msg.data as unknown as AIMoveData
        board.value = data.board as unknown as number[][]
        if (data.history) history.value = data.history
        passEvent.value = false
        break
      }
      case 'GAME_OVER': {
        overData.value = msg.data as unknown as GameOverData
        gameOver.value = true
        break
      }
    }
  }

  function send(type: string, data?: Record<string, unknown>) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return
    ws.send(JSON.stringify({ type, data }))
  }

  function joinGame(mode: GameMode, color: Color, size: number) {
    connect()
    const checkOpen = () => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        send('JOIN', { mode, color, size })
      } else if (ws && ws.readyState === WebSocket.CONNECTING) {
        setTimeout(checkOpen, 50)
      } else {
        connect()
        setTimeout(checkOpen, 100)
      }
    }
    checkOpen()
  }

  function sendMove(r: number, c: number) {
    send('MOVE', { r, c })
  }

  function sendUndo() {
    send('UNDO')
  }

  function reconnect() {
    if (ws) {
      ws.close()
    }
    ws = null
    connect()
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
    connect: () => connect(),
    joinGame,
    sendMove,
    sendUndo,
    reconnect,
  }
}
