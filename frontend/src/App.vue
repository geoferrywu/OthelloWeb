<template>
  <div class="container">
    <h1>Othello</h1>

    <div v-if="wsStatus === 'connecting'" class="connecting">
      <p>正在连接...</p>
    </div>

    <StartScreen
      v-else-if="!gameStarted"
      :wsStatus="wsStatus"
      @start="handleStart"
    />

    <template v-else>
      <Scoreboard
        :board="currentBoard"
        :currentPlayer="wsCurrentPlayer"
        :isThinking="isThinking"
      />

      <div class="status">
        <span v-if="passShown">{{ passColorName }}跳过</span>
        <span v-else-if="isThinking && wsCurrentPlayer === aiColor">
          {{ currentPlayerName }}思考中...
        </span>
        <span v-else>{{ currentPlayerName }}落子</span>
      </div>

      <template v-if="boardReady">
        <GameBoard
          :board="currentBoard"
          :currentPlayer="wsCurrentPlayer"
          :playerColor="playerColor"
          :showHint="showHint"
          :isPlayerTurn="isPlayerTurn"
          :lastMove="lastMovePos"
          :flippedCells="flippedCells"
          @place="handlePlace"
        />

        <ControlPanel
          :canUndo="canUndo"
          :gameOver="wsGameOver"
          :showHistory="showHistory"
          :showHint="showHint"
          @undo="handleUndo"
          @toggleHistory="showHistory = !showHistory"
          @toggleHint="showHint = !showHint"
          @back="handleBack"
        />
      </template>
      <div v-else class="status">棋盘数据加载中...</div>

      <HistoryPanel
        v-if="showHistory"
        :history="moveLog"
        @close="showHistory = false"
      />

      <EndModal
        v-if="wsGameOver && overData"
        :overData="overData"
        @restart="handleRestart"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useWebSocket } from './composable/useWebSocket'
import type { GameMode, Color, Player, Position } from './types'
import StartScreen from './components/StartScreen.vue'
import Scoreboard from './components/Scoreboard.vue'
import GameBoard from './components/GameBoard.vue'
import ControlPanel from './components/ControlPanel.vue'
import HistoryPanel from './components/HistoryPanel.vue'
import EndModal from './components/EndModal.vue'

const {
  status: wsStatus,
  board: wsBoard,
  currentPlayer: wsCurrentPlayer,
  history: wsHistory,
  gameOver: wsGameOver,
  overData,
  passEvent,
  flippedCells: wsFlippedCells,
  connect,
  joinGame,
  sendMove,
  sendUndo,
} = useWebSocket()

const gameStarted = ref(false)
const gameMode = ref<GameMode>('PVE')
const playerColor = ref<Player>(1)
const aiColor = ref<Player>(2)
const boardSize = ref(8)
const showHistory = ref(false)
const showHint = ref(false)
const isThinking = ref(false)
const passShown = ref(false)
const lastMovePos = ref<Position | null>(null)
const flippedCells = ref<Position[]>([])

const BLACK: Player = 1
const WHITE: Player = 2

const currentBoard = computed(() => wsBoard.value || [])
const boardReady = computed(() => currentBoard.value.length > 0)

const moveLog = computed(() => {
  return (wsHistory.value || []).map((m) => {
    const side = m.player === BLACK ? '黑' : '白'
    const text = m.position ? `${side}:${coord(m.position.r, m.position.c)}` : `${side}跳过`
    return { color: m.player, text, pass: m.position === null }
  })
})

const isPlayerTurn = computed(() => {
  return wsCurrentPlayer.value === playerColor.value && !wsGameOver.value
})

const canUndo = computed(() => {
  if (gameMode.value === 'PVE') {
    return (wsHistory.value || []).length >= 1 && !wsGameOver.value
  }
  return (wsHistory.value || []).length >= 2 && !wsGameOver.value && isPlayerTurn.value
})

const currentPlayerName = computed(() => {
  return wsCurrentPlayer.value === BLACK ? '黑方' : '白方'
})

const passColorName = computed(() => {
  return wsCurrentPlayer.value === BLACK ? '白方' : '黑方'
})

function handleStart(mode: GameMode, color: Color, size: number) {
  gameMode.value = mode
  playerColor.value = color === 'BLACK' ? BLACK : WHITE
  aiColor.value = playerColor.value === BLACK ? WHITE : BLACK
  boardSize.value = size
  gameStarted.value = true
  joinGame(mode, color, size)
}

function handlePlace(r: number, c: number) {
  if (!isPlayerTurn.value) return
  sendMove(r, c)
  lastMovePos.value = { r, c }
}

function handleUndo() {
  sendUndo()
}

function handleBack() {
  gameStarted.value = false
  showHistory.value = false
  showHint.value = false
  lastMovePos.value = null
  flippedCells.value = []
}

function handleRestart() {
  joinGame(gameMode.value, playerColor.value === BLACK ? 'BLACK' : 'WHITE', boardSize.value)
}

function coord(r: number, c: number): string {
  return String.fromCharCode(65 + c) + (r + 1)
}

onMounted(() => {
  connect()
})

watch(passEvent, (val) => {
  if (val) {
    passShown.value = true
    setTimeout(() => {
      passShown.value = false
    }, 1500)
  }
})

watch(wsCurrentPlayer, (val) => {
  if (gameMode.value === 'PVE' && val === aiColor.value && !wsGameOver.value) {
    isThinking.value = true
  } else {
    isThinking.value = false
  }
})

watch(wsBoard, () => {
  if (isThinking.value) isThinking.value = false
})

watch(wsFlippedCells, (cells) => {
  flippedCells.value = cells || []
})
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: 'Segoe UI', sans-serif;
  background: #1a1a2e;
  color: #eee;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  -webkit-user-select: none;
  user-select: none;
}

.container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
}

h1 {
  font-size: 2rem;
  letter-spacing: 4px;
  color: #e0e0e0;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
}

.connecting {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px 48px;
  background: #2a2a4a;
  border-radius: 12px;
}

.connecting p {
  font-size: 1.1rem;
  color: #aaa;
}

.status {
  font-size: 1.1rem;
  min-height: 1.5em;
  text-align: center;
}
</style>
