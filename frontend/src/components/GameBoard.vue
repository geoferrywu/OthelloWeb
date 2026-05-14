<template>
  <div v-if="boardReady" class="board" :style="{ gridTemplateColumns: `22px repeat(${size}, 1fr)` }">
    <div class="coord-label"></div>
    <div v-for="c in size" :key="'col'+c" class="coord-label">{{ colLabel(c - 1) }}</div>

    <template v-for="r in size" :key="'row'+r">
      <div class="coord-label">{{ r }}</div>
      <div
        v-for="c in size"
        :key="`${r}-${c}`"
        class="cell"
        :class="{ disabled: !isPlayerTurn }"
        @click="handleClick(r - 1, c - 1)"
      >
        <div
          v-if="cellAt(r - 1, c - 1) !== 0"
          class="disc"
          :class="[
            cellAt(r - 1, c - 1) === 1 ? 'black' : 'white',
            { flipping: isFlipped(r-1, c-1) }
          ]"
        ></div>
        <div
          v-else-if="isValidMove(r-1, c-1)"
          class="hint"
          :class="{ on: showHint && isBestMove(r-1, c-1) }"
        ></div>
      </div>
    </template>
  </div>
  <div v-else class="board-placeholder">Loading board...</div>
</template>

<script setup lang="ts">
import { computed, ref, watch, type PropType } from 'vue'
import type { Player, Position } from '../types'
import { getBestMove } from '../entities/game'

const props = defineProps<{
  board: number[][]
  currentPlayer: number
  playerColor: Player
  showHint: boolean
  isPlayerTurn: boolean
  lastMove: Position | null
  flippedCells: Position[]
}>()

const emit = defineEmits<{
  (e: 'place', r: number, c: number): void
}>()

const size = computed(() => props.board.length || 8)
const boardReady = computed(() => {
  const s = props.board.length
  if (s === 0) return false
  return props.board.every((row) => Array.isArray(row) && row.length === s)
})
const animatingFlips = ref<Set<string>>(new Set())
let clearFlipTimer: number | null = null

function cellAt(r: number, c: number): number {
  const row = props.board[r]
  if (!row || c < 0 || c >= row.length) return 0
  return row[c] ?? 0
}

function isFlipped(r: number, c: number): boolean {
  return animatingFlips.value.has(`${r},${c}`)
}

watch(() => props.flippedCells, (cells) => {
  if (cells && cells.length > 0) {
    if (clearFlipTimer !== null) {
      window.clearTimeout(clearFlipTimer)
      clearFlipTimer = null
    }
    animatingFlips.value.clear()
    for (const f of cells) {
      animatingFlips.value.add(`${f.r},${f.c}`)
    }
    clearFlipTimer = window.setTimeout(() => {
      animatingFlips.value.clear()
      clearFlipTimer = null
    }, 420)
  }
})

const validMovesSet = computed(() => {
  if (!boardReady.value) return new Set<string>()
  if (props.currentPlayer !== props.playerColor) return new Set<string>()
  const moves = new Set<string>()
  const s = size.value
  const board = props.board
  const player = props.currentPlayer as Player

  for (let r = 0; r < s; r++) {
    for (let c = 0; c < s; c++) {
      if (board[r][c] !== 0) continue
      if (getFlipsForHint(board, r, c, player, s).length > 0) {
        moves.add(`${r},${c}`)
      }
    }
  }
  return moves
})

const bestMove = computed(() => {
  if (!boardReady.value) return null
  if (!props.showHint || !props.isPlayerTurn) return null
  const s = size.value
  if (s === 0) return null
  const boardState: Player[][] = props.board.map(row => [...row]) as Player[][]
  return getBestMove(boardState, props.playerColor, s)
})

function isValidMove(r: number, c: number): boolean {
  return validMovesSet.value.has(`${r},${c}`)
}

function isBestMove(r: number, c: number): boolean {
  return bestMove.value !== null && bestMove.value.r === r && bestMove.value.c === c
}

function handleClick(r: number, c: number): void {
  if (!props.isPlayerTurn) return
  if (!isValidMove(r, c)) return
  emit('place', r, c)
}

function colLabel(c: number): string {
  return String.fromCharCode(65 + c)
}

function getFlipsForHint(board: number[][], r: number, c: number, player: Player, size: number): Position[] {
  if (board[r][c] !== 0) return []
  const opp = player === 1 ? 2 : 1
  const dirs: [number, number][] = [[-1,-1],[-1,0],[-1,1],[0,-1],[0,1],[1,-1],[1,0],[1,1]]
  const all: Position[] = []
  const inB = (r: number, c: number) => r >= 0 && r < size && c >= 0 && c < size

  for (const [dr, dc] of dirs) {
    let nr = r + dr, nc = c + dc
    const flips: Position[] = []
    while (inB(nr, nc) && board[nr][nc] === opp) {
      flips.push({ r: nr, c: nc })
      nr += dr
      nc += dc
    }
    if (flips.length > 0 && inB(nr, nc) && board[nr][nc] === player) {
      all.push(...flips)
    }
  }
  return all
}
</script>

<style scoped>
.board {
  display: grid;
  gap: 3px;
  padding: 12px;
  background: #2d6a4f;
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
}

.cell {
  width: 52px;
  height: 52px;
  background: #40916c;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  transition: background 0.15s;
}

.cell:hover { background: #52b788; }
.cell.disabled { cursor: default; pointer-events: none; }
.cell.disabled:hover { background: #40916c; }

.coord-label {
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255,255,255,0.5);
  font-size: 0.75rem;
  font-weight: 600;
  -webkit-user-select: none;
  user-select: none;
  pointer-events: none;
}

.disc {
  width: 76%;
  height: 76%;
  border-radius: 50%;
  position: absolute;
  transition: transform 0.35s ease;
}

.disc.black {
  background: radial-gradient(circle at 35% 35%, #555, #111);
  box-shadow: 2px 2px 6px rgba(0,0,0,0.5);
}

.disc.white {
  background: radial-gradient(circle at 35% 35%, #fff, #ccc);
  box-shadow: 2px 2px 6px rgba(0,0,0,0.3);
}

.disc.flipping { animation: flip 0.35s ease forwards; }

@keyframes flip {
  0%   { transform: rotateY(0deg); }
  50%  { transform: rotateY(90deg); }
  100% { transform: rotateY(0deg); }
}

.hint {
  width: 30%;
  height: 30%;
  border-radius: 50%;
  background: rgba(0,0,0,0.2);
  position: absolute;
  pointer-events: none;
}

.hint.on {
  background: #ff4444;
  box-shadow: 0 0 8px rgba(255,68,68,0.6);
  animation: pulse 1.2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(1.3); opacity: 0.7; }
}

.board-placeholder {
  min-width: 320px;
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #aaa;
}
</style>
