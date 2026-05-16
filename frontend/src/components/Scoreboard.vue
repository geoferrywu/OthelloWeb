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

const props = defineProps<{ board: number[][]; currentPlayer: number; isThinking: boolean }>()

const blackScore = computed(() => {
  let count = 0
  for (const row of props.board) for (const cell of row) if (cell === 1) count++
  return count
})

const whiteScore = computed(() => {
  let count = 0
  for (const row of props.board) for (const cell of row) if (cell === 2) count++
  return count
})
</script>

<style scoped>
.scoreboard {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  align-items: center;
  font-size: 1.02rem;
}

.player-score {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  min-width: 136px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.14);
  background: rgba(255, 255, 255, 0.05);
  transition: background 0.25s, box-shadow 0.25s;
}

.player-score.active {
  background: rgba(120, 214, 173, 0.17);
  box-shadow: 0 0 0 1px rgba(126, 241, 198, 0.28) inset, 0 8px 18px rgba(58, 177, 130, 0.18);
}

.disc-icon { width: 22px; height: 22px; border-radius: 50%; border: 2px solid #555; }
.disc-icon.black { background: #222; border-color: #444; }
.disc-icon.white { background: #f0f0f0; border-color: #ccc; }
</style>
