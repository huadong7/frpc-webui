@echo off
REM Build frpc-manager Docker image for deployment
REM Prerequisites: Docker Desktop running, Linux binary at bin\frpc-linux-amd64

set IMAGE_NAME=frpc-manager
set IMAGE_TAG=%1
if "%IMAGE_TAG%"=="" set IMAGE_TAG=latest

echo === Building frpc-manager Docker image ===
echo Image: %IMAGE_NAME%:%IMAGE_TAG%

REM Build Linux binary if needed
if not exist "bin\frpc-linux-amd64" (
    echo Building Linux binary first...
    set CGO_ENABLED=0
    set GOOS=linux
    set GOARCH=amd64
    go build -trimpath -ldflags "-s -w" -tags frpc -o bin\frpc-linux-amd64 .\cmd\frpc
)

REM Build Docker image
docker build -t %IMAGE_NAME%:%IMAGE_TAG% -f dockerfiles\Dockerfile .

echo.
echo === Build complete ===
echo Image: %IMAGE_NAME%:%IMAGE_TAG%
echo.
echo Save for transfer:
echo   docker save %IMAGE_NAME%:%IMAGE_TAG% -o frpc-manager-%IMAGE_TAG%.tar
echo.
echo Load on target machine:
echo   docker load -i frpc-manager-%IMAGE_TAG%.tar
echo.
echo Run:
echo   docker run -d --name frpc-manager -p 7400:7400 -v frpc_data:/data --restart unless-stopped %IMAGE_NAME%:%IMAGE_TAG%
pause
