:: This is a cmd file to stop service
:: Created by user
@echo off
setlocal EnableDelayedExpansion

set "BACKEND_TITLE=Othello Backend"
set "FRONTEND_TITLE=Othello Frontend"

echo Stopping Othello services...

for %%P in (5173 8080) do (
  for /f "tokens=5" %%I in ('netstat -ano ^| findstr /r /c:":%%P .*LISTENING"') do (
    taskkill /PID %%I /T /F >nul 2>&1
    if !ERRORLEVEL! EQU 0 (
      echo - Stopped PID %%I on port %%P.
    )
  )
)

taskkill /FI "WINDOWTITLE eq %BACKEND_TITLE%" /T /F >nul 2>&1
if %ERRORLEVEL%==0 (
  echo - Backend window closed.
) else (
  echo - Backend window not found.
)

taskkill /FI "WINDOWTITLE eq %FRONTEND_TITLE%" /T /F >nul 2>&1
if %ERRORLEVEL%==0 (
  echo - Frontend window closed.
) else (
  echo - Frontend window not found.
)

echo Done.
endlocal
