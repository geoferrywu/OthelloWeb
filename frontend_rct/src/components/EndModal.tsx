import { useEffect, useMemo, useRef, useState } from 'react'
import type { GameOverData } from '../types'

interface Props {
  overData: GameOverData
  actionText?: string
  onRestart: () => void
}

/**
 * 可拖拽终局弹窗。
 * 这里保留和 Vue 版本一致的“拖拽 + 居中默认定位”交互。
 */
export default function EndModal({ overData, actionText, onRestart }: Props) {
  const modalRef = useRef<HTMLDivElement | null>(null)
  const [position, setPosition] = useState({ x: 0, y: 0 })
  const [dragging, setDragging] = useState(false)
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 })

  const message = useMemo(() => {
    const { winner, blackScore, whiteScore } = overData
    if (winner === 'DRAW') return `黑${blackScore} : ${whiteScore} 白，平局`
    if (winner === 'BLACK') return `黑${blackScore} : ${whiteScore} 白，黑方获胜`
    return `白${whiteScore} : ${blackScore} 黑，白方获胜`
  }, [overData])

  const label = actionText || '再来一局'

  const clampPosition = (nextX: number, nextY: number) => {
    const modal = modalRef.current
    if (!modal) return { x: nextX, y: nextY }
    const maxX = Math.max(8, window.innerWidth - modal.offsetWidth - 8)
    const maxY = Math.max(8, window.innerHeight - modal.offsetHeight - 8)
    return {
      x: Math.min(Math.max(8, nextX), maxX),
      y: Math.min(Math.max(8, nextY), maxY),
    }
  }

  useEffect(() => {
    const modal = modalRef.current
    if (!modal) return
    const x = (window.innerWidth - modal.offsetWidth) / 2
    const y = (window.innerHeight - modal.offsetHeight) / 2
    setPosition(clampPosition(x, y))
  }, [])

  useEffect(() => {
    const onMove = (e: MouseEvent) => {
      if (!dragging) return
      const nextX = e.clientX - dragOffset.x
      const nextY = e.clientY - dragOffset.y
      setPosition(clampPosition(nextX, nextY))
    }
    const onUp = () => setDragging(false)
    const onResize = () => setPosition((prev) => clampPosition(prev.x, prev.y))

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    window.addEventListener('resize', onResize)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
      window.removeEventListener('resize', onResize)
    }
  }, [dragging, dragOffset])

  const onMouseDown = (e: React.MouseEvent) => {
    setDragging(true)
    setDragOffset({ x: e.clientX - position.x, y: e.clientY - position.y })
  }

  return (
    <div className="floating-wrap">
      <div ref={modalRef} className="modal" style={{ left: `${position.x}px`, top: `${position.y}px` }}>
        <h2 className="drag-handle" onMouseDown={onMouseDown}>对局结束</h2>
        <p>{message}</p>
        <button onClick={onRestart}>{label}</button>
      </div>
    </div>
  )
}
