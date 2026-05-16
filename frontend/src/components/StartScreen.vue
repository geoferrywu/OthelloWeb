<template>
  <div class="start-screen">
    <div class="bg-orb orb-a"></div>
    <div class="bg-orb orb-b"></div>

    <section class="panel">
      <header class="panel-header">
        <p class="eyebrow">策略对局</p>
        <h2>选择你的开局配置</h2>
      </header>

      <div class="option-group">
        <label>模式</label>
        <div class="pill-row">
          <button class="pill" :class="{ selected: mode === 'PVE' }" @click="mode = 'PVE'">人机对战</button>
          <button class="pill" :class="{ selected: mode === 'PVP' }" @click="mode = 'PVP'">双人对战</button>
        </div>
      </div>

      <div class="option-group">
        <label>执子</label>
        <div class="pill-row">
          <button class="pill" :class="{ selected: color === 'BLACK' }" @click="color = 'BLACK'">黑方先手</button>
          <button class="pill" :class="{ selected: color === 'WHITE' }" @click="color = 'WHITE'">白方后手</button>
        </div>
      </div>

      <div class="option-group">
        <label>棋盘大小</label>
        <div class="pill-row">
          <button class="pill" :class="{ selected: size === 6 }" @click="size = 6">6 x 6</button>
          <button class="pill" :class="{ selected: size === 8 }" @click="size = 8">8 x 8</button>
          <button class="pill" :class="{ selected: size === 10 }" @click="size = 10">10 x 10</button>
        </div>
      </div>

      <div v-if="mode === 'PVE'" class="option-group">
        <label>AI 算法</label>
        <div class="pill-row">
          <button class="pill" v-for="a in algorithms" :key="a" :class="{ selected: aiAlgorithm === a }" @click="aiAlgorithm = a">{{ a }}</button>
        </div>
        <div class="pill-row compact">
          <button class="pill" v-for="lv in levels" :key="lv" :class="{ selected: aiLevel === lv }" @click="aiLevel = lv">{{ lv }}</button>
        </div>
      </div>

      <button id="startBtn" @click="handleStart" :disabled="wsStatus !== 'connected'">
        {{ wsStatus === 'connected' ? '开始游戏' : '连接中...' }}
      </button>
    </section>
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
.start-screen {
  position: relative;
  width: min(760px, 94vw);
  padding: 8px;
}

.panel {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 28px;
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  background:
    linear-gradient(155deg, rgba(45, 43, 83, 0.95), rgba(25, 31, 63, 0.95));
  box-shadow:
    0 20px 55px rgba(0, 0, 0, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

.panel-header h2 {
  font-size: clamp(1.35rem, 3.2vw, 1.9rem);
  margin-top: 6px;
  color: #f6f7ff;
}

.eyebrow {
  font-size: 0.8rem;
  letter-spacing: 0.18em;
  color: #9de1c4;
  text-transform: uppercase;
}

.option-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.option-group label {
  font-size: 0.88rem;
  color: #cfd4ff;
  letter-spacing: 0.08em;
}

.pill-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pill {
  border: 1px solid rgba(255, 255, 255, 0.26);
  border-radius: 999px;
  padding: 9px 16px;
  font-size: 0.92rem;
  color: #e9ecff;
  background: rgba(255, 255, 255, 0.04);
  cursor: pointer;
  transition: 0.18s ease;
}

.pill:hover {
  transform: translateY(-1px);
  border-color: rgba(123, 238, 188, 0.8);
  box-shadow: 0 7px 18px rgba(30, 176, 123, 0.26);
}

.pill.selected {
  color: #08241d;
  border-color: #66d3a8;
  background: linear-gradient(135deg, #7ef1c6, #58c99e);
  font-weight: 700;
}

.compact .pill {
  min-width: 92px;
  text-transform: lowercase;
}

#startBtn {
  margin-top: 6px;
  border: none;
  border-radius: 12px;
  padding: 14px 18px;
  font-size: 1.03rem;
  font-weight: 700;
  color: #0a251d;
  background: linear-gradient(135deg, #8ff7cf, #5bcfa2);
  box-shadow: 0 12px 26px rgba(48, 196, 141, 0.35);
  cursor: pointer;
  transition: 0.2s ease;
}

#startBtn:hover:enabled {
  transform: translateY(-1px);
  filter: brightness(1.05);
}

#startBtn:disabled {
  background: #4f576f;
  color: #b7c0d3;
  box-shadow: none;
  cursor: default;
}

.bg-orb {
  position: absolute;
  z-index: 1;
  border-radius: 999px;
  filter: blur(44px);
  opacity: 0.45;
}

.orb-a {
  width: 220px;
  height: 220px;
  top: -36px;
  left: -46px;
  background: #66d3a8;
}

.orb-b {
  width: 180px;
  height: 180px;
  right: -24px;
  bottom: -24px;
  background: #6a86ff;
}

@media (max-width: 680px) {
  .panel {
    padding: 20px;
    gap: 18px;
  }

  .pill {
    padding: 8px 14px;
    font-size: 0.88rem;
  }
}
</style>
