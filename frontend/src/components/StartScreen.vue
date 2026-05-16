<template>
  <div class="start-screen">
    <div class="option-group">
      <label>模式</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: mode === 'PVE' }" @click="mode = 'PVE'">人机对战</div>
        <div class="pill" :class="{ selected: mode === 'PVP' }" @click="mode = 'PVP'">双人对战</div>
      </div>
    </div>

    <div class="option-group">
      <label>执子</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: color === 'BLACK' }" @click="color = 'BLACK'">黑方先手</div>
        <div class="pill" :class="{ selected: color === 'WHITE' }" @click="color = 'WHITE'">白方后手</div>
      </div>
    </div>

    <div class="option-group">
      <label>棋盘大小</label>
      <div class="pill-row">
        <div class="pill" :class="{ selected: size === 6 }" @click="size = 6">6 x 6</div>
        <div class="pill" :class="{ selected: size === 8 }" @click="size = 8">8 x 8</div>
        <div class="pill" :class="{ selected: size === 10 }" @click="size = 10">10 x 10</div>
      </div>
    </div>

    <div v-if="mode === 'PVE'" class="option-group">
      <label>AI算法</label>
      <div class="pill-row">
        <div class="pill" v-for="a in algorithms" :key="a" :class="{ selected: aiAlgorithm === a }" @click="aiAlgorithm = a">{{ a }}</div>
      </div>
      <div class="pill-row">
        <div class="pill" v-for="lv in levels" :key="lv" :class="{ selected: aiLevel === lv }" @click="aiLevel = lv">{{ lv }}</div>
      </div>
    </div>

    <button id="startBtn" @click="handleStart" :disabled="wsStatus !== 'connected'">
      {{ wsStatus === 'connected' ? '开始游戏' : '连接中...' }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { GameMode, Color, AILevel } from '../types'

defineProps<{ wsStatus: string }>()

const emit = defineEmits<{
  (e: 'start', mode: GameMode, color: Color, size: number, aiAlgorithm: string, aiLevel: AILevel): void
}>()

const mode = ref<GameMode>('PVE')
const color = ref<Color>('BLACK')
const size = ref(8)
const algorithms = ['增强博弈', '主线剪枝', '蒙特树搜', '混合博弈']
const levels: AILevel[] = ['easy', 'normal', 'hard']
const aiAlgorithm = ref(algorithms[0])
const aiLevel = ref<AILevel>('normal')

function handleStart() {
  emit('start', mode.value, color.value, size.value, aiAlgorithm.value, aiLevel.value)
}
</script>

<style scoped>
.start-screen { display: flex; flex-direction: column; align-items: center; gap: 28px; }
.option-group { display: flex; flex-direction: column; align-items: center; gap: 12px; }
.option-group label { font-size: 1rem; color: #aaa; text-transform: uppercase; letter-spacing: 2px; }
.pill-row { display: flex; gap: 8px; flex-wrap: wrap; justify-content: center; }
.pill { padding: 10px 20px; border: 2px solid #555; border-radius: 24px; cursor: pointer; font-size: 0.95rem; background: transparent; color: #ccc; transition: border-color 0.2s, background 0.2s, color 0.2s; }
.pill:hover { border-color: #52b788; color: #eee; }
.pill.selected { border-color: #52b788; background: #52b788; color: #111; font-weight: 600; }
#startBtn { padding: 14px 56px; font-size: 1.2rem; margin-top: 8px; }
#startBtn:disabled { background: #555; color: #999; cursor: default; }
</style>
