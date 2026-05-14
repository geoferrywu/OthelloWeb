<template>
  <div class="modal-overlay show">
    <div class="modal">
      <h2>Game Over</h2>
      <p>{{ message }}</p>
      <button @click="$emit('restart')">Play Again</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { GameOverData } from '../types'

const props = defineProps<{
  overData: GameOverData
}>()

defineEmits<{ (e: 'restart'): void }>()

const message = computed(() => {
  const { winner, blackScore, whiteScore } = props.overData
  if (winner === 'DRAW') return `${blackScore} : ${whiteScore} Draw`
  if (winner === 'BLACK') return `Black ${blackScore} : ${whiteScore} White - Black wins`
  return `White ${whiteScore} : ${blackScore} Black - White wins`
})
</script>

<style scoped>
.modal-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  justify-content: center;
  align-items: center;
}

.modal-overlay.show { display: flex; }

.modal {
  background: #2a2a4a;
  padding: 32px 48px;
  border-radius: 16px;
  text-align: center;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.5);
}

.modal h2 { font-size: 1.6rem; margin-bottom: 12px; }
.modal p { margin-bottom: 20px; font-size: 1.1rem; }

.modal button {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
  background: #52b788;
  color: #111;
  font-weight: 600;
  transition: background 0.2s;
}

.modal button:hover { background: #74c69d; }
</style>
