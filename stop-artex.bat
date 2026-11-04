@echo off
chcp 65001 >nul 2>&1
REM ============================================================
REM  ARTEX local auto-stop
REM
REM  Stops only what this project starts:
REM    UI/API : 8787
REM    Proxy  : 8788
REM    PostgreSQL 127.0.0.1:5437 (data: .runtime\pgdata)
REM ============================================================

setlocal EnableExtensions
set "ROOT=%~dp0"
set "PG_BIN=%ROOT%.runtime\postgres\pgsql\bin"
set "PG_DATA=%ROOT%.runtime\pgdata"

call :stop_port 8787 "ARTEX UI/API"
call :stop_port 8788 "ARTEX proxy"

echo Stopping PostgreSQL on 127.0.0.1:5437...
if exist "%PG_BIN%\pg_ctl.exe" (
    "%PG_BIN%\pg_ctl.exe" -D "%PG_DATA%" stop -m fast >nul 2>&1
    echo       done.
) else (
    echo       PostgreSQL binaries not found, skipped.
)

echo.
echo ARTEX development services stopped.
endlocal
exit /b 0

:stop_port
set "PORT=%~1"
set "LABEL=%~2"
echo Stopping %LABEL% on port %PORT%...
for /f "tokens=5" %%P in ('netstat -ano ^| findstr ":%PORT% " ^| findstr "LISTENING"') do (
    echo       Killing PID %%P ...
    taskkill /PID %%P /F >nul 2>&1
)
exit /b 0
