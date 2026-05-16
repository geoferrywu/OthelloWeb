<template>
  <div class="container">
    <h1>Othello</h1>

    <div v-if="wsStatus === 'connecting'" class="connecting"><p>正在连接...</p></div>

    <StartScreen v-else-if="!gameStarted" :wsStatus="wsStatus" @start="handleStart" />

    <template v-else>
      <section class="game-shell">
        <Scoreboard :board="currentBoard" :currentPlayer="wsCurrentPlayer" :isThinking="isThinking" />

        <div class="status">
          <span v-if="isOnlineMode && !onlineReady">{{ onlineWaitingText }}</span>
          <span v-else-if="isOnlineMode && countdown > 0">{{ onlineCountdownText }}</span>
          <span v-else-if="passShown">{{ passColorName }}跳过</span>
          <span v-else-if="isThinking && wsCurrentPlayer === aiColor">{{ currentPlayerName }}思考中...</span>
          <span v-else>{{ currentPlayerName }}落子</span>
        </div>

        <template v-if="boardReady">
          <div class="board-layout">
            <div class="left-panel">
              <aside class="side-info">
                <h3>对弈双方</h3>
                <p><strong>黑：</strong>{{ blackRole }}</p>
                <p><strong>白：</strong>{{ whiteRole }}</p>
              </aside>

              <ControlPanel
                :canUndo="canUndo"
                :gameOver="wsGameOver"
                :showHistory="showHistory"
                :showHint="showHint"
                :hintAlgorithm="hintAlgorithm"
                :hintLevel="hintLevel"
                :disableHint="isOnlineSpectator"
                @undo="handleUndo"
                @toggleHistory="showHistory = !showHistory"
                @toggleHint="toggleHint"
                @hintAlgorithmChange="changeHintAlgorithm"
                @hintLevelChange="changeHintLevel"
                @back="handleBack"
              />
            </div>

            <GameBoard
              :board="currentBoard"
              :currentPlayer="wsCurrentPlayer"
              :showHint="showHint || isOnlineSpectator"
              :showHistory="showHistory"
              :isPlayerTurn="isPlayerTurn"
              :allowPreviewMoves="isOnlineSpectator"
              :lastMove="lastMovePos"
              :flippedCells="flippedCells"
              :hintMove="hintMove"
              :historyEntries="moveLog"
              @place="handlePlace"
            />
          </div>
        </template>
        <div v-else class="status">棋盘数据加载中...</div>
      </section>

      <EndModal v-if="wsGameOver && overData && !isOnlineMode" :overData="overData" actionText="再来一局" @restart="handleRestart" />
      <EndModal v-if="wsGameOver && overData && isOnlineMode && overData.reason === 'NORMAL'" :overData="overData" :actionText="isOnlineSpectator ? '返回' : '再来一局'" @restart="handleBack" />
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
import EndModal from './components/EndModal.vue'

const { status: wsStatus, init, board: wsBoard, currentPlayer: wsCurrentPlayer, history: wsHistory, gameOver: wsGameOver, overData, passEvent, flippedCells: wsFlippedCells, hintMove, errorMessage, countdown, connect, joinGame, sendMove, sendUndo, requestHint, leaveGame } = useWebSocket()

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

const aiAlgorithm = ref('增强博弈')
const aiLevel = ref<AILevel>('normal')
const hintAlgorithm = ref('增强博弈')
const hintLevel = ref<AILevel>('normal')
const pairCode = ref('')

const BLACK: Player = 1
const WHITE: Player = 2

const currentBoard = computed(() => wsBoard.value || [])
const boardReady = computed(() => currentBoard.value.length > 0)
const isOnlineMode = computed(() => gameMode.value === 'PVP_ONLINE')
const onlineReady = computed(() => !!init.value?.online?.ready)
const isOnlineSpectator = computed(() => isOnlineMode.value && !!init.value?.online?.isSpectator)

const moveLog = computed(() => {
  return (wsHistory.value || []).map((m) => {
    const side = m.player === BLACK ? '黑' : '白'
    const base = m.position ? `${side}:${coord(m.position.r, m.position.c)}` : `${side}跳过`
    const text = m.position && m.hintTag ? `${base}_${m.hintTag}` : base
    return { color: m.player, text, pass: m.position === null }
  })
})

const isPlayerTurn = computed(() => {
  if (wsGameOver.value) return false
  if (isOnlineSpectator.value) return false
  if (gameMode.value === 'PVP') return true
  return wsCurrentPlayer.value === playerColor.value
})
const canUndo = computed(() => {
  if (gameMode.value === 'PVE') return (wsHistory.value || []).length >= 1 && !wsGameOver.value
  if (gameMode.value === 'PVP_ONLINE') return false
  return (wsHistory.value || []).length >= 2 && !wsGameOver.value && isPlayerTurn.value
})
const currentPlayerName = computed(() => wsCurrentPlayer.value === BLACK ? '黑方' : '白方')
const passColorName = computed(() => wsCurrentPlayer.value === BLACK ? '白方' : '黑方')
const onlineWaitingText = computed(() => {
  if (!isOnlineMode.value) return ''
  if (isOnlineSpectator.value) return '观战模式'
  const role = init.value?.online?.isHost ? '主玩家' : '客玩家'
  const code = init.value?.online?.pairCode || pairCode.value
  return `${role}（配对码 ${code}），等待对手加入...`
})
const onlineCountdownText = computed(() => {
  if (!isOnlineMode.value || countdown.value <= 0) return ''
  const isMyTurn = wsCurrentPlayer.value === playerColor.value
  return isMyTurn ? `请在 ${countdown.value} 秒内落子` : `对手剩余 ${countdown.value} 秒`
})

const blackRole = computed(() => {
  if (gameMode.value === 'PVP_ONLINE') {
    if (isOnlineSpectator.value) return ''
    return playerColor.value === BLACK ? '你' : '对手'
  }
  return init.value?.players?.BLACK || (gameMode.value === 'PVP' ? '玩家' : (playerColor.value === BLACK ? '玩家' : 'AI'))
})
const whiteRole = computed(() => {
  if (gameMode.value === 'PVP_ONLINE') {
    if (isOnlineSpectator.value) return ''
    return playerColor.value === WHITE ? '你' : '对手'
  }
  return init.value?.players?.WHITE || (gameMode.value === 'PVP' ? '玩家' : (playerColor.value === WHITE ? '玩家' : 'AI'))
})

function handleStart(mode: GameMode, color: Color, size: number, selectedAlgorithm: string, selectedLevel: AILevel, selectedPairCode?: string) {
  gameMode.value = mode
  playerColor.value = color === 'BLACK' ? BLACK : WHITE
  aiColor.value = playerColor.value === BLACK ? WHITE : BLACK
  boardSize.value = size
  pairCode.value = selectedPairCode || ''
  aiAlgorithm.value = selectedAlgorithm
  aiLevel.value = selectedLevel
  // 提示算法默认固定为第一个：增强博弈，不跟随AI算法
  hintAlgorithm.value = '增强博弈'
  hintLevel.value = 'normal'
  gameStarted.value = true
  joinGame(mode, color, size, selectedAlgorithm, selectedLevel, selectedPairCode)
}

function handlePlace(r: number, c: number) { if (!isPlayerTurn.value) return; sendMove(r, c); lastMovePos.value = { r, c } }
function handleUndo() { if (isOnlineMode.value) return; sendUndo(); if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value) }
function toggleHint() { showHint.value = !showHint.value; if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value) }
function changeHintAlgorithm(value: string) { hintAlgorithm.value = value; if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value) }
function changeHintLevel(value: string) { hintLevel.value = (value as AILevel); if (showHint.value) requestHint(hintAlgorithm.value, hintLevel.value) }

function handleBack() {
  if (gameStarted.value) leaveGame()
  gameStarted.value = false
  showHistory.value = false
  showHint.value = false
  lastMovePos.value = null
  flippedCells.value = []
  pairCode.value = ''
  connect()
}

function handleRestart() {
  // 重新开局时也恢复提示默认配置
  hintAlgorithm.value = '增强博弈'
  hintLevel.value = 'normal'
  joinGame(gameMode.value, playerColor.value === BLACK ? 'BLACK' : 'WHITE', boardSize.value, aiAlgorithm.value, aiLevel.value, pairCode.value || undefined)
}

function coord(r: number, c: number): string { return String.fromCharCode(65 + c) + (r + 1) }

onMounted(() => { connect() })
watch(passEvent, (val) => { if (val) { passShown.value = true; setTimeout(() => { passShown.value = false }, 1500) } })
watch(wsCurrentPlayer, (val) => {
  if (gameMode.value === 'PVE' && val === aiColor.value && !wsGameOver.value) isThinking.value = true
  else isThinking.value = false
  if (showHint.value && isPlayerTurn.value && !wsGameOver.value) requestHint(hintAlgorithm.value, hintLevel.value)
})
watch(errorMessage, (msg) => {
  if (!msg) return
  // Prevent blocking UI repaint with modal alerts during game updates.
  if (gameStarted.value) return
  window.alert(msg)
})
watch(wsGameOver, (v) => {
  if (!v || !isOnlineMode.value) return
  const reason = overData.value?.reason
  if (reason === 'NORMAL') return
  const message = overData.value?.message
  if (message) window.alert(message)
  handleBack()
})
watch(init, (val) => {
  if (!val || !isOnlineMode.value) return
  if (val.selfColor === BLACK || val.selfColor === WHITE) {
    playerColor.value = val.selfColor
    aiColor.value = val.selfColor === BLACK ? WHITE : BLACK
  }
})
watch(wsBoard, () => { if (isThinking.value) isThinking.value = false })
watch(wsFlippedCells, (cells) => { flippedCells.value = cells || [] })
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body {
  font-family: 'Segoe UI', sans-serif;
  background:
    radial-gradient(circle at 18% 14%, #2a2f66 0, rgba(42,47,102,0) 40%),
    radial-gradient(circle at 86% 84%, #1e5c4a 0, rgba(30,92,74,0) 35%),
    #12162f;
  color: #eee;
  display: block;
  min-height: 100vh;
  user-select: none;
  overflow-x: auto;
  overflow-y: auto;
}

.container {
  width: max-content;
  min-width: 1200px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  padding: 14px 0 20px;
  margin: 0 auto;
}

h1 { font-size: 2rem; letter-spacing: 4px; color: #e0e0e0; text-shadow: 0 2px 8px rgba(0, 0, 0, 0.5); }

.game-shell {
  width: max-content;
  border-radius: 18px;
  padding: 16px;
  border: 1px solid rgba(255,255,255,0.12);
  background: rgba(14, 19, 45, 0.72);
  box-shadow: 0 18px 40px rgba(0,0,0,0.45);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}

.connecting { display: flex; align-items: center; gap: 12px; padding: 24px 48px; background: #2a2a4a; border-radius: 12px; }
.connecting p { font-size: 1.1rem; color: #aaa; }

.status {
  font-size: 1.04rem;
  min-height: 1.5em;
  text-align: center;
  color: #dce9ff;
}

.board-layout {
  width: max-content;
  display: grid;
  grid-template-columns: 280px max-content;
  gap: 14px;
  align-items: stretch;
}

.left-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 280px;
}

.side-info {
  padding: 14px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.1);
  background: rgba(255, 255, 255, 0.05);
  line-height: 1.9;
  white-space: nowrap;
}

.side-info h3 {
  font-size: 0.95rem;
  margin-bottom: 6px;
  color: #d8d8d8;
}

/* 固定画布布局：不做响应式折叠。窗口更小时使用浏览器横向滚动条。 */
</style>
