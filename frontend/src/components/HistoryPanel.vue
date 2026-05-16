<template>
  <div ref="panelRef" class="history-panel" :style="panelStyle">
    <div class="history-header" @mousedown="startDrag">
      <h3>对局记录</h3>
      <button class="history-close" @click="$emit('close')">&times;</button>
    </div>
    <div class="history-list" ref="listRef">
      <div
        v-for="(m, i) in history"
        :key="i"
        class="history-item"
        :class="[m.pass ? 'pass' : (m.color === 1 ? 'black-move' : 'white-move')]"
      >{{ m.text }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, watch } from 'vue'

interface MoveEntry {
  color: number
  text: string
  pass: boolean
}

const props = defineProps<{
  history: MoveEntry[]
}>()

defineEmits<{ (e: 'close'): void }>()

const panelRef = ref<HTMLElement | null>(null)
const listRef = ref<HTMLElement | null>(null)
const panelStyle = ref({ left: '', top: '', right: '20px', position: 'fixed' as const })

let dragging = false
let dragOffX = 0
let dragOffY = 0

function startDrag(e: MouseEvent) {
  dragging = true
  const el = panelRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  dragOffX = e.clientX - rect.left
  dragOffY = e.clientY - rect.top

  const onMove = (ev: MouseEvent) => {
    if (!dragging) return
    panelStyle.value = {
      left: `${ev.clientX - dragOffX}px`,
      top: `${ev.clientY - dragOffY}px`,
      right: 'auto',
      position: 'fixed',
    }
  }

  const onUp = () => {
    dragging = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }

  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

watch(() => props.history.length, async () => {
  await nextTick()
  if (listRef.value) {
    listRef.value.scrollTop = listRef.value.scrollHeight
  }
})
</script>

<style scoped>
.history-panel {
  right: 20px;
  top: 100px;
  z-index: 100;
  width: 300px;
  height: 420px;
  background: #2a2a4a;
  border-radius: 10px;
  padding: 10px 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  position: fixed;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  cursor: grab;
  -webkit-user-select: none;
  user-select: none;
}

.history-header:active { cursor: grabbing; }

.history-header h3 {
  font-size: 0.85rem;
  color: #aaa;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin: 0;
}

.history-close {
  background: none;
  border: none;
  color: #888;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 0 2px;
  line-height: 1;
}

.history-close:hover { color: #eee; }

.history-list {
  overflow-y: auto;
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  cursor: text;
}

.history-item {
  font-size: 0.85rem;
  line-height: 1.45;
  padding: 3px 6px;
  border-radius: 4px;
  white-space: nowrap;
  overflow: visible;
  cursor: text;
  -webkit-user-select: text;
  user-select: text;
}

.history-item.black-move { color: #ddd; }
.history-item.white-move { color: #ccc; }
.history-item.pass { color: #777; font-style: italic; }
</style>

