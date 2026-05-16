<template>
  <div class="controls-wrap">
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
  </div>
</template>

<script setup lang="ts">
defineProps<{ canUndo: boolean; gameOver: boolean; showHistory: boolean; showHint: boolean; hintAlgorithm: string; hintLevel: string }>()
defineEmits<{ (e: 'undo'): void; (e: 'toggleHistory'): void; (e: 'toggleHint'): void; (e: 'back'): void; (e: 'hintAlgorithmChange', value: string): void; (e: 'hintLevelChange', value: string): void }>()

const algorithms = ['增强博弈', '主线剪枝', '蒙特树搜', '混合博弈']
</script>

<style scoped>
.controls-wrap {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
}

.btn-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}

button {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid rgba(255,255,255,0.18);
  border-radius: 10px;
  font-size: 0.92rem;
  cursor: pointer;
  background: linear-gradient(135deg, #7ef1c6, #58c99e);
  color: #09241c;
  font-weight: 700;
  transition: 0.2s ease;
}
button:hover { transform: translateY(-1px); filter: brightness(1.05); }
button:disabled { background: #555; color: #999; cursor: default; border-color: #555; }

.hint-config { display: grid; grid-template-columns: 1fr; gap: 8px; }

select {
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid rgba(255,255,255,0.24);
  background: rgba(25, 31, 63, 0.9);
  color: #e7ecff;
  width: 100%;
}
/* 固定画布模式：控制区不做响应式改列。 */
</style>
