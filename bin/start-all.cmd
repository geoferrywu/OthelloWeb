echo off
setlocal

set "ROOT=%~dp0..\\"
set "LOG_DIR=%ROOT%logs"
set "CACHE_DIR=%ROOT%.gocache"
if not exist "%LOG_DIR%" mkdir "%LOG_DIR%"
if not exist "%CACHE_DIR%" mkdir "%CACHE_DIR%"

echo Starting backend...
start "Othello Backend" cmd /k "cd /d ""%ROOT%backend"" && set ""GOCACHE=%CACHE_DIR%"" && go run main.go"

echo Starting frontend...
start "Othello Frontend" cmd /k "cd /d ""%ROOT%frontend"" && npm run dev -- --host 0.0.0.0 --port 5173"

echo.
echo Othello is starting:
echo - Frontend: http://localhost:5173
echo - Backend WebSocket: ws://localhost:8080/ws/game
echo.
echo Keep the two opened windows running while you play.

endlocal
