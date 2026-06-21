@echo off
:: Portable SIEM - Windows 7 / 8 / 8.1 Setup
:: Uses Docker Toolbox (VirtualBox-based) for older Windows versions
:: Run as Administrator

title Portable SIEM - Windows 7/8 Setup
color 0B

echo.
echo ============================================
echo   Portable SIEM - Windows 7/8 Setup
echo   Using Docker Toolbox
echo ============================================
echo.

:: Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Please run as Administrator.
    echo Right-click this file and choose "Run as administrator"
    pause
    exit /b 1
)

:: ---- Step 1: Check VirtualBox ----
echo [1/4] Checking VirtualBox...
reg query "HKEY_LOCAL_MACHINE\SOFTWARE\Oracle\VirtualBox" >nul 2>&1
if %errorlevel% neq 0 (
    echo   VirtualBox not found.
    echo   Downloading VirtualBox installer...
    echo.
    echo   Please download and install VirtualBox from:
    echo   https://www.virtualbox.org/wiki/Downloads
    echo.
    echo   Then re-run this script.
    start https://www.virtualbox.org/wiki/Downloads
    pause
    exit /b 1
) else (
    echo   VirtualBox found. OK
)

:: ---- Step 2: Check Docker Toolbox ----
echo [2/4] Checking Docker Toolbox...
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo   Docker Toolbox not found.
    echo.
    echo   Please download and install Docker Toolbox from:
    echo   https://github.com/docker/toolbox/releases/latest
    echo.
    echo   Download: DockerToolbox-XX.XX.X.exe
    echo   Install with default settings.
    echo   Then re-run this script.
    start https://github.com/docker/toolbox/releases/latest
    pause
    exit /b 1
) else (
    echo   Docker Toolbox found. OK
)

:: ---- Step 3: Start Docker Machine ----
echo [3/4] Starting Docker machine...
docker-machine status default >nul 2>&1
if %errorlevel% neq 0 (
    echo   Creating Docker machine (first time - takes a few minutes)...
    docker-machine create --driver virtualbox default
)

docker-machine start default
for /f "tokens=*" %%i in ('docker-machine env default') do %%i

echo   Docker machine ready.

:: ---- Step 4: Start SIEM ----
echo [4/4] Starting Portable SIEM...

:: Navigate to SIEM directory (two levels up from deploy\windows\)
cd /d "%~dp0..\.."

docker-compose up -d postgres redis nats

echo   Waiting for database (15 seconds)...
timeout /t 15 /nobreak >nul

docker-compose up -d siem-server siem-frontend

echo   Waiting for SIEM to start (30 seconds)...
timeout /t 30 /nobreak >nul

:: Get Docker Machine IP (not localhost on Windows 7 Toolbox)
for /f "tokens=*" %%i in ('docker-machine ip default') do set DOCKER_IP=%%i

echo.
echo ============================================
echo   Portable SIEM is running!
echo.
echo   IMPORTANT: On Docker Toolbox, use the
echo   Docker Machine IP instead of localhost:
echo.
echo   Dashboard : http://%DOCKER_IP%:3000
echo   API       : http://%DOCKER_IP%:8888/api/v1
echo   Health    : http://%DOCKER_IP%:8888/health
echo.
echo   To stop: docker-compose down
echo   To get IP again: docker-machine ip default
echo ============================================
echo.

:: Open dashboard
start http://%DOCKER_IP%:3000

pause
