@echo off
setlocal EnableDelayedExpansion

set "ROOT=%~dp0..\"

:: Load shared config (parse KEY=VALUE lines)
for /f "usebackq tokens=1,2 delims==" %%A in ("%ROOT%.env") do (
  set "line=%%A"
  if not "!line:~0,1!"=="#" if not "!line!"=="" (
    set "%%A=%%B"
  )
)

set "LOG_DIR=%ROOT%logs"
set "CACHE_DIR=%ROOT%.gocache"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"
if not exist "%CACHE_DIR%" mkdir "%CACHE_DIR%"

echo Starting backend...
start "Othello Backend" cmd /k "cd /d ""%ROOT%backend"" && set ""GOCACHE=%CACHE_DIR%"" && go run main.go"

echo Starting frontend...
start "Othello Frontend" cmd /k "cd /d ""%ROOT%frontend"" && npm run dev -- --host 0.0.0.0 --port %OTHELLO_FRONTEND_PORT%"

echo.
echo Othello is starting:
echo - Frontend: http://localhost:%OTHELLO_FRONTEND_PORT%
echo - Backend WebSocket: ws://localhost:%OTHELLO_BACKEND_PORT%/ws/game
echo.
echo Keep the two opened windows running while you play.

endlocal
