#Requires -RunAsAdministrator
# Portable SIEM - Windows 10/11 Installer
# Run as Administrator in PowerShell

param(
    [string]$SiemPort  = "8888",
    [string]$DashPort  = "3000"
)

$ErrorActionPreference = "Stop"
$SiemDir = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Portable SIEM - Windows Setup" -ForegroundColor Cyan
Write-Host "  $(Get-CimInstance Win32_OperatingSystem | Select-Object -ExpandProperty Caption)" -ForegroundColor Gray
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# ---- Helper functions ----
function Test-Command($cmd) {
    return [bool](Get-Command $cmd -ErrorAction SilentlyContinue)
}

function Install-Chocolatey {
    if (Test-Command choco) { return }
    Write-Host "[1/4] Installing Chocolatey package manager..." -ForegroundColor Yellow
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    $env:PATH += ";$env:ALLUSERSPROFILE\chocolatey\bin"
    Write-Host "  Chocolatey installed." -ForegroundColor Green
}

function Install-DockerDesktop {
    if (Test-Command docker) {
        Write-Host "[1/4] Docker already installed." -ForegroundColor Green
        return
    }

    Write-Host "[1/4] Installing Docker Desktop..." -ForegroundColor Yellow

    $winVer = [System.Environment]::OSVersion.Version
    if ($winVer.Build -lt 19041) {
        Write-Host ""
        Write-Host "  ERROR: Docker Desktop requires Windows 10 version 2004 (build 19041) or higher." -ForegroundColor Red
        Write-Host "  Your build: $($winVer.Build)" -ForegroundColor Red
        Write-Host "  For Windows 7/8/8.1 or older Windows 10, use the Docker Toolbox installer:" -ForegroundColor Yellow
        Write-Host "    deploy\windows\install-toolbox.bat" -ForegroundColor White
        exit 1
    }

    Install-Chocolatey
    choco install docker-desktop -y --no-progress
    Write-Host "  Docker Desktop installed." -ForegroundColor Green
    Write-Host "  IMPORTANT: Restart your computer, then re-run this script." -ForegroundColor Yellow
    exit 0
}

function Enable-WSL2 {
    $wsl = Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux
    if ($wsl.State -ne "Enabled") {
        Write-Host "[2/4] Enabling WSL2..." -ForegroundColor Yellow
        Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux -NoRestart | Out-Null
        Enable-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform -NoRestart | Out-Null
        wsl --set-default-version 2
        Write-Host "  WSL2 enabled." -ForegroundColor Green
    } else {
        Write-Host "[2/4] WSL2 already enabled." -ForegroundColor Green
    }
}

function Wait-DockerReady {
    Write-Host "[3/4] Waiting for Docker to be ready..." -ForegroundColor Yellow
    for ($i = 1; $i -le 30; $i++) {
        try {
            docker info 2>&1 | Out-Null
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  Docker is ready." -ForegroundColor Green
                return
            }
        } catch {}
        Write-Host "  attempt $i/30..."
        Start-Sleep 5
    }
    Write-Host "  ERROR: Docker did not start in time. Start Docker Desktop manually and retry." -ForegroundColor Red
    exit 1
}

function Start-SIEM {
    Write-Host "[4/4] Starting Portable SIEM..." -ForegroundColor Yellow
    Set-Location $SiemDir

    docker compose up -d postgres redis nats
    Write-Host "  Waiting for database..."
    Start-Sleep 10

    docker compose up -d siem-server siem-frontend

    Write-Host "  Waiting for SIEM server..."
    for ($i = 1; $i -le 20; $i++) {
        try {
            $r = Invoke-WebRequest -Uri "http://localhost:8888/health" -UseBasicParsing -TimeoutSec 2
            if ($r.StatusCode -eq 200) {
                Write-Host "  Server is up!" -ForegroundColor Green
                break
            }
        } catch {}
        Write-Host "  attempt $i/20..."
        Start-Sleep 3
    }
}

# ---- Run ----
Enable-WSL2
Install-DockerDesktop
Wait-DockerReady
Start-SIEM

$localIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.InterfaceAlias -notmatch "Loopback" } | Select-Object -First 1).IPAddress

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Portable SIEM is running!" -ForegroundColor Green
Write-Host ""
Write-Host "  Dashboard : http://localhost:3000" -ForegroundColor White
Write-Host "  API       : http://localhost:8888/api/v1" -ForegroundColor White
Write-Host "  Health    : http://localhost:8888/health" -ForegroundColor White
Write-Host ""
Write-Host "  From network: http://${localIP}:3000" -ForegroundColor White
Write-Host ""
Write-Host "  Seed test data:" -ForegroundColor Gray
Write-Host "    .\deploy\windows\seed-test-data.ps1" -ForegroundColor White
Write-Host ""
Write-Host "  Stop the SIEM:" -ForegroundColor Gray
Write-Host "    docker compose down" -ForegroundColor White
Write-Host "============================================" -ForegroundColor Cyan

# Open dashboard in browser
Start-Process "http://localhost:3000"
