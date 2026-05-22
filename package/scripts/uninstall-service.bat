@echo off
setlocal

set SERVICE_NAME=frpc-manager

echo === frpc Manager - Windows Service Uninstaller ===

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: This script must be run as Administrator.
    pause
    exit /b 1
)

echo Stopping service...
net stop %SERVICE_NAME% >nul 2>&1

REM Find frpc.exe
if exist "%~dp0frpc.exe" (
    set FRPC_BIN=%~dp0frpc.exe
) else if exist "C:\Program Files\frp\frpc.exe" (
    set FRPC_BIN=C:\Program Files\frp\frpc.exe
) else (
    echo ERROR: Cannot find frpc.exe
    pause
    exit /b 1
)

echo Uninstalling service...
"%FRPC_BIN%" service uninstall

if %errorlevel% equ 0 (
    echo Service uninstalled successfully.
) else (
    echo ERROR: Service uninstallation failed.
)

pause
