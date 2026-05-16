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

if "%OTHELLO_FRONTEND_REACT_PORT%"=="" set "OTHELLO_FRONTEND_REACT_PORT=5174"

set "BACKEND_TITLE=Othello Backend"
set "FRONTEND_VUE_TITLE=Othello Frontend (Vue)"
set "FRONTEND_REACT_TITLE=Othello Frontend (React)"

echo Stopping Othello services...

for %%P in (%OTHELLO_FRONTEND_PORT% %OTHELLO_FRONTEND_REACT_PORT% %OTHELLO_BACKEND_PORT%) do (
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

taskkill /FI "WINDOWTITLE eq %FRONTEND_VUE_TITLE%" /T /F >nul 2>&1
if %ERRORLEVEL%==0 (
  echo - Vue frontend window closed.
) else (
  echo - Vue frontend window not found.
)

taskkill /FI "WINDOWTITLE eq %FRONTEND_REACT_TITLE%" /T /F >nul 2>&1
if %ERRORLEVEL%==0 (
  echo - React frontend window closed.
) else (
  echo - React frontend window not found.
)

echo Done.
endlocal
