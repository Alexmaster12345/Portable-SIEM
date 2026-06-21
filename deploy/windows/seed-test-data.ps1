# Portable SIEM - Seed test data (Windows PowerShell)
param([string]$Server = "http://localhost:8888")

# On Docker Toolbox (Windows 7/8), get the machine IP
if ($Server -eq "http://localhost:8888") {
    try {
        $toolboxIP = docker-machine ip default 2>$null
        if ($toolboxIP) { $Server = "http://${toolboxIP}:8888" }
    } catch {}
}

Write-Host "Seeding test data to $Server" -ForegroundColor Cyan

function Send-Event($body, $label) {
    try {
        Invoke-RestMethod -Uri "$Server/api/v1/events" -Method POST `
            -ContentType "application/json" -Body $body | Out-Null
        Write-Host "  OK: $label" -ForegroundColor Green
    } catch {
        Write-Host "  FAIL: $label - $($_.Exception.Message)" -ForegroundColor Red
    }
}

$ts = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

# SSH brute force
Write-Host "`n[1/5] SSH brute force events..." -ForegroundColor Yellow
for ($i = 1; $i -le 55; $i++) {
    Send-Event (@{
        timestamp  = $ts; host = "server01"; source = "auth"
        event_type = "login_failed"; severity = "medium"
        message    = "Failed password for admin from 1.2.3.4 port $i ssh2"
        fields     = @{ src_ip = "1.2.3.4"; username = "admin" }
    } | ConvertTo-Json) "ssh_fail #$i"
}

# Successful logins
Write-Host "`n[2/5] Successful logins..." -ForegroundColor Yellow
@("alice","bob","carol") | ForEach-Object {
    $user = $_
    Send-Event (@{
        timestamp  = $ts; host = "server01"; source = "auth"
        event_type = "login_success"; severity = "info"
        message    = "Accepted publickey for $user from 10.0.0.5 port 443 ssh2"
        fields     = @{ src_ip = "10.0.0.5"; username = $user }
    } | ConvertTo-Json) "login $user"
}

# Sudo / privilege escalation
Write-Host "`n[3/5] Sudo events..." -ForegroundColor Yellow
for ($i = 1; $i -le 6; $i++) {
    Send-Event (@{
        timestamp  = $ts; host = "server01"; source = "auth"
        event_type = "sudo"; severity = "medium"
        message    = "bob : TTY=pts/0 ; USER=root ; COMMAND=/bin/bash"
        fields     = @{ username = "bob" }
    } | ConvertTo-Json) "sudo #$i"
}

# Port scan (firewall denies)
Write-Host "`n[4/5] Firewall deny events (port scan)..." -ForegroundColor Yellow
for ($port = 20; $port -le 140; $port++) {
    Send-Event (@{
        timestamp  = $ts; host = "fw-edge"; source = "network_syslog"
        event_type = "firewall_deny"; severity = "medium"
        message    = "DENY SRC=5.5.5.5 DST=10.0.0.1 DPT=$port"
        fields     = @{ src_ip = "5.5.5.5"; dst_ip = "10.0.0.1"; dst_port = "$port" }
    } | ConvertTo-Json) "fw_deny port $port"
}

# New user created
Write-Host "`n[5/5] User account created..." -ForegroundColor Yellow
Send-Event (@{
    timestamp  = $ts; host = "server02"; source = "auth"
    event_type = "user_change"; severity = "high"
    message    = "useradd: new user: name=backdoor, UID=1337"
    fields     = @{ username = "backdoor" }
} | ConvertTo-Json) "useradd backdoor"

Write-Host "`nDone! Refresh the dashboard." -ForegroundColor Cyan
Write-Host "Stats:" -ForegroundColor Gray
try {
    Invoke-RestMethod -Uri "$Server/api/v1/events/stats" | ConvertTo-Json
} catch {
    Write-Host "Could not fetch stats: $($_.Exception.Message)" -ForegroundColor Red
}
