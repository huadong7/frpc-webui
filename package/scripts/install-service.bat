@echo off
setlocal

set SERVICE_NAME=frpc-manager
set DISPLAY_NAME=frp Client Manager

echo === frpc Manager - Windows Service Installer ===

REM Check for admin rights
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: This script must be run as Administrator.
    pause
    exit /b 1
)

REM Find frpc.exe
if exist "%~dp0frpc.exe" (
    set FRPC_BIN=%~dp0frpc.exe
) else if exist "C:\Program Files\frp\frpc.exe" (
    set FRPC_BIN=C:\Program Files\frp\frpc.exe
) else (
    echo ERROR: Cannot find frpc.exe
    echo Please place frpc.exe in the same directory as this script,
    echo or in C:\Program Files\frp\
    pause
    exit /b 1
)

echo Installing service...
echo Binary: %FRPC_BIN%

"%FRPC_BIN%" service install

if %errorlevel% equ 0 (
    echo.
    echo === Installation Complete ===
    echo Start the service: net start %SERVICE_NAME%
    echo   or: "%FRPC_BIN%" service start
    echo.
    echo Web UI: http://localhost:7400
    echo Data:   %PROGRAMDATA%\frp\frpc_profiles.json
) else (
    echo ERROR: Service installation failed.
)

pause
