#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Load shared config
set -a
source "$ROOT/.env"
set +a

LOG_DIR="$ROOT/logs"
CACHE_DIR="$ROOT/.gocache"
OTHELLO_FRONTEND_REACT_PORT="${OTHELLO_FRONTEND_REACT_PORT:-5174}"

mkdir -p "$LOG_DIR" "$CACHE_DIR"

echo "Starting Othello services..."

# Start backend
(
  cd "$ROOT/backend"
  GOCACHE="$CACHE_DIR" go run main.go >"$LOG_DIR/backend.log" 2>&1
) &
BACKEND_PID=$!
echo "Backend (port $OTHELLO_BACKEND_PORT) PID: $BACKEND_PID"
echo "$BACKEND_PID" >"$ROOT/.backend.pid"

echo "Waiting for backend to be ready on port $OTHELLO_BACKEND_PORT..."
backend_ready=false
for _ in {1..60}; do
  if (echo >"/dev/tcp/127.0.0.1/$OTHELLO_BACKEND_PORT") >/dev/null 2>&1; then
    backend_ready=true
    break
  fi
  sleep 1
done

if [ "$backend_ready" != true ]; then
  echo "Backend did not become ready within 60s. Check $LOG_DIR/backend.log"
  exit 1
fi

START_TARGET="${1:-}"
if [[ -z "$START_TARGET" ]]; then
  echo ""
  echo "Select frontend to start:"
  echo "  1) Vue frontend (frontend, port $OTHELLO_FRONTEND_PORT)"
  echo "  2) React frontend (frontend_rct, port $OTHELLO_FRONTEND_REACT_PORT)"
  echo "  3) Both (default)"
  read -r -p "Enter choice (1/2/3): " START_TARGET
fi
START_TARGET="${START_TARGET:-3}"

if [[ "$START_TARGET" == "1" || "$START_TARGET" == "vue" ]]; then
  echo "Starting Vue frontend..."
  (
    cd "$ROOT/frontend"
    npm run dev -- --host 0.0.0.0 --port "$OTHELLO_FRONTEND_PORT" >"$LOG_DIR/frontend.log" 2>&1
  ) &
  FRONTEND_PID=$!
  echo "Vue frontend (port $OTHELLO_FRONTEND_PORT) PID: $FRONTEND_PID"
  echo "$FRONTEND_PID" >"$ROOT/.frontend.pid"
elif [[ "$START_TARGET" == "2" || "$START_TARGET" == "react" ]]; then
  echo "Starting React frontend..."
  (
    cd "$ROOT/frontend_rct"
    npm run dev -- --host 0.0.0.0 --port "$OTHELLO_FRONTEND_REACT_PORT" >"$LOG_DIR/frontend_rct.log" 2>&1
  ) &
  FRONTEND_RCT_PID=$!
  echo "React frontend (port $OTHELLO_FRONTEND_REACT_PORT) PID: $FRONTEND_RCT_PID"
  echo "$FRONTEND_RCT_PID" >"$ROOT/.frontend_rct.pid"
else
  echo "Starting Vue frontend..."
  (
    cd "$ROOT/frontend"
    npm run dev -- --host 0.0.0.0 --port "$OTHELLO_FRONTEND_PORT" >"$LOG_DIR/frontend.log" 2>&1
  ) &
  FRONTEND_PID=$!
  echo "Vue frontend (port $OTHELLO_FRONTEND_PORT) PID: $FRONTEND_PID"
  echo "$FRONTEND_PID" >"$ROOT/.frontend.pid"

  echo "Starting React frontend..."
  (
    cd "$ROOT/frontend_rct"
    npm run dev -- --host 0.0.0.0 --port "$OTHELLO_FRONTEND_REACT_PORT" >"$LOG_DIR/frontend_rct.log" 2>&1
  ) &
  FRONTEND_RCT_PID=$!
  echo "React frontend (port $OTHELLO_FRONTEND_REACT_PORT) PID: $FRONTEND_RCT_PID"
  echo "$FRONTEND_RCT_PID" >"$ROOT/.frontend_rct.pid"
fi

sleep 2

echo ""
echo "Othello is starting:"
echo "  - Vue Frontend:        http://localhost:$OTHELLO_FRONTEND_PORT"
echo "  - React Frontend:      http://localhost:$OTHELLO_FRONTEND_REACT_PORT"
echo "  - Backend WebSocket:   ws://localhost:$OTHELLO_BACKEND_PORT/ws/game"
echo ""
echo "Logs: $LOG_DIR/"
echo "Run bin/stop-all.sh to stop all services."
