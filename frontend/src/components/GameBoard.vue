<template>
  <div v-if="boardReady" class="board-shell">
    <div class="board" :class="{ 'history-open': showHistory }" :style="boardStyle">
      <div class="coord-label"></div>
      <div v-for="c in size" :key="'col'+c" class="coord-label">{{ colLabel(c - 1) }}</div>

      <template v-for="r in size" :key="'row'+r">
        <div class="coord-label">{{ r }}</div>
        <div v-for="c in size" :key="`${r}-${c}`" class="cell" :class="{ disabled: !isPlayerTurn }" @click="handleClick(r - 1, c - 1)">
          <div v-if="cellAt(r - 1, c - 1) !== 0" class="disc" :class="[cellAt(r - 1, c - 1) === 1 ? 'black' : 'white', { flipping: isFlipped(r-1, c-1) }]" ></div>
          <div v-else-if="isValidMove(r-1, c-1)" class="hint" :class="{ on: showHint && isBestMove(r-1, c-1) }"></div>
        </div>
      </template>

      <aside v-if="showHistory" class="board-history">
        <h4>对局记录</h4>
        <div ref="historyListRef" class="board-history-list">
          <div
            v-for="(m, i) in historyEntries"
            :key="i"
            class="board-history-item"
            :class="[m.pass ? 'pass' : (m.color === 1 ? 'black-move' : 'white-move')]"
          >
            {{ m.text }}
          </div>
        </div>
      </aside>
    </div>
  </div>
  <div v-else class="board-placeholder">Loading board...</div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick } from 'vue'
import type { Player, Position } from '../types'

const props = defineProps<{
  board: number[][]
  currentPlayer: number
  showHint: boolean
  showHistory: boolean
  isPlayerTurn: boolean
  lastMove: Position | null
  flippedCells: Position[]
  hintMove: Position | null
  historyEntries: Array<{ color: number; text: string; pass: boolean }>
}>()

const emit = defineEmits<{ (e: 'place', r: number, c: number): void }>()
const size = computed(() => props.board.length || 8)
const boardReady = computed(() => {
  const s = props.board.length
  if (s === 0) return false
  return props.board.every((row) => Array.isArray(row) && row.length === s)
})

const boardStyle = computed(() => {
  const s = size.value
  const px = s <= 6 ? 64 : s === 8 ? 52 : 42
  const historyWidth = s <= 6 ? 198 : s === 8 ? 186 : 170
  const historyGap = s <= 6 ? 8 : s === 8 ? 6 : 5
  return {
    gridTemplateColumns: `24px repeat(${s}, var(--cell-size))`,
    '--cell-size': `${px}px`,
    '--history-width': `${historyWidth}px`,
    '--history-gap': `${historyGap}px`,
  } as Record<string, string>
})

const animatingFlips = ref<Set<string>>(new Set())
const historyListRef = ref<HTMLElement | null>(null)
let clearFlipTimer: number | null = null

function cellAt(r: number, c: number): number { const row = props.board[r]; if (!row || c < 0 || c >= row.length) return 0; return row[c] ?? 0 }
function isFlipped(r: number, c: number): boolean { return animatingFlips.value.has(`${r},${c}`) }

watch(() => props.flippedCells, (cells) => {
  if (cells && cells.length > 0) {
    if (clearFlipTimer !== null) { window.clearTimeout(clearFlipTimer); clearFlipTimer = null }
    animatingFlips.value.clear()
    for (const f of cells) animatingFlips.value.add(`${f.r},${f.c}`)
    clearFlipTimer = window.setTimeout(() => { animatingFlips.value.clear(); clearFlipTimer = null }, 420)
  }
})

watch(() => props.historyEntries.length, async () => {
  await nextTick()
  if (historyListRef.value) {
    historyListRef.value.scrollTop = historyListRef.value.scrollHeight
  }
})

const validMovesSet = computed(() => {
  if (!boardReady.value) return new Set<string>()
  if (!props.isPlayerTurn) return new Set<string>()
  const moves = new Set<string>()
  const s = size.value
  const board = props.board
  const player = props.currentPlayer as Player
  for (let r = 0; r < s; r++) {
    for (let c = 0; c < s; c++) {
      if (board[r][c] !== 0) continue
      if (getFlipsForHint(board, r, c, player, s).length > 0) moves.add(`${r},${c}`)
    }
  }
  return moves
})

function isValidMove(r: number, c: number): boolean { return validMovesSet.value.has(`${r},${c}`) }
function isBestMove(r: number, c: number): boolean { return props.hintMove !== null && props.hintMove.r === r && props.hintMove.c === c }
function handleClick(r: number, c: number): void { if (!props.isPlayerTurn) return; if (!isValidMove(r, c)) return; emit('place', r, c) }
function colLabel(c: number): string { return String.fromCharCode(65 + c) }

function getFlipsForHint(board: number[][], r: number, c: number, player: Player, size: number): Position[] {
  if (board[r][c] !== 0) return []
  const opp = player === 1 ? 2 : 1
  const dirs: [number, number][] = [[-1,-1],[-1,0],[-1,1],[0,-1],[0,1],[1,-1],[1,0],[1,1]]
  const all: Position[] = []
  const inB = (r: number, c: number) => r >= 0 && r < size && c >= 0 && c < size
  for (const [dr, dc] of dirs) {
    let nr = r + dr, nc = c + dc
    const flips: Position[] = []
    while (inB(nr, nc) && board[nr][nc] === opp) { flips.push({ r: nr, c: nc }); nr += dr; nc += dc }
    if (flips.length > 0 && inB(nr, nc) && board[nr][nc] === player) all.push(...flips)
  }
  return all
}
</script>

<style scoped>
.board-shell {
  max-width: none;
  overflow: visible;
  padding-bottom: 6px;
  display: flex;
  justify-content: flex-start;
}

.board {
  display: grid;
  width: max-content;
  gap: 4px;
  padding: 14px;
  background: radial-gradient(circle at 20% 15%, #3f8a6d, #255a45 68%);
  border-radius: 16px;
  box-shadow: 0 16px 36px rgba(0, 0, 0, 0.45), inset 0 1px 0 rgba(255, 255, 255, 0.15);
  position: relative;
}

.board.history-open {
  padding-right: calc(var(--history-width) + var(--history-gap) + 10px);
}

.cell {
  width: var(--cell-size);
  height: var(--cell-size);
  background: linear-gradient(165deg, #4aa57b, #358960);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  transition: transform 0.12s ease, filter 0.12s ease;
}

.cell:hover { transform: translateY(-1px); filter: brightness(1.08); }
.cell.disabled { cursor: default; pointer-events: none; }
.cell.disabled:hover { transform: none; filter: none; }

.coord-label {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(240, 250, 245, 0.75);
  font-size: 0.76rem;
  font-weight: 700;
  user-select: none;
  pointer-events: none;
}

.disc { width: 76%; height: 76%; border-radius: 50%; position: absolute; transition: transform 0.35s ease; }
.disc.black { background: radial-gradient(circle at 35% 35%, #6a6a6a, #121212); box-shadow: 3px 3px 8px rgba(0,0,0,0.5); }
.disc.white { background: radial-gradient(circle at 35% 35%, #fff, #cfcfcf); box-shadow: 2px 2px 6px rgba(0,0,0,0.28); }
.disc.flipping { animation: flip 0.35s ease forwards; }
@keyframes flip { 0% { transform: rotateY(0deg);} 50% { transform: rotateY(90deg);} 100% { transform: rotateY(0deg);} }

.hint { width: 28%; height: 28%; border-radius: 50%; background: rgba(15, 28, 22, 0.35); position: absolute; pointer-events: none; }
.hint.on { background: #ff5e58; box-shadow: 0 0 10px rgba(255,94,88,0.72); animation: pulse 1.2s ease-in-out infinite; }
@keyframes pulse { 0%, 100% { transform: scale(1); opacity: 1; } 50% { transform: scale(1.22); opacity: 0.72; } }

.board-placeholder {
  min-width: 320px;
  min-height: 220px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #b8bfd2;
  border-radius: 14px;
  background: rgba(255,255,255,0.06);
}

.board-history {
  position: absolute;
  right: 6px;
  top: 12px;
  bottom: 12px;
  width: var(--history-width);
  padding: 10px;
  border-radius: 10px;
  border: 1px solid rgba(255,255,255,0.14);
  background: rgba(23, 79, 61, 0.5);
  backdrop-filter: blur(2px);
  display: flex;
  flex-direction: column;
}

.board-history h4 {
  font-size: 0.9rem;
  color: #d8f1e5;
  margin-bottom: 8px;
}

.board-history-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.board-history-item {
  font-size: 0.82rem;
  line-height: 1.35;
  padding: 3px 4px;
  border-radius: 4px;
  white-space: nowrap;
}

.board-history-item.black-move { color: #f2f6ff; }
.board-history-item.white-move { color: #edf8f2; }
.board-history-item.pass { color: #b8d5c7; font-style: italic; }

/* 固定画布布局：不做响应式缩放。 */
</style>
