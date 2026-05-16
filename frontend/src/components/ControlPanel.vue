<template>
  <div class="btn-row">
    <button :disabled="!canUndo || gameOver" @click="$emit('undo')">悔棋</button>
    <button @click="$emit('toggleHistory')">{{ showHistory ? '隐藏记录' : '显示记录' }}</button>
    <button @click="$emit('toggleHint')">提示: {{ showHint ? '开' : '关' }}</button>
    <button @click="$emit('back')">返回</button>
  </div>
  <div v-if="showHint" class="hint-config">
    <select :value="hintAlgorithm" @change="$emit('hintAlgorithmChange', ($event.target as HTMLSelectElement).value)">
      <option v-for="a in algorithms" :key="a" :value="a">{{ a }}</option>
    </select>
    <select :value="hintLevel" @change="$emit('hintLevelChange', ($event.target as HTMLSelectElement).value)">
      <option value="easy">easy</option>
      <option value="normal">normal</option>
      <option value="hard">hard</option>
    </select>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  canUndo: boolean
  gameOver: boolean
  showHistory: boolean
  showHint: boolean
  hintAlgorithm: string
  hintLevel: string
}>()

defineEmits<{
  (e: 'undo'): void
  (e: 'toggleHistory'): void
  (e: 'toggleHint'): void
  (e: 'back'): void
  (e: 'hintAlgorithmChange', value: string): void
  (e: 'hintLevelChange', value: string): void
}>()

const algorithms = ['增强博弈', '主线剪枝', '蒙特树搜', '混合博弈']
</script>

<style scoped>
.btn-row { display: flex; gap: 12px; }
button { padding: 10px 24px; border: none; border-radius: 8px; font-size: 1rem; cursor: pointer; background: #52b788; color: #111; font-weight: 600; transition: background 0.2s; }
button:hover { background: #74c69d; }
button:disabled { background: #555; color: #999; cursor: default; }
.hint-config { margin-top: 10px; display: flex; gap: 8px; }
select { padding: 8px 10px; border-radius: 8px; border: 1px solid #666; background: #222; color: #ddd; }
</style>
