<template>
  <div class="scoreboard">
    <div class="player-score" :class="{ active: currentPlayer === 1 }">
      <span class="disc-icon black"></span>
      <span>黑方 <strong>{{ blackScore }}</strong></span>
    </div>
    <div class="player-score" :class="{ active: currentPlayer === 2 }">
      <span class="disc-icon white"></span>
      <span>白方 <strong>{{ whiteScore }}</strong></span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  board: number[][]
  currentPlayer: number
  isThinking: boolean
}>()

const blackScore = computed(() => {
  let count = 0
  for (const row of props.board) {
    for (const cell of row) {
      if (cell === 1) count++
    }
  }
  return count
})

const whiteScore = computed(() => {
  let count = 0
  for (const row of props.board) {
    for (const cell of row) {
      if (cell === 2) count++
    }
  }
  return count
})
</script>

<style scoped>
.scoreboard { display: flex; gap: 40px; align-items: center; font-size: 1.1rem; }
.player-score { display: flex; align-items: center; gap: 10px; padding: 8px 16px; border-radius: 8px; background: rgba(255, 255, 255, 0.05); transition: background 0.3s, box-shadow 0.3s; }
.player-score.active { background: rgba(255, 255, 255, 0.12); box-shadow: 0 0 12px rgba(255, 255, 255, 0.15); }
.disc-icon { width: 24px; height: 24px; border-radius: 50%; border: 2px solid #555; }
.disc-icon.black { background: #222; border-color: #444; }
.disc-icon.white { background: #f0f0f0; border-color: #ccc; }
</style>
