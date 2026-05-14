<template>
  <div class="start-screen">
    <div class="option-group">
      <label>Mode</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: mode === 'PVE' }" @click="mode = 'PVE'">PvE</div>
        <div class="pill" :class="{ selected: mode === 'PVP' }" @click="mode = 'PVP'">PvP</div>
      </div>
    </div>

    <div class="option-group">
      <label>Color</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: color === 'BLACK' }" @click="color = 'BLACK'">Black</div>
        <div class="pill" :class="{ selected: color === 'WHITE' }" @click="color = 'WHITE'">White</div>
      </div>
    </div>

    <div class="option-group">
      <label>Board Size</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: size === 6 }" @click="size = 6">6 x 6</div>
        <div class="pill" :class="{ selected: size === 8 }" @click="size = 8">8 x 8</div>
        <div class="pill" :class="{ selected: size === 10 }" @click="size = 10">10 x 10</div>
      </div>
    </div>

    <button id="startBtn" @click="handleStart" :disabled="wsStatus !== 'connected'">
      {{ wsStatus === 'connected' ? 'Start Game' : 'Connecting...' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { GameMode, Color } from '../types'

defineProps<{
  wsStatus: string
}>()

const emit = defineEmits<{
  (e: 'start', mode: GameMode, color: Color, size: number): void
}>()

const mode = ref<GameMode>('PVE')
const color = ref<Color>('BLACK')
const size = ref(8)

function handleStart() {
  emit('start', mode.value, color.value, size.value)
}
</script>

<style scoped>
.start-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 28px;
}

.option-group {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.option-group label {
  font-size: 1rem;
  color: #aaa;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.pill-row {
  display: flex;
  gap: 8px;
}

.pill {
  padding: 10px 28px;
  border: 2px solid #555;
  border-radius: 24px;
  cursor: pointer;
  font-size: 1rem;
  background: transparent;
  color: #ccc;
  transition: border-color 0.2s, background 0.2s, color 0.2s;
}

.pill:hover { border-color: #52b788; color: #eee; }

.pill.selected {
  border-color: #52b788;
  background: #52b788;
  color: #111;
  font-weight: 600;
}

#startBtn {
  padding: 14px 56px;
  font-size: 1.2rem;
  margin-top: 8px;
}

#startBtn:disabled {
  background: #555;
  color: #999;
  cursor: default;
}
</style>
