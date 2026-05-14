<template>
  <div class="floating-wrap">
    <div
      ref="modalRef"
      class="modal"
      :style="{ left: `${position.x}px`, top: `${position.y}px` }"
    >
      <h2
        class="drag-handle"
        @mousedown="startDragMouse"
        @touchstart.prevent="startDragTouch"
      >
        对局结束
      </h2>
      <p>{{ message }}</p>
      <button @click="$emit('restart')">再来一局</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import type { GameOverData } from '../types'

const props = defineProps<{
  overData: GameOverData
}>()

defineEmits<{ (e: 'restart'): void }>()

const modalRef = ref<HTMLElement | null>(null)
const position = ref({ x: 0, y: 0 })
const dragOffset = ref({ x: 0, y: 0 })
const dragging = ref(false)

const message = computed(() => {
  const { winner, blackScore, whiteScore } = props.overData
  if (winner === 'DRAW') return `黑${blackScore} : ${whiteScore} 白，平局`
  if (winner === 'BLACK') return `黑${blackScore} : ${whiteScore} 白，黑方获胜`
  return `白${whiteScore} : ${blackScore} 黑，白方获胜`
})

function clampPosition(nextX: number, nextY: number) {
  const modal = modalRef.value
  if (!modal) return { x: nextX, y: nextY }
  const maxX = Math.max(8, window.innerWidth - modal.offsetWidth - 8)
  const maxY = Math.max(8, window.innerHeight - modal.offsetHeight - 8)
  return {
    x: Math.min(Math.max(8, nextX), maxX),
    y: Math.min(Math.max(8, nextY), maxY),
  }
}

function setDefaultPosition() {
  const modal = modalRef.value
  if (!modal) return
  const x = (window.innerWidth - modal.offsetWidth) / 2
  const y = (window.innerHeight - modal.offsetHeight) / 2
  position.value = clampPosition(x, y)
}

function startDrag(clientX: number, clientY: number) {
  dragging.value = true
  dragOffset.value = {
    x: clientX - position.value.x,
    y: clientY - position.value.y,
  }
}

function handleDrag(clientX: number, clientY: number) {
  if (!dragging.value) return
  const nextX = clientX - dragOffset.value.x
  const nextY = clientY - dragOffset.value.y
  position.value = clampPosition(nextX, nextY)
}

function endDrag() {
  dragging.value = false
}

function startDragMouse(e: MouseEvent) {
  startDrag(e.clientX, e.clientY)
}

function startDragTouch(e: TouchEvent) {
  const t = e.touches[0]
  if (!t) return
  startDrag(t.clientX, t.clientY)
}

function onMouseMove(e: MouseEvent) {
  handleDrag(e.clientX, e.clientY)
}

function onTouchMove(e: TouchEvent) {
  const t = e.touches[0]
  if (!t) return
  handleDrag(t.clientX, t.clientY)
}

function onResize() {
  position.value = clampPosition(position.value.x, position.value.y)
}

onMounted(() => {
  setDefaultPosition()
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', endDrag)
  window.addEventListener('touchmove', onTouchMove, { passive: true })
  window.addEventListener('touchend', endDrag)
  window.addEventListener('resize', onResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', endDrag)
  window.removeEventListener('touchmove', onTouchMove)
  window.removeEventListener('touchend', endDrag)
  window.removeEventListener('resize', onResize)
})
</script>

<style scoped>
.floating-wrap {
  position: fixed;
  inset: 0;
  pointer-events: none;
}

.modal {
  position: fixed;
  pointer-events: auto;
  background: #2a2a4a;
  padding: 20px 24px 24px;
  border-radius: 16px;
  text-align: center;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.5);
  min-width: 240px;
}

.drag-handle {
  font-size: 1.2rem;
  margin-bottom: 12px;
  cursor: move;
  user-select: none;
}

.modal p {
  margin-bottom: 20px;
  font-size: 1.1rem;
}

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

.modal button:hover {
  background: #74c69d;
}
</style>
