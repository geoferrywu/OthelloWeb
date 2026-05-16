interface Props {
  canUndo: boolean
  gameOver: boolean
  showHistory: boolean
  showHint: boolean
  hintAlgorithm: string
  hintLevel: string
  disableHint?: boolean
  onUndo: () => void
  onToggleHistory: () => void
  onToggleHint: () => void
  onBack: () => void
  onHintAlgorithmChange: (value: string) => void
  onHintLevelChange: (value: string) => void
}

/**
 * 控制面板：集中处理操作按钮与提示参数。
 */
export default function ControlPanel(props: Props) {
  const algorithms = ['增强博弈', '主线剪枝', '蒙特树搜', '混合博弈']

  return (
    <div className="controls-wrap">
      <div className="btn-row">
        <button disabled={!props.canUndo || props.gameOver} onClick={props.onUndo}>悔棋</button>
        <button onClick={props.onToggleHistory}>{props.showHistory ? '隐藏记录' : '显示记录'}</button>
        <button disabled={props.disableHint} onClick={props.onToggleHint}>提示: {props.showHint ? '开' : '关'}</button>
        <button onClick={props.onBack}>返回</button>
      </div>

      {props.showHint && !props.disableHint && (
        <div className="hint-config">
          <select value={props.hintAlgorithm} onChange={(e) => props.onHintAlgorithmChange(e.target.value)}>
            {algorithms.map((a) => (<option key={a} value={a}>{a}</option>))}
          </select>
          <select value={props.hintLevel} onChange={(e) => props.onHintLevelChange(e.target.value)}>
            <option value="easy">easy</option>
            <option value="normal">normal</option>
            <option value="hard">hard</option>
          </select>
        </div>
      )}
    </div>
  )
}
