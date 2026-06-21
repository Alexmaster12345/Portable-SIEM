# Portable SIEM — Windows Installation Guide

## Which installer do you need?

| Windows Version | Method | Installer |
|----------------|--------|-----------|
| Windows 7 / 8 / 8.1 | Docker Toolbox + VirtualBox | `install-toolbox.bat` |
| Windows 10 (before 2004, build < 19041) | Docker Toolbox + VirtualBox | `install-toolbox.bat` |
| Windows 10 (2004+, build ≥ 19041) | Docker Desktop + WSL2 | `install.ps1` |
| Windows 11 | Docker Desktop + WSL2 | `install.ps1` |

---

## Windows 10 (2004+) and Windows 11

### Automatic installation

Open **PowerShell as Administrator** and run:

```powershell
Set-ExecutionPolicy Bypass -Scope Process -Force
cd "D:\Portable SIEM"   # replace D: with your USB drive letter
.\deploy\windows\install.ps1
```

This will automatically:
1. Enable WSL2 and VirtualMachinePlatform features
2. Install Docker Desktop via Chocolatey
3. Start all SIEM containers
4. Open the dashboard in your browser

> **Note:** After WSL2 is enabled for the first time you may need to restart your PC, then re-run the script.

### Manual installation (step by step)

**Step 1 — Enable WSL2**

Open PowerShell as Administrator:
```powershell
wsl --install
wsl --set-default-version 2
```
Restart your PC when prompted.

**Step 2 — Install Docker Desktop**

Download from: https://www.docker.com/products/docker-desktop/

- Run the installer with default settings
- Start Docker Desktop from the Start Menu
- Wait for the whale icon in the system tray to stop animating

**Step 3 — Start the SIEM**

Open PowerShell (no admin needed):
```powershell
cd "D:\Portable SIEM"   # your USB drive letter
docker compose up -d
```

**Step 4 — Open the dashboard**

Navigate to: http://localhost:3000

---

## Windows 7 / 8 / 8.1 (and old Windows 10)

Docker Desktop does **not** support Windows 7/8. Use **Docker Toolbox** instead, which runs Docker inside a VirtualBox VM.

### Step 1 — Install VirtualBox

Download and install from: https://www.virtualbox.org/wiki/Downloads

- Choose: **Windows hosts**
- Run the installer with default settings

### Step 2 — Install Docker Toolbox

Download from: https://github.com/docker/toolbox/releases/latest

- Download: `DockerToolbox-XX.XX.X.exe`
- Run installer, keep all default options checked
- Installs: Docker, Docker Machine, Docker Compose, Kitematic

### Step 3 — Run the installer

Right-click `deploy\windows\install-toolbox.bat` → **Run as Administrator**

This will:
1. Create a VirtualBox Docker machine named `default`
2. Start all SIEM containers
3. Open the dashboard in your browser

> **IMPORTANT:** On Docker Toolbox, the SIEM is **not** at `localhost`.
> It runs at the Docker Machine IP address, typically `192.168.99.100`.
> The installer will show you the correct URL.

### Get the Docker Machine IP manually

Open **Docker Quickstart Terminal** and run:
```bash
docker-machine ip default
```

Then open: `http://<that-ip>:3000`

### Manual steps (Docker Toolbox)

Open **Docker Quickstart Terminal** (installed by Docker Toolbox):

```bash
# Navigate to USB drive
cd /d/Portable\ SIEM     # D: drive in Git Bash / MinGW

# Start the SIEM
docker-compose up -d

# Get the IP
docker-machine ip default
```

Open your browser at `http://<docker-machine-ip>:3000`

---

## Seed Test Data (populate dashboard)

**PowerShell (Windows 10/11):**
```powershell
cd "D:\Portable SIEM"
.\deploy\windows\seed-test-data.ps1
```

**Docker Quickstart Terminal (Windows 7/8):**
```bash
cd /d/Portable\ SIEM
./deploy/usb/seed-test-data.sh
```

---

## Stopping the SIEM

**PowerShell / CMD:**
```powershell
cd "D:\Portable SIEM"
docker compose down
```

**Docker Toolbox:**
```bash
docker-compose down
docker-machine stop default   # optional: stop the VM too
```

---

## Deploying the Linux Agent on Remote Hosts

From the SIEM machine, get your IP:
```powershell
ipconfig
# or on Docker Toolbox:
docker-machine ip default
```

On each Linux host you want to monitor:
```bash
# Copy the agent from the USB
scp /path/to/usb/siem-agent-linux user@target:/tmp/

# Run it
/tmp/siem-agent-linux -server http://<SIEM-IP>:8888
```

---

## Firewall Rules (Windows Defender)

If agents cannot reach the SIEM, open the required ports:

```powershell
# Run as Administrator
New-NetFirewallRule -DisplayName "SIEM Dashboard" -Direction Inbound -Protocol TCP -LocalPort 3000 -Action Allow
New-NetFirewallRule -DisplayName "SIEM API"       -Direction Inbound -Protocol TCP -LocalPort 8888 -Action Allow
New-NetFirewallRule -DisplayName "SIEM Syslog"    -Direction Inbound -Protocol UDP -LocalPort 514  -Action Allow
New-NetFirewallRule -DisplayName "SIEM Agent"     -Direction Inbound -Protocol TCP -LocalPort 9000 -Action Allow
```

---

## Troubleshooting (Windows)

**Docker Desktop won't start:**
- Make sure virtualization is enabled in BIOS (Intel VT-x / AMD-V)
- On Windows 10 Home: ensure Hyper-V is enabled via WSL2

**Port already in use:**
```powershell
netstat -ano | findstr :8888
netstat -ano | findstr :3000
# Kill the process using the PID shown
taskkill /PID <PID> /F
```

**Docker Toolbox: `docker-machine` not found:**
- Open Docker Quickstart Terminal (not regular CMD)
- Or manually: `C:\Program Files\Docker Toolbox\docker-machine.exe`

**Dashboard shows blank / no data:**
```powershell
# Check containers are running
docker ps

# Check server logs
docker logs siem-server

# Seed test data
.\deploy\windows\seed-test-data.ps1
```

**On Docker Toolbox: cannot connect to `localhost`:**
- Use the Docker Machine IP, not localhost
- Run: `docker-machine ip default` to get it

**Rebuild after an update:**
```powershell
docker compose down
docker compose build siem-server
docker compose up -d
```
