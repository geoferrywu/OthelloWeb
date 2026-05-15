# Othello 前后端分离设计方案

## 1. 项目概述

### 背景
当前 [Othello.html](Othello.html) 是一个 914 行的单文件黑白棋游戏，包含 UI 渲染、游戏逻辑和 AI 引擎全部代码。需要拆分为前后端分离架构，支持人机对战（PvE）和人人对战（PvP）两种模式。

### 技术栈
- **后端**: Go (原生 `net/http` + `gorilla/websocket`)
- **前端**: Vue 3 + Vite + TypeScript
- **通信**: WebSocket 实时双向通信
- **AI**: 前端（轻量级，仅提示功能）+ 后端（完整版，用于 PvE 落子）

## 2. 项目结构

```
othello-game/
├── backend/
│   ├── go.mod
│   ├── main.go                    # HTTP server + WebSocket handler
│   └── game/
│       ├── engine.go              # 核心游戏逻辑（棋盘、落子验证、翻转）
│       ├── ai.go                  # AI 引擎（minimax + alpha-beta）
│       ├── state.go               # 游戏状态结构体
│       └── manager.go             # 会话管理（PvE/PvP 匹配）
└── frontend/
    ├── package.json
    ├── vite.config.js
    ├── index.html
    └── src/
        ├── main.ts
        ├── App.vue
        ├── types/
        │   └── index.ts           # TypeScript 类型定义
        ├── composable/
        │   └── useWebSocket.ts    # WebSocket 连接与消息处理
        ├── components/
        │   ├── StartScreen.vue    # 模式/颜色/棋盘大小选择
        │   ├── GameBoard.vue      # 棋盘渲染、点击交互、动画
        │   ├── Scoreboard.vue     # 分数面板
        │   ├── HistoryPanel.vue   # 可拖拽落子记录
        │   ├── Modal.vue          # 游戏结束弹窗
        │   └── ControlPanel.vue   # 悔棋/记录/提示/返回按钮
        └── entities/
            └── game.ts            # 前端轻量 AI（提示用）
```

## 3. 架构设计

### 3.1 职责划分

```
┌─────────────────────────────────────────────────────────────┐
│                       前端 (Vue 3)                          │
│  • UI 渲染（棋盘、分数、历史记录、弹窗）                        │
│  • 用户交互处理（点击落子、按钮操作）                           │
│  • 轻量 AI 计算（仅用于提示功能）                              │
│  • CSS 动画（翻转、脉冲、过渡）                                │
└───────────────────────┬─────────────────────────────────────┘
                        │ WebSocket (JSON)
┌───────────────────────┴─────────────────────────────────────┐
│                       后端 (Go)                             │
│  • 游戏状态管理（唯一权威）                                    │
│  • 落子合法性验证                                             │
│  • AI 引擎（minimax + alpha-beta，完整版）                     │
│  • PvE 会话管理（单人 vs AI）                                 │
│  • PvP 会话管理（双人匹配 + 状态同步）                          │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 状态权威规则

| 职责 | 后端 | 前端 |
|------|:----:|:----:|
| 棋盘状态 | 唯一权威 | 镜像渲染 |
| 合法移动计算 | 计算并下发 | 接收展示 |
| AI 落子 | 完整版（深度 4-6） | 轻量版（提示，浅一层） |
| 胜负判断 | 最终裁决 | 弹窗展示 |
| 历史记录 | 维护 | 滚动展示 |
| UI 状态 | - | 面板显隐、拖拽位置 |

## 4. WebSocket 消息协议

### 4.1 客户端 → 服务端

| 类型 | 字段 | 说明 |
|------|------|------|
| `JOIN` | `{mode: "PVE"\|"PVP", color: "BLACK"\|"WHITE", size: 6\|8\|10}` | 加入/开始游戏 |
| `MOVE` | `{r: number, c: number}` | 落子请求 |
| `UNDO` | `{}` | 悔棋请求（仅 PvE） |
| `PING` | `{}` | 心跳保活 |

### 4.2 服务端 → 客户端

| 类型 | 字段 | 说明 |
|------|------|------|
| `INIT` | `{gameId, board, currentPlayer, players}` | 游戏初始化完成 |
| `STATE` | `{board, currentPlayer, lastMove, flipped}` | 状态更新（玩家落子后） |
| `AI_MOVE` | `{r, c, flipped}` | AI 落子通知 |
| `GAME_OVER` | `{winner, blackScore, whiteScore}` | 游戏结束 |
| `ERROR` | `{message}` | 错误信息 |
| `OPPONENT_LEFT` | `{}` | PvP: 对手断开连接 |
| `PLAYER_JOINED` | `{player}` | PvP: 另一名玩家加入 |

### 4.3 PvE 消息流示例

```
客户端                          服务端
  │                               │
  │─── JOIN {mode:"PVE",          │
  │         color:"BLACK",        │
  │         size:8} ─────────────>│
  │                               │ 初始化棋盘
  │<── INIT {gameId, board,       │
  │           currentPlayer,      │
  │           players} ───────────│
  │                               │
  │─── MOVE {r:3, c:4} ─────────>│ 验证落子
  │                               │ 计算翻转
  │<── STATE {board, currentPlayer}│ 返回新状态
  │                               │ AI 计算
  │<── AI_MOVE {r, c, flipped} ──│ AI 落子
  │                               │
```

### 4.4 PvP 消息流示例

```
客户端 A                        服务端                        客户端 B
  │                               │                               │
  │─── JOIN {mode:"PVP",          │                               │
  │         color:"BLACK"} ──────>│                               │
  │<── WAITING ───────────────────│ 等待第二名玩家                  │
  │                               │                               │
  │                               │<── JOIN {mode:"PVP",          │
  │                               │         color:"WHITE",        │
  │                               │         gameId: "..."} ───────│
  │                               │ 匹配成功                       │
  │<── INIT {gameId, board} ──────│─── INIT {gameId, board} ─────>│
  │                               │                               │
  │─── MOVE {r:3, c:4} ─────────>│                               │
  │                               │ 验证并广播                     │
  │<── STATE {board, ...} ────────│─── STATE {board, ...} ───────>│
  │                               │                               │
```

## 5. AI 设计

### 5.1 后端 AI（`backend/game/ai.go`）

| 属性 | 说明 |
|------|------|
| 算法 | Minimax + Alpha-Beta 剪枝 |
| 评估函数 | 位置权重矩阵（角落=120, 边缘=20, 近角=-40） |
| 搜索深度 | 6×6: depth 6 / 8×8: depth 4 / 10×10: depth 3 |
| 移动排序 | 按翻转数降序，优化剪枝效率 |
| 用途 | PvE 模式中 AI 的实际落子 |

### 5.2 前端 AI（`frontend/src/entities/game.ts`）

| 属性 | 说明 |
|------|------|
| 算法 | 与后端相同的 Minimax，但搜索深度减 1 |
| 用途 | 仅用于"提示"功能，高亮推荐落子位置 |
| 特点 | 纯函数调用，不依赖服务端，计算量小 |

## 6. 数据模型

### 6.1 后端 Go 结构体

```go
// game/state.go
type Player int
const (
    EMPTY Player = iota
    BLACK
    WHITE
)

type Position struct {
    R int `json:"r"`
    C int `json:"c"`
}

type GameState struct {
    Board         [][]Player `json:"board"`
    CurrentPlayer Player     `json:"currentPlayer"`
    Size          int        `json:"size"`
    History       []Move     `json:"history"`
    GameOver      bool       `json:"gameOver"`
}

type Move struct {
    Player   Player     `json:"player"`
    Position *Position  `json:"position"` // nil 表示跳过
    Flipped  []Position `json:"flipped"`
}
```

### 6.2 前端 TypeScript 类型

```typescript
// src/types/index.ts
type Player = 0 | 1 | 2; // 0=EMPTY, 1=BLACK, 2=WHITE
type GameMode = 'PVE' | 'PVP';

interface Position { r: number; c: number; }
interface BoardState { board: Player[][]; currentPlayer: Player; size: number; }
interface MoveRecord { player: Player; position: Position | null; flipped: Position[]; }
```

## 7. 前端组件设计

### 7.1 组件树

```
App.vue
├── LoadingState（连接中）
├── StartScreen.vue
│   ├── 模式选择（PvE / PvP）
│   ├── 颜色选择（黑棋先手 / 白棋）
│   ├── 棋盘大小（6×6 / 8×8 / 10×10）
│   └── 开始按钮
└── GameScreen.vue
    ├── Scoreboard.vue（黑白分数 + 当前玩家高亮）
    ├── StatusBar.vue（状态文字："黑棋落子" / "AI思考中"）
    ├── GameBoard.vue（棋盘网格 + 棋子 + 提示点 + 动画）
    ├── HistoryPanel.vue（可拖拽浮层，滚动落子记录）
    ├── ControlPanel.vue
    │   ├── 悔棋按钮
    │   ├── 历史记录按钮
    │   ├── 提示开关按钮
    │   └── 返回按钮
    └── Modal.vue（游戏结束弹窗 + 重玩按钮）
```

### 7.2 状态管理

使用 Vue 3 `reactive` + `ref` 进行组件内状态管理，通过 composable `useWebSocket` 集中处理所有 WS 消息：

```typescript
// useWebSocket 暴露的接口
interface UseWebSocketReturn {
  status: Ref<'disconnected' | 'connecting' | 'connected'>;
  gameState: Ref<BoardState | null>;
  moveLog: Ref<MoveRecord[]>;
  sendMove: (r: number, c: number) => void;
  sendUndo: () => void;
  joinGame: (mode: GameMode, color: string, size: number) => void;
}
```

## 8. 实施计划

### Phase 1: 后端基础（~2小时）
1. 初始化 Go 项目 (`go mod init othello-backend`)
2. `game/state.go` — 数据结构定义
3. `game/engine.go` — 核心游戏逻辑（从 JS 移植）
4. `game/ai.go` — minimax + alpha-beta AI
5. `game/manager.go` — 会话管理
6. `main.go` — HTTP + WebSocket 服务

### Phase 2: 前端基础（~1.5小时）
7. Vite + Vue 3 项目初始化
8. `src/types/index.ts` — TS 类型定义
9. `src/composable/useWebSocket.ts` — WS 连接封装

### Phase 3: UI 组件（~3小时）
10. `App.vue` — 主入口 + 连接状态
11. `StartScreen.vue` — 开始界面
12. `GameBoard.vue` — 棋盘组件
13. `Scoreboard.vue` — 分数面板
14. `HistoryPanel.vue` — 历史记录
15. `Modal.vue` — 弹窗组件
16. `ControlPanel.vue` — 控制面板

### Phase 4: 集成联调（~2小时）
17. `entities/game.ts` — 前端轻量 AI
18. PvE 全流程联调
19. PvP 双人对战联调
20. UI 打磨、动画、错误处理

## 9. 验证方案

1. **启动**: `bin/start-all.sh`（Linux）或 `bin\start-all.cmd`（Windows）一键启动，端口配置见 `.env`
2. **PvE 测试**: 选择 PvE 模式 → 验证开局 → 交替落子 → 验证 AI 回应 → 触发跳过 → 验证终局
3. **PvP 测试**: 开两个浏览器窗口 → 分别加入同一房间 → 验证状态同步 → 断线测试
4. **功能测试**: 悔棋、提示开关、历史记录显隐、棋盘大小切换
5. **边界测试**: 6×6 / 10×10 棋盘、双方无子可落平局、AI 连续跳过
