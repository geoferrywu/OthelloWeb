import { useState } from 'react'
import type { AILevel, Color, GameMode } from '../types'

interface Props {
  wsStatus: string
  onStart: (mode: GameMode, color: Color, size: number, aiAlgorithm: string, aiLevel: AILevel, pairCode?: string) => void
}

/**
 * 开始页：负责收集开局参数。
 * 注意：在线模式必须输入 4 位数字配对码。
 */
export default function StartScreen({ wsStatus, onStart }: Props) {
  const [mode, setMode] = useState<GameMode>('PVE')
  const [color, setColor] = useState<Color>('BLACK')
  const [size, setSize] = useState(8)
  const algorithms = ['增强博弈', '主线剪枝', '蒙特树搜', '混合博弈']
  const levels: AILevel[] = ['easy', 'normal', 'hard']
  const [aiAlgorithm, setAiAlgorithm] = useState(algorithms[0])
  const [aiLevel, setAiLevel] = useState<AILevel>('normal')
  const [pairCode, setPairCode] = useState('')

  const handleStart = () => {
    if (mode === 'PVP_ONLINE') {
      const code = pairCode.trim()
      if (!/^\d{4}$/.test(code)) return
      onStart(mode, color, size, aiAlgorithm, aiLevel, code)
      return
    }
    onStart(mode, color, size, aiAlgorithm, aiLevel)
  }

  return (
    <div className="start-screen">
      <div className="bg-orb orb-a" />
      <div className="bg-orb orb-b" />

      <section className="panel">
        <header className="panel-header">
          <p className="eyebrow">策略对局</p>
          <h2>选择你的开局配置</h2>
        </header>

        <div className="option-group">
          <label>模式</label>
          <div className="pill-row">
            <button className={`pill ${mode === 'PVE' ? 'selected' : ''}`} onClick={() => setMode('PVE')}>人机对战</button>
            <button className={`pill ${mode === 'PVP' ? 'selected' : ''}`} onClick={() => setMode('PVP')}>双人对战</button>
            <button className={`pill ${mode === 'PVP_ONLINE' ? 'selected' : ''}`} onClick={() => setMode('PVP_ONLINE')}>在线双人</button>
          </div>
        </div>

        <div className="option-group">
          <label>执子</label>
          <div className="pill-row">
            <button className={`pill ${color === 'BLACK' ? 'selected' : ''}`} onClick={() => setColor('BLACK')}>黑方先手</button>
            <button className={`pill ${color === 'WHITE' ? 'selected' : ''}`} onClick={() => setColor('WHITE')}>白方后手</button>
          </div>
        </div>

        {mode === 'PVP_ONLINE' && (
          <div className="option-group">
            <label>配对码（4位数字）</label>
            <input value={pairCode} maxLength={4} inputMode="numeric" pattern="[0-9]*" placeholder="例如 1234" onChange={(e) => setPairCode(e.target.value)} />
          </div>
        )}

        <div className="option-group">
          <label>棋盘大小</label>
          <div className="pill-row">
            <button className={`pill ${size === 6 ? 'selected' : ''}`} onClick={() => setSize(6)}>6 x 6</button>
            <button className={`pill ${size === 8 ? 'selected' : ''}`} onClick={() => setSize(8)}>8 x 8</button>
            <button className={`pill ${size === 10 ? 'selected' : ''}`} onClick={() => setSize(10)}>10 x 10</button>
          </div>
        </div>

        {mode === 'PVE' && (
          <div className="option-group">
            <label>AI 算法</label>
            <div className="pill-row">
              {algorithms.map((a) => (
                <button key={a} className={`pill ${aiAlgorithm === a ? 'selected' : ''}`} onClick={() => setAiAlgorithm(a)}>{a}</button>
              ))}
            </div>
            <div className="pill-row compact">
              {levels.map((lv) => (
                <button key={lv} className={`pill ${aiLevel === lv ? 'selected' : ''}`} onClick={() => setAiLevel(lv)}>{lv}</button>
              ))}
            </div>
          </div>
        )}

        <button id="startBtn" onClick={handleStart} disabled={wsStatus !== 'connected'}>
          {wsStatus === 'connected' ? '开始游戏' : '连接中...'}
        </button>
      </section>
    </div>
  )
}
