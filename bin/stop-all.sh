#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

# Load shared config
set -a
source "$ROOT/.env"
set +a

echo "Stopping Othello services..."

# Kill by saved PIDs
for pidfile in .backend.pid .frontend.pid; do
  pidfile_path="$ROOT/$pidfile"
  if [ -f "$pidfile_path" ]; then
    pid=$(cat "$pidfile_path")
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null && echo "- Stopped $pidfile (PID $pid)."
    else
      echo "- $pidfile (PID $pid) not running."
    fi
    rm -f "$pidfile_path"
  fi
done

# Also kill any remaining processes on the relevant ports
for port in "$OTHELLO_FRONTEND_PORT" "$OTHELLO_BACKEND_PORT"; do
  pids=$(lsof -ti :"$port" 2>/dev/null || true)
  if [ -n "$pids" ]; then
    echo "$pids" | xargs kill 2>/dev/null || true
    echo "- Killed process(es) on port $port."
  fi
done

echo "Done."
