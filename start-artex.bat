@echo off
chcp 65001 >nul 2>&1
REM ============================================================
REM  ARTEX 本地启动（与控制台同生命周期）
REM
REM  行为：
REM    1. 数据库没起来时，由本控制台拉起 PostgreSQL（5437）
REM    2. ARTEX 前台运行在本控制台（8787 / 代理 8788）
REM
REM  因此：
REM    - 关闭这个窗口 或 Ctrl-C：ARTEX 立即退出
REM    - 若数据库是本窗口拉起的，ARTEX 退出后数据库也会一并停止
REM    - 想让它常驻，请改用服务方式（nssm / 计划任务），不要用本脚本
REM
REM  端口与 SERIES（5435 / 8797 / 8798）错开，可同时运行。
REM ============================================================

setlocal EnableExtensions EnableDelayedExpansion
set "ROOT=%~dp0"
set "PG_BIN=%ROOT%.runtime\postgres\pgsql\bin"
set "PG_DATA=%ROOT%.runtime\pgdata"
set "PG_LOG=%ROOT%.runtime\pg.log"
set "PG_PORT=5437"
set "APP_PORT=8787"
set "PROXY_PORT=8788"

if not exist "%PG_BIN%\pg_isready.exe" (
    echo [ERROR] 找不到内嵌 PostgreSQL：%PG_BIN%
    exit /b 1
)
if not exist "%ROOT%artex.exe" (
    echo [ERROR] 找不到 artex.exe：%ROOT%artex.exe
    echo         先构建：go build -tags embedui -o artex.exe ./cmd/artex
    exit /b 1
)

set "STARTED_PG=0"

echo [1/2] 检查 PostgreSQL 127.0.0.1:%PG_PORT% ...
"%PG_BIN%\pg_isready.exe" -h 127.0.0.1 -p %PG_PORT% >nul 2>&1
if errorlevel 1 (
    echo       未运行，正在随本控制台拉起 PostgreSQL ...
    start "ARTEX-PG" /b cmd /c ""%PG_BIN%\postgres.exe" -D "%PG_DATA%" -p %PG_PORT% -h 127.0.0.1 >> "%PG_LOG%" 2>&1"
    set "STARTED_PG=1"
    set "PG_READY=0"
    for /l %%N in (1,1,20) do (
        if "!PG_READY!"=="0" (
            "%PG_BIN%\pg_isready.exe" -h 127.0.0.1 -p %PG_PORT% >nul 2>&1
            if not errorlevel 1 set "PG_READY=1"
            if "!PG_READY!"=="0" timeout /t 1 /nobreak >nul
        )
    )
    if "!PG_READY!"=="0" (
        echo [ERROR] PostgreSQL 20 秒内未就绪，检查日志：%PG_LOG%
        exit /b 1
    )
    echo       数据库已就绪（随本窗口退出而停止）。
) else (
    echo       已在运行（由其它进程管理，本脚本退出时不会停止它）。
)

echo [2/2] 前台运行 ARTEX：UI http://localhost:%APP_PORT%  代理 %PROXY_PORT%
echo       关闭本窗口或按 Ctrl-C 即停止。
echo.

"%ROOT%artex.exe" -addr 127.0.0.1:%APP_PORT% -proxy 127.0.0.1:%PROXY_PORT%
set "CODE=!ERRORLEVEL!"

if "%STARTED_PG%"=="1" (
    echo.
    echo [cleanup] 停止本窗口拉起的 PostgreSQL ...
    "%PG_BIN%\pg_ctl.exe" -D "%PG_DATA%" stop -m fast >nul 2>&1
)

echo.
echo ARTEX 已退出，退出码 %CODE%
endlocal & exit /b %CODE%
