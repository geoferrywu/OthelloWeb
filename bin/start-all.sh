#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Load shared config
set -a
source "$ROOT/.env"
set +a

LOG_DIR="$ROOT/logs"
CACHE_DIR="$ROOT/.gocache"

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

echo "Backend is ready. Starting frontend..."

# Start frontend
(
  cd "$ROOT/frontend"
  npm run dev -- --host 0.0.0.0 --port "$OTHELLO_FRONTEND_PORT" >"$LOG_DIR/frontend.log" 2>&1
) &
FRONTEND_PID=$!
echo "Frontend (port $OTHELLO_FRONTEND_PORT) PID: $FRONTEND_PID"
echo "$FRONTEND_PID" >"$ROOT/.frontend.pid"

sleep 2

echo ""
echo "Othello is starting:"
echo "  - Frontend:            http://localhost:$OTHELLO_FRONTEND_PORT"
echo "  - Backend WebSocket:   ws://localhost:$OTHELLO_BACKEND_PORT/ws/game"
echo ""
echo "Logs: $LOG_DIR/"
echo "Run bin/stop-all.sh to stop all services."
