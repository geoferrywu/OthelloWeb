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

echo Waiting for backend to be ready on port %OTHELLO_BACKEND_PORT%...
set "BACKEND_READY="
for /l %%I in (1,1,60) do (
  powershell -NoProfile -ExecutionPolicy Bypass -Command "try { $c = New-Object Net.Sockets.TcpClient; $c.Connect('127.0.0.1', %OTHELLO_BACKEND_PORT%); $c.Close(); exit 0 } catch { exit 1 }" >nul 2>&1
  if !errorlevel! equ 0 (
    set "BACKEND_READY=1"
    goto :backend_ready
  )
  timeout /t 1 /nobreak >nul
)

:backend_ready
if not defined BACKEND_READY (
  echo Backend did not become ready within 60s.
  echo Please check the backend window for errors.
  exit /b 1
)

if "%OTHELLO_FRONTEND_REACT_PORT%"=="" set "OTHELLO_FRONTEND_REACT_PORT=5174"

echo Backend is ready.
echo.
echo Select frontend to start:
echo [1] Vue frontend (frontend, port %OTHELLO_FRONTEND_PORT%)
echo [2] React frontend (frontend_rct, port %OTHELLO_FRONTEND_REACT_PORT%)
echo [3] Both (default)
set "START_TARGET=3"
set /p START_TARGET=Enter choice (1/2/3):
if "%START_TARGET%"=="" set "START_TARGET=3"

if "%START_TARGET%"=="1" (
  echo Starting Vue frontend...
  start "Othello Frontend (Vue)" cmd /k "cd /d ""%ROOT%frontend"" && npm run dev -- --host 0.0.0.0 --port %OTHELLO_FRONTEND_PORT%"
)

if "%START_TARGET%"=="2" (
  echo Starting React frontend...
  start "Othello Frontend (React)" cmd /k "cd /d ""%ROOT%frontend_rct"" && npm run dev -- --host 0.0.0.0 --port %OTHELLO_FRONTEND_REACT_PORT%"
)

if "%START_TARGET%"=="3" (
  echo Starting Vue frontend...
  start "Othello Frontend (Vue)" cmd /k "cd /d ""%ROOT%frontend"" && npm run dev -- --host 0.0.0.0 --port %OTHELLO_FRONTEND_PORT%"
  echo Starting React frontend...
  start "Othello Frontend (React)" cmd /k "cd /d ""%ROOT%frontend_rct"" && npm run dev -- --host 0.0.0.0 --port %OTHELLO_FRONTEND_REACT_PORT%"
)

echo.
echo Othello is starting:
echo - Vue Frontend: http://localhost:%OTHELLO_FRONTEND_PORT%
echo - React Frontend: http://localhost:%OTHELLO_FRONTEND_REACT_PORT%
echo - Backend WebSocket: ws://localhost:%OTHELLO_BACKEND_PORT%/ws/game
echo.
echo Keep the two opened windows running while you play.

endlocal
