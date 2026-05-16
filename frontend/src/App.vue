<template>
  <div class="container">
    <h1>Othello</h1>

    <div v-if="wsStatus === 'connecting'" class="connecting"><p>正在连接...</p></div>

    <StartScreen v-else-if="!gameStarted" :wsStatus="wsStatus" @start="handleStart" />

    <template v-else>
      <Scoreboard :board="currentBoard" :currentPlayer="wsCurrentPlayer" :isThinking="isThinking" />

      <div class="status">
        <span v-if="passShown">{{ passColorName }}跳过</span>
        <span v-else-if="isThinking && wsCurrentPlayer === aiColor">{{ currentPlayerName }}思考中...</span>
        <span v-else>{{ currentPlayerName }}落子</span>
      </div>

      <template v-if="boardReady">
        <div class="board-layout">
          <aside class="side-info">
            <h3>对弈双方</h3>
            <p><strong>黑：</strong>{{ blackRole }}</p>
            <p><strong>白：</strong>{{ whiteRole }}</p>
          </aside>

          <GameBoard
            :board="currentBoard"
            :currentPlayer="wsCurrentPlayer"
            :playerColor="playerColor"
            :showHint="showHint"
            :isPlayerTurn="isPlayerTurn"
            :lastMove="lastMovePos"
            :flippedCells="flippedCells"
            :hintMove="hintMove"
            @place="handlePlace"
          />
        </div>

        <ControlPanel
          :canUndo="canUndo"
          :gameOver="wsGameOver"
          :showHistory="showHistory"
          :showHint="showHint"
          :hintAlgorithm="hintAlgorithm"
          :hintLevel="hintLevel"
          @undo="handleUndo"
          @toggleHistory="showHistory = !showHistory"
          @toggleHint="toggleHint"
          @hintAlgorithmChange="changeHintAlgorithm"
          @hintLevelChange="changeHintLevel"
          @back="handleBack"
        />
      </template>
      <div v-else class="status">棋盘数据加载中...</div>

      <HistoryPanel v-if="showHistory" :history="moveLog" @close="showHistory = false" />
      <EndModal v-if="wsGameOver && overData" :overData="overData" @restart="handleRestart" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useWebSocket } from './composable/useWebSocket'
import type { GameMode, Color, Player, Position, AILevel } from './types'
import StartScreen from './components/StartScreen.vue'
import Scoreboard from './components/Scoreboard.vue'
import GameBoard from './components/GameBoard.vue'
import ControlPanel from './components/ControlPanel.vue'
import HistoryPanel from './components/HistoryPanel.vue'
import EndModal from './components/EndModal.vue'

// WebSocket 状态与接口
const {
  status: wsStatus,
  init,
  board: wsBoard,
  currentPlayer: wsCurrentPlayer,
  history: wsHistory,
  gameOver: wsGameOver,
  overData,
  passEvent,
  flippedCells: wsFlippedCells,
  hintMove,
  connect,
  joinGame,
  sendMove,
  sendUndo,
  requestHint,
} = useWebSocket()

// 对局与 UI 状态
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

// AI 设置在开局时锁定；提示设置可随时切换
const aiAlgorithm = ref('增强博弈')
const aiLevel = ref<AILevel>('normal')
const hintAlgorithm = ref('增强博弈')
const hintLevel = ref<AILevel>('normal')

const BLACK: Player = 1
const WHITE: Player = 2

const currentBoard = computed(() => wsBoard.value || [])
const boardReady = computed(() => currentBoard.value.length > 0)

// 历史记录：若命中提示点，追加 "_算法中文名_等级中文名" 后缀
const moveLog = computed(() => {
  return (wsHistory.value || []).map((m) => {
    const side = m.player === BLACK ? '黑' : '白'
    const base = m.position ? `${side}:${coord(m.position.r, m.position.c)}` : `${side}跳过`
    const text = m.position && m.hintTag ? `${base}_${m.hintTag}` : base
    return { color: m.player, text, pass: m.position === null }
  })
})

const isPlayerTurn = computed(() => wsCurrentPlayer.value === playerColor.value && !wsGameOver.value)
const canUndo = computed(() => gameMode.value === 'PVE' ? (wsHistory.value || []).length >= 1 && !wsGameOver.value : (wsHistory.value || []).length >= 2 && !wsGameOver.value && isPlayerTurn.value)
const currentPlayerName = computed(() => wsCurrentPlayer.value === BLACK ? '黑方' : '白方')
const passColorName = computed(() => wsCurrentPlayer.value === BLACK ? '白方' : '黑方')

// 玩家面板优先显示后端返回的标签（包含 AI(code, level)）
const blackRole = computed(() => init.value?.players?.BLACK || (gameMode.value === 'PVP' ? '玩家' : (playerColor.value === BLACK ? '玩家' : 'AI')))
const whiteRole = computed(() => init.value?.players?.WHITE || (gameMode.value === 'PVP' ? '玩家' : (playerColor.value === WHITE ? '玩家' : 'AI')))

function handleStart(mode: GameMode, color: Color, size: number, selectedAlgorithm: string, selectedLevel: AILevel) {
  gameMode.value = mode
  playerColor.value = color === 'BLACK' ? BLACK : WHITE
  aiColor.value = playerColor.value === BLACK ? WHITE : BLACK
  boardSize.value = size
  aiAlgorithm.value = selectedAlgorithm
  aiLevel.value = selectedLevel
  hintAlgorithm.value = selectedAlgorithm
  // 玩家提示难度固定默认中档，不跟随 AI 难度变化
  hintLevel.value = 'normal'
  gameStarted.value = true
  joinGame(mode, color, size, selectedAlgorithm, selectedLevel)
}

function handlePlace(r: number, c: number) {
  if (!isPlayerTurn.value) return
  sendMove(r, c)
  lastMovePos.value = { r, c }
}

function handleUndo() {
  sendUndo()
  // 悔棋后，若提示开关开启，按当前提示配置重新请求
  if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value)
}

function toggleHint() {
  showHint.value = !showHint.value
  if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value)
}

function changeHintAlgorithm(value: string) {
  hintAlgorithm.value = value
  if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value)
}

function changeHintLevel(value: string) {
  hintLevel.value = (value as AILevel)
  if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value)
}

function handleBack() {
  gameStarted.value = false
  showHistory.value = false
  showHint.value = false
  lastMovePos.value = null
  flippedCells.value = []
}

function handleRestart() {
  joinGame(gameMode.value, playerColor.value === BLACK ? 'BLACK' : 'WHITE', boardSize.value, aiAlgorithm.value, aiLevel.value)
}

function coord(r: number, c: number): string { return String.fromCharCode(65 + c) + (r + 1) }

onMounted(() => { connect() })

watch(passEvent, (val) => {
  if (val) {
    passShown.value = true
    setTimeout(() => { passShown.value = false }, 1500)
  }
})

watch(wsCurrentPlayer, (val) => {
  if (gameMode.value === 'PVE' && val === aiColor.value && !wsGameOver.value) isThinking.value = true
  else isThinking.value = false
  // 玩家回合且提示开关开启时，自动拉取最新提示
  if (showHint.value && val === playerColor.value && !wsGameOver.value) requestHint(hintAlgorithm.value, hintLevel.value)
})

watch(wsBoard, () => { if (isThinking.value) isThinking.value = false })
watch(wsFlippedCells, (cells) => { flippedCells.value = cells || [] })
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Segoe UI', sans-serif; background: #1a1a2e; color: #eee; display: flex; justify-content: center; align-items: center; min-height: 100vh; user-select: none; }
.container { display: flex; flex-direction: column; align-items: center; gap: 20px; }
h1 { font-size: 2rem; letter-spacing: 4px; color: #e0e0e0; text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5); }
.connecting { display: flex; align-items: center; gap: 12px; padding: 24px 48px; background: #2a2a4a; border-radius: 12px; }
.connecting p { font-size: 1.1rem; color: #aaa; }
.status { font-size: 1.1rem; min-height: 1.5em; text-align: center; }
.board-layout { position: relative; display: flex; justify-content: center; width: 100%; }
.side-info { position: absolute; right: calc(50% + 210px); top: 8px; min-width: 260px; padding: 12px 14px; border-radius: 10px; background: rgba(255, 255, 255, 0.07); line-height: 1.7; white-space: nowrap; }
.side-info h3 { font-size: 0.95rem; margin-bottom: 6px; color: #d8d8d8; }
@media (max-width: 900px) {
  .board-layout { display: flex; flex-direction: column; align-items: center; gap: 10px; }
  .side-info { position: static; width: 100%; max-width: 460px; }
}
</style>

