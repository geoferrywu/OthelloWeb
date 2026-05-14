<template>
  <div class="btn-row">
    <button :disabled="!canUndo || gameOver" @click="$emit('undo')">Undo</button>
    <button @click="$emit('toggleHistory')">
      {{ showHistory ? 'Hide History' : 'Show History' }}
    </button>
    <button @click="$emit('toggleHint')">
      Hint: {{ showHint ? 'On' : 'Off' }}
    </button>
    <button @click="$emit('back')">Back</button>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  canUndo: boolean
  gameOver: boolean
  showHistory: boolean
  showHint: boolean
}>()

defineEmits<{
  (e: 'undo'): void
  (e: 'toggleHistory'): void
  (e: 'toggleHint'): void
  (e: 'back'): void
}>()
</script>

<style scoped>
.btn-row {
  display: flex;
  gap: 12px;
}

button {
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

button:hover { background: #74c69d; }
button:disabled { background: #555; color: #999; cursor: default; }
</style>
